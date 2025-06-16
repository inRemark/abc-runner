package kafkaCases

import (
	"testing"
)

func TestProducer(t *testing.T) {
	url, topic, acks := "192.168.0.62:9092", "my_topic", 0
	InitWriterConfig(url, topic, acks)
	Produce("key", "test")
}
