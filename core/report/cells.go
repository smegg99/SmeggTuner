package report

import (
	"fmt"
	"sort"

	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
)

// reedCells lays a take's reeds against the instrument's columns; a reed the take does not carry is absent, not zero.
func reedCells(errs []target.ReedError, reedCount int, banks []session.Bank, into []int) []ReedCell {
	cells := make([]ReedCell, reedCount)
	for i := range cells {
		cells[i] = ReedCell{Reed: i + 1}
		if i < len(banks) {
			cells[i].Bank = banks[i]
		}
	}

	for _, e := range errs {
		col := e.Reed
		if into != nil {
			if e.Reed < 0 || e.Reed >= len(into) {
				continue
			}
			col = into[e.Reed]
		}
		if col < 0 || col >= reedCount {
			continue // the take carries more reeds than the instrument declares
		}

		cells[col] = ReedCell{
			Reed:    col + 1,
			Bank:    cells[col].Bank,
			Present: true,
			Curr:    e.Curr,
			Goal:    e.Goal,
			Error:   e.Error,
			InTol:   e.InTol,
		}
	}
	return cells
}

// columnsFor maps each reed of a take to its card column; nil on any uncertain mapping, so reeds
// fall back to positions. See session.BankColumns for the claim.
func columnsFor(i session.Instrument, t session.Take, reeds int) []int {
	if reeds == 0 {
		return nil
	}
	return session.BankColumns(i.Banks, i.TakeBanks(t), t.Reeds)
}

// colOf is the column a reed answers to; an unmapped reed answers to its own position.
func colOf(into []int, reed int) int {
	if reed >= 0 && reed < len(into) {
		return into[reed]
	}
	return reed
}

func beatCells(byCols map[[2]int]target.BeatError, pairs []Pair) []BeatCell {
	cells := make([]BeatCell, len(pairs))
	for i, p := range pairs {
		b, ok := byCols[[2]int{p.Low - 1, p.High - 1}]
		if !ok {
			continue
		}
		cells[i] = BeatCell{
			Present:      true,
			Curr:         b.Curr,
			Goal:         b.Goal,
			Error:        b.Error,
			CurrHz:       b.CurrHz,
			GoalHz:       b.GoalHz,
			ErrorHz:      b.ErrorHz,
			InTol:        b.InTol,
			FromEnvelope: b.FromEnvelope,
		}
	}
	return cells
}

// pairsOf picks the beat columns: every pair of adjacent card columns, plus a column of its own for
// a beat whose two ranks are next to each other in the TAKE but not on the card. A register short a
// rank beats across that gap - an LMMH card with the second 8' unheard still has the 8'-to-4' beat -
// and adjacency on the card is the only thing that gap breaks. Pairs that are adjacent in neither
// are left out: an eight-reed instrument would otherwise be twenty-eight beat columns wide.
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
