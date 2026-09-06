package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type otelcolConfig struct {
	Receivers struct {
		OTLP struct {
			Protocols struct {
				HTTP struct {
					Endpoint string `yaml:"endpoint"`
				} `yaml:"http"`
			} `yaml:"protocols"`
		} `yaml:"otlp"`
	} `yaml:"receivers"`
	Exporters struct {
		Kafka struct {
			Brokers []string `yaml:"brokers"`
			Logs    struct {
				Topic    string `yaml:"topic"`
				Encoding string `yaml:"encoding"`
			} `yaml:"logs"`
			Auth struct {
				SASL struct {
					Username  string `yaml:"username"`
					Password  string `yaml:"password"`
					Mechanism string `yaml:"mechanism"`
				} `yaml:"sasl"`
			} `yaml:"auth"`
		} `yaml:"kafka"`
	} `yaml:"exporters"`
}

func loadOtelcolKafkaConfig(configPath string) (*otelcolConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read otelcol config: %w", err)
	}

	var cfg otelcolConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse otelcol config: %w", err)
	}

	if len(cfg.Exporters.Kafka.Brokers) == 0 {
		return nil, fmt.Errorf("no kafka brokers defined in otelcol config")
	}
	if cfg.Exporters.Kafka.Logs.Topic == "" {
		return nil, fmt.Errorf("no kafka logs topic defined in otelcol config")
	}
	if cfg.Exporters.Kafka.Auth.SASL.Username == "" || cfg.Exporters.Kafka.Auth.SASL.Password == "" {
		return nil, fmt.Errorf("kafka SASL credentials missing in otelcol config")
	}

	return &cfg, nil
}

func defaultKafkaConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".nux", "otelcol-config.yaml"), nil
}

func CmdKafka() *cobra.Command {
	var (
		config    string
		brokers   string
		topic     string
		username  string
		password  string
		mechanism string
	)

	kafkaCmd := &cobra.Command{
		Use:   "kafka",
		Short: "Consume logs from the Kafka topic used by otelcol",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("config") {
				var err error
				config, err = defaultKafkaConfigPath()
				if err != nil {
					return err
				}
			}

			if _, err := os.Stat(config); err != nil {
				return fmt.Errorf("otelcol config not found at %s: %w", config, err)
			}

			cfg, err := loadOtelcolKafkaConfig(config)
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("brokers") {
				brokers = cfg.Exporters.Kafka.Brokers[0]
			}
			if !cmd.Flags().Changed("topic") {
				topic = cfg.Exporters.Kafka.Logs.Topic
			}
			if !cmd.Flags().Changed("username") {
				username = cfg.Exporters.Kafka.Auth.SASL.Username
			}
			if !cmd.Flags().Changed("password") {
				password = cfg.Exporters.Kafka.Auth.SASL.Password
			}
			if !cmd.Flags().Changed("mechanism") {
				mechanism = cfg.Exporters.Kafka.Auth.SASL.Mechanism
			}

			kcatPath, err := exec.LookPath("kcat")
			if err != nil {
				return fmt.Errorf("kcat not found on PATH: %w", err)
			}

			argsExec := []string{
				"-b", brokers,
				"-X", "security.protocol=SASL_PLAINTEXT",
				"-X", fmt.Sprintf("sasl.mechanisms=%s", mechanism),
				"-X", fmt.Sprintf("sasl.username=%s", username),
				"-X", fmt.Sprintf("sasl.password=%s", password),
				"-C",
				"-t", topic,
			}

			c := exec.Command(kcatPath, argsExec...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return fmt.Errorf("kcat exited with code %d: %w", exitErr.ExitCode(), err)
				}
				return fmt.Errorf("kcat failed: %w", err)
			}

			return nil
		},
	}

	kafkaCmd.Flags().StringVarP(&config, "config", "c", "", "otelcol config path (default: ~/.nux/otelcol-config.yaml)")
	kafkaCmd.Flags().StringVarP(&brokers, "brokers", "b", "", "Kafka brokers (default: from otelcol config)")
	kafkaCmd.Flags().StringVarP(&topic, "topic", "t", "", "Kafka topic (default: from otelcol config)")
	kafkaCmd.Flags().StringVar(&username, "username", "", "SASL username (default: from otelcol config)")
	kafkaCmd.Flags().StringVar(&password, "password", "", "SASL password (default: from otelcol config)")
	kafkaCmd.Flags().StringVar(&mechanism, "mechanism", "", "SASL mechanism (default: from otelcol config)")

	kafkaCmd.AddCommand(CmdKafkaClear())
	return kafkaCmd
}

func CmdKafkaClear() *cobra.Command {
	var (
		config    string
		brokers   string
		topic     string
		username  string
		password  string
		mechanism string
		timeout   time.Duration
	)

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete all records from the Kafka logs topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("config") {
				var err error
				config, err = defaultKafkaConfigPath()
				if err != nil {
					return err
				}
			}

			if _, err := os.Stat(config); err != nil {
				return fmt.Errorf("otelcol config not found at %s: %w", config, err)
			}

			cfg, err := loadOtelcolKafkaConfig(config)
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("brokers") {
				brokers = cfg.Exporters.Kafka.Brokers[0]
			}
			if !cmd.Flags().Changed("topic") {
				topic = cfg.Exporters.Kafka.Logs.Topic
			}
			if !cmd.Flags().Changed("username") {
				username = cfg.Exporters.Kafka.Auth.SASL.Username
			}
			if !cmd.Flags().Changed("password") {
				password = cfg.Exporters.Kafka.Auth.SASL.Password
			}
			if !cmd.Flags().Changed("mechanism") {
				mechanism = cfg.Exporters.Kafka.Auth.SASL.Mechanism
			}

			saramaCfg := sarama.NewConfig()
			saramaCfg.Version = sarama.V2_8_0_0
			saramaCfg.Net.SASL.Enable = true
			saramaCfg.Net.SASL.User = username
			saramaCfg.Net.SASL.Password = password
			saramaCfg.Net.SASL.Mechanism = sarama.SASLMechanism(mechanism)
			saramaCfg.Net.TLS.Enable = false
			saramaCfg.ClientID = "nux-kafka-clear"
			saramaCfg.Admin.Timeout = timeout

			client, err := sarama.NewClient([]string{brokers}, saramaCfg)
			if err != nil {
				return fmt.Errorf("failed to create kafka client: %w", err)
			}
			defer client.Close()

			admin, err := sarama.NewClusterAdminFromClient(client)
			if err != nil {
				return fmt.Errorf("failed to create kafka admin: %w", err)
			}
			defer admin.Close()

			partitions, err := client.Partitions(topic)
			if err != nil {
				return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
			}

			// Build DeleteRecords request per partition
			for _, p := range partitions {
				latest, err := client.GetOffset(topic, p, sarama.OffsetNewest)
				if err != nil {
					return fmt.Errorf("failed to get latest offset for %s:%d: %w", topic, p, err)
				}

				partitionOffsets := map[int32]int64{
					p: latest,
				}

				if err := admin.DeleteRecords(topic, partitionOffsets); err != nil {
					return fmt.Errorf("failed to delete records for %s:%d: %w", topic, p, err)
				}
			}

			cmd.Printf("cleared all records from topic %s\n", topic)
			return nil
		},
	}

	clearCmd.Flags().StringVarP(&config, "config", "c", "", "otelcol config path (default: ~/.nux/otelcol-config.yaml)")
	clearCmd.Flags().StringVarP(&brokers, "brokers", "b", "", "Kafka brokers (default: from otelcol config)")
	clearCmd.Flags().StringVarP(&topic, "topic", "t", "", "Kafka topic (default: from otelcol config)")
	clearCmd.Flags().StringVar(&username, "username", "", "SASL username (default: from otelcol config)")
	clearCmd.Flags().StringVar(&password, "password", "", "SASL password (default: from otelcol config)")
	clearCmd.Flags().StringVar(&mechanism, "mechanism", "", "SASL mechanism (default: from otelcol config)")
	clearCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "timeout for kafka admin operations")

	return clearCmd
}
