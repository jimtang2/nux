package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Pg struct {
	Conn *pgx.Conn
}

func (c *Pg) Connect(connStr string) error {
	if connStr == "" {
		return fmt.Errorf("postgres connection string is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close(context.Background())
		return err
	}

	c.Conn = conn
	return nil
}
