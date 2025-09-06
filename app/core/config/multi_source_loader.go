package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"redis-runner/app/core/interfaces"
	redisconfig "redis-runner/app/adapters/redis/config"
)

// EnvironmentConfigSource 环境变量配置源
type EnvironmentConfigSource struct {
	Prefix string
}

// NewEnvironmentConfigSource 创建环境变量配置源
func NewEnvironmentConfigSource(prefix string) *EnvironmentConfigSource {
	if prefix == "" {
		prefix = "REDIS_RUNNER"
	}
	return &EnvironmentConfigSource{Prefix: prefix}
}

// Load 从环境变量加载配置
func (e *EnvironmentConfigSource) Load() (interfaces.Config, error) {
	config := redisconfig.NewDefaultRedisConfig()
	config.Protocol = "redis"
	config.Mode = e.getEnvString("MODE", "standalone")

	// 基准测试配置
	config.BenchMark = redisconfig.BenchmarkConfigImpl{
		Total:       e.getEnvInt("TOTAL", 100000),
		Parallels:   e.getEnvInt("PARALLELS", 50),
		DataSize:    e.getEnvInt("DATA_SIZE", 3),
		TTL:         e.getEnvInt("TTL", 120),
		ReadPercent: e.getEnvInt("READ_PERCENT", 50),
		RandomKeys:  e.getEnvInt("RANDOM_KEYS", 0),
		Case:        e.getEnvString("CASE", "get"),
	}

	// 连接池配置
	config.Pool = redisconfig.PoolConfigImpl{
		PoolSize: e.getEnvInt("POOL_SIZE", 10),
		MinIdle:  e.getEnvInt("MIN_IDLE", 2),
	}

	// 根据模式设置连接配置
	switch config.Mode {
	case "standalone":
		config.Standalone = redisconfig.StandAloneInfo{
			Addr:     e.getEnvString("ADDR", "127.0.0.1:6379"),
			Password: e.getEnvString("PASSWORD", ""),
			Db:       e.getEnvInt("DB", 0),
		}
	case "cluster":
		addrs := e.getEnvString("ADDRS", "127.0.0.1:6371,127.0.0.1:6372,127.0.0.1:6373")
		config.Cluster = redisconfig.ClusterInfo{
			Addrs:    strings.Split(addrs, ","),
			Password: e.getEnvString("PASSWORD", ""),
		}
	case "sentinel":
		addrs := e.getEnvString("SENTINEL_ADDRS", "127.0.0.1:26371,127.0.0.1:26372,127.0.0.1:26373")
		config.Sentinel = redisconfig.SentinelInfo{
			MasterName: e.getEnvString("MASTER_NAME", "mymaster"),
			Addrs:      strings.Split(addrs, ","),
			Password:   e.getEnvString("PASSWORD", ""),
			Db:         e.getEnvInt("DB", 0),
		}
	}

	return redisconfig.NewRedisConfigAdapter(config), nil
}

// CanLoad 检查是否可以从环境变量加载
func (e *EnvironmentConfigSource) CanLoad() bool {
	// 检查关键环境变量是否存在
	_, exists := os.LookupEnv(e.envKey("MODE"))
	return exists
}

// Priority 获取优先级
func (e *EnvironmentConfigSource) Priority() int {
	return 2 // 环境变量优先级高于文件
}

// getEnvString 获取字符串环境变量
func (e *EnvironmentConfigSource) getEnvString(key, defaultValue string) string {
	if value := os.Getenv(e.envKey(key)); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func (e *EnvironmentConfigSource) getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(e.envKey(key)); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// envKey 构建环境变量键名
func (e *EnvironmentConfigSource) envKey(key string) string {
	return fmt.Sprintf("%s_%s", e.Prefix, key)
}

// CommandLineConfigSource 命令行配置源
type CommandLineConfigSource struct {
	Args []string
}

// NewCommandLineConfigSource 创建命令行配置源
func NewCommandLineConfigSource(args []string) *CommandLineConfigSource {
	return &CommandLineConfigSource{Args: args}
}

// Load 从命令行参数加载配置
func (c *CommandLineConfigSource) Load() (interfaces.Config, error) {
	config := redisconfig.NewDefaultRedisConfig()
	config.Protocol = "redis"
	config.Mode = "standalone"

	// 解析命令行参数
	argMap := c.parseArgs()

	// 解析模式
	if cluster, exists := argMap["cluster"]; exists && cluster == "true" {
		config.Mode = "cluster"
	}

	// 基准测试配置
	config.BenchMark = redisconfig.BenchmarkConfigImpl{
		Total:       c.getIntArg(argMap, "n", 100000),
		Parallels:   c.getIntArg(argMap, "c", 50),
		DataSize:    c.getIntArg(argMap, "d", 3),
		TTL:         c.getIntArg(argMap, "ttl", 120),
		ReadPercent: c.getIntArg(argMap, "R", 50),
		RandomKeys:  c.getIntArg(argMap, "r", 0),
		Case:        c.getStringArg(argMap, "t", "get"),
	}

	// 连接池配置
	config.Pool = redisconfig.PoolConfigImpl{
		PoolSize: 10,
		MinIdle:  2,
	}

	// 连接配置
	host := c.getStringArg(argMap, "h", "127.0.0.1")
	port := c.getIntArg(argMap, "p", 6379)
	password := c.getStringArg(argMap, "a", "")
	db := c.getIntArg(argMap, "db", 0)

	addr := fmt.Sprintf("%s:%d", host, port)

	if config.Mode == "cluster" {
		config.Cluster = redisconfig.ClusterInfo{
			Addrs:    []string{addr},
			Password: password,
		}
	} else {
		config.Standalone = redisconfig.StandAloneInfo{
			Addr:     addr,
			Password: password,
			Db:       db,
		}
	}

	return redisconfig.NewRedisConfigAdapter(config), nil
}

// CanLoad 检查是否可以从命令行加载
func (c *CommandLineConfigSource) CanLoad() bool {
	return len(c.Args) > 0
}

// Priority 获取优先级
func (c *CommandLineConfigSource) Priority() int {
	return 3 // 命令行参数优先级最高
}

// parseArgs 解析命令行参数
func (c *CommandLineConfigSource) parseArgs() map[string]string {
	argMap := make(map[string]string)

	for i := 0; i < len(c.Args); i++ {
		arg := c.Args[i]
		if strings.HasPrefix(arg, "-") {
			key := strings.TrimPrefix(arg, "-")
			key = strings.TrimPrefix(key, "-") // 处理 --flag 格式

			// 布尔标志
			if key == "cluster" || key == "config" {
				argMap[key] = "true"
				continue
			}

			// 键值对
			if i+1 < len(c.Args) && !strings.HasPrefix(c.Args[i+1], "-") {
				argMap[key] = c.Args[i+1]
				i++ // 跳过值
			}
		}
	}

	return argMap
}

// getStringArg 获取字符串参数
func (c *CommandLineConfigSource) getStringArg(argMap map[string]string, key, defaultValue string) string {
	if value, exists := argMap[key]; exists {
		return value
	}
	return defaultValue
}

// getIntArg 获取整数参数
func (c *CommandLineConfigSource) getIntArg(argMap map[string]string, key string, defaultValue int) int {
	if value, exists := argMap[key]; exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules []ValidationRule
}

// ValidationRule 验证规则
type ValidationRule func(interfaces.Config) error

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make([]ValidationRule, 0),
	}

	// 添加默认验证规则
	validator.AddRule(validateProtocol)
	validator.AddRule(validateMode)
	validator.AddRule(validateConnection)
	validator.AddRule(validateBenchmark)

	return validator
}

// AddRule 添加验证规则
func (v *ConfigValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate 验证配置
func (v *ConfigValidator) Validate(config interfaces.Config) error {
	for _, rule := range v.rules {
		if err := rule(config); err != nil {
			return err
		}
	}
	return nil
}

// validateProtocol 验证协议
func validateProtocol(config interfaces.Config) error {
	protocol := config.GetProtocol()
	if protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	supportedProtocols := []string{"redis"}
	for _, supported := range supportedProtocols {
		if protocol == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported protocol: %s", protocol)
}

// validateMode 验证模式
func validateMode(config interfaces.Config) error {
	// 通过适配器提取Redis配置
	if redisConfig, err := redisconfig.ExtractRedisConfig(config); err == nil {
		supportedModes := []string{"standalone", "sentinel", "cluster"}
		for _, mode := range supportedModes {
			if redisConfig.Mode == mode {
				return nil
			}
		}
		return fmt.Errorf("unsupported mode: %s", redisConfig.Mode)
	}
	return nil
}

// validateConnection 验证连接配置
func validateConnection(config interfaces.Config) error {
	connConfig := config.GetConnection()
	if connConfig == nil {
		return fmt.Errorf("connection config cannot be nil")
	}

	addresses := connConfig.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one address must be specified")
	}

	for _, addr := range addresses {
		if addr == "" {
			return fmt.Errorf("address cannot be empty")
		}
	}

	return nil
}

// validateBenchmark 验证基准测试配置
func validateBenchmark(config interfaces.Config) error {
	benchConfig := config.GetBenchmark()
	if benchConfig == nil {
		return fmt.Errorf("benchmark config cannot be nil")
	}

	if benchConfig.GetTotal() <= 0 {
		return fmt.Errorf("total requests must be positive")
	}

	if benchConfig.GetParallels() <= 0 {
		return fmt.Errorf("parallel connections must be positive")
	}

	if benchConfig.GetDataSize() <= 0 {
		return fmt.Errorf("data size must be positive")
	}

	readPercent := benchConfig.GetReadPercent()
	if readPercent < 0 || readPercent > 100 {
		return fmt.Errorf("read percent must be between 0 and 100")
	}

	return nil
}
