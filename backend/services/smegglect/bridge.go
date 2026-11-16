//go:build !production

package smegglect

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Origin marks forwarded calls as the page's; the broker must admit it (`serve --origin wails://wails`).
const Origin = "wails://wails"

var (
	bridgeURL = regexp.MustCompile(`^http://127\.0\.0\.1:\d{1,5}$`)
	// The overlay's own paths and queries; a query holds ids, flags, numbers and the overlay's version.
	bridgePath = regexp.MustCompile(`^/v1/[A-Za-z0-9_./-]*(\?[A-Za-z0-9_.=&-]*)?$`)
)

// ErrRequest means the window asked the bridge proxy for something the overlay never sends.
var ErrRequest = errors.New("smegglect: unexpected bridge request")

// BridgeRequest is one overlay request for the local Smegglect broker.
type BridgeRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   string `json:"body,omitempty"`
	Type   string `json:"type,omitempty"`
	// Base64 marks Body as base64, e.g. an attached screenshot.
	Base64 bool `json:"base64,omitempty"`
}

// BridgeReply is the broker's answer; Status 0 means no broker is running for the project.
type BridgeReply struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// Bridge forwards one overlay request to the broker with its token, which the window never sees.
func (s *Service) Bridge(ctx context.Context, req BridgeRequest) (BridgeReply, error) {
	body, err := bridgeBody(req)
	if err != nil {
		return BridgeReply{}, err
	}
	url, token, ok := s.broker()
	if !ok {
		return BridgeReply{}, nil
	}
	call, err := http.NewRequestWithContext(ctx, req.Method, url+req.Path, body)
	if err != nil {
		return BridgeReply{}, fmt.Errorf("smegglect: bridge request: %w", err)
	}
	call.Header.Set("Authorization", "Bearer "+token)
	call.Header.Set("Origin", Origin)
	if req.Type != "" {
		call.Header.Set("Content-Type", req.Type)
	}
	res, err := http.DefaultClient.Do(call)
	if err != nil {
		// A stopped broker or a call the window cancelled; neither is worth an error in the log.
		return BridgeReply{}, nil
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, MaxBytes+1))
	if err != nil {
		return BridgeReply{}, nil
	}
	if len(raw) > MaxBytes {
		return BridgeReply{}, fmt.Errorf("smegglect: bridge answer over %d bytes", MaxBytes)
	}
	return BridgeReply{Status: res.StatusCode, Body: string(raw)}, nil
}

func bridgeBody(req BridgeRequest) (io.Reader, error) {
	switch req.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut:
	default:
		return nil, ErrRequest
	}
	if !bridgePath.MatchString(req.Path) || strings.Contains(req.Path, "..") {
		return nil, ErrRequest
	}
	if len(req.Body) > base64.StdEncoding.EncodedLen(MaxBytes) {
		return nil, fmt.Errorf("smegglect: bridge request over %d bytes", MaxBytes)
	}
	if !req.Base64 {
		return strings.NewReader(req.Body), nil
	}
	raw, err := base64.StdEncoding.DecodeString(req.Body)
	if err != nil {
		return nil, fmt.Errorf("smegglect: decode: %w", err)
	}
	return bytes.NewReader(raw), nil
}

// broker reads the address and token the running broker wrote for the project; it changes on every start.
func (s *Service) broker() (string, string, bool) {
	raw, err := os.ReadFile(filepath.Join(cmp.Or(s.Root, "."), ".smegglect", "bridge.json"))
	if err != nil {
		return "", "", false
	}
	var info struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	if json.Unmarshal(raw, &info) != nil || !bridgeURL.MatchString(info.URL) || info.Token == "" {
		return "", "", false
	}
	return info.URL, info.Token, true
}
