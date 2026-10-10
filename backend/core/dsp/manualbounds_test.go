// backend/core/dsp/manualbounds_test.go
package dsp

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

func TestManualRegisterOutsideBandTableSelectsNoNote(t *testing.T) {
	for _, tc := range []struct {
		name   string
		note   tuning.Note
		offset int
	}{{"high", 120, 12}, {"low", 16, -12}, {"below_first_band", 26, -12}} {
		t.Run(tc.name, func(t *testing.T) {
			for _, continuous := range []bool{false, true} {
				cfg := EngineConfig{ReedCount: 1, ManualNote: tc.note, Continuous: continuous, Octaves: []OctaveRequest{{Offset: tc.offset, Reeds: 1}}}
				ms := runSynth(t, cfg, 0.7, 440)
				if len(ms) == 0 {
					t.Fatal("no frames")
				}
				for _, m := range ms {
					if m.Note != 0 || m.Taken {
						t.Fatalf("invalid manual layout selected %+v", m)
					}
				}
			}
		})
	}
}
