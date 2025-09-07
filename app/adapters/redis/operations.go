package redis

import (
	"fmt"
	"strings"
	"time"

	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/utils"
)

// Redis操作工厂实现

// StringGetOperationFactory GET操作工厂
type StringGetOperationFactory struct{}

func (f *StringGetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 从参数中获取键生成器
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 生成或获取键
	var key string
	if randomKeys, exists := params["random_keys"]; exists {
		if rk, ok := randomKeys.(int); ok && rk > 0 {
			key = keyGen.GenerateRandomKey("get", rk)
		} else {
			key = keyGen.GetRandomFromGenerated()
		}
	} else {
		key = keyGen.GetRandomFromGenerated()
	}

	return interfaces.Operation{
		Type: "get",
		Key:  key,
	}, nil
}

func (f *StringGetOperationFactory) GetOperationType() string {
	return "get"
}

func (f *StringGetOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// StringSetOperationFactory SET操作工厂
type StringSetOperationFactory struct{}

func (f *StringSetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 获取TTL
	ttl, ok := params["ttl"].(time.Duration)
	if !ok {
		if ttlInt, ok := params["ttl"].(int); ok {
			ttl = time.Duration(ttlInt) * time.Second
		} else {
			ttl = 120 * time.Second
		}
	}

	// 获取随机键范围
	randomKeys, ok := params["random_keys"].(int)
	if !ok {
		randomKeys = 0
	}

	// 生成键
	var key string
	if randomKeys > 0 {
		key = keyGen.GenerateRandomKey("set", randomKeys)
	} else {
		// 获取总数用于递增键
		total, ok := params["total"].(int)
		if !ok {
			total = 100000
		}
		key = keyGen.GenerateKey("set", int64(total))
	}

	// 生成值
	value := strings.Repeat("X", dataSize)

	return interfaces.Operation{
		Type:  "set",
		Key:   key,
		Value: value,
		TTL:   ttl,
	}, nil
}

func (f *StringSetOperationFactory) GetOperationType() string {
	return "set"
}

func (f *StringSetOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// StringDeleteOperationFactory DELETE操作工厂
type StringDeleteOperationFactory struct{}

func (f *StringDeleteOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 从已生成的键中随机选择
	key := keyGen.GetRandomFromGenerated()

	return interfaces.Operation{
		Type: "del",
		Key:  key,
	}, nil
}

func (f *StringDeleteOperationFactory) GetOperationType() string {
	return "del"
}

func (f *StringDeleteOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// SetGetRandomOperationFactory 混合读写操作工厂
type SetGetRandomOperationFactory struct{}

func (f *SetGetRandomOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取读写比例
	readPercent, ok := params["read_percent"].(int)
	if !ok {
		readPercent = 50
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 获取TTL
	ttl, ok := params["ttl"].(time.Duration)
	if !ok {
		if ttlInt, ok := params["ttl"].(int); ok {
			ttl = time.Duration(ttlInt) * time.Second
		} else {
			ttl = 120 * time.Second
		}
	}

	// 获取随机键范围
	randomKeys, ok := params["random_keys"].(int)
	if !ok {
		randomKeys = 0
	}

	operation := interfaces.Operation{
		Type: "set_get_random",
		Params: map[string]interface{}{
			"read_percent": readPercent,
			"data_size":    dataSize,
			"random_keys":  randomKeys,
		},
		TTL: ttl,
	}

	// 生成键
	if randomKeys > 0 {
		operation.Key = keyGen.GenerateRandomKey("mixed", randomKeys)
	} else {
		operation.Key = keyGen.GetRandomFromGenerated()
		if operation.Key == "" {
			// 如果没有已生成的键，生成一个新的
			total, ok := params["total"].(int)
			if !ok {
				total = 100000
			}
			operation.Key = keyGen.GenerateKey("mixed", int64(total))
		}
	}

	// 设置值
	operation.Value = strings.Repeat("X", dataSize)

	return operation, nil
}

func (f *SetGetRandomOperationFactory) GetOperationType() string {
	return "set_get_random"
}

func (f *SetGetRandomOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// PublishOperationFactory 发布操作工厂
type PublishOperationFactory struct{}

func (f *PublishOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 获取频道名称
	channel, ok := params["channel"].(string)
	if !ok {
		channel = "my_channel"
	}

	// 生成消息内容
	value := strings.Repeat("X", dataSize)

	return interfaces.Operation{
		Type:  "pub",
		Key:   channel, // 使用Key字段存储频道名称
		Value: value,
		Params: map[string]interface{}{
			"channel": channel,
		},
	}, nil
}

func (f *PublishOperationFactory) GetOperationType() string {
	return "pub"
}

func (f *PublishOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

// HashSetOperationFactory HSET操作工厂
type HashSetOperationFactory struct{}

func (f *HashSetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 生成键和字段
	key := keyGen.GenerateKey("hash", 0)
	field := keyGen.GetRandomFromGenerated()
	if field == "" {
		field = "field1"
	}

	// 生成值
	value := strings.Repeat("X", dataSize)

	return interfaces.Operation{
		Type:  "hset",
		Key:   key,
		Value: value,
		Params: map[string]interface{}{
			"field": field,
		},
	}, nil
}

func (f *HashSetOperationFactory) GetOperationType() string {
	return "hset"
}

func (f *HashSetOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// HashGetOperationFactory HGET操作工厂
type HashGetOperationFactory struct{}

func (f *HashGetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 从已生成的键中选择
	key := keyGen.GetRandomFromGenerated()
	if key == "" {
		key = "default_hash_key"
	}

	field := "field1" // 默认字段

	return interfaces.Operation{
		Type: "hget",
		Key:  key,
		Params: map[string]interface{}{
			"field": field,
		},
	}, nil
}

func (f *HashGetOperationFactory) GetOperationType() string {
	return "hget"
}

func (f *HashGetOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// IncrOperationFactory INCR操作工厂
type IncrOperationFactory struct{}

func (f *IncrOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 生成键
	key := keyGen.GenerateKey("counter", 0)

	return interfaces.Operation{
		Type: "incr",
		Key:  key,
	}, nil
}

func (f *IncrOperationFactory) GetOperationType() string {
	return "incr"
}

func (f *IncrOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// DecrOperationFactory DECR操作工厂
type DecrOperationFactory struct{}

func (f *DecrOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 生成键
	key := keyGen.GenerateKey("counter", 0)

	return interfaces.Operation{
		Type: "decr",
		Key:  key,
	}, nil
}

func (f *DecrOperationFactory) GetOperationType() string {
	return "decr"
}

func (f *DecrOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	return nil
}

// ListOperationFactory 列表操作工厂
type ListOperationFactory struct {
	opType string
}

func (f *ListOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 生成键
	key := keyGen.GenerateKey("list", 0)

	// 生成值（仅对PUSH操作）
	var value string
	if f.opType == "lpush" || f.opType == "rpush" {
		value = strings.Repeat("X", dataSize)
	}

	return interfaces.Operation{
		Type:  f.opType,
		Key:   key,
		Value: value,
	}, nil
}

func (f *ListOperationFactory) GetOperationType() string {
	return f.opType
}

func (f *ListOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	
	// PUSH操作需要data_size参数
	if (f.opType == "lpush" || f.opType == "rpush") {
		if _, ok := params["data_size"]; !ok {
			return fmt.Errorf("data_size parameter is required for %s operation", f.opType)
		}
	}
	
	return nil
}

// SetOperationFactory 集合操作工厂
type SetOperationFactory struct {
	opType string
}

func (f *SetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 生成键
	key := keyGen.GenerateKey("set", 0)

	// 生成值（仅对SADD/SREM操作）
	var value string
	if f.opType == "sadd" || f.opType == "srem" {
		value = strings.Repeat("X", dataSize)
	}

	return interfaces.Operation{
		Type:  f.opType,
		Key:   key,
		Value: value,
	}, nil
}

func (f *SetOperationFactory) GetOperationType() string {
	return f.opType
}

func (f *SetOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	
	// SADD/SREM操作需要data_size参数
	if (f.opType == "sadd" || f.opType == "srem") {
		if _, ok := params["data_size"]; !ok {
			return fmt.Errorf("data_size parameter is required for %s operation", f.opType)
		}
	}
	
	return nil
}

// SortedSetOperationFactory 有序集合操作工厂
type SortedSetOperationFactory struct {
	opType string
}

func (f *SortedSetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 生成键
	key := keyGen.GenerateKey("zset", 0)

	// 生成值和分数（仅对ZADD操作）
	var value string
	var score float64
	if f.opType == "zadd" {
		value = strings.Repeat("X", dataSize)
		score = float64(time.Now().UnixNano() % 1000) // 生成随机分数
	}

	return interfaces.Operation{
		Type:  f.opType,
		Key:   key,
		Value: value,
		Params: map[string]interface{}{
			"score": score,
		},
	}, nil
}

func (f *SortedSetOperationFactory) GetOperationType() string {
	return f.opType
}

func (f *SortedSetOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	
	// ZADD操作需要data_size参数
	if f.opType == "zadd" {
		if _, ok := params["data_size"]; !ok {
			return fmt.Errorf("data_size parameter is required for %s operation", f.opType)
		}
	}
	
	return nil
}

// HashOperationFactory 哈希操作工厂（扩展）
type HashOperationFactory struct {
	opType string
}

func (f *HashOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	keyGen, ok := params["key_generator"].(*utils.DefaultKeyGenerator)
	if !ok {
		return interfaces.Operation{}, fmt.Errorf("key_generator is required")
	}

	// 获取数据大小
	dataSize, ok := params["data_size"].(int)
	if !ok {
		dataSize = 3
	}

	// 生成键
	key := keyGen.GenerateKey("hash", 0)

	// 生成字段和值
	field := keyGen.GetRandomFromGenerated()
	if field == "" {
		field = "field1"
	}
	
	value := strings.Repeat("X", dataSize)

	// 对于HMSET/HMGET，生成多个字段
	if f.opType == "hmset" || f.opType == "hmget" {
		field2 := keyGen.GetRandomFromGenerated()
		if field2 == "" {
			field2 = "field2"
		}
		
		return interfaces.Operation{
			Type:  f.opType,
			Key:   key,
			Value: value,
			Params: map[string]interface{}{
				"fields": []string{field, field2},
			},
		}, nil
	}

	return interfaces.Operation{
		Type:  f.opType,
		Key:   key,
		Value: value,
		Params: map[string]interface{}{
			"field": field,
		},
	}, nil
}

func (f *HashOperationFactory) GetOperationType() string {
	return f.opType
}

func (f *HashOperationFactory) ValidateParams(params map[string]interface{}) error {
	if _, exists := params["key_generator"]; !exists {
		return fmt.Errorf("key_generator parameter is required")
	}
	
	// HMSET/HSET操作需要data_size参数
	if f.opType == "hmset" || f.opType == "hset" {
		if _, ok := params["data_size"]; !ok {
			return fmt.Errorf("data_size parameter is required for %s operation", f.opType)
		}
	}
	
	return nil
}

// SubscribeOperationFactory 订阅操作工厂
type SubscribeOperationFactory struct{}

func (f *SubscribeOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 获取频道名称
	channel, ok := params["channel"].(string)
	if !ok {
		channel = "my_channel"
	}

	return interfaces.Operation{
		Type: "sub",
		Key:  channel, // 使用Key字段存储频道名称
		Params: map[string]interface{}{
			"channel": channel,
		},
	}, nil
}

func (f *SubscribeOperationFactory) GetOperationType() string {
	return "sub"
}

func (f *SubscribeOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

// RegisterRedisOperations 注册所有Redis操作
func RegisterRedisOperations(registry *utils.OperationRegistry) {
	registry.Register("get", &StringGetOperationFactory{})
	registry.Register("set", &StringSetOperationFactory{})
	registry.Register("del", &StringDeleteOperationFactory{})
	registry.Register("set_get_random", &SetGetRandomOperationFactory{})
	registry.Register("pub", &PublishOperationFactory{})
	registry.Register("hset", &HashSetOperationFactory{})
	registry.Register("hget", &HashGetOperationFactory{})
	
	// 注册新增的操作
	registry.Register("incr", &IncrOperationFactory{})
	registry.Register("decr", &DecrOperationFactory{})
	registry.Register("lpush", &ListOperationFactory{opType: "lpush"})
	registry.Register("rpush", &ListOperationFactory{opType: "rpush"})
	registry.Register("lpop", &ListOperationFactory{opType: "lpop"})
	registry.Register("rpop", &ListOperationFactory{opType: "rpop"})
	registry.Register("sadd", &SetOperationFactory{opType: "sadd"})
	registry.Register("smembers", &SetOperationFactory{opType: "smembers"})
	registry.Register("srem", &SetOperationFactory{opType: "srem"})
	registry.Register("sismember", &SetOperationFactory{opType: "sismember"})
	registry.Register("zadd", &SortedSetOperationFactory{opType: "zadd"})
	registry.Register("zrem", &SortedSetOperationFactory{opType: "zrem"})
	registry.Register("zrange", &SortedSetOperationFactory{opType: "zrange"})
	registry.Register("zrank", &SortedSetOperationFactory{opType: "zrank"})
	registry.Register("hmset", &HashOperationFactory{opType: "hmset"})
	registry.Register("hmget", &HashOperationFactory{opType: "hmget"})
	registry.Register("hgetall", &HashOperationFactory{opType: "hgetall"})
	registry.Register("sub", &SubscribeOperationFactory{})
}
