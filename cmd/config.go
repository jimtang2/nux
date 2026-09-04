package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jimtang2/nux/lib/config"
	"github.com/spf13/cobra"
)

const (
	CONFIG_KEY              = "config"
	CONFIG_SUBPATH          = ".nux/config.yaml"
	DEFAULT_CONFIG_TEMPLATE = `nux:
  col:
    oltp_receiver_endpoint: localhost:4444
    kafka_exporter_brokers: ["localhost:9092"]
`
)

func parseConfig(cmd *cobra.Command) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find user home directory")
	}

	configPath := filepath.Join(home, CONFIG_SUBPATH)
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("cannot access directory $HOME/%s or command not initialized (run `nux init`); %w", CONFIG_SUBPATH, err)
	}

	conf, err := config.Parse(configPath)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("validate config error: %w", err)
	}

	cmd.SetContext(context.WithValue(cmd.Context(), CONFIG_KEY, conf))
	return nil
}

func getConfig(cmd *cobra.Command) *config.Config {
	return cmd.Context().Value(CONFIG_KEY).(*config.Config)
}

func createDefaultConfig(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find user home directory: %w", err)
	}

	configPath := filepath.Join(home, CONFIG_SUBPATH)
	configDir := filepath.Dir(configPath)

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			// Config exists and --force not set: do nothing
			return nil
		}
		// --force set: proceed to overwrite
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("cannot create config directory %s: %w", configDir, err)
	}

	if err := os.WriteFile(configPath, []byte(DEFAULT_CONFIG_TEMPLATE), 0o644); err != nil {
		return fmt.Errorf("cannot write config file %s: %w", configPath, err)
	}

	return nil
}
