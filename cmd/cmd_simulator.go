package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jimtang2/nux/lib/otel/util"
	"github.com/jimtang2/nux/lib/simulator"
	_ "github.com/jimtang2/nux/lib/simulator-actions/cex"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func defaultSimulatorConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".nux", "simulator-config.yaml"), nil
}

func CmdSimulator() *cobra.Command {
	var config string

	simCmd := &cobra.Command{
		Use:   "simulator",
		Short: "Start the simulator and send events to OTLP collector",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve simulator config path
			if !cmd.Flags().Changed("config") {
				var err error
				config, err = defaultSimulatorConfigPath()
				if err != nil {
					return err
				}
			}

			// Check if simulator config exists
			if _, err := os.Stat(config); err != nil {
				return fmt.Errorf("simulator config not found at %s: %w", config, err)
			}

			// Read config to get OTLP endpoint
			data, err := os.ReadFile(config)
			if err != nil {
				return fmt.Errorf("failed to read simulator config: %w", err)
			}

			var cfg struct {
				OTLPReceiverEndpoint string `yaml:"otlp_receiver_endpoint"`
			}
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("failed to parse simulator config: %w", err)
			}

			// Create simulator
			sim, err := simulator.NewSimulator(config)
			if err != nil {
				return fmt.Errorf("failed to create simulator: %w", err)
			}

			// Get OTLP endpoint from simulator config
			endpoint := sim.OTLPReceiverEndpoint()
			if endpoint == "" {
				endpoint = "http://localhost:4488/v1/logs"
			}

			cmd.Printf("starting simulator\n")
			cmd.Printf("simulator config: %s\n", config)
			cmd.Printf("OTLP collector endpoint: %s\n", endpoint)

			// Create OTLP client
			otelClient, err := util.NewOtelClientFromEndpoint(endpoint)
			if err != nil {
				return fmt.Errorf("failed to create OTLP client: %w", err)
			}
			defer otelClient.Shutdown(context.Background())

			// Start simulator in continuous mode
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			out := make(chan simulator.PlayerAction, 256)
			errCh := make(chan error, 1)

			go sim.Continuous(ctx, out, errCh)

			cmd.Println("simulator running, sending events to OTLP collector...")

			count := 0
			for {
				select {
				case pa := <-out:
					count++
					if err := otelClient.Send(ctx, pa); err != nil {
						log.Printf("failed to send event %d: %v", count, err)
					}
					// Log progress periodically
					if count%100 == 0 {
						cmd.Printf("sent %d events\n", count)
					}
				case err := <-errCh:
					return fmt.Errorf("simulator error: %w", err)
				}
			}
		},
	}

	defaultConfigPath, _ := defaultSimulatorConfigPath()
	simCmd.Flags().StringVarP(&config, "config", "c", defaultConfigPath, "simulator config path")

	return simCmd
}
