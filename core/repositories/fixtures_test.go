package repositories_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"smegg.me/smeggtuner/core/session"
)

// photo returns a real JPEG run through PrepareImage, as a chosen file would be.
func photo(t *testing.T, w, h int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 90, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	jpg, err := session.PrepareImage(&buf)
	if err != nil {
		t.Fatal(err)
	}
	return jpg
}
