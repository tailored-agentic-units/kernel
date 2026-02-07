package workflows

// ProgressFunc provides visibility into chain execution progress.
// Called after each successful step completion. Not called before the first
// step or when errors occur. Useful for progress bars, logging, or monitoring.
//
// Parameters:
//
//   - completed: Number of steps completed so far (1-indexed)
//   - total: Total number of steps in the chain
//   - state: Current accumulated state snapshot
//
// Example:
//
//	progress := func(completed, total int, state State) {
//	    fmt.Printf("Progress: %d/%d\n", completed, total)
//	}
type ProgressFunc[TContext any] func(
	completed int,
	total int,
	state TContext,
)
