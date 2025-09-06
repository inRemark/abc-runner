package commands

import (
	"log"

	"redis-runner/app/core/command"
)

// RegisterEnhancedCommands 注册所有增强版命令
func RegisterEnhancedCommands(router *command.CommandRouter) error {
	// 注册 Redis Enhanced
	redisHandler := NewRedisCommandHandler()
	if err := router.RegisterCommand("redis-enhanced", redisHandler); err != nil {
		log.Printf("Failed to register redis-enhanced command: %v", err)
		return err
	}
	log.Println("Registered enhanced command: redis-enhanced")

	// 注册 HTTP Enhanced
	httpHandler := NewHttpCommandHandler()
	if err := router.RegisterCommand("http-enhanced", httpHandler); err != nil {
		log.Printf("Failed to register http-enhanced command: %v", err)
		return err
	}
	log.Println("Registered enhanced command: http-enhanced")

	// 注册 Kafka Enhanced
	kafkaHandler := NewKafkaCommandHandler()
	if err := router.RegisterCommand("kafka-enhanced", kafkaHandler); err != nil {
		log.Printf("Failed to register kafka-enhanced command: %v", err)
		return err
	}
	log.Println("Registered enhanced command: kafka-enhanced")

	return nil
}

// GetEnhancedCommandsInfo 获取增强版命令信息
func GetEnhancedCommandsInfo() []command.CommandInfo {
	return []command.CommandInfo{
		{
			Name:        "redis-enhanced",
			Description: "Redis performance testing with advanced features",
			Version:     command.Enhanced,
			Deprecated:  false,
			Usage:       "redis-runner redis-enhanced [options]",
		},
		{
			Name:        "http-enhanced",
			Description: "HTTP load testing with enterprise features",
			Version:     command.Enhanced,
			Deprecated:  false,
			Usage:       "redis-runner http-enhanced [options]",
		},
		{
			Name:        "kafka-enhanced",
			Description: "Kafka performance testing with connection pooling",
			Version:     command.Enhanced,
			Deprecated:  false,
			Usage:       "redis-runner kafka-enhanced [options]",
		},
	}
}