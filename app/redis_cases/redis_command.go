package redisCases

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func RedisCommand(args []string) {
	flags := flag.NewFlagSet("redis", flag.ExitOnError)
	config := flags.Bool("config", false, "redis run from config file")
	cluster := flags.Bool("cluster", false, "redis cluster mode")
	h := flags.String("h", "127.0.0.1", "Server hostname (default 127.0.0.1)")
	p := flags.Int("p", 6379, "Server port (default 6379)")
	a := flags.String("a", "", "Password for Redis Auth")
	n := flags.Int("n", 100000, "Total number of requests (default 100000)")
	c := flags.Int("c", 50, "Number of parallel connections (default 50)")
	r := flags.Int("r", 0, "Random keys for SET/GET/INCR, random values for SADD, ")
	d := flags.Int("d", 3, "Data size of SET/GET value in bytes (default 3)")
	t := flags.String("t", "get", "operation set get del incr hset hget sadd srem")
	R := flags.Int("R", 50, "read operation percent (default 100%)")
	ttl := flags.Int("ttl", 120, "TTL in seconds (default 300)")
	db := flags.Int("db", 0, "Number of parallel connections (default 50)")
	err := flags.Parse(args)
	if err != nil {
		return
	}

	if *config {
		fmt.Println("Execute using configuration and result in log file...")
		Start()
		return
	}

	if *h == "" {
		fmt.Printf("Command need host")
		return
	}

	if *p == 0 {
		fmt.Printf("Command need port")
	}

	mode := "standalone"
	if *cluster {
		mode = "cluster"
	}
	fmt.Printf("Redis Info: mode:%s, host:%s, port:%d, password:%s, total:%d, parallel:%d, dataSize:%d, random:%d, readPercent:%d, ttl:%d\n",
		mode, *h, *p, *a, *n, *c, *d, *r, *R, *ttl)

	addr := *h + ":" + strconv.Itoa(*p)
	redisConfigs := new(RedisConfig)
	redisConfigs.BenchMark.Case = *t
	redisConfigs.Mode = mode
	if mode == "cluster" {
		redisConfigs.Cluster.Addrs = []string{addr}
		redisConfigs.Cluster.Password = *a
	} else {
		redisConfigs.Standalone.Addr = addr
		redisConfigs.Standalone.Password = *a
		redisConfigs.Standalone.Db = *db
	}
	redisConfigs.Pool.PoolSize = 10
	redisConfigs.Pool.MinIdle = 10

	ConfigConnect(redisConfigs)

	if strings.ToLower(*t) == "set_get_random" {
		SetGetRandomCaseCommand(*n, *c, *r, *d, *R, *ttl)
	}
	if strings.ToLower(*t) == "get" {
		GetCaseCommand(*n, *c)
	}
	if strings.ToLower(*t) == "set" {
		SetCaseCommand(*n, *c, *r, *d, *R, *ttl)
	}
	if strings.ToLower(*t) == "del" {
		DelCaseCommand(*n, *c, *r)
	}
	if strings.ToLower(*t) == "pub" {
		PubCaseCommand(*n, *c, *d)
	}
	if strings.ToLower(*t) == "sub" {
		SubCaseCommand(mode, *c)
	}
}
