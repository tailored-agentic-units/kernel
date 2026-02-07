// Package request provides protocol request types for LLM API calls.
// Each request type encapsulates protocol-specific data, provider, and model
// configuration needed for execution.
//
// Request types implement the Request interface, providing consistent
// access to protocol, headers, marshaled body, provider, and model.
//
// Use clean constructors to create requests:
//
//	chatReq := request.NewChat(provider, model, messages, options)
//	visionReq := request.NewVision(provider, model, messages, images, visionOpts, options)
//	toolsReq := request.NewTools(provider, model, messages, tools, options)
//	embeddingsReq := request.NewEmbeddings(provider, model, input, options)
package request
