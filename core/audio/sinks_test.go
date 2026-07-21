// core/audio/sinks_test.go
package audio

import (
	"sync"
	"time"
)

// recordingSink stands in for the speakers so this can run on a machine with no sound card.
// It "hears" a block the instant it is given one: nothing is ever queued.
type recordingSink struct {
	mu    sync.Mutex
	heard []float32
	rate  int
}

func (r *recordingSink) Write(samples []float32, sampleRate int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.heard = append(r.heard, samples...)
	r.rate = sampleRate
}

func (r *recordingSink) Played() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.heard)
}

func (r *recordingSink) Playing() bool { return true } // no cushion to fill

func (r *recordingSink) Buffered() int { return 0 }

func (r *recordingSink) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.heard = nil
}

func (r *recordingSink) Close() error { return nil }

func (r *recordingSink) all() ([]float32, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]float32(nil), r.heard...), r.rate
}

// resettingSink counts the flushes.
type resettingSink struct {
	recordingSink
	mu     sync.Mutex
	resets int
}

func (r *resettingSink) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resets++
}

func (r *resettingSink) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.resets
}

// drainingSink is a speaker that plays audio out over time rather than instantly; the
// playhead is looked up against how much it has actually played.
type drainingSink struct {
	mu     sync.Mutex
	queued int
	played int
}

func (d *drainingSink) Write(samples []float32, _ int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queued += len(samples)
}

func (d *drainingSink) Played() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.played
}

func (d *drainingSink) Playing() bool { return true }

func (d *drainingSink) Buffered() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.queued - d.played
}

func (d *drainingSink) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queued, d.played = 0, 0
}

func (d *drainingSink) Close() error { return nil }

// play advances the card by n samples, but never past what it has been given.
func (d *drainingSink) play(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.played += n; d.played > d.queued {
		d.played = d.queued
	}
}

// silentSink is a speaker handed audio but not playing it yet: what a real one does for the
// third of a second after Play while its cushion fills.
type silentSink struct {
	drainingSink
	on bool
}

func (s *silentSink) Playing() bool { return s.on }

// cushionSink is a speaker with a real cushion: it will not start playing until handed a
// third of a second of audio, exactly like the sound card.
type cushionSink struct {
	drainingSink
	rate int
}

func (c *cushionSink) Playing() bool {
	return c.Buffered() >= int(float64(c.rate)*speakerPrefillSeconds)
}

// leadOverRealtime is how far ahead of the wall clock the handover has run.
func (s *FileSource) leadOverRealtime(since time.Time) time.Duration {
	s.mu.Lock()
	handed := s.pos - s.from
	s.mu.Unlock()

	return s.timeAt(handed) - time.Since(since)
}
