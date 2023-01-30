package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	corev1beta1 "github.com/grid-x/ds-k8s-v2/apis/core/v1beta1"
	"github.com/sirupsen/logrus"

	"github.com/grid-x/wssh/internal"
	"github.com/grid-x/wssh/internal/devicepod"
	"github.com/grid-x/wssh/pkg/tunnel"
)

// Session is a two-way tunnel between two clients
type Session struct {
	ID             uuid.UUID
	AgentID        string
	AgentTunnelID  uuid.UUID
	DeviceID       string
	DeviceTunnelID uuid.UUID
	Signals        chan Signal
	SSHFlavor      SSHFlavor
	state          State
	log            logrus.FieldLogger
	manager        *Manager
	mutex          sync.Mutex
}

// function based FSM
// see: Rob Pike "Lexical Scanning in Go"
type stateFn func(context.Context, *Session) (context.Context, stateFn)

var tunnelCloseTimeout = 3 * time.Second

// IDfromString parses a string into a Session ID
func IDfromString(val string) (uuid.UUID, error) {
	return uuid.Parse(val)
}

// State returns this sessions state
func (s *Session) State() State {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.state
}

func (s *Session) setState(ctx context.Context, st State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state != st {
		if v := ctx.Value(internal.SessionContextKey); v != nil {
			sessionCtx := v.(internal.SessionContext)

			// decrease old state counter
			s.manager.Prom.SessionsStateDec(sessionCtx, s.state)
			// increase new unless its the final state -
			// no reason to count these as they disappear immeadiatly
			if st != StateFin {
				s.manager.Prom.SessionsStateInc(sessionCtx, st)
			}
		}
		s.state = st
		s.log.WithField("state", st).Debug("change session state")
	}
}

// Run concurrently until session finished
func (s *Session) Run(c context.Context) {
	var sessionCtx internal.SessionContext
	ctx := c
	if v := ctx.Value(internal.SessionContextKey); v != nil {
		sessionCtx = v.(internal.SessionContext)
	}
	s.manager.Prom.SessionsActiveInc(sessionCtx)
	s.manager.Prom.SessionsStateInc(sessionCtx, StateInit)
	begin := time.Now()
	for state := stateInit; state != nil; {
		ctx, state = state(ctx, s)
	}
	s.cleanup(ctx)
	s.setState(ctx, StateFin)

	duration := float64(time.Since(begin).Seconds())
	s.manager.Prom.SessionsDurationSecondsObserve(sessionCtx)(duration)
	s.manager.Prom.SessionsActiveDec(sessionCtx)
	s.log.WithField("duration", duration).Info("finish session")
}

// try to establish tunnel between Agent and Device
func stateInit(ctx context.Context, s *Session) (context.Context, stateFn) {
	s.setState(ctx, StateInit)

	if !s.manager.DeviceTunnelReady(s.ID) {
		return ctx, stateWaitDevice
	}

	return ctx, stateReady
}

// establish device tunnel
// send Signal to communicate with session request
func stateWaitDevice(ctx context.Context, s *Session) (context.Context, stateFn) {
	s.setState(ctx, StateWaitDevice)
	begin := time.Now()

	done := ctx.Done()

	t, err := s.manager.DeviceTunnel(s.ID)
	if err != nil || t == nil {
		s.log.WithError(err).Error("could not get device tunnel")
		s.Signals <- SignalFailure
		return ctx, nil
	}

	cfg := s.manager.Config()
	var sessionCtx internal.SessionContext
	if cfg.DebugDeviceSpawn {
		s.spawnDebugDevice(cfg.ListenAddr, cfg.DebugDeviceAddr)
	} else {
		s.log.Info("spawn device")

		if v := ctx.Value(internal.SessionContextKey); v != nil {
			sessionCtx = v.(internal.SessionContext)
			s.log.WithField("sessionCtx", fmt.Sprintf("%+v", sessionCtx)).Debug("get sessionCtx")
		} else {
			s.log.Error("get sessionCtx")
			return ctx, nil
		}
		cfg := s.manager.Config()
		deviceTunnel, err := s.manager.DeviceTunnel(s.ID)
		if err != nil || deviceTunnel == nil {
			s.log.WithError(err).Error("could not get device tunnel")
			return ctx, nil
		}

		keysPath := s.SSHFlavor.KeysPath()
		devicePod, err := devicepod.Create(s.log, s.manager.pods, sessionCtx, deviceTunnel, cfg.DeviceImage, cfg.ExternalAddr, cfg.DSAddr, cfg.LogLevel, keysPath, cfg.CompressLogs)
		if err != nil {
			s.log.WithError(err).Error("could not create pod")
			return ctx, nil
		}

		ctx = context.WithValue(ctx, internal.DevicePodKey, devicePod)
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			s.log.WithField("devicePod", fmt.Sprintf("%+v", devicePod)).Debug("new device pod")
		} else {
			s.log.WithField("devicePod", devicePod.Name).Info("new device pod")
		}
	}

	s.log.Debug("check device for tunnel")
	select {
	case <-done:
		s.log.Debug("canceled")
		return ctx, nil
	case sig := <-t.SigEmit:
		s.log.WithField("sig", sig).Debug("recv sig")
		if t.State() == tunnel.StateReady {
			s.log.Debug("success")
			s.manager.Prom.DeviceConnectDurationSecondsObserve(sessionCtx)(float64(time.Since(begin).Seconds()))
			s.Signals <- SignalSuccess
			return ctx, stateReady
		}
		return ctx, nil
	}
}

// pipe Device and Agent tunnels
func stateReady(ctx context.Context, s *Session) (context.Context, stateFn) {
	s.setState(ctx, StateReady)

	tunnelAgent, err := s.manager.AgentTunnel(s.ID)
	if err != nil {
		s.log.WithError(err).Error("no agent tunnel")
		return ctx, nil
	}
	tunnelDevice, err := s.manager.DeviceTunnel(s.ID)
	if err != nil {
		s.log.WithError(err).Error("no device tunnel")
		return ctx, nil
	}

	ctxCancel, cancel := context.WithCancel(internal.NewValueOnlyContext(ctx))
	defer cancel()
	done := ctxCancel.Done()

	var sessionCtx internal.SessionContext
	if v := ctx.Value(internal.SessionContextKey); v != nil {
		sessionCtx = v.(internal.SessionContext)
	}

	// copy
	go func() {
		log := s.log.WithField("routine", "copy")
		for {
			select {
			case <-done:
				log.Trace("done")
				return
			case b := <-tunnelAgent.ReadBuffer:
				log.WithField("tunnel", "agent").WithField("b", string(b)).Trace("read")
				s.manager.Prom.SessionsTransmittedBytesAdd(sessionCtx)(len(b))
				tunnelDevice.WriteBuffer <- b
			case b := <-tunnelDevice.ReadBuffer:
				log.WithField("tunnel", "device").WithField("b", string(b)).Trace("read")
				s.manager.Prom.SessionsTransmittedBytesAdd(sessionCtx)(len(b))
				tunnelAgent.WriteBuffer <- b
			}
		}
	}()

	log := s.log.WithField("routine", "signals")

signals:
	for {
		select {
		case <-done:
			log.Trace("done")
			break signals
		case sig := <-tunnelAgent.SigEmit:
			log.WithField("tunnel", "agent").WithField("sig", sig).Debug("emit")
			if sig == tunnel.SignalClose {
				break signals
			}
		case sig := <-tunnelDevice.SigEmit:
			log.WithField("tunnel", "device").WithField("sig", sig).Debug("emit")
			if sig == tunnel.SignalClose {
				break signals
			}
		}
	}

	return ctx, stateClosing
}

// close tunnels
func stateClosing(ctx context.Context, s *Session) (context.Context, stateFn) {
	s.setState(ctx, StateClosing)

	tAgent, err := s.manager.AgentTunnel(s.ID)
	if err != nil {
		s.log.WithError(err).Error("no agent tunnel")
		return ctx, nil
	}
	tDevice, err := s.manager.DeviceTunnel(s.ID)
	if err != nil {
		s.log.WithError(err).Error("no device tunnel")
		return ctx, nil
	}

	if tAgent.State() != tunnel.StateFin {
		return ctx, stateCloseAgent
	}

	if tDevice.State() != tunnel.StateFin {
		return ctx, stateCloseDevice
	}

	return ctx, nil
}

// close agent tunnel - must be in StateFin before returning to StateClosing
func stateCloseAgent(ctx context.Context, s *Session) (context.Context, stateFn) {
	s.setState(ctx, StateCloseAgent)

	t, err := s.manager.AgentTunnel(s.ID)
	if err != nil {
		s.log.WithError(err).Error("no agent tunnel")
		return ctx, nil
	}

	if err := s.closeMaybe(ctx, t); err != nil {
		s.log.WithError(err).Error("close agent tunnel")
		return ctx, nil
	}

	return ctx, stateClosing
}

// close device tunnel - must be in StateFin before returning to StateClosing
func stateCloseDevice(ctx context.Context, s *Session) (context.Context, stateFn) {
	s.setState(ctx, StateCloseDevice)

	t, err := s.manager.DeviceTunnel(s.ID)
	if err != nil {
		s.log.WithError(err).Error("no device tunnel")
		return ctx, nil
	}

	if err := s.closeMaybe(ctx, t); err != nil {
		s.log.WithError(err).Error("close device tunnel")
		return ctx, nil
	}

	return ctx, stateClosing
}

func (s *Session) closeMaybe(ctx context.Context, t *tunnel.Tunnel) error {
	log := s.log.WithField("tID", t.ID).WithField("routine", "closeMaybe")

	ctx, cancel := context.WithTimeout(context.Background(), tunnelCloseTimeout)
	defer cancel()

	done := ctx.Done()

	var err error

signals:
	for {
		select {
		case <-done:
			errMsg := "tunnel close timeout"
			err = errors.New(errMsg)
			log.Warn(errMsg)
			break signals
		case sig := <-t.SigEmit:
			log.WithField("sig", sig).Debug("recv sig")
			if t.State() == tunnel.StateFin {
				break signals
			}
		case t.SigSink <- tunnel.SignalClose:
			log.WithField("sig", tunnel.SignalClose).Debug("send sig")
		}
	}

	return err
}

func (s *Session) cleanup(ctx context.Context) {
	if v := ctx.Value(internal.DevicePodKey); v != nil {
		devicePod := v.(*corev1beta1.DevicePod)
		s.log.WithField("devicePod", devicePod.Name).Debug("delete devicePod")
		if err := s.manager.pods.Delete(context.Background(), devicePod); err != nil {
			s.log.WithField("error", err).Error("delete devicePod")
		}
	}
}
