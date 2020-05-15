package action

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func PortForward(s *service.Service, id string, portForwardCmdLocalPort string, portForwardCmdTarget string) error {
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

	local, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "127.0.0.1", portForwardCmdLocalPort))
	if err != nil {
		return err
	}
	defer local.Close()

	socketInputChannel := make(chan []byte)
	socketOutputChannel := make(chan []byte)
	sessionChannel := make(chan []byte)
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	connectionCache := make(map[string]net.Conn)

	// asynchronously accepting connections on the listen port
	go func() {
		sessionID := <-sessionChannel
		go func() {
			for {
				conn, err := local.Accept()
				if err != nil {
					return
				}

				u := uuid.New().String()
				connectionCache[u] = conn

				go readFromLocalConnection(conn, u, socketInputChannel, string(sessionID))
			}
		}()

		// Support ctrl+c to exit the forwarding
		go func() {
			for {
				<-interruptChannel
				fmt.Println("\nClosing session...")
				msg := api.NewExecuteCommandMessage(string(sessionID), []byte("exit\r\n"))
				m, err := json.Marshal(msg)
				if err != nil {
					fmt.Println("Unknown error: ", err)
					break
				}
				socketInputChannel <- m
				time.Sleep(1 * time.Second)
				// Make sure exit message is forwarded to the device.
				// The session will then be exited in background as dbclient might wait for some tcp connections to timeout
				// and shutdown gracefully afterwards. This is however something we do not want to wait for
				os.Exit(0)
			}
		}()
	}()

	// Read websocket outputs from the device and forward them to the belonging connection
	go func() {
		for {
			b := <-socketOutputChannel
			// Port forward messages are prefixed with a 36 char uuid v4 to allow multiple concurrent connections
			uuid := b[0:36]
			content := b[36:]
			c := connectionCache[string(uuid)]
			c.Write(content)
		}
	}()

	// Setup SSH connection
	// LOCALPORT will be replaced with a free port on the device by the ssh-agent
	conf := &SSHConfig{
		Client:         s.Client,
		DeviceID:       deviceID,
		WaitText:       "Connecting to the device...",
		ReadyText:      "Serving on local port " + portForwardCmdLocalPort,
		InitCommand:    fmt.Sprintf("/dbclient -p 22222 -y -L LOCALPORT:%s root@127.0.0.1", portForwardCmdTarget),
		InputChannel:   socketInputChannel,
		OutputChannel:  socketOutputChannel,
		SessionChannel: sessionChannel,
		Silent:         false,
	}

	if err := createSession(conf); err != nil {
		return err
	}

	return nil
}

func readFromLocalConnection(local net.Conn, uuid string, socketInputChannel chan []byte, sessionID string) {
	buf := make([]byte, 32*1024)
	for {
		nr, err := local.Read(buf)
		if err != nil {
			break
		}
		if nr > 0 {
			// Port forward messages are prefixed with a 36 char uuid v4 to allow multiple concurrent connections
			content := append([]byte(uuid), buf[0:nr]...)

			msg := api.NewExecuteCommandMessage(sessionID, content)
			m, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Unknown error: ", err)
				break
			}

			socketInputChannel <- m
		}
	}
}
