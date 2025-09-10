# Quick Start

[English](quickstart.md) | [中文](../zh/getting-started/quickstart.md)

## Installation

### Building from Source

```bash
# Clone repository
git clone https://github.com/your-org/abc-runner.git
cd abc-runner

# Build
make build

# Or build directly with Go
go build -o abc-runner .
```

### Using Pre-compiled Binaries

Download pre-compiled binaries suitable for your system from the [Releases page](https://github.com/your-org/abc-runner/releases).

## Basic Usage

### Redis Performance Testing

```bash
# Basic test
./abc-runner redis -h localhost -p 6379 -n 10000 -c 50

# Using configuration file
./abc-runner redis --config config/examples/redis.yaml
```

### HTTP Load Testing

```bash
# Basic test
./abc-runner http --url http://localhost:8080 -n 10000 -c 50

# Using configuration file
./abc-runner http --config config/examples/http.yaml
```

### Kafka Performance Testing

```bash
# Basic test
./abc-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5

# Using configuration file
./abc-runner kafka --config config/examples/kafka.yaml
```

## Using Aliases

```bash
# Use short aliases for quick testing
./abc-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./abc-runner h --url http://httpbin.org/get -n 100  # HTTP
./abc-runner k --broker localhost:9092 -n 100      # Kafka
```

## Viewing Help

```bash
# Show global help
./abc-runner --help

# Show specific command help
./abc-runner redis --help
./abc-runner http --help
./abc-runner kafka --help
```