# Kafka测试指南

## 支持的Kafka操作

redis-runner支持以下Kafka操作：

- **生产者测试**: 消息生产性能测试
- **消费者测试**: 消息消费性能测试
- **混合测试**: 同时进行生产和消费测试

## 配置选项

### 命令行选项

```bash
# 基本选项
--broker <broker>     Kafka broker地址
--brokers <brokers>   多个Kafka broker地址 (逗号分隔)
--topic <topic>       主题名称
--test-type <type>    测试类型: produce, consume, produce_consume
--group-id <id>       消费者组ID

# 基准测试选项
-n <requests>         总消息数 (默认: 1000)
-c <connections>      并发连接数 (默认: 10)
--duration <time>     测试持续时间 (例如: 30s, 5m) - 覆盖-n
--message-size <size> 消息大小(字节) (默认: 1024)
```

### 配置文件选项

Kafka测试支持详细的配置文件：

```yaml
kafka:
  brokers:
    - "localhost:9092"
  producer:
    acks: "all"
    batch_size: 16384
    compression: "snappy"
  consumer:
    group_id: "test-group"
    auto_offset_reset: "latest"
  benchmark:
    default_topic: "test-topic"
    message_size: 1024
    test_type: "produce"
```

## 使用示例

### 基本生产者测试

```bash
# 简单的生产者测试
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# 持续时间测试
./redis-runner kafka --broker localhost:9092 --topic test-topic --duration 60s -c 10
```

### 消费者测试

```bash
# 消费者测试
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000
```

### 混合测试

```bash
# 同时进行生产和消费测试
./redis-runner kafka --brokers localhost:9092,localhost:9093 \
  --topic high-throughput --test-type produce_consume \
  --message-size 4096 --duration 60s -c 8
```

### 高性能测试

```bash
# 高性能测试（启用压缩）
./redis-runner kafka --broker localhost:9092 --topic perf-test \
  --compression lz4 --acks all --batch-size 32768 -n 50000
```

### 使用配置文件

```bash
# 使用配置文件进行复杂测试
./redis-runner kafka --config config/examples/kafka-producer.yaml
```

## 生产者配置

### 确认机制

```yaml
producer:
  acks: "all"  # 0, 1, all
```

### 批处理

```yaml
producer:
  batch_size: 16384
  linger_ms: "5ms"
```

### 压缩

```yaml
producer:
  compression: "snappy"  # none, gzip, snappy, lz4, zstd
```

## 消费者配置

### 消费者组

```yaml
consumer:
  group_id: "redis-runner-group"
```

### 偏移量管理

```yaml
consumer:
  auto_offset_reset: "latest"  # earliest, latest
  enable_auto_commit: true
```

### 拉取配置

```yaml
consumer:
  max_poll_records: 500
  fetch_min_bytes: 1024
  fetch_max_bytes: 52428800
```

## 安全配置

### TLS加密

```yaml
security:
  tls:
    enabled: true
    cert_file: "/path/to/client.crt"
    key_file: "/path/to/client.key"
    ca_file: "/path/to/ca.crt"
```

### SASL认证

```yaml
security:
  sasl:
    enabled: true
    mechanism: "SCRAM-SHA-512"
    username: "user"
    password: "password"
```

## 结果解读

Kafka测试完成后，redis-runner会输出详细的性能报告：

- **生产者指标**:
  - 消息生产速率
  - 平均消息大小
  - 批处理效率
  - 确认延迟

- **消费者指标**:
  - 消息消费速率
  - 消费者滞后
  - 重新平衡次数

- **通用指标**:
  - 端到端延迟
  - 吞吐量
  - 错误率

## 最佳实践

1. **预热**: 在正式测试前运行短时间的预热测试
2. **主题配置**: 确保测试主题有足够的分区
3. **资源监控**: 监控Kafka集群和客户端的资源使用情况
4. **网络**: 确保网络延迟不会影响测试结果
5. **消息大小**: 根据实际使用场景选择合适的消息大小
6. **并发连接**: 根据集群性能调整并发连接数
7. **确认机制**: 根据数据一致性要求选择合适的确认机制
8. **批处理**: 调整批处理大小以平衡延迟和吞吐量