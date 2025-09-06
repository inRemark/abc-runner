package operations

import (
	"fmt"
	"sync"
)

// OperationFactory 操作工厂
type OperationFactory struct {
	operations map[OperationType]func() Operation
	mutex      sync.RWMutex
}

// OperationRegistry 操作注册器接口
type OperationRegistry interface {
	Register(operationType OperationType, factory func() Operation)
	Create(operationType OperationType) (Operation, error)
	ListSupportedOperations() []OperationType
	IsSupported(operationType OperationType) bool
}

// NewOperationFactory 创建操作工厂
func NewOperationFactory() *OperationFactory {
	factory := &OperationFactory{
		operations: make(map[OperationType]func() Operation),
	}

	// 注册默认操作
	factory.registerDefaultOperations()

	return factory
}

// registerDefaultOperations 注册默认操作
func (f *OperationFactory) registerDefaultOperations() {
	f.Register(OperationGet, func() Operation {
		return NewGetOperation()
	})

	f.Register(OperationSet, func() Operation {
		return NewSetOperation()
	})

	f.Register(OperationSetGetRandom, func() Operation {
		return NewSetGetRandomOperation()
	})

	f.Register(OperationDelete, func() Operation {
		return NewDeleteOperation()
	})

	f.Register(OperationPublish, func() Operation {
		return NewPublishOperation("")
	})

	f.Register(OperationHSet, func() Operation {
		return NewHSetOperation()
	})

	f.Register(OperationHGet, func() Operation {
		return NewHGetOperation()
	})
}

// Register 注册操作
func (f *OperationFactory) Register(operationType OperationType, factory func() Operation) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.operations[operationType] = factory
}

// Create 创建操作实例
func (f *OperationFactory) Create(operationType OperationType) (Operation, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	factoryFunc, exists := f.operations[operationType]
	if !exists {
		return nil, fmt.Errorf("unsupported operation type: %s", operationType)
	}

	return factoryFunc(), nil
}

// ListSupportedOperations 列出支持的操作类型
func (f *OperationFactory) ListSupportedOperations() []OperationType {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	var operations []OperationType
	for opType := range f.operations {
		operations = append(operations, opType)
	}

	return operations
}

// IsSupported 检查是否支持指定的操作类型
func (f *OperationFactory) IsSupported(operationType OperationType) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	_, exists := f.operations[operationType]
	return exists
}

// CreateOperationFromString 从字符串创建操作
func (f *OperationFactory) CreateOperationFromString(operationStr string) (Operation, error) {
	operationType := OperationType(operationStr)
	return f.Create(operationType)
}

// OperationBuilder 操作构建器
type OperationBuilder struct {
	factory *OperationFactory
}

// NewOperationBuilder 创建操作构建器
func NewOperationBuilder() *OperationBuilder {
	return &OperationBuilder{
		factory: NewOperationFactory(),
	}
}

// WithCustomOperation 添加自定义操作
func (b *OperationBuilder) WithCustomOperation(operationType OperationType, factory func() Operation) *OperationBuilder {
	b.factory.Register(operationType, factory)
	return b
}

// WithPublishChannel 设置发布操作的频道
func (b *OperationBuilder) WithPublishChannel(channel string) *OperationBuilder {
	b.factory.Register(OperationPublish, func() Operation {
		return NewPublishOperation(channel)
	})
	return b
}

// Build 构建操作工厂
func (b *OperationBuilder) Build() *OperationFactory {
	return b.factory
}

// OperationManager 操作管理器
type OperationManager struct {
	factory   *OperationFactory
	instances map[OperationType]Operation
	mutex     sync.RWMutex
}

// NewOperationManager 创建操作管理器
func NewOperationManager() *OperationManager {
	return &OperationManager{
		factory:   NewOperationFactory(),
		instances: make(map[OperationType]Operation),
	}
}

// GetOperation 获取操作实例（单例模式）
func (m *OperationManager) GetOperation(operationType OperationType) (Operation, error) {
	m.mutex.RLock()
	if instance, exists := m.instances[operationType]; exists {
		m.mutex.RUnlock()
		return instance, nil
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 双重检查锁定
	if instance, exists := m.instances[operationType]; exists {
		return instance, nil
	}

	// 创建新实例
	instance, err := m.factory.Create(operationType)
	if err != nil {
		return nil, err
	}

	m.instances[operationType] = instance
	return instance, nil
}

// CreateNewOperation 创建新的操作实例（每次创建新实例）
func (m *OperationManager) CreateNewOperation(operationType OperationType) (Operation, error) {
	return m.factory.Create(operationType)
}

// RegisterCustomOperation 注册自定义操作
func (m *OperationManager) RegisterCustomOperation(operationType OperationType, factory func() Operation) {
	m.factory.Register(operationType, factory)
}

// GetSupportedOperations 获取支持的操作类型
func (m *OperationManager) GetSupportedOperations() []OperationType {
	return m.factory.ListSupportedOperations()
}

// ValidateOperationType 验证操作类型
func (m *OperationManager) ValidateOperationType(operationType OperationType) error {
	if !m.factory.IsSupported(operationType) {
		supportedOps := m.factory.ListSupportedOperations()
		return fmt.Errorf("unsupported operation type '%s'. Supported operations: %v", operationType, supportedOps)
	}
	return nil
}

// BatchOperationFactory 批量操作工厂
type BatchOperationFactory struct {
	factory *OperationFactory
}

// NewBatchOperationFactory 创建批量操作工厂
func NewBatchOperationFactory() *BatchOperationFactory {
	return &BatchOperationFactory{
		factory: NewOperationFactory(),
	}
}

// CreateBatch 创建批量操作
func (b *BatchOperationFactory) CreateBatch(operationTypes []OperationType) ([]Operation, error) {
	var operations []Operation

	for _, opType := range operationTypes {
		op, err := b.factory.Create(opType)
		if err != nil {
			return nil, fmt.Errorf("failed to create operation %s: %w", opType, err)
		}
		operations = append(operations, op)
	}

	return operations, nil
}

// CreateMixed 创建混合操作（用于复杂测试场景）
func (b *BatchOperationFactory) CreateMixed(operationWeights map[OperationType]int) ([]Operation, error) {
	var operations []Operation

	for opType, weight := range operationWeights {
		for i := 0; i < weight; i++ {
			op, err := b.factory.Create(opType)
			if err != nil {
				return nil, fmt.Errorf("failed to create operation %s: %w", opType, err)
			}
			operations = append(operations, op)
		}
	}

	return operations, nil
}

// OperationChain 操作链
type OperationChain struct {
	operations []Operation
	current    int
}

// NewOperationChain 创建操作链
func NewOperationChain(operations []Operation) *OperationChain {
	return &OperationChain{
		operations: operations,
		current:    0,
	}
}

// Next 获取下一个操作
func (c *OperationChain) Next() (Operation, bool) {
	if c.current >= len(c.operations) {
		return nil, false
	}

	op := c.operations[c.current]
	c.current++
	return op, true
}

// Reset 重置操作链
func (c *OperationChain) Reset() {
	c.current = 0
}

// HasNext 检查是否有下一个操作
func (c *OperationChain) HasNext() bool {
	return c.current < len(c.operations)
}

// Length 获取操作链长度
func (c *OperationChain) Length() int {
	return len(c.operations)
}

// GetGlobalOperationFactory 获取全局操作工厂实例
var globalOperationFactory *OperationFactory
var globalFactoryOnce sync.Once

func GetGlobalOperationFactory() *OperationFactory {
	globalFactoryOnce.Do(func() {
		globalOperationFactory = NewOperationFactory()
	})
	return globalOperationFactory
}

// RegisterGlobalOperation 注册全局操作
func RegisterGlobalOperation(operationType OperationType, factory func() Operation) {
	GetGlobalOperationFactory().Register(operationType, factory)
}

// CreateGlobalOperation 创建全局操作
func CreateGlobalOperation(operationType OperationType) (Operation, error) {
	return GetGlobalOperationFactory().Create(operationType)
}

// OperationTypeFromString 从字符串转换为操作类型
func OperationTypeFromString(s string) (OperationType, error) {
	operationType := OperationType(s)

	// 验证操作类型是否有效
	validTypes := []OperationType{
		OperationGet,
		OperationSet,
		OperationSetGetRandom,
		OperationDelete,
		OperationPublish,
		OperationSubscribe,
		OperationHSet,
		OperationHGet,
	}

	for _, validType := range validTypes {
		if operationType == validType {
			return operationType, nil
		}
	}

	return "", fmt.Errorf("invalid operation type: %s", s)
}

// GetAllOperationTypes 获取所有操作类型
func GetAllOperationTypes() []OperationType {
	return []OperationType{
		OperationGet,
		OperationSet,
		OperationSetGetRandom,
		OperationDelete,
		OperationPublish,
		OperationSubscribe,
		OperationHSet,
		OperationHGet,
	}
}
