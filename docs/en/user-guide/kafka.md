# Kafka Testing Guide

[English](kafka.md) | [中文](../zh/user-guide/kafka.md)

## Supported Kafka Operations

abc-runner supports the following Kafka operations:

- **Producer Testing**: Message production performance testing
- **Consumer Testing**: Message consumption performance testing
- **Mixed Testing**: Simultaneous production and consumption testing

## Configuration Options

### Command Line Options

```bash
# Basic options
--broker <broker>     Kafka broker address
--brokers <brokers>   Multiple Kafka broker addresses (comma-separated)
--topic <topic>       Topic name
--test-type <type>    Test type: produce, consume, produce_consume
--group-id <id>       Consumer group ID

# Benchmark options
-n <requests>         Total messages (default: 1000)
-c <connections>      Concurrent connections (default: 10)
--duration <time>     Test duration (e.g., 30s, 5m) - overrides -n
--message-size <size> Message size (bytes) (default: 1024)
```

### Configuration File Options

Kafka testing supports detailed configuration files:

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

## Usage Examples

### Basic Producer Testing

```bash
# Simple producer test
./abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# Duration-based test
./abc-runner kafka --broker localhost:9092 --topic test-topic --duration 60s -c 10
```

### Consumer Testing

```bash
# Consumer test
./abc-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000
```

### Mixed Testing

```bash
# Simultaneous production and consumption test
./abc-runner kafka --brokers localhost:9092,localhost:9093 \
  --topic high-throughput --test-type produce_consume \
  --message-size 4096 --duration 60s -c 8
```

### High-Performance Testing

```bash
# High-performance test (with compression)
./abc-runner kafka --broker localhost:9092 --topic perf-test \
  --compression lz4 --acks all --batch-size 32768 -n 50000
```

### Using Configuration Files

```bash
# Complex testing using configuration file
./abc-runner kafka --config config/examples/kafka-producer.yaml
```

## Producer Configuration

### Acknowledgment Mechanism

```yaml
producer:
  acks: "all"  # 0, 1, all
```

### Batch Processing

```yaml
producer:
  batch_size: 16384
  linger_ms: "5ms"
```

### Compression

```yaml
producer:
  compression: "snappy"  # none, gzip, snappy, lz4, zstd
```

## Consumer Configuration

### Consumer Groups

```yaml
consumer:
  group_id: "abc-runner-group"
```

### Offset Management

```yaml
consumer:
  auto_offset_reset: "latest"  # earliest, latest
  enable_auto_commit: true
```

### Fetch Configuration

```yaml
consumer:
  max_poll_records: 500
  fetch_min_bytes: 1024
  fetch_max_bytes: 52428800
```

## Security Configuration

### TLS Encryption

```yaml
security:
  tls:
    enabled: true
    cert_file: "/path/to/client.crt"
    key_file: "/path/to/client.key"
    ca_file: "/path/to/ca.crt"
```

### SASL Authentication

```yaml
security:
  sasl:
    enabled: true
    mechanism: "SCRAM-SHA-512"
    username: "user"
    password: "password"
```

## Result Interpretation

After Kafka testing is completed, abc-runner will output detailed performance reports:

- **Producer Metrics**:
  - Message production rate
  - Average message size
  - Batch processing efficiency
  - Acknowledgment latency

- **Consumer Metrics**:
  - Message consumption rate
  - Consumer lag
  - Rebalance count

- **General Metrics**:
  - End-to-end latency
  - Throughput
  - Error rate

## Best Practices

1. **Warm-up**: Run short warm-up tests before formal testing
2. **Topic Configuration**: Ensure test topics have sufficient partitions
3. **Resource Monitoring**: Monitor Kafka cluster and client resource usage
4. **Network**: Ensure network latency does not affect test results
5. **Message Size**: Choose appropriate message sizes based on actual usage scenarios
6. **Concurrent Connections**: Adjust concurrent connections based on cluster performance
7. **Acknowledgment Mechanism**: Choose appropriate acknowledgment mechanisms based on data consistency requirements
8. **Batch Processing**: Adjust batch size to balance latency and throughput