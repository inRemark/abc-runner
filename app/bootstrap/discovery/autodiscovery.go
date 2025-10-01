package discovery

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"abc-runner/app/adapters/grpc"
	"abc-runner/app/adapters/http"
	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/redis"
	"abc-runner/app/adapters/tcp"
	"abc-runner/app/adapters/udp"
	"abc-runner/app/adapters/websocket"

	"abc-runner/app/commands"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"abc-runner/config"
)

// AutoDIBuilder 自动装配构建器
type AutoDIBuilder struct {
	components map[string]interface{}
	// 使用接口分离模式存储各协议工厂
	grpcFactory      interfaces.GRPCAdapterFactory
	tcpFactory       interfaces.TCPAdapterFactory
	udpFactory       interfaces.UDPAdapterFactory
	websocketFactory interfaces.WebSocketAdapterFactory
	redisFactory     interfaces.RedisAdapterFactory
	httpFactory      interfaces.HttpAdapterFactory
	kafkaFactory     interfaces.KafkaAdapterFactory
	// 保留通用查找接口，向下兼容
	factories map[string]interface{}
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
		factories:  make(map[string]interface{}),
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
	metricsConfig := metrics.DefaultMetricsConfig()
	protocolData := map[string]interface{}{
		"application": config.AppName,
		"version":     config.AutoDiscoveryVersion,
	}

	baseCollector := metrics.NewBaseCollector(metricsConfig, protocolData)

	// 直接使用新的泛型收集器
	builder.components["metrics_collector"] = baseCollector
	log.Println("✅ Metrics collector registered")

	return nil
}

// discoverProtocolAdapters 发现协议适配器 - 实现所有协议的接口分离模式
func (builder *AutoDIBuilder) discoverProtocolAdapters() error {
	log.Println("Discovering protocol adapters...")

	// 获取项目根目录
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Printf("Warning: failed to find project root: %v", err)
		projectRoot = "."
	}
	log.Printf("Using project root: %s", projectRoot)

	// 获取指标收集器
	metricsCollector := builder.getMetricsCollector()

	// 使用接口分离模式注册所有协议工厂

	// 创建并注册gRPC工厂
	builder.grpcFactory = grpc.NewAdapterFactory(metricsCollector)
	builder.factories["grpc"] = builder.grpcFactory
	builder.components["grpc_factory"] = builder.grpcFactory
	log.Printf("✅ Registered gRPC adapter factory")

	// 创建并注册TCP工厂
	builder.tcpFactory = tcp.NewAdapterFactory(metricsCollector)
	builder.factories["tcp"] = builder.tcpFactory
	builder.components["tcp_factory"] = builder.tcpFactory
	log.Printf("✅ Registered TCP adapter factory")

	// 创建并注册UDP工厂
	builder.udpFactory = udp.NewAdapterFactory(metricsCollector)
	builder.factories["udp"] = builder.udpFactory
	builder.components["udp_factory"] = builder.udpFactory
	log.Printf("✅ Registered UDP adapter factory")

	// 创建并注册WebSocket工厂
	builder.websocketFactory = websocket.NewAdapterFactory(metricsCollector)
	builder.factories["websocket"] = builder.websocketFactory
	builder.components["websocket_factory"] = builder.websocketFactory
	log.Printf("✅ Registered WebSocket adapter factory")

	// 创建并注册Redis工厂
	builder.redisFactory = redis.NewAdapterFactory(metricsCollector)
	builder.factories["redis"] = builder.redisFactory
	builder.components["redis_factory"] = builder.redisFactory
	log.Printf("✅ Registered Redis adapter factory")

	// 创建并注册HTTP工厂
	builder.httpFactory = http.NewAdapterFactory(metricsCollector)
	builder.factories["http"] = builder.httpFactory
	builder.components["http_factory"] = builder.httpFactory
	log.Printf("✅ Registered HTTP adapter factory")

	// 创建并注册Kafka工厂
	builder.kafkaFactory = kafka.NewAdapterFactory(metricsCollector)
	builder.factories["kafka"] = builder.kafkaFactory
	builder.components["kafka_factory"] = builder.kafkaFactory
	log.Printf("✅ Registered Kafka adapter factory")

	log.Printf("🎉 All implemented protocol factories registered successfully!")
	return nil
}

// registerCommandHandlers 注册命令处理器 - 支持所有协议
func (builder *AutoDIBuilder) registerCommandHandlers() error {
	log.Println("Registering command handlers...")

	// 为所有已实现的协议创建命令处理器

	// gRPC 命令处理器
	if builder.grpcFactory != nil {
		handler := commands.NewGRPCCommandHandler(builder.grpcFactory)
		builder.components["grpc_handler"] = handler
		log.Printf("✅ Registered command handler: grpc_handler")
	}

	// TCP 命令处理器
	if builder.tcpFactory != nil {
		handler := commands.NewTCPCommandHandler(builder.tcpFactory)
		builder.components["tcp_handler"] = handler
		log.Printf("✅ Registered command handler: tcp_handler")
	}

	// UDP 命令处理器
	if builder.udpFactory != nil {
		handler := commands.NewUDPCommandHandler(builder.udpFactory)
		builder.components["udp_handler"] = handler
		log.Printf("✅ Registered command handler: udp_handler")
	}

	// WebSocket 命令处理器
	if builder.websocketFactory != nil {
		handler := commands.NewWebSocketCommandHandler(builder.websocketFactory)
		builder.components["websocket_handler"] = handler
		log.Printf("✅ Registered command handler: websocket_handler")
	}

	// Redis 命令处理器
	if builder.redisFactory != nil {
		handler := commands.NewRedisCommandHandler(builder.redisFactory)
		builder.components["redis_handler"] = handler
		log.Printf("✅ Registered command handler: redis_handler")
	}

	// HTTP 命令处理器
	if builder.httpFactory != nil {
		handler := commands.NewHttpCommandHandler(builder.httpFactory)
		builder.components["http_handler"] = handler
		log.Printf("✅ Registered command handler: http_handler")
	}

	// Kafka 命令处理器
	if builder.kafkaFactory != nil {
		handler := commands.NewKafkaCommandHandler(builder.kafkaFactory)
		builder.components["kafka_handler"] = handler
		log.Printf("✅ Registered command handler: kafka_handler")
	}

	log.Printf("🎉 All implemented command handlers registered successfully!")
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

// GetFactory 获取工厂 - 支持接口分离模式
func (builder *AutoDIBuilder) GetFactory(protocol string) (interface{}, bool) {
	factory, exists := builder.factories[protocol]
	return factory, exists
}

// GetGRPCFactory 获取gRPC工厂
func (builder *AutoDIBuilder) GetGRPCFactory() (interfaces.GRPCAdapterFactory, bool) {
	if builder.grpcFactory != nil {
		return builder.grpcFactory, true
	}
	return nil, false
}

// GetAllFactories 获取所有工厂 - 返回接口列表
func (builder *AutoDIBuilder) GetAllFactories() map[string]interface{} {
	return builder.factories
}

// AddComponent 添加组件
func (builder *AutoDIBuilder) AddComponent(name string, component interface{}) {
	builder.components[name] = component
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
	validProtocols := []string{"redis", "http", "https", "kafka", "grpc", "tcp", "udp", "websocket"}

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
	markers := []string{"go.mod", "main.go", "Makefile", ".git"}

	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}

	return false
}
