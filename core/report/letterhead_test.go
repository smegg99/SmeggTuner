package report

import (
	"html/template"
	"strings"
	"testing"
)

func TestLetterheadIsOffByDefault(t *testing.T) {
	s := musette(t)

	plain := render(t, sheet(t, s))
	if strings.Contains(plain, "Smegg Accordion Service") {
		t.Fatal("the letterhead printed without being asked for")
	}

	r, err := Build(s, Options{
		Now: at,
		Letterhead: &Letterhead{
			CompanyName:    "Smegg Accordion Service",
			CompanyAddress: "ul. Warsztatowa 1\n00-001 Warszawa",
			CompanyWebsite: "smegg.me",
		},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	out := render(t, r)

	for _, want := range []string{"Smegg Accordion Service", "ul. Warsztatowa 1", "smegg.me"} {
		if !strings.Contains(out, want) {
			t.Errorf("the letterhead is missing %q", want)
		}
	}
}

// THE CARD'S OWN BRANDING IS NOT A SETTING.
//
// The letterhead used to carry an image the technician chose, which is a slot to paint over the
// app's mark with. It carries words now, and the mascot prints whatever else is on the sheet.
func TestTheLetterheadCarriesNoArtwork(t *testing.T) {
	r, err := Build(musette(t), Options{
		Now:        at,
		Letterhead: &Letterhead{CompanyName: "Smegg Accordion Service"},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	out := render(t, r)

	if strings.Contains(out, "<img") {
		t.Error("the sheet carries an <img>: the letterhead can put a picture on the card again")
	}
	if !strings.Contains(out, "mascot") {
		t.Error("the app's mark is missing from a card that carries a letterhead")
	}
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
