package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaYAMLConfigSource(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "kafka_config_test_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入测试配置
	configContent := `
protocol: kafka
brokers:
  - "localhost:9092"
  - "localhost:9093"
client_id: "test-client"
benchmark:
  total: 100
  parallels: 5
  default_topic: "test-topic"
`
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// 测试YAML配置源
	source := NewKafkaYAMLConfigSource(tmpFile.Name())
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:9092", "localhost:9093"}, config.Brokers)
	assert.Equal(t, "test-client", config.ClientID)
	assert.Equal(t, 100, config.Benchmark.Total)
	assert.Equal(t, 5, config.Benchmark.Parallels)
	assert.Equal(t, "test-topic", config.Benchmark.DefaultTopic)
}

func TestKafkaDefaultConfigSource(t *testing.T) {
	source := NewKafkaDefaultConfigSource()
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:9092"}, config.Brokers)
	assert.Equal(t, "abc-runner-kafka-client", config.ClientID)
	assert.Equal(t, 100000, config.Benchmark.Total)
	assert.Equal(t, 50, config.Benchmark.Parallels)
}

func TestKafkaEnvConfigSource(t *testing.T) {
	// 设置环境变量
	os.Setenv("KAFKA_RUNNER_BROKERS", "localhost:9094,localhost:9095")
	os.Setenv("KAFKA_RUNNER_CLIENT_ID", "env-client")
	os.Setenv("KAFKA_RUNNER_TOTAL", "200")
	defer func() {
		os.Unsetenv("KAFKA_RUNNER_BROKERS")
		os.Unsetenv("KAFKA_RUNNER_CLIENT_ID")
		os.Unsetenv("KAFKA_RUNNER_TOTAL")
	}()

	source := NewKafkaEnvConfigSource("KAFKA_RUNNER")
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:9094", "localhost:9095"}, config.Brokers)
	assert.Equal(t, "env-client", config.ClientID)
	assert.Equal(t, 200, config.Benchmark.Total)
}

func TestKafkaArgConfigSource(t *testing.T) {
	args := []string{
		"--brokers", "localhost:9096,localhost:9097",
		"--client-id", "arg-client",
		"--total", "300",
		"--parallels", "15",
		"--topic", "arg-topic",
	}

	source := NewKafkaArgConfigSource(args)
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:9096", "localhost:9097"}, config.Brokers)
	assert.Equal(t, "arg-client", config.ClientID)
	assert.Equal(t, 300, config.Benchmark.Total)
	assert.Equal(t, 15, config.Benchmark.Parallels)
	assert.Equal(t, "arg-topic", config.Benchmark.DefaultTopic)
}

func TestCreateKafkaConfigSources(t *testing.T) {
	// 测试创建配置源列表
	sources := CreateKafkaConfigSources("", []string{})
	assert.Equal(t, 4, len(sources)) // 默认配置源, YAML配置源, 环境变量配置源, 命令行参数配置源

	// 测试指定配置文件
	sources = CreateKafkaConfigSources("/path/to/config.yaml", []string{})
	assert.Equal(t, 4, len(sources)) // 默认配置源, YAML配置源, 环境变量配置源, 命令行参数配置源
}

func TestKafkaMultiSourceConfigLoader(t *testing.T) {
	// 创建配置源
	defaultSource := NewKafkaDefaultConfigSource()
	argSource := NewKafkaArgConfigSource([]string{"--total", "500"})

	// 创建加载器
	loader := NewKafkaMultiSourceConfigLoader(defaultSource, argSource)

	// 测试加载配置
	config, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, 500, config.Benchmark.Total) // 命令行参数配置源优先级更高
}