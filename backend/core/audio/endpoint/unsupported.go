// backend/core/audio/endpoint/unsupported.go

//go:build android || (!linux && !windows)

package endpoint

const supported = false

func open([]byte) (Endpoint, error) { return nil, ErrUnsupported }
