package meter

import "testing"

func BenchmarkExtractPeaks(b *testing.B) {
	spectrum := make([]float32, 1200)
	for i := range spectrum {
		spectrum[i] = float32(i * 7919 % 1009)
	}
	b.ReportAllocs()
	for b.Loop() {
		ExtractPeaks(spectrum, 100, 1100)
	}
}

func BenchmarkAssignPeaks(b *testing.B) {
	selected, targets := []int{1, 2, 3}, [3]float64{0, 10, -10}
	b.ReportAllocs()
	for b.Loop() {
		AssignPeaks([3]int{205, 200, 0}, selected, targets, 200)
		AssignPeaks([3]int{201, 0, 0}, selected, targets, 200)
	}
}
