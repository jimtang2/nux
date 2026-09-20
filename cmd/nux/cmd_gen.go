package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func defaultOtelcolConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".config/nux", "otelcol-config.yaml"), nil
}

func readOTLPEndpointFromConfig(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read otelcol config: %w", err)
	}

	var cfg struct {
		Receivers struct {
			OTLP struct {
				Protocols struct {
					HTTP struct {
						Endpoint string `yaml:"endpoint"`
					} `yaml:"http"`
				} `yaml:"protocols"`
			} `yaml:"otlp"`
		} `yaml:"receivers"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse otelcol config: %w", err)
	}

	ep := cfg.Receivers.OTLP.Protocols.HTTP.Endpoint
	if ep == "" {
		return "", fmt.Errorf("receivers.otlp.protocols.http.endpoint not found in %s", configPath)
	}

	return ep, nil
}

func CmdGen() *cobra.Command {
	var (
		otlpEndpoint    string
		otlpHTTP        bool
		otlpHTTPURLPath string
		otlpInsecure    bool
		logsCount       int
		interval        time.Duration
	)

	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate telemetry using telemetrygen",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If user did not override --otlp-endpoint, try to derive it from config
			if !cmd.Flags().Changed("otlp-endpoint") {
				configPath, err := defaultOtelcolConfigPath()
				if err == nil {
					if ep, err := readOTLPEndpointFromConfig(configPath); err == nil {
						otlpEndpoint = ep
					}
				}
			}

			telemetrygenPath, err := exec.LookPath("telemetrygen")
			if err != nil {
				return fmt.Errorf("telemetrygen not found on PATH: %w", err)
			}

			// Build command:
			// telemetrygen logs \
			//   --otlp-endpoint=<otlpEndpoint> \
			//   --otlp-http \
			//   --otlp-http-url-path=/v1/logs \
			//   --otlp-insecure \
			//   --logs=10 \
			//   --interval=10ms

			argsExec := []string{
				"logs",
				"--otlp-endpoint=" + otlpEndpoint,
			}

			if otlpHTTP {
				argsExec = append(argsExec, "--otlp-http")
			}
			if otlpHTTPURLPath != "" {
				argsExec = append(argsExec, "--otlp-http-url-path="+otlpHTTPURLPath)
			}
			if otlpInsecure {
				argsExec = append(argsExec, "--otlp-insecure")
			}
			if logsCount > 0 {
				argsExec = append(argsExec, fmt.Sprintf("--logs=%d", logsCount))
			}
			if interval > 0 {
				argsExec = append(argsExec, "--interval="+interval.String())
			}

			c := exec.Command(telemetrygenPath, argsExec...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return fmt.Errorf("telemetrygen exited with code %d: %w", exitErr.ExitCode(), err)
				}
				return fmt.Errorf("telemetrygen failed: %w", err)
			}

			return nil
		},
	}

	defaultEndpoint := "localhost:4488"
	if cfgPath, err := defaultOtelcolConfigPath(); err == nil {
		if ep, err := readOTLPEndpointFromConfig(cfgPath); err == nil {
			defaultEndpoint = ep
		}
	}

	genCmd.Flags().StringVarP(&otlpEndpoint, "otlp-endpoint", "e", defaultEndpoint, "OTLP endpoint (host:port)")
	genCmd.Flags().BoolVar(&otlpHTTP, "otlp-http", true, "Use OTLP/HTTP")
	genCmd.Flags().StringVar(&otlpHTTPURLPath, "otlp-http-url-path", "/v1/logs", "OTLP/HTTP URL path")
	genCmd.Flags().BoolVar(&otlpInsecure, "otlp-insecure", true, "Use insecure connection")
	genCmd.Flags().IntVar(&logsCount, "logs", 10, "Number of logs to generate")
	genCmd.Flags().DurationVar(&interval, "interval", 10*time.Millisecond, "Interval between logs")

	return genCmd
}
