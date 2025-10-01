# TCP 性能测试指南

[English](../en/user-guide/tcp.md) | [中文](tcp.md)

## 概述

abc-runner 的 TCP 适配器提供全面的 TCP 连接性能测试能力，支持自定义负载、连接模式和网络性能分析。

## 快速开始

### 基本 TCP 测试

```bash
# 简单的 TCP 性能测试
./abc-runner tcp --host localhost --port 8080 -n 1000 -c 10

# 使用自定义负载测试
./abc-runner tcp --host localhost --port 8080 --payload "Hello TCP" -n 1000

# 使用配置文件测试
./abc-runner tcp --config config/tcp.yaml
```

## 配置

### 连接配置

```yaml
tcp:
  host: "localhost"
  port: 8080
  timeout: "30s"
  connect_timeout: "10s"
  read_timeout: "30s"
  write_timeout: "30s"
  
  # Socket 选项
  socket:
    keep_alive: true
    keep_alive_period: "30s"
    no_delay: true
    buffer_size: 4096
    
  # TLS 配置 (用于安全连接)
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
    insecure_skip_verify: false
```

### 负载配置

```yaml
payloads:
  - type: "text"
    content: "Hello TCP Server"
    weight: 50

  - type: "binary"
    content_hex: "48656c6c6f20544350"  # "Hello TCP" 的十六进制
    weight: 30

  - type: "random" 
    size: 1024
    weight: 20
```

### 测试模式

```yaml
test_patterns:
  - name: "echo_test"
    send_data: "PING"
    expect_response: true
    response_size: 4
    weight: 60

  - name: "throughput_test"
    send_data: "DATA:${random_1024}"
    expect_response: false
    weight: 40
```

## 命令行选项

### 基本选项

```bash
./abc-runner tcp [选项]

选项:
  --host string           TCP 服务器主机 (默认 "localhost")
  --port int              TCP 服务器端口 (默认 8080)
  --payload string        要发送的数据负载
  --payload-size int      随机负载大小（字节）
  --timeout duration      连接超时 (默认 30s)
  -n, --requests int      总请求数 (默认 1000)
  -c, --connections int   并发连接数 (默认 10)
  --config string         配置文件路径
```

### 高级选项

```bash
  --tls                   启用 TLS 连接
  --cert string           客户端证书文件
  --key string            客户端私钥文件
  --ca string             CA 证书文件
  --insecure              跳过 TLS 证书验证
  --keep-alive            启用 TCP keep-alive
  --no-delay              启用 TCP_NODELAY
  --buffer-size int       Socket 缓冲区大小 (默认 4096)
```

## 测试场景

### 1. 回声服务器测试

测试基本回声功能：

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload "ECHO:Hello World" \
  -n 10000 -c 50
```

### 2. 吞吐量测试

使用大负载测试网络吞吐量：

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload-size 8192 \
  -n 5000 -c 25
```

### 3. 连接压力测试

测试连接建立和断开：

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload "CONNECT" \
  --duration 60s \
  -c 100
```

### 4. 协议测试

测试自定义协议实现：

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload "GET /status HTTP/1.1\r\nHost: localhost\r\n\r\n" \
  -n 1000 -c 20
```

### 5. 安全 TCP 测试

使用 TLS 加密测试：

```bash
./abc-runner tcp \
  --host localhost \
  --port 8443 \
  --tls \
  --cert client.crt \
  --key client.key \
  -n 1000 -c 10
```

## 负载类型

### 文本负载

```yaml
payloads:
  - type: "text"
    content: "Hello TCP Server"
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
    size: 1024
    pattern: "alphanumeric"  # alphanumeric, binary, hex
```

### 模板负载

```yaml
payloads:
  - type: "template"
    content: "ID:${sequence},DATA:${random_string_32}"
```

## 连接模式

### 单连接模式

```yaml
connection_pattern:
  type: "single"
  reuse_connection: true
  max_requests_per_connection: 1000
```

### 连接池模式

```yaml
connection_pattern:
  type: "pool"
  pool_size: 50
  max_idle: 10
  idle_timeout: "300s"
```

### 每请求连接模式

```yaml
connection_pattern:
  type: "per_request"
  connect_timeout: "5s"
  close_after_request: true
```

## 监控和指标

### 内置指标

TCP 适配器收集以下指标：

- **连接指标**: 连接成功率、连接时间
- **吞吐量指标**: 每秒发送/接收字节数
- **延迟指标**: 往返时间、连接时间
- **错误指标**: 连接错误、超时错误
- **网络指标**: 丢包率、重传

### 自定义指标

```yaml
metrics:
  collection_interval: "1s"
  custom_metrics:
    - name: "tcp_connections_active"
      type: "gauge"
    - name: "tcp_bytes_total"
      type: "counter"
```

## 网络分析

### 延迟分析

```yaml
latency_analysis:
  enabled: true
  percentiles: [50, 90, 95, 99]
  histogram_buckets: [1, 5, 10, 25, 50, 100, 250, 500, 1000]
```

### 吞吐量分析

```yaml
throughput_analysis:
  enabled: true
  measurement_interval: "1s"
  bandwidth_limit: "100MB/s"
```

## 错误处理

### 常见错误类型

1. **连接错误**: 网络连接问题
2. **超时错误**: 连接或操作超时
3. **Socket 错误**: 底层 socket 错误
4. **协议错误**: 应用协议错误
5. **资源错误**: 系统资源耗尽

### 错误恢复

```yaml
error_handling:
  retry_on_connection_error: true
  max_retries: 3
  retry_delay: "1s"
  
  fail_fast_errors:
    - "connection_refused"
    - "host_unreachable"
```

## Socket 调优

### 缓冲区大小

```yaml
socket_tuning:
  send_buffer_size: 65536
  recv_buffer_size: 65536
  write_buffer_size: 8192
  read_buffer_size: 8192
```

### TCP 选项

```yaml
tcp_options:
  tcp_nodelay: true
  tcp_keepalive: true
  tcp_keepalive_time: 7200
  tcp_keepalive_interval: 75
  tcp_keepalive_probes: 9
```

## 最佳实践

1. **连接管理**: 使用适当的连接模式
2. **缓冲区调优**: 为您的用例优化缓冲区大小
3. **超时设置**: 设置适当的超时值
4. **错误处理**: 实现强大的错误恢复
5. **资源监控**: 监控系统资源
6. **网络调优**: 为性能调优 TCP 参数
7. **安全性**: 对敏感数据使用 TLS

## 故障排除

### 常见问题

1. **连接被拒绝**: 检查 TCP 服务器是否运行，端口是否开放
2. **超时错误**: 增加超时值或检查网络延迟
3. **Socket 错误**: 检查系统限制和网络配置
4. **性能问题**: 优化缓冲区大小和连接模式
5. **TLS 错误**: 验证证书配置

### 调试模式

启用调试模式进行详细日志记录：

```bash
./abc-runner tcp --host localhost --port 8080 --debug -n 100
```

### 网络诊断

```bash
# 检查连接性
telnet localhost 8080

# 检查网络统计
netstat -an | grep 8080

# 监控网络流量
tcpdump -i any port 8080
```

## 示例

请参阅 [配置示例](../../config/examples/) 目录获取完整的 TCP 配置示例。