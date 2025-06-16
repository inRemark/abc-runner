package kafkaCases

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

var writer *kafka.Writer

func InitWriterCommand(address, topic string, acks int) *kafka.Writer {
	return initWriter(address, topic, acks)
}

func InitWriterConfig(address, topic string, acks int) *kafka.Writer {
	return initWriter(address, topic, acks)
}

func initWriter(address, topic string, acks int) *kafka.Writer {
	writer = &kafka.Writer{
		Addr:         kafka.TCP(address),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequiredAcks(acks),
	}
	return writer
}

func Produce(key, message string) (bool, bool, time.Duration) {
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(message),
		},
	)
	if err != nil {
		log.Fatalf("failed to write messages: %v", err)
	}
	return true, false, time.Since(time.Now())
}
