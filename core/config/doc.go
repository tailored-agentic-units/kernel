// Package config provides configuration management for the go-agents library.
// It defines structures for agent, model, provider, and transport configuration
// with support for human-readable duration strings and JSON serialization.
//
// Configuration files use hierarchical JSON structure with transport-based
// organization. Example:
//
//	{
//	  "name": "my-agent",
//	  "system_prompt": "You are a helpful assistant",
//	  "transport": {
//	    "provider": {
//	      "name": "ollama",
//	      "base_url": "http://localhost:11434",
//	      "model": {
//	        "name": "llama3.2:3b",
//	        "capabilities": {
//	          "chat": {"format": "openai-chat", "options": {...}}
//	        }
//	      }
//	    },
//	    "timeout": "24s",
//	    "connection_pool_size": 10
//	  }
//	}
//
// Duration values support human-readable strings ("24s", "1m", "2h") or
// numeric nanoseconds for programmatic configuration.
package config
