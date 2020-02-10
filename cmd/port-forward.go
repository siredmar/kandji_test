package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type PortForward struct {
	Command *cobra.Command
}

func NewPortForward(parent *cobra.Command, client *client.APIClient) *PortForward {
	var portForwardCmd = &cobra.Command{
		Use:                   "port-forward ID --localport=PORT --target=TARGET [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "forwards remote connections to local port",
		Example:               "# gxctl port-forward c72 --localport=8080 --target=127.0.0.1:8080",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"required argument ID not found",
					[]string{"run 'gxctl port-forward --help' for usage"},
				)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			portForwardCmdLocalPort, _ := cmd.Flags().GetString("localport")
			portForwardCmdTarget, _ := cmd.Flags().GetString("target")

			if portForwardCmdLocalPort == "" {
				return errors.E(
					errors.Invalid,
					"required parameter '--localport' not found",
					[]string{"run 'gxctl port-forward --help' for usage"},
				)
			}
			if portForwardCmdTarget == "" {
				return errors.E(
					errors.Invalid,
					"required parameter '--target' not found",
					[]string{"run 'gxctl port-forward --help' for usage"},
				)
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
						msg := ssh.NewExecuteCommandMessage(string(sessionID), []byte("exit\r\n"))
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
				Client:         client,
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
		},
	}

	portForwardCmd.Flags().StringP("localport", "l", "", "Local port for the forwarded connection")
	portForwardCmd.Flags().StringP("target", "t", "", "Target to forward traffic from")

	portForwardCmd.SetHelpTemplate(template.HelpTemplate())
	portForwardCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(portForwardCmd)

	return &PortForward{
		Command: portForwardCmd,
	}
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

			msg := ssh.NewExecuteCommandMessage(sessionID, content)
			m, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Unknown error: ", err)
				break
			}

			socketInputChannel <- m
		}
	}
}
