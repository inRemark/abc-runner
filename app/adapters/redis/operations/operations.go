package operations

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"redis-runner/app/adapters/redis/config"
)

// OperationType 操作类型常量
type OperationType string

const (
	OperationGet           OperationType = "get"
	OperationSet           OperationType = "set"
	OperationSetGetRandom  OperationType = "set_get_random"
	OperationDelete        OperationType = "delete"
	OperationPublish       OperationType = "pub"
	OperationSubscribe     OperationType = "sub"
	OperationHSet          OperationType = "hset"
	OperationHGet          OperationType = "hget"
	// 新增的操作类型
	OperationIncr          OperationType = "incr"
	OperationDecr          OperationType = "decr"
	OperationLPush         OperationType = "lpush"
	OperationRPush         OperationType = "rpush"
	OperationLPop          OperationType = "lpop"
	OperationRPop          OperationType = "rpop"
	OperationSAdd          OperationType = "sadd"
	OperationSRem          OperationType = "srem"
	OperationSMembers      OperationType = "smembers"
	OperationSIsMember     OperationType = "sismember"
	OperationZAdd          OperationType = "zadd"
	OperationZRem          OperationType = "zrem"
	OperationZRange        OperationType = "zrange"
	OperationZRank         OperationType = "zrank"
	OperationHMSet         OperationType = "hmset"
	OperationHMGet         OperationType = "hmget"
	OperationHGetAll       OperationType = "hgetall"
	OperationUnsubscribe   OperationType = "unsub"
)

// OperationResult 操作结果
type OperationResult struct {
	Success   bool
	IsRead    bool
	Duration  time.Duration
	Error     error
	Value     interface{}
	ExtraData map[string]interface{}
}

// OperationParams 操作参数
type OperationParams struct {
	Key         string
	Value       interface{}
	TTL         time.Duration
	Database    int
	Total       int
	RandomKeys  int
	DataSize    int
	ReadPercent int
	ExtraArgs   map[string]interface{}
}

// Operation 操作接口
type Operation interface {
	Execute(client redis.Cmdable, params OperationParams) OperationResult
	GetType() OperationType
	Validate(params OperationParams) error
}

// KeyGenerator 键生成器
type KeyGenerator struct {
	globalCounter int64
	genKeys       []string
	mutex         sync.RWMutex
}

// NewKeyGenerator 创建键生成器
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{
		genKeys: make([]string, 0),
	}
}

// GenerateKey 生成键
func (kg *KeyGenerator) GenerateKey(randomRange, total int) string {
	if randomRange == 0 {
		// 全局自增键模式
		keyNum := atomic.AddInt64(&kg.globalCounter, 1) - 1
		return "i:" + strconv.FormatInt(keyNum, 10)
	}
	// 随机键模式
	r := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(randomRange)
	return "r:" + strconv.Itoa(r)
}

// AddGeneratedKey 添加生成的键
func (kg *KeyGenerator) AddGeneratedKey(key string) {
	kg.mutex.Lock()
	defer kg.mutex.Unlock()
	kg.genKeys = append(kg.genKeys, key)
}

// GetRandomGeneratedKey 获取随机生成的键
func (kg *KeyGenerator) GetRandomGeneratedKey() string {
	kg.mutex.RLock()
	defer kg.mutex.RUnlock()
	
	if len(kg.genKeys) == 0 {
		return "r:0"
	}
	
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return kg.genKeys[r.Intn(len(kg.genKeys))]
}

// GetOperation Get操作实现
type GetOperation struct {
	keyGen *KeyGenerator
}

// NewGetOperation 创建Get操作
func NewGetOperation() *GetOperation {
	return &GetOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行Get操作
func (op *GetOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GetRandomGeneratedKey()
	
	start := time.Now()
	val, err := client.Get(ctx, key).Result()
	duration := time.Since(start)
	
	result := OperationResult{
		Success:  err == nil || err == redis.Nil,
		IsRead:   true,
		Duration: duration,
		Error:    err,
		Value:    val,
	}
	
	if err == redis.Nil {
		result.Error = nil // Key不存在不算错误
	}
	
	return result
}

// GetType 获取操作类型
func (op *GetOperation) GetType() OperationType {
	return OperationGet
}

// Validate 验证参数
func (op *GetOperation) Validate(params OperationParams) error {
	return nil
}

// SetOperation Set操作实现
type SetOperation struct {
	keyGen *KeyGenerator
}

// NewSetOperation 创建Set操作
func NewSetOperation() *SetOperation {
	return &SetOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行Set操作
func (op *SetOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	err := client.Set(ctx, key, value, params.TTL).Err()
	duration := time.Since(start)
	
	if err == nil {
		op.keyGen.AddGeneratedKey(key)
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Value:    nil,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *SetOperation) GetType() OperationType {
	return OperationSet
}

// Validate 验证参数
func (op *SetOperation) Validate(params OperationParams) error {
	if params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive")
	}
	return nil
}

// SetGetRandomOperation Set/Get随机操作实现
type SetGetRandomOperation struct {
	keyGen    *KeyGenerator
	setOp     *SetOperation
	getOp     *GetOperation
}

// NewSetGetRandomOperation 创建Set/Get随机操作
func NewSetGetRandomOperation() *SetGetRandomOperation {
	keyGen := NewKeyGenerator()
	return &SetGetRandomOperation{
		keyGen: keyGen,
		setOp:  &SetOperation{keyGen: keyGen},
		getOp:  &GetOperation{keyGen: keyGen},
	}
}

// Execute 执行Set/Get随机操作
func (op *SetGetRandomOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// 根据读写比例决定操作类型
	if rd.Intn(100) < params.ReadPercent {
		// 执行读操作
		result := op.getOp.Execute(client, params)
		result.IsRead = true
		return result
	} else {
		// 执行写操作
		result := op.setOp.Execute(client, params)
		result.IsRead = false
		return result
	}
}

// GetType 获取操作类型
func (op *SetGetRandomOperation) GetType() OperationType {
	return OperationSetGetRandom
}

// Validate 验证参数
func (op *SetGetRandomOperation) Validate(params OperationParams) error {
	if params.ReadPercent < 0 || params.ReadPercent > 100 {
		return fmt.Errorf("read percent must be between 0 and 100")
	}
	if params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive")
	}
	return nil
}

// DeleteOperation Delete操作实现
type DeleteOperation struct {
	keyGen *KeyGenerator
}

// NewDeleteOperation 创建Delete操作
func NewDeleteOperation() *DeleteOperation {
	return &DeleteOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行Delete操作
func (op *DeleteOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	
	start := time.Now()
	result, err := client.Del(ctx, key).Result()
	duration := time.Since(start)
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Value:    result,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *DeleteOperation) GetType() OperationType {
	return OperationDelete
}

// Validate 验证参数
func (op *DeleteOperation) Validate(params OperationParams) error {
	return nil
}

// PublishOperation Publish操作实现
type PublishOperation struct {
	channel string
}

// NewPublishOperation 创建Publish操作
func NewPublishOperation(channel string) *PublishOperation {
	if channel == "" {
		channel = "my_channel"
	}
	return &PublishOperation{
		channel: channel,
	}
}

// Execute 执行Publish操作
func (op *PublishOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	err := client.Publish(ctx, op.channel, value).Err()
	duration := time.Since(start)
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Value:    nil,
		ExtraData: map[string]interface{}{
			"channel": op.channel,
			"message": value,
		},
	}
}

// GetType 获取操作类型
func (op *PublishOperation) GetType() OperationType {
	return OperationPublish
}

// Validate 验证参数
func (op *PublishOperation) Validate(params OperationParams) error {
	if params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive")
	}
	return nil
}

// HSetOperation HSet操作实现
type HSetOperation struct {
	keyGen *KeyGenerator
}

// NewHSetOperation 创建HSet操作
func NewHSetOperation() *HSetOperation {
	return &HSetOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行HSet操作
func (op *HSetOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := params.Key
	if key == "" {
		key = op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	}
	
	field := op.keyGen.GetRandomGeneratedKey()
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	err := client.HSet(ctx, key, field, value).Err()
	duration := time.Since(start)
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Value:    nil,
		ExtraData: map[string]interface{}{
			"key":   key,
			"field": field,
		},
	}
}

// GetType 获取操作类型
func (op *HSetOperation) GetType() OperationType {
	return OperationHSet
}

// Validate 验证参数
func (op *HSetOperation) Validate(params OperationParams) error {
	if params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive")
	}
	return nil
}

// HGetOperation HGet操作实现
type HGetOperation struct {
	keyGen *KeyGenerator
}

// NewHGetOperation 创建HGet操作
func NewHGetOperation() *HGetOperation {
	return &HGetOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行HGet操作
func (op *HGetOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := params.Key
	if key == "" {
		key = op.keyGen.GetRandomGeneratedKey()
	}
	
	field := op.keyGen.GetRandomGeneratedKey()
	
	start := time.Now()
	val, err := client.HGet(ctx, key, field).Result()
	duration := time.Since(start)
	
	result := OperationResult{
		Success:  err == nil || err == redis.Nil,
		IsRead:   true,
		Duration: duration,
		Error:    err,
		Value:    val,
		ExtraData: map[string]interface{}{
			"key":   key,
			"field": field,
		},
	}
	
	if err == redis.Nil {
		result.Error = nil // Key或Field不存在不算错误
	}
	
	return result
}

// GetType 获取操作类型
func (op *HGetOperation) GetType() OperationType {
	return OperationHGet
}

// Validate 验证参数
func (op *HGetOperation) Validate(params OperationParams) error {
	return nil
}

// IncrOperation Incr操作实现
type IncrOperation struct {
	keyGen *KeyGenerator
}

// NewIncrOperation 创建Incr操作
func NewIncrOperation() *IncrOperation {
	return &IncrOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行Incr操作
func (op *IncrOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	
	start := time.Now()
	val, err := client.Incr(ctx, key).Result()
	duration := time.Since(start)
	
	if err == nil {
		op.keyGen.AddGeneratedKey(key)
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Value:    val,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *IncrOperation) GetType() OperationType {
	return OperationType("incr")
}

// Validate 验证参数
func (op *IncrOperation) Validate(params OperationParams) error {
	return nil
}

// DecrOperation Decr操作实现
type DecrOperation struct {
	keyGen *KeyGenerator
}

// NewDecrOperation 创建Decr操作
func NewDecrOperation() *DecrOperation {
	return &DecrOperation{
		keyGen: NewKeyGenerator(),
	}
}

// Execute 执行Decr操作
func (op *DecrOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	
	start := time.Now()
	val, err := client.Decr(ctx, key).Result()
	duration := time.Since(start)
	
	if err == nil {
		op.keyGen.AddGeneratedKey(key)
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Value:    val,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *DecrOperation) GetType() OperationType {
	return OperationType("decr")
}

// Validate 验证参数
func (op *DecrOperation) Validate(params OperationParams) error {
	return nil
}

// ListOperation 列表操作实现
type ListOperation struct {
	keyGen *KeyGenerator
	opType OperationType
}

// NewListOperation 创建列表操作
func NewListOperation(opType OperationType) *ListOperation {
	return &ListOperation{
		keyGen: NewKeyGenerator(),
		opType: opType,
	}
}

// Execute 执行列表操作
func (op *ListOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	var err error
	var result interface{}
	
	switch op.opType {
	case OperationType("lpush"):
		result, err = client.LPush(ctx, key, value).Result()
	case OperationType("rpush"):
		result, err = client.RPush(ctx, key, value).Result()
	case OperationType("lpop"):
		result, err = client.LPop(ctx, key).Result()
	case OperationType("rpop"):
		result, err = client.RPop(ctx, key).Result()
	default:
		return OperationResult{
			Success: false,
			Error:   fmt.Errorf("unsupported list operation: %s", op.opType),
		}
	}
	
	duration := time.Since(start)
	
	if err == nil && (op.opType == OperationType("lpush") || op.opType == OperationType("rpush")) {
		op.keyGen.AddGeneratedKey(key)
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   op.opType == OperationType("lpop") || op.opType == OperationType("rpop"),
		Duration: duration,
		Error:    err,
		Value:    result,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *ListOperation) GetType() OperationType {
	return op.opType
}

// Validate 验证参数
func (op *ListOperation) Validate(params OperationParams) error {
	if (op.opType == OperationType("lpush") || op.opType == OperationType("rpush")) && params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive for push operations")
	}
	return nil
}

// RedisSetOperation 集合操作实现
type RedisSetOperation struct {
	keyGen *KeyGenerator
	opType OperationType
}

// NewRedisSetOperation 创建集合操作
func NewRedisSetOperation(opType OperationType) *RedisSetOperation {
	return &RedisSetOperation{
		keyGen: NewKeyGenerator(),
		opType: opType,
	}
}

// Execute 执行集合操作
func (op *RedisSetOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	var err error
	var result interface{}
	
	switch op.opType {
	case OperationType("sadd"):
		result, err = client.SAdd(ctx, key, value).Result()
	case OperationType("smembers"):
		result, err = client.SMembers(ctx, key).Result()
	case OperationType("srem"):
		result, err = client.SRem(ctx, key, value).Result()
	case OperationType("sismember"):
		result, err = client.SIsMember(ctx, key, value).Result()
	default:
		return OperationResult{
			Success: false,
			Error:   fmt.Errorf("unsupported set operation: %s", op.opType),
		}
	}
	
	duration := time.Since(start)
	
	if err == nil && op.opType == OperationType("sadd") {
		op.keyGen.AddGeneratedKey(key)
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   op.opType == OperationType("smembers") || op.opType == OperationType("sismember"),
		Duration: duration,
		Error:    err,
		Value:    result,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *RedisSetOperation) GetType() OperationType {
	return op.opType
}

// Validate 验证参数
func (op *RedisSetOperation) Validate(params OperationParams) error {
	if (op.opType == OperationType("sadd") || op.opType == OperationType("srem")) && params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive for sadd and srem operations")
	}
	return nil
}

// SortedSetOperation 有序集合操作实现
type SortedSetOperation struct {
	keyGen *KeyGenerator
	opType OperationType
}

// NewSortedSetOperation 创建有序集合操作
func NewSortedSetOperation(opType OperationType) *SortedSetOperation {
	return &SortedSetOperation{
		keyGen: NewKeyGenerator(),
		opType: opType,
	}
}

// Execute 执行有序集合操作
func (op *SortedSetOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	var err error
	var result interface{}
	
	// 生成一个随机分数用于有序集合操作
	score := float64(rand.Intn(1000))
	
	switch op.opType {
	case OperationType("zadd"):
		result, err = client.ZAdd(ctx, key, &redis.Z{Score: score, Member: value}).Result()
	case OperationType("zrem"):
		result, err = client.ZRem(ctx, key, value).Result()
	case OperationType("zrange"):
		result, err = client.ZRange(ctx, key, 0, -1).Result()
	case OperationType("zrank"):
		result, err = client.ZRank(ctx, key, value).Result()
	default:
		return OperationResult{
			Success: false,
			Error:   fmt.Errorf("unsupported sorted set operation: %s", op.opType),
		}
	}
	
	duration := time.Since(start)
	
	if err == nil && (op.opType == OperationType("zadd")) {
		op.keyGen.AddGeneratedKey(key)
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   op.opType == OperationType("zrange") || op.opType == OperationType("zrank"),
		Duration: duration,
		Error:    err,
		Value:    result,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *SortedSetOperation) GetType() OperationType {
	return op.opType
}

// Validate 验证参数
func (op *SortedSetOperation) Validate(params OperationParams) error {
	if (op.opType == OperationType("zadd") || op.opType == OperationType("zrem")) && params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive for zadd and zrem operations")
	}
	return nil
}

// HashOperation 哈希操作实现
type HashOperation struct {
	keyGen *KeyGenerator
	opType OperationType
}

// NewHashOperation 创建哈希操作
func NewHashOperation(opType OperationType) *HashOperation {
	return &HashOperation{
		keyGen: NewKeyGenerator(),
		opType: opType,
	}
}

// Execute 执行哈希操作
func (op *HashOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	ctx := context.Background()
	key := op.keyGen.GenerateKey(params.RandomKeys, params.Total)
	value := strings.Repeat("X", params.DataSize)
	
	start := time.Now()
	var err error
	var result interface{}
	
	// 生成多个字段用于哈希操作
	field1 := op.keyGen.GetRandomGeneratedKey()
	field2 := op.keyGen.GetRandomGeneratedKey()
	
	switch op.opType {
	case OperationType("hset"):
		result, err = client.HSet(ctx, key, field1, value).Result()
	case OperationType("hget"):
		result, err = client.HGet(ctx, key, field1).Result()
	case OperationType("hmset"):
		fields := map[string]interface{}{
			field1: value,
			field2: value + "_2",
		}
		result, err = client.HMSet(ctx, key, fields).Result()
	case OperationType("hmget"):
		result, err = client.HMGet(ctx, key, field1, field2).Result()
	case OperationType("hgetall"):
		result, err = client.HGetAll(ctx, key).Result()
	default:
		return OperationResult{
			Success: false,
			Error:   fmt.Errorf("unsupported hash operation: %s", op.opType),
		}
	}
	
	duration := time.Since(start)
	
	if err == nil && (op.opType == OperationType("hset") || op.opType == OperationType("hmset")) {
		op.keyGen.AddGeneratedKey(key)
	}
	
	// 处理HMSET的特殊情况，它返回bool而不是int
	if op.opType == OperationType("hmset") {
		if b, ok := result.(bool); ok && b {
			result = int64(1) // 转换为int64以保持一致性
		}
	}
	
	return OperationResult{
		Success:  err == nil,
		IsRead:   op.opType == OperationType("hget") || op.opType == OperationType("hmget") || op.opType == OperationType("hgetall"),
		Duration: duration,
		Error:    err,
		Value:    result,
		ExtraData: map[string]interface{}{
			"key": key,
		},
	}
}

// GetType 获取操作类型
func (op *HashOperation) GetType() OperationType {
	return op.opType
}

// Validate 验证参数
func (op *HashOperation) Validate(params OperationParams) error {
	if (op.opType == OperationType("hset") || op.opType == OperationType("hmset")) && params.DataSize <= 0 {
		return fmt.Errorf("data size must be positive for hset and hmset operations")
	}
	return nil
}

// SubscribeOperation 订阅操作实现
type SubscribeOperation struct {
	channel string
}

// NewSubscribeOperation 创建订阅操作
func NewSubscribeOperation(channel string) *SubscribeOperation {
	if channel == "" {
		channel = "my_channel"
	}
	return &SubscribeOperation{
		channel: channel,
	}
}

// Execute 执行订阅操作
func (op *SubscribeOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	// 注意：订阅操作是阻塞的，我们需要特殊处理
	// 这里我们只是模拟订阅操作，实际的订阅应该在单独的goroutine中处理
	start := time.Now()
	
	// 模拟订阅操作的延迟
	time.Sleep(10 * time.Millisecond)
	
	duration := time.Since(start)
	
	return OperationResult{
		Success:  true,
		IsRead:   true,
		Duration: duration,
		Error:    nil,
		Value:    nil,
		ExtraData: map[string]interface{}{
			"channel": op.channel,
		},
	}
}

// GetType 获取操作类型
func (op *SubscribeOperation) GetType() OperationType {
	return OperationSubscribe
}

// Validate 验证参数
func (op *SubscribeOperation) Validate(params OperationParams) error {
	return nil
}

// UnsubscribeOperation 取消订阅操作实现
type UnsubscribeOperation struct {
	channel string
}

// NewUnsubscribeOperation 创建取消订阅操作
func NewUnsubscribeOperation(channel string) *UnsubscribeOperation {
	if channel == "" {
		channel = "my_channel"
	}
	return &UnsubscribeOperation{
		channel: channel,
	}
}

// Execute 执行取消订阅操作
func (op *UnsubscribeOperation) Execute(client redis.Cmdable, params OperationParams) OperationResult {
	start := time.Now()
	
	// 模拟取消订阅操作的延迟
	time.Sleep(5 * time.Millisecond)
	
	duration := time.Since(start)
	
	return OperationResult{
		Success:  true,
		IsRead:   false,
		Duration: duration,
		Error:    nil,
		Value:    nil,
		ExtraData: map[string]interface{}{
			"channel": op.channel,
		},
	}
}

// GetType 获取操作类型
func (op *UnsubscribeOperation) GetType() OperationType {
	return OperationType("unsub")
}

// Validate 验证参数
func (op *UnsubscribeOperation) Validate(params OperationParams) error {
	return nil
}

// OperationExecutor 操作执行器
type OperationExecutor struct {
	client redis.Cmdable
	config *config.RedisConfig
}

// NewOperationExecutor 创建操作执行器
func NewOperationExecutor(client redis.Cmdable, cfg *config.RedisConfig) *OperationExecutor {
	return &OperationExecutor{
		client: client,
		config: cfg,
	}
}

// ExecuteOperation 执行操作
func (executor *OperationExecutor) ExecuteOperation(operation Operation, params OperationParams) OperationResult {
	// 验证参数
	if err := operation.Validate(params); err != nil {
		return OperationResult{
			Success: false,
			Error:   fmt.Errorf("parameter validation failed: %w", err),
		}
	}

	// 执行操作
	return operation.Execute(executor.client, params)
}

// CreateOperationParams 创建操作参数
func CreateOperationParams(cfg *config.RedisConfig) OperationParams {
	benchmark := cfg.GetBenchmark()
	
	return OperationParams{
		Total:       benchmark.GetTotal(),
		RandomKeys:  benchmark.GetRandomKeys(),
		DataSize:    benchmark.GetDataSize(),
		ReadPercent: benchmark.GetReadPercent(),
		TTL:         benchmark.GetTTL(),
		ExtraArgs:   make(map[string]interface{}),
	}
}