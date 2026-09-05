package collector

import (
	"fmt"
	"strings"
)

func NewOtelCollectorConfig(c Config) (string, error) {
	endpoint := c.OtlpReceiverEndpoint
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	topic := "nux-topic"
	if len(c.KafkaExporterTopic) > 0 {
		topic = c.KafkaExporterTopic
	}

	brokers := []string{"localhost:9092"}
	if len(c.KafkaExporterBrokers) > 0 {
		brokers = c.KafkaExporterBrokers
	}

	brokersYAML := make([]string, len(brokers))
	for i, b := range brokers {
		brokersYAML[i] = fmt.Sprintf(`"%s"`, b)
	}

	return fmt.Sprintf(`receivers:
  otlp:
    protocols:
      http:
        endpoint: %s
exporters:
  kafka:
    brokers: [%s]
    logs:
      topic: %s
      encoding: otlp_proto
service:
  pipelines:
    logs:
      receivers: [otlp]
      exporters: [kafka]
`, endpoint, topic, strings.Join(brokersYAML, ", ")), nil
}
