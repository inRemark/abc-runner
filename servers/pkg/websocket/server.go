package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// Version 常量定义
const (
	ServerVersion = "1.0.0"
)

// WebSocketServer WebSocket服务端实现
type WebSocketServer struct {
	*common.BaseServer

	config            *WebSocketServerConfig
	httpServer        *http.Server
	upgrader          *websocket.Upgrader
	connectionManager *ConnectionManager
	mux               *http.ServeMux

	// 统计信息
	upgradeCount   int64
	broadcastCount int64
	mutex          sync.RWMutex
}

// NewWebSocketServer 创建WebSocket服务端
func NewWebSocketServer(config *WebSocketServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *WebSocketServer {
	baseServer := common.NewBaseServer("websocket", config, logger, metricsCollector)

	// 创建连接管理器
	connectionManager := NewConnectionManager(config, logger, metricsCollector)

	// 创建WebSocket升级器
	upgrader := &websocket.Upgrader{
		ReadBufferSize:   config.Upgrader.ReadBufferSize,
		WriteBufferSize:  config.Upgrader.WriteBufferSize,
		HandshakeTimeout: config.Upgrader.HandshakeTimeout,
		CheckOrigin:      createOriginChecker(config),
		Subprotocols:     config.Upgrader.Subprotocols,
	}

	// 启用压缩（如果配置）
	if config.Upgrader.EnableCompression {
		upgrader.EnableCompression = true
	}

	server := &WebSocketServer{
		BaseServer:        baseServer,
		config:            config,
		upgrader:          upgrader,
		connectionManager: connectionManager,
		mux:               http.NewServeMux(),
	}

	// 设置HTTP服务器
	server.httpServer = &http.Server{
		Addr:           config.GetAddress(),
		Handler:        server.buildHandler(),
		ReadTimeout:    config.HTTPServer.ReadTimeout,
		WriteTimeout:   config.HTTPServer.WriteTimeout,
		IdleTimeout:    config.HTTPServer.IdleTimeout,
		MaxHeaderBytes: config.HTTPServer.MaxHeaderBytes,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// Start 启动WebSocket服务端
func (ws *WebSocketServer) Start(ctx context.Context) error {
	if ws.IsRunning() {
		return fmt.Errorf("WebSocket server is already running")
	}

	ws.LogInfo("Starting WebSocket server", map[string]interface{}{
		"address": ws.config.GetAddress(),
		"path":    ws.config.Upgrader.Path,
	})

	// 设置监听器
	listener, err := net.Listen("tcp", ws.config.GetAddress())
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", ws.config.GetAddress(), err)
	}

	// 启动HTTP服务器
	go func() {
		if err := ws.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			ws.LogError("WebSocket server error", err)
		}
	}()

	ws.SetRunning(true)
	return nil
}

// Stop 停止WebSocket服务端
func (ws *WebSocketServer) Stop(ctx context.Context) error {
	if !ws.IsRunning() {
		return fmt.Errorf("WebSocket server is not running")
	}

	ws.LogInfo("Stopping WebSocket server", map[string]interface{}{
		"address": ws.config.GetAddress(),
	})

	// 关闭所有WebSocket连接
	ws.connectionManager.Shutdown()

	// 创建关闭上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ws.config.HTTPServer.ShutdownTimeout)
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := ws.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown WebSocket server: %w", err)
	}

	ws.SetRunning(false)
	return ws.Shutdown(ctx)
}

// buildHandler 构建请求处理器
func (ws *WebSocketServer) buildHandler() http.Handler {
	return ws.mux
}

// registerRoutes 注册路由
func (ws *WebSocketServer) registerRoutes() {
	// WebSocket升级端点
	ws.mux.HandleFunc(ws.config.Upgrader.Path, ws.handleWebSocketUpgrade)

	// 健康检查端点
	ws.mux.HandleFunc("/health", ws.handleHealth)

	// 指标端点
	ws.mux.HandleFunc("/metrics", ws.handleMetrics)

	// 连接统计端点
	ws.mux.HandleFunc("/stats", ws.handleStats)

	// 广播端点
	ws.mux.HandleFunc("/broadcast", ws.handleBroadcast)

	// 根路径
	ws.mux.HandleFunc("/", ws.handleRoot)
}

// handleWebSocketUpgrade 处理WebSocket升级请求
func (ws *WebSocketServer) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
	// 检查连接数限制
	if ws.connectionManager.GetConnectionCount() >= ws.config.Connection.MaxConnections {
		http.Error(w, "Maximum connections reached", http.StatusServiceUnavailable)
		return
	}

	// 升级连接
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.LogError("Failed to upgrade connection", err, map[string]interface{}{
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		})

		if ws.GetMetricsCollector() != nil {
			ws.GetMetricsCollector().RecordError("websocket", "upgrade", "upgrade_failed")
		}
		return
	}

	// 创建连接对象
	wsConn := NewConnection(conn, ws.config, ws.GetLogger(), ws.GetMetricsCollector())

	// 添加到连接管理器
	if err := ws.connectionManager.AddConnection(wsConn); err != nil {
		ws.LogError("Failed to add connection", err, map[string]interface{}{
			"connection_id": wsConn.GetID(),
		})
		wsConn.Close()
		return
	}

	// 更新统计信息
	ws.mutex.Lock()
	ws.upgradeCount++
	ws.mutex.Unlock()

	// 记录指标
	if ws.GetMetricsCollector() != nil {
		ws.GetMetricsCollector().RecordRequest("websocket", "upgrade", 0, true)
	}

	ws.LogInfo("WebSocket connection established", map[string]interface{}{
		"connection_id": wsConn.GetID(),
		"remote_addr":   r.RemoteAddr,
		"user_agent":    r.UserAgent(),
	})

	// 设置连接关闭回调
	go func() {
		<-wsConn.done
		ws.connectionManager.RemoveConnection(wsConn.GetID())
	}()
}

// handleHealth 处理健康检查请求
func (ws *WebSocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":      "ok",
		"protocol":    "websocket",
		"timestamp":   time.Now().Unix(),
		"running":     ws.IsRunning(),
		"connections": ws.connectionManager.GetConnectionCount(),
	}

	json.NewEncoder(w).Encode(health)
}

// handleMetrics 处理指标请求
func (ws *WebSocketServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := ws.GetMetrics()
	if jsonData, err := json.Marshal(metrics); err == nil {
		w.Write(jsonData)
	} else {
		ws.LogError("Failed to marshal metrics", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleStats 处理连接统计请求
func (ws *WebSocketServer) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	connections := ws.connectionManager.GetAllConnections()
	stats := map[string]interface{}{
		"total_connections":  len(connections),
		"active_connections": ws.connectionManager.GetConnectionCount(),
		"connections":        connections,
		"timestamp":          time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(stats)
}

// handleBroadcast 处理广播请求
func (ws *WebSocketServer) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取消息内容
	var request struct {
		Type    string `json:"type"` // "text" or "binary"
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.Message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// 确定消息类型
	messageType := websocket.TextMessage
	if request.Type == "binary" {
		messageType = websocket.BinaryMessage
	}

	// 广播消息
	successCount := ws.connectionManager.BroadcastMessage(messageType, []byte(request.Message))

	// 更新统计信息
	ws.mutex.Lock()
	ws.broadcastCount++
	ws.mutex.Unlock()

	// 记录指标
	if ws.GetMetricsCollector() != nil {
		ws.GetMetricsCollector().RecordRequest("websocket", "broadcast", 0, true)
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":           true,
		"broadcast_count":   successCount,
		"total_connections": ws.connectionManager.GetConnectionCount(),
		"message_type":      request.Type,
		"message_size":      len(request.Message),
		"timestamp":         time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleRoot 处理根路径请求
func (ws *WebSocketServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message":  "WebSocket Test Server",
		"protocol": "websocket",
		"version":  ServerVersion,
		"endpoints": map[string]interface{}{
			"websocket": ws.config.Upgrader.Path,
			"health":    "/health",
			"metrics":   "/metrics",
			"stats":     "/stats",
			"broadcast": "/broadcast",
		},
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetWebSocketConfig 获取WebSocket配置
func (ws *WebSocketServer) GetWebSocketConfig() *WebSocketServerConfig {
	return ws.config
}

// GetConnectionManager 获取连接管理器
func (ws *WebSocketServer) GetConnectionManager() *ConnectionManager {
	return ws.connectionManager
}

// GetMetrics 获取WebSocket服务端指标
func (ws *WebSocketServer) GetMetrics() map[string]interface{} {
	baseMetrics := ws.BaseServer.GetMetrics()

	ws.mutex.RLock()
	upgradeCount := ws.upgradeCount
	broadcastCount := ws.broadcastCount
	ws.mutex.RUnlock()

	// 添加WebSocket特定指标
	baseMetrics["upgrade_count"] = upgradeCount
	baseMetrics["broadcast_count"] = broadcastCount
	baseMetrics["websocket_path"] = ws.config.Upgrader.Path
	baseMetrics["max_connections"] = ws.config.Connection.MaxConnections
	baseMetrics["current_connections"] = ws.connectionManager.GetConnectionCount()
	baseMetrics["heartbeat_enabled"] = ws.config.Heartbeat.Enabled
	baseMetrics["compression_enabled"] = ws.config.Upgrader.EnableCompression
	baseMetrics["echo_mode"] = ws.config.Message.EchoMode

	// 连接统计
	connections := ws.connectionManager.GetAllConnections()
	totalBytesSent := int64(0)
	totalBytesRecv := int64(0)
	totalMessagesSent := int64(0)
	totalMessagesRecv := int64(0)

	for _, conn := range connections {
		totalBytesSent += conn.BytesSent
		totalBytesRecv += conn.BytesRecv
		totalMessagesSent += conn.MessagesSent
		totalMessagesRecv += conn.MessagesRecv
	}

	baseMetrics["total_bytes_sent"] = totalBytesSent
	baseMetrics["total_bytes_recv"] = totalBytesRecv
	baseMetrics["total_messages_sent"] = totalMessagesSent
	baseMetrics["total_messages_recv"] = totalMessagesRecv

	// 配置信息
	baseMetrics["config"] = map[string]interface{}{
		"read_buffer_size":  ws.config.Upgrader.ReadBufferSize,
		"write_buffer_size": ws.config.Upgrader.WriteBufferSize,
		"handshake_timeout": ws.config.Upgrader.HandshakeTimeout.String(),
		"max_message_size":  ws.config.Message.MaxMessageSize,
		"ping_interval":     ws.config.Heartbeat.PingInterval.String(),
		"pong_timeout":      ws.config.Heartbeat.PongTimeout.String(),
		"idle_timeout":      ws.config.Connection.IdleTimeout.String(),
		"cleanup_interval":  ws.config.Connection.CleanupInterval.String(),
	}

	return baseMetrics
}

// BroadcastText 广播文本消息
func (ws *WebSocketServer) BroadcastText(message string) int {
	successCount := ws.connectionManager.BroadcastMessage(websocket.TextMessage, []byte(message))

	ws.mutex.Lock()
	ws.broadcastCount++
	ws.mutex.Unlock()

	if ws.GetMetricsCollector() != nil {
		ws.GetMetricsCollector().RecordRequest("websocket", "broadcast_text", 0, true)
	}

	return successCount
}

// BroadcastBinary 广播二进制消息
func (ws *WebSocketServer) BroadcastBinary(data []byte) int {
	successCount := ws.connectionManager.BroadcastMessage(websocket.BinaryMessage, data)

	ws.mutex.Lock()
	ws.broadcastCount++
	ws.mutex.Unlock()

	if ws.GetMetricsCollector() != nil {
		ws.GetMetricsCollector().RecordRequest("websocket", "broadcast_binary", 0, true)
	}

	return successCount
}

// GetConnectionStats 获取连接统计信息
func (ws *WebSocketServer) GetConnectionStats() map[string]interface{} {
	connections := ws.connectionManager.GetAllConnections()

	stats := map[string]interface{}{
		"total_connections":  len(connections),
		"active_connections": ws.connectionManager.GetConnectionCount(),
		"max_connections":    ws.config.Connection.MaxConnections,
	}

	// 按状态统计
	stateCount := make(map[string]int)
	for _, conn := range connections {
		stateCount[conn.State]++
	}
	stats["by_state"] = stateCount

	// 计算平均指标
	if len(connections) > 0 {
		totalDuration := time.Duration(0)
		totalBytesSent := int64(0)
		totalBytesRecv := int64(0)

		for _, conn := range connections {
			totalDuration += time.Since(conn.ConnectedAt)
			totalBytesSent += conn.BytesSent
			totalBytesRecv += conn.BytesRecv
		}

		stats["avg_connection_duration"] = (totalDuration / time.Duration(len(connections))).String()
		stats["avg_bytes_sent"] = totalBytesSent / int64(len(connections))
		stats["avg_bytes_recv"] = totalBytesRecv / int64(len(connections))
	}

	return stats
}

// createOriginChecker 创建来源检查器
func createOriginChecker(config *WebSocketServerConfig) func(r *http.Request) bool {
	if !config.Upgrader.CheckOrigin {
		return func(r *http.Request) bool { return true }
	}

	allowedOrigins := make(map[string]bool)
	for _, origin := range config.Upgrader.AllowedOrigins {
		if origin == "*" {
			return func(r *http.Request) bool { return true }
		}
		allowedOrigins[origin] = true
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return allowedOrigins[origin]
	}
}
