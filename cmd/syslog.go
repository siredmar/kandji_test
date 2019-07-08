package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/pkg/term"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type Syslog struct {
	Command *cobra.Command
}

func NewSyslog(parent *cobra.Command, client *client.APIClient) *Syslog {
	var syslogCmd = &cobra.Command{
		Use:                   "syslog ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "syslog from different devices",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl syslog -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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
					msg := ssh.NewExecuteCommandMessage(sessionID, b)
					_, err = json.Marshal(msg)
					input <- b
				}
			}()

			conf := &SSHConfig{
				Client:         client,
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
		},
	}

	syslogCmd.SetHelpTemplate(template.HelpTemplate())
	syslogCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(syslogCmd)

	return &Syslog{
		Command: syslogCmd,
	}
}
