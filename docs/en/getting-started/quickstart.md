# Quick Start

[English](quickstart.md) | [中文](../zh/getting-started/quickstart.md)

## Installation

### Building from Source

```bash
# Clone repository
git clone https://github.com/your-org/redis-runner.git
cd redis-runner

# Build
make build

# Or build directly with Go
go build -o redis-runner .
```

### Using Pre-compiled Binaries

Download pre-compiled binaries suitable for your system from the [Releases page](https://github.com/your-org/redis-runner/releases).

## Basic Usage

### Redis Performance Testing

```bash
# Basic test
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50

# Using configuration file
./redis-runner redis --config config/examples/redis.yaml
```

### HTTP Load Testing

```bash
# Basic test
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# Using configuration file
./redis-runner http --config config/examples/http.yaml
```

### Kafka Performance Testing

```bash
# Basic test
./redis-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5

# Using configuration file
./redis-runner kafka --config config/examples/kafka.yaml
```

## Using Aliases

```bash
# Use short aliases for quick testing
./redis-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./redis-runner h --url http://httpbin.org/get -n 100  # HTTP
./redis-runner k --broker localhost:9092 -n 100      # Kafka
```

## Viewing Help

```bash
# Show global help
./redis-runner --help

# Show specific command help
./redis-runner redis --help
./redis-runner http --help
./redis-runner kafka --help
```