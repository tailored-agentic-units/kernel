package response

import (
	"fmt"

	"github.com/tailored-agentic-units/kernel/core/protocol"
)

// Parse parses a response based on protocol type.
// Routes to the appropriate protocol-specific parser and returns the parsed result.
// Returns an error if the protocol is unsupported or parsing fails.
func Parse(p protocol.Protocol, body []byte) (any, error) {
	switch p {
	case protocol.Chat:
		return ParseChat(body)
	case protocol.Vision:
		return ParseVision(body)
	case protocol.Tools:
		return ParseTools(body)
	case protocol.Embeddings:
		return ParseEmbeddings(body)
	case protocol.Audio:
		return ParseAudio(body)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", p)
	}
}

// ParseStreamChunk parses a streaming chunk based on protocol type.
// Routes to the appropriate protocol-specific streaming parser.
// Returns an error if the protocol doesn't support streaming or parsing fails.
func ParseStreamChunk(p protocol.Protocol, data []byte) (*StreamingChunk, error) {
	switch p {
	case protocol.Chat:
		return ParseChatStreamChunk(data)
	case protocol.Vision:
		return ParseVisionStreamChunk(data)
	case protocol.Tools:
		return ParseToolsStreamChunk(data)
	case protocol.Embeddings, protocol.Audio:
		return nil, fmt.Errorf("protocol %s does not support streaming", p)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", p)
	}
}
