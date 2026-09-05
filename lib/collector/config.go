package collector

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OtlpReceiverEndpoint string   `yaml:"otlp_receiver_endpoint"`
	KafkaExporterTopic   string   `yaml:"kafka_exporter_topic"`
	KafkaExporterBrokers []string `yaml:"kafka_exporter_brokers"`
}

func NewConfig(p string) (Config, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return Config{}, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, err
	}

	return c, nil
}

func (c Config) IsEmpty() bool {
	return len(c.OtlpReceiverEndpoint) == 0 && len(c.KafkaExporterBrokers) == 0 && len(c.KafkaExporterTopic) == 0
}
