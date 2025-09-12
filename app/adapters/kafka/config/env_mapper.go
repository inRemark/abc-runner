package config

import (
	"os"
	"strconv"
	"strings"

	"abc-runner/app/core/interfaces"
)

// KafkaEnvVarMapper Kafka环境变量映射器
type KafkaEnvVarMapper struct {
	prefix string
}

// NewKafkaEnvVarMapper 创建Kafka环境变量映射器
func NewKafkaEnvVarMapper(prefix string) *KafkaEnvVarMapper {
	if prefix == "" {
		prefix = "KAFKA_RUNNER"
	}
	return &KafkaEnvVarMapper{prefix: prefix}
}

// MapEnvVarsToConfig 将环境变量映射到配置
func (k *KafkaEnvVarMapper) MapEnvVarsToConfig(config interfaces.Config) error {
	kafkaConfig, ok := config.(*KafkaAdapterConfig)
	if !ok {
		return nil // Not a Kafka config, nothing to do
	}

	// 从环境变量加载配置项
	if brokers := os.Getenv(k.prefix + "_BROKERS"); brokers != "" {
		kafkaConfig.Brokers = strings.Split(brokers, ",")
	}

	if clientID := os.Getenv(k.prefix + "_CLIENT_ID"); clientID != "" {
		kafkaConfig.ClientID = clientID
	}

	if total := os.Getenv(k.prefix + "_TOTAL"); total != "" {
		if val, err := parseInt(total); err == nil {
			kafkaConfig.Benchmark.Total = val
		}
	}

	if parallels := os.Getenv(k.prefix + "_PARALLELS"); parallels != "" {
		if val, err := parseInt(parallels); err == nil {
			kafkaConfig.Benchmark.Parallels = val
		}
	}

	if topic := os.Getenv(k.prefix + "_TOPIC"); topic != "" {
		kafkaConfig.Benchmark.DefaultTopic = topic
	}

	return nil
}

// HasRelevantEnvVars 检查是否有相关的环境变量
func (k *KafkaEnvVarMapper) HasRelevantEnvVars() bool {
	// 检查是否有任何相关的环境变量
	envVars := []string{
		k.prefix + "_BROKERS",
		k.prefix + "_CLIENT_ID",
		k.prefix + "_TOTAL",
		k.prefix + "_PARALLELS",
		k.prefix + "_TOPIC",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// parseInt 解析整数，忽略错误
func parseInt(s string) (int, error) {
	// 移除可能的空格
	s = strings.TrimSpace(s)
	
	// 解析整数
	return strconv.Atoi(s)
}