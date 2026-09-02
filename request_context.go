package telemetry

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestContext is correlation data created at an ingress boundary.
// OpenTelemetry context stays in context.Context rather than being copied.
type RequestContext struct {
	RequestID   string
	Actor       Actor
	RequestedAt time.Time
	SourceIP    string
}

type requestContextKey struct{}
type requestActorStateKey struct{}

type requestActorState struct {
	mu    sync.RWMutex
	actor Actor
}

func (state *requestActorState) set(actor Actor) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.actor = actor
}

func (state *requestActorState) get() Actor {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.actor
}

// NewPublicRequestContext creates the untrusted public-ingress correlation.
func NewPublicRequestContext(sourceIP string) (RequestContext, error) {
	if err := validateSourceIP(sourceIP); err != nil {
		return RequestContext{}, err
	}
	return RequestContext{
		RequestID:   uuid.NewString(),
		Actor:       AnonymousActor{},
		RequestedAt: time.Now().UTC(),
		SourceIP:    sourceIP,
	}, nil
}

// NewPropagatedRequestContext restores correlation at a trusted GEUL boundary.
func NewPropagatedRequestContext(requestID string, actor Actor) (RequestContext, error) {
	if err := ValidateRequestID(requestID); err != nil {
		return RequestContext{}, err
	}
	if actor == nil {
		actor = AnonymousActor{}
	}
	return RequestContext{
		RequestID:   requestID,
		Actor:       actor,
		RequestedAt: time.Now().UTC(),
	}, nil
}

func WithRequestContext(ctx context.Context, requestContext RequestContext) context.Context {
	if requestContext.Actor == nil {
		requestContext.Actor = AnonymousActor{}
	}
	state := &requestActorState{actor: requestContext.Actor}
	ctx = context.WithValue(ctx, requestActorStateKey{}, state)
	return context.WithValue(ctx, requestContextKey{}, requestContext)
}

// WithActor returns a child context with a resolved actor. An ingress-owned
// terminal recorder can also read the resolved actor from its parent context.
func WithActor(ctx context.Context, actor Actor) context.Context {
	if actor == nil {
		actor = AnonymousActor{}
	}
	if state, ok := ctx.Value(requestActorStateKey{}).(*requestActorState); ok {
		state.set(actor)
	}
	requestContext, ok := ctx.Value(requestContextKey{}).(RequestContext)
	if !ok {
		return ctx
	}
	requestContext.Actor = actor
	return context.WithValue(ctx, requestContextKey{}, requestContext)
}

func RequestContextFrom(ctx context.Context) (RequestContext, bool) {
	requestContext, ok := ctx.Value(requestContextKey{}).(RequestContext)
	if !ok {
		return RequestContext{}, false
	}
	if state, stateOK := ctx.Value(requestActorStateKey{}).(*requestActorState); stateOK {
		requestContext.Actor = state.get()
	}
	return requestContext, true
}

func RequestIDFromContext(ctx context.Context) string {
	requestContext, ok := RequestContextFrom(ctx)
	if !ok {
		return ""
	}
	return requestContext.RequestID
}

func ValidateRequestID(requestID string) error {
	parsed, err := uuid.Parse(requestID)
	if err != nil || parsed.String() != requestID || parsed.Version() != 4 || parsed.Variant() != uuid.RFC4122 {
		return ErrInvalidRequestID
	}
	return nil
}

func validateSourceIP(sourceIP string) error {
	if sourceIP == "" {
		return nil
	}
	parsed, err := netip.ParseAddr(sourceIP)
	if err != nil || parsed.String() != sourceIP {
		return ErrInvalidSourceIP
	}
	return nil
}
