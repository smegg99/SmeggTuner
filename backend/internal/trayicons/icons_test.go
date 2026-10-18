// backend/internal/trayicons/icons_test.go

package trayicons

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"image/png"
	"io"
	"io/fs"
	"runtime"
	"testing"
)

func TestAppIconIs256PNG(t *testing.T) {
	img, err := png.Decode(bytes.NewReader(AppIcon))
	if err != nil {
		t.Fatalf("AppIcon is not a PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 256 || b.Dy() != 256 {
		t.Errorf("AppIcon is %dx%d, want 256x256", b.Dx(), b.Dy())
	}
}

func TestEveryTrayIconUsesNativeBitmap(t *testing.T) {
	for _, status := range []Status{StatusIdle} {
		for _, dark := range []bool{false, true} {
			data, err := Icon(dark, status)
			if err != nil {
				t.Fatalf("Icon(%v, %s): %v", dark, status, err)
			}
			if runtime.GOOS == "windows" {
				if len(data) < 6 || binary.LittleEndian.Uint16(data[2:]) != 1 || binary.LittleEndian.Uint16(data[4:]) != 6 {
					t.Fatalf("Icon(%v, %s) is not a six-resolution ICO", dark, status)
				}
				continue
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Icon(%v, %s) is not a PNG: %v", dark, status, err)
			}
			want := 32
			if runtime.GOOS == "darwin" {
				want = 44
			}
			if b := img.Bounds(); b.Dx() != want || b.Dy() != want {
				t.Errorf("Icon(%v, %s) is %dx%d, want %dx%d", dark, status, b.Dx(), b.Dy(), want, want)
			}
		}
	}
}

func TestEveryTrayStatePublishesLinuxSVGAndBitmapFallback(t *testing.T) {
	for _, status := range []Status{StatusIdle} {
		for _, dark := range []bool{false, true} {
			icons := ForStatus(status)
			files, name, err := icons.ThemedIcon(dark)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{name + ".svg", "scalable/status/" + name + ".svg", "22x22/status/" + name + ".svg"} {
				data, err := fs.ReadFile(files, path)
				if err != nil {
					t.Fatal(err)
				}
				decoder := xml.NewDecoder(bytes.NewReader(data))
				root := false
				for {
					token, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if start, ok := token.(xml.StartElement); ok && start.Name.Local == "svg" {
						root = true
					}
				}
				if !root {
					t.Fatalf("%s has no SVG root", path)
				}
			}
			if _, err := fs.ReadFile(files, "index.theme"); err != nil {
				t.Fatal(err)
			}
			if _, err := icons.Icon(dark); err != nil {
				t.Fatal(err)
			}
		}
	}
}
