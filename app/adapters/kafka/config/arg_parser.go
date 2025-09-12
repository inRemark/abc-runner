package config

import (
	"strings"

	"abc-runner/app/core/interfaces"
)

// KafkaArgParser Kafka命令行参数解析器
type KafkaArgParser struct{}

// NewKafkaArgParser 创建Kafka命令行参数解析器
func NewKafkaArgParser() *KafkaArgParser {
	return &KafkaArgParser{}
}

// ParseArgs 解析命令行参数
func (k *KafkaArgParser) ParseArgs(args []string, config interfaces.Config) error {
	kafkaConfig, ok := config.(*KafkaAdapterConfig)
	if !ok {
		return nil // Not a Kafka config, nothing to do
	}

	// 从命令行参数解析配置
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--brokers", "-b":
			if i+1 < len(args) {
				kafkaConfig.Brokers = strings.Split(args[i+1], ",")
				i++
			}
		case "--client-id":
			if i+1 < len(args) {
				kafkaConfig.ClientID = args[i+1]
				i++
			}
		case "--total", "-n":
			if i+1 < len(args) {
				if val, err := parseInt(args[i+1]); err == nil {
					kafkaConfig.Benchmark.Total = val
				}
				i++
			}
		case "--parallels", "-c":
			if i+1 < len(args) {
				if val, err := parseInt(args[i+1]); err == nil {
					kafkaConfig.Benchmark.Parallels = val
				}
				i++
			}
		case "--topic", "-t":
			if i+1 < len(args) {
				kafkaConfig.Benchmark.DefaultTopic = args[i+1]
				i++
			}
		}
	}

	return nil
}
