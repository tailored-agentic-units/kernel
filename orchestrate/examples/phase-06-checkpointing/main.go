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
	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

var (
	firstExecutionFailed = false
	analysisAttempts     = 0
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Multi-Stage Data Analysis with Checkpoint Recovery ===")
	fmt.Println()

	slogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogLogger := slog.New(slogHandler)
	slogObserver := observability.NewSlogObserver(slogLogger)
	observability.RegisterObserver("slog", slogObserver)

	fmt.Printf("1. Configuring observability...\n")
	fmt.Printf("  ✓ Registered slog observer\n")
	fmt.Println()

	fmt.Println("2. Loading agent configuration...")

	llamaConfig, err := agentconfig.LoadAgentConfig("examples/phase-06-checkpointing/config.llama.json")
	if err != nil {
		log.Fatalf("Failed to load llama config: %v", err)
	}

	llamaConfig.Name = "data-analyst"
	llamaConfig.SystemPrompt = `You are a scientific data analyst processing research data through multiple stages.
You provide concise summaries of each processing stage.
Keep responses to 1-2 sentences focusing on key findings or actions.`

	dataAgent, err := agent.New(llamaConfig)
	if err != nil {
		log.Fatalf("Failed to create data agent: %v", err)
	}

	fmt.Printf("  ✓ Created data-analyst agent (llama3.2:3b)\n")
	fmt.Println()

	fmt.Println("3. Creating data analysis pipeline with checkpointing...")

	graphConfig := config.DefaultGraphConfig("data-pipeline")
	graphConfig.Observer = "slog"
	graphConfig.MaxIterations = 10
	graphConfig.Checkpoint.Store = "memory"
	graphConfig.Checkpoint.Interval = 1
	graphConfig.Checkpoint.Preserve = true

	graph, err := state.NewGraph(graphConfig)
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}

	fmt.Printf("  ✓ Created state graph with checkpointing enabled\n")
	fmt.Printf("     - Checkpoint interval: Every 1 node\n")
	fmt.Printf("     - Checkpoint store: memory\n")
	fmt.Printf("     - Preserve on success: true\n")
	fmt.Println()

	fmt.Println("4. Defining pipeline stages...")

	ingestNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → STAGE 1: Data Ingestion")
		fmt.Println("     Loading research dataset...")
		time.Sleep(1 * time.Second)

		datasetName, _ := s.Get("dataset")
		prompt := fmt.Sprintf("Describe the key characteristics of the '%s' dataset being ingested.", datasetName)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := dataAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("ingestion failed: %w", err)
		}

		characteristics := response.Content()
		fmt.Printf("     Characteristics: %s\n", characteristics)
		fmt.Printf("     ✓ Stage 1 complete\n")

		return s.Set("characteristics", characteristics).Set("stage", "ingested"), nil
	})

	preprocessNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → STAGE 2: Preprocessing")
		fmt.Println("     Cleaning and normalizing data...")
		time.Sleep(1 * time.Second)

		characteristics, _ := s.Get("characteristics")
		prompt := fmt.Sprintf("What preprocessing steps are needed for data with these characteristics: %s", characteristics)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := dataAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("preprocessing failed: %w", err)
		}

		preprocessSteps := response.Content()
		fmt.Printf("     Steps: %s\n", preprocessSteps)
		fmt.Printf("     ✓ Stage 2 complete\n")

		return s.Set("preprocessing", preprocessSteps).Set("stage", "preprocessed"), nil
	})

	analyzeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → STAGE 3: Analysis")
		fmt.Println("     Running statistical analysis...")

		analysisAttempts++

		if !firstExecutionFailed && analysisAttempts == 1 {
			firstExecutionFailed = true
			fmt.Println("     ✗ SIMULATED FAILURE: Analysis process interrupted")
			return s, fmt.Errorf("analysis interrupted: simulated system failure")
		}

		time.Sleep(1 * time.Second)

		datasetName, _ := s.Get("dataset")
		prompt := fmt.Sprintf("What statistical insights can be derived from analyzing the '%s' dataset?", datasetName)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := dataAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("analysis failed: %w", err)
		}

		insights := response.Content()
		fmt.Printf("     Insights: %s\n", insights)
		fmt.Printf("     ✓ Stage 3 complete\n")

		return s.Set("insights", insights).Set("stage", "analyzed"), nil
	})

	reportNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Println("\n  → STAGE 4: Report Generation")
		fmt.Println("     Generating final report...")
		time.Sleep(1 * time.Second)

		insights, _ := s.Get("insights")
		prompt := fmt.Sprintf("Summarize these key findings in a report conclusion: %s", insights)

		messages := protocol.InitMessages(protocol.RoleUser, prompt)

		response, err := dataAgent.Chat(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("report generation failed: %w", err)
		}

		reportSummary := response.Content()
		fmt.Printf("     Summary: %s\n", reportSummary)
		fmt.Printf("     ✓ Stage 4 complete\n")

		return s.Set("report", reportSummary).Set("stage", "completed"), nil
	})

	if err := graph.AddNode("ingest", ingestNode); err != nil {
		log.Fatalf("Failed to add ingest node: %v", err)
	}
	if err := graph.AddNode("preprocess", preprocessNode); err != nil {
		log.Fatalf("Failed to add preprocess node: %v", err)
	}
	if err := graph.AddNode("analyze", analyzeNode); err != nil {
		log.Fatalf("Failed to add analyze node: %v", err)
	}
	if err := graph.AddNode("report", reportNode); err != nil {
		log.Fatalf("Failed to add report node: %v", err)
	}

	fmt.Printf("  ✓ Defined 4 pipeline stages\n")
	fmt.Printf("     - ingest → preprocess → analyze → report\n")
	fmt.Println()

	fmt.Println("5. Building pipeline graph...")

	if err := graph.AddEdge("ingest", "preprocess", nil); err != nil {
		log.Fatalf("Failed to add edge: %v", err)
	}
	if err := graph.AddEdge("preprocess", "analyze", nil); err != nil {
		log.Fatalf("Failed to add edge: %v", err)
	}
	if err := graph.AddEdge("analyze", "report", nil); err != nil {
		log.Fatalf("Failed to add edge: %v", err)
	}

	if err := graph.SetEntryPoint("ingest"); err != nil {
		log.Fatalf("Failed to set entry point: %v", err)
	}
	if err := graph.SetExitPoint("report"); err != nil {
		log.Fatalf("Failed to set exit point: %v", err)
	}

	fmt.Printf("  ✓ Pipeline graph constructed\n")
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("EXECUTION 1: Initial Run (Will Fail)")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println()

	observer := observability.NoOpObserver{}
	initialState := state.New(observer)
	initialState = initialState.Set("dataset", "climate-research-2024")

	runID := initialState.RunID
	fmt.Printf("Pipeline RunID: %s\n", runID)

	startTime := time.Now()
	finalState, err := graph.Execute(ctx, initialState)
	executionTime := time.Since(startTime)

	fmt.Println()
	if err != nil {
		fmt.Printf("❌ EXECUTION FAILED after %.2fs\n", executionTime.Seconds())
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   Checkpoint saved at: %s\n", finalState.CheckpointNode)
		fmt.Println()
	} else {
		fmt.Printf("✓ Execution completed in %.2fs\n", executionTime.Seconds())
		fmt.Println()
	}

	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("EXECUTION 2: Resume from Checkpoint")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println()

	fmt.Printf("Resuming pipeline from RunID: %s\n", runID)
	fmt.Printf("Last completed stage: %s\n", finalState.CheckpointNode)
	fmt.Println()

	fmt.Println("Note: Stages 1-2 will be skipped (already completed)")
	fmt.Println("      Execution resumes from Stage 3")
	fmt.Println()

	time.Sleep(2 * time.Second)

	resumeStartTime := time.Now()
	resumedState, err := graph.Resume(ctx, runID)
	resumeTime := time.Since(resumeStartTime)

	fmt.Println()
	if err != nil {
		log.Fatalf("❌ RESUME FAILED: %v", err)
	}

	fmt.Printf("✓ Pipeline completed successfully after resume!\n")
	fmt.Printf("   Resume execution time: %.2fs\n", resumeTime.Seconds())
	fmt.Printf("   Total time (initial + resume): %.2fs\n", (executionTime + resumeTime).Seconds())
	fmt.Printf("   Time saved by checkpointing: ~2-3s (skipped stages 1-2)\n")
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("FINAL RESULTS")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println()

	if report, exists := resumedState.Get("report"); exists {
		fmt.Printf("Report Summary:\n%s\n", report)
		fmt.Println()
	}

	if insights, exists := resumedState.Get("insights"); exists {
		fmt.Printf("Key Insights:\n%s\n", insights)
		fmt.Println()
	}

	fmt.Println("Checkpoint Demonstration Summary:")
	fmt.Println("  ✓ Initial execution failed at Stage 3")
	fmt.Println("  ✓ Checkpoint preserved progress through Stage 2")
	fmt.Println("  ✓ Resume skipped completed stages (1-2)")
	fmt.Println("  ✓ Execution continued from Stage 3")
	fmt.Println("  ✓ Pipeline completed successfully")
	fmt.Println("  ✓ Time and cost savings demonstrated")
	fmt.Println()

	fmt.Println("This example demonstrates Phase 6 checkpointing capabilities:")
	fmt.Println("  - Checkpoint save at configurable intervals")
	fmt.Println("  - State persistence across execution failures")
	fmt.Println("  - Resume execution from saved checkpoints")
	fmt.Println("  - Progress preservation (skipping completed work)")
	fmt.Println("  - Observer integration (checkpoint events)")
	fmt.Println("  - Production fault tolerance patterns")
}
