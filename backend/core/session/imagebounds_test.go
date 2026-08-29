// backend/core/session/imagebounds_test.go
package session

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestImageDimensionsAreBoundedBeforeDecoding(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	binary.BigEndian.PutUint32(raw[16:20], 1000000)
	binary.BigEndian.PutUint32(raw[20:24], 1000000)
	binary.BigEndian.PutUint32(raw[29:33], crc32.ChecksumIEEE(raw[12:29]))
	if _, err := PrepareImage(bytes.NewReader(raw)); !errors.Is(err, ErrImageHuge) {
		t.Fatalf("oversized header=%v", err)
	}
}

func TestThinImageKeepsPositiveDimensions(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 2801))); err != nil {
		t.Fatal(err)
	}
	data, err := PrepareImage(&buf)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 1 || cfg.Height != ImageMaxEdge {
		t.Fatalf("dimensions=%dx%d", cfg.Width, cfg.Height)
	}
}
