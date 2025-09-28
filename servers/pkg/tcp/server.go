package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// TCPServer TCP服务端实现
type TCPServer struct {
	*common.BaseServer
	
	config            *TCPServerConfig
	listener          net.Listener
	connectionManager *ConnectionManager
	handler           interfaces.ConnectionHandler
	
	// 并发控制
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// NewTCPServer 创建TCP服务端
func NewTCPServer(config *TCPServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *TCPServer {
	baseServer := common.NewBaseServer("tcp", config, logger, metricsCollector)
	
	server := &TCPServer{
		BaseServer:        baseServer,
		config:            config,
		connectionManager: NewConnectionManager(config.MaxConnections, logger, metricsCollector),
	}
	
	// 创建默认处理器
	server.handler = NewSimpleEchoHandler(config, logger, metricsCollector)
	
	return server
}

// Start 启动TCP服务端
func (ts *TCPServer) Start(ctx context.Context) error {
	if ts.IsRunning() {
		return fmt.Errorf("TCP server is already running")
	}
	
	// 创建监听器
	listener, err := net.Listen("tcp", ts.config.GetAddress())
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", ts.config.GetAddress(), err)
	}
	
	ts.listener = listener
	
	ts.LogInfo("Starting TCP server", map[string]interface{}{
		"address":         ts.config.GetAddress(),
		"max_connections": ts.config.MaxConnections,
		"echo_mode":       ts.config.EchoMode,
	})
	
	// 启动接受连接的协程
	ts.wg.Add(1)
	go ts.acceptConnections(ctx)
	
	ts.SetRunning(true)
	return nil
}

// Stop 停止TCP服务端
func (ts *TCPServer) Stop(ctx context.Context) error {
	if !ts.IsRunning() {
		return fmt.Errorf("TCP server is not running")
	}
	
	ts.LogInfo("Stopping TCP server", map[string]interface{}{
		"address": ts.config.GetAddress(),
	})
	
	var stopErr error
	ts.stopOnce.Do(func() {
		// 关闭监听器
		if ts.listener != nil {
			if err := ts.listener.Close(); err != nil {
				stopErr = err
			}
		}
		
		// 关闭所有连接
		ts.connectionManager.CloseAll()
		
		// 等待所有协程完成
		done := make(chan struct{})
		go func() {
			ts.wg.Wait()
			close(done)
		}()
		
		// 等待或超时
		select {
		case <-done:
			// 正常完成
		case <-time.After(10 * time.Second):
			ts.LogError("Timeout waiting for connections to close", nil)
		}
		
		ts.SetRunning(false)
	})
	
	if stopErr != nil {
		return stopErr
	}
	
	return ts.Shutdown(ctx)
}

// acceptConnections 接受连接
func (ts *TCPServer) acceptConnections(ctx context.Context) {
	defer ts.wg.Done()
	
	for {
		// 检查上下文
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		// 接受新连接
		conn, err := ts.listener.Accept()
		if err != nil {
			// 检查是否是因为监听器关闭
			if ne, ok := err.(*net.OpError); ok && ne.Op == "accept" {
				return
			}
			
			ts.LogError("Failed to accept connection", err, map[string]interface{}{
				"address": ts.config.GetAddress(),
			})
			continue
		}
		
		// 检查连接数限制
		if ts.connectionManager.GetConnectionCount() >= ts.config.MaxConnections {
			ts.LogError("Maximum connections reached, rejecting connection", nil, map[string]interface{}{
				"max_connections":    ts.config.MaxConnections,
				"current_connections": ts.connectionManager.GetConnectionCount(),
				"remote_addr":        conn.RemoteAddr().String(),
			})
			conn.Close()
			continue
		}
		
		// 启动连接处理协程
		ts.wg.Add(1)
		go ts.handleConnection(ctx, conn)
	}
}

// handleConnection 处理单个连接
func (ts *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer ts.wg.Done()
	defer conn.Close()
	
	remoteAddr := conn.RemoteAddr().String()
	
	// 创建连接对象
	connection := NewConnection(conn, ts.config, ts.GetLogger(), ts.GetMetricsCollector())
	
	// 添加到连接管理器
	if err := ts.connectionManager.AddConnection(connection); err != nil {
		ts.LogError("Failed to add connection", err, map[string]interface{}{
			"remote_addr": remoteAddr,
		})
		return
	}
	
	// 记录连接
	ts.IncrementActiveConnections()
	defer func() {
		ts.DecrementActiveConnections()
		ts.connectionManager.RemoveConnection(connection.ID)
	}()
	
	if ts.config.LogConnections {
		ts.LogInfo("New TCP connection", map[string]interface{}{
			"connection_id": connection.ID,
			"remote_addr":   remoteAddr,
			"local_addr":    conn.LocalAddr().String(),
		})
	}
	
	// 使用处理器处理连接
	if err := ts.handler.HandleConnection(ctx, conn); err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			ts.LogError("Connection handling error", err, map[string]interface{}{
				"connection_id": connection.ID,
				"remote_addr":   remoteAddr,
			})
		}
	}
	
	if ts.config.LogConnections {
		info := connection.GetInfo()
		ts.LogInfo("TCP connection closed", map[string]interface{}{
			"connection_id":   connection.ID,
			"remote_addr":     remoteAddr,
			"duration":        time.Since(connection.ConnectedAt).String(),
			"bytes_sent":      info.BytesSent,
			"bytes_recv":      info.BytesRecv,
			"messages_sent":   info.MessagesSent,
			"messages_recv":   info.MessagesRecv,
		})
	}
}

// GetMetrics 获取TCP服务端指标
func (ts *TCPServer) GetMetrics() map[string]interface{} {
	baseMetrics := ts.BaseServer.GetMetrics()
	
	// 添加TCP特定指标
	baseMetrics["max_connections"] = ts.config.MaxConnections
	baseMetrics["current_connections"] = ts.connectionManager.GetConnectionCount()
	baseMetrics["echo_mode"] = ts.config.EchoMode
	
	// 连接统计
	connectionStats := ts.GetConnectionStats()
	for k, v := range connectionStats {
		baseMetrics[k] = v
	}
	
	// 连接信息
	connections := ts.connectionManager.GetAllConnections()
	if len(connections) > 0 {
		baseMetrics["connections"] = connections
	}
	
	return baseMetrics
}

// GetConnectionManager 获取连接管理器
func (ts *TCPServer) GetConnectionManager() *ConnectionManager {
	return ts.connectionManager
}

// SetHandler 设置连接处理器
func (ts *TCPServer) SetHandler(handler interfaces.ConnectionHandler) {
	ts.handler = handler
}

// GetHandler 获取连接处理器
func (ts *TCPServer) GetHandler() interfaces.ConnectionHandler {
	return ts.handler
}

// SendMessageToConnection 向指定连接发送消息
func (ts *TCPServer) SendMessageToConnection(connectionID string, data []byte) error {
	connection, exists := ts.connectionManager.GetConnection(connectionID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connectionID)
	}
	
	return connection.SendMessage(data)
}

// BroadcastMessage 广播消息到所有连接
func (ts *TCPServer) BroadcastMessage(data []byte) error {
	connections := ts.connectionManager.GetAllConnections()
	
	var errors []error
	for _, connInfo := range connections {
		if connection, exists := ts.connectionManager.GetConnection(connInfo.ID); exists {
			if err := connection.SendMessage(data); err != nil {
				errors = append(errors, fmt.Errorf("failed to send to %s: %w", connInfo.ID, err))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d connections: %v", len(errors), errors)
	}
	
	return nil
}

// GetTCPConfig 获取TCP配置
func (ts *TCPServer) GetTCPConfig() *TCPServerConfig {
	return ts.config
}