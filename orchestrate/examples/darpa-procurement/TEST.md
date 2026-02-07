# DARPA Procurement Workflow - Test Validation Plan

## Test 1: Basic Workflow - Default Configuration

Purpose: Validate standard workflow execution with balanced mode

```sh
go run ./examples/darpa-procurement/.
```

Expected Results:
- 2 random projects processed
- 2 legal reviewers for parallel review
- Conditional routing based on cost/risk analysis
- Parallel execution of financial analysis (budget + optimization)
- Parallel execution of compliance review (legal + security for high-risk)
- Appropriate executive approval (Program Director or Deputy Director)
- Success rate tracking and budget allocation summary

Result:

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

---

## Test 2: Fast Mode - Minimal Review

Purpose: Validate fast mode with single reviewer

```sh
go run ./examples/darpa-procurement/. -mode fast -requests 1
```

Expected Results:
- 1 project processed
- Only 1 legal reviewer (fast mode auto-adjusts)
- Faster execution time
- All other workflow features still functional

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
→ Drafting procurement request: Neural Interface Brain-Computer Systems
   Neural Interface Brain-Computer Systems - SECRET, Human Performance Enhancement

→ Analyzing procurement costs...
   $32000 | Risk: MEDIUM | Route: Standard Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 8000 potential savings

→ Routing to Program Director for final approval (route: low-cost)...
  Decision: APPROVED


R&D Project: Neural Interface Brain-Computer Systems
  Classification: SECRET
  Components: 4
  Estimated Cost: $32,000
  Risk Level: MEDIUM

Final Decision:
  ✓ APPROVED by Program Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $32,000
- Total processing time: 8.7s
- Average time per request: 8.7s
```

---

## Test 3: Thorough Mode - Maximum Review

Purpose: Validate thorough analysis with maximum reviewers

```sh
go run ./examples/darpa-procurement/. -mode thorough -requests 1
```

Expected Results:
- 1 project processed
- 3 legal reviewers for comprehensive parallel review
- Longer execution time due to additional reviewers
- Higher confidence in legal consensus

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
→ Drafting procurement request: Hypersonic Flight Control Systems
   Hypersonic Flight Control Systems - TOP SECRET, Advanced Aerodynamics

→ Analyzing procurement costs...
   $380000 | Risk: HIGH | Route: Expedited Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget significantly exceeds program allocations, high financial risk identified
  Optimization: 8000 potential savings

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


R&D Project: Hypersonic Flight Control Systems
  Classification: TOP SECRET
  Components: 8
  Estimated Cost: $380,000
  Risk Level: HIGH
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $380,000
- Total processing time: 12.8s
- Average time per request: 12.8s
```

---

## Test 4: High-Volume Processing

Purpose: Validate workflow handles maximum project load

```sh
go run ./examples/darpa-procurement/. -requests 3
```

Expected Results:
- All 3 projects processed (uses different project templates)
- Mix of SECRET and TOP SECRET classifications
- Mix of approval routes (low-cost, standard, expedited, full-security)
- Budget allocation tracked across all projects
- Performance metrics for full batch

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/3 ===
→ Drafting procurement request: Quantum Entanglement Communication
   Quantum Entanglement Communication - TOP SECRET, Secure Communications

→ Analyzing procurement costs...
   $38000 | Risk: MEDIUM | Route: Standard Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 8000 potential savings

→ Routing to Program Director for final approval (route: low-cost)...
  Decision: APPROVED


R&D Project: Quantum Entanglement Communication
  Classification: TOP SECRET
  Components: 6
  Estimated Cost: $38,000
  Risk Level: MEDIUM

Final Decision:
  ✓ APPROVED by Program Director
  Award ID: PR-2024-001

=== Processing Request 2/3 ===
→ Drafting procurement request: Space-Based Sensor Networks
   Space-Based Sensor Networks - TOP SECRET, Global Persistent Surveillance

→ Analyzing procurement costs...
   $385000 | Risk: HIGH | Route: Expedited Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget significantly exceeds program allocations, high financial risk identified
  Optimization: 85000 potential savings

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: NEEDS_REVISION
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


R&D Project: Space-Based Sensor Networks
  Classification: TOP SECRET
  Components: 6
  Estimated Cost: $385,000
  Risk Level: HIGH
  Legal Review: APPROVED
  Security Review: NEEDS_REVISION

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-002

=== Processing Request 3/3 ===
→ Drafting procurement request: Neural Interface Brain-Computer Systems
   Neural Interface Brain-Computer Systems - SECRET, Human Performance Enhancement

→ Analyzing procurement costs...
   $32500 | Risk: MEDIUM | Route: Standard Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 8000 potential savings

→ Routing to Program Director for final approval (route: low-cost)...
  Decision: APPROVED


R&D Project: Neural Interface Brain-Computer Systems
  Classification: SECRET
  Components: 4
  Estimated Cost: $32,500
  Risk Level: MEDIUM

Final Decision:
  ✓ APPROVED by Program Director
  Award ID: PR-2024-003

Summary:
- Requests processed: 3
- Approved: 3 (PR-2024-001, PR-2024-002, PR-2024-003)
- Rejected: 0
- Total budget allocated: $455,500
- Total processing time: 30.8s
- Average time per request: 10.3s
```

---

## Test 5: Emergency Bypass - Skip Legal Review

Purpose: Validate emergency procurement path

```sh
go run ./examples/darpa-procurement/. -skip-legal -requests 1
```

Expected Results:
- 1 project processed
- Legal and security review stages skipped
- Direct routing to executive approval
- Faster execution (fewer workflow nodes)
- Clear indication that bypass was used

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
→ Drafting procurement request: Directed Energy Weapon Miniaturization
   Directed Energy Weapon Miniaturization - TOP SECRET, Advanced Weapons

→ Analyzing procurement costs...
   $285000 | Risk: HIGH | Route: Expedited Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget significantly exceeds program allocations, high financial risk identified
  Optimization: 7500 potential savings

→ Routing to Deputy Director for final approval (route: expedited)...
  Decision: APPROVED


R&D Project: Directed Energy Weapon Miniaturization
  Classification: TOP SECRET
  Components: 6
  Estimated Cost: $285,000
  Risk Level: HIGH

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $285,000
- Total processing time: 9.4s
- Average time per request: 9.4s
```

---

## Test 6: Checkpoint Recovery - Financial Stage Failure

Purpose: Validate checkpoint/recovery at financial analysis stage

```sh
go run ./examples/darpa-procurement/. -fail-at financial -requests 1
```

Expected Results:
- First request: Execution up to financial stage
- Simulated failure injected
- Checkpoint saved with run ID "darpa-pr-checkpoint-demo"
- Automatic recovery from checkpoint
- Workflow resumes from financial stage
- Request completes successfully after recovery
- Recovery statistics displayed

Results:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
NOTE: Failure injection enabled at stage: financial

→ Drafting procurement request: Synthetic Organism Detection Biosensors
   Synthetic Organism Detection Biosensors - SECRET, Distributed Surveillance

→ Analyzing procurement costs...
   $375000 | Risk: MEDIUM | Route: Expedited Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 85000 potential savings


✗ SIMULATED FAILURE at financial stage
Checkpoint saved: procurement_validation (runID: c44400e3-d11d-4e61-88d8-a51f01d101ae)

=== Resuming from Checkpoint ===
RunID: c44400e3-d11d-4e61-88d8-a51f01d101ae
Checkpoint: procurement_validation

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 8000 potential savings

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


Recovery Statistics:
- Checkpoint recovery successful
- State preserved across failure

R&D Project: Synthetic Organism Detection Biosensors
  Classification: SECRET
  Components: 5
  Estimated Cost: $375,000
  Risk Level: MEDIUM
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $375,000
- Total processing time: 13.8s
- Average time per request: 13.8s
```

---

## Test 7: Checkpoint Recovery - Legal Stage Failure

Purpose: Validate checkpoint/recovery at legal review stage

```sh
go run ./examples/darpa-procurement/. -fail-at legal -requests 1
```

Expected Results:
- First request: Execution through financial analysis, validation
- Simulated failure at legal review stage
- Checkpoint saved
- Automatic recovery
- Workflow resumes from legal review
- Request completes successfully

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
NOTE: Failure injection enabled at stage: legal

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

✗ SIMULATED FAILURE at legal stage
Checkpoint saved: financial_analysis (runID: 31976c31-a091-451c-9d73-18f2e900f1d3)

=== Resuming from Checkpoint ===
RunID: 31976c31-a091-451c-9d73-18f2e900f1d3
Checkpoint: financial_analysis

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


Recovery Statistics:
- Checkpoint recovery successful
- State preserved across failure

R&D Project: Hypersonic Flight Control Systems
  Classification: TOP SECRET
  Components: 6
  Estimated Cost: $385,000
  Risk Level: HIGH
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $385,000
- Total processing time: 14.0s
- Average time per request: 14.0s
```

---

## Test 8: Checkpoint Recovery - Security Stage Failure

Purpose: Validate checkpoint/recovery at security review stage

```sh
go run ./examples/darpa-procurement/. -fail-at security -requests 1
```

Expected Results:
- First request: Execution through financial and initial legal review
- Simulated failure at security review stage
- Checkpoint saved
- Automatic recovery
- Workflow resumes from security review
- Request completes successfully

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
NOTE: Failure injection enabled at stage: security

→ Drafting procurement request: Neural Interface Brain-Computer Systems
   Neural Interface Brain-Computer Systems - SECRET, Human Performance Enhancement

→ Analyzing procurement costs...
   $385000 | Risk: MEDIUM | Route: Standard Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget significantly exceeds program allocations, high financial risk identified
  Optimization: 8000 potential savings

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED

✗ SIMULATED FAILURE at security stage
Checkpoint saved: financial_analysis (runID: 9d846d73-db93-449f-811a-d29225a6591c)

=== Resuming from Checkpoint ===
RunID: 9d846d73-db93-449f-811a-d29225a6591c
Checkpoint: financial_analysis

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


Recovery Statistics:
- Checkpoint recovery successful
- State preserved across failure

R&D Project: Neural Interface Brain-Computer Systems
  Classification: SECRET
  Components: 6
  Estimated Cost: $385,000
  Risk Level: MEDIUM
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $385,000
- Total processing time: 14.4s
- Average time per request: 14.4s
```

---

## Test 9: Custom Reviewer Count

Purpose: Validate configurable reviewer parallelism

```sh
go run ./examples/darpa-procurement/. -reviewers 3 -requests 1
```

Expected Results:
- 1 project processed
- 3 legal reviewers for consensus (maximum parallelism)
- Legal review consensus based on 3 parallel reviews
- Performance impact from additional reviewer

Result:

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===
→ Drafting procurement request: Autonomous Underwater Vehicle Swarms
   Autonomous Underwater Vehicle Swarms - SECRET, Distributed Sensing Network

→ Analyzing procurement costs...
   $385000 | Risk: MEDIUM | Route: Standard Legal Review

→ Validating procurement request...
   VALIDATED

→ Conducting financial analysis (parallel: budget validation + cost optimization)...
  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 85000 potential savings

→ Conducting compliance review (parallel: 3 legal reviewers + security officer)...
  Legal Review Consensus: APPROVED
  Security Review: APPROVED
→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED


R&D Project: Autonomous Underwater Vehicle Swarms
  Classification: SECRET
  Components: 8
  Estimated Cost: $385,000
  Risk Level: MEDIUM
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $385,000
- Total processing time: 13.2s
- Average time per request: 13.2s
```

---

## Test 10: Verbose Observability Mode

Purpose: Validate detailed execution logging with SlogObserver

```sh
go run ./examples/darpa-procurement/. -v -requests 1
```

Expected Results:
- Standard workflow execution
- JSON-formatted log entries for all workflow events
- Detailed observability data including:
  - Node transitions
  - State updates
  - Agent invocations
  - Decision points
- Log entries written to stdout in addition to workflow output

Result:

**Note**: Verbose mode output is truncated for readability. JSON log entries have been condensed to show key event types.

```
DARPA Research Procurement Simulation
Initializing agents (config: examples/darpa-procurement/config.gemma.json, max_tokens: default)...

=== Processing Request 1/1 ===

[Observability Events - Sample]
{"level":"INFO","type":"graph.start","source":"darpa-procurement","data":{"entry_point":"entry","exit_points":2}}
{"level":"INFO","type":"node.start","data":{"iteration":1,"node":"entry"}}
{"level":"INFO","type":"node.complete","data":{"error":false,"node":"entry"}}
{"level":"INFO","type":"checkpoint.save","data":{"node":"entry"}}

→ Drafting procurement request: Synthetic Organism Detection Biosensors
   Synthetic Organism Detection Biosensors - SECRET, Distributed Threat Monitoring

{"level":"INFO","type":"node.complete","data":{"node":"request_drafting"}}
{"level":"INFO","type":"checkpoint.save","data":{"node":"request_drafting"}}

→ Analyzing procurement costs...
   $385000 | Risk: MEDIUM | Route: Standard Legal Review

{"level":"INFO","type":"node.complete","data":{"node":"cost_analysis"}}
{"level":"INFO","type":"checkpoint.save","data":{"node":"cost_analysis"}}

→ Validating procurement request...
   VALIDATED

{"level":"INFO","type":"node.complete","data":{"node":"procurement_validation"}}
{"level":"INFO","type":"checkpoint.save","data":{"node":"procurement_validation"}}

→ Conducting financial analysis (parallel: budget validation + cost optimization)...

{"level":"INFO","type":"parallel.start","source":"workflows.ProcessParallel","data":{"item_count":2,"worker_count":2}}
{"level":"INFO","type":"worker.start","data":{"item_index":0,"worker_id":1}}
{"level":"INFO","type":"worker.start","data":{"item_index":1,"worker_id":0}}
{"level":"INFO","type":"worker.complete","data":{"error":false,"worker_id":0}}
{"level":"INFO","type":"worker.complete","data":{"error":false,"worker_id":1}}
{"level":"INFO","type":"parallel.complete","data":{"error":false,"items_processed":2}}

  Budget: Budget aligns with program allocations, moderate financial risk identified
  Optimization: 85000 potential savings

{"level":"INFO","type":"node.complete","data":{"node":"financial_analysis"}}
{"level":"INFO","type":"checkpoint.save","data":{"node":"financial_analysis"}}

→ Conducting compliance review (parallel: 2 legal reviewers + security officer)...

{"level":"INFO","type":"parallel.start","data":{"item_count":2,"worker_count":2}}
{"level":"INFO","type":"worker.start","data":{"item_index":0,"worker_id":1}}
{"level":"INFO","type":"worker.start","data":{"item_index":1,"worker_id":0}}
{"level":"INFO","type":"worker.complete","data":{"error":false,"worker_id":0}}
{"level":"INFO","type":"worker.complete","data":{"error":false,"worker_id":1}}
{"level":"INFO","type":"parallel.complete","data":{"items_processed":2}}

  Legal Review Consensus: APPROVED
  Security Review: APPROVED

→ Routing to Deputy Director for final approval (route: full-security-review)...
  Decision: APPROVED

{"level":"INFO","type":"node.complete","data":{"node":"approval_routing"}}
{"level":"INFO","type":"edge.evaluate","data":{"from":"approval_routing","has_predicate":true,"to":"approved"}}
{"level":"INFO","type":"node.complete","data":{"node":"approved"}}
{"level":"INFO","type":"graph.complete","data":{"exit_point":"approved","iterations":7,"path_length":7}}

R&D Project: Synthetic Organism Detection Biosensors
  Classification: SECRET
  Components: 6
  Estimated Cost: $385,000
  Risk Level: MEDIUM
  Legal Review: APPROVED
  Security Review: APPROVED

Final Decision:
  ✓ APPROVED by Deputy Director
  Award ID: PR-2024-001

Summary:
- Requests processed: 1
- Approved: 1 (PR-2024-001)
- Rejected: 0
- Total budget allocated: $385,000
- Total processing time: 11.8s
- Average time per request: 11.8s
```

**Key Observability Events Captured:**
- Graph lifecycle: start → complete
- Node transitions: entry → drafting → cost analysis → validation → financial → routing → approved
- Checkpoint saves: After each node completion
- Parallel execution: Worker start/complete events for financial analysis and legal review
- Edge evaluation: Predicate evaluation and transitions

---

## Validation Checklist

After running all tests, verify:

- **Conditional Routing**: Projects route to different approval levels based on cost/risk
- **Parallel Execution**: Budget+optimization and legal+security run concurrently
- **State Management**: All project data flows correctly through stages
- **JSON Parsing**: No parsing failures with gemma model
- **Checkpoint Recovery**: All three failure injection points recover successfully
- **Mode Variations**: Fast/balanced/thorough modes adjust reviewer count appropriately
- **Emergency Bypass**: Legal review can be skipped when needed
- **Observability**: Verbose mode provides detailed execution insights
- **Scalability**: Handles 1-5 requests without issues
- **Performance**: Reasonable execution times across different configurations

---

## Expected Routing Patterns

Based on workflow logic, watch for these patterns:

- **Low Cost** (< $50,000 estimated): → Program Director
- **Standard Review** (MEDIUM risk): → Standard legal path
- **Expedited Review** (HIGH risk): → Enhanced review process
- **Full Security Review** (TOP SECRET + HIGH risk): → Deputy Director approval
- **Budget Rejection**: Projects may require revision if budget significantly exceeds allocations
