package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/tailored-agentic-units/kernel/agent"
	agentconfig "github.com/tailored-agentic-units/kernel/core/config"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/hub"
	"github.com/tailored-agentic-units/kernel/orchestrate/messaging"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== ISS Maintenance EVA - Agent Orchestration Demo ===")
	fmt.Println()

	// ============================================================================
	// 1. Load Agent Configurations
	// ============================================================================
	fmt.Println("1. Loading agent configurations...")

	// Load base configurations
	llamaConfig, err := agentconfig.LoadAgentConfig("examples/phase-01-hubs/config.llama.json")
	if err != nil {
		log.Fatalf("Failed to load llama config: %v", err)
	}

	gemmaConfig, err := agentconfig.LoadAgentConfig("examples/phase-01-hubs/config.gemma.json")
	if err != nil {
		log.Fatalf("Failed to load gemma config: %v", err)
	}

	// Override system prompts with ISS EVA operational context
	llamaConfig1 := &agentconfig.AgentConfig{
		Name: "eva-specialist-1",
		SystemPrompt: `You are a primary EVA specialist conducting external maintenance on the ISS.
Current task: Replacing cooling system component on starboard truss
EVA status: 2 hours 15 minutes elapsed, 3 hours 45 minutes remaining
Equipment: All tools accounted for, tether secured, suit systems nominal
Position: Starboard truss section S-3, 15 meters from airlock
Respond concisely in 1-2 sentences as if communicating over space-to-ground radio.`,
		Client:   llamaConfig.Client,
		Provider: llamaConfig.Provider,
		Model:    llamaConfig.Model,
	}

	llamaConfig2 := &agentconfig.AgentConfig{
		Name: "eva-specialist-2",
		SystemPrompt: `You are a secondary EVA specialist supporting external maintenance on the ISS.
Current task: Assisting cooling system repair, managing tool transfer
EVA status: 2 hours 15 minutes elapsed, 3 hours 45 minutes remaining
Equipment: Spare components secured, safety tether verified, suit nominal
Position: Starboard truss section S-2, maintaining visual contact with specialist-1
Respond concisely in 1-2 sentences as if communicating over space-to-ground radio.`,
		Client:   llamaConfig.Client,
		Provider: llamaConfig.Provider,
		Model:    llamaConfig.Model,
	}

	gemmaConfig.Name = "mission-commander"
	gemmaConfig.SystemPrompt = `You are the mission commander orchestrating an ISS EVA operation.
Mission status: Cooling system repair 40% complete, on schedule
Crew status: Both EVA specialists nominal, flight engineer monitoring
Environment: Orbital sunset in 25 minutes, next communication window in 12 minutes
Coordination: Managing EVA crew outside and support crew inside station
Respond concisely in 1 sentence as mission commander.`

	// Create flight engineer configuration
	llamaConfig3 := &agentconfig.AgentConfig{
		Name: "flight-engineer",
		SystemPrompt: `You are the flight engineer supporting EVA operations from inside the ISS.
Current task: Monitoring EVA crew vitals, managing airlock systems
Station status: All internal systems nominal, pressure stable
Support: Tools staged for retrieval, backup equipment ready
Monitoring: Tracking suit telemetry, communication relay, orbital position
Respond concisely in 1-2 sentences as flight engineer.`,
		Client:   llamaConfig.Client,
		Provider: llamaConfig.Provider,
		Model:    llamaConfig.Model,
	}

	// Create agents
	evaSpec1, err := agent.New(llamaConfig1)
	if err != nil {
		log.Fatalf("Failed to create eva-specialist-1: %v", err)
	}

	evaSpec2, err := agent.New(llamaConfig2)
	if err != nil {
		log.Fatalf("Failed to create eva-specialist-2: %v", err)
	}

	commander, err := agent.New(gemmaConfig)
	if err != nil {
		log.Fatalf("Failed to create mission-commander: %v", err)
	}

	flightEng, err := agent.New(llamaConfig3)
	if err != nil {
		log.Fatalf("Failed to create flight-engineer: %v", err)
	}

	fmt.Printf("  ✓ Created eva-specialist-1 (llama)\n")
	fmt.Printf("  ✓ Created eva-specialist-2 (llama)\n")
	fmt.Printf("  ✓ Created mission-commander (gemma)\n")
	fmt.Printf("  ✓ Created flight-engineer (llama)\n")
	fmt.Println()

	// ============================================================================
	// 2. Create Hubs
	// ============================================================================
	fmt.Println("2. Creating hubs...")

	// Configure logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	// Create EVA Hub (crew outside the station)
	evaConfig := config.DefaultHubConfig()
	evaConfig.Name = "eva-hub"
	evaConfig.Logger = logger
	evaHub := hub.New(ctx, evaConfig)
	defer evaHub.Shutdown(5 * time.Second)

	// Create ISS Hub (crew inside the station)
	issConfig := config.DefaultHubConfig()
	issConfig.Name = "iss-hub"
	issConfig.Logger = logger
	issHub := hub.New(ctx, issConfig)
	defer issHub.Shutdown(5 * time.Second)

	fmt.Printf("  ✓ Created eva-hub (EVA crew)\n")
	fmt.Printf("  ✓ Created iss-hub (ISS internal operations)\n")
	fmt.Println()

	// ============================================================================
	// 3. Create Message Handlers
	// ============================================================================

	// Channel for tracking responses
	responses := make(chan string, 10)

	// EVA Specialist 1 handler
	evaSpec1Handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("%v", msg.Data)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := evaSpec1.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Content()
		responses <- fmt.Sprintf("eva-specialist-1: %s", responseText)

		return nil, nil
	}

	// EVA Specialist 2 handler
	evaSpec2Handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("%v", msg.Data)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := evaSpec2.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Content()
		responses <- fmt.Sprintf("eva-specialist-2: %s", responseText)

		return nil, nil
	}

	// Mission Commander handler
	commanderHandler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("In %s: %v", msgCtx.HubName, msg.Data)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := commander.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Content()
		responses <- fmt.Sprintf("mission-commander (%s): %s", msgCtx.HubName, responseText)

		return nil, nil
	}

	// Flight Engineer handler
	flightEngHandler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("%v", msg.Data)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := flightEng.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Content()
		responses <- fmt.Sprintf("flight-engineer: %s", responseText)

		return nil, nil
	}

	// ============================================================================
	// 4. Register Agents with Hubs
	// ============================================================================
	fmt.Println("3. Registering agents with hubs...")

	// Register agents in EVA Hub
	if err := evaHub.RegisterAgent(evaSpec1, evaSpec1Handler); err != nil {
		log.Fatalf("Failed to register eva-specialist-1: %v", err)
	}
	if err := evaHub.RegisterAgent(evaSpec2, evaSpec2Handler); err != nil {
		log.Fatalf("Failed to register eva-specialist-2: %v", err)
	}
	if err := evaHub.RegisterAgent(commander, commanderHandler); err != nil {
		log.Fatalf("Failed to register mission-commander in eva-hub: %v", err)
	}

	// Register agents in ISS Hub
	if err := issHub.RegisterAgent(flightEng, flightEngHandler); err != nil {
		log.Fatalf("Failed to register flight-engineer: %v", err)
	}
	if err := issHub.RegisterAgent(commander, commanderHandler); err != nil {
		log.Fatalf("Failed to register mission-commander in iss-hub: %v", err)
	}

	fmt.Printf("  ✓ Registered all agents with hubs\n")
	fmt.Println()

	// ============================================================================
	// 5. Subscribe Agents to Topics
	// ============================================================================
	fmt.Println("4. Subscribing agents to topics...")

	evaHub.Subscribe(evaSpec1.ID(), "equipment")
	fmt.Printf("  ✓ eva-specialist-1 subscribed to 'equipment'\n")

	evaHub.Subscribe(evaSpec2.ID(), "safety")
	fmt.Printf("  ✓ eva-specialist-2 subscribed to 'safety'\n")

	evaHub.Subscribe(commander.ID(), "equipment")
	evaHub.Subscribe(commander.ID(), "safety")
	fmt.Printf("  ✓ mission-commander subscribed to 'equipment' and 'safety'\n")
	fmt.Println()

	// ============================================================================
	// 6. Agent-to-Agent Communication
	// ============================================================================
	fmt.Println("5. Agent-to-Agent Communication")
	fmt.Println("   eva-specialist-1 → eva-specialist-2")
	fmt.Println("   Message: I need the torque wrench, can you retrieve it from the tool bag?")

	evaHub.Send(ctx, evaSpec1.ID(), evaSpec2.ID(), "I need the torque wrench, can you retrieve it from the tool bag?")

	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// ============================================================================
	// 7. Broadcast Communication
	// ============================================================================
	fmt.Println("6. Broadcast Communication")
	fmt.Println("   mission-commander → all EVA crew")
	fmt.Println("   Message: Orbital sunset in 20 minutes, prioritize the cooling line connection")

	evaHub.Broadcast(ctx, commander.ID(), "Orbital sunset in 20 minutes, prioritize the cooling line connection")

	fmt.Printf("   %s\n", <-responses)
	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// ============================================================================
	// 8. Pub/Sub Communication
	// ============================================================================
	fmt.Println("7. Pub/Sub Communication")
	fmt.Println("   mission-commander publishes to topic 'equipment'")
	fmt.Println("   Message: Spare thermal blanket available in airlock if needed")

	evaHub.Publish(ctx, commander.ID(), "equipment", "Spare thermal blanket available in airlock if needed")

	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// ============================================================================
	// 9. Cross-Hub Communication
	// ============================================================================
	fmt.Println("8. Cross-Hub Communication")
	fmt.Println("   eva-specialist-1 → mission-commander (eva-hub) → flight-engineer (iss-hub)")
	fmt.Println("   Message: Cooling line connection complete, ready to pressurize system")

	evaHub.Send(ctx, evaSpec1.ID(), commander.ID(), "Cooling line connection complete, ready to pressurize system")
	fmt.Printf("   %s\n", <-responses)

	issHub.Send(ctx, commander.ID(), flightEng.ID(), "EVA crew ready for cooling system pressurization")
	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	// ============================================================================
	// 10. Display Metrics
	// ============================================================================
	fmt.Println("9. EVA Operation Metrics")

	evaMetrics := evaHub.Metrics()
	fmt.Printf("   EVA Hub:\n")
	fmt.Printf("     - Local Agents: %d\n", evaMetrics.LocalAgents)
	fmt.Printf("     - Messages Sent: %d\n", evaMetrics.MessagesSent)
	fmt.Printf("     - Messages Received: %d\n", evaMetrics.MessagesRecv)

	issMetrics := issHub.Metrics()
	fmt.Printf("   ISS Hub:\n")
	fmt.Printf("     - Local Agents: %d\n", issMetrics.LocalAgents)
	fmt.Printf("     - Messages Sent: %d\n", issMetrics.MessagesSent)
	fmt.Printf("     - Messages Received: %d\n", issMetrics.MessagesRecv)
	fmt.Println()

	fmt.Println("=== EVA Operation Complete ===")
}
