package kafka_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jimtang2/nux/lib/client/kafka"
	"github.com/spf13/viper"
)

func loadEnv() {
	viper.SetEnvPrefix("nux")
	viper.AutomaticEnv()
}

func TestKafka_Connect_UsingEnv(t *testing.T) {
	loadEnv()

	c := &kafka.Kafka{}
	err := c.Connect(
		viper.GetString("kafka_brokers"),
		viper.GetString("kafka_credentials"),
	)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if c.Client == nil {
		t.Fatal("Connect() did not initialize client")
	}
}

func TestKafka_ProduceConsume(t *testing.T) {
	loadEnv()

	brokers := viper.GetString("kafka_brokers")
	if brokers == "" {
		t.Skip("NUX_KAFKA_BROKERS not set, skipping test")
	}

	credentials := viper.GetString("kafka_credentials")

	c := &kafka.Kafka{}
	if err := c.Connect(brokers, credentials); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer c.Close()

	if err := c.InitAdmin(); err != nil {
		t.Fatalf("InitAdmin() error = %v", err)
	}

	if err := c.InitProducer(); err != nil {
		t.Fatalf("InitProducer() error = %v", err)
	}

	if err := c.InitConsumer(); err != nil {
		t.Fatalf("InitConsumer() error = %v", err)
	}

	topic := fmt.Sprintf("nux-test-%d", time.Now().UnixNano())

	if err := c.CreateTopic(topic, 1, 1); err != nil {
		t.Fatalf("CreateTopic() error = %v", err)
	}

	count := 2000

	messages := make([]string, count)
	for i := 0; i < count; i++ {
		messages[i] = fmt.Sprintf("msg-%d", i)
	}

	if err := c.Produce(topic, messages); err != nil {
		t.Fatalf("Produce() error = %v", err)
	}

	consumed, err := c.Consume(topic, count)
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	if len(consumed) != count {
		t.Fatalf("expected %d messages, got %d", count, len(consumed))
	}

	for i, msg := range consumed {
		expected := fmt.Sprintf("msg-%d", i)
		if msg != expected {
			t.Errorf("message %d: expected %q, got %q", i, expected, msg)
		}
	}
}
