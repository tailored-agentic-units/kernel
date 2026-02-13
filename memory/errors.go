package memory

import "errors"

// Sentinel errors for store operations.
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrLoadFailed  = errors.New("load failed")
	ErrSaveFailed  = errors.New("save failed")
)
