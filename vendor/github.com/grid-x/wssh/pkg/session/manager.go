package session

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/grid-x/wssh/internal"
	"github.com/grid-x/wssh/internal/k8s"
	"github.com/grid-x/wssh/pkg/config"
	"github.com/grid-x/wssh/pkg/tunnel"
)

// prom methods used by session FSMs
type promRepository interface {
	DeviceConnectDurationSecondsObserve(internal.SessionContext) func(float64)
	SessionsActiveInc(internal.SessionContext)
	SessionsActiveDec(internal.SessionContext)
	SessionsStateInc(internal.SessionContext, fmt.Stringer)
	SessionsStateDec(internal.SessionContext, fmt.Stringer)
	SessionsTransmittedBytesAdd(internal.SessionContext) func(int)
	SessionsDurationSecondsObserve(internal.SessionContext) func(float64)
}

// ManagerI is the managers Interface
type ManagerI interface {
	Get(sID uuid.UUID) (*Session, bool)
	Delete(sID uuid.UUID)
	NewMaybe(agentID string, deviceID string, flavor SSHFlavor) (*Session, bool, error)
}

// Manager is used to savely modify Session state
type Manager struct {
	Prom promRepository

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
	promRepo promRepository,
) *Manager {
	m := &Manager{
		Prom:     promRepo,
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
func (m *Manager) NewMaybe(agentID string, deviceID string, flavor SSHFlavor) (*Session, bool, error) {
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
		SSHFlavor:      flavor,
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
