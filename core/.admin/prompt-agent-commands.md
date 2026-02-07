# Prompt Agent Tool Commands

## Chat

### Ollama

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "In 300 words or less, describe the Go programming language" \
  -stream # optional, removing will process all at once and return
```

### Azure API Key

```sh
AZURE_API_KEY=$(. scripts/azure/utilities/get-foundry-key.sh)

go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure.json \
  -token $AZURE_API_KEY \
  -prompt "In 300 words or less, describe Kubernetes" \
  -stream
```

### Azure Entra Token

```sh
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)

go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "In 300 words or less, describe OAuth and OIDC" \
  -stream
```

## Embeddings

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.embedding.json \
  -protocol embeddings \
  -prompt "The quick brown fox jumps over the lazy dog"
```

## Vision

### Local File

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.gemma.json \
  -protocol vision \
  -images ~/Pictures/wallpapers/monks-journey.jpg \
  -prompt "Provide a comprehensive description of this image" \
  -stream
```

### Web URL

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.gemma.json \
  -protocol vision \
  -images https://ollama.com/public/ollama.png \
  -prompt "Provide a comprehensive description of this image" \
  -stream
```

### Classifier

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.classifier.json \
  -protocol vision \
  -images "~/Documents/security-classification-markings-0.png,~/Documents/security-classification-markings-1.png" \
  -prompt "Generate an agent profile optimized for detecting document classifications based on the details in the provided PDF images." \
  -stream
```

```json
{
  "name": "vision-agent",
  "system_prompt": "SYSTEM PROMPT:\n\nYou are a derivative classification expert specializing in document analysis. Your primary function is to identify and assess classification markings within documents and images.\n\n**Core Tasks:**\n- Identify classification markings in headers, footers, and document body\n- Recognize portion markings and overall document classification levels\n- Determine the highest classification level present in the material\n- Parse standard US government classification markings (UNCLASSIFIED, CONFIDENTIAL, SECRET, TOP SECRET, and associated control markings)\n\n**Analysis Method:**\n- Examine all visible text for classification indicators\n- Note marking locations (header, footer, inline, margins)\n- Apply classification hierarchy rules (highest level governs overall classification)\n- Flag unclear, non-standard, or potentially conflicting markings\n\n**Output Format:**\n- Overall document classification level\n- List of identified markings and locations\n- Confidence assessment for determination\n- Notes on any anomalies or uncertainties\n\n**Constraints:**\n- Report only what is visibly marked, do not infer classification from content\n- When uncertain, state limitations clearly\n- Do not make original classification decisions","transport": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "gemma3:4b",
        "format": "openai-standard",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7
        }
      }
    },
    "timeout": 24000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 9000000000
  }
}
```
```
