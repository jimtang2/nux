package server

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/jimtang2/nux/lib/simulator"
	_ "github.com/jimtang2/nux/lib/simulator-actions/cex"
)

func TestWSHandler_Broadcast(t *testing.T) {
	sim, err := simulator.NewSimulator("../testdata/simulator-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	handler := NewWSHandler(sim)

	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start broadcaster in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.Broadcast(ctx)
	}()

	// Connect WebSocket client using coder/websocket
	u := "ws" + server.URL[4:] // replace http with ws
	c, _, err := websocket.Dial(context.Background(), u, nil)
	if err != nil {
		t.Fatal("dial error:", err)
	}
	defer c.CloseNow()

	// Set read deadline via context
	readCtx, readCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer readCancel()

	received := 0
	maxMessages := 10

	for received < maxMessages {
		var pa simulator.PlayerAction
		err := wsjson.Read(readCtx, c, &pa)
		if err != nil {
			t.Log("read error (may be expected on ctx cancel):", err)
			break
		}

		received++
		t.Log(pa)
	}

	if received == 0 {
		t.Error("expected at least one message, got none")
	}

	cancel()
	<-errCh
}
