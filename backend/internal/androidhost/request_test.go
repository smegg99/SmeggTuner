// backend/internal/androidhost/request_test.go
package androidhost

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestReplyIsDeliveredOnce(t *testing.T) {
	b := broker{pending: make(map[uint64]chan reply)}
	b.send = func(id uint64, _ []byte) bool {
		b.complete(id, []byte(`{"value":"content://documents/one","error":""}`))
		b.complete(id, []byte(`{"value":"duplicate","error":""}`))
		return true
	}
	value, err := b.request(context.Background(), map[string]string{"kind": "save"})
	if err != nil || value != "content://documents/one" {
		t.Fatalf("reply = %q, %v", value, err)
	}
	if len(b.pending) != 0 {
		t.Fatal("completed request retained")
	}
}

func TestCancellationCleansNativeRequest(t *testing.T) {
	b := broker{pending: make(map[uint64]chan reply)}
	ctx, cancel := context.WithCancel(context.Background())
	var requestID, cancelledID uint64
	b.send = func(id uint64, data []byte) bool {
		if id != 0 {
			requestID = id
			cancel()
			return true
		}
		var command struct {
			Kind   string `json:"kind"`
			Target uint64 `json:"target"`
		}
		if err := json.Unmarshal(data, &command); err != nil {
			t.Fatal(err)
		}
		if command.Kind != "cancel" {
			t.Fatalf("unexpected command %q", command.Kind)
		}
		cancelledID = command.Target
		return true
	}
	_, err := b.request(ctx, map[string]string{"kind": "pdf"})
	if !errors.Is(err, context.Canceled) || cancelledID != requestID || requestID == 0 {
		t.Fatalf("cancellation = %d, request = %d, error = %v", cancelledID, requestID, err)
	}
	b.complete(requestID, []byte(`{"value":"late","error":""}`))
	if len(b.pending) != 0 {
		t.Fatal("cancelled request retained")
	}
}

func TestNativeFailuresAreNotCancellation(t *testing.T) {
	for _, body := range []string{`{"value":"","error":"permission denied"}`, `not json`} {
		b := broker{pending: make(map[uint64]chan reply)}
		b.send = func(id uint64, _ []byte) bool {
			b.complete(id, []byte(body))
			return true
		}
		if _, err := b.request(context.Background(), struct{}{}); err == nil {
			t.Fatalf("native failure %q reported success", body)
		}
	}
}

func TestUnavailableHostDoesNotRetainRequests(t *testing.T) {
	b := broker{pending: make(map[uint64]chan reply)}
	for _, send := range []func(uint64, []byte) bool{nil, func(uint64, []byte) bool { return false }} {
		b.send = send
		_, err := b.request(context.Background(), struct{}{})
		if !errors.Is(err, errUnavailable) || len(b.pending) != 0 {
			t.Fatalf("unavailable = %v, pending = %d", err, len(b.pending))
		}
	}
}

func TestUnicodePayloadSurvivesTransport(t *testing.T) {
	b := broker{pending: make(map[uint64]chan reply)}
	want := "Za\u017c\u00f3\u0142\u0107 \U0001f3b5"
	b.send = func(id uint64, data []byte) bool {
		var payload map[string]string
		if err := json.Unmarshal(data, &payload); err != nil || payload["name"] != want {
			t.Fatalf("payload = %q, %v", data, err)
		}
		encoded, _ := json.Marshal(reply{Value: want})
		b.complete(id, encoded)
		return true
	}
	value, err := b.request(context.Background(), map[string]string{"name": want})
	if err != nil || !strings.EqualFold(value, want) {
		t.Fatalf("unicode reply = %q, %v", value, err)
	}
}
