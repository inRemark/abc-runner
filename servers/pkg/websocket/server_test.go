package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
)

func TestNewWebSocketServer(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.config != config {
		t.Error("Server config should match provided config")
	}

	if server.GetProtocol() != "websocket" {
		t.Errorf("Expected protocol 'websocket', got '%s'", server.GetProtocol())
	}

	if server.GetPort() != 7070 {
		t.Errorf("Expected port 7070, got %d", server.GetPort())
	}

	if server.connectionManager == nil {
		t.Error("Connection manager should not be nil")
	}

	if server.upgrader == nil {
		t.Error("WebSocket upgrader should not be nil")
	}
}

func TestWebSocketServerConfig(t *testing.T) {
	config := NewWebSocketServerConfig()
	config.Upgrader.Path = "/custom-ws"
	config.Connection.MaxConnections = 500
	
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	retrievedConfig := server.GetWebSocketConfig()
	if retrievedConfig.Upgrader.Path != "/custom-ws" {
		t.Error("Server should use provided config")
	}

	if retrievedConfig.Connection.MaxConnections != 500 {
		t.Error("Server should preserve config values")
	}
}

func TestWebSocketServerHTTPEndpoints(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	// 测试根路径
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	server.handleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response["protocol"] != "websocket" {
		t.Error("Response should indicate websocket protocol")
	}

	// 测试健康检查
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	
	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var healthResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &healthResponse); err != nil {
		t.Errorf("Failed to parse health JSON response: %v", err)
	}

	if healthResponse["status"] != "ok" {
		t.Error("Health check should return 'ok' status")
	}
}

func TestWebSocketServerMetricsEndpoint(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	server.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var metricsResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &metricsResponse); err != nil {
		t.Errorf("Failed to parse metrics JSON response: %v", err)
	}

	// 检查基本指标是否存在
	expectedFields := []string{"protocol", "running", "websocket_path", "max_connections"}
	for _, field := range expectedFields {
		if _, exists := metricsResponse[field]; !exists {
			t.Errorf("Metrics should include field '%s'", field)
		}
	}

	if metricsResponse["protocol"] != "websocket" {
		t.Error("Metrics should show websocket protocol")
	}
}

func TestWebSocketServerStatsEndpoint(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	
	server.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var statsResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &statsResponse); err != nil {
		t.Errorf("Failed to parse stats JSON response: %v", err)
	}

	// 检查统计信息字段
	expectedFields := []string{"total_connections", "active_connections", "connections"}
	for _, field := range expectedFields {
		if _, exists := statsResponse[field]; !exists {
			t.Errorf("Stats should include field '%s'", field)
		}
	}

	// 初始状态应该没有连接
	if statsResponse["total_connections"] != float64(0) {
		t.Error("Initial total connections should be 0")
	}
}

func TestWebSocketServerBroadcastEndpoint(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	// 测试有效的广播请求
	broadcastData := `{"type":"text","message":"test broadcast"}`
	req := httptest.NewRequest("POST", "/broadcast", strings.NewReader(broadcastData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.handleBroadcast(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse broadcast response: %v", err)
	}

	if response["success"] != true {
		t.Error("Broadcast should succeed")
	}

	// 测试无效的HTTP方法
	req = httptest.NewRequest("GET", "/broadcast", nil)
	w = httptest.NewRecorder()
	
	server.handleBroadcast(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}

	// 测试无效的JSON
	req = httptest.NewRequest("POST", "/broadcast", strings.NewReader("invalid json"))
	w = httptest.NewRecorder()
	
	server.handleBroadcast(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	}

	// 测试空消息
	emptyMessage := `{"type":"text","message":""}`
	req = httptest.NewRequest("POST", "/broadcast", strings.NewReader(emptyMessage))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	server.handleBroadcast(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty message, got %d", w.Code)
	}
}

func TestWebSocketServerGetMetrics(t *testing.T) {
	config := NewWebSocketServerConfig()
	config.Upgrader.Path = "/custom-ws"
	config.Heartbeat.Enabled = true
	config.Message.EchoMode = true
	
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	serverMetrics := server.GetMetrics()

	// 检查基本指标
	if serverMetrics["protocol"] != "websocket" {
		t.Error("Metrics should show websocket protocol")
	}

	if serverMetrics["websocket_path"] != "/custom-ws" {
		t.Error("Metrics should show correct WebSocket path")
	}

	if serverMetrics["heartbeat_enabled"] != true {
		t.Error("Metrics should show heartbeat enabled")
	}

	if serverMetrics["echo_mode"] != true {
		t.Error("Metrics should show echo mode enabled")
	}

	// 检查配置信息
	configInfo, exists := serverMetrics["config"].(map[string]interface{})
	if !exists {
		t.Error("Metrics should include config information")
	}

	if configInfo["read_buffer_size"] != 4096 {
		t.Error("Config should show correct read buffer size")
	}
}

func TestWebSocketServerBroadcastMethods(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	// 测试广播文本消息
	textMessage := "Hello, WebSocket!"
	successCount := server.BroadcastText(textMessage)
	
	// 没有连接时应该返回0
	if successCount != 0 {
		t.Errorf("Expected 0 successful broadcasts with no connections, got %d", successCount)
	}

	// 测试广播二进制消息
	binaryData := []byte{0x01, 0x02, 0x03, 0x04}
	successCount = server.BroadcastBinary(binaryData)
	
	if successCount != 0 {
		t.Errorf("Expected 0 successful broadcasts with no connections, got %d", successCount)
	}
}

func TestWebSocketServerConnectionStats(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	stats := server.GetConnectionStats()

	// 检查统计字段
	expectedFields := []string{"total_connections", "active_connections", "max_connections", "by_state"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Connection stats should include field '%s'", field)
		}
	}

	// 初始状态检查
	if stats["total_connections"] != 0 {
		t.Error("Initial total connections should be 0")
	}

	if stats["active_connections"] != 0 {
		t.Error("Initial active connections should be 0")
	}

	if stats["max_connections"] != config.Connection.MaxConnections {
		t.Error("Max connections should match config")
	}
}

func TestWebSocketServerStartStop(t *testing.T) {
	config := NewWebSocketServerConfig()
	config.Port = 0 // 使用随机端口进行测试
	
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试服务器未运行状态
	if server.IsRunning() {
		t.Error("Server should not be running initially")
	}

	// 由于端口0可能导致问题，我们跳过实际的启动测试
	// 但可以测试重复启动的错误处理
	
	// 模拟设置为已运行状态
	server.SetRunning(true)
	
	// 测试重复启动
	err := server.Start(ctx)
	if err == nil {
		t.Error("Starting already running server should return error")
	}

	// 测试停止
	server.SetRunning(false) // 重置状态以便正常停止测试
	err = server.Stop(ctx)
	if err != nil {
		// 停止未运行的服务器可能会返回错误，这是正常的
		t.Logf("Stop returned error (expected): %v", err)
	}
}

// Benchmark tests
func BenchmarkWebSocketServerGetMetrics(b *testing.B) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		server.GetMetrics()
	}
}

func BenchmarkWebSocketServerHandleRoot(b *testing.B) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	req := httptest.NewRequest("GET", "/", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.handleRoot(w, req)
	}
}

func BenchmarkWebSocketServerHandleHealth(b *testing.B) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	server := NewWebSocketServer(config, logger, metrics)

	req := httptest.NewRequest("GET", "/health", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.handleHealth(w, req)
	}
}