package command

import (
	"context"
	"fmt"
	"log"
	"strings"

	httpCases "redis-runner/app/http_cases"
	kafkaCases "redis-runner/app/kafka_cases"
	redisCases "redis-runner/app/redis_cases"
)

// LegacyCommandWrapper 传统版本命令包装器
type LegacyCommandWrapper struct {
	*BaseCommandHandler
	legacyFunc func([]string)
	protocol   string
}

// NewLegacyCommandWrapper 创建传统版本命令包装器
func NewLegacyCommandWrapper(
	commandName string,
	description string,
	protocol string,
	legacyFunc func([]string),
) *LegacyCommandWrapper {
	baseHandler := NewBaseCommandHandler(
		commandName,
		description,
		Legacy,
		true, // 标记为已弃用
		nil,  // 传统版本不使用适配器
		nil,  // 传统版本不使用配置管理器
	)

	return &LegacyCommandWrapper{
		BaseCommandHandler: baseHandler,
		legacyFunc:         legacyFunc,
		protocol:           protocol,
	}
}

// ExecuteCommand 执行传统版本命令
func (w *LegacyCommandWrapper) ExecuteCommand(ctx context.Context, args []string) error {
	// 显示弃用警告
	w.showDeprecationWarning()

	// 调用传统函数
	w.legacyFunc(args)
	return nil
}

// showDeprecationWarning 显示弃用警告
func (w *LegacyCommandWrapper) showDeprecationWarning() {
	fmt.Printf("\n%s\n", strings.Repeat("⚠", 50))
	fmt.Printf("WARNING: Using DEPRECATED command '%s'\n", w.GetCommandName())
	fmt.Printf("Please migrate to the enhanced version: '%s-enhanced'\n", w.protocol)
	fmt.Printf("Enhanced version provides:\n")
	fmt.Printf("  ✓ Better performance with connection pooling\n")
	fmt.Printf("  ✓ Advanced metrics and monitoring\n")
	fmt.Printf("  ✓ Flexible configuration management\n")
	fmt.Printf("  ✓ Improved error handling and retry mechanisms\n")
	fmt.Printf("Migration guide: https://docs.redis-runner.com/migration\n")
	fmt.Printf("%s\n\n", strings.Repeat("⚠", 50))
}

// GetUsage 获取使用说明
func (w *LegacyCommandWrapper) GetUsage() string {
	switch w.protocol {
	case "redis":
		return w.getRedisUsage()
	case "http":
		return w.getHttpUsage()
	case "kafka":
		return w.getKafkaUsage()
	default:
		return w.BaseCommandHandler.GetUsage()
	}
}

// getRedisUsage 获取Redis使用说明
func (w *LegacyCommandWrapper) getRedisUsage() string {
	return `Usage: redis-runner redis [options]

DEPRECATED: This command is deprecated. Use 'redis-enhanced' instead.

Options:
  -h <hostname>     Server hostname (default: 127.0.0.1)
  -p <port>         Server port (default: 6379)
  -a <password>     Server password
  -n <requests>     Total number of requests (default: 1000)
  -c <connections>  Number of parallel connections (default: 10)
  -t <test>         Test case to run (default: set_get_random)

Migration: redis-runner redis-enhanced --config conf/redis.yaml`
}

// getHttpUsage 获取HTTP使用说明
func (w *LegacyCommandWrapper) getHttpUsage() string {
	return `Usage: redis-runner http [options]

DEPRECATED: This command is deprecated. Use 'http-enhanced' instead.

Options:
  --url <url>       Target URL
  -n <requests>     Total number of requests (default: 1000)
  -c <connections>  Number of parallel connections (default: 10)
  --method <method> HTTP method (default: GET)

Migration: redis-runner http-enhanced --config conf/http.yaml`
}

// getKafkaUsage 获取Kafka使用说明
func (w *LegacyCommandWrapper) getKafkaUsage() string {
	return `Usage: redis-runner kafka [options]

DEPRECATED: This command is deprecated. Use 'kafka-enhanced' instead.

Options:
  --broker <broker>   Kafka broker address
  --topic <topic>     Kafka topic name
  -n <requests>       Total number of messages (default: 1000)
  -c <connections>    Number of parallel connections (default: 3)

Migration: redis-runner kafka-enhanced --config conf/kafka.yaml`
}

// ValidateArgs 验证参数
func (w *LegacyCommandWrapper) ValidateArgs(args []string) error {
	// 传统版本参数验证较为宽松，主要由各自的实现处理
	return nil
}

// CreateLegacyWrappers 创建所有传统版本包装器
func CreateLegacyWrappers() map[string]*LegacyCommandWrapper {
	wrappers := make(map[string]*LegacyCommandWrapper)

	// Redis传统版本包装器
	redisWrapper := NewLegacyCommandWrapper(
		"redis",
		"Redis performance testing (DEPRECATED: use redis-enhanced)",
		"redis",
		redisCases.RedisCommand,
	)
	wrappers["redis"] = redisWrapper

	// HTTP传统版本包装器
	httpWrapper := NewLegacyCommandWrapper(
		"http",
		"HTTP load testing (DEPRECATED: use http-enhanced)",
		"http",
		httpCases.HttpCommand,
	)
	wrappers["http"] = httpWrapper

	// Kafka传统版本包装器
	kafkaWrapper := NewLegacyCommandWrapper(
		"kafka",
		"Kafka performance testing (DEPRECATED: use kafka-enhanced)",
		"kafka",
		kafkaCases.KafkaCommand,
	)
	wrappers["kafka"] = kafkaWrapper

	return wrappers
}

// RegisterLegacyCommands 注册传统版本命令
func RegisterLegacyCommands(router *CommandRouter) error {
	wrappers := CreateLegacyWrappers()

	for name, wrapper := range wrappers {
		if err := router.RegisterCommand(name, wrapper); err != nil {
			log.Printf("Failed to register legacy command '%s': %v", name, err)
			return err
		}
		log.Printf("Registered legacy command: %s (DEPRECATED)", name)
	}

	return nil
}

// LegacyCommandInfo 传统命令信息
type LegacyCommandInfo struct {
	Name            string `json:"name"`
	Protocol        string `json:"protocol"`
	DeprecatedSince string `json:"deprecated_since"`
	ReplacedBy      string `json:"replaced_by"`
	RemovalVersion  string `json:"removal_version"`
}

// GetLegacyCommandsInfo 获取传统命令信息
func GetLegacyCommandsInfo() []LegacyCommandInfo {
	return []LegacyCommandInfo{
		{
			Name:            "redis",
			Protocol:        "redis",
			DeprecatedSince: "v2.0.0",
			ReplacedBy:      "redis-enhanced",
			RemovalVersion:  "v3.0.0",
		},
		{
			Name:            "http",
			Protocol:        "http",
			DeprecatedSince: "v2.0.0",
			ReplacedBy:      "http-enhanced",
			RemovalVersion:  "v3.0.0",
		},
		{
			Name:            "kafka",
			Protocol:        "kafka",
			DeprecatedSince: "v2.0.0",
			ReplacedBy:      "kafka-enhanced",
			RemovalVersion:  "v3.0.0",
		},
	}
}