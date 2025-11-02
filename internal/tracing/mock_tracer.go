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
	AddEvent(ctx context.Context, event string, attributes map[string]interface{})
	SetAttribute(ctx context.Context, key string, value interface{})
}

// MockTracer provides a simple tracing implementation for development
type MockTracer struct {
	name string
}

// NewMockTracer creates a new mock tracer
func NewMockTracer(name string) *MockTracer {
	return &MockTracer{name: name}
}

// Start begins a new trace span
func (m *MockTracer) Start(ctx context.Context, operationName string) (context.Context, Span) {
	logger := log.FromContext(ctx)
	logger.Debug("Starting trace span", "tracer", m.name, "operation", operationName)

	span := &MockSpan{
		tracer:        m,
		operationName: operationName,
		startTime:     time.Now(),
	}

	// In a real implementation, we'd use a proper tracing context
	// For now, we'll store the span in a simple context value
	return context.WithValue(ctx, "current_span", span), span
}

// MockSpan represents a trace span
type MockSpan struct {
	tracer        *MockTracer
	operationName string
	startTime     time.Time
}

// End completes the trace span
func (s *MockSpan) End() {
	logger := log.Default()
	duration := time.Since(s.startTime)
	logger.Debug("Trace span ended",
		"tracer", s.tracer.name,
		"operation", s.operationName,
		"duration_ms", duration.Milliseconds())
}

// AddEvent adds an event to the current span
func (s *MockSpan) AddEvent(ctx context.Context, event string, attributes map[string]interface{}) {
	logger := log.FromContext(ctx)
	logger.Debug("Trace event",
		"tracer", s.tracer.name,
		"operation", s.operationName,
		"event", event,
		"attributes", attributes)
}

// SetAttribute sets an attribute on the current span
func (s *MockSpan) SetAttribute(ctx context.Context, key string, value interface{}) {
	logger := log.FromContext(ctx)
	logger.Debug("Trace attribute set",
		"tracer", s.tracer.name,
		"operation", s.operationName,
		"key", key,
		"value", value)
}

// GetCurrentSpan returns the current span from context
func GetCurrentSpan(ctx context.Context) Span {
	if span, ok := ctx.Value("current_span").(Span); ok {
		return span
	}
	return nil
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

// AddEvent adds an event to the current span (no-op implementation)
func (n *NoOpSpan) AddEvent(ctx context.Context, event string, attributes map[string]interface{}) {
	logger := log.FromContext(ctx)
	logger.Debug("Trace event", "tracer", "noop", "event", event, "attributes", attributes)
}

// SetAttribute sets an attribute on the current span (no-op implementation)
func (n *NoOpSpan) SetAttribute(ctx context.Context, key string, value interface{}) {
	logger := log.FromContext(ctx)
	logger.Debug("Trace attribute set", "tracer", "noop", "key", key, "value", value)
}
