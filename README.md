# redis-runner

[English](README.md) | [中文](README_zh.md)

## About

A unified performance testing tool for Redis, HTTP, and Kafka protocols.

⚠️ **Breaking Change Notice**: This version (v3.0.0) introduces breaking changes. See [Migration Guide](COMMAND_REFACTOR_MIGRATION_GUIDE.md) for upgrade instructions.

## Features

### Redis Testing

- Support for Redis cluster, sentinel, and standalone modes
- Multiple test cases: set_get_random, set_only, get_only, incr, append, etc.
- Configurable read/write ratios and TTL
- Global self-increasing or random key generation

### HTTP Testing  

- Support for GET, POST, PUT, DELETE methods
- Custom headers and request bodies
- Connection pooling and keep-alive
- Duration-based and request-count-based testing

### Kafka Testing

- Producer and consumer performance testing
- Support for multiple brokers and topics
- Configurable message sizes and compression
- Mixed produce/consume workloads

## Quick Start

### Installation

```bash
# Build from source
go build -o redis-runner .

# Or download pre-built binaries from releases
```

### Basic Usage

```bash
# Show help
./redis-runner --help

# Redis performance test
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50

# HTTP load test
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# Kafka performance test
./redis-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5
```

### Using Aliases

```bash
# Short aliases for quick testing
./redis-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./redis-runner h --url http://httpbin.org/get -n 100  # HTTP
./redis-runner k --broker localhost:9092 -n 100      # Kafka
```

## Command Reference

### Global Options

```bash
./redis-runner --help                 # Show help
./redis-runner --version              # Show version
```

### Redis Commands

```bash
# Basic Redis test
./redis-runner redis -h <host> -p <port> -n <requests> -c <connections>

# Redis with authentication
./redis-runner redis -h localhost -p 6379 -a password -n 10000 -c 50

# Redis cluster mode
./redis-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50

# Custom test case with read ratio
./redis-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80

# Using configuration file
./redis-runner redis --config conf/redis.yaml
```

### HTTP Commands

```bash
# Basic HTTP GET test
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# HTTP POST with body
./redis-runner http --url http://api.example.com/users \n  --method POST --body '{"name":"test"}' \n  --content-type application/json -n 1000 -c 20

# Duration-based test
./redis-runner http --url http://localhost:8080 --duration 60s -c 100

# Custom headers
./redis-runner http --url http://api.example.com \n  --header "Authorization:Bearer token123" \n  --header "X-API-Key:secret" -n 1000
```

### Kafka Commands

```bash
# Basic producer test
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# Consumer test
./redis-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id my-group -n 1000

# Mixed produce/consume test
./redis-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic high-throughput --test-type produce_consume \n  --message-size 4096 --duration 60s -c 8

# High-performance test with compression
./redis-runner kafka --broker localhost:9092 --topic perf-test \n  --compression lz4 --acks all --batch-size 32768 -n 50000
```

## Configuration Files

You can use YAML configuration files for complex setups:

### Redis Configuration (conf/redis.yaml)

```yaml
protocol: redis
connection:
  host: localhost
  port: 6379
  mode: standalone  # standalone, cluster, sentinel
  timeout: 30s

benchmark:
  total: 10000
  parallels: 50
  test_case: "set_get_random"
  data_size: 64
  read_ratio: 0.5
```

### HTTP Configuration (conf/http.yaml)

```yaml
protocol: http
connection:
  base_url: "http://localhost:8080"
  timeout: 30s
  max_conns_per_host: 50

benchmark:
  total: 10000
  parallels: 50
  method: "GET"
  path: "/api/test"
  headers:
    "Content-Type": "application/json"
    "Authorization": "Bearer token"
```

### Kafka Configuration (conf/kafka.yaml)

```yaml
protocol: kafka
brokers: ["localhost:9092"]
topic_configs:
  - name: "test-topic"
    partitions: 3

producer:
  batch_size: 16384
  compression: "snappy"
  required_acks: 1

consumer:
  group_id: "test-group"
  auto_offset_reset: "earliest"

benchmark:
  total: 10000
  parallels: 5
  message_size: 1024
  test_type: "produce"
```

## Migration from v2.x

This version introduces breaking changes. Key changes:

- `redis-enhanced` → `redis`
- `http-enhanced` → `http`  
- `kafka-enhanced` → `kafka`
- Simplified command structure
- Unified configuration format

See the [Migration Guide](COMMAND_REFACTOR_MIGRATION_GUIDE.md) for detailed upgrade instructions.

## Examples

### Redis Performance Testing

```bash
# Basic performance test
./redis-runner redis -h 127.0.0.1 -p 6379 -n 100000 -c 50

# Cluster mode with authentication
./redis-runner redis --mode cluster -h localhost -p 6371 \n  -a "password" -n 100000 -c 10 -d 64 --read-ratio 50

# Custom test patterns
./redis-runner redis -t incr -n 50000 -c 100  # Counter operations
./redis-runner redis -t lpush_lpop -n 10000 -c 50  # List operations
```

### HTTP Load Testing

```bash
# API endpoint testing
./redis-runner http --url http://api.example.com/health -n 10000 -c 100

# POST with JSON payload
./redis-runner http --url http://api.example.com/users \n  --method POST \n  --body '{"name":"John","email":"john@example.com"}' \n  --content-type "application/json" -n 1000 -c 20

# Load testing with ramp-up
./redis-runner http --url http://www.example.com \n  --duration 300s -c 200 --ramp-up 30s
```

### Kafka Performance Testing

```bash
# Producer throughput test
./redis-runner kafka --broker localhost:9092 --topic throughput-test \n  --message-size 1024 -n 100000 -c 10

# Consumer lag test
./redis-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id perf-test-group -n 50000

# End-to-end latency test
./redis-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic latency-test --test-type produce_consume \n  --message-size 512 --duration 120s -c 5
```

## License

[MIT](LICENSE)

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## Support

For questions and support:

- Check the [Migration Guide](COMMAND_REFACTOR_MIGRATION_GUIDE.md)
- Review command help: `./redis-runner <command> --help`
- Open an issue for bug reports or feature requests
