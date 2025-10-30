package tracing

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
)

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
	
	span := &Span{
		tracer:       m,
		operationName: operationName,
		startTime:    time.Now(),
	}
	
	// In a real implementation, we'd use a proper tracing context
	// For now, we'll store the span in a simple context value
	return context.WithValue(ctx, "current_span", span), span
}

// Span represents a trace span
type Span struct {
	tracer       *MockTracer
	operationName string
	startTime    time.Time
}

// End completes the trace span
func (s *Span) End() {
	logger := log.Default()
	duration := time.Since(s.startTime)
	logger.Debug("Trace span ended", 
		"tracer", s.tracer.name,
		"operation", s.operationName,
		"duration_ms", duration.Milliseconds())
}

// AddEvent adds an event to the current span
func (s *Span) AddEvent(ctx context.Context, event string, attributes map[string]interface{}) {
	logger := log.FromContext(ctx)
	logger.Debug("Trace event", 
		"tracer", s.tracer.name,
		"operation", s.operationName,
		"event", event,
		"attributes", attributes)
}

// SetAttribute sets an attribute on the current span
func (s *Span) SetAttribute(ctx context.Context, key string, value interface{}) {
	logger := log.FromContext(ctx)
	logger.Debug("Trace attribute set", 
		"tracer", s.tracer.name,
		"operation", s.operationName,
		"key", key,
		"value", value)
}

// GetCurrentSpan returns the current span from context
func GetCurrentSpan(ctx context.Context) *Span {
	if span, ok := ctx.Value("current_span").(*Span); ok {
		return span
	}
	return nil
}