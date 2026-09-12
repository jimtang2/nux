package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"gopkg.in/yaml.v3"
)

const (
	KafkaExporterEncodingProto = "otlp_proto"
	KafkaExporterEncodingJSON  = "otlp_json"
)

func DecodeOTLPProtoLogs(data []byte) (plog.Logs, error) {
	req := plogotlp.NewExportRequest()
	if err := req.UnmarshalProto(data); err != nil {
		return plog.Logs{}, err
	}
	return req.Logs(), nil
}

func DecodeOTLPJSONLogs(data []byte) (plog.Logs, error) {
	req := plogotlp.NewExportRequest()
	if err := req.UnmarshalJSON(data); err != nil {
		return plog.Logs{}, err
	}
	return req.Logs(), nil
}

// kafkaExporterConfig matches the shape of a single kafka exporter config:
//
//	kafka:
//	  brokers: [...]
//	  logs:
//	    topic: ...
//	    encoding: ...
//	  auth: ...
type kafkaExporterConfig struct {
	Brokers []string        `yaml:"brokers"`
	Logs    kafkaLogsConfig `yaml:"logs"`
	Auth    map[string]any  `yaml:"auth"`
}

type kafkaLogsConfig struct {
	Topic    string `yaml:"topic"`
	Encoding string `yaml:"encoding"`
}

// GetKafkaExporterEncoding reads the Kafka exporter config from the given
// collector config YAML and returns the encoding for logs (e.g. "otlp_proto"
// or "otlp_json"). It expects a single inline kafka exporter under
// exporters.kafka.
func GetKafkaExporterEncoding(configYAML string) (string, error) {
	var cfg struct {
		Exporters struct {
			Kafka kafkaExporterConfig `yaml:"kafka"`
		} `yaml:"exporters"`
	}

	if err := yaml.Unmarshal([]byte(configYAML), &cfg); err != nil {
		return "", fmt.Errorf("failed to parse collector config: %w", err)
	}

	if cfg.Exporters.Kafka.Logs.Encoding == "" {
		return "", fmt.Errorf("no kafka exporter logs.encoding found in config")
	}

	return cfg.Exporters.Kafka.Logs.Encoding, nil
}

// GetKafkaExporterTopic returns the topic name configured for logs in the
// kafka exporter. It expects a single inline kafka exporter under
// exporters.kafka.
func GetKafkaExporterTopic(configYAML string) (string, error) {
	var cfg struct {
		Exporters struct {
			Kafka kafkaExporterConfig `yaml:"kafka"`
		} `yaml:"exporters"`
	}

	if err := yaml.Unmarshal([]byte(configYAML), &cfg); err != nil {
		return "", fmt.Errorf("failed to parse collector config: %w", err)
	}

	if cfg.Exporters.Kafka.Logs.Topic == "" {
		return "", fmt.Errorf("no kafka exporter logs.topic found in config")
	}

	return cfg.Exporters.Kafka.Logs.Topic, nil
}

// ValidateConfig uses the confmap resolver to validate that the config can be
// loaded by the collector. This is a lightweight check that the YAML is valid
// and compatible with the collector's config schema.
func ValidateConfig(configYAML string) error {
	resolver, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{"yaml:" + configYAML},
		ProviderFactories: []confmap.ProviderFactory{
			yamlprovider.NewFactory(),
			fileprovider.NewFactory(),
		},
	})
	if err != nil {
		return err
	}

	_, err = resolver.Resolve(context.Background())
	return err
}
