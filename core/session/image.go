package session

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"golang.org/x/image/draw"

	// Registered for its decoder.
	_ "image/png"
)

// ImageMaxEdge is the longest side a stored photo may have.
const ImageMaxEdge = 1400

// ImageQuality is the JPEG quality a capped photo is written at.
const ImageQuality = 85

// ImageMaxBytes is the largest input file accepted, guarding against a RAW file or video picked by mistake.
const ImageMaxBytes = 32 << 20

var (
	ErrNotAnImage = errors.New("session: that file is not an image")
	ErrImageHuge  = errors.New("session: that image is too large")
)

// ImageMediaType is what the photo is served as.
const ImageMediaType = "image/jpeg"

// PrepareImage decodes a photo to prove it is one, caps it, and returns a JPEG.
func PrepareImage(r io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(r, ImageMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > ImageMaxBytes {
		return nil, fmt.Errorf("%w: over %d bytes", ErrImageHuge, ImageMaxBytes)
	}

	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotAnImage, err)
	}

	out := fit(src)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: ImageQuality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fit scales a photo down to ImageMaxEdge, keeping aspect; a smaller one is returned untouched.
func fit(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= ImageMaxEdge && h <= ImageMaxEdge {
		return src
	}

	scale := float64(ImageMaxEdge) / float64(w)
	if h > w {
		scale = float64(ImageMaxEdge) / float64(h)
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(float64(w)*scale), int(float64(h)*scale)))
	// CatmullRom: runs once per import, so favor quality.
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Src, nil)
	return dst
}
