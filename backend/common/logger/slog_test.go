// backend/common/logger/slog_test.go
package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/smegg99/s99logger"
)

func TestWailsDiagnosticsUseConfiguredSinkAndLevel(t *testing.T) {
	sink := &records{}
	var encoded bytes.Buffer
	mu.Lock()
	previous := global
	global = s99logger.New(s99logger.MultiSink(sink, s99logger.NewJSONSink(&encoded)), s99logger.Options{MinLevel: s99logger.LevelWarn})
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); global = previous; mu.Unlock() })

	framework := Wails().With("service", "session").WithGroup("shutdown")
	framework.Info("filtered")
	framework.Error("flush failed", "error", errors.New("disk full"), slog.Group("db", "path", "session.db"))
	if len(sink.values) != 1 {
		t.Fatalf("got %d records, want only the error", len(sink.values))
	}
	r := sink.values[0]
	if r.Level != s99logger.LevelError || r.MessageID != "wails.runtime" {
		t.Fatalf("framework severity/identity lost: %+v", r)
	}
	attrs := map[string]any{}
	for _, attr := range r.Attrs {
		attrs[attr.Key] = attr.Value
	}
	for key, want := range map[string]string{
		"source": "wails", "detail": "flush failed", "wails.service": "session", "wails.shutdown.db.path": "session.db",
	} {
		if attrs[key] != want {
			t.Errorf("%s = %v, want %s", key, attrs[key], want)
		}
	}
	if attrs["wails.shutdown.error"] != "disk full" {
		t.Fatalf("lost framework error: %v", attrs)
	}
	var record map[string]any
	if err := json.Unmarshal(encoded.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"detail": "flush failed", "wails.shutdown.error": "disk full", "wails.service": "session"} {
		if record[key] != want {
			t.Errorf("JSON sink lost %s: %s", key, encoded.Bytes())
		}
	}

	// The bridge follows a replacement logger instead of retaining a closed file sink.
	next := &records{}
	mu.Lock()
	global = s99logger.New(next, s99logger.Options{MinLevel: s99logger.LevelDebug})
	mu.Unlock()
	framework.Debug("reconfigured")
	if len(next.values) != 1 || next.values[0].Level != s99logger.LevelDebug || len(sink.values) != 1 {
		t.Fatal("framework logger did not follow the configured sink")
	}
}

func TestWailsHandlerGroupsDoNotMutateParents(t *testing.T) {
	parent := Wails().With("root", 1)
	left := parent.WithGroup("left").With("child", 2)
	right := parent.WithGroup("right").With("child", 3)
	for _, c := range []struct {
		log *slog.Logger
		key string
	}{{left, "wails.left.child"}, {right, "wails.right.child"}} {
		h := c.log.Handler().(*wailsHandler)
		if len(h.attrs) != 2 || h.attrs[0].Key != "wails.root" || h.attrs[1].Key != c.key {
			t.Fatalf("group attributes mutated: %+v", h.attrs)
		}
	}
	if len(parent.Handler().(*wailsHandler).attrs) != 1 {
		t.Fatal("derived logger mutated its parent")
	}
}

// A bound method's refusal (an i18n error key) is no error of the app; anything else a binding fails with is.
func TestBindingRefusalsAreNotErrors(t *testing.T) {
	sink := &records{}
	mu.Lock()
	previous := global
	global = s99logger.New(sink, s99logger.Options{MinLevel: s99logger.LevelDebug})
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); global = previous; mu.Unlock() })

	framework := Wails()
	framework.Error("Binding call failed: Bound method returned an error: session.error.noSession")
	framework.Error("Binding call failed: Bound method returned an error: open /data/x.db: permission denied")
	framework.Error("Binding call failed: no such method")
	want := []s99logger.Level{s99logger.LevelDebug, s99logger.LevelError, s99logger.LevelError}
	if len(sink.values) != len(want) {
		t.Fatalf("got %d records, want %d", len(sink.values), len(want))
	}
	for i, r := range sink.values {
		if r.Level != want[i] {
			t.Errorf("record %d level %v, want %v", i, r.Level, want[i])
		}
	}
}

// The asset server's line per served file is traffic, not news: it goes to DEBUG.
func TestAssetTrafficIsDebug(t *testing.T) {
	sink := &records{}
	mu.Lock()
	previous := global
	global = s99logger.New(sink, s99logger.Options{MinLevel: s99logger.LevelDebug})
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); global = previous; mu.Unlock() })

	Wails().Info("[AssetFileServerFS] Handling request", "url", "/_nuxt/app.js", "file", "_nuxt/app.js")
	Wails().Info("Asset server started")
	if len(sink.values) != 2 || sink.values[0].Level != s99logger.LevelDebug || sink.values[1].Level != s99logger.LevelInfo {
		t.Fatalf("records = %+v", sink.values)
	}
}
