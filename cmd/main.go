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
	}
	rootCmd.AddCommand(InitCmd())
	rootCmd.AddCommand(ColCmd())
	rootCmd.AddCommand(GenCmd())
	rootCmd.AddCommand(CmdKafka())

	return rootCmd
}
