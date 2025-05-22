package redisCases

import (
	"flag"
	"fmt"
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
		tc := *t
		fmt.Println("Execute using configuration and result in log file...")
		Start(tc)
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

	CommandConnect(mode, *h, *a, *p, *db)
	if strings.ToLower(*t) == "set_get_random" {
		DoSetGetRandomCaseCommand(mode, *n, *c, *r, *d, *R, *ttl)
	}
	if strings.ToLower(*t) == "get" {
		DoGetCaseCommand(mode, *n, *c)
	}
	if strings.ToLower(*t) == "set" {
		DoSetCaseCommand(mode, *n, *c, *r, *d, *R, *ttl)
	}
	if strings.ToLower(*t) == "del" {
		DoDelCaseCommand(mode, *n, *c, *r)
	}
	if strings.ToLower(*t) == "pub" {
		DoPubCaseCommand(mode, *n, *c, *d)
	}
	if strings.ToLower(*t) == "sub" {
		DoSubCaseCommand(mode, *c)
	}
}
