package session

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coresession "smegg.me/smeggtuner/core/session"
)

var frontend = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("the frontend"))
})

func jpegBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	img.Set(0, 0, color.RGBA{R: 200, A: 255})

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func photographed(t *testing.T, s *Service) string {
	t.Helper()

	got, err := s.SaveInstrumentSpec(coresession.Template{
		Name:       "Morino",
		Instrument: coresession.Instrument{ReedCount: 3},
	})
	if err != nil {
		t.Fatal(err)
	}

	jpg, err := coresession.PrepareImage(bytes.NewReader(jpegBytes(t)))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.templates().SetImage(got.ID, jpg); err != nil {
		t.Fatal(err)
	}
	return got.ID
}

func get(t *testing.T, s *Service, path string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	ImageMiddleware(s)(frontend).ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestAPhotographIsServedAsAPhotograph(t *testing.T) {
	s := service(t)
	id := photographed(t, s)

	w := get(t, s, "/instruments/"+id+"/image")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", ct)
	}
	if _, err := jpeg.Decode(bytes.NewReader(w.Body.Bytes())); err != nil {
		t.Fatalf("what came back is not a JPEG: %v", err)
	}

	if w.Header().Get("Etag") == "" {
		t.Fatal("no ETag, so the browser cannot revalidate and will refetch every time")
	}
}

func TestEverythingElseIsTheFrontends(t *testing.T) {
	s := service(t)

	for _, path := range []string{"/", "/index.html", "/assets/app.js", "/instruments"} {
		w := get(t, s, path)
		if w.Body.String() != "the frontend" {
			t.Fatalf("%s was intercepted: %q", path, w.Body.String())
		}
	}
}

func TestAnInstrumentWithNoPhotographIsNotAnError(t *testing.T) {
	s := service(t)
	got, err := s.SaveInstrumentSpec(coresession.Template{
		Name:       "Plain",
		Instrument: coresession.Instrument{ReedCount: 1},
	})
	if err != nil {
		t.Fatal(err)
	}

	if w := get(t, s, "/instruments/"+got.ID+"/image"); w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// A traversal id must fall through, not even a 404 that would confirm a lookup.
func TestTheImageHandlerCannotBeTalkedOutOfTheLibrary(t *testing.T) {
	s := service(t)

	for _, path := range []string{
		"/instruments/../../etc/passwd/image",
		"/instruments/a/../../b/image",
		"/instruments//image",
		"/instruments/.ssh/image",
	} {
		if _, ok := instrumentImageID(path); ok {
			t.Fatalf("%s was claimed by the image handler", path)
		}
		if body := get(t, s, path).Body.String(); body != "the frontend" {
			t.Fatalf("%s was answered by the image handler: %q", path, body)
		}
	}
}

func TestTheRevisionMovesWhenThePhotographDoes(t *testing.T) {
	s := service(t)
	id := photographed(t, s)

	first := rev(t, s, id)
	if first == 0 {
		t.Fatal("a photographed instrument has no revision")
	}

	jpg, err := coresession.PrepareImage(bytes.NewReader(jpegBytes(t)))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond) // stamp is time-based; wait so the second write gets a distinct one
	if err := s.templates().SetImage(id, jpg); err != nil {
		t.Fatal(err)
	}

	if second := rev(t, s, id); second == first {
		t.Fatal("the revision did not move, so the shelf will go on showing the old photograph")
	}
}

func rev(t *testing.T, s *Service, id string) int64 {
	t.Helper()

	all, err := s.Instruments()
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range all {
		if i.ID == id {
			return i.ImageRev
		}
	}
	t.Fatalf("no such instrument: %s", id)
	return 0
}
