// backend/core/audio/endpoint/wasapi_windows.go
package endpoint

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const supported = true

var (
	ole32            = syscall.NewLazyDLL("ole32.dll")
	coInitializeEx   = ole32.NewProc("CoInitializeEx")
	coUninitialize   = ole32.NewProc("CoUninitialize")
	coCreateInstance = ole32.NewProc("CoCreateInstance")
)

type guid struct {
	d1     uint32
	d2, d3 uint16
	d4     [8]byte
}

var (
	clsidMMDeviceEnumerator = guid{0xBCDE0395, 0xE52F, 0x467C, [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}
	iidIMMDeviceEnumerator  = guid{0xA95664D2, 0x9614, 0x4F35, [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6}}
	iidIAudioEndpointVolume = guid{0x5CDF2C82, 0x841E, 0x4546, [8]byte{0x97, 0x22, 0x0C, 0xF7, 0x40, 0x78, 0x22, 0x9A}}
)

const (
	coinitMultithreaded = 0
	clsctxAll           = 23
	eCapture, eConsole  = 1, 0

	// vtable slots
	slotRelease                 = 2
	slotGetDefaultAudioEndpoint = 4 // IMMDeviceEnumerator
	slotGetDevice               = 5
	slotActivate                = 3 // IMMDevice
	slotSetMasterScalar         = 7 // IAudioEndpointVolume
	slotGetMasterScalar         = 9
	slotSetMute                 = 14
	slotGetMute                 = 15
)

// call invokes a vtable slot of a COM object; a failed HRESULT is an error.
//
//go:uintptrescapes
func call(obj unsafe.Pointer, slot uintptr, args ...uintptr) error {
	vtbl := *(*unsafe.Pointer)(obj)
	fn := *(*uintptr)(unsafe.Add(vtbl, slot*unsafe.Sizeof(uintptr(0))))
	hr, _, _ := syscall.SyscallN(fn, append([]uintptr{uintptr(obj)}, args...)...)
	if int32(hr) < 0 {
		return fmt.Errorf("COM slot %d: HRESULT 0x%08X", slot, uint32(hr))
	}
	return nil
}

func release(obj unsafe.Pointer) {
	if obj != nil {
		_ = call(obj, slotRelease)
	}
}

// wasapiVolume is one device's IAudioEndpointVolume, used only on its own COM thread.
type wasapiVolume struct {
	jobs chan func()
	done chan struct{}
	vol  unsafe.Pointer
}

// open activates the endpoint whose id malgo keeps as a NUL-ended UTF-16 string.
func open(raw []byte) (Endpoint, error) {
	w := &wasapiVolume{jobs: make(chan func()), done: make(chan struct{})}
	ready := make(chan error, 1)
	go w.serve(endpointID(raw), ready)
	if err := <-ready; err != nil {
		return nil, err
	}
	return w, nil
}

func endpointID(raw []byte) string {
	units := make([]uint16, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		u := uint16(raw[i]) | uint16(raw[i+1])<<8
		if u == 0 {
			break
		}
		units = append(units, u)
	}
	return string(utf16.Decode(units))
}

// serve owns a COM apartment thread for the endpoint's lifetime.
func (w *wasapiVolume) serve(id string, ready chan<- error) {
	defer close(w.done)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if hr, _, _ := coInitializeEx.Call(0, coinitMultithreaded); int32(hr) < 0 {
		ready <- fmt.Errorf("CoInitializeEx: HRESULT 0x%08X", uint32(hr))
		return
	}
	defer coUninitialize.Call()
	vol, err := activate(id)
	if err != nil {
		ready <- err
		return
	}
	defer release(vol)
	w.vol = vol
	ready <- nil
	for job := range w.jobs {
		job()
	}
}

func activate(id string) (unsafe.Pointer, error) {
	var enum unsafe.Pointer
	hr, _, _ := coCreateInstance.Call(uintptr(unsafe.Pointer(&clsidMMDeviceEnumerator)), 0, clsctxAll,
		uintptr(unsafe.Pointer(&iidIMMDeviceEnumerator)), uintptr(unsafe.Pointer(&enum)))
	if int32(hr) < 0 {
		return nil, fmt.Errorf("MMDeviceEnumerator: HRESULT 0x%08X", uint32(hr))
	}
	defer release(enum)
	var dev unsafe.Pointer
	if id == "" {
		err := call(enum, slotGetDefaultAudioEndpoint, eCapture, eConsole, uintptr(unsafe.Pointer(&dev)))
		if err != nil {
			return nil, fmt.Errorf("default capture endpoint: %w", err)
		}
	} else {
		wide, err := syscall.UTF16PtrFromString(id)
		if err != nil {
			return nil, err
		}
		if err := call(enum, slotGetDevice, uintptr(unsafe.Pointer(wide)), uintptr(unsafe.Pointer(&dev))); err != nil {
			return nil, fmt.Errorf("endpoint %q: %w", id, err)
		}
	}
	defer release(dev)
	var vol unsafe.Pointer
	if err := call(dev, slotActivate, uintptr(unsafe.Pointer(&iidIAudioEndpointVolume)), clsctxAll, 0,
		uintptr(unsafe.Pointer(&vol))); err != nil {
		return nil, fmt.Errorf("endpoint volume: %w", err)
	}
	return vol, nil
}

// do runs f on the COM thread.
func (w *wasapiVolume) do(f func() error) error {
	errc := make(chan error, 1)
	select {
	case w.jobs <- func() { errc <- f() }:
		return <-errc
	case <-w.done:
		return errors.New("endpoint closed")
	}
}

func (w *wasapiVolume) Level() (float64, error) {
	var level float32
	err := w.do(func() error { return call(w.vol, slotGetMasterScalar, uintptr(unsafe.Pointer(&level))) })
	return float64(level), err
}

// SetLevel passes the scalar as float bits; Go's Windows call path mirrors integer arguments into the
// float registers, where the callee reads it.
func (w *wasapiVolume) SetLevel(level float64) error {
	bits := uintptr(math.Float32bits(float32(min(1, max(0, level)))))
	return w.do(func() error { return call(w.vol, slotSetMasterScalar, bits, 0) })
}

func (w *wasapiVolume) Muted() (bool, error) {
	var muted int32
	err := w.do(func() error { return call(w.vol, slotGetMute, uintptr(unsafe.Pointer(&muted))) })
	return muted != 0, err
}

func (w *wasapiVolume) SetMuted(muted bool) error {
	var flag uintptr
	if muted {
		flag = 1
	}
	return w.do(func() error { return call(w.vol, slotSetMute, flag, 0) })
}

func (w *wasapiVolume) Close() error {
	close(w.jobs)
	<-w.done
	return nil
}
