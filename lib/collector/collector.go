package collector

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/telemetry"
)

type Collector struct {
	collector *otelcol.Collector
}

func NewCollector(cfg Config) (*Collector, error) {
	otelColCfg, err := NewOtelCollectorConfig(cfg)
	if err != nil {
		return nil, err
	}

	factories, err := makeFactories()
	if err != nil {
		return nil, err
	}

	set := otelcol.CollectorSettings{
		Factories: func() (otelcol.Factories, error) {
			return factories, nil
		},
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{"yaml:" + otelColCfg},
				ProviderFactories: []confmap.ProviderFactory{
					yamlprovider.NewFactory(),
					fileprovider.NewFactory(),
				},
			},
		},
	}

	col, err := otelcol.NewCollector(set)
	if err != nil {
		return nil, err
	}

	return &Collector{collector: col}, nil
}

func (c *Collector) Run(ctx context.Context) error {
	return c.collector.Run(ctx)
}

func (c *Collector) Shutdown() {
	c.collector.Shutdown()
}

func makeFactories() (otelcol.Factories, error) {
	var factories otelcol.Factories
	var err error

	factories.Receivers, err = otelcol.MakeFactoryMap(
		otlpreceiver.NewFactory(),
	)
	if err != nil {
		return factories, err
	}

	factories.Exporters, err = otelcol.MakeFactoryMap(
		debugexporter.NewFactory(),
		kafkaexporter.NewFactory(),
	)
	if err != nil {
		return factories, err
	}

	factories.Processors, err = otelcol.MakeFactoryMap[processor.Factory]()
	if err != nil {
		return factories, err
	}

	factories.Extensions, err = otelcol.MakeFactoryMap[extension.Factory]()
	if err != nil {
		return factories, err
	}

	factories.Connectors, err = otelcol.MakeFactoryMap[connector.Factory]()
	if err != nil {
		return factories, err
	}

	factories.Telemetry = telemetry.NewFactory(DefaultTelemetryConfig)

	return factories, nil
}

type telemetryConfig struct{}

func (c *telemetryConfig) Validate() error { return nil }

func DefaultTelemetryConfig() component.Config {
	return &telemetryConfig{}
}
