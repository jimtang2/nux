package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

func ColCmd() *cobra.Command {
	var config string

	colCmd := &cobra.Command{
		Use:           "col",
		Short:         "Start Otel Collector",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If user did not override config, use resolved default
			if !cmd.Flags().Changed("config") {
				defaultPath, err := defaultConfigPath()
				if err != nil {
					return err
				}
				config = defaultPath
			}
			if _, err := os.Stat(config); err != nil {
				return fmt.Errorf("collector config not found at %s: %w", config, err)
			}
			if err := validateOtelcolConfig(config); err != nil {
				return err
			}
			if err := runOtelcol(config); err != nil {
				return err
			}
			return nil
		},
	}

	defaultPath, _ := defaultConfigPath()
	colCmd.Flags().StringVarP(&config, "config", "c", defaultPath, "otelcol config path")

	return colCmd
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".nux", "otelcol-config.yaml"), nil
}

func validateOtelcolConfig(configPath string) error {
	otelcolPath, err := exec.LookPath("otelcol")
	if err != nil {
		return fmt.Errorf("otelcol not found on PATH: %w", err)
	}

	cmd := exec.Command(otelcolPath, "validate", "--config", configPath)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("otelcol validate failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("otelcol validate failed: %w", err)
	}

	return nil
}

func runOtelcol(configPath string) error {
	otelcolPath, err := exec.LookPath("otelcol")
	if err != nil {
		return fmt.Errorf("otelcol not found on PATH: %w", err)
	}

	cmd := exec.Command(otelcolPath, "--config", configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("otelcol exited with code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("otelcol failed: %w", err)
	}

	return nil
}
