package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpYAMLConfigSource(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "http_config_test_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入测试配置
	configContent := `
http:
  protocol: http
  connection:
    base_url: "http://test.example.com"
    timeout: 10s
  benchmark:
    total: 100
    parallels: 5
`
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// 测试YAML配置源
	source := NewHttpYAMLConfigSource(tmpFile.Name())
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, "http://test.example.com", config.Connection.BaseURL)
	assert.Equal(t, 100, config.Benchmark.Total)
	assert.Equal(t, 5, config.Benchmark.Parallels)
}

func TestHttpDefaultConfigSource(t *testing.T) {
	source := NewHttpDefaultConfigSource()
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", config.Connection.BaseURL)
	assert.Equal(t, 1000, config.Benchmark.Total)
	assert.Equal(t, 10, config.Benchmark.Parallels)
}

func TestHttpEnvConfigSource(t *testing.T) {
	// 设置环境变量
	os.Setenv("HTTP_RUNNER_BASE_URL", "http://env.example.com")
	os.Setenv("HTTP_RUNNER_TIMEOUT", "5s")
	os.Setenv("HTTP_RUNNER_TOTAL", "200")
	defer func() {
		os.Unsetenv("HTTP_RUNNER_BASE_URL")
		os.Unsetenv("HTTP_RUNNER_TIMEOUT")
		os.Unsetenv("HTTP_RUNNER_TOTAL")
	}()

	source := NewHttpEnvConfigSource("HTTP_RUNNER")
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, "http://env.example.com", config.Connection.BaseURL)
	assert.Equal(t, 200, config.Benchmark.Total)
}

func TestHttpArgConfigSource(t *testing.T) {
	args := []string{
		"--url", "http://arg.example.com",
		"--total", "300",
		"--parallels", "15",
	}

	source := NewHttpArgConfigSource(args)
	assert.True(t, source.CanLoad())

	config, err := source.Load()
	require.NoError(t, err)
	assert.Equal(t, "http://arg.example.com", config.Connection.BaseURL)
	assert.Equal(t, 300, config.Benchmark.Total)
	assert.Equal(t, 15, config.Benchmark.Parallels)
}

func TestCreateHttpConfigSources(t *testing.T) {
	// 测试创建配置源列表
	sources := CreateHttpConfigSources("", []string{})
	// 至少应该有3个配置源：
	// 默认配置源, 环境变量配置源, 命令行参数配置源
	// 可能还有YAML配置源（如果系统找到了配置文件）
	assert.GreaterOrEqual(t, len(sources), 3)

	// 测试指定配置文件
	sources = CreateHttpConfigSources("/path/to/config.yaml", []string{})
	// 指定了配置文件路径，应该有3个配置源：
	// 默认配置源, YAML配置源, 环境变量配置源
	// 命令行参数配置源不会添加，因为args为空
	assert.Equal(t, 3, len(sources))

	// 测试指定配置文件和命令行参数
	sources = CreateHttpConfigSources("/path/to/config.yaml", []string{"--total", "100"})
	// 指定了配置文件路径和命令行参数，应该有4个配置源：
	// 默认配置源, YAML配置源, 环境变量配置源, 命令行参数配置源
	assert.Equal(t, 4, len(sources))
}

func TestHttpMultiSourceConfigLoader(t *testing.T) {
	// 创建配置源
	defaultSource := NewHttpDefaultConfigSource()
	argSource := NewHttpArgConfigSource([]string{"--total", "500"})

	// 创建加载器
	loader := NewHttpMultiSourceConfigLoader(defaultSource, argSource)

	// 测试加载配置
	config, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, 500, config.Benchmark.Total) // 命令行参数配置源优先级更高
}
