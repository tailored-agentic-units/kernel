package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tailored-agentic-units/kernel/core/config"
)

func TestClientConfig_Unmarshal(t *testing.T) {
	jsonData := `{
		"timeout": "24s",
		"retry": {
			"max_retries": 3,
			"initial_backoff": "1s",
			"max_backoff": "30s",
			"backoff_multiplier": 2.0,
			"jitter": true
		},
		"connection_pool_size": 10,
		"connection_timeout": "9s"
	}`

	var cfg config.ClientConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Timeout.ToDuration() != 24*time.Second {
		t.Errorf("got timeout %v, want 24s", cfg.Timeout.ToDuration())
	}

	if cfg.Retry.MaxRetries != 3 {
		t.Errorf("got max_retries %d, want 3", cfg.Retry.MaxRetries)
	}

	if cfg.Retry.InitialBackoff.ToDuration() != 1*time.Second {
		t.Errorf("got initial_backoff %v, want 1s", cfg.Retry.InitialBackoff.ToDuration())
	}

	if cfg.Retry.MaxBackoff.ToDuration() != 30*time.Second {
		t.Errorf("got max_backoff %v, want 30s", cfg.Retry.MaxBackoff.ToDuration())
	}

	if cfg.Retry.BackoffMultiplier != 2.0 {
		t.Errorf("got backoff_multiplier %v, want 2.0", cfg.Retry.BackoffMultiplier)
	}

	if !cfg.Retry.Jitter {
		t.Error("got jitter false, want true")
	}

	if cfg.ConnectionPoolSize != 10 {
		t.Errorf("got connection_pool_size %d, want 10", cfg.ConnectionPoolSize)
	}

	if cfg.ConnectionTimeout.ToDuration() != 9*time.Second {
		t.Errorf("got connection_timeout %v, want 9s", cfg.ConnectionTimeout.ToDuration())
	}
}

func TestClientConfig_Defaults(t *testing.T) {
	cfg := config.DefaultClientConfig()

	if cfg == nil {
		t.Fatal("DefaultClientConfig returned nil")
	}

	if cfg.Timeout.ToDuration() != 2*time.Minute {
		t.Errorf("got timeout %v, want 2m", cfg.Timeout.ToDuration())
	}

	if cfg.Retry.MaxRetries != 3 {
		t.Errorf("got max_retries %d, want 3", cfg.Retry.MaxRetries)
	}

	if cfg.Retry.InitialBackoff.ToDuration() != 1*time.Second {
		t.Errorf("got initial_backoff %v, want 1s", cfg.Retry.InitialBackoff.ToDuration())
	}

	if cfg.Retry.MaxBackoff.ToDuration() != 30*time.Second {
		t.Errorf("got max_backoff %v, want 30s", cfg.Retry.MaxBackoff.ToDuration())
	}

	if cfg.Retry.BackoffMultiplier != 2.0 {
		t.Errorf("got backoff_multiplier %v, want 2.0", cfg.Retry.BackoffMultiplier)
	}

	if !cfg.Retry.Jitter {
		t.Error("got jitter false, want true")
	}

	if cfg.ConnectionPoolSize != 10 {
		t.Errorf("got connection_pool_size %d, want 10", cfg.ConnectionPoolSize)
	}

	if cfg.ConnectionTimeout.ToDuration() != 30*time.Second {
		t.Errorf("got connection_timeout %v, want 30s", cfg.ConnectionTimeout.ToDuration())
	}
}

func TestRetryConfig_Defaults(t *testing.T) {
	cfg := config.DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("got max_retries %d, want 3", cfg.MaxRetries)
	}

	if cfg.InitialBackoff.ToDuration() != 1*time.Second {
		t.Errorf("got initial_backoff %v, want 1s", cfg.InitialBackoff.ToDuration())
	}

	if cfg.MaxBackoff.ToDuration() != 30*time.Second {
		t.Errorf("got max_backoff %v, want 30s", cfg.MaxBackoff.ToDuration())
	}

	if cfg.BackoffMultiplier != 2.0 {
		t.Errorf("got backoff_multiplier %v, want 2.0", cfg.BackoffMultiplier)
	}

	if !cfg.Jitter {
		t.Error("got jitter false, want true")
	}
}

func TestClientConfig_ConnectionPooling(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
	}{
		{
			name:     "default pool size",
			poolSize: 10,
		},
		{
			name:     "custom pool size",
			poolSize: 20,
		},
		{
			name:     "small pool size",
			poolSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ClientConfig{
				ConnectionPoolSize: tt.poolSize,
			}

			if cfg.ConnectionPoolSize != tt.poolSize {
				t.Errorf("got connection_pool_size %d, want %d", cfg.ConnectionPoolSize, tt.poolSize)
			}
		})
	}
}

func TestClientConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     *config.ClientConfig
		source   *config.ClientConfig
		expected *config.ClientConfig
	}{
		{
			name: "merge timeout",
			base: &config.ClientConfig{
				Timeout: config.Duration(1 * time.Minute),
			},
			source: &config.ClientConfig{
				Timeout: config.Duration(2 * time.Minute),
			},
			expected: &config.ClientConfig{
				Timeout: config.Duration(2 * time.Minute),
			},
		},
		{
			name: "merge retry config",
			base: &config.ClientConfig{
				Retry: config.RetryConfig{
					MaxRetries: 3,
				},
			},
			source: &config.ClientConfig{
				Retry: config.RetryConfig{
					MaxRetries: 5,
				},
			},
			expected: &config.ClientConfig{
				Retry: config.RetryConfig{
					MaxRetries: 5,
				},
			},
		},
		{
			name: "merge connection_pool_size",
			base: &config.ClientConfig{
				ConnectionPoolSize: 10,
			},
			source: &config.ClientConfig{
				ConnectionPoolSize: 20,
			},
			expected: &config.ClientConfig{
				ConnectionPoolSize: 20,
			},
		},
		{
			name: "merge connection_timeout",
			base: &config.ClientConfig{
				ConnectionTimeout: config.Duration(60 * time.Second),
			},
			source: &config.ClientConfig{
				ConnectionTimeout: config.Duration(90 * time.Second),
			},
			expected: &config.ClientConfig{
				ConnectionTimeout: config.Duration(90 * time.Second),
			},
		},
		{
			name: "zero values preserve base",
			base: &config.ClientConfig{
				Timeout:            config.Duration(1 * time.Minute),
				ConnectionPoolSize: 10,
				Retry: config.RetryConfig{
					MaxRetries: 3,
				},
			},
			source: &config.ClientConfig{
				Timeout:            0,
				ConnectionPoolSize: 0,
			},
			expected: &config.ClientConfig{
				Timeout:            config.Duration(1 * time.Minute),
				ConnectionPoolSize: 10,
				Retry: config.RetryConfig{
					MaxRetries: 3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.source)

			if tt.base.Timeout != tt.expected.Timeout {
				t.Errorf("got timeout %v, want %v", tt.base.Timeout, tt.expected.Timeout)
			}

			if tt.base.Retry.MaxRetries != tt.expected.Retry.MaxRetries {
				t.Errorf("got max_retries %d, want %d", tt.base.Retry.MaxRetries, tt.expected.Retry.MaxRetries)
			}

			if tt.base.ConnectionPoolSize != tt.expected.ConnectionPoolSize {
				t.Errorf("got connection_pool_size %d, want %d", tt.base.ConnectionPoolSize, tt.expected.ConnectionPoolSize)
			}

			if tt.base.ConnectionTimeout != tt.expected.ConnectionTimeout {
				t.Errorf("got connection_timeout %v, want %v", tt.base.ConnectionTimeout, tt.expected.ConnectionTimeout)
			}
		})
	}
}
