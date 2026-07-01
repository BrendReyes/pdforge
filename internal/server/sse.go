package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type sseEvent struct {
	Type    string `json:"type"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

type broker struct {
	mu   sync.Mutex
	subs map[string]chan sseEvent
}

func newBroker() *broker {
	return &broker{subs: make(map[string]chan sseEvent)}
}

// register creates a buffered channel for token. Must be called before the
// operation goroutine starts so no events are lost.
func (b *broker) register(token string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[token] = make(chan sseEvent, 8)
}

// publish sends an event to the token's channel. Drops silently if full.
func (b *broker) publish(token string, e sseEvent) {
	b.mu.Lock()
	ch, ok := b.subs[token]
	b.mu.Unlock()
	if !ok {
		return
	}
	select {
	case ch <- e:
	default:
	}
}

// closeToken closes and removes the channel for token.
func (b *broker) closeToken(token string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, ok := b.subs[token]; ok {
		close(ch)
		delete(b.subs, token)
	}
}

// serveSSE streams events for the token in the URL path value.
func (b *broker) serveSSE(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	b.mu.Lock()
	ch, ok := b.subs[token]
	b.mu.Unlock()
	if !ok {
		http.Error(w, "unknown token", http.StatusNotFound)
		return
	}

	// Extend the write deadline so the connection survives long-running operations.
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Now().Add(5 * time.Minute))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx proxy buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case event, open := <-ch:
			if !open {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			if event.Type == "done" || event.Type == "error" {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}
