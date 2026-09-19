package meter

// take is the record append after a measurement: recording, lock or request, not inhibited, no present slot beyond +/-58.
func (a *Analyzer) take(deviations []float64) bool {
	if !a.Recording || a.History.Inhibited || !a.History.Locked && !a.TakeRequest {
		return false
	}
	if !(Refined{Deviations: deviations}).DisplayEligible() {
		return false
	}
	a.TakeRequest = false
	a.History.Inhibited = true
	a.History.Evaluate([3]int{}, [3]float64{}, [3]int{})
	a.Display.Recorded = true
	return true
}
