package simulator_test

import (
	"context"
	"testing"
	"time"

	"github.com/jimtang2/nux/lib/simulator"
	_ "github.com/jimtang2/nux/lib/simulator-actions/cex"
)

func TestSimulator_NextNTurn(t *testing.T) {
	s, err := simulator.NewSimulator("../testdata/simulator-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	out := make(chan simulator.PlayerAction)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case pa := <-out:
				t.Log(pa)
			case <-ctx.Done():
				return
			}
		}
	}()

	err = s.NextNTurn(2, out)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(20 * time.Millisecond)
}

func TestSimulator_Continuous(t *testing.T) {
	s, err := simulator.NewSimulator("../testdata/simulator-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	out := make(chan simulator.PlayerAction)
	errCh := make(chan error)
	go s.Continuous(ctx, out, errCh)

	for {
		select {
		case pa := <-out:
			t.Log(pa)
		case err := <-errCh:
			t.Fatal(err)
		case <-ctx.Done():
			t.Logf("test ended (%v)", ctx.Err())
			return
		}
	}
}
