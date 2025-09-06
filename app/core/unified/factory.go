package unified

import (
	"fmt"
	"log"
	"sync"

	"redis-runner/app/core/interfaces"
)

// protocolAdapterFactory 协议适配器工厂实现
type protocolAdapterFactory struct {
	creators map[string]AdapterCreator
	mutex    sync.RWMutex
}

// NewProtocolAdapterFactory 创建协议适配器工厂
func NewProtocolAdapterFactory() ProtocolAdapterFactory {
	factory := &protocolAdapterFactory{
		creators: make(map[string]AdapterCreator),
		mutex:    sync.RWMutex{},
	}
	
	// 注册默认的适配器创建函数
	factory.registerDefaultCreators()
	
	return factory
}

// CreateAdapter 创建协议适配器
func (f *protocolAdapterFactory) CreateAdapter(protocol string, config interfaces.Config) (interfaces.ProtocolAdapter, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	creator, exists := f.creators[protocol]
	if !exists {
		return nil, fmt.Errorf("protocol '%s' not supported", protocol)
	}
	
	// 验证配置
	if err := f.ValidateConfig(protocol, config); err != nil {
		return nil, fmt.Errorf("invalid config for protocol '%s': %w", protocol, err)
	}
	
	// 创建适配器
	adapter, err := creator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter for protocol '%s': %w", protocol, err)
	}
	
	log.Printf("Created adapter for protocol: %s", protocol)
	return adapter, nil
}

// RegisterAdapter 注册适配器创建函数
func (f *protocolAdapterFactory) RegisterAdapter(protocol string, creator AdapterCreator) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if _, exists := f.creators[protocol]; exists {
		return fmt.Errorf("protocol '%s' already registered", protocol)
	}
	
	f.creators[protocol] = creator
	log.Printf("Registered adapter creator for protocol: %s", protocol)
	return nil
}

// GetSupportedProtocols 获取支持的协议列表
func (f *protocolAdapterFactory) GetSupportedProtocols() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	protocols := make([]string, 0, len(f.creators))
	for protocol := range f.creators {
		protocols = append(protocols, protocol)
	}
	
	return protocols
}

// ValidateConfig 验证协议配置
func (f *protocolAdapterFactory) ValidateConfig(protocol string, config interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// 验证协议匹配
	if config.GetProtocol() != "" && config.GetProtocol() != protocol {
		return fmt.Errorf("protocol mismatch: expected '%s', got '%s'", protocol, config.GetProtocol())
	}
	
	// 调用配置自身的验证方法
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	
	// 协议特定的验证
	switch protocol {
	case "redis":
		return f.validateRedisConfig(config)
	case "http":
		return f.validateHttpConfig(config)
	case "kafka":
		return f.validateKafkaConfig(config)
	default:
		// 对于未知协议，只进行基础验证
		return nil
	}
}

// registerDefaultCreators 注册默认的适配器创建函数
func (f *protocolAdapterFactory) registerDefaultCreators() {
	// Redis适配器创建函数
	f.creators["redis"] = func(config interfaces.Config) (interfaces.ProtocolAdapter, error) {
		// 这里应该引用实际的Redis适配器实现
		// 暂时返回一个mock实现
		return &MockRedisAdapter{config: config}, nil
	}
	
	// HTTP适配器创建函数
	f.creators["http"] = func(config interfaces.Config) (interfaces.ProtocolAdapter, error) {
		// 这里应该引用实际的HTTP适配器实现
		// 暂时返回一个mock实现
		return &MockHttpAdapter{config: config}, nil
	}
	
	// Kafka适配器创建函数
	f.creators["kafka"] = func(config interfaces.Config) (interfaces.ProtocolAdapter, error) {
		// 这里应该引用实际的Kafka适配器实现
		// 暂时返回一个mock实现
		return &MockKafkaAdapter{config: config}, nil
	}
	
	log.Println("Default adapter creators registered")
}

// validateRedisConfig 验证Redis配置
func (f *protocolAdapterFactory) validateRedisConfig(config interfaces.Config) error {
	connConfig := config.GetConnection()
	if connConfig == nil {
		return fmt.Errorf("connection config required for Redis")
	}
	
	addresses := connConfig.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one Redis address required")
	}
	
	// 验证地址格式
	for _, addr := range addresses {
		if addr == "" {
			return fmt.Errorf("empty Redis address not allowed")
		}
		// 这里可以添加更详细的地址格式验证
	}
	
	return nil
}

// validateHttpConfig 验证HTTP配置
func (f *protocolAdapterFactory) validateHttpConfig(config interfaces.Config) error {
	connConfig := config.GetConnection()
	if connConfig == nil {
		return fmt.Errorf("connection config required for HTTP")
	}
	
	addresses := connConfig.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one HTTP endpoint required")
	}
	
	// 验证URL格式
	for _, addr := range addresses {
		if addr == "" {
			return fmt.Errorf("empty HTTP endpoint not allowed")
		}
		// 这里可以添加URL格式验证
	}
	
	return nil
}

// validateKafkaConfig 验证Kafka配置
func (f *protocolAdapterFactory) validateKafkaConfig(config interfaces.Config) error {
	connConfig := config.GetConnection()
	if connConfig == nil {
		return fmt.Errorf("connection config required for Kafka")
	}
	
	addresses := connConfig.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one Kafka broker required")
	}
	
	// 验证broker地址格式
	for _, addr := range addresses {
		if addr == "" {
			return fmt.Errorf("empty Kafka broker address not allowed")
		}
		// 这里可以添加更详细的broker地址验证
	}
	
	return nil
}