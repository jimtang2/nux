package collector

import (
	"fmt"
	"net"
	"strconv"
)

func Validate(c Config) error {
	if err := validateEndpoint(c.OtlpReceiverEndpoint); err != nil {
		return fmt.Errorf("oltp_receiver_endpoint: %w", err)
	}

	if len(c.KafkaExporterBrokers) == 0 {
		return fmt.Errorf("kafka_exporter_brokers must not be empty")
	}

	for i, broker := range c.KafkaExporterBrokers {
		if err := validateBrokerAddress(broker); err != nil {
			return fmt.Errorf("kafka_exporter_brokers[%d]: %w", i, err)
		}
	}

	return nil
}

func validateEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint is empty")
	}

	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil
	}

	if host == "" {
		return fmt.Errorf("host is empty in endpoint %q", endpoint)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port %q in endpoint %q", portStr, endpoint)
	}

	if port <= 0 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	return nil
}

func validateBrokerAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("broker address is empty")
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid broker address %q: %w", addr, err)
	}

	if host == "" {
		return fmt.Errorf("broker host is empty in address %q", addr)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid broker port %q: %w", portStr, err)
	}

	if port <= 0 || port > 65535 {
		return fmt.Errorf("broker port must be between 1 and 65535, got %d", port)
	}

	return nil
}
