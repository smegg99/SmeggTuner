//go:build !js

// backend/core/audio/tone_test.go
package audio

import (
	"encoding/binary"
	"math"
	"testing"
	"time"
)

const referenceDuration = 43 * time.Second

// frames renders n stereo frames and returns the left and right channels.
func frames(n int, vol float64, voices ...*toneVoice) (left, right []int16) {
	out := make([]byte, n*toneChannels*2)
	renderTone(out, vol, voices...)
	for f := range n {
		left = append(left, int16(binary.LittleEndian.Uint16(out[f*4:])))
		right = append(right, int16(binary.LittleEndian.Uint16(out[f*4+2:])))
	}
	return left, right
}

// The 43 s tone has the reference's frame layout: 0.5 s in, 42 s held, 0.5 s out, ending at 2,064,000.
func TestToneEnvelopeMatchesTheReference(t *testing.T) {
	v := newToneVoice(440, referenceDuration)
	if v.frames != 2064000 {
		t.Fatalf("audible frames = %d, want 2064000", v.frames)
	}
	for _, tc := range []struct {
		frame int64
		want  float64
	}{
		{0, 0}, {12000, 0.5}, {24000, 1}, {24001, 1}, {2040000, 1},
		{2040001, 23998.0 / 24000}, {2063999, 0}, {2064000, 0}, {2100000, 0},
	} {
		if got := v.envelope(tc.frame); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("envelope(%d) = %v, want %v", tc.frame, got, tc.want)
		}
	}
}

// A stop fades from just under full to zero over 0.5 s, and never lifts a tone that is still fading in.
func TestToneStopFades(t *testing.T) {
	v := newToneVoice(440, referenceDuration)
	v.next = 100000
	v.halt()
	if got := v.envelope(100000); got != 23999.0/24000 {
		t.Fatalf("first stop frame = %v, want 23999/24000", got)
	}
	if got := v.envelope(100000 + 23999); got != 0 {
		t.Fatalf("last stop frame = %v, want 0", got)
	}
	v.next = 100000 + toneFadeFrames
	if !v.silent() {
		t.Fatal("a stopped voice is silent once its fade has run")
	}

	early := newToneVoice(440, referenceDuration)
	early.next = 6000
	early.halt()
	if got := early.envelope(6000); got != 0.25 {
		t.Fatalf("stop during the fade in = %v, want the 0.25 it already had", got)
	}
}

// Both channels carry the same sample: half scale times the volume, truncated, phased as frame*w.
func TestToneSamples(t *testing.T) {
	v := newToneVoice(441, referenceDuration)
	v.next = 30000
	left, right := frames(2000, 1, v)
	for k := range left {
		i := int64(30000 + k)
		want := int16(math.Trunc(toneScale * math.Sin(float64(i)*v.w)))
		if left[k] != want || right[k] != want {
			t.Fatalf("frame %d = %d / %d, want %d on both channels", i, left[k], right[k], want)
		}
	}

	quiet := newToneVoice(441, referenceDuration)
	quiet.next = 30000
	q, _ := frames(2000, 0.5, quiet)
	for k := range q {
		if math.Abs(float64(q[k])-float64(left[k])/2) > 1 {
			t.Fatalf("frame %d at half volume = %d, want about %d", k, q[k], left[k]/2)
		}
	}
}

// The steady part is a sine of the asked frequency: no clock correction moves it.
func TestToneFrequency(t *testing.T) {
	v := newToneVoice(440, referenceDuration)
	v.next = toneFadeFrames + 1
	left, _ := frames(ToneRate, 1, v)
	crossings := 0
	for k := 1; k < len(left); k++ {
		if (left[k-1] < 0) != (left[k] < 0) {
			crossings++
		}
	}
	if crossings < 879 || crossings > 881 {
		t.Fatalf("zero crossings in 1 s = %d, want 880 for 440 Hz", crossings)
	}
}

// A tone too short for both fades is lengthened to them; the end is silence.
func TestToneShortAndEnded(t *testing.T) {
	v := newToneVoice(440, time.Millisecond)
	if v.frames != toneMinFrames {
		t.Fatalf("frames = %d, want %d", v.frames, toneMinFrames)
	}
	v.next = v.frames + 1
	left, _ := frames(100, 1, v)
	for _, s := range left {
		if s != 0 {
			t.Fatal("an ended tone must be silent")
		}
	}
	if !v.silent() {
		t.Fatal("an ended voice is silent")
	}
}

// The callback takes the latest request: a new tone moves the sounding one to its fade, a stop fades it.
func TestTonePlayerRequests(t *testing.T) {
	p := &TonePlayer{}
	p.volBits.Store(math.Float64bits(1))
	buf := make([]byte, 480*toneChannels*2)

	p.cmd.Store(&toneCommand{freq: 440, dur: referenceDuration})
	p.fill(buf)
	first := p.voice
	if first == nil || first.next != 480 {
		t.Fatal("a play request starts a voice at frame 0")
	}

	p.cmd.Store(&toneCommand{freq: 880, dur: referenceDuration})
	p.fill(buf)
	if len(p.fading) != 1 || p.fading[0] != first || first.stop != 480 || p.voice == first {
		t.Fatal("a new tone takes over and the old one fades from where it was")
	}

	p.Stop()
	p.fill(buf)
	if p.voice == nil || p.voice.stop != 480 {
		t.Fatal("a stop fades the sounding tone from its next frame")
	}
	for range 60 {
		p.fill(buf)
	}
	if p.voice != nil || len(p.fading) != 0 {
		t.Fatal("both voices are dropped once their fades have run")
	}

	p.cmd.Store(&toneCommand{freq: 440, dur: referenceDuration})
	p.Stop()
	p.fill(buf)
	if p.voice != nil {
		t.Fatal("a stop after a play that never started leaves silence")
	}
}

// Two voices at full level sum without wrapping.
func TestToneVoicesSumWithinRange(t *testing.T) {
	a, b := newToneVoice(440, referenceDuration), newToneVoice(440, referenceDuration)
	a.next, b.next = 30000, 30000
	left, _ := frames(200, 1, a, b)
	for _, s := range left {
		if s > 2*toneScale || s < -2*toneScale {
			t.Fatalf("summed sample %d outside two half scales", s)
		}
	}
}

// A press on the note strip is at full level within its short attack and sounds a quarter second
// however soon it is let go, then fades as any tone does.
func TestAPressIsHeardOnAClick(t *testing.T) {
	v := newPressVoice(440, referenceDuration)
	if got := v.envelope(pressAttackFrames); got != 1 {
		t.Fatalf("envelope at the end of the attack = %v, want 1", got)
	}
	v.next = 2400 // let go after 50 ms
	v.halt()
	if v.stop != pressHoldFrames {
		t.Fatalf("stop fade starts at frame %d, want the hold end %d", v.stop, pressHoldFrames)
	}
	if got := v.envelope(pressHoldFrames - 1); got != 1 {
		t.Fatalf("envelope just before the hold ends = %v, want 1", got)
	}
}

// A click's release can reach the player before the audio thread has taken the press; the press
// still sounds, with the release folded into it.
func TestAReleaseBeforeThePressStartsStillSoundsIt(t *testing.T) {
	p := &TonePlayer{}
	p.volBits.Store(math.Float64bits(1))
	buf := make([]byte, 480*toneChannels*2)

	p.cmd.Store(&toneCommand{freq: 440, dur: referenceDuration, press: true})
	p.Stop()
	p.fill(buf)
	if p.voice == nil || p.voice.stop != pressHoldFrames {
		t.Fatal("a press released before it started sounds its hold and then fades")
	}
}

// Quick clicks on the note strip: every earlier tone yields and fades, none is cut off mid-wave, so the
// output never jumps by more than a sine's own step plus a fade's.
func TestQuickPressesNeverCutATone(t *testing.T) {
	p := &TonePlayer{}
	p.volBits.Store(math.Float64bits(1))
	buf := make([]byte, 480*toneChannels*2)
	var out []int16
	for i, f := range []float64{440, 494, 523, 587, 659, 698, 784} {
		p.cmd.Store(&toneCommand{freq: f, dur: referenceDuration, press: true})
		if i%2 == 1 {
			p.Stop()
		}
		for range 3 { // 30 ms apart
			p.fill(buf)
			for k := 0; k < len(buf); k += toneChannels * 2 {
				out = append(out, int16(binary.LittleEndian.Uint16(buf[k:])))
			}
		}
	}
	step := toneScale * 2 * math.Pi * 784 / ToneRate
	for k := 1; k < len(out); k++ {
		if d := math.Abs(float64(out[k]) - float64(out[k-1])); d > 2*step+toneScale/float64(yieldFrames)*2 {
			t.Fatalf("frame %d jumps by %v", k, d)
		}
	}
	if len(p.fading) > 2 {
		t.Fatalf("%d tones still fading after 30 ms clicks", len(p.fading))
	}
}
