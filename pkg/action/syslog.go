package action

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/term"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func Syslog(s *service.Service, id string) error {
	//Lookup all existing devices to validate ids and autocomplete them if necessary
	devices, err := getDevices(s.Client)
	if err != nil {
		return err
	}
	deviceIDs := devices.GetIds()

	deviceID, err := api.LookupID(id, deviceIDs)
	if err != nil {
		return err
	}

	input := make(chan []byte)
	defer close(input)
	output := make(chan []byte)
	defer close(output)
	session := make(chan []byte)
	defer close(session)

	// Forward output from SSH session
	go func() {
		for {
			b := <-output
			os.Stdout.Write(b)
		}
	}()

	t, _ := term.Open("/dev/tty")
	term.RawMode(t)

	// Forward input to SSH session
	go func() {
		sessionID := string(<-session)
		for {
			b, err := getChar(t)
			if err != nil {
				fmt.Println("Unknown error: ", err)
				break
			}
			msg := api.NewExecuteCommandMessage(sessionID, b)
			_, err = json.Marshal(msg)
			input <- b
		}
	}()

	conf := &SSHConfig{
		Client:         s.Client,
		DeviceID:       deviceID,
		WaitText:       "Connecting to the device...",
		InitCommand:    "/dbclient -p 22222 -y root@127.0.0.1 'journalctl -f'",
		InputChannel:   input,
		OutputChannel:  output,
		SessionChannel: session,
		Silent:         false,
	}

	err = createSession(conf)
	if err != nil {
		return err
	}

	t.Restore()

	return nil
}
