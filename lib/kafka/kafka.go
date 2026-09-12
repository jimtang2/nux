package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type Kafka struct {
	Client      sarama.Client
	Producer    sarama.SyncProducer
	Consumer    sarama.Consumer
	AdminClient sarama.ClusterAdmin
}

func (c *Kafka) Connect(brokersStr, credentials string) error {
	if brokersStr == "" {
		return fmt.Errorf("kafka brokers are empty")
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
		if brokers[i] == "" {
			return fmt.Errorf("kafka broker at position %d is empty", i)
		}
	}

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Net.DialTimeout = 5 * time.Second
	cfg.Net.ReadTimeout = 5 * time.Second
	cfg.Net.WriteTimeout = 5 * time.Second
	cfg.Metadata.Retry.Max = 3
	cfg.Metadata.Retry.Backoff = 500 * time.Millisecond

	if credentials != "" {
		parts := strings.SplitN(credentials, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid kafka credentials format, expected user:password")
		}

		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Handshake = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = parts[0]
		cfg.Net.SASL.Password = parts[1]
	}

	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return err
	}

	if _, err := client.Controller(); err != nil {
		_ = client.Close()
		return err
	}

	c.Client = client
	return nil
}

func (c *Kafka) InitAdmin() error {
	if c.Client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	admin, err := sarama.NewClusterAdminFromClient(c.Client)
	if err != nil {
		return err
	}

	c.AdminClient = admin
	return nil
}

func (c *Kafka) InitProducer() error {
	if c.Client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	cfg := c.Client.Config()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducerFromClient(c.Client)
	if err != nil {
		return err
	}

	c.Producer = producer
	return nil
}

func (c *Kafka) InitConsumer() error {
	if c.Client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	consumer, err := sarama.NewConsumerFromClient(c.Client)
	if err != nil {
		return err
	}

	c.Consumer = consumer
	return nil
}

func (c *Kafka) CreateTopic(name string, numPartitions int32, replicationFactor int16) error {
	if c.AdminClient == nil {
		return fmt.Errorf("kafka admin client not initialized")
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}

	return c.AdminClient.CreateTopic(name, topicDetail, false)
}

func (c *Kafka) Produce(topic string, messages []string) error {
	if c.Producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	for i, msg := range messages {
		_, _, err := c.Producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(fmt.Sprintf("key-%d", i)),
			Value: sarama.StringEncoder(msg),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Kafka) Consume(topic string, count int) ([]string, error) {
	if c.Consumer == nil {
		return nil, fmt.Errorf("kafka consumer not initialized")
	}

	partitions, err := c.Consumer.Partitions(topic)
	if err != nil {
		return nil, err
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("no partitions found for topic %s", topic)
	}

	messages := make([]string, 0, count)
	pc, err := c.Consumer.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	timeout := time.After(10 * time.Second)
	for len(messages) < count {
		select {
		case msg := <-pc.Messages():
			messages = append(messages, string(msg.Value))
		case <-timeout:
			return messages, fmt.Errorf("timeout waiting for messages, got %d/%d", len(messages), count)
		}
	}

	return messages, nil
}

func (c *Kafka) Close() error {
	var errs []error

	if c.Producer != nil {
		if err := c.Producer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.Consumer != nil {
		if err := c.Consumer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.AdminClient != nil {
		c.AdminClient.Close()
	}

	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// ClearTopic deletes all records from the given topic by calling
// DeleteRecords up to the latest offset on each partition.
func (c *Kafka) ClearTopic(topic string) error {
	if c.AdminClient == nil {
		return fmt.Errorf("kafka admin client not initialized")
	}

	if c.Client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	partitions, err := c.Client.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
	}

	if len(partitions) == 0 {
		return fmt.Errorf("no partitions found for topic %s", topic)
	}

	for _, p := range partitions {
		// Get the latest offset (high watermark).
		latest, err := c.Client.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to get latest offset for %s:%d: %w", topic, p, err)
		}

		// Delete records up to (but not including) the latest offset.
		partitionOffsets := map[int32]int64{
			p: latest,
		}

		if err := c.AdminClient.DeleteRecords(topic, partitionOffsets); err != nil {
			return fmt.Errorf("failed to delete records for %s:%d: %w", topic, p, err)
		}
	}

	return nil
}
