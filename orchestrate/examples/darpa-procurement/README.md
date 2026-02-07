# DARPA Research Procurement Workflow Demo

A comprehensive orchestration demonstration showcasing v0.1.0 patterns through a realistic DARPA research procurement simulation with dynamic R&D project generation and full workflow orchestration.

## Scenario Overview

**Context**: Defense Advanced Research Projects Agency (DARPA) multi-stage procurement system from R&D project conception through executive approval, demonstrating complete orchestration lifecycle.

This simulation models a realistic defense research procurement workflow where:
1. Research directors conceive innovative R&D projects
2. Cost analysts evaluate and budget the requests
3. Multi-stage approval workflows route requests based on cost and classification
4. Legal, security, and executive reviews ensure compliance and oversight

### Orchestration Features

This example demonstrates the following tau-orchestrate capabilities:

**Features Demonstrated:**
- **State Graphs** - 8-node workflow with sequential execution and multiple exit points
- **Conditional Routing** - Edge predicates for cost-based thresholds and decision-based branching
- **Parallel Execution** - ParallelNode with configurable FailFast behavior for financial analysis and legal reviews
- **Checkpoint Recovery** - Automatic state preservation with failure injection and resume capabilities
- **Immutable State** - Thread-safe state management flowing through all workflow nodes
- **Observability** - SlogObserver for structured JSON logging of workflow events (via `--verbose`)

**Features Not Demonstrated:**
- **Hub/Messaging** - Multi-agent coordination with message routing (see [phase-01-hubs](../phase-01-hubs/))
- **Sequential Chains** - ChainNode pattern for step-by-step processing (see [phase-04-sequential-chains](../phase-04-sequential-chains/))

### Simulation Rules

The workflow enforces these routing and review requirements:

**Cost-Based Routing Thresholds:**
- **< $50,000**: Low-cost fast track → Program Director approval (no legal review)
- **$50,000 - $199,999**: Standard legal review → Program Director approval
- **≥ $200,000**: Full compliance review (legal + security) → Deputy Director approval

**Emergency Bypass (`--skip-legal`):**
- Skips all legal and security review stages
- Routes based on cost threshold:
  - < $200,000 → Program Director (expedited)
  - ≥ $200,000 → Deputy Director (expedited)

**Failure Injection Requirements (`--fail-at`):**
- **Financial**: Works for any project (always executed)
- **Legal**: Requires cost ≥ $50,000 (triggers legal review)
- **Security**: Requires cost ≥ $200,000 (triggers security review)
- Projects below thresholds will complete successfully without reaching the failure point

**Project Cost Distribution:**
Projects are randomly generated with costs ranging from $10,000 to $500,000 based on:
- Component count (more components = higher cost)
- Classification level (TOP SECRET adds premium)
- Technology complexity (cutting-edge increases budget)

## Technical Implementation

### File Structure

```
darpa-procurement/
├── main.go                    # CLI, simulation orchestration
├── agents.go                  # Agent initialization and system prompts
├── projects.go                # R&D project templates and cost logic
├── workflow.go                # State graph construction and routing
├── responses.go               # Response structure definitions
├── parser.go                  # JSON parsing with fallback
├── config.go                  # Configuration and flag parsing
├── config.gemma.json          # Gemma model configuration
├── README.md                  # This file
├── TEST.md                    # Test validation plan
└── TECHNICAL.md               # Technical implementation guide
```

### Agent Configuration

All agents use **gemma3:4b** running in Ollama with the following configuration:

- **Base URL**: `http://localhost:11434`
- **Model**: `gemma3:4b`
- **Max Tokens**: `4096` (configurable via `--max-tokens`)
- **Temperature**: Default per model
- **Capabilities**: Chat support with structured JSON output

Each agent role has a unique system prompt defining its expertise and enforcing RFC7159-compliant JSON responses.

### Prerequisites

- **Ollama** installed and running: `ollama serve`
- **gemma3:4b** model pulled: `ollama pull gemma3:4b`

## User Inputs

### Command-Line Flags

**Simulation Parameters:**
- `--requests N` - Number of R&D projects to simulate (1-8, default: 2)
- `--config PATH` - Agent configuration file (default: `config.gemma.json`)
- `--max-tokens N` - Override max tokens for responses (default: 0, uses config value)

**Workflow Configuration:**
- `--mode [fast|balanced|thorough]` - Analysis depth (default: balanced)
  - **fast**: Single reviewer, expedited analysis
  - **balanced**: 2 reviewers, standard thresholds
  - **thorough**: 3 reviewers, comprehensive review
- `--reviewers N` - Legal/compliance reviewers (1-3, default: 2)
- `--skip-legal` - Emergency procurement bypass (skips legal/security review)
- `--fail-at [financial|legal|security]` - Inject failure for checkpoint demo

**Observability:**
- `-v, --verbose` - Enable SlogObserver with JSON event logging

## Usage

Run the simulation with default settings (2 requests, balanced mode):

```bash
go run ./examples/darpa-procurement/.
```

### Sample Output

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/2 ===
→ Drafting procurement request: Hypersonic Flight Control Systems
   Hypersonic Flight Control Systems - TOP SECRET, Advanced Aerodynamics

→ Analyzing procurement costs...
   $385000 | Risk: HIGH | Route: Expedited Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget significantly exceeds program allocations, high financial risk identified
  Optimization: 85000 potential savings

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


R&D Project: Hypersonic Flight Control Systems
  Classification: TOP SECRET
  Components: 7
  Estimated Cost: $385,000
  Risk Level: HIGH
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

=== Processing Request 2/2 ===
→ Drafting procurement request: Space-Based Sensor Networks
   Space-Based Sensor Networks - TOP SECRET, Global Surveillance and Space Situational Awareness

→ Analyzing procurement costs...
   $385000 | Risk: HIGH | Route: Expedited Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget significantly exceeds program allocations, high financial risk identified
  Optimization: 85000 potential savings

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


R&D Project: Space-Based Sensor Networks
  Classification: TOP SECRET
  Components: 6
  Estimated Cost: $385,000
  Risk Level: HIGH
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-002

Summary:
- Requests processed: 2
- Approved: 2 (PR-2024-001, PR-2024-002)
- Rejected: 0
- Total budget allocated: $770,000
- Total processing time: 24.9s
- Average time per request: 12.5s
```

For additional usage examples and comprehensive test scenarios, see [TEST.md](./TEST.md).
