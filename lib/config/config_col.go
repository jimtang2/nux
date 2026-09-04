// config_col.go
package config

import (
	"fmt"
	"strings"
)

func (c *Config) ColConfigYaml() (string, error) {
	endpoint := c.Nux.Col.OltpReceiverEndpoint
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	brokers := []string{"localhost:9092"}
	if len(c.Nux.Col.KafkaExporterBrokers) > 0 {
		brokers = c.Nux.Col.KafkaExporterBrokers
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
      topic: nux-logs
      encoding: otlp_proto
service:
  pipelines:
    logs:
      receivers: [otlp]
      exporters: [kafka]
`, endpoint, strings.Join(brokersYAML, ", ")), nil
}
