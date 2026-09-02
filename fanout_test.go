package telemetry

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestFanoutHandler(t *testing.T) {
	t.Parallel()
	var info bytes.Buffer
	var errorOnly bytes.Buffer
	handler := NewFanoutHandler(
		slog.NewJSONHandler(&errorOnly, &slog.HandlerOptions{Level: slog.LevelError}),
		slog.NewJSONHandler(&info, nil),
	)
	if !handler.Enabled(context.Background(), slog.LevelInfo) || handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("fanout Enabled did not reflect its children")
	}
	record := slog.NewRecord(time.Unix(1, 0), slog.LevelInfo, "fanout", 0)
	record.AddAttrs(slog.String("entity", "post"))
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if errorOnly.Len() != 0 || !strings.Contains(info.String(), `"entity":"post"`) {
		t.Fatalf("unexpected fanout output: error=%s info=%s", errorOnly.String(), info.String())
	}
}

func TestFanoutHandlerAttrsGroupsAndErrors(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := NewFanoutHandler(slog.NewJSONHandler(&output, nil)).
		WithAttrs([]slog.Attr{slog.String("service", "api")}).
		WithGroup("audit")
	record := slog.NewRecord(time.Unix(1, 0), slog.LevelInfo, "event", 0)
	record.AddAttrs(slog.String("action", "publish"))
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); !strings.Contains(got, `"service":"api"`) || !strings.Contains(got, `"audit":{"action":"publish"}`) {
		t.Fatalf("unexpected fanout output: %s", got)
	}

	expected := errors.New("handler failed")
	if err := NewFanoutHandler(fanoutFailingHandler{err: expected}).Handle(context.Background(), record); !errors.Is(err, expected) {
		t.Fatalf("Handle error = %v", err)
	}
}

type fanoutFailingHandler struct{ err error }

func (handler fanoutFailingHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (handler fanoutFailingHandler) Handle(context.Context, slog.Record) error { return handler.err }
func (handler fanoutFailingHandler) WithAttrs([]slog.Attr) slog.Handler        { return handler }
func (handler fanoutFailingHandler) WithGroup(string) slog.Handler             { return handler }
