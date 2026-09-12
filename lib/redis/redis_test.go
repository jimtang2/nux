package redis_test

import (
	"testing"

	"github.com/jimtang2/nux/lib/redis"
	"github.com/spf13/viper"
)

func loadEnv() {
	viper.SetEnvPrefix("nux")
	viper.AutomaticEnv()
}

func TestRedis_Connect_UsingEnv(t *testing.T) {
	loadEnv()

	c := &redis.Redis{}
	if err := c.Connect(viper.GetString("redis_url")); err != nil {
		t.Fatalf("failed to connect redis: %v", err)
	}
	if c.Client == nil {
		t.Fatal("redis client not initialized")
	}
}
