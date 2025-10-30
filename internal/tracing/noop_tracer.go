package tracing

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
)

// Tracer defines the interface for distributed tracing
type Tracer interface {
	Start(ctx context.Context, operationName string) (context.Context, Span)
}

// Span defines the interface for a trace span
type Span interface {
	End()
}

// NoOpTracer provides a no-op tracer implementation
type NoOpTracer struct{}

// NewNoOpTracer creates a new no-op tracer
func NewNoOpTracer() Tracer {
	return &NoOpTracer{}
}

// Start begins a new trace span (no-op implementation)
func (n *NoOpTracer) Start(ctx context.Context, operationName string) (context.Context, Span) {
	logger := log.FromContext(ctx)
	logger.Debug("Starting trace span", "tracer", "noop", "operation", operationName)
	
	span := &NoOpSpan{}
	return context.WithValue(ctx, "current_span", span), span
}

// NoOpSpan represents a no-op trace span
type NoOpSpan struct{}

// End completes the trace span (no-op implementation)
func (n *NoOpSpan) End() {
	logger := log.Default()
	logger.Debug("Trace span ended", "tracer", "noop")
}