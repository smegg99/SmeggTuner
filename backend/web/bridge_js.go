//go:build js

// backend/web/bridge_js.go
package web

import (
	"encoding/json"
	"syscall/js"

	"smegg.me/smeggtuner/backend/common/emit"
)

// Serve exposes the dispatcher and the source to the worker as globalThis.smeggtuner, then calls smeggtunerReady.
func Serve(d *Dispatcher, src *BrowserSource) {
	emit.SetPost(func(msg any) {
		b, err := json.Marshal(msg)
		if err != nil {
			return
		}
		js.Global().Call("smeggtunerPost", string(b))
	})
	api := js.Global().Get("Object").New()
	api.Set("call", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return callJSON(d, uint32(args[0].Int()), args[1].String())
	}))
	api.Set("push", js.FuncOf(func(_ js.Value, args []js.Value) any {
		f := args[0]
		view := js.Global().Get("Uint8Array").New(f.Get("buffer"), f.Get("byteOffset"), f.Get("byteLength"))
		b := make([]byte, view.Get("length").Int())
		js.CopyBytesToGo(b, view)
		src.Push(floatsFromBytes(b), args[1].Int())
		return nil
	}))
	js.Global().Set("smeggtuner", api)
	js.Global().Call("smeggtunerReady")
}
