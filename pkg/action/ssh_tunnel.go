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

	"github.com/gorilla/websocket"
	"github.com/grid-x/wssh/pkg/session"
	"github.com/grid-x/wssh/pkg/stream/std"
	"github.com/grid-x/wssh/pkg/stream/ws"

	"github.com/grid-x/gxctl/pkg/service"
)

var (
	// ServerSSL is true iff we should use SSL to connect to server
	ServerSSL = true
)

func SSHTunnel(s *service.Service, sn string) error {
	device, err, profile, _ := getDeviceBySN(s.Client, sn)
	if err != nil {
		return err
	}

	token, err := s.Client.GetTokenFromAuthConfig(profile)
	if err != nil {
		return err
	}

	c := &http.Client{}

	serverAddr := "ssh.staging.ds.gridx.ai:443"
	if isStaging, _ := s.Client.IsStaging(profile); !*isStaging {
		serverAddr = "ssh.ds.gridx.ai:443"
	}

	_, tID, err := NewSession(c, serverAddr, token, device.Metadata.ID)
	if err != nil {
		return err
	}

	u := url.URL{Scheme: "wss", Host: serverAddr, Path: fmt.Sprintf("/agent/tunnel/%v", tID)}

	headers := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return err
	}

	wsStream := ws.New(conn)
	defer conn.Close()
	stdStream := std.New()

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

	return nil

}

// NewSession requests a new session
func NewSession(c *http.Client, serverAddr string, token string, deviceID string) (int, int, error) {
	var body session.HTTPresponse
	err := Post(c, serverAddr, token, "/agent/session", fmt.Sprintf(`{"device": "%v"}`, deviceID), &body)

	if err != nil {
		return -1, -1, err
	}

	if body.ErrorMsg != "" {
		return -1, -1, errors.New(body.ErrorMsg)
	}

	return body.SessionID, body.TunnelID, nil
}

// Post HTTP request
func Post(c *http.Client, serverAddr string, token string, url string, payload string, v interface{}) error {
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

	return json.Unmarshal(resBody, &v)
}
