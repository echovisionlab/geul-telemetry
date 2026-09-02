package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type namedTestError struct{}

func (namedTestError) Error() string { return "secret error text" }

type failingHandler struct{}

func (failingHandler) Enabled(context.Context, slog.Level) bool   { return true }
func (failingHandler) Handle(context.Context, slog.Record) error  { return errors.New("handler failed") }
func (handler failingHandler) WithAttrs([]slog.Attr) slog.Handler { return handler }
func (handler failingHandler) WithGroup(string) slog.Handler      { return handler }

func TestNormalizingHandler(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	base := slog.NewJSONHandler(&output, nil)
	handler := NewNormalizingHandler(base)
	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("handler unexpectedly disabled")
	}

	requestContext, _ := NewPropagatedRequestContext(testRequestID, AnonymousActor{})
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})
	ctx := trace.ContextWithSpanContext(WithRequestContext(context.Background(), requestContext), spanContext)

	logger := slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("ServiceName", "api"),
		slog.String("email", "secret@example.com"),
	})).WithGroup("Request Data")
	logger.InfoContext(ctx, "handled",
		slog.Int("StatusCode", 200),
		slog.Any("err", fmt.Errorf("provider request failed: %w", namedTestError{})),
		slog.Any("payload_object", struct{ Secret string }{Secret: "hidden"}),
		slog.Group("Safe Group", slog.String("operation", "read"), slog.String("access_token", "secret")),
		slog.Group("Hidden Group", slog.String("password", "secret")),
		slog.Group("Translation",
			slog.String("ProviderType", "deepl"),
			slog.String("model", "catalog-model"),
			slog.String("sourceText", "source secret"),
			slog.String("translated-text", "translated secret"),
			slog.String("systemPrompt", "prompt secret"),
			slog.String("glossary", "glossary secret"),
			slog.String("contextPayload", "context secret"),
			slog.Group("API", slog.String("Key", "api secret")),
			slog.Group("Provider",
				slog.Group("Document",
					slog.String("ID", "document id secret"),
					slog.String("Key", "document key secret"),
					slog.String("SubmittedAt", "2026-08-21T00:00:00Z"),
				),
				slog.Group("Response",
					slog.String("Body", "response body secret"),
					slog.String("Message", "response message secret"),
					slog.String("HTTP Status Class", "5xx"),
				),
			),
		),
	)

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["service_name"] != "api" || record["request_id"] != testRequestID || record["trace_id"] != traceID.String() || record["span_id"] != spanID.String() {
		t.Fatalf("normalized record = %#v", record)
	}
	requestData, ok := record["request_data"].(map[string]any)
	if !ok || requestData["status_code"] != float64(200) || requestData["error_type"] != "wrap_error" {
		t.Fatalf("request group = %#v", record["request_data"])
	}
	translation, ok := requestData["translation"].(map[string]any)
	if !ok || translation["provider_type"] != "deepl" || translation["model"] != "catalog-model" {
		t.Fatalf("translation group = %#v", requestData["translation"])
	}
	provider, ok := translation["provider"].(map[string]any)
	if !ok {
		t.Fatalf("provider group = %#v", translation["provider"])
	}
	document, ok := provider["document"].(map[string]any)
	if !ok || document["submitted_at"] != "2026-08-21T00:00:00Z" {
		t.Fatalf("provider document group = %#v", provider["document"])
	}
	response, ok := provider["response"].(map[string]any)
	if !ok || response["http_status_class"] != "5xx" {
		t.Fatalf("provider response group = %#v", provider["response"])
	}
	if _, exists := requestData["payload_object"]; exists {
		t.Fatal("arbitrary object was not removed")
	}
	if _, exists := record["email"]; exists {
		t.Fatal("forbidden field was not removed")
	}
	if bytes.Contains(output.Bytes(), []byte("secret")) {
		t.Fatalf("sensitive value entered normalized log: %s", output.Bytes())
	}
}

func TestNormalizingHandlerUsesFullWithGroupPath(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := NewNormalizingHandler(slog.NewJSONHandler(&output, nil))
	logger := slog.New(handler).
		WithGroup("Provider").
		WithGroup("Document").
		With(
			slog.String("ID", "document id secret"),
			slog.String("Key", "document key secret"),
			slog.String("SubmittedAt", "2026-08-21T00:00:00Z"),
		)
	logger.Info("provider document resumed")

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	provider := record["provider"].(map[string]any)
	document := provider["document"].(map[string]any)
	if document["submitted_at"] != "2026-08-21T00:00:00Z" {
		t.Fatalf("provider document group = %#v", document)
	}
	if _, exists := document["id"]; exists {
		t.Fatal("nested provider document ID was not removed")
	}
	if _, exists := document["key"]; exists {
		t.Fatal("nested provider document key was not removed")
	}
	if bytes.Contains(output.Bytes(), []byte("secret")) {
		t.Fatalf("sensitive value entered WithGroup log: %s", output.Bytes())
	}
}

func TestNormalizingHandlerPreservesExplicitCorrelation(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := NewNormalizingHandler(slog.NewJSONHandler(&output, nil))
	record := slog.NewRecord(testOccurredAt, slog.LevelInfo, "explicit", 0)
	record.AddAttrs(
		slog.String("request_id", "explicit-request"),
		slog.String("trace_id", "explicit-trace"),
		slog.String("span_id", "explicit-span"),
	)
	requestContext, _ := NewPropagatedRequestContext(testRequestID, AnonymousActor{})
	if err := handler.Handle(WithRequestContext(context.Background(), requestContext), record); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("explicit-request")) {
		t.Fatalf("output = %s", output.Bytes())
	}
}

func TestRedactionUtilities(t *testing.T) {
	t.Parallel()
	for _, key := range []string{
		"authorization", "access_token", "secret_value", "identity_id", "session_id", "flow_id",
		"error_reason", "member_id", "member_display_name", "member_name", "nickname", "user_id",
	} {
		if !IsForbiddenKey(key) {
			t.Fatalf("IsForbiddenKey(%q) = false", key)
		}
	}
	if IsForbiddenKey("component") {
		t.Fatal("safe field rejected")
	}
	if NormalizeKey(" HTTPStatus-Code ") != "httpstatus_code" || NormalizeKey("actorMemberID") != "actor_member_id" || NormalizeKey("---") != "" {
		t.Fatal("NormalizeKey returned unexpected value")
	}
	if StableErrorType(namedTestError{}) != "named_test_error" || StableErrorType(&namedTestError{}) != "named_test_error" || StableErrorType(nil) != "reported_error" || StableErrorType(struct{}{}) != "reported_error" {
		t.Fatal("StableErrorType returned unexpected value")
	}
	if attribute, ok := normalizeLogAttribute(slog.String("---", "value")); ok || !attribute.Equal(slog.Attr{}) {
		t.Fatal("empty normalized key was retained")
	}
	if attribute, ok := normalizeLogAttribute(slog.String("error", "raw")); !ok || attribute.Value.String() != "string" {
		t.Fatalf("string error normalization = %#v, %v", attribute, ok)
	}
}

func TestLogRedactionMatchesCrossLanguageFixture(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("fixtures/log-redaction.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Forbidden []struct {
			Key        string `json:"key"`
			Normalized string `json:"normalized"`
		} `json:"forbidden"`
		Allowed []struct {
			Key        string `json:"key"`
			Normalized string `json:"normalized"`
		} `json:"allowed"`
	}
	if err := json.Unmarshal(contents, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, entry := range fixture.Forbidden {
		if normalized := NormalizeKey(entry.Key); normalized != entry.Normalized {
			t.Errorf("NormalizeKey(%q) = %q, want %q", entry.Key, normalized, entry.Normalized)
		}
		if !IsForbiddenKey(entry.Key) {
			t.Errorf("IsForbiddenKey(%q) = false", entry.Key)
		}
	}
	for _, entry := range fixture.Allowed {
		if normalized := NormalizeKey(entry.Key); normalized != entry.Normalized {
			t.Errorf("NormalizeKey(%q) = %q, want %q", entry.Key, normalized, entry.Normalized)
		}
		if IsForbiddenKey(entry.Key) {
			t.Errorf("IsForbiddenKey(%q) = true", entry.Key)
		}
	}
}

func TestTypedEmitters(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := NewNormalizingHandler(slog.NewJSONHandler(&output, nil))
	requestContext, _ := NewPropagatedRequestContext(testRequestID, MemberActor{MemberID: "member-1"})
	ctx := WithRequestContext(context.Background(), requestContext)
	actor, _ := ActorForRecord(MemberActor{MemberID: "member-1"})
	request := RequestRecord{
		Event: "request.completed", RecordActor: actor,
		RPCService: "geul.v1.PostService", RPCMethod: "UpdatePost", HTTPMethod: "POST",
		StatusCode: 403, DurationMS: 8, Outcome: RequestOutcomeBlocked, ErrorCode: "permission_denied", Reason: string(RequestReasonPermissionDenied),
	}
	if err := EmitRequest(ctx, handler, request); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"domain":"request"`)) || !bytes.Contains(output.Bytes(), []byte(`"actor_member_id":"member-1"`)) || !bytes.Contains(output.Bytes(), []byte(`"rpc_service":"geul.v1.PostService"`)) {
		t.Fatalf("request output = %s", output.Bytes())
	}
	if err := EmitRequest(ctx, handler, RequestRecord{}); err == nil {
		t.Fatal("invalid request record emitted")
	}

	output.Reset()
	retryCount := 2
	duration := int64(11)
	system := SystemRecord{
		Event:       EventQueueRetryAccepted,
		Correlation: Correlation{RequestID: testRequestID},
		Component:   "worker", Dependency: "postgres", Operation: "enqueue", Domain: "post",
		Queue: "post.retry", MessageID: "message-1", CommandID: "command-1",
		RetryCount: &retryCount, DurationMS: &duration, JobKind: "mesh_optimization", JobID: "job-1",
		RecordClass: "domain_audit", Action: "post.updated", Outcome: "accepted", ErrorCode: "", Reason: "",
	}
	if err := EmitSystem(context.Background(), handler, system); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"retry_count":2`)) || !bytes.Contains(output.Bytes(), []byte(`"command_id":"command-1"`)) {
		t.Fatalf("system output = %s", output.Bytes())
	}
	if err := EmitSystem(context.Background(), handler, SystemRecord{}); err == nil {
		t.Fatal("invalid system record emitted")
	}
	if err := EmitSystem(context.Background(), failingHandler{}, SystemRecord{Event: EventServiceReady, Outcome: "ready", Component: "api"}); err == nil {
		t.Fatal("handler failure was hidden")
	}
}

func TestActorMarkerMethodsAndSourceIPBranches(t *testing.T) {
	t.Parallel()
	if err := validateSourceIP(""); err != nil {
		t.Fatal(err)
	}
	if err := validateSourceIP("2001:0db8::1"); !errors.Is(err, ErrInvalidSourceIP) {
		t.Fatalf("non-canonical IP error = %v", err)
	}
	if err := validateTraceCorrelation("4bf92f3577b34da6a3ce929d0e0e4736", "bad"); err == nil {
		t.Fatal("invalid span ID accepted")
	}
}

func TestEmitterUsesExistingTimeAndCorrelation(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := slog.NewJSONHandler(&output, nil)
	record := validRequestRecord()
	record.TraceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	record.SpanID = "00f067aa0ba902b7"
	if err := EmitRequest(context.Background(), handler, record); err != nil {
		t.Fatal(err)
	}
	system := SystemRecord{
		Event: EventServiceDegraded, OccurredAt: time.Now().UTC(),
		Correlation: Correlation{RequestID: testRequestID, TraceID: record.TraceID, SpanID: record.SpanID},
		Component:   "api", Outcome: "degraded", ErrorCode: "dependency_unavailable", Reason: "dependency_unavailable",
	}
	if err := EmitSystem(context.Background(), handler, system); err != nil {
		t.Fatal(err)
	}
}
