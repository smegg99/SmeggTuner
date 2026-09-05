// backend/core/audio/tone.go
package audio

import (
	"encoding/binary"
	"math"
	"time"
)

// The reference tone generator: a 48 kHz sine at half of int16 full scale, both channels equal,
// shaped by 0.5 s linear fades and phased as frame*w with no accumulation.
const (
	ToneRate = 48000
	// ToneFade is the fade in at the start and the fade out at the end or on stop.
	ToneFade = 500 * time.Millisecond

	toneFadeFrames = ToneRate / 2
	toneScale      = 16383
	toneChannels   = 2
	// toneMinFrames keeps a tone long enough for both fades.
	toneMinFrames = 2 * toneFadeFrames

	// A note-strip press (an extension) starts at once and sounds this long however short the click.
	pressAttackFrames = ToneRate / 50
	pressHoldFrames   = ToneRate / 4
	// A new press ends the tones still sounding over this fade, so quick clicks never pile up and clip.
	yieldFrames = ToneRate / 25
)

// toneVoice is one tone, rendered by frame index from its own start.
type toneVoice struct {
	w      float64 // radians per frame
	frames int64   // audible frames, both fades included
	attack int64   // fade-in frames
	hold   int64   // the earliest frame a stop fade may start at
	stop   int64   // frame the stop fade starts at; -1 while none was asked for
	cut    int64   // frame the short yield fade starts at; -1 while none was asked for
	next   int64   // the frame the next sample is
}

func newToneVoice(freq float64, dur time.Duration) *toneVoice {
	frames := max(toneMinFrames, int64(math.Round(dur.Seconds()*ToneRate)))
	return &toneVoice{w: 2 * math.Pi * freq / ToneRate, frames: frames, attack: toneFadeFrames, stop: -1, cut: -1}
}

// newPressVoice is a note-strip tone: a short attack and a hold, so a click is heard.
func newPressVoice(freq float64, dur time.Duration) *toneVoice {
	v := newToneVoice(freq, dur)
	v.attack, v.hold = pressAttackFrames, pressHoldFrames
	return v
}

// envelope is the gain at frame i: fade in, hold, fade out, and on stop a fade from the stop frame,
// never above the gain the tone already had.
func (v *toneVoice) envelope(i int64) float64 {
	var e float64
	switch {
	case i <= v.attack:
		e = float64(i) / float64(v.attack)
	case i <= v.frames-toneFadeFrames:
		e = 1
	case i <= v.frames:
		e = float64(v.frames-1-i) / toneFadeFrames
	}
	if v.stop >= 0 {
		e = min(e, float64(toneFadeFrames-1-(i-v.stop))/toneFadeFrames)
	}
	if v.cut >= 0 {
		e = min(e, float64(yieldFrames-1-(i-v.cut))/yieldFrames)
	}
	return max(0, e)
}

// sample is frame i at volume 0..1, truncated toward zero as the reference writes it.
func (v *toneVoice) sample(i int64, vol float64) int32 {
	return int32(vol * (toneScale * v.envelope(i) * math.Sin(float64(i)*v.w)))
}

// halt starts the stop fade at the next frame, or at the end of the hold.
func (v *toneVoice) halt() {
	if v.stop < 0 {
		v.stop = max(v.next, v.hold)
	}
}

// yield fades the voice out from the next frame over yieldFrames, from whatever gain it has.
func (v *toneVoice) yield() {
	if v.cut < 0 {
		v.cut = v.next
	}
}

// silent reports whether the voice has nothing left to sound.
func (v *toneVoice) silent() bool {
	return v.next > v.frames || (v.stop >= 0 && v.next >= v.stop+toneFadeFrames) || (v.cut >= 0 && v.next >= v.cut+yieldFrames)
}

// renderTone fills out with interleaved int16 stereo frames of the voices summed; a nil voice is silence.
func renderTone(out []byte, vol float64, voices ...*toneVoice) {
	for f := 0; f < len(out)/(2*toneChannels); f++ {
		var s int32
		for _, v := range voices {
			if v != nil {
				s += v.sample(v.next, vol)
				v.next++
			}
		}
		s = min(math.MaxInt16, max(math.MinInt16, s))
		for ch := range toneChannels {
			binary.LittleEndian.PutUint16(out[(f*toneChannels+ch)*2:], uint16(int16(s)))
		}
	}
}
