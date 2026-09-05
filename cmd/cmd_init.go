package main

import (
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:               "init",
		Short:             "Initialize nux configuration",
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: false,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := createDefaultConfig(cmd, args); err != nil {
				return err
			}
			cmd.Println("✅ nux initialized (configuration written to ~/.nux)")
			return nil
		},
	}
	return initCmd
}
