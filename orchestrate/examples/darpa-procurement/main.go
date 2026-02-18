package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/tailored-agentic-units/kernel/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
)

func main() {
	config, err := ParseConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if config.Verbose {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		logger := slog.New(handler)
		observer := observability.NewSlogObserver(logger)
		observability.RegisterObserver("slog", observer)
	}

	fmt.Println("DARPA Research Procurement Simulation")

	maxTokensStr := "default"
	if config.MaxTokens > 0 {
		maxTokensStr = fmt.Sprintf("%d", config.MaxTokens)
	}
	fmt.Printf("Initializing agents (config: %s, max_tokens: %s)...\n\n", config.AgentConfig, maxTokensStr)

	ResetProjects()

	registry, err := InitializeAgents(config)
	if err != nil {
		log.Fatalf("Failed to initialize agents: %v", err)
	}

	ctx := context.Background()
	startTime := time.Now()

	var approved, rejected, revised int
	var totalCost int
	approvedIDs := []string{}

	for i := 0; i < config.Requests; i++ {
		fmt.Printf("=== Processing Request %d/%d ===\n", i+1, config.Requests)

		requestID := fmt.Sprintf("PR-2024-%03d", i+1)

		graph, err := BuildWorkflow(config, registry)
		if err != nil {
			log.Fatalf("Failed to build workflow: %v", err)
		}

		initialState := state.New(nil)
		initialState = initialState.Set("request_id", requestID)

		var finalState state.State
		var execErr error

		if config.FailAt != FailureNone && i == 0 {
			finalState, execErr = executeWithFailure(ctx, graph, initialState, config.FailAt, config)
		} else {
			finalState, execErr = graph.Execute(ctx, initialState)
		}

		if execErr != nil {
			fmt.Printf("✗ Workflow failed: %v\n\n", execErr)
			rejected++
			continue
		}

		decision, _ := finalState.Get("decision")
		projectName, _ := finalState.Get("project_name")
		classification, _ := finalState.Get("classification")
		componentCount, _ := finalState.Get("component_count")
		estimatedCost, _ := finalState.Get("estimated_cost")
		cost := estimatedCost.(int)

		fmt.Printf("\nR&D Project: %s\n", projectName)
		fmt.Printf("  Classification: %s\n", classification)
		fmt.Printf("  Components: %d\n", componentCount)
		fmt.Printf("  Estimated Cost: $%s\n", formatCost(cost))

		if riskLevel, ok := finalState.Get("risk_level"); ok {
			fmt.Printf("  Risk Level: %s\n", riskLevel)
		}

		if legalStatus, ok := finalState.Get("legal_status"); ok {
			fmt.Printf("  Legal Review: %s\n", legalStatus)
		}

		if securityStatus, ok := finalState.Get("security_status"); ok {
			fmt.Printf("  Security Review: %s\n", securityStatus)
		}

		iterations, _ := finalState.Get("iterations")
		iter := iterations.(int)

		fmt.Printf("\nFinal Decision:\n")
		switch decision {
		case "APPROVED":
			approved++
			totalCost += cost
			approvedIDs = append(approvedIDs, requestID)
			if approvalLevel, ok := finalState.Get("approval_level"); ok {
				fmt.Printf("  ✓ APPROVED by %s\n", approvalLevel)
			}
			fmt.Printf("  Award ID: %s\n", requestID)
		case "REJECTED":
			rejected++
			fmt.Printf("  ✗ REJECTED\n")
		case "NEEDS REVISION":
			revised++
			if iter >= 2 {
				fmt.Printf("  ✗ REJECTED (exceeded revision limit of 2)\n")
				rejected++
			} else {
				fmt.Printf("  ↻ NEEDS REVISION (iteration %d/2)\n", iter)
			}
		}

		fmt.Println()
	}

	duration := time.Since(startTime)
	avgTime := duration.Seconds() / float64(config.Requests)

	fmt.Println("Summary:")
	fmt.Printf("- Requests processed: %d\n", config.Requests)
	fmt.Printf("- Approved: %d", approved)
	if len(approvedIDs) > 0 {
		fmt.Printf(" (%s", approvedIDs[0])
		for i := 1; i < len(approvedIDs); i++ {
			fmt.Printf(", %s", approvedIDs[i])
		}
		fmt.Printf(")")
	}
	fmt.Println()
	fmt.Printf("- Rejected: %d\n", rejected)
	if revised > 0 {
		fmt.Printf("- Required revision: %d\n", revised)
		fmt.Printf("- Revision rate: %.0f%% (%d/%d required revision)\n",
			float64(revised)/float64(config.Requests)*100, revised, config.Requests)
	}
	if approved > 0 {
		fmt.Printf("- Total budget allocated: $%s\n", formatCost(totalCost))
	}
	fmt.Printf("- Total processing time: %.1fs\n", duration.Seconds())
	fmt.Printf("- Average time per request: %.1fs\n", avgTime)
}

func executeWithFailure(ctx context.Context, graph state.StateGraph, initialState state.State, failStage FailureStage, config *WorkflowConfig) (state.State, error) {
	fmt.Printf("NOTE: Failure injection enabled at stage: %s\n\n", failStage)

	runID := initialState.RunID

	failedState, err := graph.Execute(ctx, initialState)

	if err == nil {
		return state.State{}, fmt.Errorf("expected failure at %s stage but workflow completed successfully", failStage)
	}

	fmt.Printf("\n✗ SIMULATED FAILURE at %s stage\n", failStage)
	checkpointNode := failedState.CheckpointNode
	fmt.Printf("Checkpoint saved: %s (runID: %s)\n", checkpointNode, runID)

	fmt.Println("\n=== Resuming from Checkpoint ===")
	fmt.Printf("RunID: %s\n", runID)
	fmt.Printf("Checkpoint: %s\n", checkpointNode)
	fmt.Println()

	config.FailAt = FailureNone

	resumedState, err := graph.Resume(ctx, runID)
	if err != nil {
		return state.State{}, fmt.Errorf("resume failed: %w", err)
	}

	fmt.Println("\nRecovery Statistics:")
	fmt.Println("- Checkpoint recovery successful")
	fmt.Println("- State preserved across failure")

	return resumedState, nil
}

func formatCost(cost int) string {
	if cost >= 1000 {
		return fmt.Sprintf("%d,%03d", cost/1000, cost%1000)
	}
	return fmt.Sprintf("%d", cost)
}
