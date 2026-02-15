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
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
)

type PaperSection struct {
	Name    string
	Content string
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Research Paper Analysis Pipeline - Sequential Chains Example ===")
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

	llamaConfig, err := agentconfig.LoadAgentConfig("examples/phase-04-sequential-chains/config.llama.json")
	if err != nil {
		log.Fatalf("Failed to load llama config: %v", err)
	}

	llamaConfig.Name = "research-analyst"
	llamaConfig.SystemPrompt = `You are an expert research paper analyst.
You analyze academic papers and extract key information.
Your responses should be concise and focus on the most important points.
Always respond in 1-2 sentences with specific details.`

	analysisAgent, err := agent.New(llamaConfig)
	if err != nil {
		log.Fatalf("Failed to create analysis agent: %v", err)
	}

	fmt.Printf("  ✓ Created research-analyst agent (llama3.2:3b)\n")
	fmt.Println()

	// ============================================================================
	// 3. Prepare Research Paper Sections
	// ============================================================================
	fmt.Println("3. Preparing research paper sections...")

	sections := []PaperSection{
		{
			Name: "Abstract",
			Content: `This paper presents a novel approach to distributed consensus in blockchain networks
using adaptive sharding techniques. Our method improves transaction throughput by 3x while
maintaining security guarantees. Experimental results show significant improvements over
existing protocols in both latency and scalability.`,
		},
		{
			Name: "Introduction",
			Content: `Current blockchain systems face scalability challenges as transaction volume grows.
Traditional consensus mechanisms like Proof-of-Work and Proof-of-Stake struggle to maintain
high throughput without compromising decentralization. This paper addresses these limitations
through dynamic shard allocation based on network conditions.`,
		},
		{
			Name: "Methodology",
			Content: `We implemented an adaptive sharding protocol that monitors network load and adjusts
shard configuration in real-time. The protocol uses a reputation-based validator selection
mechanism and employs cross-shard transaction routing with optimistic execution. We tested
the system with up to 10,000 nodes across five geographic regions.`,
		},
		{
			Name: "Results",
			Content: `Our experiments show 3.2x improvement in transaction throughput compared to baseline
systems, with average latency reduced from 12 seconds to 4 seconds. The system maintained
99.9% uptime during a 30-day test period and successfully handled peak loads of 50,000
transactions per second. Cross-shard transaction overhead was minimal at 8%.`,
		},
		{
			Name: "Conclusion",
			Content: `Adaptive sharding provides a practical solution to blockchain scalability challenges.
The approach maintains security while significantly improving performance. Future work will
explore integration with zero-knowledge proofs and investigate behavior under adversarial
conditions. The protocol is ready for testnet deployment.`,
		},
	}

	fmt.Printf("  ✓ Loaded %d paper sections\n", len(sections))
	fmt.Println()

	// ============================================================================
	// 4. Configure Sequential Chain
	// ============================================================================
	fmt.Println("4. Configuring sequential analysis chain...")

	chainConfig := config.DefaultChainConfig()
	chainConfig.Observer = "slog"
	chainConfig.CaptureIntermediateStates = true

	fmt.Printf("  ✓ Chain configuration ready\n")
	fmt.Printf("    Intermediate state capture: enabled\n")
	fmt.Println()

	// ============================================================================
	// 5. Define Analysis Step Processor
	// ============================================================================
	fmt.Println("5. Defining analysis step processor...")

	stepProcessor := func(ctx context.Context, section PaperSection, s state.State) (state.State, error) {
		sectionName := section.Name

		var prompt string
		var stateKey string

		switch sectionName {
		case "Abstract":
			prompt = fmt.Sprintf("Extract the main research contribution from this abstract: %s", section.Content)
			stateKey = "main_contribution"

		case "Introduction":
			prompt = fmt.Sprintf("Identify the key problem being addressed: %s", section.Content)
			stateKey = "problem_statement"

		case "Methodology":
			prompt = fmt.Sprintf("Summarize the research method in one sentence: %s", section.Content)
			stateKey = "methodology"

		case "Results":
			prompt = fmt.Sprintf("List the top 2 quantitative results: %s", section.Content)
			stateKey = "key_results"

		case "Conclusion":
			prompt = fmt.Sprintf("What is the main future work direction mentioned: %s", section.Content)
			stateKey = "future_work"

		default:
			return s, fmt.Errorf("unknown section: %s", sectionName)
		}

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := analysisAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("analysis failed for %s: %w", sectionName, err)
		}

		analysis := response.Content()

		return s.Set(stateKey, analysis), nil
	}

	fmt.Printf("  ✓ Step processor defined\n")
	fmt.Println()

	// ============================================================================
	// 6. Define Progress Callback
	// ============================================================================
	fmt.Println("6. Configuring progress tracking...")

	totalSteps := len(sections)

	progressCallback := func(completed int, total int, s state.State) {
		percentage := (completed * 100) / total
		fmt.Printf("\n  Progress: Step %d/%d complete (%d%%)\n", completed, total, percentage)
	}

	fmt.Printf("  ✓ Progress callback configured\n")
	fmt.Println()

	// ============================================================================
	// 7. Execute Sequential Analysis Chain
	// ============================================================================
	fmt.Println("7. Executing sequential analysis pipeline...")
	fmt.Println()

	initialState := state.New(slogObserver)
	initialState = initialState.Set("paper_title", "Adaptive Sharding for Blockchain Scalability")
	initialState = initialState.Set("analysis_start", time.Now().Format(time.RFC3339))

	fmt.Println("  Starting analysis of 5 paper sections...")
	fmt.Println()

	startTime := time.Now()

	result, err := workflows.ProcessChain(
		ctx,
		chainConfig,
		sections,
		initialState,
		stepProcessor,
		progressCallback,
	)
	if err != nil {
		log.Fatalf("Analysis pipeline failed: %v", err)
	}

	duration := time.Since(startTime)

	fmt.Println()
	fmt.Println("  ✓ Analysis pipeline completed")
	fmt.Println()

	// ============================================================================
	// 8. Display Analysis Results
	// ============================================================================
	fmt.Println("8. Analysis Results")
	fmt.Println()

	paperTitle, _ := result.Final.Get("paper_title")
	fmt.Printf("   Paper: %s\n", paperTitle)
	fmt.Println()

	fmt.Println("   Key Findings:")
	fmt.Println()

	contribution, _ := result.Final.Get("main_contribution")
	fmt.Printf("   Main Contribution:\n     %s\n\n", contribution)

	problem, _ := result.Final.Get("problem_statement")
	fmt.Printf("   Problem Statement:\n     %s\n\n", problem)

	methodology, _ := result.Final.Get("methodology")
	fmt.Printf("   Methodology:\n     %s\n\n", methodology)

	results, _ := result.Final.Get("key_results")
	fmt.Printf("   Key Results:\n     %s\n\n", results)

	futureWork, _ := result.Final.Get("future_work")
	fmt.Printf("   Future Work:\n     %s\n\n", futureWork)

	// ============================================================================
	// 9. Display State Evolution
	// ============================================================================
	fmt.Println("9. State Evolution Analysis")
	fmt.Println()

	if len(result.Intermediate) > 0 {
		fmt.Printf("   Total states captured: %d (initial + %d processing steps)\n", len(result.Intermediate), result.Steps)
		fmt.Println()

		fmt.Println("   State progression:")
		sectionNames := []string{"Abstract", "Introduction", "Methodology", "Results", "Conclusion"}
		for i := range result.Intermediate {
			if i == 0 {
				fmt.Printf("     [%d] Initial state (paper metadata)\n", i)
			} else {
				fmt.Printf("     [%d] After processing: %s\n", i, sectionNames[i-1])
			}
		}
		fmt.Println()
	}

	// ============================================================================
	// 10. Execution Metrics
	// ============================================================================
	fmt.Println("10. Execution Metrics")
	fmt.Printf("    Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("    Steps Completed: %d/%d\n", result.Steps, totalSteps)
	fmt.Printf("    Intermediate States Captured: %d\n", len(result.Intermediate))
	fmt.Printf("    Average Time per Step: %v\n", (duration / time.Duration(result.Steps)).Round(time.Millisecond))
	fmt.Println()

	fmt.Println("=== Research Paper Analysis Complete ===")
}
