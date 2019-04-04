package messaging

import (
	"sync"

	"github.com/gorilla/websocket"
)

// WebsocketWriter protected websocket writes with a mutex to avoid concurrent wrties
type WebsocketWriter struct {
	Socket *websocket.Conn // websocket connection of the player
	mu     sync.Mutex
}

// NewWebsocketWriter creates a new websocket writer
func NewWebsocketWriter(c *websocket.Conn) *WebsocketWriter {
	w := &WebsocketWriter{
		Socket: c,
	}
	return w
}

// WriteJSON writes JSON to the socket
func (w *WebsocketWriter) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Socket.WriteJSON(v)
}

// WriteMessage writes a plain message to the socket
func (w *WebsocketWriter) WriteMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Socket.WriteMessage(messageType, data)
}
