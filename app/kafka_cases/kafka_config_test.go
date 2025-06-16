package kafkaCases

import (
	"fmt"
	"testing"
)

func TestKafkaConfig(t *testing.T) {
	config, err := LoadKafkaConfig("../../conf/kafka.yaml")
	if err != nil {
		t.Fatalf("Failed to load Kafka config: %v", err)
	}
	// 打印部分配置，验证是否正确加载
	fmt.Printf("Brokers: %v\n", config.Brokers)
	fmt.Printf("Topic: %s\n", config.Topic)
	fmt.Printf("Producer: %v\n", config.Producer)
	fmt.Printf("Consumer: %v\n", config.Consumer)
	fmt.Printf("TSL: %v\n", config.TLS)
	fmt.Printf("SASL: %v\n", config.SASL)
}
