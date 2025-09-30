package websocket

import (
	"testing"
	"time"
)

func TestNewWebSocketServerConfig(t *testing.T) {
	config := NewWebSocketServerConfig()

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// 测试基础配置
	if config.GetProtocol() != "websocket" {
		t.Errorf("Expected protocol 'websocket', got '%s'", config.GetProtocol())
	}

	if config.GetHost() != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", config.GetHost())
	}

	if config.GetPort() != 7070 {
		t.Errorf("Expected port 7070, got %d", config.GetPort())
	}

	// 测试升级器配置
	if config.Upgrader.Path != "/ws" {
		t.Errorf("Expected path '/ws', got '%s'", config.Upgrader.Path)
	}

	if config.Upgrader.ReadBufferSize != 4096 {
		t.Errorf("Expected read buffer size 4096, got %d", config.Upgrader.ReadBufferSize)
	}

	if config.Upgrader.WriteBufferSize != 4096 {
		t.Errorf("Expected write buffer size 4096, got %d", config.Upgrader.WriteBufferSize)
	}

	// 测试连接配置
	if config.Connection.MaxConnections != 1000 {
		t.Errorf("Expected max connections 1000, got %d", config.Connection.MaxConnections)
	}

	// 测试心跳配置
	if !config.Heartbeat.Enabled {
		t.Error("Expected heartbeat to be enabled")
	}

	if config.Heartbeat.PingInterval != 30*time.Second {
		t.Errorf("Expected ping interval 30s, got %v", config.Heartbeat.PingInterval)
	}

	// 测试消息配置
	if !config.Message.EchoMode {
		t.Error("Expected echo mode to be enabled")
	}

	if config.Message.MaxMessageSize != 1024*1024 {
		t.Errorf("Expected max message size 1MB, got %d", config.Message.MaxMessageSize)
	}
}

func TestWebSocketServerConfigValidate(t *testing.T) {
	// 测试有效配置
	validConfig := NewWebSocketServerConfig()
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// 测试无效端口
	invalidPortConfig := NewWebSocketServerConfig()
	invalidPortConfig.BaseConfig.Port = 0
	if err := invalidPortConfig.Validate(); err == nil {
		t.Error("Config with invalid port should fail validation")
	}

	// 测试无效路径
	invalidPathConfig := NewWebSocketServerConfig()
	invalidPathConfig.Upgrader.Path = ""
	if err := invalidPathConfig.Validate(); err == nil {
		t.Error("Config with empty path should fail validation")
	}

	// 测试无效缓冲区大小
	invalidBufferConfig := NewWebSocketServerConfig()
	invalidBufferConfig.Upgrader.ReadBufferSize = 0
	if err := invalidBufferConfig.Validate(); err == nil {
		t.Error("Config with zero read buffer size should fail validation")
	}

	// 测试无效最大连接数
	invalidMaxConnConfig := NewWebSocketServerConfig()
	invalidMaxConnConfig.Connection.MaxConnections = 0
	if err := invalidMaxConnConfig.Validate(); err == nil {
		t.Error("Config with zero max connections should fail validation")
	}

	// 测试心跳配置
	invalidHeartbeatConfig := NewWebSocketServerConfig()
	invalidHeartbeatConfig.Heartbeat.Enabled = true
	invalidHeartbeatConfig.Heartbeat.PingInterval = 0
	if err := invalidHeartbeatConfig.Validate(); err == nil {
		t.Error("Config with zero ping interval should fail validation when heartbeat is enabled")
	}

	// 测试消息大小限制
	invalidMsgSizeConfig := NewWebSocketServerConfig()
	invalidMsgSizeConfig.Message.MaxMessageSize = 0
	if err := invalidMsgSizeConfig.Validate(); err == nil {
		t.Error("Config with zero max message size should fail validation")
	}

	// 测试消息大小过大
	tooLargeMsgSizeConfig := NewWebSocketServerConfig()
	tooLargeMsgSizeConfig.Message.MaxMessageSize = 200 * 1024 * 1024 // 200MB
	if err := tooLargeMsgSizeConfig.Validate(); err == nil {
		t.Error("Config with too large message size should fail validation")
	}
}

func TestWebSocketServerConfigClone(t *testing.T) {
	originalConfig := NewWebSocketServerConfig()
	originalConfig.Upgrader.AllowedOrigins = []string{"http://example.com", "https://example.org"}
	originalConfig.Upgrader.Subprotocols = []string{"chat", "echo"}

	clonedConfig := originalConfig.Clone().(*WebSocketServerConfig)

	// 测试基本字段克隆
	if clonedConfig.GetProtocol() != originalConfig.GetProtocol() {
		t.Error("Protocol should be cloned correctly")
	}

	if clonedConfig.GetPort() != originalConfig.GetPort() {
		t.Error("Port should be cloned correctly")
	}

	// 测试切片深拷贝
	if len(clonedConfig.Upgrader.AllowedOrigins) != len(originalConfig.Upgrader.AllowedOrigins) {
		t.Error("AllowedOrigins slice should be cloned correctly")
	}

	// 修改原始配置，确保克隆的配置不受影响
	originalConfig.Upgrader.AllowedOrigins[0] = "http://modified.com"
	if clonedConfig.Upgrader.AllowedOrigins[0] == "http://modified.com" {
		t.Error("Cloned config should not be affected by changes to original")
	}

	// 测试子协议切片
	if len(clonedConfig.Upgrader.Subprotocols) != len(originalConfig.Upgrader.Subprotocols) {
		t.Error("Subprotocols slice should be cloned correctly")
	}
}

func TestConnectionInfo(t *testing.T) {
	info := ConnectionInfo{
		ID:           "test-conn-1",
		RemoteAddr:   "127.0.0.1:12345",
		UserAgent:    "test-client/1.0",
		ConnectedAt:  time.Now(),
		BytesSent:    1024,
		BytesRecv:    2048,
		MessagesSent: 10,
		MessagesRecv: 15,
		State:        "connected",
	}

	if info.ID != "test-conn-1" {
		t.Errorf("Expected ID 'test-conn-1', got '%s'", info.ID)
	}

	if info.BytesSent != 1024 {
		t.Errorf("Expected bytes sent 1024, got %d", info.BytesSent)
	}

	if info.State != "connected" {
		t.Errorf("Expected state 'connected', got '%s'", info.State)
	}
}

func TestMessageInfo(t *testing.T) {
	data := []byte("test message")
	info := MessageInfo{
		ConnectionID: "test-conn-1",
		Type:         1, // TextMessage
		Direction:    "in",
		Size:         len(data),
		Timestamp:    time.Now(),
		Data:         data,
	}

	if info.ConnectionID != "test-conn-1" {
		t.Errorf("Expected connection ID 'test-conn-1', got '%s'", info.ConnectionID)
	}

	if info.Type != 1 {
		t.Errorf("Expected message type 1, got %d", info.Type)
	}

	if info.Direction != "in" {
		t.Errorf("Expected direction 'in', got '%s'", info.Direction)
	}

	if info.Size != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), info.Size)
	}

	if string(info.Data) != "test message" {
		t.Errorf("Expected data 'test message', got '%s'", string(info.Data))
	}
}

func TestHeartbeatInfo(t *testing.T) {
	now := time.Now()
	info := HeartbeatInfo{
		ConnectionID: "test-conn-1",
		Type:         "ping",
		Timestamp:    now,
	}

	if info.ConnectionID != "test-conn-1" {
		t.Errorf("Expected connection ID 'test-conn-1', got '%s'", info.ConnectionID)
	}

	if info.Type != "ping" {
		t.Errorf("Expected type 'ping', got '%s'", info.Type)
	}

	if info.Timestamp != now {
		t.Error("Timestamp should match")
	}

	// 测试pong消息（带延迟）
	pongInfo := HeartbeatInfo{
		ConnectionID: "test-conn-1",
		Type:         "pong",
		Timestamp:    now.Add(10 * time.Millisecond),
		Latency:      10 * time.Millisecond,
	}

	if pongInfo.Type != "pong" {
		t.Errorf("Expected type 'pong', got '%s'", pongInfo.Type)
	}

	if pongInfo.Latency != 10*time.Millisecond {
		t.Errorf("Expected latency 10ms, got %v", pongInfo.Latency)
	}
}

// Benchmark tests
func BenchmarkNewWebSocketServerConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewWebSocketServerConfig()
	}
}

func BenchmarkWebSocketServerConfigValidate(b *testing.B) {
	config := NewWebSocketServerConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}

func BenchmarkWebSocketServerConfigClone(b *testing.B) {
	config := NewWebSocketServerConfig()
	config.Upgrader.AllowedOrigins = []string{"http://example.com", "https://example.org"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config.Clone()
	}
}