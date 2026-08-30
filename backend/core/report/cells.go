package report

import (
	"fmt"
	"sort"

	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
)

// reedCells lays a take's slots against the instrument's columns in unit; a slot that did not sound is absent, not zero.
func reedCells(errs []target.ReedError, reedCount int, banks []session.Bank, into []int, unit target.Unit) []ReedCell {
	cells := make([]ReedCell, reedCount)
	for i := range cells {
		cells[i] = ReedCell{Reed: i + 1}
		if i < len(banks) {
			cells[i].Bank = banks[i]
		}
	}

	for _, e := range errs {
		e = e.In(unit)
		col := e.Reed
		if into != nil {
			if e.Reed < 0 || e.Reed >= len(into) {
				continue
			}
			col = into[e.Reed]
		}
		if col < 0 || col >= reedCount {
			continue // the take carries more slots than the instrument declares
		}

		cells[col] = ReedCell{
			Reed:    col + 1,
			Bank:    cells[col].Bank,
			Present: true,
			HasGoal: e.HasGoal,
			Curr:    e.Curr,
			Goal:    e.Goal,
			Error:   e.Error,
			InTol:   e.InTol,
		}
	}
	return cells
}

// colOf is the column a slot answers to; an unmapped slot answers to its own position.
func colOf(into []int, slot int) int {
	if slot >= 0 && slot < len(into) {
		return into[slot]
	}
	return slot
}

func beatCells(byCols map[[2]int]target.BeatError, pairs []Pair) []BeatCell {
	cells := make([]BeatCell, len(pairs))
	for i, p := range pairs {
		b, ok := byCols[[2]int{p.Low - 1, p.High - 1}]
		if !ok {
			continue
		}
		cells[i] = BeatCell{
			Present: true,
			HasGoal: b.HasGoal,
			Curr:    b.Curr,
			Goal:    b.Goal,
			Error:   b.Error,
			CurrHz:  b.CurrHz,
			GoalHz:  b.GoalHz,
			ErrorHz: b.ErrorHz,
			InTol:   b.InTol,
		}
	}
	return cells
}

// pairsOf picks the beat columns: every pair of adjacent card columns, plus a column of its own for
// a beat whose two slots are neighbours in the TAKE but not on the card. A register short a rank
// beats across that gap (an LH switch on an LMMH card beats 16' against 4'), and adjacency on the
// card is the only thing that gap breaks. Pairs that are adjacent in neither are left out: an
// eight-reed instrument would otherwise be twenty-eight beat columns wide.
func pairsOf(reedCount int, beats []map[[2]int]target.BeatError) []Pair {
	seen := make(map[[2]int]Pair)
	add := func(lo, hi int) {
		if lo < 0 || hi <= lo {
			return
		}
		seen[[2]int{lo, hi}] = Pair{Key: fmt.Sprintf("%d-%d", lo+1, hi+1), Low: lo + 1, High: hi + 1}
	}
	for lo := 0; lo+1 < reedCount; lo++ {
		add(lo, lo+1)
	}
	for _, byCols := range beats {
		for cols, b := range byCols {
			if b.High != b.Low+1 {
				continue
			}
			add(cols[0], cols[1])
		}
	}

	pairs := make([]Pair, 0, len(seen))
	for _, p := range seen {
		pairs = append(pairs, p)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Low != pairs[j].Low {
			return pairs[i].Low < pairs[j].Low
		}
		return pairs[i].High < pairs[j].High
	})
	return pairs
}
