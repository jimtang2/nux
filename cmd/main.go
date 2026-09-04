package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "nux",
		Short:             "Nux",
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: false,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if excludePersistentPreRun(cmd, args) {
				return nil
			}
			if err := loadConfigContext(cmd, args); err != nil {
				return err
			}
			return nil
		},
	}
	// cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(InitCmd())
	rootCmd.AddCommand(ColCmd())
	return rootCmd
}

func excludePersistentPreRun(cmd *cobra.Command, args []string) bool {
	if cmd.Use == "init" || cmd.CommandPath() == "nux init" {
		return true
	}
	return false
}

func loadConfigContext(cmd *cobra.Command, args []string) error {
	if err := parseConfig(cmd); err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	return nil
}

/*
nux
|- init
|- col
|  |- start
|	|- stop
|  |- status
|
|- run

*/
