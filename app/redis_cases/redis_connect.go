package redisCases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisCmd redis.Cmdable
var redisDb *redis.Client
var redisCluster *redis.ClusterClient
var redisConfigs *RedisConfig

func standalone(config *RedisConfig) *redis.Client {
	standalone := redis.NewClient(&redis.Options{
		Addr:         config.Standalone.Addr,
		DB:           config.Standalone.Db,
		Password:     config.Standalone.Password,
		PoolSize:     config.Pool.PoolSize,
		MinIdleConns: config.Pool.MinIdle,
	})
	return standalone
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
func ConfigConnect(redisConfigs *RedisConfig) {
	if redisConfigs.BenchMark.Case == "sub" {
		clientSub(redisConfigs)
	} else {
		client(redisConfigs)
	}
}

func client(redisConfigs *RedisConfig) {
	if redisConfigs.Mode == "cluster" {
		redisCmd = cluster(redisConfigs)
	} else {
		if redisConfigs.Mode == "standalone" {
			redisCmd = standalone(redisConfigs)
		} else if redisConfigs.Mode == "sentinel" {
			redisCmd = sentinel(redisConfigs)
		}
	}
	_, err := redisCmd.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("redis client connect failed: %v", err)
	}

	log.Printf("redis client connect successfully")
}

func clientSub(redisConfigs *RedisConfig) {
	if redisConfigs.Mode == "cluster" {
		redisCluster = cluster(redisConfigs)
		_, err := redisCluster.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("redis sub client connect failed: %v", err)
		}
	} else {
		if redisConfigs.Mode == "standalone" {
			redisDb = standalone(redisConfigs)
		} else if redisConfigs.Mode == "sentinel" {
			redisDb = sentinel(redisConfigs)
		}
		_, err := redisDb.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("redis sub client connect failed: %v", err)
		}
	}
	log.Printf("redis connect successfully")
}

func Get(key string) (result string) {
	val, err := redisCmd.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	return val
}

func Set(key string, value interface{}, duration time.Duration) {
	redisCmd.Set(ctx, key, value, duration)
}

func Del(key string) {
	redisCmd.Del(ctx, key)
}

func HSet(key, field string, value interface{}) {
	redisCmd.HSet(ctx, key, field, value)
}

func HSetTNX(key, field string, value interface{}) {
	redisCmd.HSetNX(ctx, key, field, value)
}

func HGet(key string, field string) interface{} {
	return redisCmd.HGet(ctx, key, field)
}

func HGetAll() {

}

func Pub(value interface{}) {
	channelName := "my_channel"
	err := redisCmd.Publish(ctx, channelName, value).Err()
	if err != nil {
		fmt.Printf("Failed to publish message cluster:  %v \n", err)
		return
	}
}

func Sub(mode string, index int) {
	channelName := "my_channel"
	if mode == "cluster" {
		sub := redisCluster.Subscribe(ctx, channelName)
		defer func(sub *redis.PubSub) {
			err := sub.Close()
			if err != nil {
				log.Fatalf("redis connect failed: %v", err)
			}
		}(sub)
		for msg := range sub.Channel() {
			fmt.Printf("Received_message_%d: %s \n", index, msg.Payload)
		}
	} else {
		sub := redisDb.Subscribe(ctx, channelName)
		defer func(sub *redis.PubSub) {
			err := sub.Close()
			if err != nil {
				log.Fatalf("redis connect failed: %v", err)
			}
		}(sub)
		for msg := range sub.Channel() {
			fmt.Printf("Received_message_%d: %s \n", index, msg.Payload)
		}
	}
}
