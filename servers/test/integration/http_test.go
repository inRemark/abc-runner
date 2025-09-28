package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
	httpPkg "abc-runner/servers/pkg/http"
)

// TestHTTPServerIntegration HTTP服务端集成测试
func TestHTTPServerIntegration(t *testing.T) {
	// 创建日志和指标收集器
	logger := logging.NewLogger("info")
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建配置
	config := httpPkg.NewHTTPServerConfig()
	config.BaseConfig.Port = 18080 // 使用测试端口
	
	// 创建服务端
	server := httpPkg.NewHTTPServer(config, logger, metricsCollector)
	
	// 启动服务端
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	
	// 确保服务端关闭
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Stop(shutdownCtx)
	}()
	
	// 等待服务端启动
	time.Sleep(100 * time.Millisecond)
	
	// 测试健康检查端点
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18080/health")
		if err != nil {
			t.Fatalf("Failed to call health endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
	
	// 测试根端点
	t.Run("RootEndpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18080/")
		if err != nil {
			t.Fatalf("Failed to call root endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
	
	// 测试指标端点
	t.Run("MetricsEndpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18080/metrics")
		if err != nil {
			t.Fatalf("Failed to call metrics endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

// BenchmarkHTTPServer HTTP服务端基准测试
func BenchmarkHTTPServer(b *testing.B) {
	// 创建日志和指标收集器
	logger := logging.NewLogger("error") // 减少日志输出
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建配置
	config := httpPkg.NewHTTPServerConfig()
	config.BaseConfig.Port = 18081 // 使用不同端口
	
	// 创建服务端
	server := httpPkg.NewHTTPServer(config, logger, metricsCollector)
	
	// 启动服务端
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := server.Start(ctx); err != nil {
		b.Fatalf("Failed to start HTTP server: %v", err)
	}
	
	// 确保服务端关闭
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Stop(shutdownCtx)
	}()
	
	// 等待服务端启动
	time.Sleep(100 * time.Millisecond)
	
	// 基准测试
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{Timeout: 5 * time.Second}
		for pb.Next() {
			resp, err := client.Get("http://localhost:18081/health")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}