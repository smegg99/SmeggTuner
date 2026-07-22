package report

import (
	"fmt"
	"slices"
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

// columnsFor maps each reed of a take to its card column: each reed claims a register bank in its
// own octave (see session.AssignBanks), so a take short a rank still lands in the right columns.
// Nil on any uncertain mapping so reeds fall back to positions. See services/record.banksOf for the
// screen's copy.
func columnsFor(i session.Instrument, t session.Take, reeds int) []int {
	if len(i.Banks) == 0 || t.Register == "" || reeds == 0 {
		return nil
	}
	r, ok := i.Register(t.Register)
	if !ok {
		return nil
	}
	assigned := session.AssignBanks(r.Banks, t.Reeds)
	if assigned == nil {
		return nil
	}

	into := make([]int, len(assigned))
	for n, b := range assigned {
		col := slices.Index(i.Banks, b)
		if col < 0 {
			return nil // a rank the instrument does not have
		}
		into[n] = col
	}
	return into
}

func beatCells(byPair map[string]target.BeatError, pairs []Pair) []BeatCell {
	cells := make([]BeatCell, len(pairs))
	for i, p := range pairs {
		b, ok := byPair[p.Key]
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

// pairsOf picks the beat columns: adjacent reed pairs plus any envelope-measured pair not among them, so a merged note's beat still gets a column.
func pairsOf(reedCount int, beats []map[string]target.BeatError) []Pair {
	seen := make(map[string]Pair)
	for lo := 1; lo < reedCount; lo++ {
		p := Pair{Key: fmt.Sprintf("%d-%d", lo, lo+1), Low: lo, High: lo + 1}
		seen[p.Key] = p
	}
	for _, byPair := range beats {
		for key, b := range byPair {
			if _, ok := seen[key]; ok || !b.FromEnvelope {
				continue
			}
			seen[key] = Pair{Key: key, Low: b.Low + 1, High: b.High + 1}
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
