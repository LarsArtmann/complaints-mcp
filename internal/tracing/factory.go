package tracing

import (
	"os"
	"strconv"
)

// TracerType represents the type of tracer to use
type TracerType string

const (
	TracerTypeMock TracerType = "mock"
	TracerTypeReal TracerType = "real"
)

// TracerConfig holds configuration for tracer creation
type TracerConfig struct {
	Type        TracerType
	ServiceName string
	JaegerURL  string
	SampleRate  float64
}

// DefaultTracerConfig returns default tracer configuration
func DefaultTracerConfig() TracerConfig {
	tracerType := TracerTypeMock
	if os.Getenv("TRACER_TYPE") == "real" {
		tracerType = TracerTypeReal
	}

	sampleRate := 1.0
	if envSample := os.Getenv("TRACE_SAMPLE_RATE"); envSample != "" {
		if parsed, err := strconv.ParseFloat(envSample, 64); err == nil {
			sampleRate = parsed
		}
	}

	return TracerConfig{
		Type:        tracerType,
		ServiceName: "complaints-mcp",
		JaegerURL:  getEnvOrDefault("JAEGER_URL", "http://localhost:14268/api/traces"),
		SampleRate:  sampleRate,
	}
}

// NewTracer creates a tracer based on configuration
func NewTracer(config TracerConfig) Tracer {
	switch config.Type {
	case TracerTypeReal:
		return NewRealTracer(config.ServiceName)
	case TracerTypeMock:
		fallthrough
	default:
		return NewMockTracer(config.ServiceName)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}