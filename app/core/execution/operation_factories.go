package execution

import (
	"fmt"

	"abc-runner/app/core/interfaces"
	httpConfig "abc-runner/app/adapters/http/config"
	kafkaConfig "abc-runner/app/adapters/kafka/config"
)

// HttpOperationFactory HTTP操作工厂
type HttpOperationFactory struct {
	config *httpConfig.HttpAdapterConfig
}

// NewHttpOperationFactory 创建HTTP操作工厂
func NewHttpOperationFactory(config *httpConfig.HttpAdapterConfig) OperationFactory {
	return &HttpOperationFactory{config: config}
}

func (h *HttpOperationFactory) CreateOperation(jobID int, config BenchmarkConfig) interfaces.Operation {
	// 从配置中选择请求配置
	var reqConfig httpConfig.HttpRequestConfig
	if len(h.config.Requests) > 0 {
		// 轮询选择请求配置
		reqConfig = h.config.Requests[jobID%len(h.config.Requests)]
	} else {
		// 默认GET请求
		reqConfig = httpConfig.HttpRequestConfig{
			Method: "GET",
			Path:   "/",
		}
	}

	return interfaces.Operation{
		Type: "http_request",
		Key:  fmt.Sprintf("http_job_%d", jobID),
		Params: map[string]interface{}{
			"method":  reqConfig.Method,
			"path":    reqConfig.Path,
			"headers": reqConfig.Headers,
			"body":    reqConfig.Body,
			"weight":  reqConfig.Weight,
		},
	}
}

// RedisOperationFactory Redis操作工厂
type RedisOperationFactory struct {
	config interfaces.Config
}

// NewRedisOperationFactory 创建Redis操作工厂
func NewRedisOperationFactory(config interfaces.Config) OperationFactory {
	return &RedisOperationFactory{config: config}
}

func (r *RedisOperationFactory) CreateOperation(jobID int, config BenchmarkConfig) interfaces.Operation {
	benchmark := r.config.GetBenchmark()
	
	// 根据读写比例决定操作类型
	isRead := (jobID % 100) < benchmark.GetReadPercent()
	
	var opType string
	var key, value string
	
	// 生成键
	if benchmark.GetRandomKeys() > 0 {
		key = fmt.Sprintf("key_%d", jobID%benchmark.GetRandomKeys())
	} else {
		key = fmt.Sprintf("key_%d", jobID)
	}
	
	if isRead {
		opType = "get"
	} else {
		opType = "set"
		// 生成指定大小的值
		dataSize := benchmark.GetDataSize()
		if dataSize <= 0 {
			dataSize = 64
		}
		value = generateRandomValue(dataSize)
	}

	operation := interfaces.Operation{
		Type:  opType,
		Key:   key,
		Value: value,
		TTL:   benchmark.GetTTL(),
		Params: map[string]interface{}{
			"operation_type": opType,
			"job_id":         jobID,
			"is_read":        isRead,
		},
	}

	return operation
}

// KafkaOperationFactory Kafka操作工厂
type KafkaOperationFactory struct {
	config *kafkaConfig.KafkaAdapterConfig
}

// NewKafkaOperationFactory 创建Kafka操作工厂
func NewKafkaOperationFactory(config *kafkaConfig.KafkaAdapterConfig) OperationFactory {
	return &KafkaOperationFactory{config: config}
}

func (k *KafkaOperationFactory) CreateOperation(jobID int, config BenchmarkConfig) interfaces.Operation {
	benchmark := k.config.Benchmark
	
	// 根据测试类型确定操作
	var opType string
	switch benchmark.TestCase {
	case "consumer", "consume":
		opType = "consume_message"
	case "producer", "produce":
		opType = "produce_message"
	default:
		// 根据读写比例决定
		if (jobID % 100) < benchmark.ReadPercent {
			opType = "consume_message"
		} else {
			opType = "produce_message"
		}
	}
	
	// 生成消息
	key := fmt.Sprintf("key_%d", jobID)
	value := fmt.Sprintf("message_%d", jobID)
	
	// 如果有指定数据大小，生成相应大小的值
	if benchmark.DataSize > 0 {
		value = generateRandomValue(benchmark.DataSize)
	}

	return interfaces.Operation{
		Type:  opType,
		Key:   key,
		Value: value,
		Params: map[string]interface{}{
			"topic":        benchmark.DefaultTopic,
			"partition":    jobID % 3, // 简单分区策略
			"job_id":       jobID,
			"test_type":    benchmark.TestCase,
			"message_size": benchmark.DataSize,
		},
	}
}

// generateRandomValue 生成指定大小的随机值
func generateRandomValue(size int) string {
	if size <= 0 {
		return ""
	}
	
	// 简单的值生成逻辑
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, size)
	
	for i := 0; i < size; i++ {
		result[i] = charset[i%len(charset)]
	}
	
	return string(result)
}