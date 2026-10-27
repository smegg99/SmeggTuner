// backend/web/dispatch_test.go
package web

import (
	"encoding/json"
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/common/serviceerror"
	"smegg.me/smeggtuner/backend/services/tuner"
)

type fakeService struct{}

var errFake = &serviceerror.ServiceError{Key: "fake.error"}

func (*fakeService) Double(n int) (int, error) { return 2 * n, nil }
func (*fakeService) Fail() error               { return errFake }
func (*fakeService) Nothing()                  {}

func fakeID(method string) uint32 {
	return MethodID("smegg.me/smeggtuner/backend/web.fakeService." + method)
}

func TestDispatcherIDsMatchGeneratedBindings(t *testing.T) {
	d := NewDispatcher(tuner.New(nil, nil, nil))
	// Values from frontend/bindings/smegg.me/smeggtuner/backend/services/tuner/service.ts.
	for name, id := range map[string]uint32{"Freeze": 2505020343, "AdoptCorrection": 1447629746} {
		if !d.Has(id) {
			t.Fatalf("%s: id %d not registered", name, id)
		}
	}
}

func TestDispatcherReturnsValue(t *testing.T) {
	d := NewDispatcher(&fakeService{})
	got, err := d.Call(fakeID("Double"), []json.RawMessage{json.RawMessage("21")})
	if err != nil || string(got) != "42" {
		t.Fatalf("Double(21) = %s, %v; want 42", got, err)
	}
	got, err = d.Call(fakeID("Nothing"), nil)
	if err != nil || string(got) != "null" {
		t.Fatalf("Nothing() = %s, %v; want null", got, err)
	}
}

func TestDispatcherReturnsErrorAsRuntimeError(t *testing.T) {
	d := NewDispatcher(&fakeService{})
	_, err := d.Call(fakeID("Fail"), nil)
	var ce *CallError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want a *CallError", err)
	}
	if ce.Kind != "RuntimeError" || ce.Message != "fake.error" || string(ce.Cause) != `{"key":"fake.error"}` {
		t.Fatalf("CallError = %+v (cause %s)", ce, ce.Cause)
	}
}

func TestDispatcherUnknownID(t *testing.T) {
	_, err := NewDispatcher(&fakeService{}).Call(1, nil)
	var ce *CallError
	if !errors.As(err, &ce) || ce.Kind != "ReferenceError" {
		t.Fatalf("err = %v, want a ReferenceError", err)
	}
}

func TestDispatcherBadArgumentIsTypeError(t *testing.T) {
	_, err := NewDispatcher(&fakeService{}).Call(fakeID("Double"), []json.RawMessage{json.RawMessage(`"x"`)})
	var ce *CallError
	if !errors.As(err, &ce) || ce.Kind != "TypeError" {
		t.Fatalf("err = %v, want a TypeError", err)
	}
}
