package dsp

import "testing"

func TestRingWraps(t *testing.T) {
	r := NewRing(8)
	r.Write([]float32{1, 2, 3, 4, 5})
	r.Write([]float32{6, 7, 8, 9, 10}) // overwrites 1,2
	if r.Len() != 8 {
		t.Fatalf("len = %d", r.Len())
	}
	dst := make([]float64, 4)
	n := r.Tail(4, dst)
	if n != 4 || dst[0] != 7 || dst[3] != 10 {
		t.Fatalf("tail = %v (n=%d)", dst, n)
	}
}

func TestRingPartial(t *testing.T) {
	r := NewRing(16)
	r.Write([]float32{1, 2, 3})
	dst := make([]float64, 8)
	n := r.Tail(8, dst)
	if n != 3 || dst[0] != 1 || dst[2] != 3 {
		t.Fatalf("partial tail n=%d dst=%v", n, dst)
	}
}
