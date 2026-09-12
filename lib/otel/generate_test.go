package otel_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/jimtang2/nux/lib/kafka"
	"github.com/jimtang2/nux/lib/otel"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

func loadEnv() {
	viper.SetEnvPrefix("nux")
	viper.AutomaticEnv()
}

// newTestLoggerProvider creates an OTLP/HTTP logger provider for tests.
// Returns the provider and logger; caller must shutdown the provider.
func newTestLoggerProvider(ctx context.Context, endpoint string) (*sdklog.LoggerProvider, log.Logger, error) {
	exporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithEndpointURL(endpoint),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("nux-generator-test"),
		),
	)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	processor := sdklog.NewBatchProcessor(
		exporter,
		sdklog.WithMaxQueueSize(2048),
		sdklog.WithExportMaxBatchSize(512),
		sdklog.WithExportInterval(1*time.Second),
		sdklog.WithExportTimeout(10*time.Second),
	)

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)

	logger := provider.Logger("nux-generator")

	return provider, logger, nil
}

// newTestLogRecord creates a log.Record with test data.
func newTestLogRecord(i int) log.Record {
	var rec log.Record
	rec.SetTimestamp(time.Now())
	rec.SetObservedTimestamp(time.Now())
	rec.SetSeverity(log.SeverityInfo)
	rec.SetSeverityText("INFO")
	rec.SetBody(attribute.StringValue(fmt.Sprintf(`{"event":"test-log","i":%d}`, i)))
	rec.AddAttributes(
		attribute.String("test.id", fmt.Sprintf("run-%d", i)),
		attribute.String("event.name", "test_log"),
	)
	return rec
}

// startCollectorForTest starts a collector from the given config path and
// returns a shutdown function. It fails the test if the collector cannot start.
func startCollectorForTest(t *testing.T, configPath string) (*otel.Collector, func()) {
	t.Helper()

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read collector config: %v", err)
	}

	c, err := otel.NewCollector(string(configBytes))
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	errCh := make(chan error, 1)
	go func() {
		defer wg.Done()
		if err := c.Run(ctx); err != nil && err != context.Canceled {
			errCh <- err
		}
	}()

	// Give the collector a moment to start.
	time.Sleep(200 * time.Millisecond)

	shutdown := func() {
		c.Shutdown()
		cancel()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case err := <-errCh:
			t.Fatalf("collector run failed: %v", err)
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("collector did not exit after shutdown")
		}
	}

	return c, shutdown
}

// startKafkaConsumerForTest starts a consumer on the given topic and returns
// a function that waits for at least wantRecords log records to arrive (or timeout)
// and returns the total count.
func startKafkaConsumerForTest(t *testing.T, brokers, credentials, topic string, encoding string) func(ctx context.Context, wantRecords int) (int, error) {
	t.Helper()

	kc := &kafka.Kafka{}
	if err := kc.Connect(brokers, credentials); err != nil {
		t.Fatalf("kafka connect failed: %v", err)
	}

	if err := kc.InitConsumer(); err != nil {
		t.Fatalf("kafka init consumer failed: %v", err)
	}

	if err := kc.InitAdmin(); err != nil {
		t.Fatalf("kafka init admin failed: %v", err)
	}

	if err := kc.CreateTopic(topic, 1, 1); err != nil {
		if s := err.Error(); s != "" && s != "topic already exists" {
			// t.Logf("create topic error (may be already exists): %v", err)
		}
	}

	pc, err := kc.Consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("failed to consume partition: %v", err)
	}

	decode := func(data []byte) (plog.Logs, error) {
		switch encoding {
		case otel.KafkaExporterEncodingJSON:
			return otel.DecodeOTLPJSONLogs(data)
		default:
			return otel.DecodeOTLPProtoLogs(data)
		}
	}

	waitForRecords := func(ctx context.Context, wantRecords int) (int, error) {
		defer pc.Close()
		defer kc.Close()

		totalLogRecords := 0
		timeout := time.After(10 * time.Second)

		for totalLogRecords < wantRecords {
			select {
			case msg := <-pc.Messages():
				logs, err := decode(msg.Value)
				if err != nil {
					t.Logf("failed to decode OTLP logs: %v", err)
					continue
				}

				for i := range logs.ResourceLogs().Len() {
					rl := logs.ResourceLogs().At(i)
					for j := range rl.ScopeLogs().Len() {
						sl := rl.ScopeLogs().At(j)
						totalLogRecords += sl.LogRecords().Len()
					}
				}

			case <-timeout:
				return totalLogRecords, fmt.Errorf("timeout waiting for logs, got %d log records", totalLogRecords)
			case <-ctx.Done():
				return totalLogRecords, ctx.Err()
			}
		}

		return totalLogRecords, nil
	}

	return waitForRecords
}

func TestGenerateLog(t *testing.T) {
	loadEnv()

	brokers := viper.GetString("kafka_brokers")
	if brokers == "" {
		t.Skip("NUX_KAFKA_BROKERS not set, skipping test")
	}

	credentials := viper.GetString("kafka_credentials")

	configPath := "../testdata/otel-collector-config-kafka.yaml"

	c, shutdownCollector := startCollectorForTest(t, configPath)
	defer shutdownCollector()

	topic, err := c.KafkaExporterTopic()
	if err != nil {
		t.Fatalf("failed to get kafka exporter topic: %v", err)
	}

	encoding, err := c.KafkaExporterEncoding()
	if err != nil {
		t.Fatalf("failed to get kafka exporter encoding: %v", err)
	}

	// Connect to Kafka and clear the topic before the test.
	kc := &kafka.Kafka{}
	if err := kc.Connect(brokers, credentials); err != nil {
		t.Fatalf("kafka connect failed: %v", err)
	}
	if err := kc.InitAdmin(); err != nil {
		t.Fatalf("kafka init admin failed: %v", err)
	}

	// Ensure topic exists (create if needed).
	if err := kc.CreateTopic(topic, 1, 1); err != nil {
		if s := err.Error(); s != "" && s != "topic already exists" {
			// t.Logf("create topic error (may be already exists): %v", err)
		}
	}

	// Clear existing records so this test only sees newly emitted logs.
	if err := kc.ClearTopic(topic); err != nil {
		t.Fatalf("failed to clear topic %s: %v", topic, err)
	}

	// Defer a cleanup that clears the topic again after the test (optional).
	t.Cleanup(func() {
		if err := kc.ClearTopic(topic); err != nil {
			t.Logf("failed to clear topic %s in cleanup: %v", topic, err)
		}
		_ = kc.Close()
	})

	waitForRecords := startKafkaConsumerForTest(t, brokers, credentials, topic, encoding)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create OTel logger provider and logger directly.
	provider, logger, err := newTestLoggerProvider(ctx, "http://localhost:4318/v1/logs")
	if err != nil {
		t.Fatalf("failed to create logger provider: %v", err)
	}
	defer provider.Shutdown(ctx)

	count := 500

	// Emit logs using the logger directly.
	for i := range count {
		rec := newTestLogRecord(i)
		logger.Emit(ctx, rec)
	}

	// Give exporter/collector time to flush.
	time.Sleep(500 * time.Millisecond)

	// Wait for records.
	total, err := waitForRecords(ctx, count)
	if err != nil {
		t.Logf("waitForRecords error: %v", err)
	}

	if total != count {
		t.Fatalf("expected %d log records, got %d", count, total)
	}
}
