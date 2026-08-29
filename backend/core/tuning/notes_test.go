package tuning

import (
	"math"
	"testing"
)

func almost(t *testing.T, got, want, tol float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s: got %v want %v (tol %v)", msg, got, want, tol)
	}
}

func TestFreq(t *testing.T) {
	almost(t, NoteA4.Freq(440), 440.0, 1e-9, "A4@440")
	almost(t, MinNote.Freq(440), 20.60172231, 1e-6, "E0@440")
	almost(t, MaxNote.Freq(440), 8372.018090, 1e-4, "C9@440")
	almost(t, Note(60).Freq(440), 261.6255653, 1e-6, "C4@440")
	almost(t, NoteA4.Freq(442), 442.0, 1e-9, "A4@442")
}

func TestCentsRoundTrip(t *testing.T) {
	almost(t, Cents(441, 440), 3.930158, 1e-5, "441 vs 440")
	almost(t, Cents(440, 440), 0, 1e-12, "identity")
	f := FreqAtCents(440, 17.5)
	almost(t, Cents(f, 440), 17.5, 1e-9, "round trip")
}

func TestNames(t *testing.T) {
	cases := []struct {
		n    Note
		s    ScaleNaming
		want string
	}{
		{69, NamingCDEFGAB, "A4"},
		{71, NamingCDEFGAB, "B4"},
		{71, NamingCDEFGAH, "H4"},
		{70, NamingCDEFGAH, "B4"}, // German: A# is written B
		{60, NamingDoReMi, "Do4"},
		{61, NamingDoReMi, "Do#4"},
		{16, NamingCDEFGAB, "E0"},
		{120, NamingCDEFGAB, "C9"},
	}
	for _, c := range cases {
		if got := c.n.Name(c.s); got != c.want {
			t.Fatalf("Name(%d,%d)=%q want %q", c.n, c.s, got, c.want)
		}
	}
}

func TestNearest(t *testing.T) {
	n, dev := Nearest(442.6, 440)
	if n != NoteA4 {
		t.Fatalf("nearest of 442.6 = %v want A4", n)
	}
	almost(t, dev, Cents(442.6, 440), 1e-9, "dev")
	n, _ = Nearest(1.0, 440) // below range clamps
	if n != MinNote {
		t.Fatalf("clamp low: %v", n)
	}
	n, _ = Nearest(20000, 440)
	if n != MaxNote {
		t.Fatalf("clamp high: %v", n)
	}
}

func TestManualStepRules(t *testing.T) {
	cases := []struct {
		from  Note
		steps int
		want  Note
	}{
		{0, 1, NoteA4}, {0, -12, NoteA4}, {NoteA4, 1, 70}, {NoteA4, -12, 57},
		{MinNote, -1, MinNote}, {MinNote + 5, -12, MinNote}, {MaxNote, 1, MaxNote}, {MaxNote - 5, 12, MaxNote},
		{24, -1, 23},
	}
	for _, c := range cases {
		if got := c.from.Step(c.steps); got != c.want {
			t.Errorf("Note(%d).Step(%d) = %d, want %d", c.from, c.steps, got, c.want)
		}
	}
	if ManualStart(0) != NoteA4 || ManualStart(MaxNote+1) != NoteA4 || ManualStart(40) != 40 {
		t.Fatal("entering manual pins the shown note, else A4")
	}
}

func TestPitchNameHasNoOctave(t *testing.T) {
	for s, want := range map[ScaleNaming]string{NamingCDEFGAB: "A#", NamingCDEFGAH: "B", NamingDoReMi: "La#", NamingPolish: "Ais"} {
		if got := Note(70).PitchName(s); got != want {
			t.Errorf("naming %d: %q, want %q", s, got, want)
		}
	}
}

// The recordings are named in this notation: sharps spelled out, and H where English writes B.
func TestPolishNames(t *testing.T) {
	cases := map[Note]string{
		60: "C4",
		61: "Cis4", // not C#
		70: "Ais4", // A sharp, not B
		71: "H4",   // B natural is H
		56: "Gis3",
		59: "H3",
	}
	for note, want := range cases {
		if got := note.Name(NamingPolish); got != want {
			t.Errorf("Note(%d).Name(NamingPolish) = %q, want %q", note, got, want)
		}
	}
}
