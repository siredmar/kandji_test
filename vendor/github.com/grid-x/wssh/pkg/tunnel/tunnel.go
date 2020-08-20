package tunnel

import (
	"context"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Tunnel is a connection from a Client to the Server
type Tunnel struct {
	ID          int
	ReadBuffer  chan []byte
	WriteBuffer chan []byte
	SigEmit     chan Signal
	SigSink     chan Signal
	log         logrus.FieldLogger
	conn        *websocket.Conn
	state       State
	stateMutex  sync.Mutex
	connMutex   sync.Mutex
}

// IDfromString parses a string into a Tunnel ID
func IDfromString(val string) int {
	v, _ := strconv.Atoi(val)
	return v
}

// State returns the current State of a Tunnel
func (t *Tunnel) State() State {
	t.stateMutex.Lock()
	defer t.stateMutex.Unlock()
	return t.state
}

func (t *Tunnel) setState(s State) {
	t.stateMutex.Lock()
	defer t.stateMutex.Unlock()
	if t.state != s {
		t.state = s
		t.log.WithField("state", s).Debug("change tunnel state")
	}
}

// Conn returns the underlying raw connection
func (t *Tunnel) Conn() *websocket.Conn {
	return t.conn
}

// SetConn sets the underlying raw connection
func (t *Tunnel) SetConn(c *websocket.Conn) {
	t.connMutex.Lock()
	defer t.connMutex.Unlock()
	t.conn = c
}

// function based FSM
// see: Rob Pike "Lexical Scanning in Go"
type stateFn func(context.Context, *Tunnel) stateFn

// Run concurrently until tunnel finished
func (t *Tunnel) Run(ctx context.Context) {
	for state := stateInit; state != nil; {
		state = state(ctx, t)
	}
	t.setState(StateFin)
	t.SigEmit <- SignalClose
}

// init: establish connection
func stateInit(ctx context.Context, t *Tunnel) stateFn {
	t.setState(StateInit)
	t.connMutex.Lock()
	defer t.connMutex.Unlock()
	// t.conn.SetPingHandler(func(appData string) error {
	// 	log.Debug("recv ping")
	// 	if err := t.conn.WriteMessage(websocket.PongMessage, nil); err != nil {
	// 		closeErr := websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway)
	// 		log.WithError(err).WithField("closeErr", closeErr).Error("send pong")
	// 		t.Signals <- SignalClose
	// 	}
	// 	log.Debug("send pong")
	// 	return nil // TODO: what happens with the returned error?
	// })
	return stateReady
}

// ready: exchange data
func stateReady(ctx context.Context, t *Tunnel) stateFn {
	t.setState(StateReady)
	t.connMutex.Lock()
	defer t.connMutex.Unlock()

	conn := t.Conn()

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()
	done := ctx.Done()

	// read
	go func() {
		for {
			select {
			case <-done:
				t.log.WithField("routine", "read").Trace("done")
				return
			default:
				_, b, err := conn.ReadMessage()
				if err != nil {
					unexpected := websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure)
					_l := t.log.WithField("unexpected", unexpected).WithError(err)
					if unexpected {
						_l.Error("read")
					} else {
						_l.Debug("read")
					}
					cancel()
				} else {
					t.log.WithField("b", string(b)).Trace("read")
					t.ReadBuffer <- b
				}
			}
		}
	}()

	// write
	go func() {
		for {
			select {
			case <-done:
				t.log.WithField("routine", "write").Trace("done")
				return
			case msg := <-t.WriteBuffer:
				err := conn.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					t.log.WithError(err).Error("writeSocket")
				} else {
					t.log.WithField("message", string(msg)).Trace("writeSocket")
				}
			}
		}
	}()

	t.SigEmit <- SignalOpen

signals:
	for {
		select {
		case sig := <-t.SigSink:
			t.log.WithField("sig", sig).Debug("recv")
			if sig == SignalClose {
				break signals
			}
		case <-done:
			t.log.WithField("routine", "signals").Trace("done")
			break signals
		}
	}

	return stateClosing
}

func stateClosing(ctx context.Context, t *Tunnel) stateFn {
	t.setState(StateClosing)
	t.connMutex.Lock()
	defer func() {
		t.connMutex.Unlock()
	}()

	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	if err := t.conn.WriteMessage(websocket.CloseMessage, msg); err != nil {
		t.log.WithError(err).Error("close connection")
		return nil
	}

	return nil
}
