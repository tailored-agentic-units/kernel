# Phase 1: Hub & Messaging Example - ISS Maintenance EVA

This example demonstrates the core hub and messaging capabilities of tau-orchestrate through an International Space Station (ISS) maintenance EVA (Extravehicular Activity) scenario.

## Overview

The example simulates a coordinated ISS maintenance operation where:
- **EVA crew** operates outside the station performing cooling system repairs
- **ISS internal crew** supports the operation from inside the station
- **Mission commander** coordinates both teams across both operational hubs

This scenario demonstrates all Phase 1 orchestration patterns:
1. **Agent-to-Agent Communication** - Direct peer communication
2. **Broadcast Communication** - One-to-many within a hub
3. **Pub/Sub Communication** - Topic-based message distribution
4. **Cross-Hub Communication** - Coordination across multiple hubs

## Architecture

### Agents (4 total)

| Agent | Model | Role | Hub Membership |
|-------|-------|------|----------------|
| `eva-specialist-1` | llama3.2:3b | Primary spacewalker | eva-hub |
| `eva-specialist-2` | llama3.2:3b | Secondary spacewalker | eva-hub |
| `mission-commander` | gemma3:4b | Mission coordinator | eva-hub, iss-hub |
| `flight-engineer` | llama3.2:3b | Internal support | iss-hub |

### Hubs (2 total)

```
┌─────────────────────────────────────────────────────────────────┐
│                          EVA Hub                                │
│  (Crew operating outside the station)                           │
│                                                                 │
│  Agents:                                                        │
│  • eva-specialist-1 (subscribed to: equipment)                  │
│  • eva-specialist-2 (subscribed to: safety)                     │
│  • mission-commander (subscribed to: equipment, safety)         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                          ISS Hub                                │
│  (Crew operating inside the station)                            │
│                                                                 │
│  Agents:                                                        │
│  • flight-engineer                                              │
│  • mission-commander                                            │
└─────────────────────────────────────────────────────────────────┘
```

### Topic Subscriptions

```
Topics in EVA Hub:
  equipment → eva-specialist-1, mission-commander
  safety    → eva-specialist-2, mission-commander
```

## Communication Patterns

### 1. Agent-to-Agent Communication

**Scenario:** EVA specialist requests tool from teammate

```
┌────────────────────────────────────────────────┐
│              EVA Hub                           │
│                                                │
│  eva-specialist-1 ──────────────────┐          │
│   "Need torque wrench"              │          │
│                                     │          │
│                                     ▼          │
│                          eva-specialist-2      │
│                           "Retrieving tool"    │
│                                                │
└────────────────────────────────────────────────┘
```

**Code:**
```go
evaHub.Send(ctx, evaSpec1.ID(), evaSpec2.ID(),
    "I need the torque wrench, can you retrieve it from the tool bag?")
```

**Message Flow:**
1. `eva-specialist-1` sends direct message to `eva-specialist-2`
2. Hub routes message only to the specified recipient
3. `eva-specialist-2` processes message and responds

---

### 2. Broadcast Communication

**Scenario:** Commander announces orbital sunset to entire EVA crew

```
┌─────────────────────────────────────────────────────────────┐
│                        EVA Hub                              │
│                                                             │
│                  mission-commander                          │
│                   "Orbital sunset                           │
│                    in 20 minutes"                           │
│                         │                                   │
│         ┌───────────────┴───────────────┐                   │
│         ▼                               ▼                   │
│  eva-specialist-1              eva-specialist-2             │
│   "Acknowledged"                "Roger that"                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Code:**
```go
evaHub.Broadcast(ctx, commander.ID(),
    "Orbital sunset in 20 minutes, prioritize the cooling line connection")
```

**Message Flow:**
1. `mission-commander` broadcasts to all agents in eva-hub
2. Hub delivers message to all registered agents (excluding sender)
3. Both `eva-specialist-1` and `eva-specialist-2` receive and respond

---

### 3. Pub/Sub Communication

**Scenario:** Commander publishes equipment update to subscribed agents

```
┌──────────────────────────────────────────────────────────────┐
│                        EVA Hub                               │
│                                                              │
│                  mission-commander                           │
│                  Publish("equipment",                        │
│                   "Thermal blanket available")               │
│                         │                                    │
│                         │ (sender filtered out)              │
│                         │                                    │
│                         ▼                                    │
│                  eva-specialist-1                            │
│                  (subscribed to "equipment")                 │
│                  "Noted"                                     │
│                                                              │
│  eva-specialist-2 (subscribed to "safety")                   │
│  does NOT receive - not subscribed to "equipment"            │
│                                                              │
│  mission-commander (subscribed to "equipment")               │
│  does NOT receive - sender is filtered out                   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

**Code:**
```go
// Subscriptions established during initialization
evaHub.Subscribe(evaSpec1.ID(), "equipment")
evaHub.Subscribe(evaSpec2.ID(), "safety")
evaHub.Subscribe(commander.ID(), "equipment")
evaHub.Subscribe(commander.ID(), "safety")

// Publish to topic
evaHub.Publish(ctx, commander.ID(), "equipment",
    "Spare thermal blanket available in airlock if needed")
```

**Message Flow:**
1. `mission-commander` publishes to "equipment" topic
2. Hub filters out sender and delivers only to other subscribers
3. `eva-specialist-1` receives (subscribed to "equipment")
4. `eva-specialist-2` does NOT receive (subscribed to "safety" only)
5. `mission-commander` does NOT receive (sender is filtered out)

---

### 4. Cross-Hub Communication

**Scenario:** EVA crew reports completion, commander relays to internal crew

```
┌─────────────────────────────┐      ┌─────────────────────────────┐
│         EVA Hub             │      │         ISS Hub             │
│                             │      │                             │
│  eva-specialist-1           │      │                             │
│   "Cooling line complete"   │      │                             │
│         │                   │      │                             │
│         ▼                   │      │                             │
│  mission-commander ─────────┼──────┼────► mission-commander      │
│  (eva-hub context)          │      │     (iss-hub context)       │
│                             │      │              │              │
│                             │      │              ▼              │
│                             │      │       flight-engineer       │
│                             │      │      "Starting pressurize"  │
│                             │      │                             │
└─────────────────────────────┘      └─────────────────────────────┘
```

**Code:**
```go
// EVA specialist reports to commander in eva-hub
evaHub.Send(ctx, evaSpec1.ID(), commander.ID(),
    "Cooling line connection complete, ready to pressurize system")

// Commander relays to flight engineer in iss-hub
issHub.Send(ctx, commander.ID(), flightEng.ID(),
    "EVA crew ready for cooling system pressurization")
```

**Message Flow:**
1. `eva-specialist-1` sends to `mission-commander` in eva-hub
2. `mission-commander` receives in eva-hub context
3. `mission-commander` (registered in both hubs) sends in iss-hub
4. `flight-engineer` receives in iss-hub and responds

---

## Prerequisites

### 1. Ollama with Models

The example requires [Ollama](https://ollama.ai) running with two models:
- `llama3.2:3b` - For EVA specialists and flight engineer
- `gemma3:4b` - For mission commander

**Quick Start with Docker Compose:**

```bash
# From repository root
docker-compose up -d

# Verify models are pulled
docker exec tau-orchestrate-ollama ollama list
```

The Docker Compose configuration automatically:
- Starts Ollama server on port 11434
- Pulls both required models
- Configures GPU access (if available)

### 2. Environment

- Go 1.23 or later
- Docker (for Ollama container)
- NVIDIA GPU (optional, improves performance)

## Running the Example

### Option 1: Using go run

```bash
# From repository root
go run examples/phase-01-hubs/main.go
```

### Option 2: Build and run

```bash
# Build
go build -o bin/phase-01-hubs examples/phase-01-hubs/main.go

# Run
./bin/phase-01-hubs
```

## Expected Output

```
=== ISS Maintenance EVA - Agent Orchestration Demo ===

1. Loading agent configurations...
  ✓ Created eva-specialist-1 (llama)
  ✓ Created eva-specialist-2 (llama)
  ✓ Created mission-commander (gemma)
  ✓ Created flight-engineer (llama)

2. Creating hubs...
  ✓ Created eva-hub (EVA crew)
  ✓ Created iss-hub (ISS internal operations)

3. Registering agents with hubs...
  ✓ Registered all agents with hubs

4. Subscribing agents to topics...
  ✓ eva-specialist-1 subscribed to 'equipment'
  ✓ eva-specialist-2 subscribed to 'safety'
  ✓ mission-commander subscribed to 'equipment' and 'safety'

5. Agent-to-Agent Communication
   eva-specialist-1 → eva-specialist-2
   Message: I need the torque wrench, can you retrieve it from the tool bag?
   eva-specialist-2: Roger, retrieving torque wrench from tool bag now.

6. Broadcast Communication
   mission-commander → all EVA crew
   Message: Orbital sunset in 20 minutes, prioritize the cooling line connection
   eva-specialist-1: Acknowledged, prioritizing cooling line connection before sunset.
   eva-specialist-2: Copy that, focusing on cooling line completion.

7. Pub/Sub Communication
   mission-commander publishes to topic 'equipment'
   Message: Spare thermal blanket available in airlock if needed
   eva-specialist-1: Noted, thermal blanket location confirmed.

8. Cross-Hub Communication
   eva-specialist-1 → mission-commander (eva-hub) → flight-engineer (iss-hub)
   Message: Cooling line connection complete, ready to pressurize system
   mission-commander (eva-hub): Copy that, relaying to internal crew for pressurization.
   flight-engineer: Initiating cooling system pressurization sequence now.

9. EVA Operation Metrics
   EVA Hub:
     - Local Agents: 3
     - Messages Sent: 4
     - Messages Received: 6
   ISS Hub:
     - Local Agents: 2
     - Messages Sent: 1
     - Messages Received: 1

=== EVA Operation Complete ===
```

## Configuration

### Agent Configurations

- **`config.llama.json`** - Base configuration for llama3.2:3b agents
- **`config.gemma.json`** - Base configuration for gemma3:4b agent

Both configurations set `max_tokens: 150` to keep responses concise for the demonstration.

### Customization

**Adjust response length:**

Edit the `max_tokens` value in the config files:

```json
{
  "capabilities": {
    "chat": {
      "options": {
        "max_tokens": 150  // Increase for longer responses
      }
    }
  }
}
```

**Modify system prompts:**

System prompts are set in `main.go` with operational context. Edit the `SystemPrompt` fields in the agent configurations to change agent behavior and context awareness.

## Key Concepts Demonstrated

### Hub Registration
Agents can be registered with multiple hubs, enabling cross-hub coordination:
```go
evaHub.RegisterAgent(commander, commanderHandler)
issHub.RegisterAgent(commander, commanderHandler)
```

### Topic-Based Messaging
Agents subscribe to topics of interest, receiving only relevant messages:
```go
evaHub.Subscribe(evaSpec1.ID(), "equipment")
evaHub.Publish(ctx, commander.ID(), "equipment", message)
```

**Important:** The sender is automatically filtered out and does NOT receive their own published messages.

### Message Context
Handlers receive context about which hub delivered the message:
```go
func handler(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) {
    fmt.Printf("Received in %s hub\n", msgCtx.HubName)
}
```

### Metrics and Observability
Hubs track communication metrics:
```go
metrics := evaHub.Metrics()
fmt.Printf("Messages sent: %d\n", metrics.MessagesSent)
```

## What's Next

This example demonstrates Phase 1 capabilities (Hub & Messaging). Future phases will add:

- **Phase 2**: State management with LangGraph-inspired state graphs
- **Phase 3**: Workflow patterns (sequential chains, parallel execution, conditional routing)
- **Phase 4**: Advanced observability (trace capture, decision logging, confidence scoring)

## Troubleshooting

**Agents not responding:**
- Verify Ollama is running: `curl http://localhost:11434/api/tags`
- Check models are available: `ollama list`
- Ensure config files point to correct Ollama URL

**Timeout errors:**
- Models running on GPU may respond faster
- Increase timeout in code if running on CPU
- Reduce `max_tokens` for faster responses

**Connection refused:**
- Verify Ollama container is running: `docker ps`
- Check port 11434 is not blocked by firewall
- Ensure Docker network is properly configured
