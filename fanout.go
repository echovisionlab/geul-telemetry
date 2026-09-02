package telemetry

import (
	"context"
	"log/slog"
)

type fanoutHandler struct {
	handlers []slog.Handler
}

// NewFanoutHandler sends each record to every enabled handler. Service-local
// telemetry setup can compose stdout and OTLP without reimplementing slog
// handler semantics.
func NewFanoutHandler(handlers ...slog.Handler) slog.Handler {
	return &fanoutHandler{handlers: handlers}
}

func (handler *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, child := range handler.handlers {
		if child.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (handler *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, child := range handler.handlers {
		if child.Enabled(ctx, record.Level) {
			if err := child.Handle(ctx, record.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (handler *fanoutHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	children := make([]slog.Handler, len(handler.handlers))
	for index, child := range handler.handlers {
		children[index] = child.WithAttrs(attributes)
	}
	return &fanoutHandler{handlers: children}
}

func (handler *fanoutHandler) WithGroup(name string) slog.Handler {
	children := make([]slog.Handler, len(handler.handlers))
	for index, child := range handler.handlers {
		children[index] = child.WithGroup(name)
	}
	return &fanoutHandler{handlers: children}
}
