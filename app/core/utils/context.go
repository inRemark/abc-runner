package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// DefaultKeyGenerator 默认键生成器实现
type DefaultKeyGenerator struct {
	counter       int64
	generatedKeys []string
	mutex         sync.RWMutex
}

// NewDefaultKeyGenerator 创建默认键生成器
func NewDefaultKeyGenerator() *DefaultKeyGenerator {
	return &DefaultKeyGenerator{
		generatedKeys: make([]string, 0),
	}
}

// GenerateKey 生成递增键
func (g *DefaultKeyGenerator) GenerateKey(operationType string, index int64) string {
	keyNum := atomic.AddInt64(&g.counter, 1) - 1
	key := fmt.Sprintf("%s:i:%d", operationType, keyNum)

	g.mutex.Lock()
	g.generatedKeys = append(g.generatedKeys, key)
	g.mutex.Unlock()

	return key
}

// GenerateRandomKey 生成随机键
func (g *DefaultKeyGenerator) GenerateRandomKey(operationType string, maxRange int) string {
	if maxRange <= 0 {
		return g.GenerateKey(operationType, 0)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(maxRange)
	key := fmt.Sprintf("%s:r:%d", operationType, randomNum)

	g.mutex.Lock()
	g.generatedKeys = append(g.generatedKeys, key)
	g.mutex.Unlock()

	return key
}

// GetGeneratedKeys 获取已生成的键列表
func (g *DefaultKeyGenerator) GetGeneratedKeys() []string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	result := make([]string, len(g.generatedKeys))
	copy(result, g.generatedKeys)
	return result
}

// GetRandomFromGenerated 从已生成的键中随机选择一个
func (g *DefaultKeyGenerator) GetRandomFromGenerated() string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if len(g.generatedKeys) == 0 {
		return "default:key:0"
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return g.generatedKeys[r.Intn(len(g.generatedKeys))]
}

// Reset 重置键生成器
func (g *DefaultKeyGenerator) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.counter = 0
	g.generatedKeys = make([]string, 0)
}

// TestContextImpl 测试上下文实现
type TestContextImpl struct {
	adapter          interfaces.ProtocolAdapter
	config           interfaces.Config
	metricsCollector interfaces.MetricsCollector
	keyGenerator     interfaces.KeyGenerator
	cancelled        bool
	cancelMutex      sync.RWMutex
}

// NewTestContext 创建测试上下文
func NewTestContext(
	adapter interfaces.ProtocolAdapter,
	config interfaces.Config,
	metricsCollector interfaces.MetricsCollector,
	keyGenerator interfaces.KeyGenerator,
) *TestContextImpl {
	return &TestContextImpl{
		adapter:          adapter,
		config:           config,
		metricsCollector: metricsCollector,
		keyGenerator:     keyGenerator,
	}
}

// GetAdapter 获取协议适配器
func (t *TestContextImpl) GetAdapter() interfaces.ProtocolAdapter {
	return t.adapter
}

// GetConfig 获取配置
func (t *TestContextImpl) GetConfig() interfaces.Config {
	return t.config
}

// GetMetricsCollector 获取指标收集器
func (t *TestContextImpl) GetMetricsCollector() interfaces.MetricsCollector {
	return t.metricsCollector
}

// GetKeyGenerator 获取键生成器
func (t *TestContextImpl) GetKeyGenerator() interfaces.KeyGenerator {
	return t.keyGenerator
}

// Cancel 取消测试
func (t *TestContextImpl) Cancel() {
	t.cancelMutex.Lock()
	defer t.cancelMutex.Unlock()
	t.cancelled = true
}

// IsCancelled 检查是否已取消
func (t *TestContextImpl) IsCancelled() bool {
	t.cancelMutex.RLock()
	defer t.cancelMutex.RUnlock()
	return t.cancelled
}

// OperationRegistry 操作注册表
type OperationRegistry struct {
	factories map[string]interfaces.OperationFactory
	mutex     sync.RWMutex
}

// NewOperationRegistry 创建操作注册表
func NewOperationRegistry() *OperationRegistry {
	return &OperationRegistry{
		factories: make(map[string]interfaces.OperationFactory),
	}
}

// Register 注册操作工厂
func (r *OperationRegistry) Register(operationType string, factory interfaces.OperationFactory) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.factories[operationType] = factory
}

// GetFactory 获取操作工厂
func (r *OperationRegistry) GetFactory(operationType string) (interfaces.OperationFactory, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	factory, exists := r.factories[operationType]
	return factory, exists
}

// GetSupportedOperations 获取支持的操作类型
func (r *OperationRegistry) GetSupportedOperations() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	operations := make([]string, 0, len(r.factories))
	for opType := range r.factories {
		operations = append(operations, opType)
	}
	return operations
}

// CreateOperation 创建操作
func (r *OperationRegistry) CreateOperation(operationType string, params map[string]interface{}) (interfaces.Operation, error) {
	factory, exists := r.GetFactory(operationType)
	if !exists {
		return interfaces.Operation{}, fmt.Errorf("unsupported operation type: %s", operationType)
	}

	return factory.CreateOperation(params)
}

// ValidateOperation 验证操作参数
func (r *OperationRegistry) ValidateOperation(operationType string, params map[string]interface{}) error {
	factory, exists := r.GetFactory(operationType)
	if !exists {
		return fmt.Errorf("unsupported operation type: %s", operationType)
	}

	return factory.ValidateParams(params)
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"network",
			"broken pipe",
			"connection reset",
		},
	}
}

// IsRetryableError 检查错误是否可重试
func (r *RetryConfig) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, retryableErr := range r.RetryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}
	return false
}

// GetDelay 获取重试延迟
func (r *RetryConfig) GetDelay(attempt int) time.Duration {
	delay := float64(r.InitialDelay) * math.Pow(r.BackoffFactor, float64(attempt))
	if delay > float64(r.MaxDelay) {
		delay = float64(r.MaxDelay)
	}
	return time.Duration(delay)
}

// contains 检查字符串是否包含子串
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	total     int64
	current   int64
	startTime time.Time
	mutex     sync.RWMutex
}

// NewProgressTracker 创建进度跟踪器
func NewProgressTracker(total int64) *ProgressTracker {
	return &ProgressTracker{
		total:     total,
		startTime: time.Now(),
	}
}

// Update 更新进度
func (p *ProgressTracker) Update(increment int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.current += increment
}

// GetProgress 获取进度信息
func (p *ProgressTracker) GetProgress() (current, total int64, percentage float64, eta time.Duration) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	current = p.current
	total = p.total

	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}

	if current > 0 {
		elapsed := time.Since(p.startTime)
		rate := float64(current) / elapsed.Seconds()
		if rate > 0 {
			remaining := float64(total - current)
			eta = time.Duration(remaining/rate) * time.Second
		}
	}

	return
}

// IsCompleted 检查是否完成
func (p *ProgressTracker) IsCompleted() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.current >= p.total
}
