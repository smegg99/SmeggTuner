package meter

import "testing"

func retentionCMajor() [3]int { return [3]int{36, 40, 43} }

func retentionState(notes [3]int, label string, change, last uint32) ChordRetention {
	return ChordRetention{Notes: notes, Label: label, ChangeTick: change, LastTick: last}
}

// retentionFilter fails the test unless the filter reads exactly the supplied ticks.
func retentionFilter(t *testing.T, s *ChordRetention, notes [3]int, label string, ticks ...uint32) ChordRetained {
	t.Helper()
	used := 0
	out := s.Filter(notes, label, func() uint32 {
		used++
		if used > len(ticks) {
			t.Fatalf("filter read tick %d of %d", used, len(ticks))
		}
		return ticks[used-1]
	})
	if used != len(ticks) {
		t.Fatalf("filter read %d ticks, supplied %d", used, len(ticks))
	}
	return out
}

func retentionExpect(t *testing.T, got, want ChordRetention) {
	t.Helper()
	if got != want {
		t.Errorf("state %+v, want %+v", got, want)
	}
}

func TestRetentionUnchangedNotesSkipChangeAge(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 100, 100)
	if out := retentionFilter(t, &s, retentionCMajor(), "new label", 1000, 1001); out.RetainedPrevious {
		t.Error("unchanged tuple was retained")
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "new label", 100, 1001))
}

func TestRetentionChangeAgeIsStrictlyAbove750(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 100, 100)
	out := retentionFilter(t, &s, [3]int{36, 40, 44}, "new", 850, 851, 852, 853)
	if out.RetainedPrevious || out.Notes != [3]int{36, 40, 44} {
		t.Errorf("750 ms retained: %+v", out)
	}
	retentionExpect(t, s, retentionState([3]int{36, 40, 44}, "new", 851, 853))

	s = retentionState(retentionCMajor(), "C", 100, 100)
	out = retentionFilter(t, &s, [3]int{36, 40, 44}, "new", 851, 852, 853)
	if !out.RetainedPrevious || out.Notes != retentionCMajor() || out.Label != "C" {
		t.Errorf("751 ms did not retain the stored pair: %+v", out)
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 100, 853))
}

func TestRetentionMatchesArePositional(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 100, 100)
	retentionFilter(t, &s, [3]int{48, 52, 55}, "next", 1000, 1001, 1002)
	retentionExpect(t, s, retentionState([3]int{48, 52, 55}, "next", 1000, 1002))

	s = retentionState(retentionCMajor(), "C", 100, 100)
	if out := retentionFilter(t, &s, [3]int{43, 0, 36}, "fixture", 1000, 1001, 1002); out.RetainedPrevious {
		t.Error("cross-extreme equalities counted")
	}

	s = retentionState([3]int{}, "", 100, 100)
	out := retentionFilter(t, &s, [3]int{36, 0, 0}, UnknownChord, 851, 852, 853)
	if !out.RetainedPrevious || out.Notes != [3]int{} || out.Label != "" {
		t.Errorf("zero equalities or the empty label: %+v", out)
	}
	retentionExpect(t, s, retentionState([3]int{}, "", 100, 853))

	s = retentionState([3]int{36, 36, 40}, "fixture", 100, 100)
	if out := retentionFilter(t, &s, [3]int{36, 48, 52}, UnknownChord, 851, 852, 853); !out.RetainedPrevious {
		t.Error("duplicate equalities did not count twice")
	}
}

func TestRetentionUnknownLabel(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 100, 100)
	out := retentionFilter(t, &s, [3]int{36, 40, 0}, UnknownChord, 1000, 1001, 1002)
	if !out.RetainedPrevious || out.Label != "C" || s.LastTick != 1002 {
		t.Errorf("unknown candidate was not replaced first: %+v %+v", out, s)
	}

	s = retentionState(retentionCMajor(), "C", 100, 100)
	out = retentionFilter(t, &s, [3]int{}, UnknownChord, 1000, 1001)
	if out.Notes != [3]int{} || out.Label != UnknownChord || out.ExpiredStored {
		t.Errorf("unknown output: %+v", out)
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 0, 0))

	s = retentionState(retentionCMajor(), UnknownChord, 100, 100)
	if out := retentionFilter(t, &s, retentionCMajor(), UnknownChord, 1000); out.Notes != [3]int{} {
		t.Errorf("unchanged unknown kept notes: %+v", out)
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), UnknownChord, 0, 0))
}

func TestRetentionExpiry(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 100, 100)
	if out := retentionFilter(t, &s, [3]int{}, UnknownChord, 3099, 3100); out.ExpiredStored {
		t.Error("3000 ms expired")
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 0, 0))

	s = retentionState(retentionCMajor(), "C", 100, 100)
	if out := retentionFilter(t, &s, [3]int{}, UnknownChord, 3100, 3101); !out.ExpiredStored {
		t.Error("3001 ms did not expire")
	}
	retentionExpect(t, s, retentionState([3]int{}, "", 0, 0))

	s = retentionState(retentionCMajor(), "C", 100, 100)
	out := retentionFilter(t, &s, [3]int{36, 40, 44}, UnknownChord, 4000, 4001, 4002)
	if !out.RetainedPrevious || !out.ExpiredStored || out.Notes != retentionCMajor() || out.Label != "C" {
		t.Errorf("expiry erased the retained pair: %+v", out)
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 100, 4002))

	s = retentionState(retentionCMajor(), "C", 100, 100)
	if out := retentionFilter(t, &s, [3]int{48, 52, 55}, "next", 4000, 4001, 4002); !out.ExpiredStored || out.RetainedPrevious {
		t.Errorf("new tuple after expiry: %+v", out)
	}
	retentionExpect(t, s, retentionState([3]int{48, 52, 55}, "next", 4000, 4002))
}

func TestRetentionZeroChangeTickIsInitialized(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 0, 0)
	retentionFilter(t, &s, retentionCMajor(), "C", 500, 501, 502)
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 501, 502))

	s = retentionState(retentionCMajor(), "C", 0, 0)
	if out := retentionFilter(t, &s, [3]int{36, 40, 44}, "new", 751, 752, 753, 754); !out.RetainedPrevious {
		t.Error("751 ms from a zero change tick did not retain")
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 753, 754))
}

func TestRetentionSequentialCalls(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 100, 100)
	retentionFilter(t, &s, [3]int{36, 40, 44}, "first", 500, 500, 500, 500)
	retentionExpect(t, s, retentionState([3]int{36, 40, 44}, "first", 500, 500))
	if out := retentionFilter(t, &s, [3]int{36, 40, 45}, "second", 1251, 1251, 1251); out.Notes != [3]int{36, 40, 44} {
		t.Errorf("second call did not retain the first tuple: %+v", out)
	}
	if s.ChangeTick != 500 {
		t.Errorf("change tick %d, want 500", s.ChangeTick)
	}
	retentionFilter(t, &s, [3]int{48, 52, 55}, "third", 1300, 1300, 1300)
	retentionExpect(t, s, retentionState([3]int{48, 52, 55}, "third", 1300, 1300))
}

func TestRetentionAgesWrapAsUint32(t *testing.T) {
	s := retentionState(retentionCMajor(), "C", 4294967196, 4294967196)
	if out := retentionFilter(t, &s, retentionCMajor(), "C", 50, 51); out.ExpiredStored {
		t.Error("a 150 ms age across the wrap expired")
	}
	retentionExpect(t, s, retentionState(retentionCMajor(), "C", 4294967196, 51))
}
