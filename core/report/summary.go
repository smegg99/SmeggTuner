package report

func summarize(rows []Row, tol, beatTol float64) Summary {
	s := Summary{Tolerance: tol, BeatTol: beatTol, Notes: len(rows)}
	for i, r := range rows {
		if i == 0 || r.Note < s.MinNote {
			s.MinNote, s.MinName = r.Note, r.Name
		}
		if i == 0 || r.Note > s.MaxNote {
			s.MaxNote, s.MaxName = r.Note, r.Name
		}
		if r.Merged {
			s.Merged++
		}
		if r.Derived {
			s.Derived++
		}
		if r.Manual {
			s.Manual++
		}
		if r.OutOfTol > 0 {
			s.NotesOut++
		}
		for _, c := range r.Reeds {
			if !c.Present {
				continue
			}
			s.Reeds++
			if !c.InTol {
				s.ReedsOut++
			}
		}
		for _, c := range r.Beats {
			if !c.Present {
				continue
			}
			s.Beats++
			if !c.InTol {
				s.BeatsOut++
			}
		}
	}
	return s
}
