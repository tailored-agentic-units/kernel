package config

import "time"

// ClientConfig defines the configuration for the HTTP client layer.
// It includes timeout settings, retry behavior, and connection pooling parameters.
type ClientConfig struct {
	Timeout            Duration    `json:"timeout"`
	Retry              RetryConfig `json:"retry"`
	ConnectionPoolSize int         `json:"connection_pool_size"`
	ConnectionTimeout  Duration    `json:"connection_timeout"`
}

// RetryConfig configures retry behavior for failed requests.
// Implements exponential backoff with jitter for transient failures.
type RetryConfig struct {
	MaxRetries        int      `json:"max_retries"`
	InitialBackoff    Duration `json:"initial_backoff"`
	MaxBackoff        Duration `json:"max_backoff"`
	BackoffMultiplier float64  `json:"backoff_multiplier"`
	Jitter            bool     `json:"jitter"`
}

// DefaultClientConfig creates a ClientConfig with default values.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:            Duration(2 * time.Minute),
		Retry:              DefaultRetryConfig(),
		ConnectionPoolSize: 10,
		ConnectionTimeout:  Duration(30 * time.Second),
	}
}

// DefaultRetryConfig creates a RetryConfig with default values.
// Retries up to 3 times with exponential backoff starting at 1s, capped at 30s.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    Duration(time.Second),
		MaxBackoff:        Duration(30 * time.Second),
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// Merge combines the source ClientConfig into this ClientConfig.
// Positive values from source override the current values. Zero values are ignored.
func (c *ClientConfig) Merge(source *ClientConfig) {
	if source.Timeout > 0 {
		c.Timeout = source.Timeout
	}

	if source.Retry.MaxRetries > 0 {
		c.Retry.MaxRetries = source.Retry.MaxRetries
	}

	if source.Retry.InitialBackoff > 0 {
		c.Retry.InitialBackoff = source.Retry.InitialBackoff
	}

	if source.Retry.MaxBackoff > 0 {
		c.Retry.MaxBackoff = source.Retry.MaxBackoff
	}

	if source.Retry.BackoffMultiplier > 0 {
		c.Retry.BackoffMultiplier = source.Retry.BackoffMultiplier
	}

	// Jitter is boolean, always take source value if explicitly set
	c.Retry.Jitter = source.Retry.Jitter

	if source.ConnectionPoolSize > 0 {
		c.ConnectionPoolSize = source.ConnectionPoolSize
	}

	if source.ConnectionTimeout > 0 {
		c.ConnectionTimeout = source.ConnectionTimeout
	}
}
