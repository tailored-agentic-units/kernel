package main

import (
	"flag"
	"fmt"
)

type WorkflowMode string

const (
	ModeFast     WorkflowMode = "fast"
	ModeBalanced WorkflowMode = "balanced"
	ModeThorough WorkflowMode = "thorough"
)

type FailureStage string

const (
	FailureNone      FailureStage = ""
	FailureFinancial FailureStage = "financial"
	FailureLegal     FailureStage = "legal"
	FailureSecurity  FailureStage = "security"
)

type WorkflowConfig struct {
	AgentConfig string
	MaxTokens   int
	Requests    int
	Reviewers   int
	Mode        WorkflowMode
	SkipLegal   bool
	FailAt      FailureStage
	Verbose     bool
}

func ParseConfig() (*WorkflowConfig, error) {
	config := &WorkflowConfig{}

	var modeStr string
	var failAtStr string

	flag.StringVar(&config.AgentConfig, "config", "examples/darpa-procurement/config.gemma.json", "Path to agent configuration file")
	flag.IntVar(&config.MaxTokens, "max-tokens", 0, "Override max tokens for agent responses")
	flag.IntVar(&config.Requests, "requests", 2, "Number of R&D projects to simulate (1-5)")
	flag.IntVar(&config.Reviewers, "reviewers", 2, "Number of legal/compliance reviewers for parallel review (1-3)")
	flag.StringVar(&modeStr, "mode", "balanced", "Analysis depth: fast, balanced, or thorough")
	flag.BoolVar(&config.SkipLegal, "skip-legal", false, "Emergency procurement bypass (skips legal/security review)")
	flag.StringVar(&failAtStr, "fail-at", "", "Inject failure at stage for checkpoint demo: financial, legal, or security")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose mode with SlogObserver")

	flag.Parse()

	maxRequests := len(projectTemplates)
	if config.Requests < 1 || config.Requests > maxRequests {
		return nil, fmt.Errorf("requests must be between 1 and %d (number of available project templates), got %d", maxRequests, config.Requests)
	}

	if config.Reviewers < 1 || config.Reviewers > 3 {
		return nil, fmt.Errorf("reviewers must be between 1 and 3, got %d", config.Reviewers)
	}

	switch modeStr {
	case "fast":
		config.Mode = ModeFast
		config.Reviewers = 1
	case "balanced":
		config.Mode = ModeBalanced
	case "thorough":
		config.Mode = ModeThorough
		if config.Reviewers < 2 {
			config.Reviewers = 3
		}
	default:
		return nil, fmt.Errorf("mode must be fast, balanced, or thorough, got %s", modeStr)
	}

	switch failAtStr {
	case "":
		config.FailAt = FailureNone
	case "financial":
		config.FailAt = FailureFinancial
	case "legal":
		config.FailAt = FailureLegal
	case "security":
		config.FailAt = FailureSecurity
	default:
		return nil, fmt.Errorf("fail-at must be empty, financial, legal, or security, got %s", failAtStr)
	}

	return config, nil
}

func (c *WorkflowConfig) ObserverName() string {
	if c.Verbose {
		return "slog"
	}
	return "noop"
}
