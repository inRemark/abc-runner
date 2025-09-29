package tcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/servers/pkg/interfaces"
)

// Connection TCP连接封装
type Connection struct {
	ID          string
	conn        net.Conn
	RemoteAddr  string
	LocalAddr   string
	ConnectedAt time.Time

	// 统计信息
	BytesSent    int64
	BytesRecv    int64
	MessagesSent int64
	MessagesRecv int64

	// 控制
	closed    bool
	closeMu   sync.RWMutex
	closeChan chan struct{}

	// 配置
	config  *TCPServerConfig
	logger  interfaces.Logger
	metrics interfaces.MetricsCollector
}

// NewConnection 创建新连接
func NewConnection(conn net.Conn, config *TCPServerConfig, logger interfaces.Logger, metrics interfaces.MetricsCollector) *Connection {
	return &Connection{
		ID:          GenerateConnectionID(),
		conn:        conn,
		RemoteAddr:  conn.RemoteAddr().String(),
		LocalAddr:   conn.LocalAddr().String(),
		ConnectedAt: time.Now(),
		closeChan:   make(chan struct{}),
		config:      config,
		logger:      logger,
		metrics:     metrics,
	}
}

// Handle 处理连接
func (c *Connection) Handle(ctx context.Context) error {
	defer c.Close()

	// 设置连接参数
	if tcpConn, ok := c.conn.(*net.TCPConn); ok {
		if err := tcpConn.SetKeepAlive(c.config.KeepAlive); err != nil {
			c.logger.Warn("Failed to set keep alive", map[string]interface{}{
				"connection_id": c.ID,
				"error":         err.Error(),
			})
		}

		if err := tcpConn.SetNoDelay(c.config.NoDelay); err != nil {
			c.logger.Warn("Failed to set no delay", map[string]interface{}{
				"connection_id": c.ID,
				"error":         err.Error(),
			})
		}
	}

	// 设置初始超时
	if err := c.conn.SetDeadline(time.Now().Add(c.config.ConnectionTimeout)); err != nil {
		return fmt.Errorf("failed to set connection deadline: %w", err)
	}

	// 创建读取器
	reader := bufio.NewReaderSize(c.conn, c.config.ReadBufferSize)

	// 处理消息循环
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.closeChan:
			return nil
		default:
			// 处理单个消息
			if err := c.handleMessage(reader); err != nil {
				if err == io.EOF || c.isClosed() {
					return nil
				}
				return fmt.Errorf("message handling error: %w", err)
			}
		}
	}
}

// handleMessage 处理单个消息
func (c *Connection) handleMessage(reader *bufio.Reader) error {
	// 设置读取超时
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// 读取消息长度（前4字节）
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(reader, lengthBytes); err != nil {
		return err
	}

	// 解析消息长度
	messageLength := int(lengthBytes[0])<<24 | int(lengthBytes[1])<<16 | int(lengthBytes[2])<<8 | int(lengthBytes[3])

	// 验证消息长度
	if messageLength <= 0 || messageLength > c.config.MaxMessageSize {
		return fmt.Errorf("invalid message length: %d", messageLength)
	}

	// 读取消息内容
	messageData := make([]byte, messageLength)
	if _, err := io.ReadFull(reader, messageData); err != nil {
		return err
	}

	// 更新统计
	atomic.AddInt64(&c.BytesRecv, int64(4+messageLength))
	atomic.AddInt64(&c.MessagesRecv, 1)

	// 记录消息
	if c.config.LogMessages && c.logger != nil {
		c.logger.Debug("Message received", map[string]interface{}{
			"connection_id": c.ID,
			"size":          messageLength,
			"data":          string(messageData),
		})
	}

	// 记录指标
	if c.metrics != nil {
		c.metrics.RecordRequest("tcp", "message_recv", time.Since(time.Now()), true)
	}

	// 处理消息（回显模式）
	if c.config.EchoMode {
		return c.sendResponse(messageData)
	}

	return nil
}

// sendResponse 发送响应
func (c *Connection) sendResponse(data []byte) error {
	// 应用响应延迟
	if c.config.ResponseDelay > 0 {
		time.Sleep(c.config.ResponseDelay)
	}

	// 设置写入超时
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// 构建响应（长度+数据）
	messageLength := len(data)
	lengthBytes := []byte{
		byte(messageLength >> 24),
		byte(messageLength >> 16),
		byte(messageLength >> 8),
		byte(messageLength),
	}

	// 发送长度
	if _, err := c.conn.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// 发送数据
	if _, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	// 更新统计
	atomic.AddInt64(&c.BytesSent, int64(4+messageLength))
	atomic.AddInt64(&c.MessagesSent, 1)

	// 记录消息
	if c.config.LogMessages && c.logger != nil {
		c.logger.Debug("Message sent", map[string]interface{}{
			"connection_id": c.ID,
			"size":          messageLength,
		})
	}

	// 记录指标
	if c.metrics != nil {
		c.metrics.RecordRequest("tcp", "message_sent", time.Since(time.Now()), true)
	}

	return nil
}

// SendMessage 发送自定义消息
func (c *Connection) SendMessage(data []byte) error {
	if c.isClosed() {
		return fmt.Errorf("connection is closed")
	}

	return c.sendResponse(data)
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.closeChan)

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

// isClosed 检查连接是否已关闭
func (c *Connection) isClosed() bool {
	c.closeMu.RLock()
	defer c.closeMu.RUnlock()
	return c.closed
}

// GetInfo 获取连接信息
func (c *Connection) GetInfo() ConnectionInfo {
	return ConnectionInfo{
		ID:           c.ID,
		RemoteAddr:   c.RemoteAddr,
		LocalAddr:    c.LocalAddr,
		ConnectedAt:  c.ConnectedAt,
		BytesSent:    atomic.LoadInt64(&c.BytesSent),
		BytesRecv:    atomic.LoadInt64(&c.BytesRecv),
		MessagesSent: atomic.LoadInt64(&c.MessagesSent),
		MessagesRecv: atomic.LoadInt64(&c.MessagesRecv),
	}
}

// SimpleEchoHandler 简单回显处理器
type SimpleEchoHandler struct {
	config  *TCPServerConfig
	logger  interfaces.Logger
	metrics interfaces.MetricsCollector
}

// NewSimpleEchoHandler 创建简单回显处理器
func NewSimpleEchoHandler(config *TCPServerConfig, logger interfaces.Logger, metrics interfaces.MetricsCollector) *SimpleEchoHandler {
	return &SimpleEchoHandler{
		config:  config,
		logger:  logger,
		metrics: metrics,
	}
}

// HandleConnection 处理连接（简单实现）
func (h *SimpleEchoHandler) HandleConnection(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	connectionID := GenerateConnectionID()
	remoteAddr := conn.RemoteAddr().String()

	if h.config.LogConnections && h.logger != nil {
		h.logger.Info("New TCP connection", map[string]interface{}{
			"connection_id": connectionID,
			"remote_addr":   remoteAddr,
		})
	}

	// 设置连接参数
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(h.config.KeepAlive)
		tcpConn.SetNoDelay(h.config.NoDelay)
	}

	buffer := make([]byte, h.config.BufferSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))

			// 读取数据
			n, err := conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					goto ConnectionClosed
				}
				return fmt.Errorf("read error: %w", err)
			}

			if n == 0 {
				continue
			}

			data := buffer[:n]

			// 记录指标
			if h.metrics != nil {
				h.metrics.RecordRequest("tcp", "echo", 0, true)
			}

			// 应用延迟
			if h.config.ResponseDelay > 0 {
				time.Sleep(h.config.ResponseDelay)
			}

			// 回显数据
			if h.config.EchoMode {
				conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))

				if _, err := conn.Write(data); err != nil {
					return fmt.Errorf("write error: %w", err)
				}
			}
		}
	}

ConnectionClosed:
	if h.config.LogConnections && h.logger != nil {
		h.logger.Info("TCP connection closed", map[string]interface{}{
			"connection_id": connectionID,
			"remote_addr":   remoteAddr,
		})
	}

	return nil
}

// GetProtocol 获取协议名称
func (h *SimpleEchoHandler) GetProtocol() string {
	return "tcp"
}
