package util_test

import (
	"context"
	"testing"
	"time"

	"github.com/jimtang2/nux/lib/otel/util"
	"github.com/jimtang2/nux/lib/simulator"
	_ "github.com/jimtang2/nux/lib/simulator-actions/cex"
)

func TestOtelUtil_SimulatorOtelClient(t *testing.T) {
	s, err := simulator.NewSimulator("../../testdata/simulator-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(4 * time.Second)
		cancel()
	}()

	out := make(chan simulator.PlayerAction)
	errCh := make(chan error)
	go s.Continuous(ctx, out, errCh)

	// Create OTLP client
	// Config file path can be empty to use default http://localhost:4488
	// or point to ~/.nux/simulator-config.yaml which should have collector_endpoint
	otelClient, err := util.NewOtelClient("")
	if err != nil {
		t.Fatalf("failed to create otel client: %v", err)
	}
	defer otelClient.Shutdown(ctx)

	for {
		select {
		case pa := <-out:
			t.Log(pa)
			// Send to OTLP collector
			if err := otelClient.Send(ctx, pa); err != nil {
				t.Logf("failed to send event: %v", err)
				// Don't fail the test on send errors; just log
			}
		case err := <-errCh:
			t.Fatal(err)
		case <-ctx.Done():
			t.Logf("test ended (%v)", ctx.Err())
			return
		}
	}
}
