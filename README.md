# abc-runner

[English](README.md) | [中文](README.zh.md)

## About

A unified performance testing tool for Redis, HTTP, and Kafka protocols.

⚠️ **Breaking Change Notice**: This version (v0.1.0) introduces breaking changes. See [Migration Guide](docs/CHANGELOG.md) for upgrade instructions.

## Features

### Redis Testing

- Support for Redis cluster, sentinel, and standalone modes
- Multiple test cases: set_get_random, set_only, get_only, incr, decr, lpush, rpush, lpop, rpop, sadd, smembers, zadd, zrange, hset, hget, hmset, hmget, hgetall, pub, sub, etc.
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
go build -o abc-runner .

# Or download pre-built binaries from releases
```

### Basic Usage

```bash
# Show help
./abc-runner --help

# Redis performance test
./abc-runner redis -h localhost -p 6379 -n 10000 -c 50

# HTTP load test
./abc-runner http --url http://localhost:8080 -n 10000 -c 50

# Kafka performance test
./abc-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5
```

### Using Aliases

```bash
# Short aliases for quick testing
./abc-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./abc-runner h --url http://httpbin.org/get -n 100  # HTTP
./abc-runner k --broker localhost:9092 -n 100      # Kafka
```

## Command Reference

### Global Options

```bash
./abc-runner --help                 # Show help
./abc-runner --version              # Show version
```

### Redis Commands

```bash
# Basic Redis test
./abc-runner redis -h <host> -p <port> -n <requests> -c <connections>

# Redis with authentication
./abc-runner redis -h localhost -p 6379 -a password -n 10000 -c 50

# Redis cluster mode
./abc-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50

# Custom test case with read ratio
./abc-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80

# Using configuration file
./abc-runner redis --config config/redis.yaml

# Using configuration file with core configuration
./abc-runner redis --config config/redis.yaml --core-config config/core.yaml
```

Supported Redis test cases (`-t` option):
- `get`: Simple GET operations
- `set`: Simple SET operations
- `set_get_random`: Mixed SET/GET operations with configurable read ratio
- `delete`: DEL operations
- `incr`: INCR operations (increment counters)
- `decr`: DECR operations (decrement counters)
- `lpush`: LPUSH operations (push to left of list)
- `rpush`: RPUSH operations (push to right of list)
- `lpop`: LPOP operations (pop from left of list)
- `rpop`: RPOP operations (pop from right of list)
- `sadd`: SADD operations (add to set)
- `smembers`: SMEMBERS operations (get all members of set)
- `srem`: SREM operations (remove from set)
- `sismember`: SISMEMBER operations (check set membership)
- `zadd`: ZADD operations (add to sorted set)
- `zrange`: ZRANGE operations (get range from sorted set)
- `zrem`: ZREM operations (remove from sorted set)
- `zrank`: ZRANK operations (get rank in sorted set)
- `hset`: HSET operations (set hash field)
- `hget`: HGET operations (get hash field)
- `hmset`: HMSET operations (set multiple hash fields)
- `hmget`: HMGET operations (get multiple hash fields)
- `hgetall`: HGETALL operations (get all hash fields)
- `pub`: PUBLISH operations (publish to channel)
- `sub`: SUBSCRIBE operations (subscribe to channel)

### HTTP Commands

```bash
# Basic HTTP GET test
./abc-runner http --url http://localhost:8080 -n 10000 -c 50

# HTTP POST with body
./abc-runner http --url http://api.example.com/users \n  --method POST --body '{"name":"test"}' \n  --content-type application/json -n 1000 -c 20

# Duration-based test
./abc-runner http --url http://localhost:8080 --duration 60s -c 100

# Custom headers
./abc-runner http --url http://api.example.com \n  --header "Authorization:Bearer token123" \n  --header "X-API-Key:secret" -n 1000

# Using configuration file with core configuration
./abc-runner http --config config/http.yaml --core-config config/core.yaml
```

### Kafka Commands

```bash
# Basic producer test
./abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# Consumer test
./abc-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id my-group -n 1000

# Mixed produce/consume test
./abc-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic high-throughput --test-type produce_consume \n  --message-size 4096 --duration 60s -c 8

# High-performance test with compression
./abc-runner kafka --broker localhost:9092 --topic perf-test \n  --compression lz4 --acks all --batch-size 32768 -n 50000

# Using configuration file with core configuration
./abc-runner kafka --config config/kafka.yaml --core-config config/core.yaml
```

## Configuration Files

You can use YAML configuration files for complex setups:

### Core Configuration (config/core.yaml)

The core configuration file contains common settings shared across all protocols:

```yaml
core:
  # Logging configuration
  logging:
    level: "info"              # Log level: debug, info, warn, error
    format: "json"             # Log format: json, text
    output: "stdout"           # Output target: stdout, file, or file path
    file_path: "./logs"        # Log file directory
    max_size: "100MB"          # Maximum size of a single log file
    max_age: 7                 # Number of days to retain log files
    max_backups: 5             # Maximum number of backup files
    compress: true             # Whether to compress old log files

  # Report configuration
  reports:
    enabled: true              # Whether to enable reports
    formats: ["console"]       # Report formats: console, json, csv, text, all
    output_dir: "./reports"    # Report output directory
    file_prefix: "benchmark"   # Report file prefix
    include_timestamp: true    # Include timestamp in filename
    enable_console_report: true # Enable detailed console report
    overwrite_existing: false  # Whether to overwrite existing files

  # Monitoring configuration
  monitoring:
    enabled: true              # Whether to enable monitoring
    metrics_interval: "5s"     # Metrics collection interval
    prometheus:
      enabled: false           # Whether to enable Prometheus export
      port: 9090               # Prometheus export port
    statsd:
      enabled: false           # Whether to enable StatsD export
      host: "localhost:8125"   # StatsD server address

  # Global connection configuration
  connection:
    timeout: "30s"             # Default connection timeout
    keep_alive: "30s"          # Connection keep-alive time
    max_idle_conns: 100        # Maximum number of idle connections
    idle_conn_timeout: "90s"   # Idle connection timeout
```

### Redis Configuration (config/redis.yaml)

```yaml
redis:
  mode: "standalone"    # Options: standalone, sentinel, cluster
  benchmark:
    total: 10000              # 10000 requests default
    parallels: 50             # 2 parallel default
    random_keys: 50           # 0:Incremental key, >0:random key range is [0, r]
    read_percent: 50          # 50% read and 50 write default
    data_size: 3              # 3 bytes default
    ttl: 120                  # 120 seconds default
    case: "set_get_random"    # operations: set_get_random, set, get, del, pub, sub
  pool:
    pool_size: 10
    min_idle: 2
  standalone:
    addr: 127.0.0.1:6379
    password: "pwd@redis"
    db: 0
  sentinel:
    master_name: "mymaster"
    addrs:
      - "127.0.0.1:26371"
      - "127.0.0.1:26372"
      - "127.0.0.1:26373"
    password: "pwd@redis"
    db: 0
  cluster:
    addrs:
      - "127.0.0.1:6371"
      - "127.0.0.1:6372"
      - "127.0.0.1:6373"
    password: "pwd@redis"
```

### HTTP Configuration (config/http.yaml)

```yaml
http:
  connection:
    base_url: "http://localhost:8080"
    timeout: 30s
    keep_alive: 90s
    max_idle_conns: 50
    max_conns_per_host: 20
    idle_conn_timeout: 90s
    disable_compression: false
  benchmark:
    total: 100000
    parallels: 50
    duration: "5m"
    ramp_up: "30s"
    data_size: 1024
    ttl: 0s
    read_percent: 70
    random_keys: 0
    test_case: "mixed_operations"
    timeout: 30s
  requests:
    - method: "GET"
      path: "/api/users"
      headers:
        Accept: "application/json"
      weight: 100
```

### Kafka Configuration (config/kafka.yaml)

```yaml
kafka:
  brokers:
    - "localhost:9092"
  client_id: "abc-runner-kafka-client"
  version: "2.8.0"
  producer:
    acks: "all"
    retries: 3
    batch_size: 16384
    linger_ms: "5ms"
    compression: "snappy"
    idempotence: true
    max_in_flight: 5
    request_timeout: "30s"
    write_timeout: "10s"
    read_timeout: "10s"
  benchmark:
    default_topic: "benchmark-topic"
    message_size_range:
      min: 100
      max: 10240
    batch_sizes: [1, 10, 100, 1000]
    partition_strategy: "round_robin"
    total: 100000
    parallels: 50
    data_size: 1024
    ttl: 0
    read_percent: 50
    random_keys: 10000
    test_case: "produce"
    timeout: "30s"
```

## Documentation

For detailed documentation, please see the following resources:

- [Architecture Overview](docs/en/architecture/overview.md) - System architecture and design principles | [架构概述](docs/zh/architecture/overview.md)
- [Component Documentation](docs/en/architecture/components.md) - Detailed component documentation | [组件详解](docs/zh/architecture/components.md)
- [Quick Start Guide](docs/en/getting-started/quickstart.md) - Getting started quickly | [快速开始](docs/zh/getting-started/quickstart.md)
- [Redis Testing Guide](docs/en/user-guide/redis.md) - Redis-specific features and usage | [Redis测试指南](docs/zh/user-guide/redis.md)
- [HTTP Testing Guide](docs/en/user-guide/http.md) - HTTP-specific features and usage | [HTTP测试指南](docs/zh/user-guide/http.md)
- [Kafka Testing Guide](docs/en/user-guide/kafka.md) - Kafka-specific features and usage | [Kafka测试指南](docs/zh/user-guide/kafka.md)
- [Contributing Guide](docs/en/developer-guide/contributing.md) - Guidelines for contributing | [贡献指南](docs/zh/developer-guide/contributing.md)
- [Extending abc-runner](docs/en/developer-guide/extending.md) - How to extend the tool | [扩展abc-runner](docs/zh/developer-guide/extending.md)

## Packaging and Distribution

### Release Packages

Pre-built release packages are available for download from the [releases page](https://github.com/your-org/abc-runner/releases). Each release includes:

- Platform-specific binaries for macOS, Linux, and Windows
- Configuration file templates
- Documentation and license files

### Building from Source

```bash
# Clone the repository
git clone https://github.com/your-org/abc-runner.git
cd abc-runner

# Build for current platform
make build

# Build for all supported platforms
make build-all

# Create release packages
make release

# Create release packages with specific version
VERSION=1.0.0 make release
```

For detailed information about the packaging process, see [Packaging Guide](docs/packaging-guide.md).

## Migration from v0.0.x

This version introduces breaking changes. Key changes:

- `redis-enhanced` → `redis`
- `http-enhanced` → `http`  
- `kafka-enhanced` → `kafka`
- Simplified command structure
- Unified configuration format

See the [Migration Guide](docs/CHANGELOG.md) for detailed upgrade instructions.

## Examples

### Redis Performance Testing

```bash
# Basic performance test
./abc-runner redis -h 127.0.0.1 -p 6379 -n 100000 -c 50

# Cluster mode with authentication
./abc-runner redis --mode cluster -h localhost -p 6371 \n  -a "password" -n 100000 -c 10 -d 64 --read-ratio 50

# Custom test patterns
./abc-runner redis -t incr -n 50000 -c 100  # Counter operations
./abc-runner redis -t lpush_lpop -n 10000 -c 50  # List operations
```

### HTTP Load Testing

```bash
# API endpoint testing
./abc-runner http --url http://api.example.com/health -n 10000 -c 100

# POST with JSON payload
./abc-runner http --url http://api.example.com/users \n  --method POST \n  --body '{"name":"John","email":"john@example.com"}' \n  --content-type "application/json" -n 1000 -c 20

# Load testing with ramp-up
./abc-runner http --url http://www.example.com \n  --duration 300s -c 200 --ramp-up 30s
```

### Kafka Performance Testing

```bash
# Producer throughput test
./abc-runner kafka --broker localhost:9092 --topic throughput-test \n  --message-size 1024 -n 100000 -c 10

# Consumer lag test
./abc-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id perf-test-group -n 50000

# End-to-end latency test
./abc-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic latency-test --test-type produce_consume \n  --message-size 512 --duration 120s -c 5
```

## License

[Apache License 2.0](LICENSE)

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## Support

For questions and support:

- Check the [Migration Guide](docs/CHANGELOG.md)
- Review command help: `./abc-runner <command> --help`
- Open an issue for bug reports or feature requests

## Documentation

This project maintains documentation in both English and Chinese. For guidelines on maintaining multilingual documentation, see the [Document Translation Guide](docs/maintenance/document-translation-guide.md).