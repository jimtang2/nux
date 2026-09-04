package col_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jimtang2/nux/lib/config"
	"github.com/jimtang2/nux/lib/otel/col"
)

func TestOtelCollector_Start_KafkaExporter(t *testing.T) {
	configPath := filepath.Join("testdata", "nux-config.yaml")

	nuxConfig, err := config.Parse(configPath)
	if err != nil {
		t.Fatalf("failed to parse nux config: %v", err)
	}

	c, err := col.NewCollector(nuxConfig)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
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
}
