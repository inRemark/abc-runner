package websocket

import (
	"testing"
	"time"

	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
)

func TestConnectionState(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateConnecting, "connecting"},
		{StateConnected, "connected"},
		{StateClosing, "closing"},
		{StateClosed, "closed"},
	}

	for _, test := range tests {
		if test.state.String() != test.expected {
			t.Errorf("Expected state %d to return '%s', got '%s'", 
				test.state, test.expected, test.state.String())
		}
	}

	// 测试未知状态
	unknownState := ConnectionState(999)
	if unknownState.String() != "unknown" {
		t.Errorf("Expected unknown state to return 'unknown', got '%s'", unknownState.String())
	}
}

func TestGenerateConnectionID(t *testing.T) {
	// 生成多个ID，确保它们是唯一的
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateConnectionID()
		
		// 检查ID格式
		if len(id) < 5 {
			t.Errorf("Connection ID too short: '%s'", id)
		}
		
		if id[:3] != "ws-" {
			t.Errorf("Connection ID should start with 'ws-': '%s'", id)
		}
		
		// 检查唯一性
		if ids[id] {
			t.Errorf("Duplicate connection ID generated: '%s'", id)
		}
		ids[id] = true
	}
}

func TestIsTextMessage(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
		name     string
	}{
		{[]byte("hello world"), true, "simple text"},
		{[]byte("Hello\nWorld\t!"), true, "text with newlines and tabs"},
		{[]byte("你好世界"), true, "UTF-8 text"},
		{[]byte{0x00, 0x01, 0x02}, false, "binary data with control chars"},
		{[]byte{0xFF, 0xFE}, false, "binary data"},
		{[]byte(""), true, "empty data"},
		{[]byte("test\r\n"), true, "text with CRLF"},
	}

	for _, test := range tests {
		result := isTextMessage(test.data)
		if result != test.expected {
			t.Errorf("Test '%s': expected %v, got %v for data %v", 
				test.name, test.expected, result, test.data)
		}
	}
}

func TestNewConnectionManager(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)

	if cm == nil {
		t.Fatal("Connection manager should not be nil")
	}

	if cm.maxConnections != config.Connection.MaxConnections {
		t.Errorf("Expected max connections %d, got %d", 
			config.Connection.MaxConnections, cm.maxConnections)
	}

	if cm.cleanupInterval != config.Connection.CleanupInterval {
		t.Errorf("Expected cleanup interval %v, got %v", 
			config.Connection.CleanupInterval, cm.cleanupInterval)
	}

	if len(cm.connections) != 0 {
		t.Error("New connection manager should have no connections")
	}

	// 测试清理
	cm.Shutdown()
}

func TestConnectionManagerBasicOperations(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)
	defer cm.Shutdown()

	// 测试初始状态
	if cm.GetConnectionCount() != 0 {
		t.Error("Initial connection count should be 0")
	}

	// 创建模拟连接
	mockConn := &Connection{
		id:           "test-conn-1",
		remoteAddr:   "127.0.0.1:12345",
		connectedAt:  time.Now(),
		state:        StateConnected,
		done:         make(chan struct{}),
		config:       config,
		logger:       logger,
		metricsCollector: metrics,
	}

	// 测试添加连接
	err := cm.AddConnection(mockConn)
	if err != nil {
		t.Errorf("Failed to add connection: %v", err)
	}

	if cm.GetConnectionCount() != 1 {
		t.Error("Connection count should be 1 after adding")
	}

	// 测试获取连接
	retrievedConn, exists := cm.GetConnection("test-conn-1")
	if !exists {
		t.Error("Connection should exist")
	}

	if retrievedConn.id != "test-conn-1" {
		t.Error("Retrieved connection should have correct ID")
	}

	// 测试获取所有连接
	allConnections := cm.GetAllConnections()
	if len(allConnections) != 1 {
		t.Error("Should have exactly one connection")
	}

	// 测试移除连接
	cm.RemoveConnection("test-conn-1")
	if cm.GetConnectionCount() != 0 {
		t.Error("Connection count should be 0 after removal")
	}

	_, exists = cm.GetConnection("test-conn-1")
	if exists {
		t.Error("Connection should not exist after removal")
	}
}

func TestConnectionManagerMaxConnections(t *testing.T) {
	config := NewWebSocketServerConfig()
	config.Connection.MaxConnections = 2 // 设置最大连接数为2
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)
	defer cm.Shutdown()

	// 添加第一个连接
	conn1 := &Connection{
		id:    "conn-1",
		state: StateConnected,
		done:  make(chan struct{}),
		config: config,
		logger: logger,
		metricsCollector: metrics,
	}
	err := cm.AddConnection(conn1)
	if err != nil {
		t.Errorf("Failed to add first connection: %v", err)
	}

	// 添加第二个连接
	conn2 := &Connection{
		id:    "conn-2", 
		state: StateConnected,
		done:  make(chan struct{}),
		config: config,
		logger: logger,
		metricsCollector: metrics,
	}
	err = cm.AddConnection(conn2)
	if err != nil {
		t.Errorf("Failed to add second connection: %v", err)
	}

	// 尝试添加第三个连接（应该失败）
	conn3 := &Connection{
		id:    "conn-3",
		state: StateConnected,
		done:  make(chan struct{}),
		config: config,
		logger: logger,
		metricsCollector: metrics,
	}
	err = cm.AddConnection(conn3)
	if err == nil {
		t.Error("Adding third connection should fail due to max connections limit")
	}

	if cm.GetConnectionCount() != 2 {
		t.Errorf("Connection count should be 2, got %d", cm.GetConnectionCount())
	}
}

func TestConnectionManagerBroadcast(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)
	defer cm.Shutdown()

	// 添加多个模拟连接
	for i := 0; i < 3; i++ {
		conn := &Connection{
			id:          GenerateConnectionID(),
			state:       StateConnected,
			sendQueue:   make(chan []byte, 10),
			done:        make(chan struct{}),
			config:      config,
			logger:      logger,
			metricsCollector: metrics,
		}
		cm.AddConnection(conn)
	}

	// 测试广播
	message := []byte("broadcast test message")
	successCount := cm.BroadcastMessage(1, message) // TextMessage

	if successCount != 3 {
		t.Errorf("Expected 3 successful broadcasts, got %d", successCount)
	}

	// 验证每个连接都收到了消息
	for _, conn := range cm.connections {
		select {
		case receivedMessage := <-conn.sendQueue:
			if string(receivedMessage) != string(message) {
				t.Error("Received message doesn't match sent message")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Connection should have received broadcast message")
		}
	}
}

func TestConnectionManagerCloseAll(t *testing.T) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)

	// 添加多个连接
	connCount := 5
	for i := 0; i < connCount; i++ {
		conn := &Connection{
			id:    GenerateConnectionID(),
			state: StateConnected,
			done:  make(chan struct{}),
			config: config,
			logger: logger,
			metricsCollector: metrics,
		}
		cm.AddConnection(conn)
	}

	if cm.GetConnectionCount() != connCount {
		t.Errorf("Expected %d connections, got %d", connCount, cm.GetConnectionCount())
	}

	// 关闭所有连接
	cm.CloseAll()

	if cm.GetConnectionCount() != 0 {
		t.Error("All connections should be closed")
	}

	allConnections := cm.GetAllConnections()
	if len(allConnections) != 0 {
		t.Error("No connections should remain")
	}
}

// Benchmark tests
func BenchmarkGenerateConnectionID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateConnectionID()
	}
}

func BenchmarkIsTextMessage(b *testing.B) {
	testData := []byte("This is a test message for benchmarking")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		isTextMessage(testData)
	}
}

func BenchmarkConnectionManagerAddRemove(b *testing.B) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)
	defer cm.Shutdown()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn := &Connection{
			id:    GenerateConnectionID(),
			state: StateConnected,
			config: config,
			logger: logger,
			metricsCollector: metrics,
		}
		
		cm.AddConnection(conn)
		cm.RemoveConnection(conn.id)
	}
}

func BenchmarkConnectionManagerBroadcast(b *testing.B) {
	config := NewWebSocketServerConfig()
	logger := logging.NewLogger("info")
	metrics := monitoring.NewMetricsCollector()

	cm := NewConnectionManager(config, logger, metrics)
	defer cm.Shutdown()

	// 添加100个连接进行基准测试
	for i := 0; i < 100; i++ {
		conn := &Connection{
			id:        GenerateConnectionID(),
			state:     StateConnected,
			sendQueue: make(chan []byte, 100),
			done:      make(chan struct{}),
			config:    config,
			logger:    logger,
			metricsCollector: metrics,
		}
		cm.AddConnection(conn)
	}

	message := []byte("benchmark test message")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.BroadcastMessage(1, message)
	}
}