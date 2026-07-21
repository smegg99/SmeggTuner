package tuner

import (
	"context"
	"errors"
	"sync/atomic"

	"smegg.me/smeggtuner/core/dsp"
	audiosvc "smegg.me/smeggtuner/services/audio"
)

// run is one engine's state. dead is set by finish before it emits, so a run
// that ended on its own stops counting as running while stop can still reach it via s.run.
type run struct {
	engine *dsp.Engine
	cancel context.CancelFunc
	events chan dsp.Measurement
	done   chan struct{}
	isMic  bool
	source string
	dead   atomic.Bool
	drops  atomic.Int64
}

// offer hands a Measurement to the emitter without blocking the DSP goroutine;
// a full queue drops and counts the newest (another arrives in ~85 ms).
func (r *run) offer(m dsp.Measurement) {
	select {
	case r.events <- m:
	default:
		r.drops.Add(1)
	}
}

// Recorder receives the measurement stream besides the frontend; nil is legal.
type Recorder interface {
	OnMeasurement(dsp.Measurement)
}

// errorKey maps the engine's exit error to an i18n key: a keyed error keeps its
// key, a raw one is device-lost on the mic path and file-unreadable on the file path.
func (r *run) errorKey(err error) string {
	var se *ServiceError
	if errors.As(err, &se) {
		return se.Key
	}
	if r.isMic {
		return ErrDeviceLost.Key
	}
	return audiosvc.ErrFileUnreadable.Key
}
