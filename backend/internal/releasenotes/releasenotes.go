// backend/internal/releasenotes/releasenotes.go
package releasenotes

import _ "embed"

//go:generate cp ../../../CHANGELOG.md CHANGELOG.md

// Changelog contains the bundled release notes in each supported language.
//
//go:embed CHANGELOG.md
var Changelog string
