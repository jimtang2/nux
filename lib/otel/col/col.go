package col

import (
	"context"

	"github.com/jimtang2/nux/lib/config"
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

type Col struct {
	collector *otelcol.Collector
}

func NewCollector(cfg *config.Config) (*Col, error) {
	colYAML, err := cfg.ColConfigYaml()
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
				URIs: []string{"yaml:" + colYAML},
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

	return &Col{collector: col}, nil
}

func (c *Col) Run(ctx context.Context) error {
	return c.collector.Run(ctx)
}

func (c *Col) Shutdown() {
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

	factories.Telemetry = telemetry.NewFactory(defaultTelemetryConfig)

	return factories, nil
}

type telemetryConfig struct{}

func (c *telemetryConfig) Validate() error { return nil }

func defaultTelemetryConfig() component.Config {
	return &telemetryConfig{}
}
