package unified

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/unified/config"
)

// unifiedConfigManager 统一配置管理器实现
type unifiedConfigManager struct {
	loaders map[string]config.ConfigLoader
	sources map[string]config.ConfigSource
	cache   map[string]interfaces.Config
	mutex   sync.RWMutex
}

// NewUnifiedConfigManager 创建统一配置管理器
func NewUnifiedConfigManager() config.UnifiedConfigManager {
	manager := &unifiedConfigManager{
		loaders: make(map[string]config.ConfigLoader),
		sources: make(map[string]config.ConfigSource),
		cache:   make(map[string]interfaces.Config),
		mutex:   sync.RWMutex{},
	}
	
	// 注册默认的配置加载器
	manager.registerDefaultLoaders()
	
	return manager
}

// LoadConfig 加载配置
func (m *unifiedConfigManager) LoadConfig(source config.ConfigSource, protocol string) (interfaces.Config, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 生成缓存键
	cacheKey := fmt.Sprintf("%s:%s:%s", protocol, source.GetType(), source.GetPath())
	
	// 检查缓存
	if cached, exists := m.cache[cacheKey]; exists {
		log.Printf("Config loaded from cache: %s", cacheKey)
		return cached.Clone(), nil
	}
	
	// 获取加载器
	loader, exists := m.loaders[protocol]
	if !exists {
		return nil, fmt.Errorf("no config loader found for protocol: %s", protocol)
	}
	
	// 验证配置源
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("config source validation failed: %w", err)
	}
	
	// 从源加载配置
	cfg, err := loader.LoadFromSource(source)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from source: %w", err)
	}
	
	// 验证配置
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	// 缓存配置
	m.cache[cacheKey] = cfg.Clone()
	
	log.Printf("Config loaded successfully: %s", cacheKey)
	return cfg, nil
}

// MergeConfigs 合并多个配置
func (m *unifiedConfigManager) MergeConfigs(configs ...interfaces.Config) (interfaces.Config, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configs to merge")
	}
	
	if len(configs) == 1 {
		return configs[0].Clone(), nil
	}
	
	// 基础配置（第一个配置作为基础）
	base := configs[0].Clone()
	baseUnified, ok := base.(*config.UnifiedConfig)
	if !ok {
		return nil, fmt.Errorf("config is not UnifiedConfig type")
	}
	
	// 逐个合并后续配置
	for i := 1; i < len(configs); i++ {
		if err := m.mergeIntoBase(baseUnified, configs[i]); err != nil {
			return nil, fmt.Errorf("failed to merge config %d: %w", i, err)
		}
	}
	
	// 验证合并后的配置
	if err := m.ValidateConfig(baseUnified); err != nil {
		return nil, fmt.Errorf("merged config validation failed: %w", err)
	}
	
	return baseUnified, nil
}

// ValidateConfig 验证配置
func (m *unifiedConfigManager) ValidateConfig(cfg interfaces.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// 调用配置自身的验证方法
	if err := cfg.Validate(); err != nil {
		return err
	}
	
	// 获取协议特定的加载器进行验证
	protocol := cfg.GetProtocol()
	if loader, exists := m.loaders[protocol]; exists {
		if err := loader.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("protocol-specific validation failed: %w", err)
		}
	}
	
	return nil
}

// GetDefaultConfig 获取默认配置
func (m *unifiedConfigManager) GetDefaultConfig(protocol string) (interfaces.Config, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 检查缓存
	cacheKey := fmt.Sprintf("%s:default", protocol)
	if cached, exists := m.cache[cacheKey]; exists {
		return cached.Clone(), nil
	}
	
	// 获取加载器
	loader, exists := m.loaders[protocol]
	if !exists {
		return nil, fmt.Errorf("no config loader found for protocol: %s", protocol)
	}
	
	// 获取默认配置
	defaultCfg := loader.GetDefaultConfig()
	
	// 缓存默认配置
	m.cache[cacheKey] = defaultCfg.Clone()
	
	log.Printf("Default config created for protocol: %s", protocol)
	return defaultCfg, nil
}

// SaveConfig 保存配置
func (m *unifiedConfigManager) SaveConfig(cfg interfaces.Config, destination string) error {
	// 创建目录
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// 根据文件扩展名选择格式
	ext := strings.ToLower(filepath.Ext(destination))
	
	switch ext {
	case ".yaml", ".yml":
		return m.saveConfigAsYAML(cfg, destination)
	case ".json":
		return m.saveConfigAsJSON(cfg, destination)
	default:
		return fmt.Errorf("unsupported config format: %s", ext)
	}
}

// RegisterConfigLoader 注册配置加载器
func (m *unifiedConfigManager) RegisterConfigLoader(protocol string, loader config.ConfigLoader) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.loaders[protocol]; exists {
		return fmt.Errorf("config loader for protocol '%s' already registered", protocol)
	}
	
	m.loaders[protocol] = loader
	log.Printf("Config loader registered for protocol: %s", protocol)
	return nil
}

// GetProtocolConfig 获取特定协议的配置
func (m *unifiedConfigManager) GetProtocolConfig(protocol string, source config.ConfigSource) (interfaces.Config, error) {
	// 先尝试从源加载
	if source != nil {
		return m.LoadConfig(source, protocol)
	}
	
	// 尝试从默认位置加载
	defaultPath := fmt.Sprintf("conf/%s.yaml", protocol)
	if _, err := os.Stat(defaultPath); err == nil {
		fileSource := &FileConfigSource{Path: defaultPath}
		return m.LoadConfig(fileSource, protocol)
	}
	
	// 返回默认配置
	return m.GetDefaultConfig(protocol)
}

// registerDefaultLoaders 注册默认的配置加载器
func (m *unifiedConfigManager) registerDefaultLoaders() {
	// Redis配置加载器
	m.loaders["redis"] = &RedisConfigLoader{}
	
	// HTTP配置加载器
	m.loaders["http"] = &HttpConfigLoader{}
	
	// Kafka配置加载器
	m.loaders["kafka"] = &KafkaConfigLoader{}
	
	log.Println("Default config loaders registered")
}

// mergeIntoBase 将配置合并到基础配置中
func (m *unifiedConfigManager) mergeIntoBase(base *config.UnifiedConfig, source interfaces.Config) error {
	sourceUnified, ok := source.(*config.UnifiedConfig)
	if !ok {
		return fmt.Errorf("source config is not UnifiedConfig type")
	}
	
	// 合并协议（源覆盖基础）
	if sourceUnified.Protocol != "" {
		base.Protocol = sourceUnified.Protocol
	}
	
	// 合并连接配置
	if sourceUnified.Connection != nil {
		if base.Connection == nil {
			base.Connection = sourceUnified.Connection.Clone()
		} else {
			m.mergeConnectionConfig(base.Connection, sourceUnified.Connection)
		}
	}
	
	// 合并基准测试配置
	if sourceUnified.Benchmark != nil {
		if base.Benchmark == nil {
			base.Benchmark = sourceUnified.Benchmark.Clone()
		} else {
			m.mergeBenchmarkConfig(base.Benchmark, sourceUnified.Benchmark)
		}
	}
	
	// 合并全局配置
	if sourceUnified.Global != nil {
		if base.Global == nil {
			base.Global = sourceUnified.Global.Clone()
		} else {
			m.mergeGlobalConfig(base.Global, sourceUnified.Global)
		}
	}
	
	// 合并元数据
	if base.Metadata == nil {
		base.Metadata = make(map[string]interface{})
	}
	for k, v := range sourceUnified.Metadata {
		base.Metadata[k] = v
	}
	
	return nil
}

// mergeConnectionConfig 合并连接配置
func (m *unifiedConfigManager) mergeConnectionConfig(base, source *config.UnifiedConnectionConfig) {
	// 合并地址列表（源覆盖基础）
	if len(source.Addresses) > 0 {
		base.Addresses = append([]string{}, source.Addresses...)
	}
	
	// 合并凭据（源覆盖基础）
	if len(source.Credentials) > 0 {
		if base.Credentials == nil {
			base.Credentials = make(map[string]string)
		}
		for k, v := range source.Credentials {
			base.Credentials[k] = v
		}
	}
	
	// 合并连接池配置
	if source.Pool != nil {
		if base.Pool == nil {
			base.Pool = source.Pool.Clone()
		} else {
			m.mergePoolConfig(base.Pool, source.Pool)
		}
	}
	
	// 合并超时设置
	if source.Timeout > 0 {
		base.Timeout = source.Timeout
	}
	
	// 合并TLS配置
	if source.TLS != nil {
		base.TLS = &config.TLSConfig{
			Enabled:            source.TLS.Enabled,
			CertFile:           source.TLS.CertFile,
			KeyFile:            source.TLS.KeyFile,
			CAFile:             source.TLS.CAFile,
			InsecureSkipVerify: source.TLS.InsecureSkipVerify,
		}
	}
}

// mergeBenchmarkConfig 合并基准测试配置
func (m *unifiedConfigManager) mergeBenchmarkConfig(base, source *config.UnifiedBenchmarkConfig) {
	if source.Total > 0 {
		base.Total = source.Total
	}
	if source.Parallels > 0 {
		base.Parallels = source.Parallels
	}
	if source.DataSize > 0 {
		base.DataSize = source.DataSize
	}
	if source.TTL > 0 {
		base.TTL = source.TTL
	}
	if source.ReadPercent >= 0 {
		base.ReadPercent = source.ReadPercent
	}
	if source.RandomKeys > 0 {
		base.RandomKeys = source.RandomKeys
	}
	if source.TestCase != "" {
		base.TestCase = source.TestCase
	}
	if source.Duration > 0 {
		base.Duration = source.Duration
	}
	if source.RateLimit > 0 {
		base.RateLimit = source.RateLimit
	}
	if source.WarmupTime > 0 {
		base.WarmupTime = source.WarmupTime
	}
}

// mergePoolConfig 合并连接池配置
func (m *unifiedConfigManager) mergePoolConfig(base, source *config.UnifiedPoolConfig) {
	if source.PoolSize > 0 {
		base.PoolSize = source.PoolSize
	}
	if source.MinIdle >= 0 {
		base.MinIdle = source.MinIdle
	}
	if source.MaxIdle >= 0 {
		base.MaxIdle = source.MaxIdle
	}
	if source.IdleTimeout > 0 {
		base.IdleTimeout = source.IdleTimeout
	}
	if source.ConnectionTimeout > 0 {
		base.ConnectionTimeout = source.ConnectionTimeout
	}
	if source.MaxLifetime > 0 {
		base.MaxLifetime = source.MaxLifetime
	}
	if source.RetryInterval > 0 {
		base.RetryInterval = source.RetryInterval
	}
	if source.MaxRetries > 0 {
		base.MaxRetries = source.MaxRetries
	}
}

// mergeGlobalConfig 合并全局配置
func (m *unifiedConfigManager) mergeGlobalConfig(base, source *config.GlobalConfig) {
	if source.LogLevel != "" {
		base.LogLevel = source.LogLevel
	}
	if source.OutputFormat != "" {
		base.OutputFormat = source.OutputFormat
	}
	if source.DefaultProtocol != "" {
		base.DefaultProtocol = source.DefaultProtocol
	}
	
	// 合并别名
	if base.Aliases == nil {
		base.Aliases = make(map[string]string)
	}
	for k, v := range source.Aliases {
		base.Aliases[k] = v
	}
	
	// 合并布尔值配置
	base.MetricsEnabled = source.MetricsEnabled
	base.Debug = source.Debug
}

// saveConfigAsYAML 保存配置为YAML格式
func (m *unifiedConfigManager) saveConfigAsYAML(cfg interfaces.Config, destination string) error {
	// 这里应该使用YAML库来序列化配置
	// 暂时使用简化实现
	log.Printf("Saving config as YAML to: %s", destination)
	return fmt.Errorf("YAML save not implemented yet")
}

// saveConfigAsJSON 保存配置为JSON格式
func (m *unifiedConfigManager) saveConfigAsJSON(cfg interfaces.Config, destination string) error {
	// 这里应该使用JSON库来序列化配置
	// 暂时使用简化实现
	log.Printf("Saving config as JSON to: %s", destination)
	return fmt.Errorf("JSON save not implemented yet")
}