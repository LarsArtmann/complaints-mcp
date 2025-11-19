package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Resolution represents the resolution of a complaint
// This is a value object that encapsulates both WHEN and WHO resolved a complaint
// By combining these into a single type, we eliminate the split-brain anti-pattern
// where ResolvedAt and ResolvedBy could be out of sync
type Resolution struct {
	timestamp  time.Time // When the complaint was resolved (not a pointer!)
	resolvedBy AgentName // Who resolved it (strong type, not string!)
}

// NewResolution creates a new Resolution with validation
// Both timestamp and resolver are required - no partial resolution states possible
func NewResolution(resolvedBy AgentName) (Resolution, error) {
	// Validate that resolver is not empty
	if resolvedBy.IsEmpty() {
		return Resolution{}, fmt.Errorf("resolver name cannot be empty")
	}

	// Use current time for resolution timestamp
	// This enforces that resolution time is always NOW (no backdating)
	now := time.Now()

	return Resolution{
		timestamp:  now,
		resolvedBy: resolvedBy,
	}, nil
}

// NewResolutionWithTime creates a Resolution with a specific timestamp
// Used for deserialization from storage, not for normal resolution flow
func NewResolutionWithTime(timestamp time.Time, resolvedBy AgentName) (Resolution, error) {
	// Validate that resolver is not empty
	if resolvedBy.IsEmpty() {
		return Resolution{}, fmt.Errorf("resolver name cannot be empty")
	}

	// Validate that timestamp is not zero
	if timestamp.IsZero() {
		return Resolution{}, fmt.Errorf("resolution timestamp cannot be zero")
	}

	return Resolution{
		timestamp:  timestamp,
		resolvedBy: resolvedBy,
	}, nil
}

// Timestamp returns when the complaint was resolved
func (r Resolution) Timestamp() time.Time {
	return r.timestamp
}

// ResolvedBy returns who resolved the complaint
func (r Resolution) ResolvedBy() AgentName {
	return r.resolvedBy
}

// IsZero checks if this is a zero-value Resolution
// Used internally for validation
func (r Resolution) IsZero() bool {
	return r.timestamp.IsZero() && r.resolvedBy.IsEmpty()
}

// String returns a human-readable representation
func (r Resolution) String() string {
	return fmt.Sprintf("Resolved by %s at %s",
		r.resolvedBy.String(),
		r.timestamp.Format(time.RFC3339))
}

// MarshalJSON implements json.Marshaler for JSON serialization
// Converts the value object to JSON-friendly format
func (r Resolution) MarshalJSON() ([]byte, error) {
	// Create a struct with exported fields for JSON marshaling
	aux := struct {
		Timestamp  time.Time `json:"timestamp"`
		ResolvedBy string    `json:"resolved_by"`
	}{
		Timestamp:  r.timestamp,
		ResolvedBy: r.resolvedBy.String(),
	}

	return json.Marshal(aux)
}

// UnmarshalJSON implements json.Unmarshaler for JSON deserialization
// Reconstructs the value object from JSON data with validation
func (r *Resolution) UnmarshalJSON(data []byte) error {
	// Unmarshal into a struct with exported fields
	aux := struct {
		Timestamp  time.Time `json:"timestamp"`
		ResolvedBy string    `json:"resolved_by"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal resolution: %w", err)
	}

	// Reconstruct AgentName value object from string
	resolvedBy, err := NewAgentName(aux.ResolvedBy)
	if err != nil {
		return fmt.Errorf("invalid resolver name in resolution: %w", err)
	}

	// Create Resolution with validation
	resolution, err := NewResolutionWithTime(aux.Timestamp, resolvedBy)
	if err != nil {
		return fmt.Errorf("invalid resolution data: %w", err)
	}

	*r = resolution
	return nil
}

// Equal checks if two Resolutions are equal
// Useful for testing and comparisons
func (r Resolution) Equal(other Resolution) bool {
	return r.timestamp.Equal(other.timestamp) &&
		r.resolvedBy.String() == other.resolvedBy.String()
}
