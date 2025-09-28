package discovery

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"abc-runner/app/adapters/http"
	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/redis"
	"abc-runner/app/adapters/grpc"
	"abc-runner/app/adapters/tcp"
	"abc-runner/app/adapters/udp"
	"abc-runner/app/adapters/websocket"
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
	
	// 直接使用新的泛型收集器
	builder.components["metrics_collector"] = baseCollector
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
		{"tcp", &TCPAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
		{"udp", &UDPAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
		{"grpc", &GRPCAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
		{"websocket", &WebSocketAdapterFactory{metricsCollector: builder.getMetricsCollector()}},
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
		
		// 使用具体的命令处理器创建函数
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
		case "tcp":
			handler := commands.NewTCPCommandHandler(factory)
			builder.components[handlerName] = handler
		case "udp":
			handler := commands.NewUDPCommandHandler(factory)
			builder.components[handlerName] = handler
		case "grpc":
			handler := commands.NewGRPCCommandHandler(factory)
			builder.components[handlerName] = handler
		case "websocket":
			handler := commands.NewWebSocketCommandHandler(factory)
			builder.components[handlerName] = handler
		default:
			log.Printf("⚠️  Unknown protocol: %s", protocolName)
			continue
		}
		
		log.Printf("✅ Registered command handler: %s", handlerName)
	}
	
	return nil
}

// getMetricsCollector 获取指标收集器
func (builder *AutoDIBuilder) getMetricsCollector() interfaces.DefaultMetricsCollector {
	if collector, exists := builder.components["metrics_collector"]; exists {
		return collector.(interfaces.DefaultMetricsCollector)
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

// AddComponent 添加组件
func (builder *AutoDIBuilder) AddComponent(name string, component interface{}) {
	builder.components[name] = component
}

// RedisAdapterFactory Redis适配器工厂
type RedisAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *RedisAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 直接创建Redis适配器
	adapter := redis.NewRedisAdapter(f.metricsCollector)
	
	// 使用适配器模式包装返回的适配器，使其实现新接口
	return &RedisAdapterWrapper{
		baseAdapter:      adapter,
		metricsCollector: f.metricsCollector,
	}
}

func (f *RedisAdapterFactory) GetProtocolName() string {
	return "redis"
}

// HttpAdapterFactory HTTP适配器工厂
type HttpAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *HttpAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 创建兼容性包装器 - 暂时返回nil，需要先修复HTTP适配器
	return nil
}

func (f *HttpAdapterFactory) GetProtocolName() string {
	return "http"
}

// KafkaAdapterFactory Kafka适配器工厂
type KafkaAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *KafkaAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 创建兼容性包装器 - 暂时返回nil，需要先修复Kafka适配器
	return nil
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

// 协议适配器包装器，用于统一新旧接口

// RedisAdapterWrapper Redis适配器包装器
type RedisAdapterWrapper struct {
	baseAdapter      *redis.RedisAdapter
	metricsCollector interfaces.MetricsCollector[map[string]interface{}]
}

func (w *RedisAdapterWrapper) Connect(ctx context.Context, config interfaces.Config) error {
	return w.baseAdapter.Connect(ctx, config)
}

func (w *RedisAdapterWrapper) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return w.baseAdapter.Execute(ctx, operation)
}

func (w *RedisAdapterWrapper) Close() error {
	return w.baseAdapter.Close()
}

func (w *RedisAdapterWrapper) GetProtocolMetrics() map[string]interface{} {
	return w.baseAdapter.GetProtocolMetrics()
}

func (w *RedisAdapterWrapper) HealthCheck(ctx context.Context) error {
	return w.baseAdapter.HealthCheck(ctx)
}

func (w *RedisAdapterWrapper) GetProtocolName() string {
	return w.baseAdapter.GetProtocolName()
}

func (w *RedisAdapterWrapper) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return w.metricsCollector
}

// HttpAdapterWrapper HTTP适配器包装器
type HttpAdapterWrapper struct {
	baseAdapter      *http.HttpAdapter
	metricsCollector interfaces.MetricsCollector[map[string]interface{}]
}

func (w *HttpAdapterWrapper) Connect(ctx context.Context, config interfaces.Config) error {
	return w.baseAdapter.Connect(ctx, config)
}

func (w *HttpAdapterWrapper) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return w.baseAdapter.Execute(ctx, operation)
}

func (w *HttpAdapterWrapper) Close() error {
	return w.baseAdapter.Close()
}

func (w *HttpAdapterWrapper) GetProtocolMetrics() map[string]interface{} {
	return w.baseAdapter.GetProtocolMetrics()
}

func (w *HttpAdapterWrapper) HealthCheck(ctx context.Context) error {
	return w.baseAdapter.HealthCheck(ctx)
}

func (w *HttpAdapterWrapper) GetProtocolName() string {
	return w.baseAdapter.GetProtocolName()
}

func (w *HttpAdapterWrapper) GetMetricsCollector() interfaces.MetricsCollector[map[string]interface{}] {
	return w.metricsCollector
}

// KafkaAdapterWrapper Kafka适配器包装器
type KafkaAdapterWrapper struct {
	baseAdapter      *kafka.KafkaAdapter
	metricsCollector interfaces.MetricsCollector[map[string]interface{}]
}

func (w *KafkaAdapterWrapper) Connect(ctx context.Context, config interfaces.Config) error {
	return w.baseAdapter.Connect(ctx, config)
}

func (w *KafkaAdapterWrapper) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return w.baseAdapter.Execute(ctx, operation)
}

func (w *KafkaAdapterWrapper) Close() error {
	return w.baseAdapter.Close()
}

func (w *KafkaAdapterWrapper) GetProtocolMetrics() map[string]interface{} {
	return w.baseAdapter.GetProtocolMetrics()
}

func (w *KafkaAdapterWrapper) HealthCheck(ctx context.Context) error {
	return w.baseAdapter.HealthCheck(ctx)
}

func (w *KafkaAdapterWrapper) GetProtocolName() string {
	return w.baseAdapter.GetProtocolName()
}

func (w *KafkaAdapterWrapper) GetMetricsCollector() interfaces.MetricsCollector[map[string]interface{}] {
	return w.metricsCollector
}

// TCPAdapterFactory TCP适配器工厂
type TCPAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *TCPAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 创建TCP适配器实例
	tcpAdapter := tcp.NewTCPAdapter(f.metricsCollector)
	return tcpAdapter
}

func (f *TCPAdapterFactory) GetProtocolName() string {
	return "tcp"
}

// UDPAdapterFactory UDP适配器工厂
type UDPAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *UDPAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 创建UDP适配器实例
	udpAdapter := udp.NewUDPAdapter(f.metricsCollector)
	return udpAdapter
}

func (f *UDPAdapterFactory) GetProtocolName() string {
	return "udp"
}

// GRPCAdapterFactory gRPC适配器工厂
type GRPCAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *GRPCAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 创建gRPC适配器实例
	grpcAdapter := grpc.NewGRPCAdapter(f.metricsCollector)
	return grpcAdapter
}

func (f *GRPCAdapterFactory) GetProtocolName() string {
	return "grpc"
}

// WebSocketAdapterFactory WebSocket适配器工厂
type WebSocketAdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

func (f *WebSocketAdapterFactory) CreateAdapter() interfaces.ProtocolAdapter {
	// 创建WebSocket适配器实例
	websocketAdapter := websocket.NewWebSocketAdapter(f.metricsCollector)
	return websocketAdapter
}

func (f *WebSocketAdapterFactory) GetProtocolName() string {
	return "websocket"
}
