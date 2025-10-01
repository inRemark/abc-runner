package utils

import (
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	// 保存原始工作目录
	// originalWd, _ := os.Getwd()
	// defer os.Chdir(originalWd)

	// // 创建临时目录结构用于测试
	// tempDir := t.TempDir()
	// configDir := filepath.Join(tempDir, "config")
	// os.Mkdir(configDir, 0755)

	// // 创建测试配置文件
	// redisConfigPath := filepath.Join(configDir, "redis.yaml")
	// httpConfigPath := filepath.Join(configDir, "http.yaml")
	// kafkaConfigPath := filepath.Join(configDir, "kafka.yaml")

	// os.WriteFile(redisConfigPath, []byte("test: redis"), 0644)
	// os.WriteFile(httpConfigPath, []byte("test: http"), 0644)
	// os.WriteFile(kafkaConfigPath, []byte("test: kafka"), 0644)

	// // 更改工作目录到临时目录
	// os.Chdir(tempDir)

	// // 测试Redis配置文件查找
	// redisPath := FindConfigFile("redis")
	// expectedRedisPath := "config/redis.yaml"
	// if redisPath != expectedRedisPath {
	// 	t.Errorf("Expected Redis config path %s, got %s", expectedRedisPath, redisPath)
	// }

	// // 测试HTTP配置文件查找
	// httpPath := FindConfigFile("http")
	// expectedHttpPath := "config/http.yaml"
	// if httpPath != expectedHttpPath {
	// 	t.Errorf("Expected HTTP config path %s, got %s", expectedHttpPath, httpPath)
	// }

	// // 测试Kafka配置文件查找
	// kafkaPath := FindConfigFile("kafka")
	// expectedKafkaPath := "config/kafka.yaml"
	// if kafkaPath != expectedKafkaPath {
	// 	t.Errorf("Expected Kafka config path %s, got %s", expectedKafkaPath, kafkaPath)
	// }

	// 测试未知协议
	// unknownPath := FindConfigFile("unknown")
	// if unknownPath != "" {
	// 	t.Errorf("Expected empty path for unknown protocol, got %s", unknownPath)
	// }
}
