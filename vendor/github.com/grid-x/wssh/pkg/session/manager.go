package session

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/grid-x/wssh/internal/k8s"
	"github.com/grid-x/wssh/pkg/config"
	"github.com/grid-x/wssh/pkg/tunnel"
)

// ManagerI is the managers Interface
type ManagerI interface {
	Delete(sID uuid.UUID)
	NewMaybe(agentID string, deviceID string) (*Session, bool, error)
}

// Manager is used to savely modify Session state
type Manager struct {
	cfg              *config.ServerConfig
	log              logrus.FieldLogger
	sessions         map[uuid.UUID]*Session
	pods             *k8s.PodsRepository
	tunnel           *tunnel.Manager
	listenAddr       string
	debugDeviceSpawn bool
	debugDeviceAddr  string
}

// NewManager returns a new Manager
func NewManager(
	cfg *config.ServerConfig,
	log logrus.FieldLogger,
	tunnelManager *tunnel.Manager,
	pods *k8s.PodsRepository,
) *Manager {
	m := &Manager{
		cfg:      cfg,
		tunnel:   tunnelManager,
		log:      log,
		sessions: make(map[uuid.UUID]*Session),
		pods:     pods,
	}

	return m
}

// Config returns the server config
func (m *Manager) Config() config.ServerConfig {
	return *m.cfg
}

// Get a Session by its ID
func (m *Manager) Get(sID uuid.UUID) (*Session, bool) {
	s, ok := m.sessions[sID]
	return s, ok
}

// Delete a Session by its ID
func (m *Manager) Delete(sID uuid.UUID) {
	m.log.WithField("sID", sID).Debug("delete session")
	delete(m.sessions, sID)
}

// NewMaybe returns a Session
func (m *Manager) NewMaybe(agentID string, deviceID string) (*Session, bool, error) {
	// TODO: implement session reconnect
	// TODO: check if clients exist
	tunnelAgent, err := m.tunnel.New()
	if err != nil {
		return nil, false, err
	}
	tunnelDevice, err := m.tunnel.New()
	if err != nil {
		return nil, false, err
	}

	// new session
	sID := m.nextID()
	s := &Session{
		ID:             sID,
		AgentID:        agentID,
		AgentTunnelID:  tunnelAgent.ID,
		DeviceID:       deviceID,
		DeviceTunnelID: tunnelDevice.ID,
		Signals:        make(chan Signal),
		log:            m.log.WithField("sID", sID),
		manager:        m,
	}

	m.sessions[sID] = s

	return s, true, nil
}

// AgentTunnel returns the Agent Tunnel for a specific Session
func (m *Manager) AgentTunnel(sID uuid.UUID) (*tunnel.Tunnel, error) {
	s, ok := m.Get(sID)
	if !ok {
		return nil, fmt.Errorf("session with sID '%v' does not exist", sID)
	}
	t, err := m.tunnel.Get(s.AgentTunnelID)
	return t, err
}

// DeviceTunnel returns the Device Tunnel for a specific Session
func (m *Manager) DeviceTunnel(sID uuid.UUID) (*tunnel.Tunnel, error) {
	s, ok := m.Get(sID)
	if !ok {
		return nil, fmt.Errorf("session with sID '%v' does not exist", sID)
	}
	t, err := m.tunnel.Get(s.DeviceTunnelID)
	return t, err
}

// AgentTunnelReady returns true iff the Agent Tunnel is ready for a specific Session
func (m *Manager) AgentTunnelReady(sID uuid.UUID) bool {
	t, err := m.AgentTunnel(sID)
	if err != nil {
		return false
	}
	return t.State() == tunnel.StateReady
}

// DeviceTunnelReady returns true iff the Device Tunnel is ready for a specific Session
func (m *Manager) DeviceTunnelReady(sID uuid.UUID) bool {
	t, err := m.DeviceTunnel(sID)
	if err != nil {
		return false
	}
	return t.State() == tunnel.StateReady
}

func (m *Manager) nextID() uuid.UUID {
	return uuid.New()
}
