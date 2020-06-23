package uid

import (
	"sync/atomic"
)

// LID represents a process-wide unique ID.
type LID int32

// NewID generates a new, process-wide unique ID.
func NewLID() LID {
	return LID(atomic.AddInt32(&Next, 1))
}
