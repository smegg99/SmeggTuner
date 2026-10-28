// backend/common/logger/slog.go
package logger

import (
	"context"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/smegg99/s99logger"
)

// Wails routes framework diagnostics through the configured application sinks and severity filter.
func Wails() *slog.Logger {
	return slog.New(&wailsHandler{group: "wails."})
}

const wailsMessage s99logger.MessageID = "wails.runtime"

// wailsDebug is whether the console takes debug lines; the logger itself always runs at debug for the
// report, so Wails' per-call debug lines are dropped here unless someone reads them.
var wailsDebug atomic.Bool

type wailsHandler struct {
	attrs []s99logger.Attr
	group string
}

func (*wailsHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo || wailsDebug.Load()
}

func (h *wailsHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}
	attrs := append(slices.Clone(h.attrs), s99logger.String("source", "wails"), s99logger.String("detail", r.Message))
	r.Attrs(func(a slog.Attr) bool {
		attrs = appendSlogAttr(attrs, h.group, a)
		return true
	})
	event := s99logger.NewEvent(wailsMessage, attrs...)
	log := read()
	level := r.Level
	if refusal(r.Message) || traffic(r.Message) {
		level = slog.LevelDebug
	}
	if level < slog.LevelInfo && !wailsDebug.Load() {
		return nil
	}
	switch {
	case level >= slog.LevelError:
		log.Error(event, ctx)
	case level >= slog.LevelWarn:
		log.Warn(event, ctx)
	case level >= slog.LevelInfo:
		log.Info(event, ctx)
	default:
		log.Debug(event, ctx)
	}
	return nil
}

// errorKey is a service error's i18n key, the form every deliberate refusal takes (session.error.noSession).
var errorKey = regexp.MustCompile(`^[a-z][A-Za-z0-9]*(\.[A-Za-z0-9]+)+$`)

// refusal reports Wails' line for a bound method that returned one of the app's own refusals: the
// frontend shows it and the service logs it where it matters, so it is no error of the app.
func refusal(message string) bool {
	_, after, ok := strings.Cut(message, "Bound method returned an error: ")
	return ok && errorKey.MatchString(strings.TrimSpace(after))
}

// traffic reports Wails' line for each file its asset server hands the webview: one per request at INFO,
// which buries the app's own log, while a failed request still reaches it as an error.
func traffic(message string) bool {
	return strings.HasSuffix(message, "] Handling request")
}

func (h *wailsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := &wailsHandler{attrs: slices.Clone(h.attrs), group: h.group}
	for _, attr := range attrs {
		next.attrs = appendSlogAttr(next.attrs, h.group, attr)
	}
	return next
}

func (h *wailsHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &wailsHandler{attrs: slices.Clone(h.attrs), group: h.group + name + "."}
}

func appendSlogAttr(attrs []s99logger.Attr, prefix string, attr slog.Attr) []s99logger.Attr {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return attrs
	}
	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			prefix += attr.Key + "."
		}
		for _, child := range attr.Value.Group() {
			attrs = appendSlogAttr(attrs, prefix, child)
		}
		return attrs
	}
	value := attr.Value.Any()
	if err, ok := value.(error); ok {
		value = err.Error()
	}
	return append(attrs, s99logger.Any(prefix+attr.Key, value))
}
