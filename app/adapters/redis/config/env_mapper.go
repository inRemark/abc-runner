package config

import (
	"os"
	"strconv"
	"strings"

	"abc-runner/app/core/interfaces"
)

// RedisEnvVarMapper Redis环境变量映射器
type RedisEnvVarMapper struct {
	prefix string
}

// NewRedisEnvVarMapper 创建Redis环境变量映射器
func NewRedisEnvVarMapper(prefix string) *RedisEnvVarMapper {
	if prefix == "" {
		prefix = "REDIS_RUNNER"
	}
	return &RedisEnvVarMapper{prefix: prefix}
}

// MapEnvVarsToConfig 将环境变量映射到配置
func (r *RedisEnvVarMapper) MapEnvVarsToConfig(config interfaces.Config) error {
	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		return nil // Not a Redis config, nothing to do
	}

	// 基本配置
	if protocol := os.Getenv(r.prefix + "_PROTOCOL"); protocol != "" {
		redisConfig.Protocol = protocol
	}

	if mode := os.Getenv(r.prefix + "_MODE"); mode != "" {
		redisConfig.Mode = mode
	}

	// 基准测试配置
	if total := os.Getenv(r.prefix + "_TOTAL"); total != "" {
		if val, err := strconv.Atoi(total); err == nil {
			redisConfig.BenchMark.Total = val
		}
	}

	if parallels := os.Getenv(r.prefix + "_PARALLELS"); parallels != "" {
		if val, err := strconv.Atoi(parallels); err == nil {
			redisConfig.BenchMark.Parallels = val
		}
	}

	if dataSize := os.Getenv(r.prefix + "_DATA_SIZE"); dataSize != "" {
		if val, err := strconv.Atoi(dataSize); err == nil {
			redisConfig.BenchMark.DataSize = val
		}
	}

	if ttl := os.Getenv(r.prefix + "_TTL"); ttl != "" {
		if val, err := strconv.Atoi(ttl); err == nil {
			redisConfig.BenchMark.TTL = val
		}
	}

	if readPercent := os.Getenv(r.prefix + "_READ_PERCENT"); readPercent != "" {
		if val, err := strconv.Atoi(readPercent); err == nil {
			redisConfig.BenchMark.ReadPercent = val
		}
	}

	if randomKeys := os.Getenv(r.prefix + "_RANDOM_KEYS"); randomKeys != "" {
		if val, err := strconv.Atoi(randomKeys); err == nil {
			redisConfig.BenchMark.RandomKeys = val
		}
	}

	if testCase := os.Getenv(r.prefix + "_CASE"); testCase != "" {
		redisConfig.BenchMark.Case = testCase
	}

	// 连接池配置
	if poolSize := os.Getenv(r.prefix + "_POOL_SIZE"); poolSize != "" {
		if val, err := strconv.Atoi(poolSize); err == nil {
			redisConfig.Pool.PoolSize = val
		}
	}

	if minIdle := os.Getenv(r.prefix + "_MIN_IDLE"); minIdle != "" {
		if val, err := strconv.Atoi(minIdle); err == nil {
			redisConfig.Pool.MinIdle = val
		}
	}

	if maxIdle := os.Getenv(r.prefix + "_MAX_IDLE"); maxIdle != "" {
		if val, err := strconv.Atoi(maxIdle); err == nil {
			redisConfig.Pool.MaxIdle = val
		}
	}

	// 单机配置
	if addr := os.Getenv(r.prefix + "_ADDR"); addr != "" {
		redisConfig.Standalone.Addr = addr
	}

	if password := os.Getenv(r.prefix + "_PASSWORD"); password != "" {
		redisConfig.Standalone.Password = password
	}

	if db := os.Getenv(r.prefix + "_DB"); db != "" {
		if val, err := strconv.Atoi(db); err == nil {
			redisConfig.Standalone.Db = val
		}
	}

	// 哨兵配置
	if masterName := os.Getenv(r.prefix + "_SENTINEL_MASTER_NAME"); masterName != "" {
		redisConfig.Sentinel.MasterName = masterName
	}

	if sentinelAddrs := os.Getenv(r.prefix + "_SENTINEL_ADDRS"); sentinelAddrs != "" {
		redisConfig.Sentinel.Addrs = strings.Split(sentinelAddrs, ",")
	}

	if sentinelPassword := os.Getenv(r.prefix + "_SENTINEL_PASSWORD"); sentinelPassword != "" {
		redisConfig.Sentinel.Password = sentinelPassword
	}

	if sentinelDb := os.Getenv(r.prefix + "_SENTINEL_DB"); sentinelDb != "" {
		if val, err := strconv.Atoi(sentinelDb); err == nil {
			redisConfig.Sentinel.Db = val
		}
	}

	// 集群配置
	if clusterAddrs := os.Getenv(r.prefix + "_CLUSTER_ADDRS"); clusterAddrs != "" {
		redisConfig.Cluster.Addrs = strings.Split(clusterAddrs, ",")
	}

	if clusterPassword := os.Getenv(r.prefix + "_CLUSTER_PASSWORD"); clusterPassword != "" {
		redisConfig.Cluster.Password = clusterPassword
	}

	return nil
}

// HasRelevantEnvVars 检查是否有相关的环境变量
func (r *RedisEnvVarMapper) HasRelevantEnvVars() bool {
	// 检查是否有任何相关的环境变量
	envVars := []string{
		r.prefix + "_PROTOCOL",
		r.prefix + "_MODE",
		r.prefix + "_ADDR",
		r.prefix + "_TOTAL",
		r.prefix + "_PARALLELS",
		r.prefix + "_DATA_SIZE",
		r.prefix + "_TTL",
		r.prefix + "_READ_PERCENT",
		r.prefix + "_RANDOM_KEYS",
		r.prefix + "_CASE",
		r.prefix + "_POOL_SIZE",
		r.prefix + "_MIN_IDLE",
		r.prefix + "_MAX_IDLE",
		r.prefix + "_SENTINEL_MASTER_NAME",
		r.prefix + "_SENTINEL_ADDRS",
		r.prefix + "_SENTINEL_PASSWORD",
		r.prefix + "_SENTINEL_DB",
		r.prefix + "_CLUSTER_ADDRS",
		r.prefix + "_CLUSTER_PASSWORD",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}