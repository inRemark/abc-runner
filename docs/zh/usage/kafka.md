# Kafka测试指南

[English](kafka.md) | [中文](kafka.zh.md)

本指南涵盖了abc-runner的Kafka特定功能和使用模式。

## 基本Kafka测试

### 生产者测试

测试Kafka生产者性能：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5
```

### 消费者测试

测试Kafka消费者性能：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000
```

### 混合生产者/消费者测试

测试端到端Kafka性能：

```bash
./abc-runner kafka --brokers localhost:9092,localhost:9093 \
  --topic high-throughput --test-type produce_consume \
  --message-size 4096 --duration 60s -c 8
```

## Kafka配置

### 单个代理

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000
```

### 多个代理

```bash
./abc-runner kafka --brokers localhost:9092,localhost:9093,localhost:9094 \
  --topic test-topic -n 10000
```

## 测试类型

### 仅生产（默认）

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type produce -n 10000
```

### 仅消费

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id test-group -n 1000
```

### 生产和消费

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type produce_consume -n 10000
```

## 消息配置

### 消息大小

控制生产的消息大小：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --message-size 1024 -n 10000
```

### 消息压缩

使用不同的压缩算法进行测试：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --compression lz4 --message-size 4096 -n 10000
```

## 配置文件示例

```yaml
# kafka.yaml
protocol: kafka
brokers: ["localhost:9092"]
topic_configs:
  - name: "test-topic"
    partitions: 3

producer:
  batch_size: 16384
  compression: "snappy"
  required_acks: 1
  max_message_bytes: 1048576

consumer:
  group_id: "test-group"
  auto_offset_reset: "earliest"
  fetch_min_bytes: 1
  fetch_max_wait: 500ms

benchmark:
  total: 10000
  parallels: 5
  message_size: 1024
  test_type: "produce"
```

使用配置文件运行：

```bash
./abc-runner kafka --config kafka.yaml
```

## 高级功能

### 安全配置

测试启用SSL/TLS的Kafka集群：

```bash
./abc-runner kafka --broker localhost:9093 --topic test-topic \
  --tls-enabled --tls-skip-verify -n 1000
```

### SASL认证

测试使用SASL/PLAIN认证：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --sasl-enabled --sasl-user user --sasl-password pass -n 1000
```

### 消费者组管理

使用特定的消费者组设置进行测试：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id performance-test \
  --consumer-group-reset earliest -n 1000
```

## 性能调优

### 批量大小

优化生产者的批量大小以提高吞吐量：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --batch-size 32768 --message-size 1024 -n 100000
```

### 确认设置

测试不同的确认设置：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --acks all --message-size 1024 -n 10000
```

### 并发控制

调整并行连接数：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  -n 100000 -c 20
```

### 基于时长的测试

运行特定时长的测试而不是固定数量的消息：

```bash
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --duration 300s -c 10
```