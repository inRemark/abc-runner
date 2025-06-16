package kafkaCases

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

var reader *kafka.Reader

func InitReaderCommand(address, topic, groupId string) *kafka.Reader {
	return initReader(address, topic, groupId)
}
func InitReaderConfig(url, topic, groupId string) *kafka.Reader {
	return initReader(url, topic, groupId)
}
func initReader(address, topic, groupId string) *kafka.Reader {
	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:           []string{address},
		Topic:             topic,
		GroupID:           groupId,
		MinBytes:          1,                 // Default: 1
		MaxBytes:          10240,             // Default: 1MB
		MaxWait:           10 * time.Second,  // Default: 10s
		QueueCapacity:     100,               // defaults to 100
		ReadBatchTimeout:  10 * time.Second,  // Default: 10s
		StartOffset:       kafka.FirstOffset, // Default: FirstOffset
		RebalanceTimeout:  30 * time.Second,  // Default: 30s
		HeartbeatInterval: 3 * time.Second,   // Default: 3s
		CommitInterval:    0 * time.Second,   // Default: 0
		SessionTimeout:    30 * time.Second,  // Default: 30s
		MaxAttempts:       3,                 // default 3
	})
	return reader
}

func Consume(rdr *kafka.Reader, index int) {
	defer func() {
		if err := rdr.Close(); err != nil {
			log.Printf("failed to close reader: %v", err)
		}
	}()
	log.Printf("Starting Kafka consumer %d...\n", index)

	for {
		msg, err := rdr.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("error reading message: %v", err)
		}
		fmt.Printf("Received via reader= %d, key=%s, value=%s, partition=%d, offset=%d\n", index,
			string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
		time.Sleep(100 * time.Millisecond)
	}
}
