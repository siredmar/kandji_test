package ssh

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// ConnectionRepository provides methods to manage connections
type ConnectionRepository struct {
	logger log.FieldLogger
	uuid   GetUUID

	// cache holds a map of connections
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]*Connection
}

// Connection represents a ssh session
type Connection struct {
	PodName string
	Conn    *websocket.Conn
}

// NewConnectionRepository creates a new connection repository with the given settings
func NewConnectionRepository(
	logger log.FieldLogger,
	uuid GetUUID,
) (*ConnectionRepository, error) {
	r := &ConnectionRepository{
		logger: logger,
		uuid:   uuid,
		cache:  make(map[string]*Connection),
	}

	return r, nil
}

// Add adds a new connection
func (r *ConnectionRepository) Add(connection *Connection, deviceID string) {
	r.cache[deviceID] = connection
}

// Remove deletes a connection
func (r *ConnectionRepository) Remove(deviceID string) {
	delete(r.cache, deviceID)
}

// Get gets the connection with the given address
func (r *ConnectionRepository) Get(deviceID string) (conn *Connection) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.cache[deviceID]
}

// GetUUID retrieves a random v4 uuid
func (r *ConnectionRepository) GetUUID() uuid.UUID {
	return r.uuid()
}
