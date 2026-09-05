package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jimtang2/nux/lib/collector"
	"github.com/jimtang2/nux/lib/generator"
	"github.com/spf13/cobra"
)

type Config struct {
	Collector collector.Config
	Generator generator.Config
}

const (
	CONFIG_PATH                       = ".nux"
	COLLECTOR_CONFIG_KEY              = "collector-config"
	COLLECTOR_CONFIG                  = "collector-config.yaml"
	DEFAULT_COLLECTOR_CONFIG_TEMPLATE = `collector:
  otlp_receiver_endpoint: localhost:4444
  kafka_exporter_brokers: ["localhost:9092"]`
	GENERATOR_CONFIG_KEY              = "generator-config"
	GENERATOR_CONFIG                  = "generator-config.yaml"
	DEFAULT_GENERATOR_CONFIG_TEMPLATE = `templates: []`
)

func parseConfig(cmd *cobra.Command) error {
	if err := parseCollectorConfig(cmd); err != nil {
		return err
	}
	if err := parseGeneratorConfig(cmd); err != nil {
		return err
	}
	return nil
}

func parseCollectorConfig(cmd *cobra.Command) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	p := filepath.Join(home, COLLECTOR_CONFIG)
	if _, err := os.Stat(p); err != nil {
		return err
	}
	cfg, err := collector.NewConfig(p)
	if err != nil {
		return err
	}
	if err := collector.Validate(cfg); err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(cmd.Context(), COLLECTOR_CONFIG_KEY, cfg))
	return nil
}

func parseGeneratorConfig(cmd *cobra.Command) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	p := filepath.Join(home, GENERATOR_CONFIG)
	if _, err := os.Stat(p); err != nil {
		return err
	}
	cfg, err := generator.NewConfig(p)
	if err != nil {
		return err
	}
	if err := generator.Validate(cfg); err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(cmd.Context(), GENERATOR_CONFIG_KEY, cfg))
	return nil
}

func getCollectorConfig(cmd *cobra.Command) collector.Config {
	return cmd.Context().Value(COLLECTOR_CONFIG_KEY).(collector.Config)
}

func getGeneratorConfig(cmd *cobra.Command) generator.Config {
	return cmd.Context().Value(GENERATOR_CONFIG_KEY).(generator.Config)
}

func createDefaultConfig(cmd *cobra.Command, args []string) error {
	if err := createDefaultCollectorConfig(cmd, args); err != nil {
		return err
	}
	if err := createDefaultGeneratorConfig(cmd, args); err != nil {
		return err
	}
	return nil
}

func createDefaultCollectorConfig(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, CONFIG_PATH, COLLECTOR_CONFIG)
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configPath); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return nil
		}
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, []byte(DEFAULT_COLLECTOR_CONFIG_TEMPLATE), 0o644); err != nil {
		return err
	}
	return nil
}

func createDefaultGeneratorConfig(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, CONFIG_PATH, GENERATOR_CONFIG)
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configPath); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return nil
		}
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, []byte(DEFAULT_GENERATOR_CONFIG_TEMPLATE), 0o644); err != nil {
		return err
	}
	return nil
}
