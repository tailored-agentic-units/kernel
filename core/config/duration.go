package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a wrapper around time.Duration that supports JSON unmarshaling
// from both string format ("24s", "1m", "2h") and numeric nanoseconds.
//
// Example JSON formats:
//
//	"timeout": "24s"           // string format
//	"timeout": 24000000000     // numeric nanoseconds
//
// When marshaling to JSON, Duration outputs string format.
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler for Duration.
// It accepts both string duration format ("24s", "1m", "2h") and numeric nanoseconds.
// Returns an error if the value cannot be parsed as either format.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		parsed, err := time.ParseDuration(str)
		if err != nil {
			return fmt.Errorf("invalid duration string %q: %w", str, err)
		}
		*d = Duration(parsed)
		return nil
	}

	var num int64
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("duration must be a string (e.g., \"2m\") or number (nanoseconds)")
	}
	*d = Duration(num)
	return nil
}

// MarshalJSON implements json.Marshaler for Duration.
// It outputs the duration in string format (e.g., "24s", "1m30s").
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// ToDuration converts Duration to time.Duration.
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}
