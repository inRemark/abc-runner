# WebSocket 性能测试指南

[English](../en/user-guide/websocket.md) | [中文](websocket.md)

## 概述

abc-runner 的 WebSocket 适配器提供全面的 WebSocket 连接性能测试能力，支持实时通信测试、自定义协议和消息模式。

## 快速开始

### 基本 WebSocket 测试

```bash
# 简单的 WebSocket 性能测试
./abc-runner websocket --url ws://localhost:8080/ws -n 1000 -c 10

# 使用自定义消息测试
./abc-runner websocket --url ws://localhost:8080/ws --message "Hello WebSocket" -n 1000

# 使用配置文件测试
./abc-runner websocket --config config/websocket.yaml
```

## 配置

### 连接配置

```yaml
websocket:
  url: "ws://localhost:8080/ws"
  timeout: "30s"
  handshake_timeout: "10s"
  read_buffer_size: 4096
  write_buffer_size: 4096
  
  # TLS 配置 (用于 wss://)
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
    insecure_skip_verify: false

  # 握手头部
  headers:
    Origin: "http://localhost:8080"
    Sec-WebSocket-Protocol: "chat"
    Authorization: "Bearer token123"
```

### 消息配置

```yaml
messages:
  - type: "text"
    content: "Hello WebSocket"
    weight: 70

  - type: "binary"
    content_base64: "SGVsbG8gV2ViU29ja2V0"
    weight: 20

  - type: "ping"
    weight: 10
```

### 测试模式

```yaml
test_patterns:
  - name: "echo_test"
    send_message: "ping"
    expect_response: true
    response_timeout: "5s"
    weight: 50

  - name: "broadcast_test"
    send_message: "broadcast:Hello everyone"
    expect_response: false
    weight: 30

  - name: "subscribe_test"
    send_message: "subscribe:channel1"
    expect_response: true
    keep_listening: true
    weight: 20
```

## 命令行选项

### 基本选项

```bash
./abc-runner websocket [选项]

选项:
  --url string            WebSocket URL (ws:// 或 wss://)
  --message string        要发送的消息
  --message-type string   消息类型: text, binary, ping, pong (默认 "text")
  --timeout duration      连接超时 (默认 30s)
  -n, --requests int      总消息数 (默认 1000)
  -c, --connections int   并发连接数 (默认 10)
  --config string         配置文件路径
```

### 高级选项

```bash
  --tls                   启用 TLS (wss://)
  --cert string           客户端证书文件
  --key string            客户端私钥文件
  --ca string             CA 证书文件
  --insecure              跳过 TLS 证书验证
  --header strings        自定义头部 (key:value 格式)
  --subprotocol strings   WebSocket 子协议
  --origin string         Origin 头部值
```

## 测试场景

### 1. 回声测试

测试基本消息回声功能：

```bash
./abc-runner websocket \
  --url ws://localhost:8080/echo \
  --message "Hello Echo" \
  -n 10000 -c 50
```

### 2. 聊天应用测试

使用多个连接测试聊天应用：

```bash
./abc-runner websocket \
  --url ws://localhost:8080/chat \
  --header "Authorization:Bearer token123" \
  --message "Hello Chat Room" \
  -n 5000 -c 100
```

### 3. 实时数据流测试

测试实时数据流：

```bash
./abc-runner websocket \
  --url ws://localhost:8080/stream \
  --subprotocol "data-stream" \
  --message '{"type":"subscribe","channel":"market-data"}' \
  -n 1000 -c 20
```

### 4. 二进制数据测试

测试二进制消息传输：

```bash
./abc-runner websocket \
  --url ws://localhost:8080/binary \
  --message-type binary \
  --message "$(echo -n 'Binary Data' | base64)" \
  -n 2000 -c 25
```

### 5. 安全 WebSocket 测试

使用 TLS 加密测试：

```bash
./abc-runner websocket \
  --url wss://localhost:8443/secure \
  --tls \
  --cert client.crt \
  --key client.key \
  -n 1000 -c 10
```

## 消息类型

### 文本消息

```yaml
messages:
  - type: "text"
    content: "纯文本消息"
```

### 二进制消息

```yaml
messages:
  - type: "binary"
    content_base64: "SGVsbG8gV29ybGQ="  # Base64 编码
```

### JSON 消息

```yaml
messages:
  - type: "text"
    content: |
      {
        "type": "message",
        "data": "Hello JSON",
        "timestamp": "2025-01-02T10:00:00Z"
      }
```

### 控制消息

```yaml
messages:
  - type: "ping"
    content: "ping data"
  
  - type: "pong"
    content: "pong data"
```

## 监控和指标

### 内置指标

WebSocket 适配器收集以下指标：

- **连接指标**: 连接成功率、握手时间
- **消息指标**: 每秒发送/接收消息数
- **延迟指标**: 回声消息的往返时间
- **错误指标**: 连接错误、消息错误
- **吞吐量指标**: 每秒发送/接收字节数

### 实时监控

```yaml
monitoring:
  enabled: true
  interval: "1s"
  metrics:
    - connection_count
    - message_rate
    - error_rate
    - latency_p95
```

## 连接管理

### 连接生命周期

```yaml
connection:
  # 连接建立
  connect_timeout: "10s"
  handshake_timeout: "5s"
  
  # 保活设置
  ping_interval: "30s"
  pong_timeout: "10s"
  
  # 重连设置
  auto_reconnect: true
  reconnect_delay: "5s"
  max_reconnect_attempts: 3
```

### 连接池

```yaml
pool:
  max_connections: 100
  idle_timeout: "300s"
  connection_lifetime: "3600s"
```

## 错误处理

### 常见错误类型

1. **连接错误**: 网络连接问题
2. **握手错误**: HTTP 升级失败
3. **协议错误**: WebSocket 协议违规
4. **消息错误**: 无效消息格式
5. **超时错误**: 操作超时

### 错误恢复

```yaml
error_handling:
  retry_on_failure: true
  max_retries: 3
  retry_delay: "1s"
  
  ignore_errors:
    - "close_normal"
    - "close_going_away"
```

## 最佳实践

1. **连接管理**: 使用适当的连接限制
2. **消息大小**: 考虑消息大小和缓冲区限制
3. **错误处理**: 实现适当的重连逻辑
4. **安全性**: 在生产环境中使用 WSS
5. **监控**: 监控连接健康和消息速率
6. **资源管理**: 正确清理连接
7. **协议合规**: 遵循 WebSocket 协议标准

## 故障排除

### 常见问题

1. **连接被拒绝**: 检查 WebSocket 服务器是否运行
2. **握手失败**: 验证 URL 和头部
3. **协议错误**: 检查消息格式和协议合规性
4. **超时错误**: 增加超时值
5. **TLS 错误**: 验证证书配置

### 调试模式

启用调试模式进行详细日志记录：

```bash
./abc-runner websocket --url ws://localhost:8080/ws --debug -n 100
```

## 示例

请参阅 [配置示例](../../config/examples/) 目录获取完整的 WebSocket 配置示例。