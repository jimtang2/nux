package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jimtang2/simulator"
	_ "github.com/jimtang2/simulator-actions/cex"
	"github.com/jimtang2/simulator/otel"
	"github.com/spf13/cobra"
)

func defaultSimulatorConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".config/nux", "simulator-config.yaml"), nil
}

func tlsConfigFromCA(caPath string) (*tls.Config, error) {
	if caPath == "" {
		return nil, nil
	}

	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("--cacert: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("--cacert contains no valid PEM certificates")
	}

	return &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}, nil
}

func CmdSimulator() *cobra.Command {
	var config string
	var otlpURL string
	var cacert string

	simCmd := &cobra.Command{
		Use:   "simulator",
		Short: "Start the simulator and send events to OTLP collector",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("config") {
				var err error
				config, err = defaultSimulatorConfigPath()
				if err != nil {
					return err
				}
			}

			if _, err := os.Stat(config); err != nil {
				return fmt.Errorf("simulator config not found at %s: %w", config, err)
			}

			tlsCfg, err := tlsConfigFromCA(cacert)
			if err != nil {
				cmd.Println(err)
				tlsCfg = nil
			}

			sim, err := simulator.NewSimulator(config)
			if err != nil {
				return fmt.Errorf("failed to create simulator: %w", err)
			}

			cmd.Printf("starting simulator\n")
			cmd.Printf("simulator config: %s\n", config)
			cmd.Printf("OTLP collector endpoint: %s\n", otlpURL)

			otelClient, err := otel.NewOtelClient(otlpURL, tlsCfg)
			if err != nil {
				return fmt.Errorf("failed to create OTLP client: %w", err)
			}
			defer otelClient.Shutdown(context.Background())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			out := make(chan simulator.Event, 256)
			errCh := make(chan error, 1)

			go sim.Continuous(ctx, out, errCh)

			cmd.Println("simulator running, sending events to OTLP collector...")

			count := 0
			for {
				select {
				case event := <-out:
					count++
					if err := otelClient.Send(ctx, event); err != nil {
						log.Printf("failed to send event %d: %v", count, err)
					}
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
	simCmd.Flags().StringVar(&otlpURL, "otlp-url", "", "OTLP collector URL (http or https)")
	simCmd.Flags().StringVar(&cacert, "cacert", "", "PEM file of CA used to verify the collector")
	_ = simCmd.MarkFlagRequired("otlp-url")

	return simCmd
}
