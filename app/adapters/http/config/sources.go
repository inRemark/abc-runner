package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"abc-runner/app/core/utils"
	"gopkg.in/yaml.v3"
)

// HttpConfigSource HTTP配置源接口
type HttpConfigSource interface {
	Load() (*HttpAdapterConfig, error)
	CanLoad() bool
	Priority() int
}

// HttpYAMLConfigSource HTTP YAML文件配置源
type HttpYAMLConfigSource struct {
	FilePath string
}

// NewHttpYAMLConfigSource 创建HTTP YAML配置源
func NewHttpYAMLConfigSource(filePath string) *HttpYAMLConfigSource {
	return &HttpYAMLConfigSource{FilePath: filePath}
}

// Load 加载配置
func (y *HttpYAMLConfigSource) Load() (*HttpAdapterConfig, error) {
	data, err := os.ReadFile(y.FilePath)
	if err != nil {
		return nil, err
	}

	var configWrapper struct {
		HTTP *HttpAdapterConfig `yaml:"http"`
	}

	err = yaml.Unmarshal(data, &configWrapper)
	if err != nil {
		return nil, err
	}

	return configWrapper.HTTP, nil
}

// CanLoad 检查是否可以加载
func (y *HttpYAMLConfigSource) CanLoad() bool {
	_, err := os.Stat(y.FilePath)
	return err == nil
}

// Priority 获取优先级
func (y *HttpYAMLConfigSource) Priority() int {
	return 50
}

// HttpEnvConfigSource HTTP环境变量配置源
type HttpEnvConfigSource struct {
	Prefix string
}

// NewHttpEnvConfigSource 创建HTTP环境变量配置源
func NewHttpEnvConfigSource(prefix string) *HttpEnvConfigSource {
	if prefix == "" {
		prefix = "HTTP_RUNNER"
	}
	return &HttpEnvConfigSource{Prefix: prefix}
}

// Load 从环境变量加载配置
func (e *HttpEnvConfigSource) Load() (*HttpAdapterConfig, error) {
	config := LoadDefaultHttpConfig()

	// 从环境变量加载配置项
	if baseURL := os.Getenv(e.Prefix + "_BASE_URL"); baseURL != "" {
		config.Connection.BaseURL = baseURL
	}

	if timeout := os.Getenv(e.Prefix + "_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			config.Connection.Timeout = val
		}
	}

	if maxIdleConns := os.Getenv(e.Prefix + "_MAX_IDLE_CONNS"); maxIdleConns != "" {
		if val, err := parseInt(maxIdleConns); err == nil {
			config.Connection.MaxIdleConns = val
		}
	}

	if maxConnsPerHost := os.Getenv(e.Prefix + "_MAX_CONNS_PER_HOST"); maxConnsPerHost != "" {
		if val, err := parseInt(maxConnsPerHost); err == nil {
			config.Connection.MaxConnsPerHost = val
		}
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

	if method := os.Getenv(e.Prefix + "_METHOD"); method != "" {
		config.Benchmark.Method = method
	}

	if path := os.Getenv(e.Prefix + "_PATH"); path != "" {
		config.Benchmark.Path = path
	}

	return config, nil
}

// CanLoad 检查是否可以加载
func (e *HttpEnvConfigSource) CanLoad() bool {
	// 检查是否有任何相关的环境变量
	envVars := []string{
		e.Prefix + "_BASE_URL",
		e.Prefix + "_TIMEOUT",
		e.Prefix + "_TOTAL",
		e.Prefix + "_PARALLELS",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// Priority 获取优先级
func (e *HttpEnvConfigSource) Priority() int {
	return 80
}

// HttpArgConfigSource HTTP命令行参数配置源
type HttpArgConfigSource struct {
	Args []string
}

// NewHttpArgConfigSource 创建HTTP命令行参数配置源
func NewHttpArgConfigSource(args []string) *HttpArgConfigSource {
	return &HttpArgConfigSource{Args: args}
}

// Load 从命令行参数加载配置
func (a *HttpArgConfigSource) Load() (*HttpAdapterConfig, error) {
	config := LoadDefaultHttpConfig()

	// 从命令行参数解析配置
	for i := 0; i < len(a.Args); i++ {
		switch a.Args[i] {
		case "--url", "-u":
			if i+1 < len(a.Args) {
				config.Connection.BaseURL = a.Args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(a.Args) {
				if t, err := time.ParseDuration(a.Args[i+1]); err == nil {
					config.Connection.Timeout = t
				}
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
		case "--method", "-m":
			if i+1 < len(a.Args) {
				config.Benchmark.Method = a.Args[i+1]
				i++
			}
		case "--path", "-p":
			if i+1 < len(a.Args) {
				config.Benchmark.Path = a.Args[i+1]
				i++
			}
		}
	}

	return config, nil
}

// CanLoad 检查是否可以加载
func (a *HttpArgConfigSource) CanLoad() bool {
	return len(a.Args) > 0
}

// Priority 获取优先级
func (a *HttpArgConfigSource) Priority() int {
	return 100
}

// HttpDefaultConfigSource HTTP默认配置源
type HttpDefaultConfigSource struct{}

// NewHttpDefaultConfigSource 创建HTTP默认配置源
func NewHttpDefaultConfigSource() *HttpDefaultConfigSource {
	return &HttpDefaultConfigSource{}
}

// Load 加载默认配置
func (d *HttpDefaultConfigSource) Load() (*HttpAdapterConfig, error) {
	return LoadDefaultHttpConfig(), nil
}

// CanLoad 检查是否可以加载
func (d *HttpDefaultConfigSource) CanLoad() bool {
	return true
}

// Priority 获取优先级
func (d *HttpDefaultConfigSource) Priority() int {
	return 1
}

// CreateHttpConfigSources 创建HTTP配置源列表
func CreateHttpConfigSources(configFile string, args []string) []HttpConfigSource {
	sources := make([]HttpConfigSource, 0)

	// 1. 默认配置源（最低优先级）
	sources = append(sources, NewHttpDefaultConfigSource())

	// 2. YAML配置文件
	if configFile != "" {
		sources = append(sources, NewHttpYAMLConfigSource(configFile))
	} else {
		// 使用统一的配置文件查找机制
		foundPath := utils.FindConfigFile("http")
		if foundPath != "" {
			sources = append(sources, NewHttpYAMLConfigSource(foundPath))
		}
	}

	// 3. 环境变量配置源
	sources = append(sources, NewHttpEnvConfigSource("HTTP_RUNNER"))

	// 4. 命令行参数配置源（最高优先级）
	if len(args) > 0 {
		sources = append(sources, NewHttpArgConfigSource(args))
	}

	return sources
}

// parseInt 解析整数，忽略错误
func parseInt(s string) (int, error) {
	val, err := parseIntStrict(s)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// parseIntStrict 严格解析整数
func parseIntStrict(s string) (int, error) {
	// 移除可能的空格
	s = strings.TrimSpace(s)
	
	// 解析整数
	return strconv.Atoi(s)
}