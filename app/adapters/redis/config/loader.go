package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// MultiSourceConfigLoader 多源配置加载器
type MultiSourceConfigLoader struct {
	sources []ConfigSource
}

// NewMultiSourceConfigLoader 创建多源配置加载器
func NewMultiSourceConfigLoader(sources ...ConfigSource) *MultiSourceConfigLoader {
	return &MultiSourceConfigLoader{sources: sources}
}

// AddSource 添加配置源
func (m *MultiSourceConfigLoader) AddSource(source ConfigSource) {
	m.sources = append(m.sources, source)
}

// GetSources 获取配置源列表
func (m *MultiSourceConfigLoader) GetSources() []ConfigSource {
	return m.sources
}

// Load 从多个源加载配置，按优先级排序
func (m *MultiSourceConfigLoader) Load() (*RedisConfig, error) {
	// 按优先级排序（高优先级在前）
	sources := make([]ConfigSource, len(m.sources))
	copy(sources, m.sources)
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() > sources[j].Priority()
	})

	// 尝试从每个源加载配置
	for _, source := range sources {
		if source.CanLoad() {
			config, err := source.Load()
			if err != nil {
				continue
			}
			if config != nil {
				return config, nil
			}
		}
	}

	// 如果所有源都失败，返回默认配置
	return NewDefaultRedisConfig(), nil
}

// YAMLConfigSource YAML文件配置源
type YAMLConfigSource struct {
	FilePath string
}

// NewYAMLConfigSource 创建YAML配置源
func NewYAMLConfigSource(filePath string) *YAMLConfigSource {
	return &YAMLConfigSource{FilePath: filePath}
}

// Load 加载配置
func (y *YAMLConfigSource) Load() (*RedisConfig, error) {
	data, err := os.ReadFile(y.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", y.FilePath, err)
	}

	var configWrapper struct {
		Redis RedisConfig `yaml:"redis"`
	}

	err = yaml.Unmarshal(data, &configWrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", y.FilePath, err)
	}

	config := &configWrapper.Redis
	if config.Protocol == "" {
		config.Protocol = "redis"
	}

	return config, nil
}

// CanLoad 检查是否可以加载
func (y *YAMLConfigSource) CanLoad() bool {
	_, err := os.Stat(y.FilePath)
	return err == nil
}

// Priority 获取优先级（YAML文件优先级中等）
func (y *YAMLConfigSource) Priority() int {
	return 50
}

// EnvConfigSource 环境变量配置源
type EnvConfigSource struct {
	Prefix string
}

// NewEnvConfigSource 创建环境变量配置源
func NewEnvConfigSource(prefix string) *EnvConfigSource {
	if prefix == "" {
		prefix = "REDIS_RUNNER"
	}
	return &EnvConfigSource{Prefix: prefix}
}

// Load 从环境变量加载配置
func (e *EnvConfigSource) Load() (*RedisConfig, error) {
	config := NewDefaultRedisConfig()

	// 基本配置
	if protocol := os.Getenv(e.Prefix + "_PROTOCOL"); protocol != "" {
		config.Protocol = protocol
	}

	if mode := os.Getenv(e.Prefix + "_MODE"); mode != "" {
		config.Mode = mode
	}

	// 基准测试配置
	if total := os.Getenv(e.Prefix + "_TOTAL"); total != "" {
		if val, err := strconv.Atoi(total); err == nil {
			config.BenchMark.Total = val
		}
	}

	if parallels := os.Getenv(e.Prefix + "_PARALLELS"); parallels != "" {
		if val, err := strconv.Atoi(parallels); err == nil {
			config.BenchMark.Parallels = val
		}
	}

	if dataSize := os.Getenv(e.Prefix + "_DATA_SIZE"); dataSize != "" {
		if val, err := strconv.Atoi(dataSize); err == nil {
			config.BenchMark.DataSize = val
		}
	}

	if ttl := os.Getenv(e.Prefix + "_TTL"); ttl != "" {
		if val, err := strconv.Atoi(ttl); err == nil {
			config.BenchMark.TTL = val
		}
	}

	if readPercent := os.Getenv(e.Prefix + "_READ_PERCENT"); readPercent != "" {
		if val, err := strconv.Atoi(readPercent); err == nil {
			config.BenchMark.ReadPercent = val
		}
	}

	if randomKeys := os.Getenv(e.Prefix + "_RANDOM_KEYS"); randomKeys != "" {
		if val, err := strconv.Atoi(randomKeys); err == nil {
			config.BenchMark.RandomKeys = val
		}
	}

	if testCase := os.Getenv(e.Prefix + "_CASE"); testCase != "" {
		config.BenchMark.Case = testCase
	}

	// 连接池配置
	if poolSize := os.Getenv(e.Prefix + "_POOL_SIZE"); poolSize != "" {
		if val, err := strconv.Atoi(poolSize); err == nil {
			config.Pool.PoolSize = val
		}
	}

	if minIdle := os.Getenv(e.Prefix + "_MIN_IDLE"); minIdle != "" {
		if val, err := strconv.Atoi(minIdle); err == nil {
			config.Pool.MinIdle = val
		}
	}

	if maxIdle := os.Getenv(e.Prefix + "_MAX_IDLE"); maxIdle != "" {
		if val, err := strconv.Atoi(maxIdle); err == nil {
			config.Pool.MaxIdle = val
		}
	}

	// 单机配置
	if addr := os.Getenv(e.Prefix + "_ADDR"); addr != "" {
		config.Standalone.Addr = addr
	}

	if password := os.Getenv(e.Prefix + "_PASSWORD"); password != "" {
		config.Standalone.Password = password
	}

	if db := os.Getenv(e.Prefix + "_DB"); db != "" {
		if val, err := strconv.Atoi(db); err == nil {
			config.Standalone.Db = val
		}
	}

	// 哨兵配置
	if masterName := os.Getenv(e.Prefix + "_SENTINEL_MASTER_NAME"); masterName != "" {
		config.Sentinel.MasterName = masterName
	}

	if sentinelAddrs := os.Getenv(e.Prefix + "_SENTINEL_ADDRS"); sentinelAddrs != "" {
		config.Sentinel.Addrs = strings.Split(sentinelAddrs, ",")
	}

	if sentinelPassword := os.Getenv(e.Prefix + "_SENTINEL_PASSWORD"); sentinelPassword != "" {
		config.Sentinel.Password = sentinelPassword
	}

	if sentinelDb := os.Getenv(e.Prefix + "_SENTINEL_DB"); sentinelDb != "" {
		if val, err := strconv.Atoi(sentinelDb); err == nil {
			config.Sentinel.Db = val
		}
	}

	// 集群配置
	if clusterAddrs := os.Getenv(e.Prefix + "_CLUSTER_ADDRS"); clusterAddrs != "" {
		config.Cluster.Addrs = strings.Split(clusterAddrs, ",")
	}

	if clusterPassword := os.Getenv(e.Prefix + "_CLUSTER_PASSWORD"); clusterPassword != "" {
		config.Cluster.Password = clusterPassword
	}

	return config, nil
}

// CanLoad 检查是否可以加载（环境变量始终可用）
func (e *EnvConfigSource) CanLoad() bool {
	// 检查是否有任何相关的环境变量
	envVars := []string{
		e.Prefix + "_PROTOCOL",
		e.Prefix + "_MODE",
		e.Prefix + "_ADDR",
		e.Prefix + "_TOTAL",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// Priority 获取优先级（环境变量优先级较高）
func (e *EnvConfigSource) Priority() int {
	return 80
}

// ArgConfigSource 命令行参数配置源
type ArgConfigSource struct {
	Args map[string]string
}

// NewArgConfigSource 创建命令行参数配置源
func NewArgConfigSource(args map[string]string) *ArgConfigSource {
	return &ArgConfigSource{Args: args}
}

// Load 从命令行参数加载配置
func (a *ArgConfigSource) Load() (*RedisConfig, error) {
	config := NewDefaultRedisConfig()

	if protocol, exists := a.Args["protocol"]; exists {
		config.Protocol = protocol
	}

	if mode, exists := a.Args["mode"]; exists {
		config.Mode = mode
	}

	if total, exists := a.Args["total"]; exists {
		if val, err := strconv.Atoi(total); err == nil {
			config.BenchMark.Total = val
		}
	}

	if parallels, exists := a.Args["parallels"]; exists {
		if val, err := strconv.Atoi(parallels); err == nil {
			config.BenchMark.Parallels = val
		}
	}

	if dataSize, exists := a.Args["data_size"]; exists {
		if val, err := strconv.Atoi(dataSize); err == nil {
			config.BenchMark.DataSize = val
		}
	}

	if ttl, exists := a.Args["ttl"]; exists {
		if val, err := strconv.Atoi(ttl); err == nil {
			config.BenchMark.TTL = val
		}
	}

	if readPercent, exists := a.Args["read_percent"]; exists {
		if val, err := strconv.Atoi(readPercent); err == nil {
			config.BenchMark.ReadPercent = val
		}
	}

	if randomKeys, exists := a.Args["random_keys"]; exists {
		if val, err := strconv.Atoi(randomKeys); err == nil {
			config.BenchMark.RandomKeys = val
		}
	}

	if testCase, exists := a.Args["case"]; exists {
		config.BenchMark.Case = testCase
	}

	if addr, exists := a.Args["addr"]; exists {
		config.Standalone.Addr = addr
	}

	if password, exists := a.Args["password"]; exists {
		config.Standalone.Password = password
	}

	if db, exists := a.Args["db"]; exists {
		if val, err := strconv.Atoi(db); err == nil {
			config.Standalone.Db = val
		}
	}

	return config, nil
}

// CanLoad 检查是否可以加载
func (a *ArgConfigSource) CanLoad() bool {
	return len(a.Args) > 0
}

// Priority 获取优先级（命令行参数优先级最高）
func (a *ArgConfigSource) Priority() int {
	return 100
}

// DefaultConfigSource 默认配置源
type DefaultConfigSource struct{}

// NewDefaultConfigSource 创建默认配置源
func NewDefaultConfigSource() *DefaultConfigSource {
	return &DefaultConfigSource{}
}

// Load 加载默认配置
func (d *DefaultConfigSource) Load() (*RedisConfig, error) {
	return NewDefaultRedisConfig(), nil
}

// CanLoad 检查是否可以加载（默认配置始终可用）
func (d *DefaultConfigSource) CanLoad() bool {
	return true
}

// Priority 获取优先级（默认配置优先级最低）
func (d *DefaultConfigSource) Priority() int {
	return 1
}

// CreateStandardLoader 创建标准配置加载器
func CreateStandardLoader(configPath string, args []string) ConfigLoader {
	loader := NewMultiSourceConfigLoader()

	// 添加默认配置源（最低优先级）
	loader.AddSource(NewDefaultConfigSource())

	// 添加YAML文件配置源
	if configPath != "" {
		loader.AddSource(NewYAMLConfigSource(configPath))
	}

	// 尝试查找配置文件
	configPaths := []string{
		"conf/redis.yaml",
		"config/redis.yaml",
		"redis.yaml",
		filepath.Join(os.Getenv("HOME"), ".redis-runner", "redis.yaml"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			loader.AddSource(NewYAMLConfigSource(path))
			break
		}
	}

	// 添加环境变量配置源
	loader.AddSource(NewEnvConfigSource("REDIS_RUNNER"))

	// 添加命令行参数配置源（最高优先级）
	if len(args) > 0 {
		loader.AddSource(NewCommandLineConfigSource(args))
	}

	return loader
}

// LoadConfigWithOverrides 加载配置并应用覆盖
func LoadConfigWithOverrides(configPath string, args []string, overrides map[string]interface{}) (*RedisConfig, error) {
	loader := CreateStandardLoader(configPath, args)
	config, err := loader.Load()
	if err != nil {
		return nil, err
	}

	// 应用覆盖配置
	if len(overrides) > 0 {
		config = applyOverrides(config, overrides)
	}

	return config, nil
}

// applyOverrides 应用覆盖配置
func applyOverrides(config *RedisConfig, overrides map[string]interface{}) *RedisConfig {
	redisConfig := config
	cloned := redisConfig.Clone()

	for key, value := range overrides {
		switch key {
		case "protocol":
			if v, ok := value.(string); ok {
				cloned.Protocol = v
			}
		case "mode":
			if v, ok := value.(string); ok {
				cloned.Mode = v
			}
		case "total":
			if v, ok := value.(int); ok {
				cloned.BenchMark.Total = v
			}
		case "parallels":
			if v, ok := value.(int); ok {
				cloned.BenchMark.Parallels = v
			}
		case "data_size":
			if v, ok := value.(int); ok {
				cloned.BenchMark.DataSize = v
			}
		case "ttl":
			if v, ok := value.(int); ok {
				cloned.BenchMark.TTL = v
			}
		case "read_percent":
			if v, ok := value.(int); ok {
				cloned.BenchMark.ReadPercent = v
			}
		case "random_keys":
			if v, ok := value.(int); ok {
				cloned.BenchMark.RandomKeys = v
			}
		case "case":
			if v, ok := value.(string); ok {
				cloned.BenchMark.Case = v
			}
		case "addr":
			if v, ok := value.(string); ok {
				cloned.Standalone.Addr = v
			}
		case "password":
			if v, ok := value.(string); ok {
				cloned.Standalone.Password = v
			}
		case "db":
			if v, ok := value.(int); ok {
				cloned.Standalone.Db = v
			}
		}
	}

	return cloned
}