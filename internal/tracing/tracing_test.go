package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTracerConfig(t *testing.T) {
	config := DefaultTracerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "complaints-mcp", config.ServiceName)
	assert.Equal(t, TracerTypeReal, config.Type)
}

func TestNewTracer(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	assert.NotNil(t, tracer)
}

func TestMockTracer(t *testing.T) {
	tracer := NewMockTracer("test-service")

	assert.NotNil(t, tracer)

	// Test Start method
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test-operation")

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)

	// Test span End method (should not panic)
	assert.NotPanics(t, func() {
		span.End()
	})
}

func TestRealTracer(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewRealTracer(config.ServiceName)

	assert.NotNil(t, tracer)

	// Test Start method
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test-operation")

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)

	// Test span End method (should not panic)
	assert.NotPanics(t, func() {
		span.End()
	})

	// Test Close method
	err := tracer.Close()
	require.NoError(t, err)
}

func TestTracerInterface(t *testing.T) {
	// Ensure both mock and real tracers implement the Tracer interface
	var _ Tracer = NewMockTracer("test-service")
	var _ Tracer = NewRealTracer("test-service")
}
