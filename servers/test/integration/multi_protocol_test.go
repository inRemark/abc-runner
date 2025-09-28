package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
	"abc-runner/servers/pkg/grpc"
	httpPkg "abc-runner/servers/pkg/http"
	"abc-runner/servers/pkg/tcp"
	"abc-runner/servers/pkg/udp"
)

// TestMultiProtocolServers 多协议服务端集成测试
func TestMultiProtocolServers(t *testing.T) {
	// 创建共享的日志和指标收集器
	logger := logging.NewLogger("info")
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 创建并启动HTTP服务端
	httpConfig := httpPkg.NewHTTPServerConfig()
	httpConfig.BaseConfig.Port = 28080
	httpServer := httpPkg.NewHTTPServer(httpConfig, logger, metricsCollector)
	
	if err := httpServer.Start(ctx); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	defer httpServer.Stop(context.Background())
	
	// 创建并启动TCP服务端
	tcpConfig := tcp.NewTCPServerConfig()
	tcpConfig.BaseConfig.Port = 29090
	tcpServer := tcp.NewTCPServer(tcpConfig, logger, metricsCollector)
	
	if err := tcpServer.Start(ctx); err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer tcpServer.Stop(context.Background())
	
	// 创建并启动UDP服务端
	udpConfig := udp.NewUDPServerConfig()
	udpConfig.BaseConfig.Port = 29091
	udpServer := udp.NewUDPServer(udpConfig, logger, metricsCollector)
	
	if err := udpServer.Start(ctx); err != nil {
		t.Fatalf("Failed to start UDP server: %v", err)
	}
	defer udpServer.Stop(context.Background())
	
	// 创建并启动gRPC服务端
	grpcConfig := grpc.NewGRPCServerConfig()
	grpcConfig.BaseConfig.Port = 50052
	grpcServer := grpc.NewGRPCServer(grpcConfig, logger, metricsCollector)
	
	if err := grpcServer.Start(ctx); err != nil {
		t.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer grpcServer.Stop(context.Background())
	
	// 等待所有服务端启动
	time.Sleep(200 * time.Millisecond)
	
	// 测试所有服务端是否运行
	t.Run("AllServersRunning", func(t *testing.T) {
		// 测试HTTP服务端
		if !httpServer.IsRunning() {
			t.Error("HTTP server should be running")
		}
		
		// 测试TCP服务端
		if !tcpServer.IsRunning() {
			t.Error("TCP server should be running")
		}
		
		// 测试UDP服务端
		if !udpServer.IsRunning() {
			t.Error("UDP server should be running")
		}
		
		// 测试gRPC服务端
		if !grpcServer.IsRunning() {
			t.Error("gRPC server should be running")
		}
	})
	
	// 测试HTTP端点
	t.Run("HTTPEndpoints", func(t *testing.T) {
		// 健康检查
		resp, err := http.Get("http://localhost:28080/health")
		if err != nil {
			t.Fatalf("Failed to call HTTP health endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("HTTP health check failed, status: %d", resp.StatusCode)
		}
		
		// 指标端点
		resp, err = http.Get("http://localhost:28080/metrics")
		if err != nil {
			t.Fatalf("Failed to call HTTP metrics endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("HTTP metrics endpoint failed, status: %d", resp.StatusCode)
		}
	})
	
	// 测试gRPC端点
	t.Run("GRPCEndpoints", func(t *testing.T) {
		// 服务信息
		resp, err := http.Get("http://localhost:50052/")
		if err != nil {
			t.Fatalf("Failed to call gRPC service info: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("gRPC service info failed, status: %d", resp.StatusCode)
		}
		
		// 健康检查
		resp, err = http.Get("http://localhost:50052/grpc.health.v1.Health/Check")
		if err != nil {
			t.Fatalf("Failed to call gRPC health check: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("gRPC health check failed, status: %d", resp.StatusCode)
		}
	})
	
	// 测试指标收集
	t.Run("MetricsCollection", func(t *testing.T) {
		metrics := metricsCollector.GetMetrics()
		
		if metrics["total_requests"] == nil {
			t.Error("Expected total_requests metric to be present")
		}
		
		if metrics["start_time"] == nil {
			t.Error("Expected start_time metric to be present")
		}
	})
}

// TestServerPortConflicts 测试端口冲突处理
func TestServerPortConflicts(t *testing.T) {
	logger := logging.NewLogger("error")
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建第一个HTTP服务端
	config1 := httpPkg.NewHTTPServerConfig()
	config1.BaseConfig.Port = 38080
	server1 := httpPkg.NewHTTPServer(config1, logger, metricsCollector)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := server1.Start(ctx); err != nil {
		t.Fatalf("Failed to start first HTTP server: %v", err)
	}
	defer server1.Stop(context.Background())
	
	// 等待第一个服务端启动
	time.Sleep(100 * time.Millisecond)
	
	// 尝试在同一端口启动第二个HTTP服务端
	config2 := httpPkg.NewHTTPServerConfig()
	config2.BaseConfig.Port = 38080 // 相同端口
	server2 := httpPkg.NewHTTPServer(config2, logger, metricsCollector)
	
	// 这应该失败
	if err := server2.Start(ctx); err == nil {
		t.Error("Expected second server to fail due to port conflict")
		server2.Stop(context.Background())
	}
}

// TestServerGracefulShutdown 测试优雅关闭
func TestServerGracefulShutdown(t *testing.T) {
	logger := logging.NewLogger("info")
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建HTTP服务端
	config := httpPkg.NewHTTPServerConfig()
	config.BaseConfig.Port = 48080
	server := httpPkg.NewHTTPServer(config, logger, metricsCollector)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	
	// 等待服务端启动
	time.Sleep(100 * time.Millisecond)
	
	// 验证服务端正在运行
	if !server.IsRunning() {
		t.Fatal("Server should be running")
	}
	
	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	
	if err := server.Stop(shutdownCtx); err != nil {
		t.Errorf("Failed to gracefully stop server: %v", err)
	}
	
	// 验证服务端已停止
	if server.IsRunning() {
		t.Error("Server should be stopped")
	}
}

// BenchmarkMultiProtocolServers 多协议服务端基准测试
func BenchmarkMultiProtocolServers(b *testing.B) {
	logger := logging.NewLogger("error")
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建并启动HTTP服务端
	httpConfig := httpPkg.NewHTTPServerConfig()
	httpConfig.BaseConfig.Port = 58080
	httpServer := httpPkg.NewHTTPServer(httpConfig, logger, metricsCollector)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := httpServer.Start(ctx); err != nil {
		b.Fatalf("Failed to start HTTP server: %v", err)
	}
	defer httpServer.Stop(context.Background())
	
	// 创建并启动gRPC服务端
	grpcConfig := grpc.NewGRPCServerConfig()
	grpcConfig.BaseConfig.Port = 50053
	grpcServer := grpc.NewGRPCServer(grpcConfig, logger, metricsCollector)
	
	if err := grpcServer.Start(ctx); err != nil {
		b.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer grpcServer.Stop(context.Background())
	
	// 等待服务端启动
	time.Sleep(100 * time.Millisecond)
	
	b.ResetTimer()
	
	// 并发测试HTTP和gRPC
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{Timeout: 5 * time.Second}
		for pb.Next() {
			// 交替请求HTTP和gRPC
			if pb.Next() {
				// HTTP请求
				resp, err := client.Get("http://localhost:58080/health")
				if err == nil {
					resp.Body.Close()
				}
			}
			
			if pb.Next() {
				// gRPC请求
				resp, err := client.Get("http://localhost:50053/")
				if err == nil {
					resp.Body.Close()
				}
			}
		}
	})
}