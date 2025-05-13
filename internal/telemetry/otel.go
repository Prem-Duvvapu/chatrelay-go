package telemetry

import (
    "context"
    "log"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer(serviceName string) func(context.Context) error {
    exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
    if err != nil {
        log.Fatalf("failed to initialize stdouttrace exporter: %v", err)
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)

    log.Println("OpenTelemetry initialized")

    return tp.Shutdown
}
