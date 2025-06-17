package redisCases

import (
	"strconv"
	"testing"
)

func TestRedisCommand(t *testing.T) {
	mode, h, p, a, db, n, c, d, r, R, ttl, testCase :=
		"standalone", "127.0.0.1", 6371, "pwd@redis", 0, 100000, 50, 5, 200000, 100, 300, "get"

	addr := h + ":" + strconv.Itoa(p)
	redisConfigs := new(RedisConfig)
	redisConfigs.BenchMark.Case = testCase
	redisConfigs.Mode = mode
	if mode == "cluster" {
		redisConfigs.Cluster.Addrs = []string{addr}
		redisConfigs.Cluster.Password = a
	} else {
		redisConfigs.Standalone.Addr = addr
		redisConfigs.Standalone.Password = a
		redisConfigs.Standalone.Db = db
	}
	redisConfigs.Pool.PoolSize = 10
	redisConfigs.Pool.MinIdle = 10
	ConfigConnect(redisConfigs)
	SetGetRandomCaseCommand(n, c, r, d, R, ttl)
}
