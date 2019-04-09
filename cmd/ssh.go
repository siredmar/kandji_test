package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/pkg/term"
	"github.com/spf13/cobra"
	"github.com/tj/go-spin"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type SSH struct {
	Command *cobra.Command
}

func NewSSH(parent *cobra.Command) *SSH {
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
			initCommand, _ := cmd.Flags().GetString("command")

			info, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("Unknown error")
			}
			if (info.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
				// Piped input
				return fmt.Errorf("Piped input is currently not supported")
			}

			client := client.NewAPIClient()

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
				for {
					b, err := getChar(t)
					if err != nil {
						fmt.Println("Unknown error: ", err)
						break
					}
					input <- b
				}
			}()

			err = createSession(client, deviceID, "Setting up ssh infrastructure...", initCommand, input, output)
			if err != nil {
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

func createSession(client *client.APIClient, deviceID, waitText, initCommand string, inputChannel, outputChannel chan []byte) error {
	connectedChannel := make(chan int)
	defer close(connectedChannel)

	s := spin.New()
	s.Set(spin.Box1)
	go func() {
		for {
			select {
			case <-connectedChannel:
				fmt.Print("\r\n")
				fmt.Print("Connection established!\r\n")
				return
			default:
				fmt.Printf("\r\033[36m%s\033[m %s", waitText, s.Next())
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	endpoint := fmt.Sprintf("%s/%s/ssh", api.DevicesEndpoint, deviceID)

	additionalHeaders := make(map[string]string)
	if initCommand == "" {
		initCommand = "/dbclient -y root@127.0.0.1"
	}
	additionalHeaders["command"] = initCommand

	conn, err := client.GetWebsocketConnection(endpoint, additionalHeaders)
	if err != nil {
		return err
	}
	defer conn.Close()

	websocketWriter := NewWebsocketWriter(conn)

	var processId string

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
			err = json.Unmarshal(message, &messageType)
			if err != nil {
				fmt.Println("Unknown message received. Closing connection...")
			}

			switch messageType.Type {
			case ssh.ProcessOutputMessageType:
				var output ssh.ProcessOutputMessage
				err = json.Unmarshal(message, &output)
				if err != nil {
					fmt.Println("Not able to unmarshall process output. Closing connection...")
				}
				outputChannel <- output.Data
			case ssh.ProcessCreatedMessageType:
				connectedChannel <- 0

				var created ssh.ProcessCreatedMessage
				err = json.Unmarshal(message, &created)
				if err != nil {
					fmt.Println("Not able to unmarshall process created. Closing connection...")
				}
				processId = created.ID
			case ssh.ProcessTerminatedMessageType:
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
			b := <-inputChannel
			m := ssh.NewExecuteCommandMessage(processId, b)
			err = websocketWriter.WriteJSON(m)
			if err != nil {
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

func getChar(t *term.Term) ([]byte, error) {
	bytes := make([]byte, 3)

	var n int
	n, err := t.Read(bytes)
	if err != nil {
		return nil, err
	}

	if n == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".
		if bytes[2] == 65 {
			// Up
			bytes = []byte{27, 91, 65}
		} else if bytes[2] == 66 {
			// Down
			bytes = []byte{27, 91, 66}
		} else if bytes[2] == 67 {
			// Right
			bytes = []byte{27, 91, 67}
		} else if bytes[2] == 68 {
			// Left
			bytes = []byte{27, 91, 68}
		}

		return bytes, nil
	}

	return bytes[:1], nil
}
