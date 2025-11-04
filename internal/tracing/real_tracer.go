package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// RealTracer implements production-ready tracing with OpenTelemetry
type RealTracer struct {
	tracer trace.Tracer
}

// RealSpan adapts OpenTelemetry span to our Span interface
type RealSpan struct {
	span trace.Span
}

// NewRealTracer creates a new production tracer with Jaeger export
func NewRealTracer(serviceName string) *RealTracer {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		// Fallback to stdout exporter
		exp, _ = jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost("localhost")))
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			"",
			semconv.ServiceName(serviceName),
		)),
	)

	// Register as global tracer provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &RealTracer{
		tracer: otel.Tracer(serviceName),
	}
}

// Start creates a new span
func (t *RealTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, name)
	return ctx, &RealSpan{span: span}
}

// Close shuts down tracer
func (t *RealTracer) Close() error {
	// Shutdown tracer provider
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(context.Background())
	}
	return nil
}

// RealSpan adapter methods
func (rs *RealSpan) End() {
	rs.span.End()
}

func (rs *RealSpan) AddEvent(ctx context.Context, event string, attributes map[string]any) {
	// Convert map[string]interface{} to attribute slice
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", v)))
	}
	rs.span.AddEvent(event, trace.WithAttributes(attrs...))
}

func (rs *RealSpan) SetAttribute(ctx context.Context, key string, value any) {
	rs.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}
