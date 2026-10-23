// backend/internal/androidhost/request.go
package androidhost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var errUnavailable = errors.New("native Android host is unavailable")

type reply struct {
	Value string `json:"value"`
	Error string `json:"error"`
}

type broker struct {
	mu      sync.Mutex
	next    uint64
	pending map[uint64]chan reply
	send    func(uint64, []byte) bool
}

var host = broker{pending: make(map[uint64]chan reply)}

// CacheDirectory returns Android's private, disposable storage directory.
func CacheDirectory(ctx context.Context) (string, error) {
	return host.request(ctx, struct {
		Kind string `json:"kind"`
	}{"cache"})
}

func (b *broker) request(ctx context.Context, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	b.mu.Lock()
	send := b.send
	if send == nil {
		b.mu.Unlock()
		return "", errUnavailable
	}
	b.next++
	id := b.next
	result := make(chan reply, 1)
	b.pending[id] = result
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		delete(b.pending, id)
		b.mu.Unlock()
	}()
	if !send(id, data) {
		return "", errUnavailable
	}
	select {
	case <-ctx.Done():
		cancel, _ := json.Marshal(struct {
			Kind   string `json:"kind"`
			Target uint64 `json:"target"`
		}{"cancel", id})
		send(0, cancel)
		return "", ctx.Err()
	case r := <-result:
		if r.Error != "" {
			return "", fmt.Errorf("native Android operation: %s", r.Error)
		}
		return r.Value, nil
	}
}

func (b *broker) complete(id uint64, data []byte) {
	var r reply
	if err := json.Unmarshal(data, &r); err != nil {
		r.Error = fmt.Sprintf("invalid native reply: %v", err)
	}
	b.mu.Lock()
	result := b.pending[id]
	delete(b.pending, id)
	b.mu.Unlock()
	if result != nil {
		result <- r
	}
}

// SaveDocument asks Android for a writable document URI; an empty result means cancellation.
func SaveDocument(ctx context.Context, name, title, mime string) (string, error) {
	return host.request(ctx, struct {
		Kind  string `json:"kind"`
		Name  string `json:"name"`
		Title string `json:"title"`
		MIME  string `json:"mime"`
	}{"save", name, title, mime})
}

// WriteDocument copies a finished local file into the document selected by the user.
func WriteDocument(ctx context.Context, uri, path string) error {
	_, err := host.request(ctx, struct {
		Kind string `json:"kind"`
		URI  string `json:"uri"`
		Path string `json:"path"`
	}{"write", uri, path})
	return err
}

// OpenDocument hands a document to an Android viewer with a scoped read grant.
func OpenDocument(ctx context.Context, uri, mime string) error {
	_, err := host.request(ctx, struct {
		Kind string `json:"kind"`
		URI  string `json:"uri"`
		MIME string `json:"mime"`
	}{"open", uri, mime})
	return err
}

// PrintDocument hands self-contained HTML to Android's print UI, including Save as PDF.
func PrintDocument(ctx context.Context, html []byte, landscape bool, name string) error {
	_, err := host.request(ctx, struct {
		Kind      string `json:"kind"`
		HTML      string `json:"html"`
		Landscape bool   `json:"landscape"`
		Name      string `json:"name"`
	}{"print", string(html), landscape, name})
	return err
}
