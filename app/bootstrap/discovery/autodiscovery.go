package discovery

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"abc-runner/app/adapters/http"
	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/redis"
	"abc-runner/app/commands"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// AutoDIBuilder 自动装配构建器
type AutoDIBuilder struct {
	components map[string]interface{}
	factories  map[string]AdapterFactory
}

// AdapterFactory 适配器工厂接口
type AdapterFactory interface {
	CreateAdapter() interfaces.ProtocolAdapter
	GetProtocolName() string
}

// Component 组件信息
type Component struct {
	Name     string
	Type     reflect.Type
	Instance interface{}
	Path     string
}

// NewAutoDIBuilder 创建自动装配构建器
func NewAutoDIBuilder() *AutoDIBuilder {
	return &AutoDIBuilder{
		components: make(map[string]interface{}),
		factories:  make(map[string]AdapterFactory),
	}
}

// Build 构建并装配所有组件
func (builder *AutoDIBuilder) Build() error {
	log.Println("Starting automatic dependency injection...")
	
	// 1. 创建指标收集器
	if err := builder.setupMetricsCollector(); err != nil {
		return fmt.Errorf("failed to setup metrics collector: %w", err)
	}
	
	// 2. 发现并注册协议适配器
	if err := builder.discoverProtocolAdapters(); err != nil {
		return fmt.Errorf("failed to discover protocol adapters: %w", err)
	}
	
	// 3. 注册命令处理器
	if err := builder.registerCommandHandlers(); err != nil {
		return fmt.Errorf("failed to register command handlers: %w", err)
	}
	
	log.Printf("Auto DI completed. Registered %d components", len(builder.components))
	return nil
}

// setupMetricsCollector 设置指标收集器
func (builder *AutoDIBuilder) setupMetricsCollector() error {
	// 创建通用指标收集器
	config := metrics.DefaultMetricsConfig()
	protocolData := map[string]interface{}{
		"application": "abc-runner",
		"version":     "3.0.0",
	}
	
	baseCollector := metrics.NewBaseCollector(config, protocolData)
	
	// 创建适配器，用于新旧接口桥接
	metricsAdapter := &MetricsAdapter{
		baseCollector: baseCollector,
	}
	
	builder.components["metrics_collector"] = metricsAdapter
	log.Println("✅ Metrics collector registered")
	
	return nil
}

// discoverProtocolAdapters 发现协议适配器
func (builder *AutoDIBuilder) discoverProtocolAdapters() error {
	log.Println("Discovering protocol adapters...")
	
	// 获取项目根目录
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Printf("Warning: failed to find project root: %v", err)
		projectRoot = "."
	}
	log.Printf("Using project root: %s", projectRoot)
	
	// 基于约定注册已知的协议适配器
	protocols := []struct {
		name        string
		factory     AdapterFactory
	}{
		{"redis", &RedisAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
		{"http", &HttpAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
		{"kafka", &KafkaAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
	}
	
	for _, protocol := range protocols {
		// 检查协议目录是否存在
		protocolPath := filepath.Join(projectRoot, "app", "adapters", protocol.name)
		if _, err := os.Stat(protocolPath); err == nil {
			builder.factories[protocol.name] = protocol.factory
			builder.components[protocol.name+"_factory"] = protocol.factory
			log.Printf("✅ Discovered protocol: %s (path: %s)", protocol.name, protocolPath)
		} else {
			log.Printf("⚠️  Protocol directory not found: %s", protocolPath)
			// 即使目录不存在，也注册协议（对于简化实现）
			builder.factories[protocol.name] = protocol.factory
			builder.components[protocol.name+"_factory"] = protocol.factory
			log.Printf("✅ Registered protocol without directory check: %s", protocol.name)
		}
	}
	
	return nil
}

// registerCommandHandlers 注册命令处理器
func (builder *AutoDIBuilder) registerCommandHandlers() error {
	log.Println("Registering command handlers...")
	
	// 为每个发现的协议创建命令处理器
	for protocolName, factory := range builder.factories {
		handlerName := protocolName + "_handler"
		
		// 使用新的简化命令处理器
		switch protocolName {
		case "redis":
			handler := commands.NewRedisCommandHandler(factory)
			builder.components[handlerName] = handler
		case "http":
			handler := commands.NewHttpCommandHandler(factory)
			builder.components[handlerName] = handler
		case "kafka":
			handler := commands.NewKafkaCommandHandler(factory)
			builder.components[handlerName] = handler
		}
		
		log.Printf("✅ Registered command handler: %s", handlerName)
	}
	
	return nil
}

// getMetricsCollector 获取指标收集器
func (builder *AutoDIBuilder) getMetricsCollector() interfaces.MetricsCollector {
	if collector, exists := builder.components["metrics_collector"]; exists {
		return collector.(interfaces.MetricsCollector)
	}
	return nil
}

// GetComponent 获取组件
func (builder *AutoDIBuilder) GetComponent(name string) (interface{}, bool) {
	component, exists := builder.components[name]
	return component, exists
}

// GetFactory 获取工厂
func (builder *AutoDIBuilder) GetFactory(protocol string) (AdapterFactory, bool) {
	factory, exists := builder.factories[protocol]
	return factory, exists
}

// GetAllFactories 获取所有工厂
func (builder *AutoDIBuilder) GetAllFactories() map[string]AdapterFactory {
	return builder.factories
}

// MetricsAdapter 指标适配器（新旧接口桥接）
type MetricsAdapter struct {
	baseCollector *metrics.BaseCollector[map[string]interface{}]
}

func (m *MetricsAdapter) RecordOperation(result *interfaces.OperationResult) {
	m.baseCollector.Record(result)
}

func (m *MetricsAdapter) GetMetrics() *interfaces.Metrics {
	snapshot := m.baseCollector.Snapshot()
	
	return &interfaces.Metrics{
		TotalOps:   snapshot.Core.Operations.Total,
		SuccessOps: snapshot.Core.Operations.Success,
		FailedOps:  snapshot.Core.Operations.Failed,
		ReadOps:    snapshot.Core.Operations.Read,
		WriteOps:   snapshot.Core.Operations.Write,
		AvgLatency: snapshot.Core.Latency.Average,
		MinLatency: snapshot.Core.Latency.Min,
		MaxLatency: snapshot.Core.Latency.Max,
		P90Latency: snapshot.Core.Latency.P90,
		P95Latency: snapshot.Core.Latency.P95,
		P99Latency: snapshot.Core.Latency.P99,
		ErrorRate:  float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100,
		RPS:        int32(snapshot.Core.Throughput.RPS),
	}
}

func (m *MetricsAdapter) Reset() {
	m.baseCollector.Reset()
}

func (m *MetricsAdapter) Export() map[string]interface{} {
	snapshot := m.baseCollector.Snapshot()
	
	return map[string]interface{}{
		"total_ops":    snapshot.Core.Operations.Total,
		"success_ops":  snapshot.Core.Operations.Success,
		"failed_ops":   snapshot.Core.Operations.Failed,
		"success_rate": snapshot.Core.Operations.Rate,
		"rps":          snapshot.Core.Throughput.RPS,
		"avg_latency":  int64(snapshot.Core.Latency.Average),
		"p95_latency":  int64(snapshot.Core.Latency.P95),
		"p99_latency":  int64(snapshot.Core.Latency.P99),
		"protocol_data": snapshot.Protocol,
	}
}

// GetSnapshot 获取新指标快照
func (m *MetricsAdapter) GetSnapshot() *metrics.MetricsSnapshot[map[string]interface{}] {
	return m.baseCollector.Snapshot()
}

// RedisAdapterFactory Redis适配器工厂
type RedisAdapterFactory struct {
	metricsCollector interfaces.MetricsCollector
}

func (f *RedisAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	return redis.NewRedisAdapter(f.metricsCollector)
}

func (f *RedisAdapterFactory) GetProtocolName() string {
	return "redis"
}

// HttpAdapterFactory HTTP适配器工厂
type HttpAdapterFactory struct {
	metricsCollector interfaces.MetricsCollector
}

func (f *HttpAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	return http.NewHttpAdapter(f.metricsCollector)
}

func (f *HttpAdapterFactory) GetProtocolName() string {
	return "http"
}

// KafkaAdapterFactory Kafka适配器工厂
type KafkaAdapterFactory struct {
	metricsCollector interfaces.MetricsCollector
}

func (f *KafkaAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	return kafka.NewKafkaAdapter(f.metricsCollector)
}

func (f *KafkaAdapterFactory) GetProtocolName() string {
	return "kafka"
}

// ScanAdapterDirectories 扫描适配器目录
func ScanAdapterDirectories(basePath string) ([]string, error) {
	var protocols []string
	
	adaptersPath := filepath.Join(basePath, "app", "adapters")
	
	entries, err := os.ReadDir(adaptersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read adapters directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			protocolName := entry.Name()
			
			// 检查是否包含必需的文件
			adapterFile := filepath.Join(adaptersPath, protocolName, "adapter.go")
			if _, err := os.Stat(adapterFile); err == nil {
				protocols = append(protocols, protocolName)
				log.Printf("Found protocol directory: %s", protocolName)
			}
		}
	}
	
	return protocols, nil
}

// IsValidProtocolName 检查是否是有效的协议名称
func IsValidProtocolName(name string) bool {
	validProtocols := []string{"redis", "http", "https", "kafka", "grpc", "tcp", "udp"}
	
	name = strings.ToLower(name)
	for _, valid := range validProtocols {
		if name == valid {
			return true
		}
	}
	
	return false
}

// findProjectRoot 查找项目根目录
func findProjectRoot() (string, error) {
	// 获取当前执行文件的路径
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// 获取执行文件所在目录
	execDir := filepath.Dir(execPath)
	
	// 从执行文件目录往上查找项目根目录
	currentDir := execDir
	for {
		// 检查是否存在项目标识文件
		if checkProjectMarkers(currentDir) {
			return currentDir, nil
		}
		
		// 向上一级目录
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// 已到达根目录
			break
		}
		currentDir = parentDir
	}
	
	// 如果在bin目录中，尝试上一级目录
	if filepath.Base(execDir) == "bin" {
		parentDir := filepath.Dir(execDir)
		if checkProjectMarkers(parentDir) {
			return parentDir, nil
		}
	}
	
	// 如果找不到，尝试使用当前工作目录
	workDir, err := os.Getwd()
	if err == nil {
		// 从当前工作目录往上查找
		currentDir = workDir
		for {
			if checkProjectMarkers(currentDir) {
				return currentDir, nil
			}
			
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir {
				break
			}
			currentDir = parentDir
		}
	}
	
	return "", fmt.Errorf("project root not found")
}

// checkProjectMarkers 检查项目标识文件
func checkProjectMarkers(dir string) bool {
	markers := []string{"go.mod", "main.go", "app", "Makefile"}
	
	for _, marker := range markers {
		markerPath := filepath.Join(dir, marker)
		if _, err := os.Stat(markerPath); err == nil {
			return true
		}
	}
	
	return false
}