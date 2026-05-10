// Package tracing инициализирует OpenTelemetry для Gin.
package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init OTLP HTTP exporter и TracerProvider. Возвращает shutdown.
func Init(ctx context.Context, enabled bool, serviceName, otlpEndpoint string) (func(context.Context) error, error) {
	if !enabled {
		return func(context.Context) error { return nil }, nil
	}
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(trimScheme(otlpEndpoint)),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", serviceName)),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	return func(c context.Context) error {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}, nil
}

func trimScheme(endpoint string) string {
	for _, p := range []string{"https://", "http://"} {
		if len(endpoint) > len(p) && endpoint[:len(p)] == p {
			return endpoint[len(p):]
		}
	}
	return endpoint
}
