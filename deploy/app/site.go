// deploy/app/site.go

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// inlineScript is Nuxt's runtime-config script; the CSP allows exactly the ones the pages carry, by hash.
var inlineScript = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// encodings are the precompressed copies scripts/web.sh writes beside each file, best first.
var encodings = []struct{ name, ext string }{{"br", ".br"}, {"gzip", ".gz"}}

type site struct {
	dir string
	csp string
}

func newSite(dir string) (*site, error) {
	hashes := map[string]bool{}
	for _, page := range []string{"index.html", "200.html", "404.html"} {
		data, err := os.ReadFile(filepath.Join(dir, page))
		if err != nil {
			return nil, err
		}
		for _, m := range inlineScript.FindAllSubmatch(data, -1) {
			sum := sha256.Sum256(m[1])
			hashes["'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'"] = true
		}
	}
	var scripts []string
	for hash := range hashes {
		scripts = append(scripts, hash)
	}
	slices.Sort(scripts)
	csp := fmt.Sprintf("default-src 'self'; script-src 'self' 'wasm-unsafe-eval' %s; style-src 'self' 'unsafe-inline'; "+
		"img-src 'self' data: blob:; font-src 'self' data:; connect-src 'self'; worker-src 'self'; manifest-src 'self'; "+
		"object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'", strings.Join(scripts, " "))
	return &site{dir: dir, csp: csp}, nil
}

func (s *site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := path.Clean("/" + r.URL.Path)
	if name == "/" {
		name = "/index.html"
	}
	status := http.StatusOK
	file := filepath.Join(s.dir, filepath.FromSlash(name))
	if info, err := os.Stat(file); err != nil || info.IsDir() {
		// A path without an extension is a route of the single page; anything else is a missing file.
		name, status = "/200.html", http.StatusOK
		if path.Ext(r.URL.Path) != "" {
			name, status = "/404.html", http.StatusNotFound
		}
		file = filepath.Join(s.dir, filepath.FromSlash(name))
	}
	s.headers(w, name)
	s.serve(w, r, file, status)
}

func (s *site) headers(w http.ResponseWriter, name string) {
	h := w.Header()
	h.Set("Content-Security-Policy", s.csp)
	h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	h.Set("Permissions-Policy", "microphone=(self), camera=(), geolocation=(), payment=(), usb=()")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// Hashed chunks never change under one name; pages, the service worker and the engine revalidate every time.
	if strings.HasPrefix(name, "/_nuxt/") {
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		h.Set("Cache-Control", "no-cache")
	}
}

// serve sends file, or its precompressed copy when the client takes that encoding.
func (s *site) serve(w http.ResponseWriter, r *http.Request, file string, status int) {
	original, err := os.Stat(file)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if kind := mime.TypeByExtension(filepath.Ext(file)); kind != "" {
		w.Header().Set("Content-Type", kind)
	}
	sent := file
	for _, enc := range encodings {
		if _, err := os.Stat(file + enc.ext); err == nil {
			w.Header().Add("Vary", "Accept-Encoding")
			if accepts(r.Header.Get("Accept-Encoding"), enc.name) {
				w.Header().Set("Content-Encoding", enc.name)
				sent = file + enc.ext
				break
			}
		}
	}
	f, err := os.Open(sent)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	if status != http.StatusOK {
		w.WriteHeader(status)
		if r.Method == http.MethodGet {
			_, _ = f.WriteTo(w)
		}
		return
	}
	http.ServeContent(w, r, "", original.ModTime(), f)
}

// accepts reports whether an Accept-Encoding header takes name with a nonzero quality.
func accepts(header, name string) bool {
	for _, part := range strings.Split(header, ",") {
		token, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(strings.TrimSpace(token), name) {
			continue
		}
		q := strings.ReplaceAll(strings.TrimSpace(params), " ", "")
		return q != "q=0" && q != "q=0.0" && q != "q=0.00" && q != "q=0.000"
	}
	return false
}
