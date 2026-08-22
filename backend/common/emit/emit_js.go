//go:build js

// backend/common/emit/emit_js.go
package emit

import "sync/atomic"

var post atomic.Pointer[func(msg any)]

// SetPost installs the worker's message channel to the main thread.
func SetPost(fn func(msg any)) { post.Store(&fn) }

// Post sends one message to the main thread; dropped until SetPost runs.
func Post(msg any) {
	if fn := post.Load(); fn != nil {
		(*fn)(msg)
	}
}

// Emit sends an event message, the shape the main thread hands to dispatchWailsEvent.
func Emit(name string, data any) {
	Post(map[string]any{"type": "event", "name": name, "data": data})
}
