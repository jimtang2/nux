package config_test

import (
	"strings"
	"testing"

	"github.com/jimtang2/nux/lib/config"
)

func TestConfig_CreateColConfig(t *testing.T) {
	c, err := config.Parse("testdata/nux-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	yamlStr, err := c.ColConfigYaml()
	if err != nil {
		t.Fatal(err)
	}

	// Basic structure checks
	if !strings.Contains(yamlStr, "receivers:") {
		t.Error("YAML missing 'receivers' section")
	}
	if !strings.Contains(yamlStr, "exporters:") {
		t.Error("YAML missing 'exporters' section")
	}
	if !strings.Contains(yamlStr, "service:") {
		t.Error("YAML missing 'service' section")
	}

	// OTLP receiver checks
	if !strings.Contains(yamlStr, "otlp:") {
		t.Error("YAML missing 'otlp' receiver")
	}
	if !strings.Contains(yamlStr, "endpoint: localhost:4444") {
		t.Errorf("expected endpoint localhost:4444, got:\n%s", yamlStr)
	}

	// Kafka exporter checks
	if !strings.Contains(yamlStr, "kafka:") {
		t.Error("YAML missing 'kafka' exporter")
	}
	if !strings.Contains(yamlStr, `brokers: ["localhost:9092"]`) {
		t.Errorf("expected brokers [\"localhost:9092\"], got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "topic: nux-logs") {
		t.Error("expected topic: nux-logs")
	}
	if !strings.Contains(yamlStr, "encoding: otlp_proto") {
		t.Error("expected encoding: otlp_proto")
	}

	// Pipeline checks
	if !strings.Contains(yamlStr, "pipelines:") {
		t.Error("YAML missing 'pipelines' section")
	}
	if !strings.Contains(yamlStr, "receivers: [otlp]") {
		t.Error("expected receivers: [otlp]")
	}
	if !strings.Contains(yamlStr, "exporters: [kafka]") {
		t.Error("expected exporters: [kafka]")
	}
}
