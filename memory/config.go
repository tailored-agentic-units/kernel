package memory

// Config holds memory store initialization parameters.
type Config struct {
	Path string `json:"path,omitempty"` // FileStore root directory; empty disables memory.
}

// DefaultConfig returns the default memory configuration (disabled).
func DefaultConfig() Config {
	return Config{}
}

// Merge applies non-zero values from source into c.
func (c *Config) Merge(source *Config) {
	if source.Path != "" {
		c.Path = source.Path
	}
}

// NewStore creates a Store from configuration. Returns nil Store when Path
// is empty, indicating memory is disabled.
func NewStore(cfg *Config) (Store, error) {
	if cfg.Path == "" {
		return nil, nil
	}
	return NewFileStore(cfg.Path), nil
}
