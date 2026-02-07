package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func parseJSON[T any](content string) (T, error) {
	var result T

	err1 := json.Unmarshal([]byte(content), &result)
	if err1 == nil {
		return result, nil
	}

	extracted := extractJSONFromMarkdown(content)
	err2 := json.Unmarshal([]byte(extracted), &result)
	if err2 == nil {
		return result, nil
	}

	return result, fmt.Errorf(
		"failed to parse JSON response: %w",
		errors.Join(
			fmt.Errorf("direct parse: %w", err1),
			fmt.Errorf("markdown extraction: %w", err2),
		),
	)
}

func extractJSONFromMarkdown(content string) string {
	re := regexp.MustCompile(`(?s)` + "`" + `{3}(?:json)?\s*(.+?)\s*` + "`" + `{3}`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return content
}
