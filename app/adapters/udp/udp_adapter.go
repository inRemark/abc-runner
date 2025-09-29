package udp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/adapters/udp/config"
	"abc-runner/app/core/interfaces"
)

// UDPAdapter UDP协议适配器
type UDPAdapter struct {
	config           *config.UDPConfig
	conn             net.Conn
	packetConn       net.PacketConn
	metricsCollector interfaces.DefaultMetricsCollector
	mu               sync.RWMutex
	isConnected      bool

	// 统计信息
	sentPackets      int64
	receivedPackets  int64
	lostPackets      int64
	duplicatePackets int64
}

// NewUDPAdapter 创建UDP适配器
func NewUDPAdapter(metricsCollector interfaces.DefaultMetricsCollector) *UDPAdapter {
	return &UDPAdapter{
		metricsCollector: metricsCollector,
		isConnected:      false,
	}
}

// Connect 初始化连接
func (u *UDPAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// 类型断言配置
	udpConfig, ok := cfg.(*config.UDPConfig)
	if !ok {
		return fmt.Errorf("invalid config type for UDP adapter")
	}

	u.config = udpConfig

	// 验证配置
	if err := udpConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 根据模式创建不同的连接
	switch udpConfig.UDPSpecific.PacketMode {
	case "unicast":
		if err := u.setupUnicastConnection(); err != nil {
			return fmt.Errorf("failed to setup unicast connection: %w", err)
		}
	case "broadcast":
		if err := u.setupBroadcastConnection(); err != nil {
			return fmt.Errorf("failed to setup broadcast connection: %w", err)
		}
	case "multicast":
		if err := u.setupMulticastConnection(); err != nil {
			return fmt.Errorf("failed to setup multicast connection: %w", err)
		}
	default:
		return fmt.Errorf("unsupported packet mode: %s", udpConfig.UDPSpecific.PacketMode)
	}

	u.isConnected = true
	return nil
}

// setupUnicastConnection 设置单播连接
func (u *UDPAdapter) setupUnicastConnection() error {
	address := fmt.Sprintf("%s:%d", u.config.Connection.Address, u.config.Connection.Port)

	// 创建UDP连接
	conn, err := net.DialTimeout("udp", address, u.config.Connection.Timeout)
	if err != nil {
		return fmt.Errorf("failed to dial UDP address %s: %w", address, err)
	}

	u.conn = conn
	return nil
}

// setupBroadcastConnection 设置广播连接
func (u *UDPAdapter) setupBroadcastConnection() error {
	// 创建UDP PacketConn用于广播
	packetConn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return fmt.Errorf("failed to listen for broadcast: %w", err)
	}

	u.packetConn = packetConn
	return nil
}

// setupMulticastConnection 设置组播连接
func (u *UDPAdapter) setupMulticastConnection() error {
	groupAddr, err := net.ResolveUDPAddr("udp",
		fmt.Sprintf("%s:%d", u.config.UDPSpecific.MulticastGroup, u.config.Connection.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	// 创建组播连接
	conn, err := net.ListenMulticastUDP("udp", nil, groupAddr)
	if err != nil {
		return fmt.Errorf("failed to listen multicast: %w", err)
	}

	u.packetConn = conn
	return nil
}

// Execute 执行操作
func (u *UDPAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	result := &interfaces.OperationResult{
		Success:  false,
		Duration: 0,
		IsRead:   false,
		Error:    nil,
		Value:    nil,
		Metadata: make(map[string]interface{}),
	}

	// 检查连接状态
	if !u.isConnected {
		result.Error = fmt.Errorf("adapter not connected")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// 根据操作类型执行不同的操作
	switch operation.Type {
	case "packet_send":
		result, err := u.executePacketSend(ctx, operation)
		result.Duration = time.Since(startTime)
		if u.metricsCollector != nil {
			u.metricsCollector.Record(result)
		}
		return result, err
	case "packet_receive":
		result, err := u.executePacketReceive(ctx, operation)
		result.Duration = time.Since(startTime)
		if u.metricsCollector != nil {
			u.metricsCollector.Record(result)
		}
		return result, err
	case "echo_udp":
		result, err := u.executeEchoUDP(ctx, operation)
		result.Duration = time.Since(startTime)
		if u.metricsCollector != nil {
			u.metricsCollector.Record(result)
		}
		return result, err
	case "multicast":
		result, err := u.executeMulticast(ctx, operation)
		result.Duration = time.Since(startTime)
		if u.metricsCollector != nil {
			u.metricsCollector.Record(result)
		}
		return result, err
	default:
		result.Error = fmt.Errorf("unsupported operation type: %s", operation.Type)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
}

// executePacketSend 执行数据包发送
func (u *UDPAdapter) executePacketSend(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   false,
		Metadata: make(map[string]interface{}),
	}

	// 构造测试数据
	testData := u.buildTestData(operation)

	// 设置超时
	deadline := time.Now().Add(u.config.Connection.Timeout)

	var n int
	var err error

	switch u.config.UDPSpecific.PacketMode {
	case "unicast":
		if u.conn != nil {
			u.conn.SetWriteDeadline(deadline)
			n, err = u.conn.Write(testData)
		} else {
			err = fmt.Errorf("unicast connection not available")
		}
	case "broadcast":
		if u.packetConn != nil {
			broadcastAddr := &net.UDPAddr{
				IP:   net.IPv4bcast,
				Port: u.config.Connection.Port,
			}
			n, err = u.packetConn.WriteTo(testData, broadcastAddr)
		} else {
			err = fmt.Errorf("broadcast connection not available")
		}
	case "multicast":
		if u.packetConn != nil {
			multicastAddr, _ := net.ResolveUDPAddr("udp",
				fmt.Sprintf("%s:%d", u.config.UDPSpecific.MulticastGroup, u.config.Connection.Port))
			n, err = u.packetConn.WriteTo(testData, multicastAddr)
		} else {
			err = fmt.Errorf("multicast connection not available")
		}
	}

	if err != nil {
		result.Error = fmt.Errorf("failed to send packet: %w", err)
		return result, result.Error
	}

	atomic.AddInt64(&u.sentPackets, 1)

	result.Success = true
	result.Value = n
	result.Metadata["sent_bytes"] = n
	result.Metadata["packet_mode"] = u.config.UDPSpecific.PacketMode
	result.Metadata["packet_id"] = atomic.LoadInt64(&u.sentPackets)

	return result, nil
}

// executePacketReceive 执行数据包接收
func (u *UDPAdapter) executePacketReceive(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true,
		Metadata: make(map[string]interface{}),
	}

	buffer := make([]byte, u.config.BenchMark.DataSize*2)
	deadline := time.Now().Add(u.config.Connection.Timeout)

	var n int
	var addr net.Addr
	var err error

	if u.conn != nil {
		// 单播接收
		u.conn.SetReadDeadline(deadline)
		n, err = u.conn.Read(buffer)
	} else if u.packetConn != nil {
		// 广播/组播接收
		u.packetConn.SetReadDeadline(deadline)
		n, addr, err = u.packetConn.ReadFrom(buffer)
	} else {
		err = fmt.Errorf("no connection available for receiving")
	}

	if err != nil {
		result.Error = fmt.Errorf("failed to receive packet: %w", err)
		return result, result.Error
	}

	atomic.AddInt64(&u.receivedPackets, 1)

	result.Success = true
	result.Value = buffer[:n]
	result.Metadata["received_bytes"] = n
	result.Metadata["packet_mode"] = u.config.UDPSpecific.PacketMode
	result.Metadata["packet_id"] = atomic.LoadInt64(&u.receivedPackets)

	if addr != nil {
		result.Metadata["sender_address"] = addr.String()
	}

	return result, nil
}

// executeEchoUDP 执行UDP回显测试
func (u *UDPAdapter) executeEchoUDP(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true, // 回显测试主要是验证响应
		Metadata: make(map[string]interface{}),
	}

	// 构造测试数据
	testData := u.buildTestData(operation)

	// 发送数据
	sendResult, err := u.executePacketSend(ctx, operation)
	if err != nil {
		result.Error = fmt.Errorf("failed to send echo packet: %w", err)
		return result, result.Error
	}

	// 等待响应
	receiveOp := interfaces.Operation{
		Type:   "packet_receive",
		Key:    operation.Key,
		Params: operation.Params,
	}

	receiveResult, err := u.executePacketReceive(ctx, receiveOp)
	if err != nil {
		result.Error = fmt.Errorf("failed to receive echo response: %w", err)
		return result, result.Error
	}

	// 验证响应数据
	receivedData := receiveResult.Value.([]byte)
	dataMatch := len(receivedData) == len(testData)
	if dataMatch {
		for i, b := range testData {
			if i >= len(receivedData) || receivedData[i] != b {
				dataMatch = false
				break
			}
		}
	}

	result.Success = true
	result.Value = receivedData
	result.Metadata["sent_bytes"] = sendResult.Value
	result.Metadata["received_bytes"] = len(receivedData)
	result.Metadata["data_match"] = dataMatch
	result.Metadata["echo_verified"] = dataMatch

	return result, nil
}

// executeMulticast 执行组播测试
func (u *UDPAdapter) executeMulticast(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if u.config.UDPSpecific.PacketMode != "multicast" {
		return nil, fmt.Errorf("multicast operation requires multicast mode")
	}

	// 直接使用packet_send的逻辑，但标记为组播
	result, err := u.executePacketSend(ctx, operation)
	if err != nil {
		return result, err
	}

	result.Metadata["multicast_group"] = u.config.UDPSpecific.MulticastGroup
	result.Metadata["ttl"] = u.config.UDPSpecific.TTL

	return result, nil
}

// buildTestData 构造测试数据
func (u *UDPAdapter) buildTestData(operation interfaces.Operation) []byte {
	// 如果操作中包含自定义数据，使用它
	if operation.Value != nil {
		if data, ok := operation.Value.([]byte); ok {
			return data
		}
		if str, ok := operation.Value.(string); ok {
			return []byte(str)
		}
	}

	// 否则生成默认大小的测试数据
	data := make([]byte, u.config.BenchMark.DataSize)

	// 添加包标识和序列号
	packetId := atomic.LoadInt64(&u.sentPackets)
	for i := range data {
		if i < 8 {
			// 前8字节作为包ID
			data[i] = byte((packetId >> (8 * (7 - i))) & 0xFF)
		} else {
			// 其余字节为测试数据
			data[i] = byte(i % 256)
		}
	}

	return data
}

// Close 关闭连接
func (u *UDPAdapter) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	var errs []error

	if u.conn != nil {
		if err := u.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close conn: %w", err))
		}
		u.conn = nil
	}

	if u.packetConn != nil {
		if err := u.packetConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close packetConn: %w", err))
		}
		u.packetConn = nil
	}

	u.isConnected = false

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// GetProtocolMetrics 获取协议特定指标
func (u *UDPAdapter) GetProtocolMetrics() map[string]interface{} {
	sentPackets := atomic.LoadInt64(&u.sentPackets)
	receivedPackets := atomic.LoadInt64(&u.receivedPackets)

	lossRate := float64(0)
	if sentPackets > 0 {
		lossRate = float64(sentPackets-receivedPackets) / float64(sentPackets) * 100
	}

	metrics := map[string]interface{}{
		"sent_packets":      sentPackets,
		"received_packets":  receivedPackets,
		"lost_packets":      sentPackets - receivedPackets,
		"duplicate_packets": atomic.LoadInt64(&u.duplicatePackets),
		"packet_loss_rate":  lossRate,
	}

	// 只有在配置存在时才添加配置相关指标
	if u.config != nil {
		metrics["packet_mode"] = u.config.UDPSpecific.PacketMode
		metrics["ttl"] = u.config.UDPSpecific.TTL
		metrics["buffer_size"] = u.config.Connection.BufferSize
		metrics["multicast_group"] = u.config.UDPSpecific.MulticastGroup
	} else {
		metrics["packet_mode"] = "unknown"
		metrics["ttl"] = 0
		metrics["buffer_size"] = 0
		metrics["multicast_group"] = ""
	}

	return metrics
}

// HealthCheck 健康检查
func (u *UDPAdapter) HealthCheck(ctx context.Context) error {
	if !u.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	// 对于UDP，健康检查主要是验证连接对象存在
	if u.conn == nil && u.packetConn == nil {
		return fmt.Errorf("no UDP connection available")
	}

	return nil
}

// GetProtocolName 获取协议名称
func (u *UDPAdapter) GetProtocolName() string {
	return "udp"
}

// GetMetricsCollector 获取指标收集器
func (u *UDPAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return u.metricsCollector
}
