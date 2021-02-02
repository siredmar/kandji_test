package action

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/grid-x/wssh/pkg/session"
	"github.com/grid-x/wssh/pkg/stream/std"
	"github.com/grid-x/wssh/pkg/stream/ws"

	errs "github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
	"github.com/grid-x/gxctl/pkg/spinner"
)

var (
	// ServerSSL is true iff we should use SSL to connect to server
	ServerSSL = true
)

func SSHTunnel(s *service.Service, quiet bool, sn string) error {
	if quiet {
		spinner.Disable()
	}
	checkAgent := spinner.New("check ssh agent")
	if !sshAgentAvailable() {
		checkAgent.Fail()
		return errs.E(
			errs.Service,
			"ssh-agent unavailable. Please ensure it is running",
		)
	}
	checkAgent.Ok()

	getDevice := spinner.New("device")
	device, err, profile, _ := getDeviceBySN(s.Client, sn)
	if err != nil {
		getDevice.Fail()
		return err
	}
	if v, ok := device.Metadata.Labels["core.gridx.ai/hardware"]; ok && v == "virtual" {
		getDevice.Fail()
		err := errs.E(
			errs.Invalid,
			fmt.Sprintf("can't connect to virtual device %v", device.Metadata.ID),
		)
		return err
	}
	if !device.IsOnline() {
		getDevice.Fail()
		err := errs.E(
			errs.Invalid,
			fmt.Sprintf("device %v is offline", device.Metadata.ID),
		)
		return err
	}
	getDevice.Ok()

	auth := spinner.New("auth")
	token, err := s.Client.GetTokenFromAuthConfig(profile)
	if err != nil {
		auth.Fail()
		return err
	}
	auth.Ok()

	c := &http.Client{}

	serverAddr := "ssh.staging.ds.gridx.ai:443"
	if isStaging, _ := s.Client.IsStaging(profile); !*isStaging {
		serverAddr = "ssh.ds.gridx.ai:443"
	}

	startSession := spinner.New("start session")
	_, tID, err := NewSession(c, serverAddr, token, device.Metadata.ID)
	if err != nil {
		startSession.Fail()
		return err
	}
	startSession.Ok()

	u := url.URL{Scheme: "wss", Host: serverAddr, Path: fmt.Sprintf("/agent/tunnel/%v", tID)}

	headers := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
	}

	connectTunnel := spinner.New("connect to tunnel")
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		connectTunnel.Fail()
		return err
	}

	wsStream := ws.New(conn)
	defer conn.Close()
	stdStream := std.New()
	connectTunnel.Ok()

	go func() {
		for {
			if _, err := io.Copy(wsStream, stdStream); err != nil {
				if err == io.EOF {
					continue
				}
				return
			}
		}
	}()

	for {
		if _, err := io.Copy(stdStream, wsStream); err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}
	}
}

// NewSession requests a new session
func NewSession(c *http.Client, serverAddr string, token string, deviceID string) (*uuid.UUID, *uuid.UUID, error) {
	var body session.HTTPresponse
	err := Post(c, serverAddr, token, "/agent/session", fmt.Sprintf(`{"device": "%v"}`, deviceID), &body)

	if err != nil {
		return nil, nil, err
	}

	if body.ErrorMsg != "" {
		return nil, nil, errors.New(body.ErrorMsg)
	}

	return &body.SessionID, &body.TunnelID, nil
}

// Post HTTP request
func Post(c *http.Client, serverAddr string, token string, url string, payload string, v *session.HTTPresponse) error {
	var reqURL string
	if ServerSSL {
		reqURL = "https://" + serverAddr + url
	} else {
		reqURL = "http://" + serverAddr + url
	}
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resBody, &v); err != nil {
		return fmt.Errorf("Unknown error: " + string(resBody))
	}

	return nil
}

func sshAgentAvailable() bool {
	cmd := exec.Command("ssh-add", "-l")
	_, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 2 {
				return false
			}
			return true
		}
	}
	return true
}
