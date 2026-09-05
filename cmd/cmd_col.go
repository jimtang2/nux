package main

import (
	"github.com/spf13/cobra"
)

func ColCmd() *cobra.Command {
	colCmd := &cobra.Command{
		Use:           "col",
		Short:         "OpenTelemetry collector commands",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := findColProc(cmd)
			if err != nil {
				return err
			}
			cfg := getCollectorConfig(cmd)
			cmd.Printf(`otel collector pid: %v
otlp receiver endpoints:
  - %v/v1/logs POST
  - %v/v1/metrics POST
  - %v/v1/traces POST
kafka exporter brokers: %v
`,
				pid,
				cfg.OtlpReceiverEndpoint,
				cfg.OtlpReceiverEndpoint,
				cfg.OtlpReceiverEndpoint,
				cfg.KafkaExporterBrokers)
			return nil
		},
	}
	colCmd.AddCommand(ColStartCmd())
	colCmd.AddCommand(ColStopCmd())
	return colCmd
}

func ColStartCmd() *cobra.Command {
	var daemon bool

	startCmd := &cobra.Command{
		Use:          "start",
		Short:        "Start the OpenTelemetry collector service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if daemon {
				return runCollectorDaemon(cmd)
			}
			pid, err := startColProc(cmd)
			if err != nil {
				return err
			}
			cmd.Printf("otel collector pid: %v\n", pid)
			return nil
		},
	}
	startCmd.Flags().BoolVar(&daemon, "daemon", false, "Run as background daemon (internal use)")
	return startCmd
}

func ColStopCmd() *cobra.Command {
	stopCmd := &cobra.Command{
		Use:          "stop",
		Short:        "Stop the OpenTelemetry collector service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := stopColProc(cmd); err != nil {
				return err
			}
			cmd.Println("otel collector pid: -1 (closed)")
			return nil
		},
	}
	return stopCmd
}
