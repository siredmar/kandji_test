package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	ssh "github.com/grid-x/ds-api/pkg/ssh"
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

			err = createSession(client, deviceID)
			if err != nil {
				return err
			}

			return nil
		},
	}

	sshCmd.SetHelpTemplate(template.HelpTemplate())
	sshCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(sshCmd)

	return &SSH{
		Command: sshCmd,
	}
}

func createSession(client *client.APIClient, deviceID string) error {
	fmt.Println("Setting up ssh infrastructure...")
	endpoint := fmt.Sprintf("%s/%s/ssh", api.DevicesEndpoint, deviceID)

	conn, err := client.GetWebsocketConnection(endpoint)
	if err != nil {
		return err
	}

	// Support Proxy
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		<-interruptChannel
		os.Exit(1)
	}()

	var processId string

	go func() {
		defer conn.Close()
		for {
			var messageType ssh.MessageType

			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Connection closed")
			}
			err = json.Unmarshal(message, &messageType)
			if err != nil {
				fmt.Println("Unknown message received. Closing connection...")
				os.Exit(0)
			}

			switch messageType.Type {
			case ssh.ProcessOutputMessageType:
				var output ssh.ProcessOutputMessage
				err = json.Unmarshal(message, &output)
				if err != nil {
					fmt.Println("Not able to unmarshall process output. Closing connection...")
					os.Exit(0)
				}
				os.Stdout.Write(output.Data)
			case ssh.ProcessCreatedMessageType:
				var created ssh.ProcessCreatedMessage
				err = json.Unmarshal(message, &created)
				if err != nil {
					fmt.Println("Not able to unmarshall process created. Closing connection...")
					os.Exit(0)
				}
				processId = created.ID
			case ssh.ProcessTerminatedMessageType:
				os.Exit(0)
			default:
				fmt.Println("Received an unknown message type")
			}
		}
	}()

	for {
		select {
		default:
			var msg = make([]byte, 1024)
			size, err := os.Stdin.Read(msg)
			if err == io.EOF {
				return nil
			} else if err != nil {
				panic(err)
			} else {
				m := ssh.NewExecuteCommandMessage(processId, msg[0:size])
				conn.WriteJSON(m)
			}
		}
	}
}
