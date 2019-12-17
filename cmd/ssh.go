package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/pkg/term"
	"github.com/spf13/cobra"
	"github.com/tj/go-spin"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type SSH struct {
	Command *cobra.Command
}

type SSHConfig struct {
	Client         *client.APIClient
	DeviceID       string
	WaitText       string
	ReadyText      string
	InitCommand    string
	InputChannel   chan []byte
	OutputChannel  chan []byte
	SessionChannel chan []byte
	Silent         bool
}

func NewSSH(parent *cobra.Command, client *client.APIClient) *SSH {
	var sshCmd = &cobra.Command{
		Use:                   "ssh ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "ssh to different devices",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl ssh -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("Unknown error")
			}
			if (info.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
				// Piped input
				return fmt.Errorf("Piped input is currently not supported")
			}

			//Lookup all existing devices to validate ids and autocomplete them if necessary
			devices, err := getDevices(client)
			if err != nil {
				return err
			}
			deviceIDs := devices.GetIds()

			deviceID, err := api.LookupID(args[0], deviceIDs)
			if err != nil {
				return err
			}

			input := make(chan []byte)
			defer close(input)
			output := make(chan []byte)
			defer close(output)
			session := make(chan []byte)
			defer close(session)

			t, _ := term.Open("/dev/tty")
			term.RawMode(t)

			// Forward output from SSH session
			go func() {
				for {
					b := <-output

					os.Stdout.Write(b)
				}
			}()

			// Forward input to SSH session
			go func() {
				sessionID := string(<-session)
				for {
					b, err := getChar(t)
					if err != nil {
						fmt.Println("Unknown error: ", err)
						break
					}
					msg := ssh.NewExecuteCommandMessage(sessionID, b)
					m, err := json.Marshal(msg)
					if err != nil {
						fmt.Println("Unknown error: ", err)
						break
					}

					input <- m
				}
			}()

			initCommand := "/dbclient -p 22222 -y root@127.0.0.1"
			flagCommand, _ := cmd.Flags().GetString("command")
			if flagCommand != "" {
				initCommand += " " + flagCommand
			}
			conf := &SSHConfig{
				Client:         client,
				DeviceID:       deviceID,
				WaitText:       "Setting up ssh infrastructure...",
				InitCommand:    initCommand,
				InputChannel:   input,
				OutputChannel:  output,
				SessionChannel: session,
				Silent:         false,
			}

			if err := createSession(conf); err != nil {
				return err
			}

			t.Restore()

			return nil
		},
	}

	sshCmd.SetHelpTemplate(template.HelpTemplate())
	sshCmd.SetUsageTemplate(template.UsageTemplate())
	sshCmd.Flags().StringP("command", "c", "", "specify a command to issue")
	parent.AddCommand(sshCmd)

	return &SSH{
		Command: sshCmd,
	}
}

func createSession(conf *SSHConfig) error {
	connectedChannel := make(chan int)
	defer close(connectedChannel)

	s := spin.New()
	s.Set(spin.Box1)
	go func() {
		if conf.Silent {
			for {
				// Ignore it ;)
				<-connectedChannel
			}
		}
		for {
			select {
			case <-connectedChannel:
				fmt.Fprint(os.Stderr, "\r\n")
				fmt.Fprint(os.Stderr, "Connection established!\r\n")
				fmt.Fprint(os.Stderr, conf.ReadyText+"\r\n")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r\033[36m%s\033[m %s", conf.WaitText, s.Next())
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	endpoint := fmt.Sprintf("%s/%s/ssh", api.DevicesEndpoint, conf.DeviceID)

	additionalHeaders := make(map[string]string)
	additionalHeaders["command"] = conf.InitCommand

	conn, err := conf.Client.GetWebsocketConnection(endpoint, additionalHeaders)
	if err != nil {
		return err
	}
	defer conn.Close()

	websocketWriter := NewWebsocketWriter(conn)

	// Exit channel
	exitChannel := make(chan int)
	defer close(exitChannel)

	go func() {
		for {
			// Websocket connection will close after 60s of inactivity. Send a keepalive every 30s.
			time.Sleep(30 * time.Second)
			err := websocketWriter.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				break
			}
		}
	}()

	// Read messages from the device and send it to the client
	go func() {
		for {
			var messageType ssh.MessageType

			_, message, err := conn.ReadMessage()

			if err != nil {
				fmt.Println("Connection closed")
				exitChannel <- 0
				break
			}
			if err := json.Unmarshal(message, &messageType); err != nil {
				fmt.Println("Unknown message received. Closing connection...")
			}

			switch messageType.Type {
			case ssh.ProcessOutputMessageType:

				var output ssh.ProcessOutputMessage
				if err := json.Unmarshal(message, &output); err != nil {
					fmt.Println("Not able to unmarshall process output. Closing connection...")
				}
				conf.OutputChannel <- output.Data
			case ssh.WriteToFileMessageType:
				conf.OutputChannel <- message
			case ssh.ProcessCreatedMessageType:
				connectedChannel <- 0

				var created ssh.ProcessCreatedMessage
				if err := json.Unmarshal(message, &created); err != nil {
					fmt.Println("Not able to unmarshall process created. Closing connection...")
				}
				conf.SessionChannel <- []byte(created.ID)
			case ssh.ProcessTerminatedMessageType:
				exitChannel <- 0
				return
			case ssh.ErrorMessageType:
				fmt.Println("Unknown error encountered... Closing")
				exitChannel <- 0
				return
			default:
				fmt.Println("Received an unknown message type")
			}
		}
	}()

	// Handle input and send it to the device
	go func() {
		for {
			b := <-conf.InputChannel
			if err := websocketWriter.WriteMessage(websocket.TextMessage, b); err != nil {
				exitChannel <- 0
				break
			}
		}
	}()

	<-exitChannel
	return nil
}

type WebsocketWriter struct {
	Socket *websocket.Conn // websocket connection of the player
	mu     sync.Mutex
}

func NewWebsocketWriter(c *websocket.Conn) *WebsocketWriter {
	w := &WebsocketWriter{
		Socket: c,
	}
	return w
}

func (w *WebsocketWriter) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Socket.WriteJSON(v)
}

func (w *WebsocketWriter) WriteMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Socket.WriteMessage(messageType, data)
}

func getChar(r io.Reader) ([]byte, error) {
	bytes := make([]byte, 4096)
	n, err := r.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes[:n], nil
}
