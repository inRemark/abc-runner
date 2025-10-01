package connection

import (
	"fmt"
	"testing"
	"time"

	"abc-runner/app/adapters/websocket/config"
)

// TestNewWebSocketConnectionPool 测试连接池创建
func TestNewWebSocketConnectionPool(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.WebSocketConfig
		expectErr bool
	}{
		{
			name:      "Valid config",
			config:    createValidConfig(),
			expectErr: false,
		},
		{
			name:      "Nil config",
			config:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewWebSocketConnectionPool(tt.config)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectErr && pool == nil {
				t.Errorf("Expected pool to be created but got nil")
			}

			if pool != nil {
				// 清理
				pool.Close()
			}
		})
	}
}

// TestWebSocketConnectionPoolStats 测试连接池统计
func TestWebSocketConnectionPoolStats(t *testing.T) {
	config := createValidConfig()
	pool, err := NewWebSocketConnectionPool(config)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	stats := pool.GetStats()
	if stats == nil {
		t.Errorf("Expected stats but got nil")
	}

	expectedKeys := []string{
		"max_connections",
		"current_connections",
		"active_connections",
		"available_connections",
		"total_created",
		"closed",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stats key '%s' but not found", key)
		}
	}

	// 检查基本值
	if stats["max_connections"] != config.Connection.Pool.PoolSize {
		t.Errorf("Expected max_connections to be %d but got %v",
			config.Connection.Pool.PoolSize, stats["max_connections"])
	}

	if stats["closed"].(bool) {
		t.Errorf("Expected pool to not be closed initially")
	}
}

// TestWebSocketConnectionPoolClose 测试连接池关闭
func TestWebSocketConnectionPoolClose(t *testing.T) {
	config := createValidConfig()
	pool, err := NewWebSocketConnectionPool(config)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	// 检查初始状态
	stats := pool.GetStats()
	if stats["closed"].(bool) {
		t.Errorf("Expected pool to not be closed initially")
	}

	// 关闭连接池
	err = pool.Close()
	if err != nil {
		t.Errorf("Expected no error when closing pool, got: %v", err)
	}

	// 检查关闭后的状态
	stats = pool.GetStats()
	if !stats["closed"].(bool) {
		t.Errorf("Expected pool to be closed after Close()")
	}

	// 关闭后尝试获取连接应该失败
	_, err = pool.GetConnection()
	if err == nil {
		t.Errorf("Expected error when getting connection from closed pool")
	}
}

// TestWebSocketConnectionStats 测试连接统计
func TestWebSocketConnectionStats(t *testing.T) {
	// 创建模拟连接
	conn := &WebSocketConnection{
		ID:        "test_conn_1",
		URL:       "ws://localhost:7070/ws",
		isActive:  true,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		// 初始化心跳相关字段
		lastPingTime: time.Now(),
		lastPongTime: time.Now(),
		missedPongs:  0,
	}

	stats := conn.GetStats()
	if stats == nil {
		t.Errorf("Expected stats but got nil")
	}

	expectedKeys := []string{
		"id",
		"url",
		"is_active",
		"created_at",
		"last_used",
		"messages_sent",
		"messages_recv",
		"bytes_sent",
		"bytes_recv",
		"last_ping_time",
		"last_pong_time",
		"missed_pongs",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stats key '%s' but not found", key)
		}
	}

	// 检查基本值
	if stats["id"] != conn.ID {
		t.Errorf("Expected ID to be %s but got %v", conn.ID, stats["id"])
	}

	if stats["url"] != conn.URL {
		t.Errorf("Expected URL to be %s but got %v", conn.URL, stats["url"])
	}

	if stats["is_active"] != conn.isActive {
		t.Errorf("Expected is_active to be %v but got %v", conn.isActive, stats["is_active"])
	}
}

// TestWebSocketConnectionConcurrency 测试连接并发安全
func TestWebSocketConnectionConcurrency(t *testing.T) {
	conn := &WebSocketConnection{
		ID:       "test_conn_concurrent",
		URL:      "ws://localhost:8080/ws",
		isActive: true,
		sendChan: make(chan []byte, 100),
		done:     make(chan struct{}),
	}

	// 并发发送消息
	concurrency := 10
	messageCount := 100

	// 启动发送协程
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			for j := 0; j < messageCount; j++ {
				// 模拟发送消息
				message := []byte(fmt.Sprintf("worker_%d_message_%d", workerID, j))
				select {
				case conn.sendChan <- message:
					// 消息发送成功
				case <-time.After(time.Second):
					t.Errorf("Timeout sending message from worker %d", workerID)
				}
			}
		}(i)
	}

	// 并发获取统计信息
	statsChan := make(chan map[string]interface{}, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			stats := conn.GetStats()
			statsChan <- stats
		}()
	}

	// 收集统计结果
	for i := 0; i < concurrency; i++ {
		select {
		case stats := <-statsChan:
			if stats == nil {
				t.Errorf("Expected stats but got nil from concurrent access")
			}
		case <-time.After(time.Second):
			t.Errorf("Timeout getting stats from concurrent access")
		}
	}

	// 清理
	close(conn.done)
}

// TestConnectionHeartbeat 测试连接心跳功能
func TestConnectionHeartbeat(t *testing.T) {
	conn := &WebSocketConnection{
		ID:           "test_heartbeat_conn",
		URL:          "ws://localhost:7070/ws",
		isActive:     true,
		lastPingTime: time.Now(),
		lastPongTime: time.Now(),
		missedPongs:  0,
	}

	// 测试心跳状态更新
	originalPingTime := conn.lastPingTime
	time.Sleep(time.Millisecond) // 确保时间差异

	// 模拟发送心跳
	conn.lastPingTime = time.Now()

	if !conn.lastPingTime.After(originalPingTime) {
		t.Errorf("Expected ping time to be updated")
	}

	// 模拟接收Pong响应
	originalMissedPongs := conn.missedPongs
	conn.missedPongs = 5 // 模拟错过的心跳

	// 验证missedPongs已经增加
	if conn.missedPongs == originalMissedPongs {
		t.Errorf("Expected missed pongs to be different from original value")
	}

	// 模拟收到Pong，重置计数
	conn.lastPongTime = time.Now()
	conn.missedPongs = 0

	if conn.missedPongs != 0 {
		t.Errorf("Expected missed pongs to be reset to 0 but got %d", conn.missedPongs)
	}
}

// createValidConfig 创建有效的测试配置
func createValidConfig() *config.WebSocketConfig {
	return &config.WebSocketConfig{
		Protocol: "websocket",
		Connection: config.ConnectionConfig{
			URL:     "ws://localhost:7070/ws",
			Timeout: 30 * time.Second,
			Pool: config.PoolConfig{
				PoolSize:          5,
				MinIdle:           1,
				MaxIdle:           3,
				IdleTimeout:       300 * time.Second,
				ConnectionTimeout: 30 * time.Second,
			},
		},
		BenchMark: config.BenchmarkConfig{
			Total:     100,
			Parallels: 10,
			DataSize:  1024,
			Duration:  60 * time.Second,
		},
		WebSocketSpecific: config.WebSocketSpecificConfig{
			MessageType:          "text",
			Compression:          false,
			AutoReconnect:        true,
			ReconnectInterval:    5 * time.Second,
			MaxReconnectAttempts: 3,
			BufferSize:           4096,
			WriteTimeout:         10 * time.Second,
			ReadTimeout:          10 * time.Second,
			Heartbeat: config.HeartbeatConfig{
				Enabled:   true,
				Interval:  30 * time.Second,
				Timeout:   10 * time.Second,
				MaxMissed: 3,
			},
		},
	}
}
