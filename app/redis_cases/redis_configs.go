package redisCases

import (
	"gopkg.in/yaml.v2"
	"os"
)

type StandAloneInfo struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type SentinelInfo struct {
	MasterName string   `yaml:"master_name"`
	Addrs      []string `yaml:"addrs"`
	Password   string   `yaml:"password"`
	Db         int      `yaml:"db"`
}

type ClusterInfo struct {
	Addrs    []string `yaml:"addrs"`
	Password string   `yaml:"password"`
}

type PoolInfo struct {
	PoolSize int `yaml:"pool_size"`
	MinIdle  int `yaml:"min_idle"`
}

type BenchMarkInfo struct {
	DataSize    int    `yaml:"data_size"`
	Parallels   int    `yaml:"parallels"`
	Total       int    `yaml:"total"`
	TTL         int    `yaml:"ttl"`
	ReadPercent int    `yaml:"read_percent"`
	RandomKeys  int    `yaml:"random_keys"`
	Case        string `yaml:"case"`
}

type RedisConfig struct {
	Mode       string         `yaml:"mode"`
	BenchMark  BenchMarkInfo  `yaml:"benchmark"`
	Pool       PoolInfo       `yaml:"pool"`
	Standalone StandAloneInfo `yaml:"standalone"`
	Sentinel   SentinelInfo   `yaml:"sentinel"`
	Cluster    ClusterInfo    `yaml:"cluster"`
}

func LoadConfig() (*RedisConfig, error) {

	data, err := os.ReadFile("conf/redis-config.yaml")
	if err != nil {
		return nil, err
	}

	var redisConfig struct {
		Redis RedisConfig `yaml:"redis"`
	}
	err = yaml.Unmarshal(data, &redisConfig)
	if err != nil {
		return nil, err
	}

	return &redisConfig.Redis, nil
}
