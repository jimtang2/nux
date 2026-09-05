package collector_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jimtang2/nux/lib/collector"
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

func TestConfig_CreateYaml(t *testing.T) {
	c, err := collector.NewConfig("../testdata/collector-config-example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	otelColCfg, err := collector.NewOtelCollectorConfig(c)
	if err != nil {
		t.Fatal(err)
	}

	// Basic structure checks
	if !strings.Contains(otelColCfg, "receivers:") {
		t.Error("YAML missing 'receivers' section")
	}
	if !strings.Contains(otelColCfg, "exporters:") {
		t.Error("YAML missing 'exporters' section")
	}
	if !strings.Contains(otelColCfg, "service:") {
		t.Error("YAML missing 'service' section")
	}

	// OTLP receiver checks
	if !strings.Contains(otelColCfg, "otlp:") {
		t.Error("YAML missing 'otlp' receiver")
	}
	if !strings.Contains(otelColCfg, "endpoint: localhost:4444") {
		t.Errorf("expected endpoint localhost:4444, got:\n%s", otelColCfg)
	}

	// Kafka exporter checks
	if !strings.Contains(otelColCfg, "kafka:") {
		t.Error("YAML missing 'kafka' exporter")
	}
	if !strings.Contains(otelColCfg, `brokers: ["localhost:9092"]`) {
		t.Errorf("expected brokers [\"localhost:9092\"], got:\n%s", otelColCfg)
	}
	if !strings.Contains(otelColCfg, "topic: nux-logs") {
		t.Error("expected topic: nux-logs")
	}
	if !strings.Contains(otelColCfg, "encoding: otlp_proto") {
		t.Error("expected encoding: otlp_proto")
	}

	// Pipeline checks
	if !strings.Contains(otelColCfg, "pipelines:") {
		t.Error("YAML missing 'pipelines' section")
	}
	if !strings.Contains(otelColCfg, "receivers: [otlp]") {
		t.Error("expected receivers: [otlp]")
	}
	if !strings.Contains(otelColCfg, "exporters: [kafka]") {
		t.Error("expected exporters: [kafka]")
	}
}

type TestCollector interface {
	Run(context.Context) error
	Shutdown()
}

type testCollectorFactory func() (TestCollector, error)

func TestCollectorFactory(t *testing.T) {
	tests := []struct {
		name    string
		newCol  testCollectorFactory
		wantErr bool
	}{
		{
			name: "OtelBasic",
			newCol: func() (TestCollector, error) {
				set := otelcol.CollectorSettings{
					Factories: func() (otelcol.Factories, error) {
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

						factories.Telemetry = telemetry.NewFactory(collector.DefaultTelemetryConfig)

						return factories, nil
					},
					ConfigProviderSettings: otelcol.ConfigProviderSettings{
						ResolverSettings: confmap.ResolverSettings{
							URIs: []string{"file:../testdata/otel-collector-config.yaml"},
							ProviderFactories: []confmap.ProviderFactory{
								fileprovider.NewFactory(),
							},
						},
					},
				}

				return otelcol.NewCollector(set)
			},
			wantErr: false,
		},
		{
			name: "CustomKafkaConfig",
			newCol: func() (TestCollector, error) {
				nuxConfig, err := collector.NewConfig("../testdata/collector-config-example.yaml")
				if err != nil {
					return nil, err
				}
				return collector.NewCollector(nuxConfig)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := tt.newCol()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to create collector: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("expected error creating collector, got nil")
			}

			var wg sync.WaitGroup
			wg.Add(1)

			errCh := make(chan error, 1)
			go func() {
				defer wg.Done()
				errCh <- c.Run(context.Background())
			}()

			time.Sleep(100 * time.Millisecond)

			c.Shutdown()

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
		})
	}

}
