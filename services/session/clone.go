// Nothing handed out may share memory with the active session: Wails marshals on the calling
// goroutine while takes are appended from the engine's. Exception: the goal curve is handed out by
// pointer (services/tuner reads it live) and never edited in place - mutators clone, edit, swap.
package session

import (
	"errors"
	"io/fs"
	"os"

	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
)

func cloneCurve(c *target.Curve) *target.Curve {
	if c == nil {
		return nil
	}
	out := *c
	out.Anchors = make([]target.Anchor, len(c.Anchors))
	for i, a := range c.Anchors {
		out.Anchors[i] = target.Anchor{
			Note:  a.Note,
			Reeds: append([]float64(nil), a.Reeds...),
		}
	}
	return &out
}

// cloneInstrument deep-copies down to banks so a snapshot never shares the live session's slices.
func cloneInstrument(i coresession.Instrument) coresession.Instrument {
	i.Banks = append([]coresession.Bank(nil), i.Banks...)
	i.Registers = append([]coresession.Register(nil), i.Registers...)
	for n := range i.Registers {
		i.Registers[n].Banks = append([]coresession.Bank(nil), i.Registers[n].Banks...)
	}
	return i
}

func cloneTakes(ts []coresession.Take) []coresession.Take {
	out := make([]coresession.Take, len(ts))
	for i, t := range ts {
		t.Reeds = append([]dsp.ReedMeasure(nil), t.Reeds...)
		t.Beats = append([]dsp.BeatMeasure(nil), t.Beats...)
		out[i] = t
	}
	return out
}

func cloneSession(s *coresession.Session) *coresession.Session {
	out := *s
	out.Instrument = cloneInstrument(s.Instrument)
	out.Curve = cloneCurve(s.Curve)
	out.Takes = cloneTakes(s.Takes)
	return &out
}

// dspReed is a reed with a deviation and no frequency (typed, not heard).
func dspReed(cents float64) dsp.ReedMeasure {
	return dsp.ReedMeasure{DevCents: cents}
}

func isNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist)
}
