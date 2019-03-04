package ssh

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Session represents a ssh session
type Session struct {
	ID         string
	ClientConn *websocket.Conn
	DeviceConn *websocket.Conn
}

// SessionRepository provides methods to manage sessions
type SessionRepository struct {
	logger log.FieldLogger
	uuid   GetUUID

	// cache holds a map of sessions
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]*Session
}

// NewSessionRepository creates a new session repository with the given settings
func NewSessionRepository(
	logger log.FieldLogger,
	uuid GetUUID,
) (*SessionRepository, error) {
	r := &SessionRepository{
		logger: logger,
		uuid:   uuid,
		cache:  make(map[string]*Session),
	}

	return r, nil
}

// Add adds a new session
func (r *SessionRepository) Add(sess *Session) {
	r.cache[sess.ID] = sess
}

// Remove deletes a session
func (r *SessionRepository) Remove(session string) {
	delete(r.cache, session)
}

// Get gets a session
func (r *SessionRepository) Get(session string) *Session {
	return r.cache[session]
}

// GetByConnection gets the session with the given connection
func (r *SessionRepository) GetByConnection(sconn *websocket.Conn) []*Session {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var w []*Session
	for _, v := range r.cache {
		if v.DeviceConn == sconn {
			w = append(w, v)
		}
	}
	return w
}

// GetUUID retrieves a random v4 uuid
func (r *SessionRepository) GetUUID() uuid.UUID {
	return r.uuid()
}
