package messaging

import (
	"fmt"
	"time"

	nats "github.com/nats-io/go-nats"
)

// Some commonly used NATS subjects
const (
	SSHSubject   = "ssh"
	DeviceChan   = "device"
	ClientChan   = "client"
	StatusChan   = "status"
	RegisterChan = "register"
)

// NATSRepository manages the access to nats
type NATSRepository struct {
	nc *nats.EncodedConn
}

// NewNATSRepository creates a new NATS repository containing a open connection to NATS server
func NewNATSRepository(nats *nats.EncodedConn) (*NATSRepository, error) {
	r := &NATSRepository{
		nc: nats,
	}
	return r, nil
}

// RegisterAtDevice registers a a new client session at the device
func (m *NATSRepository) RegisterAtDevice(deviceID string, sessionID string, initCommand string, timeout time.Duration) error {
	var resp string

	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.register.2a598bf6-7591-45f1-979f-1fd07f97704c
	subject := fmt.Sprintf("%s.%s.%s.%s", SSHSubject, deviceID, RegisterChan, sessionID)

	if m.nc != nil {
		err := m.nc.Request(subject, fmt.Sprintf("%s$%s", sessionID, initCommand), &resp, timeout)
		if err != nil {
			return err
		}
	}
	if resp == "OK" {
		return nil
	}

	return fmt.Errorf("Incorrect answer")
}

// ListenForClientReg listens for new client registrations
func (m *NATSRepository) ListenForClientReg(deviceID string, sessionC chan string) (*nats.Subscription, error) {
	var sub *nats.Subscription
	var err error

	// Use * to catch all session ID's
	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.register.*
	subject := fmt.Sprintf("%s.%s.%s.*", SSHSubject, deviceID, RegisterChan)

	if m.nc != nil {
		sub, err = m.nc.Subscribe(subject, func(msg *nats.Msg) {
			// Return session ID to channel
			sessionC <- string(msg.Data)
			m.nc.Publish(msg.Reply, []byte("OK"))
		})
		return nil, err
	}
	return sub, nil
}

// PublishSSHMessageToClient sends a message device -> client
func (m *NATSRepository) PublishSSHMessageToClient(deviceID string, sessionID string, b []byte) error {
	return m.publish(deviceID, sessionID, ClientChan, b)
}

// PublishSSHMessageToDevice sends a message client -> device
func (m *NATSRepository) PublishSSHMessageToDevice(deviceID string, sessionID string, b []byte) error {
	return m.publish(deviceID, sessionID, DeviceChan, b)
}

func (m *NATSRepository) publish(deviceID string, sessionID string, channel string, b []byte) error {
	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.messages.2a598bf6-7591-45f1-979f-1fd07f97704c.device
	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.messages.2a598bf6-7591-45f1-979f-1fd07f97704c.client
	subject := fmt.Sprintf("%s.%s.%s.%s.%s", SSHSubject, deviceID, "messages", sessionID, channel)

	if m.nc != nil {
		if err := m.nc.Publish(subject, b); err != nil {
			return err
		}
	}

	return nil
}

// SubscribeToSSHMessagesForDevice subscribes for message client -> device
func (m *NATSRepository) SubscribeToSSHMessagesForDevice(deviceID string, sshC chan []byte) (*nats.Subscription, error) {
	return m.subscribe(deviceID, "*", DeviceChan, sshC)
}

// SubscribeToSSHMessagesForClient subscribes for message device -> client
func (m *NATSRepository) SubscribeToSSHMessagesForClient(deviceID string, sessionID string, sshC chan []byte) (*nats.Subscription, error) {
	return m.subscribe(deviceID, sessionID, ClientChan, sshC)
}

func (m *NATSRepository) subscribe(deviceID string, sessionID string, channel string, sshC chan []byte) (*nats.Subscription, error) {
	var sub *nats.Subscription
	var err error

	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.messages.2a598bf6-7591-45f1-979f-1fd07f97704c.device
	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.messages.2a598bf6-7591-45f1-979f-1fd07f97704c.client
	subject := fmt.Sprintf("%s.%s.%s.%s.%s", SSHSubject, deviceID, "messages", sessionID, channel)

	if m.nc != nil {
		sub, err = m.nc.Subscribe(subject, func(msg *nats.Msg) {
			sshC <- msg.Data
		})
		return nil, err
	}
	return sub, nil
}

// ListenForClientPing listens for a ping by a client
func (m *NATSRepository) ListenForClientPing(deviceID string) (*nats.Subscription, error) {
	var sub *nats.Subscription
	var err error

	// Use * to catch all session ID's
	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.status.*
	subject := fmt.Sprintf("%s.%s.%s.*", SSHSubject, deviceID, StatusChan)

	if m.nc != nil {
		sub, err = m.nc.Subscribe(subject, func(msg *nats.Msg) {
			m.nc.Publish(msg.Reply, []byte("Pong"))
		})
		return nil, err
	}
	return sub, nil
}

// SendPingToDevice sends a ping to a device
func (m *NATSRepository) SendPingToDevice(deviceID string, sessionID string, timeout time.Duration) error {
	var input, resp string
	input = "Ping"

	// eg: ssh.f61f78fb-c52e-4df9-8b1a-2219b2188bd4.status.2a598bf6-7591-45f1-979f-1fd07f97704c
	subject := fmt.Sprintf("%s.%s.%s.%s", SSHSubject, deviceID, StatusChan, sessionID)

	if m.nc != nil {
		err := m.nc.Request(subject, input, &resp, timeout)
		if err != nil {
			return err
		}
	}
	if resp == "Pong" {
		return nil
	}

	return fmt.Errorf("Incorrect answer")
}
