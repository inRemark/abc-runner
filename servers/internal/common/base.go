package common

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"abc-runner/servers/pkg/interfaces"
)

// BaseServer 通用服务端基础实现
type BaseServer struct {
	protocol         string
	host             string
	port             int
	config           interfaces.ServerConfig
	logger           interfaces.Logger
	metricsCollector interfaces.MetricsCollector
	healthChecker    interfaces.HealthChecker

	// 运行状态管理
	running    bool
	runningMu  sync.RWMutex
	StartTime  time.Time // 导出字段
	stopChan   chan struct{}
	ctx        context.Context
	cancelFunc context.CancelFunc

	// 连接统计
	activeConnections int64
	totalConnections  int64
	connectionsMu     sync.RWMutex
}

// NewBaseServer 创建基础服务端
func NewBaseServer(protocol string, config interfaces.ServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *BaseServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &BaseServer{
		protocol:         protocol,
		host:             config.GetHost(),
		port:             config.GetPort(),
		config:           config,
		logger:           logger,
		metricsCollector: metricsCollector,
		stopChan:         make(chan struct{}),
		ctx:              ctx,
		cancelFunc:       cancel,
	}
}

// GetProtocol 获取协议名称
func (bs *BaseServer) GetProtocol() string {
	return bs.protocol
}

// GetAddress 获取监听地址
func (bs *BaseServer) GetAddress() string {
	return bs.host
}

// GetPort 获取监听端口
func (bs *BaseServer) GetPort() int {
	return bs.port
}

// IsRunning 检查服务端是否正在运行
func (bs *BaseServer) IsRunning() bool {
	bs.runningMu.RLock()
	defer bs.runningMu.RUnlock()
	return bs.running
}

// GetConfig 获取服务端配置
func (bs *BaseServer) GetConfig() interfaces.ServerConfig {
	return bs.config
}

// SetRunning 设置运行状态
func (bs *BaseServer) SetRunning(running bool) {
	bs.runningMu.Lock()
	defer bs.runningMu.Unlock()

	if running && !bs.running {
		bs.StartTime = time.Now()
		bs.logger.Info(fmt.Sprintf("%s server started", bs.protocol), map[string]interface{}{
			"protocol": bs.protocol,
			"address":  bs.GetFullAddress(),
		})
	} else if !running && bs.running {
		uptime := time.Since(bs.StartTime)
		bs.logger.Info(fmt.Sprintf("%s server stopped", bs.protocol), map[string]interface{}{
			"protocol": bs.protocol,
			"uptime":   uptime.String(),
		})
	}

	bs.running = running
}

// GetFullAddress 获取完整地址
func (bs *BaseServer) GetFullAddress() string {
	return fmt.Sprintf("%s:%d", bs.host, bs.port)
}

// IncrementActiveConnections 增加活跃连接数
func (bs *BaseServer) IncrementActiveConnections() {
	bs.connectionsMu.Lock()
	defer bs.connectionsMu.Unlock()
	bs.activeConnections++
	bs.totalConnections++

	if bs.metricsCollector != nil {
		bs.metricsCollector.RecordConnection(bs.protocol, "open")
	}
}

// DecrementActiveConnections 减少活跃连接数
func (bs *BaseServer) DecrementActiveConnections() {
	bs.connectionsMu.Lock()
	defer bs.connectionsMu.Unlock()
	if bs.activeConnections > 0 {
		bs.activeConnections--
	}

	if bs.metricsCollector != nil {
		bs.metricsCollector.RecordConnection(bs.protocol, "close")
	}
}

// GetConnectionStats 获取连接统计
func (bs *BaseServer) GetConnectionStats() map[string]interface{} {
	bs.connectionsMu.RLock()
	defer bs.connectionsMu.RUnlock()

	return map[string]interface{}{
		"active_connections": bs.activeConnections,
		"total_connections":  bs.totalConnections,
	}
}

// GetMetrics 获取服务端指标
func (bs *BaseServer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 基础信息
	metrics["protocol"] = bs.protocol
	metrics["address"] = bs.GetFullAddress()
	metrics["running"] = bs.IsRunning()

	if !bs.StartTime.IsZero() {
		metrics["start_time"] = bs.StartTime
		if bs.IsRunning() {
			metrics["uptime"] = time.Since(bs.StartTime).String()
		}
	}

	// 连接统计
	connectionStats := bs.GetConnectionStats()
	for k, v := range connectionStats {
		metrics[k] = v
	}

	// 指标收集器数据
	if bs.metricsCollector != nil {
		collectorMetrics := bs.metricsCollector.GetMetrics()
		for k, v := range collectorMetrics {
			metrics[k] = v
		}
	}

	return metrics
}

// HealthCheck 健康检查
func (bs *BaseServer) HealthCheck(ctx context.Context) error {
	if !bs.IsRunning() {
		return fmt.Errorf("%s server is not running", bs.protocol)
	}

	if bs.healthChecker != nil {
		return bs.healthChecker.Check(ctx)
	}

	return nil
}

// RecordRequest 记录请求
func (bs *BaseServer) RecordRequest(operation string, duration time.Duration, success bool) {
	if bs.metricsCollector != nil {
		bs.metricsCollector.RecordRequest(bs.protocol, operation, duration, success)
	}
}

// RecordError 记录错误
func (bs *BaseServer) RecordError(operation string, errorType string) {
	if bs.metricsCollector != nil {
		bs.metricsCollector.RecordError(bs.protocol, operation, errorType)
	}
}

// LogInfo 记录信息日志
func (bs *BaseServer) LogInfo(msg string, fields ...map[string]interface{}) {
	if bs.logger != nil {
		bs.logger.Info(msg, fields...)
	}
}

// LogError 记录错误日志
func (bs *BaseServer) LogError(msg string, err error, fields ...map[string]interface{}) {
	if bs.logger != nil {
		bs.logger.Error(msg, err, fields...)
	}
}

// LogDebug 记录调试日志
func (bs *BaseServer) LogDebug(msg string, fields ...map[string]interface{}) {
	if bs.logger != nil {
		bs.logger.Debug(msg, fields...)
	}
}

// GetLogger 获取日志记录器
func (bs *BaseServer) GetLogger() interfaces.Logger {
	return bs.logger
}

// GetMetricsCollector 获取指标收集器
func (bs *BaseServer) GetMetricsCollector() interfaces.MetricsCollector {
	return bs.metricsCollector
}

// GetContext 获取上下文
func (bs *BaseServer) GetContext() context.Context {
	return bs.ctx
}

// GetStopChannel 获取停止信号通道
func (bs *BaseServer) GetStopChannel() <-chan struct{} {
	return bs.stopChan
}

// Shutdown 优雅关闭
func (bs *BaseServer) Shutdown(ctx context.Context) error {
	bs.SetRunning(false)
	bs.cancelFunc()
	close(bs.stopChan)
	return nil
}

// BaseConfig 通用配置基础实现
type BaseConfig struct {
	Protocol string `yaml:"protocol" json:"protocol"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
}

// GetProtocol 获取协议名称
func (bc *BaseConfig) GetProtocol() string {
	return bc.Protocol
}

// GetHost 获取监听主机
func (bc *BaseConfig) GetHost() string {
	if bc.Host == "" {
		return "localhost"
	}
	return bc.Host
}

// GetPort 获取监听端口
func (bc *BaseConfig) GetPort() int {
	return bc.Port
}

// GetAddress 获取完整监听地址
func (bc *BaseConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", bc.GetHost(), bc.Port)
}

// Validate 验证基础配置
func (bc *BaseConfig) Validate() error {
	if bc.Port <= 0 || bc.Port > 65535 {
		return fmt.Errorf("invalid port: %d, must be between 1 and 65535", bc.Port)
	}

	if bc.Protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	return nil
}

// Clone 克隆基础配置
func (bc *BaseConfig) Clone() interfaces.ServerConfig {
	return &BaseConfig{
		Protocol: bc.Protocol,
		Host:     bc.Host,
		Port:     bc.Port,
	}
}

// ToJSON 转换为JSON字符串
func (bc *BaseConfig) ToJSON() (string, error) {
	data, err := json.Marshal(bc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (bc *BaseConfig) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), bc)
}
