package telemetry

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// NormalizingHandler enforces key spelling, correlation injection, and the
// GEUL denylist before a record reaches stdout or an OTLP bridge.
type NormalizingHandler struct {
	next       slog.Handler
	attributes []slog.Attr
	groups     []string
}

func NewNormalizingHandler(next slog.Handler) slog.Handler {
	return &NormalizingHandler{next: next}
}

func (handler *NormalizingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.next.Enabled(ctx, level)
}

func (handler *NormalizingHandler) Handle(ctx context.Context, record slog.Record) error {
	normalized := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	seenRequestID := false
	seenTraceID := false
	seenSpanID := false
	for _, attribute := range handler.attributes {
		normalized.AddAttrs(attribute)
		seenRequestID = seenRequestID || attribute.Key == "request_id"
		seenTraceID = seenTraceID || attribute.Key == "trace_id"
		seenSpanID = seenSpanID || attribute.Key == "span_id"
	}
	recordAttributes := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attribute slog.Attr) bool {
		attribute, ok := normalizeLogAttributeAtPath(attribute, handler.groups)
		if !ok {
			return true
		}
		if len(handler.groups) == 0 {
			seenRequestID = seenRequestID || attribute.Key == "request_id"
			seenTraceID = seenTraceID || attribute.Key == "trace_id"
			seenSpanID = seenSpanID || attribute.Key == "span_id"
		}
		recordAttributes = append(recordAttributes, attribute)
		return true
	})
	normalized.AddAttrs(wrapLogGroups(recordAttributes, handler.groups)...)
	correlation := CorrelationFromContext(ctx)
	if correlation.RequestID != "" && !seenRequestID {
		normalized.AddAttrs(slog.String("request_id", correlation.RequestID))
	}
	if correlation.TraceID != "" && !seenTraceID {
		normalized.AddAttrs(slog.String("trace_id", correlation.TraceID))
	}
	if correlation.SpanID != "" && !seenSpanID {
		normalized.AddAttrs(slog.String("span_id", correlation.SpanID))
	}
	return handler.next.Handle(ctx, normalized)
}

func (handler *NormalizingHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	normalized := make([]slog.Attr, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute, ok := normalizeLogAttributeAtPath(attribute, handler.groups); ok {
			normalized = append(normalized, attribute)
		}
	}
	bound := append([]slog.Attr{}, handler.attributes...)
	bound = append(bound, wrapLogGroups(normalized, handler.groups)...)
	return &NormalizingHandler{next: handler.next, attributes: bound, groups: append([]string{}, handler.groups...)}
}

func (handler *NormalizingHandler) WithGroup(name string) slog.Handler {
	groups := append([]string{}, handler.groups...)
	groups = append(groups, NormalizeKey(name))
	return &NormalizingHandler{next: handler.next, attributes: append([]slog.Attr{}, handler.attributes...), groups: groups}
}

func wrapLogGroups(attributes []slog.Attr, groups []string) []slog.Attr {
	if len(attributes) == 0 || len(groups) == 0 {
		return attributes
	}
	wrapped := append([]slog.Attr{}, attributes...)
	for index := len(groups) - 1; index >= 0; index-- {
		values := make([]any, len(wrapped))
		for attributeIndex, attribute := range wrapped {
			values[attributeIndex] = attribute
		}
		wrapped = []slog.Attr{slog.Group(groups[index], values...)}
	}
	return wrapped
}

func IsForbiddenKey(key string) bool {
	key = NormalizeKey(key)
	if _, forbidden := forbiddenLogKeys[key]; forbidden {
		return true
	}
	for _, suffix := range forbiddenLogSuffixes {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}
	for _, prefix := range forbiddenLogPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func normalizeLogAttribute(attribute slog.Attr) (slog.Attr, bool) {
	return normalizeLogAttributeAtPath(attribute, nil)
}

func normalizeLogAttributeAtPath(attribute slog.Attr, parentPath []string) (slog.Attr, bool) {
	attribute.Value = attribute.Value.Resolve()
	key := NormalizeKey(attribute.Key)
	path := append(append([]string{}, parentPath...), key)
	if key == "" || IsForbiddenKey(key) || IsForbiddenKey(strings.Join(path, "_")) {
		return slog.Attr{}, false
	}
	if key == "error" || key == "err" {
		return slog.String("error_type", StableErrorType(attribute.Value.Any())), true
	}
	if attribute.Value.Kind() == slog.KindGroup {
		children := attribute.Value.Group()
		normalized := make([]any, 0, len(children))
		for _, child := range children {
			if child, ok := normalizeLogAttributeAtPath(child, path); ok {
				normalized = append(normalized, child)
			}
		}
		if len(normalized) == 0 {
			return slog.Attr{}, false
		}
		return slog.Group(key, normalized...), true
	}
	if attribute.Value.Kind() == slog.KindAny {
		return slog.Attr{}, false
	}
	attribute.Key = key
	return attribute, true
}

func StableErrorType(value any) string {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return "reported_error"
	}
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	if valueType.Name() == "" {
		return "reported_error"
	}
	return NormalizeKey(valueType.Name())
}

func NormalizeKey(key string) string {
	key = strings.TrimSpace(key)
	var output strings.Builder
	output.Grow(len(key) + 4)
	lastUnderscore := false
	var previous rune
	for index, current := range key {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			if unicode.IsUpper(current) && index > 0 && !lastUnderscore && (unicode.IsLower(previous) || unicode.IsDigit(previous)) {
				output.WriteByte('_')
			}
			output.WriteRune(unicode.ToLower(current))
			lastUnderscore = false
		} else if output.Len() > 0 && !lastUnderscore {
			output.WriteByte('_')
			lastUnderscore = true
		}
		previous = current
	}
	return strings.Trim(output.String(), "_")
}

// EmitRequest validates and emits a typed request record. Callers intentionally
// choose whether a handler failure is fail-open at their boundary.
func EmitRequest(ctx context.Context, handler slog.Handler, record RequestRecord) error {
	if record.OccurredAt.IsZero() {
		record.OccurredAt = time.Now().UTC()
	}
	if record.RequestID == "" {
		record.Correlation = CorrelationFromContext(ctx)
	}
	if err := record.Validate(); err != nil {
		return err
	}
	logRecord := slog.NewRecord(record.OccurredAt, slog.LevelInfo, "Request completed", 0)
	logRecord.AddAttrs(requestRecordAttributes(record)...)
	return handler.Handle(ctx, logRecord)
}

func requestRecordAttributes(record RequestRecord) []slog.Attr {
	attributes := []slog.Attr{
		slog.String("domain", "request"),
		slog.String("event", record.Event),
		slog.Time("occurred_at", record.OccurredAt),
		slog.String("request_id", record.RequestID),
		slog.String("actor_kind", string(record.Kind)),
		slog.Int("status_code", record.StatusCode),
		slog.Int64("duration_ms", record.DurationMS),
		slog.String("outcome", string(record.Outcome)),
	}
	attributes = appendOptionalString(attributes, "trace_id", record.TraceID)
	attributes = appendOptionalString(attributes, "span_id", record.SpanID)
	attributes = appendOptionalString(attributes, "actor_member_id", record.MemberID)
	attributes = appendOptionalString(attributes, "actor_service", record.Service)
	attributes = appendOptionalString(attributes, "http_method", record.HTTPMethod)
	attributes = appendOptionalString(attributes, "http_route", record.HTTPRoute)
	attributes = appendOptionalString(attributes, "rpc_service", record.RPCService)
	attributes = appendOptionalString(attributes, "rpc_method", record.RPCMethod)
	attributes = appendOptionalString(attributes, "error_code", record.ErrorCode)
	return appendOptionalString(attributes, "reason", record.Reason)
}

func EmitSystem(ctx context.Context, handler slog.Handler, record SystemRecord) error {
	if record.OccurredAt.IsZero() {
		record.OccurredAt = time.Now().UTC()
	}
	if record.RequestID == "" {
		record.Correlation = CorrelationFromContext(ctx)
	}
	if err := record.Validate(); err != nil {
		return err
	}
	level := slog.LevelInfo
	switch record.Outcome {
	case "failed":
		level = slog.LevelError
	case "degraded", "requeued":
		level = slog.LevelWarn
	}
	logRecord := slog.NewRecord(record.OccurredAt, level, "System event", 0)
	attributes := []slog.Attr{slog.String("event", string(record.Event)), slog.Time("occurred_at", record.OccurredAt)}
	attributes = appendOptionalString(attributes, "request_id", record.RequestID)
	attributes = appendOptionalString(attributes, "trace_id", record.TraceID)
	attributes = appendOptionalString(attributes, "span_id", record.SpanID)
	attributes = appendOptionalString(attributes, "component", record.Component)
	attributes = appendOptionalString(attributes, "dependency", record.Dependency)
	attributes = appendOptionalString(attributes, "operation", record.Operation)
	attributes = appendOptionalString(attributes, "domain", record.Domain)
	attributes = appendOptionalString(attributes, "queue", record.Queue)
	attributes = appendOptionalString(attributes, "message_id", record.MessageID)
	attributes = appendOptionalString(attributes, "command_id", record.CommandID)
	attributes = appendOptionalString(attributes, "job_kind", record.JobKind)
	attributes = appendOptionalString(attributes, "job_id", record.JobID)
	attributes = appendOptionalString(attributes, "entity_type", record.EntityType)
	attributes = appendOptionalString(attributes, "entity_id", record.EntityID)
	attributes = appendOptionalString(attributes, "target_locale", record.TargetLocale)
	attributes = appendOptionalString(attributes, "record_class", string(record.RecordClass))
	attributes = appendOptionalString(attributes, "action", record.Action)
	attributes = appendOptionalString(attributes, "outcome", record.Outcome)
	attributes = appendOptionalString(attributes, "error_code", record.ErrorCode)
	attributes = appendOptionalString(attributes, "reason", record.Reason)
	attributes = appendOptionalString(attributes, "error_classification", record.ErrorClassification)
	if record.RetryCount != nil {
		attributes = append(attributes, slog.Int("retry_count", *record.RetryCount))
	}
	if record.DurationMS != nil {
		attributes = append(attributes, slog.Int64("duration_ms", *record.DurationMS))
	}
	logRecord.AddAttrs(attributes...)
	return handler.Handle(ctx, logRecord)
}

func appendOptionalString(attributes []slog.Attr, key, value string) []slog.Attr {
	if value == "" {
		return attributes
	}
	return append(attributes, slog.String(key, value))
}
