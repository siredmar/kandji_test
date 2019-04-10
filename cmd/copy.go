package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
		Use:                   "copy ID --source=SOURCE --destination=DESTINATION [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "forwards files to devices",
		Example:               "# gxctl copy c72 --src=/tmp/test.conf --target=/opt/test.conf",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl copy -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			copyCmdSource, _ := cmd.Flags().GetString("source")
			copyCmdDestination, _ := cmd.Flags().GetString("destination")

			if copyCmdSource == "" {
				return errors.MissingParameter("source", "gxctl copy -h")
			}
			if copyCmdDestination == "" {
				return errors.MissingParameter("destination", "gxctl copy-h")
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

			srcFile, err := os.Open(copyCmdSource)
			if err != nil {
				return err
			}

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
					nr, err := srcFile.Read(buf)
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
					if nr > 0 {
						msg := ssh.NewWriteToFileMessage(sessionID, tmpFileName, buf[0:nr], false)
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

			err = createSession(conf)
			if err != nil {
				return err
			}

			// At this point we've copied the file into the container... Time for some SCP magic to happen
			conf = &SSHConfig{
				Client:         client,
				DeviceID:       deviceID,
				WaitText:       "Connecting to the device...",
				InitCommand:    fmt.Sprintf("/scp -S /dbclient /%s root@127.0.0.1:%s", tmpFileName, copyCmdDestination),
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

	copyCmd.Flags().StringP("source", "s", "", "Source path of the file")
	copyCmd.Flags().StringP("destination", "d", "", "Destination path of the file")

	copyCmd.SetHelpTemplate(template.HelpTemplate())
	copyCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(copyCmd)

	return &Copy{
		Command: copyCmd,
	}
}
