package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/pkg/term"
	"github.com/spf13/cobra"

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

			fmt.Println("Setting up ssh infrastructure...")
			err = createSession(client, deviceID, initCommand, true)
			if err != nil {
				return err
			}

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

func createSession(client *client.APIClient, deviceID, initCommand string, handleInput bool) error {
	endpoint := fmt.Sprintf("%s/%s/ssh", api.DevicesEndpoint, deviceID)

	additionalHeaders := make(map[string]string)
	additionalHeaders["command"] = initCommand

	conn, err := client.GetWebsocketConnection(endpoint, additionalHeaders)
	defer conn.Close()

	websocketWriter := NewWebsocketWriter(conn)

	if err != nil {
		return err
	}

	var processId string

	// Common controls
	var crtlC = []byte("\x03")

	// Exit channel
	exitChannel := make(chan int)
	defer close(exitChannel)

	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := websocketWriter.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				break
			}
		}
	}()

	go func() {
		// Read messages from the device
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
				os.Stdout.Write(output.Data)
			case ssh.ProcessCreatedMessageType:
				var created ssh.ProcessCreatedMessage
				err = json.Unmarshal(message, &created)
				if err != nil {
					fmt.Println("Not able to unmarshall process created. Closing connection...")
				}
				processId = created.ID
			case ssh.ProcessTerminatedMessageType:
			default:
				fmt.Println("Received an unknown message type")
			}
		}
	}()

	// Support ctrl+c
	interruptChannel := make(chan os.Signal, 1)
	defer close(interruptChannel)

	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		for {
			<-interruptChannel
			m := ssh.NewExecuteCommandMessage(processId, crtlC)
			websocketWriter.WriteJSON(m)
		}
	}()

	go func() {
		if !handleInput {
			return
		}

		// Handler for user input
		info, _ := os.Stdin.Stat()
		if (info.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
			// Piped input - Return as other handler will read it
			return
		} else {
			// User input
			for {
				b, _ := getChar()

				m := ssh.NewExecuteCommandMessage(processId, b)
				err = websocketWriter.WriteJSON(m)
				if err != nil {
					fmt.Println("Unknown error: ", err)
					exitChannel <- 0
				}
			}
		}
	}()

	/*
		go func() {
			// Handler for piped in input
			info, _ := os.Stdin.Stat()
			if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
				// User input - Return as other handler will read it
				return
			} else {
				// Piped input
					for {
						// Wait until process is created
						if processId != "" {
							break
						}
					}
					for {
						reader := bufio.NewReader(os.Stdin)
						input, err := reader.ReadBytes('\n')
						if err != nil && err == io.EOF {
							// EOF reached, logout and exit
							msg := append([]byte("exit"), lineFeed...)
							m := ssh.NewExecuteCommandMessage(processId, msg)
							err := websocketWriter.WriteJSON(m)
							if err != nil {
								fmt.Println("Unknown error: ", err)
								exitChannel <- 0
							}
							break
						}

						m := ssh.NewExecuteCommandMessage(processId, input)
						err = websocketWriter.WriteJSON(m)
						if err != nil {
							fmt.Println("Unknown error: ", err)
							exitChannel <- 0
						}
					}

			}
		}()
	*/

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

func getChar() ([]byte, error) {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)

	var numRead int
	numRead, err := t.Read(bytes)
	if err != nil {
		return nil, err
	}

	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
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
	t.Restore()
	t.Close()

	return bytes[:1], nil
}
