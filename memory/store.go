// Package memory provides the context composition pipeline for the TAU kernel.
// It manages a hierarchical key-value namespace backed by pluggable storage,
// with session-scoped caching and progressive loading.
package memory

import "context"

// Store translates between external storage and the internal key-value namespace.
// Implementations are stateless — they perform I/O on each call without caching.
type Store interface {
	// List returns all available keys in the store.
	List(ctx context.Context) ([]string, error)
	// Load retrieves entries for the specified keys.
	Load(ctx context.Context, keys ...string) ([]Entry, error)
	// Save persists entries to storage, creating or overwriting as needed.
	Save(ctx context.Context, entries ...Entry) error
	// Delete removes entries from storage. Missing keys are ignored.
	Delete(ctx context.Context, keys ...string) error
}
