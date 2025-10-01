package operations

import (
	"context"
	"fmt"
	"net"
	"time"

	"abc-runner/app/adapters/udp/config"
	"abc-runner/app/core/interfaces"
)

// UDPOperations UDP操作执行器 - 遵循统一架构模式
type UDPOperations struct {
	conn             net.Conn
	packetConn       net.PacketConn
	config           *config.UDPConfig
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewUDPOperations 创建UDP操作执行器
func NewUDPOperations(
	conn net.Conn,
	packetConn net.PacketConn,
	config *config.UDPConfig,
	metricsCollector interfaces.DefaultMetricsCollector,
) *UDPOperations {
	return &UDPOperations{
		conn:             conn,
		packetConn:       packetConn,
		config:           config,
		metricsCollector: metricsCollector,
	}
}

// ExecuteOperation 执行UDP操作 - 统一操作入口
func (u *UDPOperations) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   u.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	var opErr error
	switch operation.Type {
	case "packet_send":
		opErr = u.executePacketSend(ctx, operation, result)
	case "packet_receive":
		opErr = u.executePacketReceive(ctx, operation, result)
	case "echo_test":
		opErr = u.executeEchoTest(ctx, operation, result)
	case "broadcast":
		opErr = u.executeBroadcast(ctx, operation, result)
	case "multicast":
		opErr = u.executeMulticast(ctx, operation, result)
	default:
		opErr = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Success = opErr == nil
	result.Error = opErr
	result.Duration = time.Since(startTime)

	// 添加操作特定元数据
	for k, v := range operation.Metadata {
		result.Metadata[k] = v
	}
	result.Metadata["protocol"] = "udp"
	result.Metadata["operation_type"] = operation.Type
	result.Metadata["packet_mode"] = u.config.UDPSpecific.PacketMode
	result.Metadata["execution_time_ms"] = float64(result.Duration.Nanoseconds()) / 1e6

	return result, opErr
}

// executePacketSend 执行数据包发送
func (u *UDPOperations) executePacketSend(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	testData := u.buildTestData(operation)
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("UDP_PACKET_%d", time.Now().Unix()))
	}

	var n int
	var err error

	switch u.config.UDPSpecific.PacketMode {
	case "unicast":
		if u.conn != nil {
			n, err = u.conn.Write(testData)
		} else {
			err = fmt.Errorf("unicast connection not available")
		}
	case "broadcast", "multicast":
		if u.packetConn != nil {
			addr := &net.UDPAddr{
				IP:   net.IPv4bcast,
				Port: u.config.Connection.Port,
			}
			n, err = u.packetConn.WriteTo(testData, addr)
		} else {
			err = fmt.Errorf("packet connection not available")
		}
	default:
		err = fmt.Errorf("unsupported packet mode: %s", u.config.UDPSpecific.PacketMode)
	}

	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}

	result.Value = n
	result.Metadata["sent_bytes"] = n
	result.Metadata["packet_size"] = len(testData)
	return nil
}

// executePacketReceive 执行数据包接收
func (u *UDPOperations) executePacketReceive(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	bufferSize := 4096
	if size, ok := operation.Params["buffer_size"].(int); ok && size > 0 {
		bufferSize = size
	}

	buffer := make([]byte, bufferSize)
	var n int
	var err error

	if u.packetConn != nil {
		n, _, err = u.packetConn.ReadFrom(buffer)
	} else if u.conn != nil {
		n, err = u.conn.Read(buffer)
	} else {
		err = fmt.Errorf("no connection available")
	}

	if err != nil {
		return fmt.Errorf("failed to receive packet: %w", err)
	}

	receivedData := buffer[:n]
	result.Value = receivedData
	result.Metadata["received_bytes"] = n
	result.Metadata["buffer_size"] = bufferSize
	return nil
}

// executeEchoTest 执行回显测试
func (u *UDPOperations) executeEchoTest(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// 简化实现：发送后尝试接收
	if err := u.executePacketSend(ctx, operation, result); err != nil {
		return err
	}

	// 清理result，准备接收
	sentBytes := result.Metadata["sent_bytes"]
	result.Metadata = make(map[string]interface{})

	if err := u.executePacketReceive(ctx, operation, result); err != nil {
		return err
	}

	result.Metadata["sent_bytes"] = sentBytes
	result.Metadata["echo_test"] = true
	return nil
}

// executeBroadcast 执行广播操作
func (u *UDPOperations) executeBroadcast(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	return u.executePacketSend(ctx, operation, result)
}

// executeMulticast 执行组播操作
func (u *UDPOperations) executeMulticast(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	return u.executePacketSend(ctx, operation, result)
}

// buildTestData 构造测试数据
func (u *UDPOperations) buildTestData(operation interfaces.Operation) []byte {
	if operation.Value != nil {
		switch v := operation.Value.(type) {
		case []byte:
			return v
		case string:
			return []byte(v)
		}
	}

	if u.config != nil && u.config.BenchMark.DataSize > 0 {
		data := make([]byte, u.config.BenchMark.DataSize)
		for i := range data {
			data[i] = byte('A' + (i % 26))
		}
		return data
	}

	return []byte("UDP_TEST_DATA")
}

// isReadOperation 判断是否为读操作
func (u *UDPOperations) isReadOperation(operationType string) bool {
	readOperations := map[string]bool{
		"packet_receive": true,
		"echo_test":      true,
		"packet_send":    false,
		"broadcast":      false,
		"multicast":      false,
	}
	return readOperations[operationType]
}

// GetSupportedOperations 获取支持的操作类型
func (u *UDPOperations) GetSupportedOperations() []string {
	return []string{
		"packet_send",
		"packet_receive",
		"echo_test",
		"broadcast",
		"multicast",
	}
}
