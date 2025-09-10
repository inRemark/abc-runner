package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"abc-runner/app/core/utils"
	"gopkg.in/yaml.v2"
)

// KafkaConfigSource Kafka配置源接口
type KafkaConfigSource interface {
	Load() (*KafkaAdapterConfig, error)
	CanLoad() bool
	Priority() int
}

// KafkaYAMLConfigSource Kafka YAML文件配置源
type KafkaYAMLConfigSource struct {
	FilePath string
}

// NewKafkaYAMLConfigSource 创建Kafka YAML配置源
func NewKafkaYAMLConfigSource(filePath string) *KafkaYAMLConfigSource {
	return &KafkaYAMLConfigSource{FilePath: filePath}
}

// Load 加载配置
func (y *KafkaYAMLConfigSource) Load() (*KafkaAdapterConfig, error) {
	data, err := os.ReadFile(y.FilePath)
	if err != nil {
		return nil, err
	}

	var configWrapper struct {
		Kafka *KafkaAdapterConfig `yaml:"kafka"`
	}

	err = yaml.Unmarshal(data, &configWrapper)
	if err != nil {
		return nil, err
	}

	if configWrapper.Kafka == nil {
		return nil, fmt.Errorf("invalid kafka config file format")
	}

	return configWrapper.Kafka, nil
}

// CanLoad 检查是否可以加载
func (y *KafkaYAMLConfigSource) CanLoad() bool {
	_, err := os.Stat(y.FilePath)
	return err == nil
}

// Priority 获取优先级
func (y *KafkaYAMLConfigSource) Priority() int {
	return 50
}

// KafkaEnvConfigSource Kafka环境变量配置源
type KafkaEnvConfigSource struct {
	Prefix string
}

// NewKafkaEnvConfigSource 创建Kafka环境变量配置源
func NewKafkaEnvConfigSource(prefix string) *KafkaEnvConfigSource {
	if prefix == "" {
		prefix = "KAFKA_RUNNER"
	}
	return &KafkaEnvConfigSource{Prefix: prefix}
}

// Load 从环境变量加载配置
func (e *KafkaEnvConfigSource) Load() (*KafkaAdapterConfig, error) {
	// 使用配置加载器获取默认配置
	loader := NewConfigLoader()
	config := loader.LoadFromDefault()

	// 从环境变量加载配置项
	if brokers := os.Getenv(e.Prefix + "_BROKERS"); brokers != "" {
		config.Brokers = strings.Split(brokers, ",")
	}

	if clientID := os.Getenv(e.Prefix + "_CLIENT_ID"); clientID != "" {
		config.ClientID = clientID
	}

	if total := os.Getenv(e.Prefix + "_TOTAL"); total != "" {
		if val, err := parseInt(total); err == nil {
			config.Benchmark.Total = val
		}
	}

	if parallels := os.Getenv(e.Prefix + "_PARALLELS"); parallels != "" {
		if val, err := parseInt(parallels); err == nil {
			config.Benchmark.Parallels = val
		}
	}

	if topic := os.Getenv(e.Prefix + "_TOPIC"); topic != "" {
		config.Benchmark.DefaultTopic = topic
	}

	return config, nil
}

// CanLoad 检查是否可以加载
func (e *KafkaEnvConfigSource) CanLoad() bool {
	// 检查是否有任何相关的环境变量
	envVars := []string{
		e.Prefix + "_BROKERS",
		e.Prefix + "_CLIENT_ID",
		e.Prefix + "_TOTAL",
		e.Prefix + "_PARALLELS",
		e.Prefix + "_TOPIC",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// Priority 获取优先级
func (e *KafkaEnvConfigSource) Priority() int {
	return 80
}

// KafkaArgConfigSource Kafka命令行参数配置源
type KafkaArgConfigSource struct {
	Args []string
}

// NewKafkaArgConfigSource 创建Kafka命令行参数配置源
func NewKafkaArgConfigSource(args []string) *KafkaArgConfigSource {
	return &KafkaArgConfigSource{Args: args}
}

// Load 从命令行参数加载配置
func (a *KafkaArgConfigSource) Load() (*KafkaAdapterConfig, error) {
	// 使用配置加载器获取默认配置
	loader := NewConfigLoader()
	config := loader.LoadFromDefault()

	// 从命令行参数解析配置
	for i := 0; i < len(a.Args); i++ {
		switch a.Args[i] {
		case "--brokers", "-b":
			if i+1 < len(a.Args) {
				config.Brokers = strings.Split(a.Args[i+1], ",")
				i++
			}
		case "--client-id":
			if i+1 < len(a.Args) {
				config.ClientID = a.Args[i+1]
				i++
			}
		case "--total", "-n":
			if i+1 < len(a.Args) {
				if val, err := parseInt(a.Args[i+1]); err == nil {
					config.Benchmark.Total = val
				}
				i++
			}
		case "--parallels", "-c":
			if i+1 < len(a.Args) {
				if val, err := parseInt(a.Args[i+1]); err == nil {
					config.Benchmark.Parallels = val
				}
				i++
			}
		case "--topic", "-t":
			if i+1 < len(a.Args) {
				config.Benchmark.DefaultTopic = a.Args[i+1]
				i++
			}
		}
	}

	return config, nil
}

// CanLoad 检查是否可以加载
func (a *KafkaArgConfigSource) CanLoad() bool {
	return len(a.Args) > 0
}

// Priority 获取优先级
func (a *KafkaArgConfigSource) Priority() int {
	return 100
}

// KafkaDefaultConfigSource Kafka默认配置源
type KafkaDefaultConfigSource struct{}

// NewKafkaDefaultConfigSource 创建Kafka默认配置源
func NewKafkaDefaultConfigSource() *KafkaDefaultConfigSource {
	return &KafkaDefaultConfigSource{}
}

// Load 加载默认配置
func (d *KafkaDefaultConfigSource) Load() (*KafkaAdapterConfig, error) {
	loader := NewConfigLoader()
	return loader.LoadFromDefault(), nil
}

// CanLoad 检查是否可以加载
func (d *KafkaDefaultConfigSource) CanLoad() bool {
	return true
}

// Priority 获取优先级
func (d *KafkaDefaultConfigSource) Priority() int {
	return 1
}

// CreateKafkaConfigSources 创建Kafka配置源列表
func CreateKafkaConfigSources(configFile string, args []string) []KafkaConfigSource {
	sources := make([]KafkaConfigSource, 0)

	// 1. 默认配置源（最低优先级）
	sources = append(sources, NewKafkaDefaultConfigSource())

	// 2. YAML配置文件
	if configFile != "" {
		sources = append(sources, NewKafkaYAMLConfigSource(configFile))
	} else {
		// 使用统一的配置文件查找机制
		foundPath := utils.FindConfigFile("kafka")
		if foundPath != "" {
			sources = append(sources, NewKafkaYAMLConfigSource(foundPath))
		}
	}

	// 3. 环境变量配置源
	sources = append(sources, NewKafkaEnvConfigSource("KAFKA_RUNNER"))

	// 4. 命令行参数配置源（最高优先级）
	if len(args) > 0 {
		sources = append(sources, NewKafkaArgConfigSource(args))
	}

	return sources
}

// parseInt 解析整数，忽略错误
func parseInt(s string) (int, error) {
	// 移除可能的空格
	s = strings.TrimSpace(s)
	
	// 解析整数
	return strconv.Atoi(s)
}