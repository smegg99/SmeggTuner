// Photographs go over the asset server, not bindings: a JSON-RPC binding would base64 every JPEG.
package session

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ImagePrefix is the path prefix this middleware claims.
const ImagePrefix = "/instruments/"

const imageSuffix = "/image"

// ImageMiddleware serves instrument photographs as asset-server middleware; a package func, so Wails never binds it.
func ImageMiddleware(s *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return imageHandler(s, next)
	}
}

func imageHandler(s *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := instrumentImageID(r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		jpg, rev, err := s.templates().Image(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// ETag lets ServeContent answer If-None-Match with 304 so re-renders revalidate instead of refetching.
		w.Header().Set("ETag", fmt.Sprintf(`"%d-%d"`, rev, len(jpg)))
		w.Header().Set("Cache-Control", "no-cache")

		http.ServeContent(w, r, "image.jpg", time.Unix(0, rev), bytes.NewReader(jpg))
	})
}

// instrumentImageID extracts the id from /instruments/<id>/image; any other shape falls through, so a traversal id is never served.
func instrumentImageID(path string) (string, bool) {
	if !strings.HasPrefix(path, ImagePrefix) || !strings.HasSuffix(path, imageSuffix) {
		return "", false
	}

	id := strings.TrimSuffix(strings.TrimPrefix(path, ImagePrefix), imageSuffix)
	if id == "" || strings.ContainsAny(id, "/.\\") {
		return "", false
	}
	return id, true
}
