# UDP 性能测试指南

[English](../en/user-guide/udp.md) | [中文](udp.md)

## 概述

abc-runner 的 UDP 适配器提供全面的 UDP 连接性能测试能力，支持数据报传输测试、丢包分析和网络可靠性评估。

## 快速开始

### 基本 UDP 测试

```bash
# 简单的 UDP 性能测试
./abc-runner udp --host localhost --port 8080 -n 1000 -c 10

# 使用自定义负载测试
./abc-runner udp --host localhost --port 8080 --payload "Hello UDP" -n 1000

# 使用配置文件测试
./abc-runner udp --config config/udp.yaml
```

## 配置

### 连接配置

```yaml
udp:
  host: "localhost"
  port: 8080
  timeout: "30s"
  read_timeout: "5s"
  write_timeout: "5s"
  
  # Socket 选项
  socket:
    buffer_size: 65536
    read_buffer_size: 65536
    write_buffer_size: 65536
    
  # 数据包选项
  packet:
    max_size: 1472  # MTU - IP 头部 - UDP 头部
    fragment_threshold: 1400
```

### 负载配置

```yaml
payloads:
  - type: "text"
    content: "Hello UDP Server"
    weight: 50

  - type: "binary"
    content_hex: "48656c6c6f20555350"  # "Hello UDP" 的十六进制
    weight: 30

  - type: "random"
    size: 512
    weight: 20
```

### 测试模式

```yaml
test_patterns:
  - name: "ping_test"
    send_data: "PING:${timestamp}"
    expect_response: true
    response_timeout: "1s"
    weight: 60

  - name: "broadcast_test"
    send_data: "BROADCAST:${random_string}"
    expect_response: false
    weight: 40
```

## 命令行选项

### 基本选项

```bash
./abc-runner udp [选项]

选项:
  --host string           UDP 服务器主机 (默认 "localhost")
  --port int              UDP 服务器端口 (默认 8080)
  --payload string        要发送的数据负载
  --payload-size int      随机负载大小（字节）
  --timeout duration      操作超时 (默认 30s)
  -n, --requests int      总数据包数 (默认 1000)
  -c, --connections int   并发发送者数 (默认 10)
  --config string         配置文件路径
```

### 高级选项

```bash
  --packet-size int       最大数据包大小 (默认 1472)
  --buffer-size int       Socket 缓冲区大小 (默认 65536)
  --no-fragment           不允许数据包分片
  --broadcast             启用广播模式
  --multicast string      多播组地址
  --ttl int               数据包生存时间 (默认 64)
```

## 测试场景

### 1. 回声服务器测试

测试基本 UDP 回声功能：

```bash
./abc-runner udp \
  --host localhost \
  --port 8080 \
  --payload "ECHO:Hello World" \
  -n 10000 -c 50
```

### 2. 丢包测试

使用高频发送测试丢包：

```bash
./abc-runner udp \
  --host localhost \
  --port 8080 \
  --payload-size 1024 \
  --duration 60s \
  -c 100
```

### 3. 吞吐量测试

使用大数据包测试 UDP 吞吐量：

```bash
./abc-runner udp \
  --host localhost \
  --port 8080 \
  --packet-size 1472 \
  --payload-size 1400 \
  -n 5000 -c 25
```

### 4. 广播测试

测试 UDP 广播功能：

```bash
./abc-runner udp \
  --host 192.168.1.255 \
  --port 8080 \
  --broadcast \
  --payload "BROADCAST:Hello Network" \
  -n 1000 -c 5
```

### 5. 多播测试

测试 UDP 多播功能：

```bash
./abc-runner udp \
  --multicast 224.0.0.1 \
  --port 8080 \
  --payload "MULTICAST:Hello Group" \
  -n 1000 -c 10
```

## 负载类型

### 文本负载

```yaml
payloads:
  - type: "text"
    content: "Hello UDP Server"
```

### 二进制负载

```yaml
payloads:
  - type: "binary"
    content_hex: "deadbeef"
```

### 随机负载

```yaml
payloads:
  - type: "random"
    size: 512
    pattern: "binary"  # binary, hex, alphanumeric
```

### 结构化负载

```yaml
payloads:
  - type: "structured"
    format: "json"
    content: |
      {
        "id": "${sequence}",
        "timestamp": "${timestamp}",
        "data": "${random_string_32}"
      }
```

## 数据包配置

### 大小优化

```yaml
packet_config:
  # 针对网络 MTU 优化
  max_packet_size: 1472  # 以太网 MTU (1500) - IP 头部 (20) - UDP 头部 (8)
  
  # 分片设置
  allow_fragmentation: false
  fragment_threshold: 1400
  
  # 巨型帧（如果支持）
  jumbo_frames: false
  jumbo_packet_size: 9000
```

### 服务质量

```yaml
qos:
  dscp: 0  # 区分服务代码点
  tos: 0   # 服务类型
  priority: 0
```

## 监控和指标

### 内置指标

UDP 适配器收集以下指标：

- **数据包指标**: 发送/接收数据包数、丢包率
- **吞吐量指标**: 每秒发送/接收字节数
- **延迟指标**: 往返时间（用于回声测试）
- **错误指标**: 发送错误、接收错误
- **网络指标**: 抖动、数据包重排

### 丢包分析

```yaml
packet_loss_analysis:
  enabled: true
  sequence_tracking: true
  duplicate_detection: true
  reorder_detection: true
  
  # 丢包率阈值
  warning_threshold: 1.0   # 1% 丢包
  critical_threshold: 5.0  # 5% 丢包
```

### 抖动分析

```yaml
jitter_analysis:
  enabled: true
  measurement_window: "10s"
  percentiles: [50, 90, 95, 99]
```

## 网络配置

### 广播模式

```yaml
broadcast:
  enabled: true
  address: "192.168.1.255"
  interface: "eth0"  # 可选：指定接口
```

### 多播模式

```yaml
multicast:
  enabled: true
  group: "224.0.0.1"
  interface: "0.0.0.0"  # 监听所有接口
  ttl: 1  # 多播 TTL
```

### Socket 选项

```yaml
socket_options:
  so_reuseaddr: true
  so_reuseport: true
  so_broadcast: true
  so_rcvbuf: 65536
  so_sndbuf: 65536
```

## 错误处理

### 常见错误类型

1. **网络错误**: 网络不可达、主机不可达
2. **Socket 错误**: 缓冲区溢出、socket 创建错误
3. **超时错误**: 发送/接收超时
4. **数据包错误**: 数据包过大、分片错误
5. **配置错误**: 无效地址、端口冲突

### 错误恢复

```yaml
error_handling:
  retry_on_error: false  # UDP 是即发即弃
  ignore_send_errors: false
  ignore_receive_errors: false
  
  # 特定错误处理
  handle_icmp_errors: true
  handle_buffer_full: true
```

## 性能调优

### 缓冲区优化

```yaml
performance:
  # 增加缓冲区大小以获得高吞吐量
  socket_buffer_size: 1048576  # 1MB
  
  # 批处理
  batch_size: 100
  batch_timeout: "1ms"
  
  # CPU 优化
  cpu_affinity: [0, 1, 2, 3]
  use_kernel_bypass: false  # DPDK 支持（如果可用）
```

### 速率限制

```yaml
rate_limiting:
  enabled: true
  packets_per_second: 10000
  bytes_per_second: 10485760  # 10MB/s
  burst_size: 1000
```

## 最佳实践

1. **数据包大小**: 保持数据包在 MTU 大小以下以避免分片
2. **错误处理**: 不要依赖传递保证
3. **缓冲区管理**: 使用适当的缓冲区大小
4. **速率控制**: 实现速率限制以避免网络过载
5. **监控**: 监控丢包和抖动
6. **网络调优**: 优化网络堆栈参数
7. **测试环境**: 在现实网络条件下测试

## 故障排除

### 常见问题

1. **丢包**: 检查网络容量和缓冲区大小
2. **高抖动**: 调查网络拥塞
3. **发送错误**: 检查 socket 缓冲区大小和速率限制
4. **接收错误**: 验证服务器正在监听和处理数据包
5. **防火墙问题**: 确保 UDP 端口开放

### 调试模式

启用调试模式进行详细日志记录：

```bash
./abc-runner udp --host localhost --port 8080 --debug -n 100
```

### 网络诊断

```bash
# 检查 UDP 连接性
nc -u localhost 8080

# 监控 UDP 流量
tcpdump -i any udp port 8080

# 检查 socket 统计
ss -u -a | grep 8080

# 检查丢包
ping -c 100 hostname
```

## 可靠性测试

### 丢包模拟

```yaml
reliability_testing:
  simulate_packet_loss: true
  loss_rate: 1.0  # 1% 丢包

  simulate_jitter: true
  jitter_range: "1ms-10ms"
  
  simulate_reordering: true
  reorder_rate: 0.1  # 0.1% 数据包重排
```

### 压力测试

```yaml
stress_testing:
  flood_test: true
  max_rate: 100000  # 每秒数据包数
  duration: "60s"
  
  buffer_overflow_test: true
  large_packet_test: true
  concurrent_senders: 100
```

## 示例

请参阅 [配置示例](../../config/examples/) 目录获取完整的 UDP 配置示例。