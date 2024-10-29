package otelutils

import (
	"context"

	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func SpanTrackEvent(ctx context.Context, event string, attributes map[string]string) {
	trace.SpanFromContext(ctx).AddEvent(event, trace.WithAttributes(lo.MapToSlice(attributes, func(k string, v string) attribute.KeyValue {
		return attribute.KeyValue{Key: attribute.Key(k), Value: attribute.StringValue(v)}
	})...))
}

func TraceIdFromContext(ctx context.Context) string {
	return trace.SpanContextFromContext(ctx).TraceID().String()
}
