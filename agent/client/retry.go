package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/tailored-agentic-units/kernel/core/config"
)

// HTTPStatusError represents an HTTP error with status code and response body.
// Used to distinguish HTTP errors from other types of errors for retry logic.
type HTTPStatusError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *HTTPStatusError) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("HTTP %d: %s - %s", e.StatusCode, e.Status, string(e.Body))
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
}

// isRetryableError determines if an error should trigger a retry attempt.
// Returns true for transient failures that might succeed on retry:
// - HTTP 429 (rate limit), 502 (bad gateway), 503 (service unavailable), 504 (gateway timeout)
// - Network operation errors (connection failures, timeouts)
// - Temporary DNS errors
//
// Returns false for:
// - Context cancellation/deadline errors (user-initiated or timeout)
// - HTTP client errors (4xx except 429)
// - Other permanent failures
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Never retry context errors (cancelled by user or timed out)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for HTTP status errors - retry on transient server issues
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 429 || // Rate limit
			httpErr.StatusCode == 502 || // Bad gateway
			httpErr.StatusCode == 503 || // Service unavailable
			httpErr.StatusCode == 504 // Gateway timeout
	}

	// Check for network operation errors (connection refused, timeout, etc.)
	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return true
	}

	// Check for DNS errors - retry if temporary or timeout
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary() || dnsErr.Timeout()
	}

	// Check for URL errors - unwrap and check underlying error
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return isRetryableError(urlErr.Err)
	}

	// Default to not retrying unknown errors
	return false
}

// calculateBackoff computes exponential backoff duration with optional jitter.
// Uses exponential growth: initialBackoff * (2^attempt).
// Applies ±25% jitter if enabled to prevent thundering herd.
// Caps result at maxBackoff to prevent excessive delays.
func calculateBackoff(attempt int, cfg config.RetryConfig) time.Duration {
	// Cap attempt to prevent overflow
	maxAttempt := min(attempt, 10)

	// Calculate exponential backoff: initialBackoff * (2^attempt)
	delay := time.Duration(cfg.InitialBackoff) * time.Duration(1<<uint(maxAttempt))

	// Apply jitter (±25% randomization) if enabled
	if cfg.Jitter {
		jitterRange := delay / 4
		jitter := time.Duration(rand.Int63n(int64(jitterRange)*2)) - jitterRange
		delay += jitter
	}

	// Cap at MaxBackoff
	return min(delay, time.Duration(cfg.MaxBackoff))
}

// doWithRetry executes an operation with retry logic.
// Retries only on transient failures (determined by isRetryableError).
// Uses exponential backoff with optional jitter between retries.
// Respects context cancellation during operation and backoff.
//
// Returns the successful result or the last error encountered.
func doWithRetry[T any](
	ctx context.Context,
	cfg config.RetryConfig,
	operation func(context.Context) (T, error),
) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation before retry
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("operation cancelled: %w", err)
		}

		// Execute operation
		result, lastErr = operation(ctx)
		if lastErr == nil {
			return result, nil
		}

		// Check if error is retryable
		if !isRetryableError(lastErr) {
			return result, lastErr
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxRetries {
			delay := calculateBackoff(attempt, cfg)

			select {
			case <-time.After(delay):
				// Continue to next retry
			case <-ctx.Done():
				return result, fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	return result, fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}
