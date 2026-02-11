package tools

import "errors"

// Sentinel errors for the tools registry.
var (
	ErrNotFound      = errors.New("tool not found")
	ErrAlreadyExists = errors.New("tool already registered")
	ErrEmptyName     = errors.New("tool name is empty")
)
