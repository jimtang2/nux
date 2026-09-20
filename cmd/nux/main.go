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
	rootCmd.AddCommand(CmdInit())
	rootCmd.AddCommand(CmdCol())
	rootCmd.AddCommand(CmdGen())
	kafkaCmd := CmdKafka()
	kafkaTopicCmd := CmdKafkaTopic()
	kafkaTopicCmd.AddCommand(CmdKafkaTopicDelete())
	kafkaCmd.AddCommand(kafkaTopicCmd)
	kafkaCmd.AddCommand(CmdKafkaClear())
	rootCmd.AddCommand(kafkaCmd)
	rootCmd.AddCommand(CmdSimulator())
	return rootCmd
}
