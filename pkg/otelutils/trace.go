package otelutils

import (
	"context"
	"os"
	"strconv"

	"github.com/helin0815/gowebmonitor/pkg/log"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc/credentials"
)

const (
	OtelEndpointEnvVar = "OTEL_EXPORTER_OTLP_ENDPOINT"
	OtelInsecureEnvVar = "OTEL_EXPORTER_OTLP_INSECURE"
)

func newTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	endpoint := os.Getenv(OtelEndpointEnvVar)
	if endpoint == "" {
		log.Warnf("OTEL_EXPORTER_OTLP_ENDPOINT not set, skipping Opentelemtry tracing")
		return nil, nil
	}

	insecure, err := strconv.ParseBool(os.Getenv(OtelInsecureEnvVar))
	if err != nil {
		insecure = true
	}

	grpcOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if insecure {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
	} else {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exporter, err := otlptracegrpc.New(ctx, grpcOpts...)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newResource() *resource.Resource {
	serviceName := os.Getenv("CHJ_APP_NAME")
	if serviceName == "" {
		serviceName = "lsego-service"
	}

	r, err := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)))
	if err != nil {
		log.Panicf("unable to merge resource: %v", err)
	}

	return r
}

func InitProvider(ctx context.Context) func(ctx context.Context) {
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(newResource()),
	}

	exp, err := newTraceExporter(ctx)
	if err != nil {
		log.Warnf("failed to initialize exporter")
	}

	if exp != nil {
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	return func(ctx context.Context) {
		_ = tp.Shutdown(ctx)
		if exp != nil {
			_ = exp.Shutdown(ctx)
		}
	}
}
