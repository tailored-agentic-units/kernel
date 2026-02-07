package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
	"github.com/tailored-agentic-units/kernel/agent"
	agentconfig "github.com/tailored-agentic-units/kernel/core/config"
)

type Document struct {
	ID      string
	Title   string
	Content string
	Version int
	Status  string
}

type Analysis struct {
	Analyst string
	Type    string
	Finding string
	Issues  []string
}

type Review struct {
	Reviewer string
	Approved bool
	Comments string
	Score    int
}

type Decision struct {
	Approved      bool
	Reason        string
	RecommendedChange string
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Technical Document Review Workflow ===")
	fmt.Println("Demonstrating: Chain → Parallel → Conditional routing with state management")
	fmt.Println()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("1. Loading agent configurations...")

	llamaConfig, err := agentconfig.LoadAgentConfig("examples/phase-07-conditional-routing/config.llama.json")
	if err != nil {
		log.Fatalf("Failed to load llama config: %v", err)
	}

	gemmaConfig, err := agentconfig.LoadAgentConfig("examples/phase-07-conditional-routing/config.gemma.json")
	if err != nil {
		log.Fatalf("Failed to load gemma config: %v", err)
	}

	techAnalystCfg := &agentconfig.AgentConfig{
		Name: "technical-analyst",
		SystemPrompt: `You are a technical analyst reviewing documentation.
Analyze technical accuracy, implementation details, and code examples.
Identify any technical errors, unclear explanations, or missing information.
Keep your analysis concise (2-3 sentences) and list specific issues.`,
		Client:   llamaConfig.Client,
		Provider: llamaConfig.Provider,
		Model:    llamaConfig.Model,
	}

	securityAnalystCfg := &agentconfig.AgentConfig{
		Name: "security-analyst",
		SystemPrompt: `You are a security analyst reviewing documentation.
Analyze security implications, vulnerability disclosures, and security best practices.
Identify any security concerns, missing warnings, or dangerous patterns.
Keep your analysis concise (2-3 sentences) and list specific issues.`,
		Client:   gemmaConfig.Client,
		Provider: gemmaConfig.Provider,
		Model:    gemmaConfig.Model,
	}

	businessAnalystCfg := &agentconfig.AgentConfig{
		Name: "business-analyst",
		SystemPrompt: `You are a business analyst reviewing documentation.
Analyze business value, user impact, and clarity for non-technical readers.
Identify any unclear business justification or missing user perspective.
Keep your analysis concise (2-3 sentences) and list specific issues.`,
		Client:   llamaConfig.Client,
		Provider: llamaConfig.Provider,
		Model:    llamaConfig.Model,
	}

	reviewer1Cfg := &agentconfig.AgentConfig{
		Name: "reviewer-alpha",
		SystemPrompt: `You are an experienced technical reviewer.
Review the document and prior analyses. Provide approval or rejection with justification.
Be thorough but fair. Respond in 2-3 sentences with clear approval/rejection.
Start response with "APPROVE:" or "REJECT:" followed by reasoning.`,
		Client:   gemmaConfig.Client,
		Provider: gemmaConfig.Provider,
		Model:    gemmaConfig.Model,
	}

	reviewer2Cfg := &agentconfig.AgentConfig{
		Name: "reviewer-beta",
		SystemPrompt: `You are a senior technical reviewer focused on quality.
Review the document and prior analyses. Provide approval or rejection with justification.
Focus on overall quality and completeness. Respond in 2-3 sentences with clear approval/rejection.
Start response with "APPROVE:" or "REJECT:" followed by reasoning.`,
		Client:   llamaConfig.Client,
		Provider: llamaConfig.Provider,
		Model:    llamaConfig.Model,
	}

	reviewer3Cfg := &agentconfig.AgentConfig{
		Name: "reviewer-gamma",
		SystemPrompt: `You are a principal engineer reviewing documentation.
Review the document and prior analyses. Provide approval or rejection with justification.
Focus on technical depth and accuracy. Respond in 2-3 sentences with clear approval/rejection.
Start response with "APPROVE:" or "REJECT:" followed by reasoning.`,
		Client:   gemmaConfig.Client,
		Provider: gemmaConfig.Provider,
		Model:    gemmaConfig.Model,
	}

	techAnalyst, err := agent.New(techAnalystCfg)
	if err != nil {
		log.Fatalf("Failed to create technical-analyst: %v", err)
	}

	secAnalyst, err := agent.New(securityAnalystCfg)
	if err != nil {
		log.Fatalf("Failed to create security-analyst: %v", err)
	}

	bizAnalyst, err := agent.New(businessAnalystCfg)
	if err != nil {
		log.Fatalf("Failed to create business-analyst: %v", err)
	}

	reviewer1, err := agent.New(reviewer1Cfg)
	if err != nil {
		log.Fatalf("Failed to create reviewer-alpha: %v", err)
	}

	reviewer2, err := agent.New(reviewer2Cfg)
	if err != nil {
		log.Fatalf("Failed to create reviewer-beta: %v", err)
	}

	reviewer3, err := agent.New(reviewer3Cfg)
	if err != nil {
		log.Fatalf("Failed to create reviewer-gamma: %v", err)
	}

	fmt.Println("   ✓ Agents created: 3 analysts + 3 reviewers")
	fmt.Println()

	fmt.Println("2. Configuring stateful workflow...")

	graphCfg := config.DefaultGraphConfig("document-review-workflow")
	graphCfg.Checkpoint = config.CheckpointConfig{
		Store:    "memory",
		Interval: 1,
		Preserve: false,
	}

	chainCfg := config.DefaultChainConfig()
	parallelCfg := config.DefaultParallelConfig()
	conditionalCfg := config.DefaultConditionalConfig()

	graph, err := state.NewGraph(graphCfg)
	if err != nil {
		log.Fatal(err)
	}

	document := Document{
		ID:      "DOC-2025-001",
		Title:   "API Authentication System Design",
		Content: "This document describes the implementation of JWT-based authentication with OAuth2 integration. The system provides secure token management, refresh token rotation, and supports multiple identity providers including Google, GitHub, and enterprise SAML.",
		Version: 1,
		Status:  "pending",
	}

	analysisAgents := []struct {
		name    string
		analyst agent.Agent
		atype   string
	}{
		{"technical-analyst", techAnalyst, "Technical"},
		{"security-analyst", secAnalyst, "Security"},
		{"business-analyst", bizAnalyst, "Business"},
	}

	analyzeProcessor := func(ctx context.Context, item struct {
		name    string
		analyst agent.Agent
		atype   string
	}, s state.State) (state.State, error) {
		doc, _ := s.Get("document")
		currentDoc := doc.(Document)

		logger.Info("sequential analysis", "analyst", item.name, "type", item.atype)

		prompt := fmt.Sprintf("Analyze this document:\n\nTitle: %s\n\nContent: %s\n\nProvide your %s analysis.",
			currentDoc.Title, currentDoc.Content, strings.ToLower(item.atype))

		response, err := item.analyst.Chat(ctx, prompt)
		if err != nil {
			return s, fmt.Errorf("analysis failed: %w", err)
		}

		content := response.Content()
		issues := []string{}
		if strings.Contains(strings.ToLower(content), "issue") ||
			strings.Contains(strings.ToLower(content), "concern") ||
			strings.Contains(strings.ToLower(content), "missing") {
			issues = append(issues, "flagged by analyst")
		}

		analysis := Analysis{
			Analyst: item.name,
			Type:    item.atype,
			Finding: content,
			Issues:  issues,
		}

		analysesKey := "analyses"
		var analyses []Analysis
		if existing, ok := s.Get(analysesKey); ok {
			analyses = existing.([]Analysis)
		}
		analyses = append(analyses, analysis)

		return s.Set(analysesKey, analyses), nil
	}

	analyzeNode := workflows.ChainNode(
		chainCfg,
		analysisAgents,
		analyzeProcessor,
		nil,
	)

	reviewAgents := []struct {
		name     string
		reviewer agent.Agent
	}{
		{"reviewer-alpha", reviewer1},
		{"reviewer-beta", reviewer2},
		{"reviewer-gamma", reviewer3},
	}

	reviewProcessor := func(ctx context.Context, item struct {
		name     string
		reviewer agent.Agent
	}) (Review, error) {
		logger.Info("concurrent review", "reviewer", item.name)

		prompt := "Review this document for approval. Consider prior analyses and provide clear APPROVE or REJECT decision with reasoning."

		response, err := item.reviewer.Chat(ctx, prompt)
		if err != nil {
			return Review{}, fmt.Errorf("review failed: %w", err)
		}

		content := response.Content()
		approved := strings.HasPrefix(strings.ToUpper(content), "APPROVE")

		score := 50
		if approved {
			score = 85
		}

		return Review{
			Reviewer: item.name,
			Approved: approved,
			Comments: content,
			Score:    score,
		}, nil
	}

	reviewAggregator := func(results []Review, currentState state.State) state.State {
		logger.Info("aggregating reviews", "count", len(results))

		approvedCount := 0
		totalScore := 0
		for _, r := range results {
			if r.Approved {
				approvedCount++
			}
			totalScore += r.Score
		}

		avgScore := totalScore / len(results)
		consensus := float64(approvedCount)/float64(len(results)) >= 0.66

		return currentState.
			Set("reviews", results).
			Set("consensus", consensus).
			Set("average_score", avgScore).
			Set("approved_count", approvedCount)
	}

	reviewNode := workflows.ParallelNode(
		parallelCfg,
		reviewAgents,
		reviewProcessor,
		nil,
		reviewAggregator,
	)

	decisionPredicate := func(s state.State) (string, error) {
		consensus, ok := s.Get("consensus")
		if !ok {
			return "reject", nil
		}

		if consensus.(bool) {
			return "approve", nil
		}

		revisionCount, ok := s.Get("revision_count")
		if !ok || revisionCount.(int) < 2 {
			return "revise", nil
		}

		return "reject", nil
	}

	decisionRoutes := workflows.Routes[state.State]{
		Handlers: map[string]workflows.RouteHandler[state.State]{
			"approve": func(ctx context.Context, s state.State) (state.State, error) {
				logger.Info("decision: document approved")

				doc, _ := s.Get("document")
				currentDoc := doc.(Document)
				currentDoc.Status = "approved"

				avgScore, _ := s.Get("average_score")
				approvedCount, _ := s.Get("approved_count")

				decision := Decision{
					Approved: true,
					Reason:   fmt.Sprintf("Consensus reached - %d of 3 reviewers approved (avg score: %d)", approvedCount, avgScore),
				}

				return s.
					Set("document", currentDoc).
					Set("decision", decision).
					Set("workflow_complete", true), nil
			},
			"revise": func(ctx context.Context, s state.State) (state.State, error) {
				logger.Info("decision: revision requested")

				doc, _ := s.Get("document")
				currentDoc := doc.(Document)
				currentDoc.Status = "revision-needed"
				currentDoc.Version++

				avgScore, _ := s.Get("average_score")

				decision := Decision{
					Approved:          false,
					Reason:            fmt.Sprintf("Insufficient consensus (avg score: %d) - revision required", avgScore),
					RecommendedChange: "Address reviewer concerns and resubmit",
				}

				revisionCountKey := "revision_count"
				var revisionCount int
				if existing, ok := s.Get(revisionCountKey); ok {
					revisionCount = existing.(int)
				}
				revisionCount++

				return s.
					Set("document", currentDoc).
					Set("decision", decision).
					Set(revisionCountKey, revisionCount).
					Set("workflow_complete", false), nil
			},
			"reject": func(ctx context.Context, s state.State) (state.State, error) {
				logger.Info("decision: document rejected")

				doc, _ := s.Get("document")
				currentDoc := doc.(Document)
				currentDoc.Status = "rejected"

				revisionCount, _ := s.Get("revision_count")

				decision := Decision{
					Approved: false,
					Reason:   fmt.Sprintf("Maximum revisions (%d) reached without consensus", revisionCount),
				}

				return s.
					Set("document", currentDoc).
					Set("decision", decision).
					Set("workflow_complete", true), nil
			},
		},
	}

	decisionNode := workflows.ConditionalNode(
		conditionalCfg,
		decisionPredicate,
		decisionRoutes,
	)

	finalizeNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		logger.Info("finalizing workflow")

		doc, _ := s.Get("document")
		currentDoc := doc.(Document)

		decision, _ := s.Get("decision")
		currentDecision := decision.(Decision)

		logger.Info("workflow finalized",
			"doc_id", currentDoc.ID,
			"version", currentDoc.Version,
			"status", currentDoc.Status,
			"approved", currentDecision.Approved,
		)

		return s.Set("finalized", true), nil
	})

	if err := graph.AddNode("analyze", analyzeNode); err != nil {
		log.Fatal(err)
	}
	if err := graph.AddNode("review", reviewNode); err != nil {
		log.Fatal(err)
	}
	if err := graph.AddNode("decision", decisionNode); err != nil {
		log.Fatal(err)
	}
	if err := graph.AddNode("finalize", finalizeNode); err != nil {
		log.Fatal(err)
	}

	if err := graph.AddEdge("analyze", "review", nil); err != nil {
		log.Fatal(err)
	}

	if err := graph.AddEdge("review", "decision", state.KeyExists("consensus")); err != nil {
		log.Fatal(err)
	}

	workflowCompletePredicate := func(s state.State) bool {
		complete, ok := s.Get("workflow_complete")
		if !ok {
			return false
		}
		return complete.(bool)
	}

	if err := graph.AddEdge("decision", "finalize", workflowCompletePredicate); err != nil {
		log.Fatal(err)
	}

	if err := graph.AddEdge("decision", "analyze", state.Not(workflowCompletePredicate)); err != nil {
		log.Fatal(err)
	}

	if err := graph.SetEntryPoint("analyze"); err != nil {
		log.Fatal(err)
	}
	if err := graph.SetExitPoint("finalize"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("   ✓ Graph configured with conditional routing + revision loop")
	fmt.Println()

	fmt.Println("3. Executing stateful workflow...")
	fmt.Println()

	initialState := state.New(nil).Set("document", document)

	finalState, err := graph.Execute(ctx, initialState)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("=== Workflow Complete ===")
	fmt.Println()

	if doc, ok := finalState.Get("document"); ok {
		d := doc.(Document)
		fmt.Printf("Document: %s (v%d)\n", d.ID, d.Version)
		fmt.Printf("  Title: %s\n", d.Title)
		fmt.Printf("  Status: %s\n", d.Status)
		fmt.Println()
	}

	if analyses, ok := finalState.Get("analyses"); ok {
		analysesList := analyses.([]Analysis)
		fmt.Printf("Analyses Completed: %d\n", len(analysesList))
		for _, a := range analysesList {
			fmt.Printf("  [%s] %s\n", a.Type, a.Analyst)
			fmt.Printf("    Finding: %s\n", a.Finding)
			if len(a.Issues) > 0 {
				fmt.Printf("    Issues: %v\n", a.Issues)
			}
		}
		fmt.Println()
	}

	if reviews, ok := finalState.Get("reviews"); ok {
		reviewsList := reviews.([]Review)
		fmt.Printf("Reviews Completed: %d\n", len(reviewsList))
		approvedCount, _ := finalState.Get("approved_count")
		avgScore, _ := finalState.Get("average_score")
		fmt.Printf("  Approved: %d of %d (avg score: %d)\n", approvedCount, len(reviewsList), avgScore)
		for _, r := range reviewsList {
			status := "✗ REJECTED"
			if r.Approved {
				status = "✓ APPROVED"
			}
			fmt.Printf("  [%s] %s\n", status, r.Reviewer)
			fmt.Printf("    Comments: %s\n", r.Comments)
		}
		fmt.Println()
	}

	if decision, ok := finalState.Get("decision"); ok {
		d := decision.(Decision)
		status := "REJECTED"
		if d.Approved {
			status = "APPROVED"
		}
		fmt.Printf("Final Decision: %s\n", status)
		fmt.Printf("  Reason: %s\n", d.Reason)
		if d.RecommendedChange != "" {
			fmt.Printf("  Recommendation: %s\n", d.RecommendedChange)
		}
		fmt.Println()
	}

	if revCount, ok := finalState.Get("revision_count"); ok {
		fmt.Printf("Revisions: %d\n", revCount)
		fmt.Println()
	}

	fmt.Println("Workflow Features Demonstrated:")
	fmt.Println("  ✓ ChainNode - Sequential analysis by 3 specialists")
	fmt.Println("  ✓ ParallelNode - Concurrent review by 3 reviewers")
	fmt.Println("  ✓ ConditionalNode - Decision routing (approve/revise/reject)")
	fmt.Println("  ✓ State Management - Document, analyses, reviews, decisions")
	fmt.Println("  ✓ Conditional Edges - Workflow loops based on state")
	fmt.Println("  ✓ Checkpointing - State persisted after each node")
}
