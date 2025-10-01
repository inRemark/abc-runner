package udp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"abc-runner/app/adapters/udp/config"
	"abc-runner/app/adapters/udp/operations"
	"abc-runner/app/core/interfaces"
)

// UDPAdapter UDP协议适配器 - 遵循统一架构模式
// 职责：连接管理、状态维护、健康检查
type UDPAdapter struct {
	config           *config.UDPConfig
	conn             net.Conn
	packetConn       net.PacketConn
	udpOperations    *operations.UDPOperations
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

	// 创建 UDP操作执行器
	u.udpOperations = operations.NewUDPOperations(u.conn, u.packetConn, u.config, u.metricsCollector)

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

// Execute 执行操作 - 使用执行器处理
func (u *UDPAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !u.isConnected {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: 0,
			Error:    fmt.Errorf("adapter not connected"),
		}, fmt.Errorf("adapter not connected")
	}

	// 委托给UDP操作执行器处理
	return u.udpOperations.ExecuteOperation(ctx, operation)
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
