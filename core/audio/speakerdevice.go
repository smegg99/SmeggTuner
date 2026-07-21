// core/audio/speakerdevice.go
package audio

import (
	"encoding/binary"
	"math"

	"github.com/gen2brain/malgo"
)

// Reset throws away whatever is queued and re-arms the prefill, so the next run starts with a
// full cushion. Without it, a Stop drained the ring but left the device started, so the next
// play skipped the prefill and broke up from the first block. Stopping the device joins the
// audio thread, which is the only reason it is legal to touch read from here.
func (s *Speaker) Reset() {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	if s.device != nil && s.started {
		_ = s.device.Stop()
		s.started = false
	}
	s.playing.Store(false)
	s.read.Store(0)
	s.write.Store(0)
}

// Write queues a block for the card, opening the device on the first one. It never blocks and
// never fails loudly: a machine with no output device must still be able to measure a recording.
func (s *Speaker) Write(samples []float32, sampleRate int) {
	// NOT skipped when muted - a muted speaker still plays (silence), keeping the played-count
	// clock moving. See SetMuted.
	if len(samples) == 0 || sampleRate <= 0 {
		return
	}
	if !s.open(sampleRate) {
		return
	}

	buf := s.buf.Load()
	if buf == nil || len(buf.data) == 0 {
		return
	}
	size := int64(len(buf.data))

	// Only this goroutine moves write, so it is stored once, at the end: the consumer sees a
	// whole block appear, never half of one.
	w := s.write.Load()

	for _, v := range samples {
		buf.data[w%size] = v
		w++
	}
	s.write.Store(w)

	// The producer lapped the consumer: push it forward to the oldest surviving sample. The
	// CAS cannot drag the consumer backwards if it moved while we looked.
	if r := s.read.Load(); w-r > size {
		s.drops.Add(w - r - size)
		s.read.CompareAndSwap(r, w-size)
	}

	s.startWhenFilled(w-s.read.Load(), sampleRate)
}

// startWhenFilled lets the card have the device once the cushion is full, so its first pull
// cannot land on an empty ring. See speakerPrefillSeconds.
func (s *Speaker) startWhenFilled(buffered int64, sampleRate int) {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	if s.device == nil || s.started {
		return
	}
	if buffered < int64(float64(sampleRate)*speakerPrefillSeconds) {
		return
	}
	if err := s.device.Start(); err != nil {
		return
	}
	s.started = true
	s.playing.Store(true)
}

// open brings up a device at the RECORDING's own rate and reports whether there is one to play
// through. The rate is the file's, not a fixed 48 kHz: resampling to play would be inventing
// audio, the one thing this app may never do. A file with a new rate gets a new device and ring.
func (s *Speaker) open(sampleRate int) bool {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	if s.ctx == nil {
		return false // closed
	}
	if s.device != nil && s.rate == sampleRate {
		return true
	}

	// Uninit joins the audio thread, so once it returns no callback is running and the ring
	// can be replaced safely.
	if s.device != nil {
		s.device.Uninit()
		s.device = nil
		s.started = false
		s.playing.Store(false)
	}

	cfg := malgo.DefaultDeviceConfig(malgo.Playback)
	cfg.Playback.Format = malgo.FormatF32
	cfg.Playback.Channels = 1
	cfg.SampleRate = uint32(sampleRate)

	// Conservative, with a period big enough to survive a scheduler. malgo's low-latency
	// default asks for a few-millisecond period that a Go program with a GC cannot service
	// reliably, and each dropped period is a hole. We are drawing a needle over a recording,
	// so latency is worth nothing here and robustness everything: 30 ms, four deep.
	cfg.PerformanceProfile = malgo.Conservative
	cfg.PeriodSizeInMilliseconds = 30
	cfg.Periods = 4

	dev, err := malgo.InitDevice(s.ctx.Context, cfg, malgo.DeviceCallbacks{Data: s.pull})
	if err != nil {
		return false
	}

	s.buf.Store(&ring{data: make([]float32, int(float64(sampleRate)*speakerBufferSeconds))})
	s.read.Store(0)
	s.write.Store(0)

	// NOT started here; startWhenFilled does that once the ring holds enough.
	s.device = dev
	s.rate = sampleRate
	return true
}

// pull runs on the AUDIO THREAD. It takes no locks and allocates nothing. An underrun is
// written as silence, never as a repeat of the last block.
func (s *Speaker) pull(output, _ []byte, frameCount uint32) {
	buf := s.buf.Load()
	muted := s.muted.Load()

	r := s.read.Load() // only this thread moves it
	w := s.write.Load()

	// Mute is a gain of zero: the ring is consumed at the same rate either way, so the
	// played-count keeps its meaning. See SetMuted.
	gain := float32(math.Float64frombits(s.gain.Load()))
	if muted {
		gain = 0
	}

	for i := uint32(0); i < frameCount; i++ {
		var v float32

		switch {
		case buf == nil || len(buf.data) == 0:
			// nothing to play through yet
		case r < w:
			v = buf.data[r%int64(len(buf.data))] * gain
			r++
		default:
			s.underruns.Add(1)
		}

		binary.LittleEndian.PutUint32(output[i*4:], math.Float32bits(v))
	}

	s.read.Store(r)
}

func (s *Speaker) Close() error {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	if s.device != nil {
		s.device.Uninit() // joins the audio thread; pull cannot be running after this
		s.device = nil
		s.started = false
		s.playing.Store(false)
	}
	if s.ctx == nil {
		return nil
	}

	err := s.ctx.Uninit()
	s.ctx.Free()
	s.ctx = nil
	return err
}
