package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Nux struct {
		Col struct {
			OltpReceiverEndpoint string   `yaml:"oltp_receiver_endpoint"`
			KafkaExporterBrokers []string `yaml:"kafka_exporter_brokers"`
		} `yaml:"col"`
	} `yaml:"nux"`
}

func Parse(p string) (*Config, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
