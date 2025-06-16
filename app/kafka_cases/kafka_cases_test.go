package kafkaCases

import (
	"testing"
)

func TestConsumeCase(t *testing.T) {
	url, topic, groupId, c := "192.168.0.62:9092", "my_topic", "groupId", 1
	DoConsumeCase(url, topic, groupId, c)
}

func TestProduceCase(t *testing.T) {
	url, topic, acks := "192.168.0.62:9092", "my_topic", 0
	InitWriterConfig(url, topic, acks)
	DoProduceCase("console", 100, 3, 3)
}
