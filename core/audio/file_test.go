// core/audio/file_test.go
package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	gaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func writeTestWAV(t *testing.T, path string, freq float64, sr, seconds int) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := wav.NewEncoder(f, sr, 16, 1, 1)
	n := sr * seconds
	buf := &gaudio.IntBuffer{
		Format:         &gaudio.Format{NumChannels: 1, SampleRate: sr},
		SourceBitDepth: 16,
		Data:           make([]int, n),
	}
	for i := 0; i < n; i++ {
		buf.Data[i] = int(30000 * math.Sin(2*math.Pi*freq*float64(i)/float64(sr)))
	}
	if err := enc.Write(buf); err != nil {
		t.Fatal(err)
	}
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestFileSourceReadsBack(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a440.wav")
	writeTestWAV(t, p, 440, 48000, 1)
	src, err := NewFileSource(p, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if src.Info().SampleRate != 48000 || src.Info().Realtime {
		t.Fatalf("info: %+v", src.Info())
	}
	ch, err := src.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var total int
	var peak float64
	for b := range ch {
		total += len(b.Samples)
		for _, v := range b.Samples {
			if a := math.Abs(float64(v)); a > peak {
				peak = a
			}
		}
	}
	if total != 48000 {
		t.Fatalf("total = %d", total)
	}
	if peak < 0.8 || peak > 1.0 {
		t.Fatalf("peak = %v (16-bit 30000 should be ~0.9155)", peak)
	}
}

func TestFileSourceMissing(t *testing.T) {
	if _, err := NewFileSource("/nonexistent.wav", false, false); err == nil {
		t.Fatal("expected error")
	}
}

// A block must never alias the source's decoded array: the consumer owns and overwrites it.
func TestFileSourceBlocksDoNotAliasSource(t *testing.T) {
	p := filepath.Join(t.TempDir(), "alias.wav")
	writeTestWAV(t, p, 440, 48000, 1)
	src, err := NewFileSource(p, false, false)
	if err != nil {
		t.Fatal(err)
	}
	ch, err := src.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for b := range ch {
		for i := range b.Samples {
			b.Samples[i] = 0
		}
	}
	var peak float64
	for _, v := range src.samples {
		if a := math.Abs(float64(v)); a > peak {
			peak = a
		}
	}
	if peak < 0.8 {
		t.Fatalf("consumer writes reached the source array: peak = %v", peak)
	}
}

// go-audio passes the fmt channel count through unchecked, so the downmix must reject a bad format itself.
func TestDownmixMonoRejectsDegenerateFormat(t *testing.T) {
	cases := map[string]*gaudio.IntBuffer{
		"zero channels": {
			Format:         &gaudio.Format{NumChannels: 0, SampleRate: 48000},
			SourceBitDepth: 16,
			Data:           make([]int, 64),
		},
		"zero bit depth": {
			Format:         &gaudio.Format{NumChannels: 1, SampleRate: 48000},
			SourceBitDepth: 0,
			Data:           make([]int, 64),
		},
		"negative bit depth": {
			Format:         &gaudio.Format{NumChannels: 1, SampleRate: 48000},
			SourceBitDepth: -8,
			Data:           make([]int, 64),
		},
	}
	for name, buf := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := downmixMono(buf); err == nil {
				t.Fatal("expected an error for a degenerate format")
			}
		})
	}
}

func TestFileSourceMalformedWAV(t *testing.T) {
	cases := map[string]struct{ channels, bitDepth uint16 }{
		"zero channels":  {0, 16},
		"zero bit depth": {1, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "bad.wav")
			writeMalformedWAV(t, p, tc.channels, tc.bitDepth)
			if _, err := NewFileSource(p, false, false); err == nil {
				t.Fatal("expected an error for a malformed wav header")
			}
		})
	}
}

// writeMalformedWAV hand-builds a WAV with a degenerate channel count or bit depth: no encoder produces it, but a file picker can hand it to us.
func writeMalformedWAV(t *testing.T, path string, channels, bitDepth uint16) {
	t.Helper()
	const frames = 64
	dataBytes := uint32(frames * 2)
	var b bytes.Buffer
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+dataBytes))
	b.WriteString("WAVE")
	b.WriteString("fmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16)) // chunk size
	binary.Write(&b, binary.LittleEndian, uint16(1))  // PCM
	binary.Write(&b, binary.LittleEndian, channels)
	binary.Write(&b, binary.LittleEndian, uint32(48000)) // sample rate
	binary.Write(&b, binary.LittleEndian, uint32(96000)) // byte rate
	binary.Write(&b, binary.LittleEndian, uint16(2))     // block align
	binary.Write(&b, binary.LittleEndian, bitDepth)
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, dataBytes)
	b.Write(make([]byte, dataBytes))
	if err := os.WriteFile(path, b.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestFileSourceLoop(t *testing.T) {
	p := filepath.Join(t.TempDir(), "short.wav")
	writeTestWAV(t, p, 440, 48000, 1)
	src, err := NewFileSource(p, false, true)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := src.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var total int
	for b := range ch {
		total += len(b.Samples)
		if total > 48000*2 { // looped past EOF at least once
			cancel()
		}
	}
	if total <= 48000 {
		t.Fatalf("loop did not restart: %d samples", total)
	}
}
