package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
	"abc-runner/servers/pkg/grpc"
	"abc-runner/servers/pkg/http"
	"abc-runner/servers/pkg/interfaces"
	"abc-runner/servers/pkg/tcp"
	"abc-runner/servers/pkg/udp"
	"abc-runner/servers/pkg/websocket"
)

type ServerInfo struct {
	Name   string
	Server interfaces.Server
	Config interfaces.ServerConfig
}

func main() {
	var (
		httpPort      = flag.Int("http-port", 8080, "HTTP server port")
		tcpPort       = flag.Int("tcp-port", 9090, "TCP server port")
		udpPort       = flag.Int("udp-port", 9091, "UDP server port")
		grpcPort      = flag.Int("grpc-port", 50051, "gRPC server port")
		websocketPort = flag.Int("websocket-port", 7070, "WebSocket server port")
		host          = flag.String("host", "localhost", "Server host for all protocols")
		logLevel      = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		protocols     = flag.String("protocols", "all", "Protocols to start (all,http,tcp,udp,grpc,websocket)")
		help          = flag.Bool("help", false, "Show help information")
		version       = flag.Bool("version", false, "Show version information")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	// 初始化日志
	logger := logging.NewLogger(*logLevel)
	logger.Info("Starting multi-protocol server suite", map[string]interface{}{
		"protocols": *protocols,
		"host":      *host,
		"log_level": *logLevel,
	})

	// 创建指标收集器
	metricsCollector := monitoring.NewMetricsCollector()

	// 创建服务端
	servers := createServers(*protocols, *host, *httpPort, *tcpPort, *udpPort, *grpcPort, *websocketPort, logger, metricsCollector)

	if len(servers) == 0 {
		logger.Fatal("No servers to start", nil)
		os.Exit(1)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动所有服务端
	if err := startAllServers(ctx, servers, logger); err != nil {
		logger.Fatal("Failed to start servers", err)
		os.Exit(1)
	}

	// 显示启动信息
	showStartupInfo(servers, logger)

	// 等待中断信号
	waitForShutdown(ctx, cancel, servers, logger)
}

// createServers 创建服务端实例
func createServers(protocols, host string, httpPort, tcpPort, udpPort, grpcPort, websocketPort int, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) []ServerInfo {
	var servers []ServerInfo

	// HTTP服务端
	if protocols == "all" || protocols == "http" || contains(protocols, "http") {
		httpConfig := http.NewHTTPServerConfig()
		httpConfig.BaseConfig.Host = host
		httpConfig.BaseConfig.Port = httpPort

		httpServer := http.NewHTTPServer(httpConfig, logger, metricsCollector)
		servers = append(servers, ServerInfo{
			Name:   "HTTP",
			Server: httpServer,
			Config: httpConfig,
		})
	}

	// TCP服务端
	if protocols == "all" || protocols == "tcp" || contains(protocols, "tcp") {
		tcpConfig := tcp.NewTCPServerConfig()
		tcpConfig.BaseConfig.Host = host
		tcpConfig.BaseConfig.Port = tcpPort

		tcpServer := tcp.NewTCPServer(tcpConfig, logger, metricsCollector)
		servers = append(servers, ServerInfo{
			Name:   "TCP",
			Server: tcpServer,
			Config: tcpConfig,
		})
	}

	// UDP服务端
	if protocols == "all" || protocols == "udp" || contains(protocols, "udp") {
		udpConfig := udp.NewUDPServerConfig()
		udpConfig.BaseConfig.Host = host
		udpConfig.BaseConfig.Port = udpPort

		udpServer := udp.NewUDPServer(udpConfig, logger, metricsCollector)
		servers = append(servers, ServerInfo{
			Name:   "UDP",
			Server: udpServer,
			Config: udpConfig,
		})
	}

	// gRPC服务端
	if protocols == "all" || protocols == "grpc" || contains(protocols, "grpc") {
		grpcConfig := grpc.NewGRPCServerConfig()
		grpcConfig.BaseConfig.Host = host
		grpcConfig.BaseConfig.Port = grpcPort

		grpcServer := grpc.NewGRPCServer(grpcConfig, logger, metricsCollector)
		servers = append(servers, ServerInfo{
			Name:   "gRPC",
			Server: grpcServer,
			Config: grpcConfig,
		})
	}

	// WebSocket服务端
	if protocols == "all" || protocols == "websocket" || contains(protocols, "websocket") {
		websocketConfig := websocket.NewWebSocketServerConfig()
		websocketConfig.BaseConfig.Host = host
		websocketConfig.BaseConfig.Port = websocketPort

		websocketServer := websocket.NewWebSocketServer(websocketConfig, logger, metricsCollector)
		servers = append(servers, ServerInfo{
			Name:   "WebSocket",
			Server: websocketServer,
			Config: websocketConfig,
		})
	}

	return servers
}

// startAllServers 启动所有服务端
func startAllServers(ctx context.Context, servers []ServerInfo, logger interfaces.Logger) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(servers))

	for _, serverInfo := range servers {
		wg.Add(1)
		go func(si ServerInfo) {
			defer wg.Done()

			if err := si.Server.Start(ctx); err != nil {
				errChan <- fmt.Errorf("failed to start %s server: %w", si.Name, err)
				return
			}

			logger.Info(fmt.Sprintf("%s server started successfully", si.Name), map[string]interface{}{
				"protocol": si.Config.GetProtocol(),
				"address":  si.Config.GetAddress(),
			})
		}(serverInfo)
	}

	// 等待所有服务端启动
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 检查启动错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// showStartupInfo 显示启动信息
func showStartupInfo(servers []ServerInfo, logger interfaces.Logger) {
	logger.Info("All servers started successfully", map[string]interface{}{
		"server_count": len(servers),
		"pid":          os.Getpid(),
	})

	fmt.Println("\n🚀 abc-runner Multi-Protocol Server Suite")
	fmt.Println("=" + strings.Repeat("=", 50))

	for _, serverInfo := range servers {
		fmt.Printf("✅ %s Server: %s\n", serverInfo.Name, serverInfo.Config.GetAddress())
	}

	fmt.Println("\n📊 Available Endpoints:")
	for _, serverInfo := range servers {
		switch serverInfo.Config.GetProtocol() {
		case "http":
			fmt.Printf("   %s: http://%s/health (health check)\n", serverInfo.Name, serverInfo.Config.GetAddress())
			fmt.Printf("   %s: http://%s/metrics (metrics)\n", serverInfo.Name, serverInfo.Config.GetAddress())
		case "grpc":
			fmt.Printf("   %s: http://%s/ (service info)\n", serverInfo.Name, serverInfo.Config.GetAddress())
			fmt.Printf("   %s: http://%s/TestService/Echo (echo)\n", serverInfo.Name, serverInfo.Config.GetAddress())
		case "websocket":
			fmt.Printf("   %s: http://%s/health (health check)\n", serverInfo.Name, serverInfo.Config.GetAddress())
			fmt.Printf("   %s: http://%s/metrics (metrics)\n", serverInfo.Name, serverInfo.Config.GetAddress())
			fmt.Printf("   %s: ws://%s/ws (websocket endpoint)\n", serverInfo.Name, serverInfo.Config.GetAddress())
		default:
			fmt.Printf("   %s: %s://%s (echo server)\n", serverInfo.Name, serverInfo.Config.GetProtocol(), serverInfo.Config.GetAddress())
		}
	}

	fmt.Println("\n⚡ Press Ctrl+C to stop all servers")
	fmt.Println()
}

// waitForShutdown 等待关闭信号
func waitForShutdown(ctx context.Context, cancel context.CancelFunc, servers []ServerInfo, logger interfaces.Logger) {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	// 开始优雅关闭
	logger.Info("Initiating graceful shutdown of all servers...")
	fmt.Println("\n🛑 Shutting down all servers...")

	// 创建关闭超时上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 并行关闭所有服务端
	var wg sync.WaitGroup
	for _, serverInfo := range servers {
		wg.Add(1)
		go func(si ServerInfo) {
			defer wg.Done()

			if err := si.Server.Stop(shutdownCtx); err != nil {
				logger.Error(fmt.Sprintf("Error stopping %s server", si.Name), err)
				fmt.Printf("❌ Error stopping %s server: %v\n", si.Name, err)
			} else {
				fmt.Printf("✅ %s server stopped\n", si.Name)
			}
		}(serverInfo)
	}

	// 等待所有服务端关闭
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All servers shutdown completed successfully")
		fmt.Println("✅ All servers stopped successfully")
	case <-time.After(25 * time.Second):
		logger.Error("Timeout waiting for servers to stop", nil)
		fmt.Println("⚠️  Timeout waiting for servers to stop")
	}

	cancel()
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`Multi-Protocol Server Suite for abc-runner

USAGE:
    multi-server [OPTIONS]

OPTIONS:
    -host <host>           Server host for all protocols (default: localhost)
    -http-port <port>      HTTP server port (default: 8080)
    -tcp-port <port>       TCP server port (default: 9090)
    -udp-port <port>       UDP server port (default: 9091)
    -grpc-port <port>      gRPC server port (default: 50051)
    -websocket-port <port> WebSocket server port (default: 7070)
    -protocols <list>      Protocols to start: all,http,tcp,udp,grpc,websocket (default: all)
    -log-level <level>     Log level: debug, info, warn, error (default: info)
    -help                  Show this help message
    -version               Show version information

EXAMPLES:
    # Start all servers with default ports
    multi-server

    # Start only HTTP and WebSocket servers
    multi-server -protocols http,websocket

    # Start all servers on different host
    multi-server -host 0.0.0.0

    # Start with custom ports
    multi-server -http-port 8888 -websocket-port 8899

    # Start with debug logging
    multi-server -log-level debug

SUPPORTED PROTOCOLS:
    - HTTP:      RESTful API server with health checks and metrics
    - TCP:       Connection-oriented echo server with keep-alive
    - UDP:       Connectionless packet server with loss simulation
    - gRPC:      RPC server with streaming support
    - WebSocket: Real-time bidirectional communication server

FEATURES:
    - Unified management of multiple protocol servers
    - Graceful startup and shutdown
    - Centralized logging and metrics
    - Health monitoring for all protocols
    - Easy testing environment setup

SIGNALS:
    SIGINT, SIGTERM  - Graceful shutdown of all servers
`)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Println("Multi-Protocol Server Suite")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Built for: abc-runner performance testing framework")
	fmt.Println("Protocols: HTTP, TCP, UDP, gRPC, WebSocket")

	// 显示构建信息（如果可用）
	if buildDate := os.Getenv("BUILD_DATE"); buildDate != "" {
		fmt.Printf("Build Date: %s\n", buildDate)
	}

	if gitCommit := os.Getenv("GIT_COMMIT"); gitCommit != "" {
		fmt.Printf("Git Commit: %s\n", gitCommit)
	}
}

// 工具函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
