# Kafka Testing Guide

This guide covers Kafka-specific features and usage patterns for redis-runner.

## Basic Kafka Testing

### Producer Testing

Test Kafka producer performance:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5
```

### Consumer Testing

Test Kafka consumer performance:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000
```

### Mixed Producer/Consumer Testing

Test end-to-end Kafka performance:

```bash
./redis-runner kafka --brokers localhost:9092,localhost:9093 \
  --topic high-throughput --test-type produce_consume \
  --message-size 4096 --duration 60s -c 8
```

## Kafka Configuration

### Single Broker

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000
```

### Multiple Brokers

```bash
./redis-runner kafka --brokers localhost:9092,localhost:9093,localhost:9094 \
  --topic test-topic -n 10000
```

## Test Types

### Produce Only (Default)

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type produce -n 10000
```

### Consume Only

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id test-group -n 1000
```

### Produce and Consume

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type produce_consume -n 10000
```

## Message Configuration

### Message Size

Control the size of messages produced:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --message-size 1024 -n 10000
```

### Message Compression

Test with different compression algorithms:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --compression lz4 --message-size 4096 -n 10000
```

## Configuration File Example

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

Run with configuration:
```bash
./redis-runner kafka --config kafka.yaml
```

## Advanced Features

### Security Configuration

Test with SSL/TLS enabled Kafka clusters:

```bash
./redis-runner kafka --broker localhost:9093 --topic test-topic \
  --tls-enabled --tls-skip-verify -n 1000
```

### SASL Authentication

Test with SASL/PLAIN authentication:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --sasl-enabled --sasl-user user --sasl-password pass -n 1000
```

### Consumer Group Management

Test with specific consumer group settings:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id performance-test \
  --consumer-group-reset earliest -n 1000
```

## Performance Tuning

### Batch Size

Optimize producer batch size for throughput:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --batch-size 32768 --message-size 1024 -n 100000
```

### Acknowledgement Settings

Test with different acknowledgement settings:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --acks all --message-size 1024 -n 10000
```

### Concurrency Control

Adjust the number of parallel connections:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  -n 100000 -c 20
```

### Duration-Based Testing

Run tests for a specific duration instead of a fixed number of messages:

```bash
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --duration 300s -c 10
```