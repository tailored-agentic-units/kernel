package kernel

import "errors"

// ErrMaxIterations is returned by Run when the loop exhausts its iteration
// budget without the agent producing a final response.
var ErrMaxIterations = errors.New("max iterations reached")
