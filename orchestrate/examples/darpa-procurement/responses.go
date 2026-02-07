package main

type ProcurementRequest struct {
	ProjectSummary    string   `json:"project_summary"`
	TechnicalReqs     []string `json:"technical_requirements"`
	Components        []string `json:"components"`
	Justification     string   `json:"justification"`
}

type CostAnalysis struct {
	EstimatedCost int      `json:"estimated_cost"`
	RiskLevel     string   `json:"risk_level"`
	Breakdown     []string `json:"cost_breakdown"`
	Route         string   `json:"recommended_route"`
	Reasoning     string   `json:"reasoning"`
}

type ValidationResult struct {
	Status   string   `json:"status"`
	Findings []string `json:"findings"`
	Concerns []string `json:"concerns"`
}

type BudgetValidation struct {
	Approved   bool     `json:"approved"`
	Assessment string   `json:"assessment"`
	Concerns   []string `json:"concerns"`
	Risk       string   `json:"financial_risk"`
}

type CostOptimization struct {
	Savings      int      `json:"potential_savings"`
	Alternatives []string `json:"alternatives"`
	Impact       string   `json:"capability_impact"`
}

type LegalReview struct {
	Decision   string   `json:"decision"`
	Reasoning  string   `json:"reasoning"`
	Concerns   []string `json:"concerns"`
	Compliance bool     `json:"far_compliant"`
}

type SecurityReview struct {
	Decision   string   `json:"decision"`
	Assessment string   `json:"assessment"`
	Clearance  string   `json:"clearance_level"`
	Concerns   []string `json:"concerns"`
}

type ExecutiveDecision struct {
	Decision      string   `json:"decision"`
	Justification string   `json:"justification"`
	Conditions    []string `json:"conditions"`
}
