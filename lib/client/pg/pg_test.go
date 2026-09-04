package pg_test

import (
	"testing"

	"github.com/jimtang2/nux/lib/client/pg"
	"github.com/spf13/viper"
)

func loadEnv() {
	viper.SetEnvPrefix("nux")
	viper.AutomaticEnv()
}

func TestPostgres_Connect_UsingEnv(t *testing.T) {
	loadEnv()

	c := &pg.Pg{}
	if err := c.Connect(viper.GetString("postgres_url")); err != nil {
		t.Fatalf("failed to connect pg: %v", err)
	}
	if c.Conn == nil {
		t.Fatal("postgres client not initialized")
	}
}
