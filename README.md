# abc-runner

[English](README.md) | [中文](README.zh.md)

## About

A unified performance testing tool for Redis, HTTP, and Kafka protocols.

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

## Building from Source

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

```bash
config/core.yaml
config.http.yaml
config/redis.yaml
config/kafka.yaml
config/tcp.yaml
config/udp.yaml
config/grpc.yaml
config/websocket.yaml
```

## Documentation

For detailed documentation, please see the following resources:

- [Architecture Overview](docs/en/architecture/overview.md) - System architecture and design principles | [Architecture Overview](docs/en/architecture/overview.md)
- [Component Documentation](docs/en/architecture/components.md) - Detailed component documentation | [Component Documentation](docs/en/architecture/components.md)
- [Quick Start Guide](docs/en/getting-started/quickstart.md) - Getting started quickly | [Quick Start Guide](docs/en/getting-started/quickstart.md)
- [Redis Testing Guide](docs/en/user-guide/redis.md) - Redis-specific features and usage | [Redis Testing Guide](docs/en/user-guide/redis.md)
- [HTTP Testing Guide](docs/en/user-guide/http.md) - HTTP-specific features and usage | [HTTP Testing Guide](docs/en/user-guide/http.md)
- [Kafka Testing Guide](docs/en/user-guide/kafka.md) - Kafka-specific features and usage | [Kafka Testing Guide](docs/en/user-guide/kafka.md)
- [Contributing Guide](docs/en/developer-guide/contributing.md) - Guidelines for contributing | [Contributing Guide](docs/en/developer-guide/contributing.md)
- [Extending abc-runner](docs/en/developer-guide/extending.md) - How to extend the tool | [Extending abc-runner](docs/en/developer-guide/extending.md)

## License

[Apache License 2.0](LICENSE)

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## Support

For questions and support:

- Check the [Migration Guide](docs/CHANGELOG.md)
- Review command help: `./abc-runner <command> --help`
- Open an issue for bug reports or feature requests
