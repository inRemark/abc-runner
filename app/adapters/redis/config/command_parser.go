package config

import (
	"strconv"
	"strings"
)

// CommandLineConfigSource 命令行配置源
type CommandLineConfigSource struct {
	Args []string
}

// NewCommandLineConfigSource 创建命令行配置源
func NewCommandLineConfigSource(args []string) *CommandLineConfigSource {
	return &CommandLineConfigSource{Args: args}
}

// Load 从命令行参数加载配置
func (c *CommandLineConfigSource) Load() (*RedisConfig, error) {
	config := NewDefaultRedisConfig()
	argsMap := c.parseArgs()

	// 解析协议
	if protocol, exists := argsMap["protocol"]; exists {
		config.Protocol = protocol
	}

	// 解析Redis模式
	if mode, exists := argsMap["mode"]; exists {
		config.Mode = mode
	} else {
		// 根据参数推断模式
		if _, hasCluster := argsMap["cluster"]; hasCluster {
			config.Mode = "cluster"
		}
	}

	// 解析基准测试配置
	if total, exists := argsMap["n"]; exists {
		if val, err := strconv.Atoi(total); err == nil {
			config.BenchMark.Total = val
		}
	}

	if parallels, exists := argsMap["c"]; exists {
		if val, err := strconv.Atoi(parallels); err == nil {
			config.BenchMark.Parallels = val
		}
	}

	if dataSize, exists := argsMap["d"]; exists {
		if val, err := strconv.Atoi(dataSize); err == nil {
			config.BenchMark.DataSize = val
		}
	}

	if ttl, exists := argsMap["ttl"]; exists {
		if val, err := strconv.Atoi(ttl); err == nil {
			config.BenchMark.TTL = val
		}
	}

	if readPercent, exists := argsMap["R"]; exists {
		if val, err := strconv.Atoi(readPercent); err == nil {
			config.BenchMark.ReadPercent = val
		}
	}

	if randomKeys, exists := argsMap["r"]; exists {
		if val, err := strconv.Atoi(randomKeys); err == nil {
			config.BenchMark.RandomKeys = val
		}
	}

	if testCase, exists := argsMap["t"]; exists {
		config.BenchMark.Case = testCase
	}

	// 解析连接配置
	host := "localhost"
	if h, exists := argsMap["h"]; exists {
		host = h
	}

	port := "6379"
	if p, exists := argsMap["p"]; exists {
		port = p
	}

	addr := host + ":" + port

	switch config.Mode {
	case "cluster":
		// 集群模式
		if addresses, exists := argsMap["cluster-addrs"]; exists {
			config.Cluster.Addrs = strings.Split(addresses, ",")
		} else {
			// 如果没有指定集群地址，使用默认地址列表
			config.Cluster.Addrs = []string{addr}
		}
		if password, exists := argsMap["a"]; exists {
			config.Cluster.Password = password
		}
	case "sentinel":
		// 哨兵模式
		if masterName, exists := argsMap["sentinel-master"]; exists {
			config.Sentinel.MasterName = masterName
		}
		if addresses, exists := argsMap["sentinel-addrs"]; exists {
			config.Sentinel.Addrs = strings.Split(addresses, ",")
		} else {
			config.Sentinel.Addrs = []string{addr}
		}
		if password, exists := argsMap["a"]; exists {
			config.Sentinel.Password = password
		}
		if db, exists := argsMap["db"]; exists {
			if val, err := strconv.Atoi(db); err == nil {
				config.Sentinel.Db = val
			}
		}
	default:
		// 单机模式
		config.Standalone.Addr = addr
		if password, exists := argsMap["a"]; exists {
			config.Standalone.Password = password
		}
		if db, exists := argsMap["db"]; exists {
			if val, err := strconv.Atoi(db); err == nil {
				config.Standalone.Db = val
			}
		}
	}

	return config, nil
}

// parseArgs 解析命令行参数
func (c *CommandLineConfigSource) parseArgs() map[string]string {
	argsMap := make(map[string]string)

	for i := 0; i < len(c.Args); i++ {
		arg := c.Args[i]
		
		// 处理标志参数
		if arg == "--cluster" {
			argsMap["cluster"] = "true"
			continue
		}
		
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			if i+1 < len(c.Args) && !strings.HasPrefix(c.Args[i+1], "-") {
				argsMap[key] = c.Args[i+1]
				i++
			} else {
				argsMap[key] = "true"
			}
			continue
		}
		
		if strings.HasPrefix(arg, "-") {
			key := strings.TrimPrefix(arg, "-")
			if i+1 < len(c.Args) && !strings.HasPrefix(c.Args[i+1], "-") {
				argsMap[key] = c.Args[i+1]
				i++
			} else {
				argsMap[key] = "true"
			}
			continue
		}
	}

	return argsMap
}

// CanLoad 检查是否可以加载
func (c *CommandLineConfigSource) CanLoad() bool {
	return len(c.Args) > 0
}

// Priority 获取优先级（命令行参数优先级最高）
func (c *CommandLineConfigSource) Priority() int {
	return 100
}