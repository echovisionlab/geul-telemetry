package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const MessageRequestIDHeader = "x-request-id"

// InjectCorrelation writes W3C trace context and the independent request ID
// to a trusted internal transport carrier.
func InjectCorrelation(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		carrier.Set(MessageRequestIDHeader, requestID)
	}
}

// ExtractCorrelation restores W3C trace context and, when valid, the request
// ID from a trusted internal transport carrier.
func ExtractCorrelation(ctx context.Context, carrier propagation.TextMapCarrier, actor Actor) context.Context {
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	requestContext, err := NewPropagatedRequestContext(carrier.Get(MessageRequestIDHeader), actor)
	if err != nil {
		return ctx
	}
	return WithRequestContext(ctx, requestContext)
}
