package tuner

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"

	"smegg.me/smeggtuner/backend/core/dsp"
)

// run is one engine's state. dead is set by finish before it emits, so a run
// that ended on its own stops counting as running while stop can still reach it via s.run.
type run struct {
	engine       *dsp.Engine
	cancel       context.CancelFunc
	events       chan dsp.Measurement
	done         chan struct{}
	isMic        bool
	source       string
	dead         atomic.Bool
	drops        atomic.Int64
	ctx          context.Context
	observations chan dsp.Measurement
}

// offer preserves observer order with bounded backpressure; only UI updates may be dropped.
func (r *run) offer(m dsp.Measurement) {
	select {
	case r.observations <- m:
	case <-r.ctx.Done():
		return
	}
	select {
	case r.events <- m:
	default:
		r.drops.Add(1)
	}
}

func (r *run) observe(observer func(dsp.Measurement)) {
	for m := range r.observations {
		observer(m)
	}
}

// reconfigure hands the engine a whole new config; the engine normalizes the layout of its copy in place,
// so it gets its own.
func (r *run) reconfigure(cfg dsp.EngineConfig) {
	cfg.Octaves = slices.Clone(cfg.Octaves)
	r.engine.Update(func(c *dsp.EngineConfig) { *c = cfg })
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
	if errors.Is(err, dsp.ErrSampleRate) {
		return ErrSampleRate.Key
	}
	if r.isMic {
		return ErrDeviceLost.Key
	}
	return ErrFileUnreadable.Key
}
