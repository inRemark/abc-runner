# Quick Start Guide

This guide will help you get started with redis-runner quickly.

## Installation

### Building from Source

```bash
# Clone the repository
git clone https://github.com/your-org/redis-runner.git
cd redis-runner

# Build the binary
go build -o redis-runner .

# Run the tool
./redis-runner --help
```

### Using Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/your-org/redis-runner/releases).

## Basic Usage

### Redis Performance Testing

```bash
# Basic Redis test
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50

# Redis with authentication
./redis-runner redis -h localhost -p 6379 -a password -n 10000 -c 50

# Redis cluster mode
./redis-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50
```

### HTTP Load Testing

```bash
# Basic HTTP GET test
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# HTTP POST with body
./redis-runner http --url http://api.example.com/users \
  --method POST --body '{"name":"test"}' \
  --content-type application/json -n 1000 -c 20
```

### Kafka Performance Testing

```bash
# Basic producer test
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# Consumer test
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000
```

## Using Configuration Files

You can use YAML configuration files for complex setups. See the [configuration documentation](configuration.md) for more details.

### Redis Configuration Example

```yaml
# redis.yaml
protocol: redis
connection:
  host: localhost
  port: 6379
  mode: standalone
  timeout: 30s

benchmark:
  total: 10000
  parallels: 50
  test_case: "set_get_random"
  data_size: 64
  read_ratio: 0.5
```

Run with configuration file:
```bash
./redis-runner redis --config redis.yaml
```

## Command Aliases

For quicker testing, you can use short aliases:

```bash
# Short aliases for quick testing
./redis-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./redis-runner h --url http://httpbin.org/get -n 100  # HTTP
./redis-runner k --broker localhost:9092 -n 100      # Kafka
```

## Viewing Help

```bash
# Show general help
./redis-runner --help

# Show help for specific protocol
./redis-runner redis --help
./redis-runner http --help
./redis-runner kafka --help
```