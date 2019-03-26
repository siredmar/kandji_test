package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type Syslog struct {
	Command *cobra.Command
}

func NewSyslog(parent *cobra.Command) *Syslog {
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

			fmt.Println("Connecting to the device...")
			err = createSession(client, deviceID, "journalctl -f", false)
			if err != nil {
				return err
			}

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
