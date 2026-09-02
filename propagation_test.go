package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestInjectAndExtractCorrelation(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })

	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})
	requestContext, err := NewPropagatedRequestContext(testRequestID, MemberActor{MemberID: "member-1"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := trace.ContextWithSpanContext(WithRequestContext(context.Background(), requestContext), spanContext)
	carrier := propagation.MapCarrier{}
	InjectCorrelation(ctx, carrier)
	if carrier.Get(MessageRequestIDHeader) != testRequestID || carrier.Get("traceparent") == "" {
		t.Fatalf("carrier = %#v", carrier)
	}

	extracted := ExtractCorrelation(context.Background(), carrier, SystemActor{ServiceName: ServiceTranscoder})
	correlation := CorrelationFromContext(extracted)
	if correlation.RequestID != testRequestID || correlation.TraceID != traceID.String() || correlation.SpanID != spanID.String() {
		t.Fatalf("correlation = %#v", correlation)
	}
	requestValue, ok := RequestContextFrom(extracted)
	if !ok || requestValue.Actor.Kind() != ActorKindSystem {
		t.Fatalf("request context = %#v, %v", requestValue, ok)
	}
}

func TestExtractCorrelationIgnoresInvalidRequestID(t *testing.T) {
	t.Parallel()
	carrier := propagation.MapCarrier{MessageRequestIDHeader: "bad"}
	ctx := ExtractCorrelation(context.Background(), carrier, AnonymousActor{})
	if _, ok := RequestContextFrom(ctx); ok {
		t.Fatal("invalid propagated request ID was accepted")
	}
}
