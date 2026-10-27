// backend/web/dispatch.go

// Package web is the web edition's Go side: the engine in a browser worker, called through the generated Wails bindings.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"reflect"
)

// CallError is the error shape the Wails runtime throws: kind picks the JS error class, cause carries the Go error as JSON.
type CallError struct {
	Kind    string          `json:"kind"`
	Message string          `json:"message"`
	Cause   json.RawMessage `json:"cause,omitempty"`
}

func (e *CallError) Error() string { return e.Message }

// MethodID is the binding ID Wails gives "<pkgpath>.<Type>.<Method>" (FNV-32a, pkg/application/bindings.go).
func MethodID(fqn string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fqn))
	return h.Sum32()
}

var (
	errorType   = reflect.TypeFor[error]()
	contextType = reflect.TypeFor[context.Context]()
)

type method struct {
	fn         reflect.Value
	in         []reflect.Type
	hasContext bool
}

// Dispatcher calls bound service methods by their binding ID.
type Dispatcher struct{ methods map[uint32]method }

// NewDispatcher binds every exported method of each service, a pointer to a named struct.
func NewDispatcher(services ...any) *Dispatcher {
	d := &Dispatcher{methods: map[uint32]method{}}
	for _, svc := range services {
		v := reflect.ValueOf(svc)
		elem := v.Type().Elem()
		for i := range v.NumMethod() {
			m := v.Type().Method(i)
			fn := v.Method(i)
			bm := method{fn: fn}
			for j := range fn.Type().NumIn() {
				bm.in = append(bm.in, fn.Type().In(j))
			}
			if len(bm.in) > 0 && bm.in[0] == contextType {
				bm.hasContext, bm.in = true, bm.in[1:]
			}
			d.methods[MethodID(elem.PkgPath()+"."+elem.Name()+"."+m.Name)] = bm
		}
	}
	return d
}

// Has reports whether id is bound.
func (d *Dispatcher) Has(id uint32) bool { _, ok := d.methods[id]; return ok }

// Call decodes args, runs the method and encodes its result the way the Wails runtime does.
func (d *Dispatcher) Call(id uint32, args []json.RawMessage) (json.RawMessage, error) {
	m, ok := d.methods[id]
	if !ok {
		return nil, &CallError{Kind: "ReferenceError", Message: fmt.Sprintf("unknown bound method id %d", id)}
	}
	if len(args) != len(m.in) {
		return nil, &CallError{Kind: "TypeError", Message: fmt.Sprintf("got %d arguments, want %d", len(args), len(m.in))}
	}
	in := make([]reflect.Value, 0, len(args)+1)
	if m.hasContext {
		in = append(in, reflect.ValueOf(context.Background()))
	}
	for i, raw := range args {
		p := reflect.New(m.in[i])
		if err := json.Unmarshal(raw, p.Interface()); err != nil {
			return nil, &CallError{Kind: "TypeError", Message: fmt.Sprintf("could not parse argument #%d: %s", i, err)}
		}
		in = append(in, p.Elem())
	}
	var values []any
	var errs []error
	for _, out := range m.fn.Call(in) {
		if out.Type() == errorType {
			if !out.IsNil() {
				errs = append(errs, out.Interface().(error))
			}
			continue
		}
		values = append(values, out.Interface())
	}
	if len(errs) > 0 {
		return nil, runtimeError(errs)
	}
	var result any
	switch len(values) {
	case 0:
	case 1:
		result = values[0]
	default:
		result = values
	}
	return json.Marshal(result)
}

func runtimeError(errs []error) *CallError {
	ce := &CallError{Kind: "RuntimeError", Message: errors.Join(errs...).Error()}
	if len(errs) == 1 {
		if b, err := json.Marshal(errs[0]); err == nil {
			ce.Cause = b
		}
	}
	return ce
}
