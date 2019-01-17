package cmd

import (
	"github.com/spf13/cobra"
)

type Get struct {
	Command *cobra.Command
}

func NewGet(parent *cobra.Command) *Get {
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get different resources",
		Long:  `TODO`,
	}

	getCmd.PersistentFlags().StringP("device-id", "d", "", "specify device id")
	getCmd.PersistentFlags().StringP("output", "o", "", "print the result in a different format. Currently supported json")
	parent.AddCommand(getCmd)

	return &Get{
		Command: getCmd,
	}
}
