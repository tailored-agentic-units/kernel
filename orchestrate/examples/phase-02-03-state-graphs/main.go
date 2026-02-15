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
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Software Deployment Pipeline - State Graph Example ===")
	fmt.Println()

	// ============================================================================
	// 1. Configure Observer
	// ============================================================================
	fmt.Println("1. Configuring observability...")

	slogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogLogger := slog.New(slogHandler)
	slogObserver := observability.NewSlogObserver(slogLogger)
	observability.RegisterObserver("slog", slogObserver)

	fmt.Printf("  ✓ Registered slog observer\n")
	fmt.Println()

	// ============================================================================
	// 2. Load Agent Configuration
	// ============================================================================
	fmt.Println("2. Loading agent configuration...")

	llamaConfig, err := agentconfig.LoadAgentConfig("examples/phase-02-03-state-graphs/config.llama.json")
	if err != nil {
		log.Fatalf("Failed to load llama config: %v", err)
	}

	llamaConfig.Name = "deployment-manager"
	llamaConfig.SystemPrompt = `You are an expert DevOps deployment manager responsible for software deployments.
You analyze deployment requests and provide technical assessments.
Your responses should be concise and focus on technical details.
Always respond in 1-2 sentences with specific technical information.`

	deploymentAgent, err := agent.New(llamaConfig)
	if err != nil {
		log.Fatalf("Failed to create deployment agent: %v", err)
	}

	fmt.Printf("  ✓ Created deployment-manager agent (llama3.2:3b)\n")
	fmt.Println()

	// ============================================================================
	// 3. Create State Graph
	// ============================================================================
	fmt.Println("3. Creating deployment pipeline state graph...")

	graphConfig := config.DefaultGraphConfig("deployment-pipeline")
	graphConfig.Observer = "slog"
	graphConfig.MaxIterations = 10

	graph, err := state.NewGraph(graphConfig)
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}

	fmt.Printf("  ✓ Created state graph with observer\n")
	fmt.Println()

	// ============================================================================
	// 4. Define Pipeline Nodes
	// ============================================================================
	fmt.Println("4. Defining pipeline nodes...")

	planNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → PLAN: Analyzing deployment requirements...")

		appName, _ := s.Get("app_name")
		targetEnv, _ := s.Get("target_env")

		prompt := fmt.Sprintf("Analyze deployment plan for application '%s' to '%s' environment. What are the key considerations?", appName, targetEnv)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := deploymentAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("plan failed: %w", err)
		}

		planDetails := response.Content()
		fmt.Printf("     Plan: %s\n", planDetails)

		return s.Set("plan", planDetails).Set("status", "planned"), nil
	})

	buildNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → BUILD: Compiling and creating artifacts...")

		appName, _ := s.Get("app_name")

		prompt := fmt.Sprintf("What artifacts should be built for '%s' application deployment? List 2-3 key artifacts.", appName)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := deploymentAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("build failed: %w", err)
		}

		artifacts := response.Content()
		fmt.Printf("     Artifacts: %s\n", artifacts)

		return s.Set("artifacts", artifacts).Set("status", "built"), nil
	})

	testNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → TEST: Running automated test suite...")

		retryCount, exists := s.Get("retry_count")
		if !exists {
			retryCount = 0
		}

		attempts := retryCount.(int)

		prompt := fmt.Sprintf("Evaluate test results for deployment (attempt %d). Should tests pass (yes) or need fixes (no)?", attempts+1)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := deploymentAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("test execution failed: %w", err)
		}

		testResult := response.Content()
		fmt.Printf("     Test Result: %s\n", testResult)

		return s.Set("test_result", testResult).Set("status", "tested"), nil
	})

	fixNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → FIX: Addressing test failures...")

		retryCount, exists := s.Get("retry_count")
		if !exists {
			retryCount = 0
		}

		attempts := retryCount.(int) + 1

		testResult, _ := s.Get("test_result")
		prompt := fmt.Sprintf("Test failed: %s. What fix should be applied (attempt %d)?", testResult, attempts)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := deploymentAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("fix failed: %w", err)
		}

		fixDetails := response.Content()
		fmt.Printf("     Fix Applied: %s\n", fixDetails)

		return s.Set("fix_details", fixDetails).Set("retry_count", attempts).Set("status", "fixed"), nil
	})

	deployNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → DEPLOY: Deploying to target environment...")

		targetEnv, _ := s.Get("target_env")
		artifacts, _ := s.Get("artifacts")

		prompt := fmt.Sprintf("Confirm deployment to '%s' with artifacts: %s. Provide deployment confirmation.", targetEnv, artifacts)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := deploymentAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("deployment failed: %w", err)
		}

		deploymentConfirm := response.Content()
		fmt.Printf("     Deployment: %s\n", deploymentConfirm)

		return s.Set("deployment_result", deploymentConfirm).Set("status", "deployed"), nil
	})

	rollbackNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → ROLLBACK: Maximum retry attempts exceeded, rolling back...")

		retryCount, _ := s.Get("retry_count")

		prompt := fmt.Sprintf("Deployment failed after %d attempts. Describe rollback procedure.", retryCount)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := deploymentAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("rollback failed: %w", err)
		}

		rollbackDetails := response.Content()
		fmt.Printf("     Rollback: %s\n", rollbackDetails)

		return s.Set("rollback_details", rollbackDetails).Set("status", "rolled_back"), nil
	})

	graph.AddNode("plan", planNode)
	graph.AddNode("build", buildNode)
	graph.AddNode("test", testNode)
	graph.AddNode("fix", fixNode)
	graph.AddNode("deploy", deployNode)
	graph.AddNode("rollback", rollbackNode)

	fmt.Printf("  ✓ Added 6 nodes (plan, build, test, fix, deploy, rollback)\n")
	fmt.Println()

	// ============================================================================
	// 5. Define Pipeline Edges
	// ============================================================================
	fmt.Println("5. Defining pipeline transitions...")

	graph.AddEdge("plan", "build", state.AlwaysTransition())
	graph.AddEdge("build", "test", state.AlwaysTransition())

	testsPassed := func(s state.State) bool {
		result, exists := s.Get("test_result")
		if !exists {
			return false
		}
		testStr := fmt.Sprintf("%v", result)
		return len(testStr) > 0 && (testStr[0] == 'y' || testStr[0] == 'Y' || testStr[0] == 'P' || testStr[0] == 'p')
	}

	testsFailedWithRetriesLeft := func(s state.State) bool {
		retryCount, exists := s.Get("retry_count")
		if !exists {
			retryCount = 0
		}
		return retryCount.(int) < 3 && !testsPassed(s)
	}

	maxRetriesExceeded := func(s state.State) bool {
		retryCount, exists := s.Get("retry_count")
		if !exists {
			return false
		}
		return retryCount.(int) >= 3
	}

	graph.AddEdge("test", "deploy", testsPassed)
	graph.AddEdge("test", "fix", testsFailedWithRetriesLeft)
	graph.AddEdge("test", "rollback", maxRetriesExceeded)
	graph.AddEdge("fix", "test", state.AlwaysTransition())

	fmt.Printf("  ✓ Added 6 edges with conditional routing\n")
	fmt.Println()

	// ============================================================================
	// 6. Configure Entry and Exit Points
	// ============================================================================
	fmt.Println("6. Configuring entry and exit points...")

	graph.SetEntryPoint("plan")
	graph.SetExitPoint("deploy")
	graph.SetExitPoint("rollback")

	fmt.Printf("  ✓ Entry point: plan\n")
	fmt.Printf("  ✓ Exit points: deploy, rollback\n")
	fmt.Println()

	// ============================================================================
	// 7. Execute Deployment Pipeline
	// ============================================================================
	fmt.Println("7. Executing deployment pipeline...")
	fmt.Println()

	initialState := state.New(slogObserver)
	initialState = initialState.Set("app_name", "cloud-api-service")
	initialState = initialState.Set("target_env", "production")
	initialState = initialState.Set("retry_count", 0)

	fmt.Println("  Initial deployment request:")
	fmt.Printf("    Application: cloud-api-service\n")
	fmt.Printf("    Environment: production\n")
	fmt.Println()

	startTime := time.Now()

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}

	duration := time.Since(startTime)

	fmt.Println()
	fmt.Println("  ✓ Pipeline execution completed")
	fmt.Println()

	// ============================================================================
	// 8. Display Results
	// ============================================================================
	fmt.Println("8. Deployment Results")
	fmt.Println()

	status, _ := finalState.Get("status")
	fmt.Printf("   Final Status: %s\n", status)
	fmt.Println()

	if status == "deployed" {
		fmt.Println("   ✓ DEPLOYMENT SUCCESSFUL")
		deploymentResult, _ := finalState.Get("deployment_result")
		fmt.Printf("   Details: %s\n", deploymentResult)
	} else if status == "rolled_back" {
		fmt.Println("   ✗ DEPLOYMENT FAILED - ROLLED BACK")
		rollbackDetails, _ := finalState.Get("rollback_details")
		fmt.Printf("   Details: %s\n", rollbackDetails)
		retryCount, _ := finalState.Get("retry_count")
		fmt.Printf("   Retry Attempts: %d\n", retryCount)
	}
	fmt.Println()

	// ============================================================================
	// 9. Execution Metrics
	// ============================================================================
	fmt.Println("9. Execution Metrics")
	fmt.Printf("   Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("   Max Iterations Allowed: %d\n", graphConfig.MaxIterations)
	fmt.Println()

	fmt.Println("=== Deployment Pipeline Complete ===")
}
