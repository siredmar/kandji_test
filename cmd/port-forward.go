package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type PortForward struct {
	Command *cobra.Command
}

func NewPortForward(parent *cobra.Command) *PortForward {
	var portForwardCmd = &cobra.Command{
		Use:                   "port-forward ID --localport=PORT --target=TARGET [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "forwards remote connections to local port",
		Example:               "# gxctl port-forward c72 --localport=8080 --target=127.0.0.1:8080",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl port-forward -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			portForwardCmdLocalPort, _ := cmd.Flags().GetString("localport")
			portForwardCmdTarget, _ := cmd.Flags().GetString("target")

			if portForwardCmdLocalPort == "" {
				return errors.MissingParameter("localport", "gxctl port-forward -h")
			}
			if portForwardCmdTarget == "" {
				return errors.MissingParameter("target", "gxctl port-forward -h")
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

			local, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "127.0.0.1", portForwardCmdLocalPort))
			if err != nil {
				return err
			}
			defer local.Close()

			socketInputChannel := make(chan []byte)
			socketOutputChannel := make(chan []byte)

			connectionCache := make(map[string]net.Conn)

			go func() {
				for {
					conn, err := local.Accept()
					if err != nil {
						return
					}

					u := uuid.New().String()
					connectionCache[u] = conn

					go handleConnection(conn, u, socketInputChannel, socketOutputChannel)
				}
			}()

			// Support ctrl+c to exit

			interruptChannel := make(chan os.Signal, 1)
			defer close(interruptChannel)

			signal.Notify(interruptChannel, os.Interrupt)
			go func() {
				for {
					<-interruptChannel
					socketInputChannel <- []byte("exit\r\n")
					time.Sleep(1 * time.Second)
					// Make sure exit is forwarded to the device. The session will then be exited in background
					os.Exit(0)
				}
			}()

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

			// LOCALPORT will be replaced with a free port on the device by the ssh-agent
			sshCommand := fmt.Sprintf("/dbclient -y -L LOCALPORT:%s root@127.0.0.1", portForwardCmdTarget)
			err = createSession(client, deviceID, "Connecting to the device...", sshCommand, socketInputChannel, socketOutputChannel)
			if err != nil {
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

func handleConnection(local net.Conn, uuid string, socketInputChannel, socketOutputChannel chan []byte) {
	go readFromLocalConnection(local, uuid, socketInputChannel)
}

func readFromLocalConnection(local net.Conn, uuid string, socketInputChannel chan []byte) {
	buf := make([]byte, 32*1024)
	for {
		nr, err := local.Read(buf)
		if err != nil {
			break
		}
		if nr > 0 {
			// Port forward messages are prefixed with a 36 char uuid v4 to allow multiple concurrent connections
			msg := append([]byte(uuid), buf[0:nr]...)
			socketInputChannel <- msg
		}
	}
}
