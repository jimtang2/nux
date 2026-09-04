package col_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/telemetry"
)

type telemetryConfig struct{}

func (c *telemetryConfig) Validate() error { return nil }

func defaultTelemetryConfig() component.Config {
	return &telemetryConfig{}
}

func makeMinTestFactories() (otelcol.Factories, error) {
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

func TestOtelCollector_Start_MinimumConfig(t *testing.T) {
	configPath := filepath.Join("testdata", "otel-col-min.yaml")

	// Ensure config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file does not exist: %s", configPath)
	}

	factories, err := makeMinTestFactories()
	if err != nil {
		t.Fatalf("failed to build factories: %v", err)
	}

	set := otelcol.CollectorSettings{
		Factories: func() (otelcol.Factories, error) {
			return factories, nil
		},
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{"file:" + configPath},
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
				},
			},
		},
	}

	col, err := otelcol.NewCollector(set)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	errCh := make(chan error, 1)
	go func() {
		defer wg.Done()
		errCh <- col.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)

	col.Shutdown()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("collector run failed: %v", err)
		}
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("collector did not exit after shutdown")
	}
}
