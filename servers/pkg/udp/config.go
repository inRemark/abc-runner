package udp

import (
	"fmt"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// UDPServerConfig UDP服务端配置
type UDPServerConfig struct {
	*common.BaseConfig `yaml:",inline"`
	
	// UDP特定配置
	BufferSize      int           `yaml:"buffer_size" json:"buffer_size"`
	MaxPacketSize   int           `yaml:"max_packet_size" json:"max_packet_size"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	
	// 行为配置
	EchoMode        bool    `yaml:"echo_mode" json:"echo_mode"`
	ResponseDelay   time.Duration `yaml:"response_delay" json:"response_delay"`
	PacketLossRate  float64 `yaml:"packet_loss_rate" json:"packet_loss_rate"`
	
	// 多播配置
	EnableMulticast bool   `yaml:"enable_multicast" json:"enable_multicast"`
	MulticastGroup  string `yaml:"multicast_group" json:"multicast_group"`
	MulticastTTL    int    `yaml:"multicast_ttl" json:"multicast_ttl"`
	
	// 广播配置
	EnableBroadcast bool `yaml:"enable_broadcast" json:"enable_broadcast"`
	
	// 日志配置
	LogPackets bool `yaml:"log_packets" json:"log_packets"`
}

// NewUDPServerConfig 创建UDP服务端配置
func NewUDPServerConfig() *UDPServerConfig {
	return &UDPServerConfig{
		BaseConfig: &common.BaseConfig{
			Protocol: "udp",
			Host:     "localhost",
			Port:     9091,
		},
		BufferSize:      4096,
		MaxPacketSize:   65507, // UDP最大数据包大小
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		EchoMode:        true,
		ResponseDelay:   0,
		PacketLossRate:  0.0,
		EnableMulticast: false,
		MulticastGroup:  "224.0.0.1",
		MulticastTTL:    1,
		EnableBroadcast: false,
		LogPackets:      false,
	}
}

// Validate 验证UDP配置
func (c *UDPServerConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}
	
	if c.BufferSize <= 0 {
		return fmt.Errorf("buffer_size must be positive")
	}
	
	if c.MaxPacketSize <= 0 || c.MaxPacketSize > 65507 {
		return fmt.Errorf("max_packet_size must be between 1 and 65507")
	}
	
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}
	
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}
	
	if c.PacketLossRate < 0.0 || c.PacketLossRate > 1.0 {
		return fmt.Errorf("packet_loss_rate must be between 0.0 and 1.0")
	}
	
	if c.MulticastTTL < 0 || c.MulticastTTL > 255 {
		return fmt.Errorf("multicast_ttl must be between 0 and 255")
	}
	
	return nil
}

// Clone 克隆UDP配置
func (c *UDPServerConfig) Clone() interfaces.ServerConfig {
	clone := *c
	clone.BaseConfig = c.BaseConfig.Clone().(*common.BaseConfig)
	return &clone
}

// PacketInfo 数据包信息
type PacketInfo struct {
	RemoteAddr string    `json:"remote_addr"`
	LocalAddr  string    `json:"local_addr"`
	Size       int       `json:"size"`
	Timestamp  time.Time `json:"timestamp"`
	Direction  string    `json:"direction"` // "in" or "out"
	Data       []byte    `json:"data,omitempty"`
	Dropped    bool      `json:"dropped,omitempty"`
}

// UDPStats UDP统计信息
type UDPStats struct {
	PacketsReceived int64   `json:"packets_received"`
	PacketsSent     int64   `json:"packets_sent"`
	PacketsDropped  int64   `json:"packets_dropped"`
	BytesReceived   int64   `json:"bytes_received"`
	BytesSent       int64   `json:"bytes_sent"`
	ErrorCount      int64   `json:"error_count"`
	StartTime       time.Time `json:"start_time"`
}

// PacketHandler 数据包处理器接口
type PacketHandler interface {
	HandlePacket(packet []byte, remoteAddr string) ([]byte, error)
}

// EchoPacketHandler 回显数据包处理器
type EchoPacketHandler struct {
	config *UDPServerConfig
	logger interfaces.Logger
}

// NewEchoPacketHandler 创建回显数据包处理器
func NewEchoPacketHandler(config *UDPServerConfig, logger interfaces.Logger) *EchoPacketHandler {
	return &EchoPacketHandler{
		config: config,
		logger: logger,
	}
}

// HandlePacket 处理数据包
func (h *EchoPacketHandler) HandlePacket(packet []byte, remoteAddr string) ([]byte, error) {
	// 验证数据包大小
	if len(packet) > h.config.MaxPacketSize {
		return nil, fmt.Errorf("packet too large: %d bytes", len(packet))
	}
	
	// 记录接收的数据包
	if h.config.LogPackets && h.logger != nil {
		h.logger.Debug("UDP packet received", map[string]interface{}{
			"remote_addr": remoteAddr,
			"size":        len(packet),
			"data":        string(packet),
		})
	}
	
	// 回显模式：返回相同的数据
	if h.config.EchoMode {
		return packet, nil
	}
	
	// 非回显模式：返回确认消息
	response := fmt.Sprintf("ACK: received %d bytes from %s", len(packet), remoteAddr)
	return []byte(response), nil
}