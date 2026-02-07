package mock_test

import (
	"context"
	"testing"

	"github.com/tailored-agentic-units/kernel/agent/mock"
	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/agent/providers"
)

func TestNewMockProvider(t *testing.T) {
	provider := mock.NewMockProvider()

	if provider == nil {
		t.Fatal("NewMockProvider returned nil")
	}
}

func TestMockProvider_Name(t *testing.T) {
	provider := mock.NewMockProvider()

	if provider.Name() != "mock-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "mock-provider")
	}
}

func TestMockProvider_Endpoint(t *testing.T) {
	customEndpoints := map[protocol.Protocol]string{
		protocol.Chat:   "/chat",
		protocol.Vision: "/vision",
	}

	provider := mock.NewMockProvider(
		mock.WithBaseURL("https://custom.api"),
		mock.WithEndpointMapping(customEndpoints),
	)

	endpoint, err := provider.Endpoint(protocol.Chat)

	if err != nil {
		t.Fatalf("Endpoint failed: %v", err)
	}

	if endpoint != "https://custom.api/chat" {
		t.Errorf("got endpoint %q, want %q", endpoint, "https://custom.api/chat")
	}
}

func TestMockProvider_PrepareRequest(t *testing.T) {
	expectedRequest := &providers.Request{
		URL:     "https://test.api/chat",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    []byte(`{"test":"data"}`),
	}

	provider := mock.NewMockProvider(
		mock.WithPrepareResponse(expectedRequest, nil),
	)

	request, err := provider.PrepareRequest(context.Background(), protocol.Chat, []byte(`{}`), nil)

	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request != expectedRequest {
		t.Error("returned different request than configured")
	}
}

func TestMockProvider_Marshal(t *testing.T) {
	expectedBody := []byte(`{"model":"test-model"}`)

	provider := mock.NewMockProvider(
		mock.WithMarshalResponse(expectedBody, nil),
	)

	body, err := provider.Marshal(protocol.Chat, nil)

	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(body) != string(expectedBody) {
		t.Errorf("got body %q, want %q", string(body), string(expectedBody))
	}
}

func TestMockProvider_BaseURL(t *testing.T) {
	provider := mock.NewMockProvider(
		mock.WithBaseURL("https://custom.api"),
	)

	if provider.BaseURL() != "https://custom.api" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://custom.api")
	}
}
