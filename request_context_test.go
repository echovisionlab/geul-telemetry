package telemetry

import (
	"context"
	"errors"
	"testing"
)

const testRequestID = "018f47a2-8a3d-4e17-9d42-6f12c89b1234"

func TestPublicRequestContextAndActorResolution(t *testing.T) {
	t.Parallel()
	requestContext, err := NewPublicRequestContext("192.0.2.4")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRequestID(requestContext.RequestID); err != nil {
		t.Fatalf("generated request ID: %v", err)
	}
	ctx := WithRequestContext(context.Background(), requestContext)
	child := WithActor(ctx, MemberActor{MemberID: "member-1"})
	for _, current := range []context.Context{ctx, child} {
		got, ok := RequestContextFrom(current)
		if !ok || got.Actor.Kind() != ActorKindMember || RequestIDFromContext(current) != requestContext.RequestID {
			t.Fatalf("resolved context = %#v, %v", got, ok)
		}
	}
}

func TestRequestContextValidationAndMissingValues(t *testing.T) {
	t.Parallel()
	if _, err := NewPublicRequestContext("not-an-ip"); !errors.Is(err, ErrInvalidSourceIP) {
		t.Fatalf("NewPublicRequestContext() error = %v", err)
	}
	if err := ValidateRequestID("BAD"); !errors.Is(err, ErrInvalidRequestID) {
		t.Fatalf("ValidateRequestID() error = %v", err)
	}
	if err := ValidateRequestID("018f47a2-8a3d-1e17-9d42-6f12c89b1234"); !errors.Is(err, ErrInvalidRequestID) {
		t.Fatalf("ValidateRequestID() version error = %v", err)
	}
	if err := ValidateRequestID("018f47a2-8a3d-4e17-7d42-6f12c89b1234"); !errors.Is(err, ErrInvalidRequestID) {
		t.Fatalf("ValidateRequestID() variant error = %v", err)
	}
	if _, err := NewPropagatedRequestContext("bad", AnonymousActor{}); !errors.Is(err, ErrInvalidRequestID) {
		t.Fatalf("NewPropagatedRequestContext() error = %v", err)
	}
	propagated, err := NewPropagatedRequestContext(testRequestID, nil)
	if err != nil || propagated.Actor.Kind() != ActorKindAnonymous {
		t.Fatalf("NewPropagatedRequestContext() = %#v, %v", propagated, err)
	}
	ctx := WithRequestContext(context.Background(), RequestContext{RequestID: testRequestID})
	if requestContext, ok := RequestContextFrom(ctx); !ok || requestContext.Actor.Kind() != ActorKindAnonymous {
		t.Fatalf("WithRequestContext() = %#v, %v", requestContext, ok)
	}
	if RequestIDFromContext(context.Background()) != "" {
		t.Fatal("missing context returned request ID")
	}
	if _, ok := RequestContextFrom(context.Background()); ok {
		t.Fatal("missing request context reported present")
	}
	if got := WithActor(context.Background(), nil); got == nil {
		t.Fatal("WithActor returned nil context")
	}
}
