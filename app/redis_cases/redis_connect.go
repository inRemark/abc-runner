package redisCases

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisDb *redis.Client
var redisCluster *redis.ClusterClient
var redisConfigs *RedisConfig

func standalone(config *RedisConfig) *redis.Client {
	redisDb = redis.NewClient(&redis.Options{
		Addr:         config.Standalone.Addr,
		DB:           config.Standalone.Db,
		Password:     config.Standalone.Password,
		PoolSize:     config.Pool.PoolSize,
		MinIdleConns: config.Pool.MinIdle,
	})
	return redisDb
}

func sentinel(config *RedisConfig) *redis.Client {

	sentinel := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    config.Sentinel.MasterName,
		SentinelAddrs: config.Sentinel.Addrs,
		Password:      config.Sentinel.Password,
		DB:            config.Sentinel.Db,
		PoolSize:      config.Pool.PoolSize,
		MinIdleConns:  config.Pool.MinIdle,
	})
	return sentinel
}

func cluster(config *RedisConfig) *redis.ClusterClient {
	cluster := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Cluster.Addrs,
		Password:     config.Cluster.Password,
		PoolSize:     config.Pool.PoolSize,
		MinIdleConns: config.Pool.MinIdle,
	})

	return cluster
}
func ConfigConnect() {
	var err error
	redisConfigs, err = LoadConfig()
	if err != nil {
		log.Fatalf("redis-config load failed: %v", err)
	}
	connect(redisConfigs)
}

func CommandConnect(mode, h, a string, p, db int) {
	addr := h + ":" + strconv.Itoa(p)
	redisConfigs := new(RedisConfig)
	redisConfigs.Pool.PoolSize = 10
	redisConfigs.Pool.MinIdle = 10
	redisConfigs.Mode = mode

	redisConfigs.Standalone.Addr = addr
	redisConfigs.Standalone.Password = a
	redisConfigs.Standalone.Db = db

	redisConfigs.Cluster.Addrs = []string{addr}
	redisConfigs.Cluster.Password = a
	connect(redisConfigs)
}

func connect(redisConfigs *RedisConfig) {

	if redisConfigs.Mode == "cluster" {
		redisCluster = cluster(redisConfigs)
		_, err := redisCluster.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("redis connect failed: %v", err)
		}
	} else {
		if redisConfigs.Mode == "standalone" {
			redisDb = standalone(redisConfigs)
		} else if redisConfigs.Mode == "sentinel" {
			redisDb = sentinel(redisConfigs)
		}
		_, err := redisDb.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("redis connect failed: %v", err)
		}
	}

	log.Printf("redis connect successfully")
}

func Get(mode string, key string) (result string) {
	if mode == "cluster" {
		val, err := redisCluster.Get(ctx, key).Result()
		if err != nil {
			return ""
		}
		return val
	} else {
		val, err := redisDb.Get(ctx, key).Result()
		if err != nil {
			return ""
		}
		return val
	}
}

func Set(mode string, key string, value interface{}, duration time.Duration) {
	if mode == "cluster" {
		redisCluster.Set(ctx, key, value, duration)
	} else {
		redisDb.Set(ctx, key, value, duration)
	}
}

func Del(mode string, key string) {
	if mode == "cluster" {
		redisCluster.Del(ctx, key)
	} else {
		redisDb.Del(ctx, key)
	}
}

func Pub(mode string, value interface{}) {
	channelName := "mychannel"
	if mode == "cluster" {
		err := redisCluster.Publish(ctx, channelName, value).Err()
		if err != nil {
			fmt.Printf("Failed to publish message cluster:  %v \n", err)
			return
		}
	} else {
		err := redisDb.Publish(ctx, channelName, value).Err()
		if err != nil {
			fmt.Printf("Failed to publish message standalone: %v \n", err)
			return
		}
	}
}

func Sub(mode string, index int) {
	channelName := "mychannel"
	if mode == "cluster" {
		pubsub := redisCluster.Subscribe(ctx, channelName)
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			fmt.Printf("Received_message_%d: %s \n", index, msg.Payload)
		}
	} else {
		pubsub := redisDb.Subscribe(ctx, channelName)
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			fmt.Printf("Received_message_%d: %s \n", index, msg.Payload)
		}
	}
}
