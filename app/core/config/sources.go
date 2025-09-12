package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"abc-runner/app/core/interfaces"
)

// EnvironmentConfigSource 环境变量配置源
type EnvironmentConfigSource struct {
	Prefix string
}

// NewEnvironmentConfigSource 创建环境变量配置源
func NewEnvironmentConfigSource(prefix string) *EnvironmentConfigSource {
	if prefix == "" {
		prefix = "ABC_RUNNER"
	}
	return &EnvironmentConfigSource{Prefix: prefix}
}

// Load 从环境变量加载配置
func (e *EnvironmentConfigSource) Load() (interfaces.Config, error) {
	// 这是一个通用的环境变量配置源，实际实现应该由各协议适配器提供
	// 这里返回一个简单的实现用于兼容性
	return nil, nil
}

// CanLoad 检查是否可以从环境变量加载
func (e *EnvironmentConfigSource) CanLoad() bool {
	// 检查关键环境变量是否存在
	_, exists := os.LookupEnv(e.envKey("PROTOCOL"))
	return exists
}

// Priority 获取优先级
func (e *EnvironmentConfigSource) Priority() int {
	return 2 // 环境变量优先级高于文件
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
	// 这是一个通用的命令行配置源，实际实现应该由各协议适配器提供
	// 这里返回一个简单的实现用于兼容性
	return nil, nil
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
