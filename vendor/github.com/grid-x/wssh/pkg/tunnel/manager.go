package tunnel

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Manager is used to savely modify Tunnel state
type Manager struct {
	tunnels map[uuid.UUID]*Tunnel
	log     logrus.FieldLogger
	n       int
}

// NewManager returns a new Manager
func NewManager(log logrus.FieldLogger) *Manager {
	m := &Manager{
		tunnels: make(map[uuid.UUID]*Tunnel),
		log:     log,
		n:       1,
	}

	return m
}

// Get a Tunnel by its ID
func (m *Manager) Get(tID uuid.UUID) (*Tunnel, error) {
	t, ok := m.tunnels[tID]
	if !ok {
		return nil, fmt.Errorf("tunnel with tID %v does not exist", tID)
	}

	return t, nil
}

// New returns a new Tunnel
func (m *Manager) New() (*Tunnel, error) {
	tID := m.nextID()
	_, ok := m.tunnels[tID]
	if ok {
		return nil, fmt.Errorf("tunnel with tID %v already exists", tID)
	}

	t := &Tunnel{
		ID:          tID,
		ReadBuffer:  make(chan []byte),
		WriteBuffer: make(chan []byte),
		SigEmit:     make(chan Signal),
		SigSink:     make(chan Signal),
		log:         m.log.WithField("tID", tID),
	}
	m.tunnels[tID] = t

	return t, nil
}

func (m *Manager) nextID() uuid.UUID {
	return uuid.New()
}
