package meter

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"testing"
)

// transformTone rounds a sum of {amplitude, radians per sample, phase} cosines, so the input is integral.
func transformTone(samples int, tones ...[3]float64) []float32 {
	src := make([]float32, samples)
	for j := range src {
		var v float64
		for _, t := range tones {
			v += float64(t[0] * math.Cos(float64(t[1]*float64(j))+t[2]))
		}
		src[j] = float32(math.Round(v))
	}
	return src
}

// transformDigest is FNV-1a over the little-endian IEEE bits of every value in order.
func transformDigest[T float32 | float64](tables ...[]T) string {
	h := fnv.New64a()
	for _, table := range tables {
		if err := binary.Write(h, binary.LittleEndian, table); err != nil {
			panic(err)
		}
	}
	return fmt.Sprintf("%016x", h.Sum64())
}

func TestTransformTwiddleGolden(t *testing.T) {
	want := map[int]string{
		1 << 3:  "f9d36bb99768eea4",
		1 << 4:  "f3d1befeb21b3158",
		1 << 5:  "b3a30fc6c44504bd",
		1 << 6:  "e0bf2c4e8c79a2a7",
		1 << 7:  "c9663bc1c3d424cc",
		1 << 8:  "8a54a3089b38cd19",
		1 << 9:  "26af036d145b354f",
		1 << 10: "e7a017ffba98d451",
		1 << 11: "706ca67dc3b27ddc",
		1 << 12: "69252a4d3b485cc0",
		1 << 13: "c0b392f60d5a8f0c",
		1 << 14: "e4b29c116f1e6ed5",
		1 << 15: "268950de2769703c",
		1 << 16: "8f4ae86311fc30dc",
		1 << 17: "341fbe22eb372407",
		1 << 18: "2148dd3006204c95",
		1 << 19: "97e450b42f28d095",
		1 << 20: "c6b5177fa2037a9c",
		1 << 21: "436624b0f027a2ce",
	}
	for size := 8; size <= 1<<21; size *= 2 {
		got := transformDigest(NewTransform(size).twiddles...)
		if got != want[size] {
			t.Errorf("size %d: twiddle digest %s, want %s", size, got, want[size])
		}
	}
}

func TestTransformPackedGolden(t *testing.T) {
	const a4, a3 = 2 * math.Pi * 440 / 48000, 2 * math.Pi * 220 / 48000
	cases := []struct {
		size, samples int
		tones         [][3]float64
		want          string
	}{
		{1024, 600, [][3]float64{{12000, a4 * 20, 0.3}}, "8cf6c1d0ffde62c9"},
		{2048, 2048, [][3]float64{{9000, a4 * 10, 0}, {7000, a4 * 12.6, 1.1}, {5000, a4 * 15, 2.2}}, "f5f825ed06ca443a"},
		{262144, 65536, [][3]float64{{16000, a4, 0.5}, {4000, 2 * a4, 1.5}, {1000, 3 * a4, 2.5}}, "8fbe909df7ae34bc"},
		{262144, 65536, [][3]float64{{10000, a3, 0}, {10000, a3 * 1.0035, 0.7}, {6000, a4 * 1.4983, 0}}, "0e740f7f5facb7e3"},
		{524288, 131072, [][3]float64{{12000, a4, 0}, {12000, a4 * 1.0034, 0.2}, {3000, 2 * a4, 0.4}}, "7a6dd5265e8eed08"},
		{524288, 131072, [][3]float64{{20000, a3 / 4, 0.9}, {8000, a3 / 2, 0}}, "86478514d27779db"},
	}
	plans := map[int]*Transform{}
	for i, c := range cases {
		if plans[c.size] == nil {
			plans[c.size] = NewTransform(c.size)
		}
		packed := plans[c.size].Packed(make([]float32, c.size), transformTone(c.samples, c.tones...))
		got := transformDigest(packed)
		if got != c.want {
			t.Errorf("case %d (size %d): packed digest %s, want %s", i, c.size, got, c.want)
		}
	}
}
