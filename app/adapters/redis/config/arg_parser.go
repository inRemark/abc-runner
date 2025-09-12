package config

import (
	"strconv"
	"strings"

	"abc-runner/app/core/interfaces"
)

// RedisArgParser Redis命令行参数解析器
type RedisArgParser struct{}

// NewRedisArgParser 创建Redis命令行参数解析器
func NewRedisArgParser() *RedisArgParser {
	return &RedisArgParser{}
}

// ParseArgs 解析命令行参数
func (r *RedisArgParser) ParseArgs(args []string, config interfaces.Config) error {
	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		return nil // Not a Redis config, nothing to do
	}

	// 从命令行参数解析配置
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--protocol":
			if i+1 < len(args) {
				redisConfig.Protocol = args[i+1]
				i++
			}
		case "--mode":
			if i+1 < len(args) {
				redisConfig.Mode = args[i+1]
				i++
			}
		case "--total", "-n":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.BenchMark.Total = val
				}
				i++
			}
		case "--parallels", "-c":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.BenchMark.Parallels = val
				}
				i++
			}
		case "--data-size":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.BenchMark.DataSize = val
				}
				i++
			}
		case "--ttl":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.BenchMark.TTL = val
				}
				i++
			}
		case "--read-percent":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.BenchMark.ReadPercent = val
				}
				i++
			}
		case "--random-keys":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.BenchMark.RandomKeys = val
				}
				i++
			}
		case "--case":
			if i+1 < len(args) {
				redisConfig.BenchMark.Case = args[i+1]
				i++
			}
		case "--addr":
			if i+1 < len(args) {
				redisConfig.Standalone.Addr = args[i+1]
				i++
			}
		case "--password":
			if i+1 < len(args) {
				redisConfig.Standalone.Password = args[i+1]
				i++
			}
		case "--db":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.Standalone.Db = val
				}
				i++
			}
		case "--sentinel-master-name":
			if i+1 < len(args) {
				redisConfig.Sentinel.MasterName = args[i+1]
				i++
			}
		case "--sentinel-addrs":
			if i+1 < len(args) {
				redisConfig.Sentinel.Addrs = strings.Split(args[i+1], ",")
				i++
			}
		case "--sentinel-password":
			if i+1 < len(args) {
				redisConfig.Sentinel.Password = args[i+1]
				i++
			}
		case "--sentinel-db":
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					redisConfig.Sentinel.Db = val
				}
				i++
			}
		case "--cluster-addrs":
			if i+1 < len(args) {
				redisConfig.Cluster.Addrs = strings.Split(args[i+1], ",")
				i++
			}
		case "--cluster-password":
			if i+1 < len(args) {
				redisConfig.Cluster.Password = args[i+1]
				i++
			}
		}
	}

	return nil
}
