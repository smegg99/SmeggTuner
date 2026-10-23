// backend/internal/documents/documents.go
package documents

import "strings"

// SaveOptions describes the native destination picker.
type SaveOptions struct {
	Name    string
	Title   string
	Filter  string
	Pattern string
	MIME    string
}

// IsURI distinguishes Android document handles from local filesystem paths.
func IsURI(path string) bool { return strings.HasPrefix(path, "content://") }
