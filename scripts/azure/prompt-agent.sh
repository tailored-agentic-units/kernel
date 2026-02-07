curl -X POST https://eastus.api.cognitive.microsoft.com/openai/deployments/o3-mini/chat/completions?api-version=2025-01-01-preview -s \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AZURE_FOUNDRY_KEY" \
  -d '{          
    "messages": [
      {              
        "role": "user",                                                                           
        "content": "How do you feel about being integrated into an agentic system written in Go?"
      }
    ],                              
    "max_completion_tokens": 100000,
    "model": "o3-mini"
  }' |
  jq -r '.choices[0].message.content'
