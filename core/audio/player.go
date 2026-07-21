// core/audio/player.go
package audio

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
)

const (
	playerSampleRate = 48000
	// 20 ms attack/release ramp avoids clicks.
	playerRampStep = 1.0 / (0.020 * playerSampleRate)
	playerAmp      = 0.4
)

// TonePlayer plays a sine reference tone through the default output device. Play is
// non-blocking; the device is created lazily on first Play and kept until Close. The data
// callback reads freqBits/untilNano/volBits atomically and takes no locks, so mu may be held
// across malgo Init/Start/Uninit (Uninit joins the audio thread).
type TonePlayer struct {
	mu        sync.Mutex
	ctx       *malgo.AllocatedContext
	device    *malgo.Device
	freqBits  atomic.Uint64 // math.Float64bits of the target frequency
	untilNano atomic.Int64  // deadline as UnixNano; 0 or past = silence
	volBits   atomic.Uint64 // math.Float64bits of the output level, 0..1
	// phase and gain are touched only by the device callback.
	phase float64
	gain  float64
}

func NewTonePlayer() (*TonePlayer, error) {
	ctx, err := newMalgoContext()
	if err != nil {
		return nil, err
	}
	p := &TonePlayer{ctx: ctx}
	p.volBits.Store(math.Float64bits(1)) // full by default; SetVolume turns it down
	return p, nil
}

// SetVolume sets the tone's output level, 0..1, on top of the fixed headroom (playerAmp).
// Read atomically by the callback, so a sounding tone changes level without a gap.
func (p *TonePlayer) SetVolume(v float64) {
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	p.volBits.Store(math.Float64bits(v))
}

// Play starts (or retargets) a tone of freq hertz detuned by ppm parts per million, lasting
// dur. Returns immediately; a running tone picks up the new frequency and deadline with no gap.
func (p *TonePlayer) Play(freq float64, dur time.Duration, ppm float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ctx == nil {
		return errors.New("tone player closed")
	}
	p.freqBits.Store(math.Float64bits(freq * (1 + ppm*1e-6)))
	p.untilNano.Store(time.Now().Add(dur).UnixNano())
	if p.device != nil {
		return nil // already running; new freq/deadline picked up by callback
	}
	cfg := malgo.DefaultDeviceConfig(malgo.Playback)
	cfg.Playback.Format = malgo.FormatF32
	cfg.Playback.Channels = 1
	cfg.SampleRate = playerSampleRate
	dev, err := malgo.InitDevice(p.ctx.Context, cfg, malgo.DeviceCallbacks{
		Data: func(output, _ []byte, frameCount uint32) {
			freq := math.Float64frombits(p.freqBits.Load())
			vol := math.Float64frombits(p.volBits.Load())
			active := time.Now().UnixNano() < p.untilNano.Load()
			for i := uint32(0); i < frameCount; i++ {
				if active {
					if p.gain < 1 {
						p.gain += playerRampStep
					}
				} else if p.gain > 0 {
					p.gain -= playerRampStep
				}
				if p.gain < 0 {
					p.gain = 0
				}
				v := playerAmp * vol * p.gain * math.Sin(p.phase)
				p.phase += 2 * math.Pi * freq / playerSampleRate
				if p.phase > 2*math.Pi {
					p.phase -= 2 * math.Pi
				}
				binary.LittleEndian.PutUint32(output[i*4:], math.Float32bits(float32(v)))
			}
		},
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

// Stop ends the current tone; the release ramp still runs so it fades instead of clicking.
// The device stays ready for the next Play.
func (p *TonePlayer) Stop() {
	p.untilNano.Store(0)
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
