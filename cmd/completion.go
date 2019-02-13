package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type Completion struct {
	Command *cobra.Command
}

func NewCompletion(root *cobra.Command) *Completion {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Show bash completions",
		Long:  `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := root.GenBashCompletion(os.Stdout)
			if err != nil {
				return err
			}

			return nil
		},
	}

	root.AddCommand(completionCmd)

	return &Completion{
		Command: completionCmd,
	}
}
