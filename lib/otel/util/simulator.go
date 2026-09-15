package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jimtang2/simulator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"gopkg.in/yaml.v3"
)

// otelClientConfig matches the expected YAML structure for the simulator client.
type otelClientConfig struct {
	CollectorEndpoint string `yaml:"collector_endpoint"`
}

// OtelClient is a client that sends simulator.Event events as OTLP log records.
type OtelClient struct {
	provider *sdklog.LoggerProvider
	exporter *otlploghttp.Exporter
	logger   log.Logger
}

// NewOtelClient creates a new OTLP log client from a config file.
// The config file should contain:
//
//	collector_endpoint: http://localhost:4488
//
// If the file doesn't exist or the endpoint is not specified, it defaults to http://localhost:4488.
func NewOtelClient(configPath string) (*OtelClient, error) {
	endpoint := "http://localhost:4488/v1/logs"

	// Try to read endpoint from config file
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var cfg otelClientConfig
			if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.CollectorEndpoint != "" {
				endpoint = cfg.CollectorEndpoint
			}
		}
	}

	return NewOtelClientFromEndpoint(endpoint)
}

// NewOtelClientFromEndpoint creates a new OTLP log client with the given endpoint.
func NewOtelClientFromEndpoint(endpoint string) (*OtelClient, error) {
	ctx := context.Background()

	// Create OTLP HTTP exporter
	exporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("nux-simulator"),
		),
	)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Use simple processor (no batching) for immediate export
	processor := sdklog.NewSimpleProcessor(exporter)

	// Create logger provider
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)

	// Create logger
	logger := provider.Logger("nux-simulator")

	return &OtelClient{
		provider: provider,
		exporter: exporter,
		logger:   logger,
	}, nil
}

// Send sends a simulator.Event as an OTLP log record.
func (c *OtelClient) Send(ctx context.Context, e simulator.Event) error {
	// Marshal attributes to JSON for the log body
	bodyJSON, err := json.Marshal(e.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	// Create log record
	var rec log.Record
	rec.SetTimestamp(time.Now())
	rec.SetObservedTimestamp(time.Now())
	rec.SetSeverity(log.SeverityInfo)
	rec.SetSeverityText("INFO")
	rec.SetBody(attribute.StringValue(string(bodyJSON)))

	// Add structured attributes
	rec.AddAttributes(
		attribute.String("event.name", e.ActionName),
		attribute.Int64("simulator.turn", e.Turn),
		attribute.Int("user.id", e.PlayerID),
	)

	// Emit the log record
	c.logger.Emit(ctx, rec)

	return nil
}

// ForceFlush flushes any buffered log records.
func (c *OtelClient) ForceFlush(ctx context.Context) error {
	return c.provider.ForceFlush(ctx)
}

// Shutdown gracefully shuts down the OTLP client.
func (c *OtelClient) Shutdown(ctx context.Context) error {
	return c.provider.Shutdown(ctx)
}
