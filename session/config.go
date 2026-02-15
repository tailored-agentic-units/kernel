package session

// Config holds session initialization parameters. Currently empty — serves as
// an extension point for future session backends.
type Config struct{}

// DefaultConfig returns the default session configuration.
func DefaultConfig() Config {
	return Config{}
}

// Merge applies non-zero values from source into c.
func (c *Config) Merge(source *Config) {}

// New creates a Session from configuration. Currently returns an in-memory session.
func New(cfg *Config) (Session, error) {
	return NewMemorySession(), nil
}
