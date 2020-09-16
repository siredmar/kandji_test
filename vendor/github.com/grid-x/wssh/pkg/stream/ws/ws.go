package ws

import (
	"errors"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

// WS wraps a websocket connection to fulfill the Stream interface.
// FIXME: this is currently broken and unused on the server side.
//        it does work fine for the client side though...
type WS struct {
	conn       *websocket.Conn
	reader     io.Reader
	mutexRead  sync.Mutex
	mutexWrite sync.Mutex
}

// New WS stream, using an open websocket connection.
// TODO: open websocket connection from supplied options
func New(conn *websocket.Conn) *WS {
	return &WS{
		conn: conn,
	}
}

// Read from websocket into buffer.
func (ws *WS) Read(b []byte) (int, error) {
	ws.mutexRead.Lock()
	defer ws.mutexRead.Unlock()

	for {
		reader, err := ws.Reader()
		if err != nil {
			return 0, err
		}

		n, err := reader.Read(b)
		if err != nil {
			ws.reader = nil
			if err == io.EOF {
				if n > 0 {
					return n, nil
				}
				continue
			}
		}
		return n, err
	}

}

// Reader returns a websocket reader.
func (ws *WS) Reader() (io.Reader, error) {
	if ws.reader == nil {
		msgType, reader, err := ws.conn.NextReader()
		if err != nil {
			return nil, err
		}

		if msgType != websocket.BinaryMessage {
			return nil, errors.New("websocket read non-binary message")
		}
		ws.reader = reader
	}

	return ws.reader, nil
}

// Write from buffer into websocket.
func (ws *WS) Write(b []byte) (int, error) {
	ws.mutexWrite.Lock()
	defer ws.mutexWrite.Unlock()

	writer, err := ws.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	n, err := writer.Write(b)
	writer.Close()

	return n, err
}

// Close the websocket connection.
func (ws *WS) Close() error {
	return ws.conn.Close()
}
