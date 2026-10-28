// deploy/app/site_test.go

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const config = `window.__NUXT__={config:{}}`

// fixture lays out a site the way scripts/web.sh does: pages with Nuxt's inline script, a hashed chunk, the engine and
// precompressed copies beside some files.
func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	page := `<html><script type="module" src="/_nuxt/a.js"></script><script>` + config + `</script>` +
		`<script type="application/json" id="__NUXT_DATA__">[]</script></html>`
	files := map[string]string{
		"index.html": page, "200.html": page, "404.html": "<html>missing</html>",
		"_nuxt/a.js": "plain", "_nuxt/a.js.br": "brotli", "_nuxt/a.js.gz": "gzip",
		"engine/smeggtuner.wasm": "wasm", "engine/smeggtuner.wasm.br": "wasm-brotli",
	}
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func get(t *testing.T, s *site, target, encoding string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, target, nil)
	if encoding != "" {
		r.Header.Set("Accept-Encoding", encoding)
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestTheCSPAllowsOnlyThePagesInlineScriptAndWasm(t *testing.T) {
	s, err := newSite(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(config))
	want := "script-src 'self' 'wasm-unsafe-eval' 'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	var scripts string
	for _, directive := range strings.Split(s.csp, ";") {
		if strings.HasPrefix(strings.TrimSpace(directive), "script-src ") {
			scripts = strings.TrimSpace(directive)
		}
	}
	if scripts != want {
		t.Fatalf("script-src = %q, want %q", scripts, want)
	}
	w := get(t, s, "/", "")
	if w.Code != 200 || w.Header().Get("Content-Security-Policy") != s.csp || w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("/ = %d %v", w.Code, w.Header())
	}
	if !strings.Contains(w.Header().Get("Permissions-Policy"), "microphone=(self)") {
		t.Fatalf("permissions = %q", w.Header().Get("Permissions-Policy"))
	}
}

func TestPrecompressedCopiesFollowAcceptEncoding(t *testing.T) {
	s, _ := newSite(fixture(t))
	for _, c := range []struct{ accept, encoding, body string }{
		{"gzip, deflate, br", "br", "brotli"},
		{"gzip", "gzip", "gzip"},
		{"br;q=0, gzip", "gzip", "gzip"},
		{"", "", "plain"},
	} {
		w := get(t, s, "/_nuxt/a.js", c.accept)
		if w.Header().Get("Content-Encoding") != c.encoding || w.Body.String() != c.body {
			t.Errorf("Accept-Encoding %q: got %q %q", c.accept, w.Header().Get("Content-Encoding"), w.Body.String())
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "javascript") || w.Header().Get("Vary") == "" {
			t.Errorf("Accept-Encoding %q: headers %v", c.accept, w.Header())
		}
		if !strings.Contains(w.Header().Get("Cache-Control"), "immutable") {
			t.Errorf("a hashed chunk must be immutable: %v", w.Header())
		}
	}
}

func TestTheEngineIsWasmAndRevalidates(t *testing.T) {
	s, _ := newSite(fixture(t))
	w := get(t, s, "/engine/smeggtuner.wasm", "br")
	if w.Header().Get("Content-Type") != "application/wasm" || w.Header().Get("Content-Encoding") != "br" ||
		w.Header().Get("Cache-Control") != "no-cache" || w.Body.String() != "wasm-brotli" {
		t.Fatalf("engine = %v %q", w.Header(), w.Body.String())
	}
}

func TestRoutesGetTheAppAndMissingFilesA404(t *testing.T) {
	s, _ := newSite(fixture(t))
	if w := get(t, s, "/settings", ""); w.Code != 200 || !strings.Contains(w.Body.String(), config) {
		t.Fatalf("route = %d %q", w.Code, w.Body.String())
	}
	if w := get(t, s, "/_nuxt/gone.js", ""); w.Code != 404 || w.Body.String() != "<html>missing</html>" {
		t.Fatalf("missing file = %d %q", w.Code, w.Body.String())
	}
	if w := get(t, s, "/../../etc/passwd", ""); w.Code != 404 && w.Code != 200 || strings.Contains(w.Body.String(), "root:") {
		t.Fatalf("traversal = %d %q", w.Code, w.Body.String())
	}
}

func TestOnlyReadsAreServed(t *testing.T) {
	s, _ := newSite(fixture(t))
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST = %d", w.Code)
	}
}
