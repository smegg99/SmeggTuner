package report

import (
	"html/template"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLetterheadIsOffByDefault(t *testing.T) {
	s := musette(t)

	plain := render(t, sheet(t, s))
	if strings.Contains(plain, "Smegg Accordion Service") {
		t.Fatal("the letterhead printed without being asked for")
	}

	logo, err := LoadLogo(pngFile(t))
	if err != nil {
		t.Fatalf("LoadLogo: %v", err)
	}
	r, err := Build(s, Options{
		Now: at,
		Letterhead: &Letterhead{
			CompanyName:    "Smegg Accordion Service",
			CompanyAddress: "ul. Warsztatowa 1\n00-001 Warszawa",
			CompanyWebsite: "smegg.me",
			Logo:           logo,
		},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	out := render(t, r)

	for _, want := range []string{"Smegg Accordion Service", "ul. Warsztatowa 1", "smegg.me", `src="data:image/png;base64,`} {
		if !strings.Contains(out, want) {
			t.Errorf("the letterhead is missing %q", want)
		}
	}
	if strings.Contains(out, "ZgotmplZ") {
		t.Error("the logo was filtered out of the img tag rather than embedded")
	}
}

func TestLoadLogo(t *testing.T) {
	if uri, err := LoadLogo(""); err != nil || uri != "" {
		t.Errorf("LoadLogo(\"\") = %q, %v; want no logo and no error", uri, err)
	}
	if _, err := LoadLogo(filepath.Join(t.TempDir(), "gone.png")); err != ErrLogoUnreadable {
		t.Errorf("a missing logo = %v, want ErrLogoUnreadable", err)
	}

	text := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(text, []byte("this is not a logo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadLogo(text); err != ErrLogoNotImage {
		t.Errorf("a text file as a logo = %v, want ErrLogoNotImage", err)
	}
}

func pngFile(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 20, G: 30, B: 40, A: 255})

	path := filepath.Join(t.TempDir(), "logo.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

// The footer carries the app's own site as a real link, so a PDF keeps it clickable.
func TestTheFooterLinksHome(t *testing.T) {
	out := render(t, sheet(t, musette(t)))

	if !strings.Contains(out, `<a href="https://smeggtuner.com">smeggtuner.com</a>`) {
		t.Error("the footer does not link to the app's site")
	}
}

// A missing scheme becomes https; anything that is not a web address is not linked.
func TestTheLetterheadSiteBecomesALink(t *testing.T) {
	link := funcs["site"].(func(string) template.URL)

	for in, want := range map[string]template.URL{
		"smeggtuner.com":         "https://smeggtuner.com",
		"  smeggtuner.com  ":     "https://smeggtuner.com",
		"http://example.org":     "http://example.org",
		"https://example.org/x":  "https://example.org/x",
		"javascript://%0aalert1": "",
		"ftp://files.example":    "",
		"":                       "",
	} {
		if got := link(in); got != want {
			t.Errorf("site(%q) = %q, want %q", in, got, want)
		}
	}
}
