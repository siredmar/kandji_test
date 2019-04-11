package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type Copy struct {
	Command *cobra.Command
}

func NewCopy(parent *cobra.Command) *Copy {
	var copyCmd = &cobra.Command{
		Use:                   "copy ID SOURCE DESTINATION [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "forwards files to devices",
		Example:               "# gxctl copy /tmp/test.conf c72:/opt/test.conf",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.MissingParameter("ID", "gxctl copy -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			copyCmdSource := args[1]
			copyCmdDestination := args[2]

			if strings.Contains(copyCmdSource, ":") && strings.Contains(copyCmdDestination, ":") {
				return errors.ServerError("You cannot specify the device in both source and destination parameters")
			}

			var id, source, dest string

			if strings.Contains(copyCmdSource, ":") {
				// Direction: Device -> Client
				return errors.NotImplementedError("Copy from device to the client is not yet supported")
			}
			source = copyCmdSource

			if strings.Contains(copyCmdSource, ":") {
				// Direction: Client -> Device
				splitted := strings.Split(copyCmdSource, ":")
				id = splitted[0]
				dest = splitted[1]

				return errors.NotImplementedError("Copy from device to the client is not yet supported")
			}

			client := client.NewAPIClient()

			//Lookup all existing devices to validate ids and autocomplete them if necessary
			devices, err := getDevices(client)
			if err != nil {
				return err
			}
			deviceIDs := devices.GetIds()

			deviceID, err := api.LookupID(id, deviceIDs)
			if err != nil {
				return err
			}

			srcFile, err := os.Open(source)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			socketInputChannel := make(chan []byte)
			socketOutputChannel := make(chan []byte)
			sessionChannel := make(chan []byte)

			// Ignore output messages but make sure channel is not blocked
			go func() {
				for {
					<-socketOutputChannel
				}
			}()

			tmpFileName := uuid.New().String()

			go func() {
				sessionID := string(<-sessionChannel)

				// Issue inital create file message
				msg := ssh.NewCreateFileMessage(sessionID, tmpFileName, false)
				m, err := json.Marshal(msg)
				if err != nil {
					fmt.Println("Unknown error: ", err)
					return
				}
				socketInputChannel <- m

				buf := make([]byte, 32*1024)

				for {
					n, err := srcFile.Read(buf)
					if err != nil {
						msg := ssh.NewWriteToFileMessage(sessionID, tmpFileName, nil, true)
						m, err := json.Marshal(msg)
						if err != nil {
							fmt.Println("Unknown error: ", err)
							break
						}
						socketInputChannel <- m
						break
					}
					if n > 0 {
						msg := ssh.NewWriteToFileMessage(sessionID, tmpFileName, buf[0:n], false)
						m, err := json.Marshal(msg)
						if err != nil {
							fmt.Println("Unknown error: ", err)
							break
						}
						socketInputChannel <- m
					}
				}
			}()

			conf := &SSHConfig{
				Client:         client,
				DeviceID:       deviceID,
				WaitText:       "Connecting to the device...",
				InitCommand:    "",
				InputChannel:   socketInputChannel,
				OutputChannel:  socketOutputChannel,
				SessionChannel: sessionChannel,
				Silent:         false,
			}

			if err := createSession(conf); err != nil {
				return err
			}

			// At this point we've copied the file into the container... Time for some SCP magic to happen
			conf = &SSHConfig{
				Client:         client,
				DeviceID:       deviceID,
				WaitText:       "Connecting to the device...",
				InitCommand:    fmt.Sprintf("/scp -S /dbclient /%s root@127.0.0.1:%s", tmpFileName, dest),
				InputChannel:   socketInputChannel,
				OutputChannel:  socketOutputChannel,
				SessionChannel: sessionChannel,
				Silent:         true,
			}

			go func() {
				// Unblock sessionChannel for SCP Session
				<-sessionChannel
			}()

			if err := createSession(conf); err != nil {
				return err
			}

			fmt.Println("All done!")
			return nil
		},
	}

	copyCmd.SetHelpTemplate(template.HelpTemplate())
	copyCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(copyCmd)

	return &Copy{
		Command: copyCmd,
	}
}
