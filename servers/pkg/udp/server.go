package udp

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// UDPServer UDP服务端实现
type UDPServer struct {
	*common.BaseServer
	
	config     *UDPServerConfig
	conn       *net.UDPConn
	handler    PacketHandler
	stats      *UDPStats
	
	// 控制
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// NewUDPServer 创建UDP服务端
func NewUDPServer(config *UDPServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *UDPServer {
	baseServer := common.NewBaseServer("udp", config, logger, metricsCollector)
	
	server := &UDPServer{
		BaseServer: baseServer,
		config:     config,
		handler:    NewEchoPacketHandler(config, logger),
		stats: &UDPStats{
			StartTime: time.Now(),
		},
	}
	
	return server
}

// Start 启动UDP服务端
func (us *UDPServer) Start(ctx context.Context) error {
	if us.IsRunning() {
		return fmt.Errorf("UDP server is already running")
	}
	
	// 解析地址
	addr, err := net.ResolveUDPAddr("udp", us.config.GetAddress())
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", us.config.GetAddress(), err)
	}
	
	// 创建UDP连接
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP %s: %w", us.config.GetAddress(), err)
	}
	
	us.conn = conn
	
	us.LogInfo("Starting UDP server", map[string]interface{}{
		"address":           us.config.GetAddress(),
		"echo_mode":         us.config.EchoMode,
		"packet_loss_rate":  us.config.PacketLossRate,
		"enable_multicast":  us.config.EnableMulticast,
		"enable_broadcast":  us.config.EnableBroadcast,
	})
	
	// 配置多播
	if us.config.EnableMulticast {
		if err := us.setupMulticast(); err != nil {
			us.LogError("Failed to setup multicast", err)
			// 继续运行，但记录警告
		}
	}
	
	// 配置广播
	if us.config.EnableBroadcast {
		if err := us.setupBroadcast(); err != nil {
			us.LogError("Failed to setup broadcast", err)
			// 继续运行，但记录警告
		}
	}
	
	// 启动数据包处理协程
	us.wg.Add(1)
	go us.handlePackets(ctx)
	
	us.SetRunning(true)
	return nil
}

// Stop 停止UDP服务端
func (us *UDPServer) Stop(ctx context.Context) error {
	if !us.IsRunning() {
		return fmt.Errorf("UDP server is not running")
	}
	
	us.LogInfo("Stopping UDP server", map[string]interface{}{
		"address": us.config.GetAddress(),
	})
	
	var stopErr error
	us.stopOnce.Do(func() {
		// 关闭UDP连接
		if us.conn != nil {
			if err := us.conn.Close(); err != nil {
				stopErr = err
			}
		}
		
		// 等待处理协程完成
		done := make(chan struct{})
		go func() {
			us.wg.Wait()
			close(done)
		}()
		
		// 等待或超时
		select {
		case <-done:
			// 正常完成
		case <-time.After(5 * time.Second):
			us.LogError("Timeout waiting for packet handler to stop", nil)
		}
		
		us.SetRunning(false)
	})
	
	if stopErr != nil {
		return stopErr
	}
	
	return us.Shutdown(ctx)
}

// handlePackets 处理数据包
func (us *UDPServer) handlePackets(ctx context.Context) {
	defer us.wg.Done()
	
	buffer := make([]byte, us.config.BufferSize)
	
	for {
		// 检查上下文
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		// 设置读取超时
		if err := us.conn.SetReadDeadline(time.Now().Add(us.config.ReadTimeout)); err != nil {
			us.LogError("Failed to set read deadline", err)
			continue
		}
		
		// 读取数据包
		n, remoteAddr, err := us.conn.ReadFromUDP(buffer)
		if err != nil {
			// 检查是否是因为连接关闭
			if ne, ok := err.(*net.OpError); ok && ne.Op == "read" {
				return
			}
			
			us.LogError("Failed to read UDP packet", err)
			atomic.AddInt64(&us.stats.ErrorCount, 1)
			continue
		}
		
		if n == 0 {
			continue
		}
		
		// 更新统计
		atomic.AddInt64(&us.stats.PacketsReceived, 1)
		atomic.AddInt64(&us.stats.BytesReceived, int64(n))
		
		// 复制数据包内容
		packet := make([]byte, n)
		copy(packet, buffer[:n])
		
		// 记录指标
		if us.GetMetricsCollector() != nil {
			us.GetMetricsCollector().RecordRequest("udp", "packet_recv", 0, true)
		}
		
		// 模拟丢包
		if us.shouldDropPacket() {
			atomic.AddInt64(&us.stats.PacketsDropped, 1)
			
			if us.config.LogPackets {
				us.LogDebug("Packet dropped (simulated loss)", map[string]interface{}{
					"remote_addr": remoteAddr.String(),
					"size":        n,
				})
			}
			continue
		}
		
		// 处理数据包
		go us.processPacket(packet, remoteAddr)
	}
}

// processPacket 处理单个数据包
func (us *UDPServer) processPacket(packet []byte, remoteAddr *net.UDPAddr) {
	start := time.Now()
	
	// 使用处理器处理数据包
	response, err := us.handler.HandlePacket(packet, remoteAddr.String())
	if err != nil {
		us.LogError("Failed to handle packet", err, map[string]interface{}{
			"remote_addr": remoteAddr.String(),
			"size":        len(packet),
		})
		
		atomic.AddInt64(&us.stats.ErrorCount, 1)
		
		if us.GetMetricsCollector() != nil {
			us.GetMetricsCollector().RecordRequest("udp", "packet_handle", time.Since(start), false)
		}
		return
	}
	
	// 发送响应（如果有）
	if response != nil && len(response) > 0 {
		if err := us.sendResponse(response, remoteAddr); err != nil {
			us.LogError("Failed to send response", err, map[string]interface{}{
				"remote_addr": remoteAddr.String(),
				"size":        len(response),
			})
			atomic.AddInt64(&us.stats.ErrorCount, 1)
		}
	}
	
	// 记录指标
	if us.GetMetricsCollector() != nil {
		us.GetMetricsCollector().RecordRequest("udp", "packet_handle", time.Since(start), true)
	}
}

// sendResponse 发送响应
func (us *UDPServer) sendResponse(data []byte, remoteAddr *net.UDPAddr) error {
	// 应用响应延迟
	if us.config.ResponseDelay > 0 {
		time.Sleep(us.config.ResponseDelay)
	}
	
	// 设置写入超时
	if err := us.conn.SetWriteDeadline(time.Now().Add(us.config.WriteTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}
	
	// 发送数据
	n, err := us.conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write UDP packet: %w", err)
	}
	
	// 更新统计
	atomic.AddInt64(&us.stats.PacketsSent, 1)
	atomic.AddInt64(&us.stats.BytesSent, int64(n))
	
	// 记录发送的数据包
	if us.config.LogPackets {
		us.LogDebug("UDP packet sent", map[string]interface{}{
			"remote_addr": remoteAddr.String(),
			"size":        n,
		})
	}
	
	// 记录指标
	if us.GetMetricsCollector() != nil {
		us.GetMetricsCollector().RecordRequest("udp", "packet_sent", 0, true)
	}
	
	return nil
}

// shouldDropPacket 检查是否应该丢弃数据包
func (us *UDPServer) shouldDropPacket() bool {
	if us.config.PacketLossRate <= 0 {
		return false
	}
	
	return rand.Float64() < us.config.PacketLossRate
}

// setupMulticast 设置多播
func (us *UDPServer) setupMulticast() error {
	// 解析多播地址
	multicastAddr, err := net.ResolveUDPAddr("udp", us.config.MulticastGroup+":"+fmt.Sprintf("%d", us.config.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve multicast address: %w", err)
	}
	
	us.LogInfo("Multicast enabled", map[string]interface{}{
		"group": us.config.MulticastGroup,
		"ttl":   us.config.MulticastTTL,
	})
	
	// 这里可以添加具体的多播配置逻辑
	_ = multicastAddr
	
	return nil
}

// setupBroadcast 设置广播
func (us *UDPServer) setupBroadcast() error {
	us.LogInfo("Broadcast enabled")
	
	// 这里可以添加具体的广播配置逻辑
	
	return nil
}

// GetMetrics 获取UDP服务端指标
func (us *UDPServer) GetMetrics() map[string]interface{} {
	baseMetrics := us.BaseServer.GetMetrics()
	
	// 添加UDP特定指标
	baseMetrics["echo_mode"] = us.config.EchoMode
	baseMetrics["packet_loss_rate"] = us.config.PacketLossRate
	baseMetrics["enable_multicast"] = us.config.EnableMulticast
	baseMetrics["enable_broadcast"] = us.config.EnableBroadcast
	baseMetrics["max_packet_size"] = us.config.MaxPacketSize
	
	// 统计信息
	baseMetrics["packets_received"] = atomic.LoadInt64(&us.stats.PacketsReceived)
	baseMetrics["packets_sent"] = atomic.LoadInt64(&us.stats.PacketsSent)
	baseMetrics["packets_dropped"] = atomic.LoadInt64(&us.stats.PacketsDropped)
	baseMetrics["bytes_received"] = atomic.LoadInt64(&us.stats.BytesReceived)
	baseMetrics["bytes_sent"] = atomic.LoadInt64(&us.stats.BytesSent)
	baseMetrics["error_count"] = atomic.LoadInt64(&us.stats.ErrorCount)
	
	return baseMetrics
}

// SendPacket 发送数据包到指定地址
func (us *UDPServer) SendPacket(data []byte, remoteAddr string) error {
	if !us.IsRunning() {
		return fmt.Errorf("UDP server is not running")
	}
	
	addr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve address %s: %w", remoteAddr, err)
	}
	
	return us.sendResponse(data, addr)
}

// GetHandler 获取数据包处理器
func (us *UDPServer) GetHandler() PacketHandler {
	return us.handler
}

// SetHandler 设置数据包处理器
func (us *UDPServer) SetHandler(handler PacketHandler) {
	us.handler = handler
}

// GetUDPConfig 获取UDP配置
func (us *UDPServer) GetUDPConfig() *UDPServerConfig {
	return us.config
}

// GetStats 获取统计信息
func (us *UDPServer) GetStats() UDPStats {
	return UDPStats{
		PacketsReceived: atomic.LoadInt64(&us.stats.PacketsReceived),
		PacketsSent:     atomic.LoadInt64(&us.stats.PacketsSent),
		PacketsDropped:  atomic.LoadInt64(&us.stats.PacketsDropped),
		BytesReceived:   atomic.LoadInt64(&us.stats.BytesReceived),
		BytesSent:       atomic.LoadInt64(&us.stats.BytesSent),
		ErrorCount:      atomic.LoadInt64(&us.stats.ErrorCount),
		StartTime:       us.stats.StartTime,
	}
}