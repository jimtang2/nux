package otel_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jimtang2/nux/lib/otel"
)

func TestCollectorFactory(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantErr    bool
	}{
		{
			name:       "OtelBasic",
			configPath: "../testdata/otel-collector-config.yaml",
			wantErr:    false,
		},
		{
			name:       "CustomKafkaConfig",
			configPath: "../testdata/otel-collector-config-kafka.yaml",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := os.ReadFile(tt.configPath)
			if err != nil {
				t.Fatal(err)
			}
			c, err := otel.NewCollector(string(config))
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

			time.Sleep(50 * time.Millisecond)

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
