# gRPC 性能测试指南

[English](../en/user-guide/grpc.md) | [中文](grpc.md)

## 概述

abc-runner 的 gRPC 适配器提供全面的 gRPC 服务性能测试能力，支持连接池、流式操作和自定义元数据处理。

## 快速开始

### 基本 gRPC 测试

```bash
# 简单的 gRPC 性能测试
./abc-runner grpc --target localhost:9090 -n 1000 -c 10

# 使用自定义方法测试
./abc-runner grpc --target localhost:9090 --method "/example.Service/TestMethod" -n 1000

# 使用配置文件测试
./abc-runner grpc --config config/grpc.yaml
```

## 配置

### 连接配置

```yaml
grpc:
  target: "localhost:9090"
  timeout: "30s"
  max_recv_msg_size: 4194304
  max_send_msg_size: 4194304
  enable_reflection: true
  
  # TLS 配置
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
    insecure_skip_verify: false

  # 连接池
  pool:
    size: 10
    max_idle: 5
```

### 请求配置

```yaml
requests:
  - method: "/example.Service/UnaryMethod"
    payload: |
      {
        "message": "test",
        "value": 123
      }
    metadata:
      authorization: "Bearer token123"
      x-custom-header: "value"
    weight: 100

  - method: "/example.Service/StreamMethod"
    stream_type: "client_stream"
    payload: |
      {
        "data": "streaming test"
      }
    weight: 50
```

## 命令行选项

### 基本选项

```bash
./abc-runner grpc [选项]

选项:
  --target string         gRPC 服务器目标 (默认 "localhost:9090")
  --method string         要调用的 gRPC 方法
  --payload string        请求负载 (JSON 格式)
  --timeout duration      请求超时 (默认 30s)
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
  --reflection            启用服务器反射
  --metadata strings      自定义元数据 (key:value 格式)
```

## 测试场景

### 1. 一元 RPC 测试

测试简单的请求-响应 RPC 调用：

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/GetUser" \
  --payload '{"id": 123}' \
  -n 10000 -c 50
```

### 2. 服务器流测试

测试服务器端流式 RPC：

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/ListUsers" \
  --stream-type server_stream \
  -n 1000 -c 20
```

### 3. 客户端流测试

测试客户端流式 RPC：

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/CreateUsers" \
  --stream-type client_stream \
  -n 500 -c 10
```

### 4. 双向流测试

测试双向流式 RPC：

```bash  
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/ChatService" \
  --stream-type bidi_stream \
  -n 1000 -c 5
```

### 5. 带认证的负载测试

使用认证元数据进行测试：

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/SecureMethod" \
  --metadata "authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -n 5000 -c 25
```

## 监控和指标

### 内置指标

gRPC 适配器收集以下指标：

- **响应时间**: 请求/响应延迟
- **吞吐量**: 每秒请求数 (RPS)
- **错误率**: 失败请求百分比
- **连接指标**: 活跃连接数、连接错误
- **流指标**: 流持续时间、消息计数

### 自定义指标

```yaml
metrics:
  custom_labels:
    service: "example-service"
    environment: "production"
  
  collection_interval: "1s"
  export_interval: "10s"
```

## 错误处理

### 常见错误类型

1. **连接错误**: 网络连接问题
2. **认证错误**: 无效凭据或过期令牌
3. **方法错误**: 无效方法名或 protobuf 定义
4. **超时错误**: 请求超时
5. **资源错误**: 服务器资源耗尽

### 错误配置

```yaml
error_handling:
  retry_attempts: 3
  retry_delay: "1s"
  ignore_errors:
    - "UNAVAILABLE"
    - "RESOURCE_EXHAUSTED"
```

## 最佳实践

1. **连接管理**: 使用合适的连接池大小
2. **TLS 配置**: 在生产环境中启用 TLS
3. **负载大小**: 考虑消息大小限制
4. **流式传输**: 对大数据传输使用流式传输
5. **元数据**: 包含必要的认证和跟踪元数据
6. **错误处理**: 实现适当的重试逻辑
7. **监控**: 监控关键性能指标

## 故障排除

### 常见问题

1. **连接被拒绝**: 检查 gRPC 服务器是否运行
2. **方法未找到**: 验证方法名和服务定义
3. **认证失败**: 检查元数据和凭据
4. **超时错误**: 增加超时值或检查服务器性能
5. **TLS 错误**: 验证证书配置

### 调试模式

启用调试模式进行详细日志记录：

```bash
./abc-runner grpc --target localhost:9090 --debug -n 100
```

## 示例

请参阅 [配置示例](../../config/examples/) 目录获取完整的 gRPC 配置示例。