package main

import (
	"context"
	"fmt"

	"github.com/tailored-agentic-units/kernel/core/response"
	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/state"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
)


func BuildWorkflow(wc *WorkflowConfig, registry *AgentRegistry) (state.StateGraph, error) {
	graphConfig := config.GraphConfig{
		Name:     "darpa-procurement",
		Observer: wc.ObserverName(),
		Checkpoint: config.CheckpointConfig{
			Store:    "memory",
			Interval: 1,
			Preserve: true,
		},
		MaxIterations: 10,
	}

	graph, err := state.NewGraph(graphConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	entryNode := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s, nil
	})
	graph.AddNode("entry", entryNode)
	graph.SetEntryPoint("entry")

	draftingNode := createDraftingNode(registry)
	graph.AddNode("request_drafting", draftingNode)

	costNode := createCostAnalysisNode(registry)
	graph.AddNode("cost_analysis", costNode)

	validationNode := createValidationNode(registry)
	graph.AddNode("procurement_validation", validationNode)

	financialNode := createFinancialAnalysisNode(wc, registry)
	graph.AddNode("financial_analysis", financialNode)

	routingNode := createRoutingNode(wc, registry)
	graph.AddNode("approval_routing", routingNode)

	approvedExit := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s, nil
	})
	graph.AddNode("approved", approvedExit)
	graph.SetExitPoint("approved")

	rejectedExit := state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		return s, nil
	})
	graph.AddNode("rejected", rejectedExit)
	graph.SetExitPoint("rejected")

	graph.AddEdge("entry", "request_drafting", nil)
	graph.AddEdge("request_drafting", "cost_analysis", nil)
	graph.AddEdge("cost_analysis", "procurement_validation", nil)
	graph.AddEdge("procurement_validation", "financial_analysis", nil)
	graph.AddEdge("financial_analysis", "approval_routing", nil)

	approvedPredicate := func(s state.State) bool {
		decision, _ := s.Get("decision")
		return decision == "APPROVED"
	}
	graph.AddEdge("approval_routing", "approved", approvedPredicate)

	rejectedPredicate := func(s state.State) bool {
		decision, _ := s.Get("decision")
		return decision == "REJECTED"
	}
	graph.AddEdge("approval_routing", "rejected", rejectedPredicate)

	revisionPredicate := func(s state.State) bool {
		decision, _ := s.Get("decision")
		iterations, _ := s.Get("iterations")
		iter, ok := iterations.(int)
		if !ok {
			return false
		}
		return decision == "NEEDS REVISION" && iter < 2
	}
	graph.AddEdge("approval_routing", "procurement_validation", revisionPredicate)

	maxRetriesPredicate := func(s state.State) bool {
		decision, _ := s.Get("decision")
		iterations, _ := s.Get("iterations")
		iter, ok := iterations.(int)
		if !ok {
			return false
		}
		return decision == "NEEDS REVISION" && iter >= 2
	}
	graph.AddEdge("approval_routing", "rejected", maxRetriesPredicate)

	return graph, nil
}

func createDraftingNode(registry *AgentRegistry) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		project := GetRandomProject()
		fmt.Printf("→ Drafting procurement request: %s\n", project.Name)

		prompt := fmt.Sprintf(`Draft a procurement request for the following R&D project:

Project: %s
Category: %s
Classification: %s
Description: %s
Required Components: %d

Provide your response in your directed JSON format.`,
			project.Name,
			project.Category,
			project.Classification,
			project.Description,
			project.ComponentCount())

		response, err := registry.ResearchDirector.Chat(ctx, prompt)
		if err != nil {
			return s, fmt.Errorf("research director failed: %w", err)
		}

		request, err := parseJSON[ProcurementRequest](response.Content())
		if err != nil {
			return s, fmt.Errorf("failed to parse procurement request: %w", err)
		}

		fmt.Printf("   %s\n\n", request.ProjectSummary)

		newState := s.
			Set("project_name", project.Name).
			Set("project_category", project.Category).
			Set("classification", string(project.Classification)).
			Set("component_count", project.ComponentCount()).
			Set("complexity_score", project.ComplexityScore()).
			Set("procurement_request", request).
			Set("iterations", 0)

		return newState, nil
	})
}

func createCostAnalysisNode(registry *AgentRegistry) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Printf("→ Analyzing procurement costs...\n")

		procReq, _ := s.Get("procurement_request")
		request := procReq.(ProcurementRequest)
		classification, _ := s.Get("classification")
		componentCount, _ := s.Get("component_count")
		complexityScore, _ := s.Get("complexity_score")

		prompt := fmt.Sprintf(`Analyze this procurement request and provide a realistic budget estimate:

Classification: %s
Component Count: %d
Complexity Score: %d

Project Summary: %s
Technical Requirements: %v
Components: %v
Justification: %s

Provide your response in your directed JSON format.`,
			classification,
			componentCount,
			complexityScore,
			request.ProjectSummary,
			request.TechnicalReqs,
			request.Components,
			request.Justification)

		response, err := registry.CostAnalyst.Chat(ctx, prompt)
		if err != nil {
			return s, fmt.Errorf("cost analyst failed: %w", err)
		}

		analysis, err := parseJSON[CostAnalysis](response.Content())
		if err != nil {
			return s, fmt.Errorf("failed to parse cost analysis: %w", err)
		}

		fmt.Printf("   $%d | Risk: %s | Route: %s\n\n", analysis.EstimatedCost, analysis.RiskLevel, analysis.Route)

		newState := s.
			Set("cost_analysis", analysis).
			Set("estimated_cost", analysis.EstimatedCost).
			Set("risk_level", analysis.RiskLevel)

		return newState, nil
	})
}

func createValidationNode(registry *AgentRegistry) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		iterations, _ := s.Get("iterations")
		iter := iterations.(int)

		if iter > 0 {
			fmt.Printf("→ Validating revised procurement request (revision %d)...\n", iter)
		} else {
			fmt.Printf("→ Validating procurement request...\n")
		}

		procReq, _ := s.Get("procurement_request")
		request := procReq.(ProcurementRequest)
		classification, _ := s.Get("classification")
		componentCount, _ := s.Get("component_count")

		iterationNote := ""
		if iter > 0 {
			iterationNote = fmt.Sprintf("\n\nNOTE: This is revision %d based on previous feedback. Verify all revisions have been addressed.", iter)
		}

		prompt := fmt.Sprintf(`Validate this procurement request for technical completeness and compliance:

Classification: %s
Component Count: %d%s

Project Summary: %s
Technical Requirements: %v
Components: %v

Provide your response in your directed JSON format.`,
			classification,
			componentCount,
			iterationNote,
			request.ProjectSummary,
			request.TechnicalReqs,
			request.Components)

		response, err := registry.ProcurementSpecialist.Chat(ctx, prompt)
		if err != nil {
			return s, fmt.Errorf("procurement specialist failed: %w", err)
		}

		validation, err := parseJSON[ValidationResult](response.Content())
		if err != nil {
			return s, fmt.Errorf("failed to parse validation result: %w", err)
		}

		fmt.Printf("   %s\n\n", validation.Status)

		newState := s.Set("validation_result", validation)

		return newState, nil
	})
}

func createFinancialAnalysisNode(wc *WorkflowConfig, registry *AgentRegistry) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		fmt.Printf("→ Conducting financial analysis (parallel: budget validation + cost optimization)...\n")

		procReq, _ := s.Get("procurement_request")
		request := procReq.(ProcurementRequest)
		estimatedCost, _ := s.Get("estimated_cost")
		riskLevel, _ := s.Get("risk_level")

		type AnalysisTask struct {
			Name   string
			Prompt string
		}

		tasks := []AnalysisTask{
			{
				Name: "budget",
				Prompt: fmt.Sprintf(`Validate this procurement budget against program allocations:

Estimated Cost: $%d
Risk Level: %s

Project Summary: %s
Justification: %s

Provide your response in your directed JSON format.`,
					estimatedCost,
					riskLevel,
					request.ProjectSummary,
					request.Justification),
			},
			{
				Name: "optimizer",
				Prompt: fmt.Sprintf(`Analyze this procurement for cost optimization opportunities:

Estimated Cost: $%d
Risk Level: %s

Project Summary: %s
Components: %v

Provide your response in your directed JSON format.`,
					estimatedCost,
					riskLevel,
					request.ProjectSummary,
					request.Components),
			},
		}

		type AnalysisResult struct {
			Name       string
			Validation *BudgetValidation
			Optimization *CostOptimization
		}

		processor := func(ctx context.Context, task AnalysisTask) (AnalysisResult, error) {
			var resp *response.ChatResponse
			var err error

			switch task.Name {
			case "budget":
				resp, err = registry.BudgetAnalyst.Chat(ctx, task.Prompt)
				if err != nil {
					return AnalysisResult{}, fmt.Errorf("budget analyst failed: %w", err)
				}
				validation, parseErr := parseJSON[BudgetValidation](resp.Content())
				if parseErr != nil {
					return AnalysisResult{}, fmt.Errorf("failed to parse budget validation: %w", parseErr)
				}
				return AnalysisResult{Name: "budget", Validation: &validation}, nil

			case "optimizer":
				resp, err = registry.CostOptimizer.Chat(ctx, task.Prompt)
				if err != nil {
					return AnalysisResult{}, fmt.Errorf("cost optimizer failed: %w", err)
				}
				optimization, parseErr := parseJSON[CostOptimization](resp.Content())
				if parseErr != nil {
					return AnalysisResult{}, fmt.Errorf("failed to parse cost optimization: %w", parseErr)
				}
				return AnalysisResult{Name: "optimizer", Optimization: &optimization}, nil

			default:
				return AnalysisResult{}, fmt.Errorf("unknown task: %s", task.Name)
			}
		}

		aggregator := func(results []AnalysisResult, currentState state.State) state.State {
			newState := currentState
			for _, result := range results {
				if result.Validation != nil {
					newState = newState.Set("budget_validation", *result.Validation)
				}
				if result.Optimization != nil {
					newState = newState.Set("cost_optimization", *result.Optimization)
				}
			}
			return newState
		}

		failFast := true
		parallelCfg := config.ParallelConfig{
			Observer:    wc.ObserverName(),
			FailFastNil: &failFast,
			MaxWorkers:  2,
		}

		parallelNode := workflows.ParallelNode(parallelCfg, tasks, processor, nil, aggregator)
		newState, err := parallelNode.Execute(ctx, s)
		if err != nil {
			return s, err
		}

		budgetVal, _ := newState.Get("budget_validation")
		budget := budgetVal.(BudgetValidation)
		costOpt, _ := newState.Get("cost_optimization")
		optimization := costOpt.(CostOptimization)

		fmt.Printf("  Budget: %s\n", budget.Assessment)
		fmt.Printf("  Optimization: %d potential savings\n\n", optimization.Savings)

		if wc.FailAt == FailureFinancial {
			return newState, fmt.Errorf("simulated failure at financial stage")
		}

		return newState, nil
	})
}

func createRoutingNode(wc *WorkflowConfig, registry *AgentRegistry) state.StateNode {
	return state.NewFunctionNode(func(ctx context.Context, s state.State) (state.State, error) {
		estimatedCost, _ := s.Get("estimated_cost")
		classification, _ := s.Get("classification")
		cost := estimatedCost.(int)
		classLevel := classification.(string)

		if wc.SkipLegal {
			if cost < 200000 {
				return routeToExecutive(ctx, s, registry.ProgramDirector, "Program Director", cost, "expedited")
			} else {
				return routeToExecutive(ctx, s, registry.DeputyDirector, "Deputy Director", cost, "expedited")
			}
		}

		if cost < 50000 {
			return routeToExecutive(ctx, s, registry.ProgramDirector, "Program Director", cost, "low-cost")
		}

		if cost < 200000 {
			newState, err := performLegalReview(ctx, s, wc, registry, classLevel, false)
			if err != nil {
				return s, err
			}

			legalStatus, _ := newState.Get("legal_status")
			if legalStatus == "REJECTED" {
				return newState.Set("decision", "REJECTED"), nil
			}
			if legalStatus == "NEEDS REVISION" {
				return handleRevision(newState)
			}

			return routeToExecutive(ctx, newState, registry.ProgramDirector, "Program Director", cost, "standard-legal")
		}

		newState, err := performLegalReview(ctx, s, wc, registry, classLevel, true)
		if err != nil {
			return s, err
		}

		legalStatus, _ := newState.Get("legal_status")
		securityStatus, _ := newState.Get("security_status")

		if legalStatus == "REJECTED" || securityStatus == "REJECTED" {
			return newState.Set("decision", "REJECTED"), nil
		}

		if legalStatus == "NEEDS REVISION" || securityStatus == "NEEDS REVISION" {
			return handleRevision(newState)
		}

		return routeToExecutive(ctx, newState, registry.DeputyDirector, "Deputy Director", cost, "full-security-review")
	})
}

func performLegalReview(ctx context.Context, s state.State, wc *WorkflowConfig, registry *AgentRegistry, classification string, includeSecurityReview bool) (state.State, error) {
	reviewerCount := len(registry.LegalReviewers)
	if includeSecurityReview {
		fmt.Printf("→ Conducting compliance review (parallel: %d legal reviewers + security officer)...\n", reviewerCount)
	} else {
		fmt.Printf("→ Conducting legal review (parallel: %d reviewers)...\n", reviewerCount)
	}

	procReq, _ := s.Get("procurement_request")
	request := procReq.(ProcurementRequest)
	cost, _ := s.Get("estimated_cost")

	legalPrompt := fmt.Sprintf(`Review this procurement request for legal compliance:

Classification: %s
Estimated Cost: $%d

Project Summary: %s
Justification: %s

Provide your response in your directed JSON format.`,
		classification,
		cost,
		request.ProjectSummary,
		request.Justification)

	type LegalTask struct {
		ReviewerIndex int
	}

	tasks := make([]LegalTask, len(registry.LegalReviewers))
	for i := range tasks {
		tasks[i] = LegalTask{ReviewerIndex: i}
	}

	processor := func(ctx context.Context, task LegalTask) (LegalReview, error) {
		reviewer := registry.LegalReviewers[task.ReviewerIndex]
		response, err := reviewer.Chat(ctx, legalPrompt)
		if err != nil {
			return LegalReview{}, fmt.Errorf("legal reviewer %d failed: %w", task.ReviewerIndex+1, err)
		}

		review, parseErr := parseJSON[LegalReview](response.Content())
		if parseErr != nil {
			return LegalReview{}, fmt.Errorf("failed to parse legal review %d: %w", task.ReviewerIndex+1, parseErr)
		}

		return review, nil
	}

	aggregator := func(results []LegalReview, currentState state.State) state.State {
		decisions := make([]string, len(results))
		for i, result := range results {
			decisions[i] = result.Decision
		}

		consensus := calculateConsensus(decisions)

		return currentState.
			Set("legal_status", consensus).
			Set("legal_reviews", results)
	}

	failFast := false
	parallelCfg := config.ParallelConfig{
		Observer:    wc.ObserverName(),
		FailFastNil: &failFast,
		MaxWorkers:  len(registry.LegalReviewers),
	}

	parallelNode := workflows.ParallelNode(parallelCfg, tasks, processor, nil, aggregator)
	newState, err := parallelNode.Execute(ctx, s)
	if err != nil {
		return s, err
	}

	legalStatus, _ := newState.Get("legal_status")
	fmt.Printf("  Legal Review Consensus: %s\n", legalStatus)

	if wc.FailAt == FailureLegal {
		return newState, fmt.Errorf("simulated failure at legal stage")
	}

	if includeSecurityReview {
		securityPrompt := fmt.Sprintf(`Review this procurement request for security compliance:

Classification: %s
Estimated Cost: $%d

Project Summary: %s

Provide your response in your directed JSON format.`,
			classification,
			cost,
			request.ProjectSummary)

		response, err := registry.SecurityOfficer.Chat(ctx, securityPrompt)
		if err != nil {
			return newState, fmt.Errorf("security officer failed: %w", err)
		}

		security, parseErr := parseJSON[SecurityReview](response.Content())
		if parseErr != nil {
			return newState, fmt.Errorf("failed to parse security review: %w", parseErr)
		}

		fmt.Printf("  Security Review: %s\n", security.Decision)

		if wc.FailAt == FailureSecurity {
			return newState, fmt.Errorf("simulated failure at security stage")
		}

		newState = newState.
			Set("security_status", security.Decision).
			Set("security_review", security)
	} else {
		fmt.Println()
	}

	return newState, nil
}

func routeToExecutive(ctx context.Context, s state.State, executive interface {
	Chat(context.Context, string, ...map[string]any) (*response.ChatResponse, error)
}, title string, cost int, route string) (state.State, error) {
	fmt.Printf("→ Routing to %s for final approval (route: %s)...\n", title, route)

	procReq, _ := s.Get("procurement_request")
	request := procReq.(ProcurementRequest)
	riskLevel, _ := s.Get("risk_level")
	legalStatus, ok := s.Get("legal_status")
	if !ok {
		legalStatus = "N/A"
	}

	prompt := fmt.Sprintf(`Review this procurement request for final approval:

Route: %s
Estimated Cost: $%d
Risk Level: %s
Legal Status: %s

Project Summary: %s
Justification: %s

Make your decision: APPROVED, NEEDS_REVISION, or REJECTED.
Provide justification for your decision.`,
		route,
		cost,
		riskLevel,
		legalStatus,
		request.ProjectSummary,
		request.Justification)

	response, err := executive.Chat(ctx, prompt)
	if err != nil {
		return s, fmt.Errorf("%s approval failed: %w", title, err)
	}

	decision, parseErr := parseJSON[ExecutiveDecision](response.Content())
	if parseErr != nil {
		return s, fmt.Errorf("failed to parse executive decision: %w", parseErr)
	}

	fmt.Printf("  Decision: %s\n\n", decision.Decision)

	newState := s.
		Set("executive_decision", decision).
		Set("approval_level", title).
		Set("decision", decision.Decision)

	return newState, nil
}

func handleRevision(s state.State) (state.State, error) {
	iterations, _ := s.Get("iterations")
	iter := iterations.(int)

	return s.
		Set("iterations", iter+1).
		Set("decision", "NEEDS REVISION"), nil
}

func calculateConsensus(statuses []string) string {
	rejectedCount := 0
	revisionCount := 0
	approvedCount := 0

	for _, status := range statuses {
		switch status {
		case "REJECTED":
			rejectedCount++
		case "NEEDS REVISION":
			revisionCount++
		case "APPROVED":
			approvedCount++
		}
	}

	if rejectedCount > 0 {
		return "REJECTED"
	}
	if revisionCount > len(statuses)/2 {
		return "NEEDS REVISION"
	}
	return "APPROVED"
}
