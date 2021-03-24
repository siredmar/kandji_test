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
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/grid-x/wssh/pkg/session"
	"github.com/grid-x/wssh/pkg/stream/std"
	"github.com/grid-x/wssh/pkg/stream/ws"

	"github.com/grid-x/gxctl/pkg/api"
	errs "github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
	"github.com/grid-x/gxctl/pkg/spinner"
)

var (
	// ServerSSL is true iff we should use SSL to connect to server
	ServerSSL = true
)

func SSHTunnel(s *service.Service, quiet, skipConfigCheck bool, sn string) error {
	if quiet {
		spinner.Disable()
	}
	checkConfig := spinner.New("check config")
	if err := SSHConfigCheck(s); err != nil {
		if skipConfigCheck {
			checkConfig.Warn()
		} else {
			checkConfig.Fail()
			return errs.E(
				err,
				"edit SSH config so it matches 'gxctl ssh setup' output OR invoke 'gxctl ssh tunnel …' with --skip-config-check flag",
			)
		}
	} else {
		checkConfig.Ok()
	}

	checkAgent := spinner.New("check ssh agent")
	if !sshAgentAvailable() {
		checkAgent.Fail()
		return errs.E(
			errs.Service,
			"      ssh-agent unavailable. Please ensure it is running",
		)
	}
	checkAgent.Ok()

	auth := spinner.New("auth")
	var sb strings.Builder
	ok := false
	for i, p := range s.Client.Auth.Profiles {
		err := p.Auth.Token.Validate()
		if err == nil {
			ok = true
			sb.WriteString(fmt.Sprintf("        %s: \033[32mOK\033[39m", p.Name))
		} else {
			sb.WriteString(fmt.Sprintf("        %s: \033[33m%s\033[39m", p.Name, err))
		}

		if i < len(s.Client.Auth.Profiles)-1 {
			sb.WriteString("\n")
		}
	}
	if ok {
		auth.Ok()
	} else {
		auth.Fail()
	}

	auth.Write(sb.String())

	getDevice := spinner.New("device")
	responseList, err := getDeviceBySN(s.Client, sn)
	if err != nil {
		getDevice.Fail()
		return err
	}

	var device api.Device
	var profile string

	for _, r := range responseList {
		if len(r.device.Devices) > 1 {
			getDevice.Fail()
			err := errs.E(
				errs.Invalid,
				fmt.Sprintf("        more than one device found"),
			)
			return err
		} else if len(r.device.Devices) == 1 {
			if !device.IsEmpty() {
				getDevice.Fail()
				err := errs.E(
					errs.Invalid,
					fmt.Sprintf("        more than one device found"),
				)
				return err
			}
			device = r.device.Devices[0]
			profile = r.profile
		}
	}

	if device.IsEmpty() {
		getDevice.Fail()
		err := errs.E(
			errs.Invalid,
			fmt.Sprintf("        no device found"),
		)
		return err
	}

	if v, ok := device.Metadata.Labels["core.gridx.ai/hardware"]; ok && v == "virtual" {
		getDevice.Fail()
		err := errs.E(
			errs.Invalid,
			fmt.Sprintf("        can't connect to virtual device %v", device.Metadata.ID),
		)
		return err
	}
	if !device.IsOnline() {
		getDevice.Fail()
		err := errs.E(
			errs.Invalid,
			fmt.Sprintf("        device %v is offline", device.Metadata.ID),
		)
		return err
	}
	getDevice.Ok()

	c := &http.Client{}

	serverAddr := "ssh.staging.ds.gridx.ai:443"
	if isStaging := s.Client.IsStaging(profile); !isStaging {
		serverAddr = "ssh.ds.gridx.ai:443"
	}

	tokenInject := spinner.New("inject token")
	token, err := s.Client.GetTokenFromAuthConfig(profile)
	if err != nil {
		tokenInject.Fail()
		return err
	}
	tokenInject.Ok()

	startSession := spinner.New("start session")
	_, tID, err := NewSession(c, serverAddr, token.String(), device.Metadata.ID)
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
