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