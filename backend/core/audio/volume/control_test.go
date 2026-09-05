// backend/core/audio/volume/control_test.go
package volume

import (
	"math"
	"testing"
)

// pct is an RMS at p percent of full scale.
func pct(p float32) float32 { return p / 100 * fullScale }

func near(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-6 }

// room learns a floor at 0.1 % of full scale.
func room(c *Control) {
	for range 4 {
		c.Step(Interval{RMS: pct(0.1), Peak: 100}, false, 0.5)
	}
}

func TestAWeakNoteRaisesTheLevel(t *testing.T) {
	var c Control
	room(&c)
	got, set := c.Step(Interval{RMS: pct(2), Peak: 2000}, false, 0.5)
	if !set || !near(got, 0.525) {
		t.Fatalf("weak note: %v %v, want +0.025", got, set)
	}
}

// Without a signal twice the floor nothing moves: steady room noise is not a note.
func TestTheFloorIsNotASignal(t *testing.T) {
	var c Control
	for range 6 {
		if got, set := c.Step(Interval{RMS: pct(2), Peak: 2000}, false, 0.5); set {
			t.Fatalf("a steady level moved the volume to %v", got)
		}
	}
}

func TestThreeLoudIntervalsLowerTheLevel(t *testing.T) {
	var c Control
	room(&c)
	loudIv := Interval{RMS: pct(30), Peak: 20000}
	for i, want := range []float32{0.5, 0.5, 0.45, 0.4} {
		level := float32(0.5)
		if i == 3 {
			level = 0.45
		}
		got, _ := c.Step(loudIv, false, level)
		if !near(got, want) {
			t.Fatalf("loud interval %d: %v, want %v", i+1, got, want)
		}
	}
}

func TestTwoClippedIntervalsLowerByATenth(t *testing.T) {
	var c Control
	room(&c)
	clip := Interval{RMS: pct(30), Peak: 32767, Clipped: true}
	if got, set := c.Step(clip, false, 0.6); set {
		t.Fatalf("one clipped interval moved the level to %v", got)
	}
	if got, set := c.Step(clip, false, 0.6); !set || !near(got, 0.5) {
		t.Fatalf("second clipped interval: %v %v, want 0.5", got, set)
	}
	if got, _ := c.Step(clip, false, 0.5); !near(got, 0.4) {
		t.Fatalf("every further clipped interval lowers again: %v", got)
	}
	// The hold of one interval is spent before it is tested, so it never blocks the next raise.
	if got, set := c.Step(Interval{RMS: pct(2), Peak: 2000}, false, 0.4); !set || !near(got, 0.425) {
		t.Fatalf("raise after clipping: %v %v", got, set)
	}
}

func TestTheLevelStaysWithinATenthAndOne(t *testing.T) {
	var c Control
	if got, set := c.Step(Interval{}, false, 0.05); !set || got != minLevel {
		t.Fatalf("a device under 10 %% = %v %v, want lifted to 0.1", got, set)
	}
	room(&c)
	if got, set := c.Step(Interval{RMS: pct(2), Peak: 2000}, false, 0.999); set || got != 1 {
		t.Fatalf("raise past 1 = %v %v, want clamped and too small to set", got, set)
	}
}

func TestAMutedEndpointIsLeftAlone(t *testing.T) {
	var c Control
	room(&c)
	before := c.floor
	if got, set := c.Step(Interval{RMS: pct(2), Peak: 2000}, true, 0.05); set || got != 0.05 {
		t.Fatalf("muted: %v %v", got, set)
	}
	if c.floor != before {
		t.Fatal("a muted interval moved the floor")
	}
}

func TestTheNoiseFloor(t *testing.T) {
	var c Control
	c.learn(Interval{RMS: 3})
	if c.floor != 0 {
		t.Fatal("an interval under 0.01 % of full scale was learned")
	}
	c.learn(Interval{RMS: 100, Peak: 300})
	if c.floor != 100 {
		t.Fatalf("first floor %v, want the RMS", c.floor)
	}
	steps := []struct {
		iv   Interval
		want float32
	}{
		{Interval{RMS: 50, Peak: 100}, 100 + 0.15*(50-100)},                    // quieter: fast
		{Interval{RMS: 200, Peak: 900}, 92.5 + 0.05*(200-92.5)},                // quiet louder: slow
		{Interval{RMS: 2000, Peak: 9000}, 97.875 + 0.001*(2000-97.875)},        // a note: hardly
		{Interval{RMS: 120, Peak: 4000}, 99.777125 + 0.001*(120-99.777125)},    // a high peak is not quiet
		{Interval{RMS: 120, Clipped: true}, 99.797348 + 0.001*(120-99.797348)}, // nor is a flat top
	}
	for i, s := range steps {
		c.learn(s.iv)
		if math.Abs(float64(c.floor-s.want)) > 1e-3 {
			t.Fatalf("step %d: floor %v, want %v", i, c.floor, s.want)
		}
	}
	c.Step(Interval{Restart: true, RMS: 40, Peak: 100}, false, 0.5)
	if c.floor != 40 {
		t.Fatalf("a restart relearns from the interval: %v", c.floor)
	}
}
