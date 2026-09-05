//go:build !js

// backend/core/audio/player.go
package audio

import (
	"errors"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
)

// TonePlayer plays the reference tone (tone.go) through the default output device at the nominal
// 48 kHz clock. The device opens on the first Play and stays open until Close. The data callback takes
// no locks: requests reach it through cmd and volBits, so mu may be held across malgo Init/Start/Uninit.
type TonePlayer struct {
	mu      sync.Mutex
	ctx     *malgo.AllocatedContext
	device  *malgo.Device
	cmd     atomic.Pointer[toneCommand] // the latest request, taken by the callback
	volBits atomic.Uint64               // math.Float64bits of the output level, 0..1
	// voice, fading and mix are touched only by the device callback.
	voice  *toneVoice
	fading []*toneVoice // every earlier voice until its fade has run; none is ever cut off mid-wave
	mix    []*toneVoice
}

// toneCommand starts a new tone, or with stop set fades the sounding one out; the latest one wins,
// except that a stop folds into a press the callback has not taken yet (released sets).
type toneCommand struct {
	freq     float64
	dur      time.Duration
	stop     bool
	press    bool
	released bool
}

func NewTonePlayer() (*TonePlayer, error) {
	ctx, err := newMalgoContext()
	if err != nil {
		return nil, err
	}
	p := &TonePlayer{ctx: ctx}
	p.volBits.Store(math.Float64bits(1))
	return p, nil
}

// SetVolume scales the half-scale tone by v, 0..1; a sounding tone changes level on the next buffer.
func (p *TonePlayer) SetVolume(v float64) {
	p.volBits.Store(math.Float64bits(min(1, max(0, v))))
}

// Play starts a tone of freq hertz sounding for dur, fades included; a sounding tone fades out under
// it. Returns immediately.
func (p *TonePlayer) Play(freq float64, dur time.Duration) error {
	return p.start(&toneCommand{freq: freq, dur: dur})
}

// Press is Play for the note strip: the tone starts at once and a Stop lets it sound its hold first.
func (p *TonePlayer) Press(freq float64, dur time.Duration) error {
	return p.start(&toneCommand{freq: freq, dur: dur, press: true})
}

func (p *TonePlayer) start(c *toneCommand) error {
	if !(c.freq > 0) || c.freq >= ToneRate/2 {
		return errors.New("tone player: frequency outside the output band")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ctx == nil {
		return errors.New("tone player closed")
	}
	p.cmd.Store(c)
	if p.device != nil {
		return nil
	}
	cfg := malgo.DefaultDeviceConfig(malgo.Playback)
	cfg.Playback.Format = malgo.FormatS16
	cfg.Playback.Channels = toneChannels
	cfg.SampleRate = ToneRate
	// malgo's low-latency default period is a few milliseconds, which a Go program with a GC cannot
	// always fill; each missed one is a crackle. 20 ms, three deep, still answers a click at once.
	cfg.PerformanceProfile = malgo.Conservative
	cfg.PeriodSizeInMilliseconds = 20
	cfg.Periods = 3
	dev, err := malgo.InitDevice(p.ctx.Context, cfg, malgo.DeviceCallbacks{
		Data: func(output, _ []byte, _ uint32) { p.fill(output) },
	})
	if err != nil {
		return err
	}
	if err := dev.Start(); err != nil {
		dev.Uninit()
		return err
	}
	p.device = dev
	return nil
}

// fill runs on the audio thread: it takes the latest request, then renders every voice still sounding.
func (p *TonePlayer) fill(output []byte) {
	if c := p.cmd.Swap(nil); c != nil {
		p.take(c)
	}
	p.mix = append(append(p.mix[:0], p.voice), p.fading...)
	renderTone(output, math.Float64frombits(p.volBits.Load()), p.mix...)
	p.fading = slices.DeleteFunc(p.fading, (*toneVoice).silent)
	if p.voice != nil && p.voice.silent() {
		p.voice = nil
	}
}

// take applies a request: a new tone moves the sounding one to the fading ones, a press makes them all
// yield, and a stop fades the sounding one.
func (p *TonePlayer) take(c *toneCommand) {
	if p.voice != nil {
		p.voice.halt()
	}
	if c.stop {
		return
	}
	if c.press {
		for _, v := range p.fading {
			v.yield()
		}
		if p.voice != nil {
			p.voice.yield()
		}
	}
	if p.voice != nil {
		p.fading = append(p.fading, p.voice)
	}
	if c.press {
		p.voice = newPressVoice(c.freq, c.dur)
	} else {
		p.voice = newToneVoice(c.freq, c.dur)
	}
	if c.released {
		p.voice.halt()
	}
}

// Stop fades the sounding tone out over ToneFade; the device stays ready for the next Play.
func (p *TonePlayer) Stop() {
	for {
		c := p.cmd.Load()
		if c == nil || !c.press {
			if p.cmd.CompareAndSwap(c, &toneCommand{stop: true}) {
				return
			}
			continue
		}
		released := *c
		released.released = true
		if p.cmd.CompareAndSwap(c, &released) {
			return
		}
	}
}

// Close releases the playback device and audio context. The player must not be used after.
func (p *TonePlayer) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.device != nil {
		p.device.Uninit()
		p.device = nil
	}
	if p.ctx != nil {
		_ = p.ctx.Uninit()
		p.ctx.Free()
		p.ctx = nil
	}
}
