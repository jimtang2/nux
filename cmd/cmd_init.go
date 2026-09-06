package main

import (
	"os"
	"path/filepath"

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
			if err := setup(); err != nil {
				return err
			}
			cmd.Println("✅ nux initialized (configuration written to ~/.nux)")
			return nil
		},
	}
	return initCmd
}

func setup() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(home, ".nux"), 0o700); err != nil {
		return err
	}
	return nil
}
