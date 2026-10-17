package tuner

import (
	"slices"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/core/dsp"
)

// Mode is what the tuner measures while no session is open.
type Mode string

const (
	ModeTones   Mode = "tones"
	ModeOctaves Mode = "octaves"
	ModeChords  Mode = "chords"
)

func (m Mode) valid() bool { return m == ModeTones || m == ModeOctaves || m == ModeChords }

// boxes is a set of small positive ids as a bit mask, so a panel stays comparable.
type boxes uint8

func boxesOf(ids []int) boxes {
	var b boxes
	for _, id := range ids {
		b = b.with(id, true)
	}
	return b
}

func (b boxes) with(id int, on bool) boxes {
	if on {
		return b | 1<<(id-1)
	}
	return b &^ (1 << (id - 1))
}

func (b boxes) ids() []int {
	var out []int
	for id := 1; id <= 8; id++ {
		if b&(1<<(id-1)) != 0 {
			out = append(out, id)
		}
	}
	return out
}

// panel is the tuner's own measuring setup while no session is open. octaves holds the upper octave
// boxes as reference box ids: 2 is +12, 3 is +24; the base is always on.
type panel struct {
	mode    Mode
	reeds   boxes
	octaves boxes
}

var defaultPanel = panel{mode: ModeTones, reeds: boxesOf([]int{1}), octaves: boxesOf([]int{2})}

func readPanel(t appconfig.Tuner) panel {
	pn := panel{mode: Mode(t.Mode), reeds: boxesOf(t.ReedBoxes)}
	for _, off := range t.OctaveOffsets {
		pn.octaves = pn.octaves.with(1+off/12, true)
	}
	return pn
}

// offsets is the upper octave boxes in semitones above the base.
func (pn panel) offsets() []int {
	var out []int
	for _, id := range pn.octaves.ids() {
		out = append(out, 12*(id-1))
	}
	return out
}

func (pn panel) write(t *appconfig.Tuner) {
	t.Mode, t.ReedBoxes, t.OctaveOffsets = string(pn.mode), pn.reeds.ids(), pn.offsets()
}

// apply sets the engine up for the panel's mode: chords pick their own notes in automatic and the
// octave and chord modes run untransposed, as the reference does.
func (pn panel) apply(c dsp.EngineConfig) dsp.EngineConfig {
	switch pn.mode {
	case ModeChords:
		c.Chords, c.ManualNote, c.Transpose, c.ReedCount = true, autoNote, 0, 3
	case ModeOctaves:
		c.Transpose = 0
		c.Octaves = []dsp.OctaveRequest{{Offset: 0, Reeds: 1}}
		for _, off := range pn.offsets() {
			c.Octaves = append(c.Octaves, dsp.OctaveRequest{Offset: off, Reeds: 1})
		}
		c.ReedCount = len(c.Octaves)
	default:
		c.Reeds = pn.reeds.ids()
		c.ReedCount = len(c.Reeds)
	}
	return c
}

// SetMode switches the measuring mode used while no session is open: chords also clear the manual
// note. Refused while a session is open, since its register decides.
func (s *Service) SetMode(mode string) error {
	m := Mode(mode)
	if !m.valid() {
		return rejected("mode", mode, ErrInvalidMode)
	}
	return s.set("mode", func(c *dsp.EngineConfig, pn *panel, open bool) error {
		if open {
			return ErrSessionDecides
		}
		pn.mode = m
		if m == ModeChords {
			c.ManualNote = autoNote
		}
		return nil
	}, writePanel)
}

// SetReedBox turns one of tones mode's reed boxes 1..3 on or off; the last one on stays on.
func (s *Service) SetReedBox(box int, on bool) error {
	if box < 1 || box > 3 {
		return rejected("reed_box", box, ErrInvalidBox)
	}
	return s.setBox("reed_box", func(pn *panel) *boxes { return &pn.reeds }, box, on)
}

// SetOctave turns the octave 12 or 24 semitones above the base on or off in octaves mode; one of
// them stays on.
func (s *Service) SetOctave(offset int, on bool) error {
	if !slices.Contains([]int{12, 24}, offset) {
		return rejected("octave", offset, ErrInvalidBox)
	}
	return s.setBox("octave", func(pn *panel) *boxes { return &pn.octaves }, 1+offset/12, on)
}

func (s *Service) setBox(setting string, of func(*panel) *boxes, id int, on bool) error {
	return s.set(setting, func(_ *dsp.EngineConfig, pn *panel, _ bool) error {
		b := of(pn)
		next := b.with(id, on)
		if next == 0 {
			return ErrLastBox
		}
		*b = next
		return nil
	}, writePanel)
}

func writePanel(fc *appconfig.Config, _ dsp.EngineConfig, pn panel) { pn.write(&fc.Tuner) }
