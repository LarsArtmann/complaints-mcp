package domain

import (
	"fmt"
	"sync"
	"time"
)

// ThreadSafeComplaint provides thread-safe operations on a Complaint entity
// This separates thread safety concerns from the pure domain entity
type ThreadSafeComplaint struct {
	complaint *Complaint
	mu        sync.RWMutex
}

// NewThreadSafeComplaint creates a thread-safe wrapper for a complaint
func NewThreadSafeComplaint(complaint *Complaint) *ThreadSafeComplaint {
	return &ThreadSafeComplaint{
		complaint: complaint,
	}
}

// Complaint returns a copy of the underlying complaint
// Returns a copy to prevent external mutation
func (tsc *ThreadSafeComplaint) Complaint() Complaint {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()
	return *tsc.complaint
}

// Resolve marks the complaint as resolved - thread-safe
func (tsc *ThreadSafeComplaint) Resolve(resolvedBy string) error {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	
	// Check if already resolved for better error message
	if tsc.complaint.IsResolved() {
		var ts string
		if tsc.complaint.ResolvedAt != nil {
			ts = tsc.complaint.ResolvedAt.Format(time.RFC3339)
		} else {
			ts = "<unknown>"
		}
		return fmt.Errorf("complaint already resolved by %s at %s", 
			tsc.complaint.ResolvedBy, ts)
	}
	
	// Delegate to domain method to enforce state machine rules
	return tsc.complaint.Resolve(resolvedBy)
}

// TransitionState performs a state transition - thread-safe
func (tsc *ThreadSafeComplaint) TransitionState(newState ResolutionState, resolvedBy string) error {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	
	// Check if transition is allowed
	if !tsc.complaint.ResolutionState.CanTransitionTo(newState) {
		return fmt.Errorf("invalid state transition from %s to %s", 
			tsc.complaint.ResolutionState, newState)
	}
	
	// Handle resolution state
	if newState.IsResolved() && resolvedBy == "" {
		return fmt.Errorf("resolver name required for resolved state")
	}
	
	// Perform state transition
	tsc.complaint.ResolutionState = newState
	
	if newState.IsResolved() {
		now := time.Now()
		tsc.complaint.ResolvedAt = &now
		tsc.complaint.ResolvedBy = resolvedBy
	}
	
	return nil
}

// IsResolved checks if the complaint is resolved - thread-safe
func (tsc *ThreadSafeComplaint) IsResolved() bool {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()
	return tsc.complaint.IsResolved()
}

// ResolutionState returns the current resolution state - thread-safe
func (tsc *ThreadSafeComplaint) ResolutionState() ResolutionState {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()
	return tsc.complaint.ResolutionState
}