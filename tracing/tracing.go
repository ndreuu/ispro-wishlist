package tracing

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var TracerProvider *sdktrace.TracerProvider

func Init(serviceName, endpoint string) error {
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpointURL("http://"+endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  5 * time.Minute,
		}),
	)
	if err != nil {
		return err
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return err
	}

	TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(TracerProvider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("OTEL ERROR: %v", err)
	}))

	log.Printf("Tracing initialized for %s, endpoint: %s", serviceName, endpoint)
	return nil
}

func Shutdown() error {
	if TracerProvider != nil {
		return TracerProvider.Shutdown(context.Background())
	}
	return nil
}

func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
