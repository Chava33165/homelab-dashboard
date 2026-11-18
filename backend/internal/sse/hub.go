package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type message struct {
	event string
	data  []byte
}

type client struct {
	ch chan message
}

// Hub multiplexa eventos de servidor->cliente sobre un unico endpoint SSE. Se eligio
// SSE sobre WebSockets porque todo el trafico relevante es servidor->cliente; las
// mutaciones (start/stop/restart, etc) van por REST normal. Ver CLAUDE.md.
type Hub struct {
	mu      sync.Mutex
	clients map[*client]bool
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*client]bool)}
}

// Publish manda `data` (serializado a JSON) como un evento `event` a todos los clientes
// conectados. Si un cliente esta lento y su buffer se llena, se descarta ese evento
// para ese cliente en vez de bloquear al resto.
func (h *Hub) Publish(event string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := message{event: event, data: payload}

	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.ch <- msg:
		default:
		}
	}
}

// ServeHTTP atiende una conexion SSE. El caller debe haber validado auth antes de llamar esto.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming no soportado", http.StatusInternalServerError)
		return
	}

	c := &client{ch: make(chan message, 16)}
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepAlive.C:
			fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case msg := <-c.ch:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg.event, msg.data)
			flusher.Flush()
		}
	}
}
