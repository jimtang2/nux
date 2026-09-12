package server

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/jimtang2/nux/lib/simulator"
)

// WSHandler broadcasts simulator events to all connected WebSocket clients.
type WSHandler struct {
	sim     *simulator.Simulator
	clients map[*wsConn]struct{}
	mu      sync.RWMutex
}

type wsConn struct {
	conn *websocket.Conn
	send chan simulator.PlayerAction
}

// NewWSHandler creates a new WebSocket handler that broadcasts simulator.PlayerAction events.
func NewWSHandler(sim *simulator.Simulator) *WSHandler {
	h := &WSHandler{
		sim:     sim,
		clients: make(map[*wsConn]struct{}),
	}
	return h
}

// ServeHTTP implements http.Handler for WebSocket connections.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}

	c := &wsConn{
		conn: conn,
		send: make(chan simulator.PlayerAction, 256),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	go h.writer(c)
	go h.reader(c)
}

// Broadcast starts broadcasting simulator events to all connected clients.
func (h *WSHandler) Broadcast(ctx context.Context) error {
	out := make(chan simulator.PlayerAction, 256)
	errCh := make(chan error, 1)

	go h.sim.Continuous(ctx, out, errCh)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				log.Println("simulator error:", err)
				return err
			}
		case pa := <-out:
			h.broadcastAction(pa)
		}
	}
}

func (h *WSHandler) broadcastAction(pa simulator.PlayerAction) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- pa:
		default:
			// Client buffer full; skip this message.
		}
	}
}

func (h *WSHandler) writer(c *wsConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case pa, ok := <-c.send:
			if !ok {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := wsjson.Write(ctx, c.conn, pa)
			cancel()
			if err != nil {
				log.Println("write message error:", err)
				return
			}

		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				log.Println("ping error:", err)
				return
			}
		}
	}
}

func (h *WSHandler) reader(c *wsConn) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		close(c.send)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	// CloseRead tells the connection we only write and don't read messages.
	// This ensures ping/pong/close frames are still handled.
	ctx := c.conn.CloseRead(context.Background())
	<-ctx.Done()
}
