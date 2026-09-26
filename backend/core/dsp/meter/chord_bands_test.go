package meter

import "testing"

func chordFrequencies() *NoteTable {
	f := EqualTemperament(440, 0)
	return &f
}

func TestBindBandsKeepsSlotOrderAndOwnGeometry(t *testing.T) {
	f := chordFrequencies()
	bands := BindBands([3]int{36, 0, 60}, f, 524288, 44100, 1)
	for slot, note := range map[int]int{0: 36, 2: 60} {
		if want := NewNoteBand(f[note-1], f[note], f[note+1], 524288, 44100, 1); bands[slot] != want {
			t.Errorf("slot %d: got %+v, want %+v", slot, bands[slot], want)
		}
	}
	if bands[1] != (NoteBand{}) {
		t.Errorf("missing slot was filled: %+v", bands[1])
	}
	if bands[2].BinCount() <= bands[0].BinCount() {
		t.Errorf("higher note band is not wider: %d <= %d", bands[2].BinCount(), bands[0].BinCount())
	}
	if BindBands([3]int{}, f, 524288, 44100, 1) != [3]NoteBand{} {
		t.Error("zero tuple bound a band")
	}
}

func TestBindBandsHasNoTrialRestrictionOrBaseRequirement(t *testing.T) {
	f := chordFrequencies()
	for _, notes := range [][3]int{{36, 40, 43}, {0, 12, 24}} {
		bands := BindBands(notes, f, 524288, 44100, 1)
		for slot, note := range notes {
			want := NoteBand{}
			if note != 0 {
				want = NewNoteBand(f[note-1], f[note], f[note+1], 524288, 44100, 1)
			}
			if bands[slot] != want {
				t.Errorf("%v slot %d: got %+v, want %+v", notes, slot, bands[slot], want)
			}
		}
	}
}
