# Docker Deployment Guide

[English](docker.md) | [中文](../zh/deployment/docker.md)

## Docker Images

redis-runner provides official Docker images for quick deployment and testing.

## Pulling Images

```bash
# Pull the latest version
docker pull redis-runner/redis-runner:latest

# Pull a specific version
docker pull redis-runner/redis-runner:v0.2.0
```

## Basic Usage

### Running Redis Tests

```bash
# Run basic Redis tests
docker run --rm redis-runner/redis-runner redis -h host.docker.internal -p 6379 -n 1000 -c 10

# Use custom configuration file
docker run --rm -v $(pwd)/config:/config redis-runner/redis-runner redis --config /config/redis.yaml
```

### Running HTTP Tests

```bash
# Run HTTP tests
docker run --rm redis-runner/redis-runner http --url http://host.docker.internal:8080 -n 1000 -c 10
```

### Running Kafka Tests

```bash
# Run Kafka tests
docker run --rm redis-runner/redis-runner kafka --broker host.docker.internal:9092 --topic test -n 1000 -c 5
```

## Docker Compose

Create a `docker-compose.yml` file to orchestrate the test environment:

```yaml
version: '3.8'

services:
  redis-runner:
    image: redis-runner/redis-runner:latest
    depends_on:
      - redis
      - kafka
    volumes:
      - ./config:/config
      - ./reports:/reports
    command: redis --config /config/redis.yaml

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  zookeeper:
    image: bitnami/zookeeper:latest
    ports:
      - "2181:2181"
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes

  kafka:
    image: bitnami/kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
      - ALLOW_PLAINTEXT_LISTENER=yes
```

### Running the Test Environment

```bash
# Start all services
docker-compose up -d

# Run Redis tests
docker-compose run --rm redis-runner redis -h redis -p 6379 -n 1000 -c 10

# Run Kafka tests
docker-compose run --rm redis-runner kafka --broker kafka:9092 --topic test -n 1000 -c 5

# Stop all services
docker-compose down
```

## Custom Images

### Building Custom Images

Create a `Dockerfile`:

```dockerfile
FROM redis-runner/redis-runner:latest

# Copy custom configuration
COPY config/ /config/

# Set working directory
WORKDIR /app

# Set default command
ENTRYPOINT ["redis-runner"]
CMD ["--help"]
```

Build the image:

```bash
docker build -t my-redis-runner .
```

### Multi-stage Build

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

# Install dependencies
RUN go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o redis-runner .

# Run stage
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from build stage
COPY --from=builder /app/redis-runner .

# Copy configuration files
COPY config/ /config/

# Set port
EXPOSE 8080

# Run application
CMD ["./redis-runner"]
```

## Persistent Data

### Report Data

```bash
# Mount report directory
docker run --rm -v $(pwd)/reports:/reports redis-runner/redis-runner redis -h redis -p 6379
```

### Configuration Files

```bash
# Mount configuration directory
docker run --rm -v $(pwd)/config:/config redis-runner/redis-runner redis --config /config/redis.yaml
```

## Environment Variables

The following environment variables are supported:

```bash
docker run --rm \
  -e REDIS_HOST=redis \
  -e REDIS_PORT=6379 \
  -e KAFKA_BROKERS=kafka:9092 \
  redis-runner/redis-runner redis -n 1000 -c 10
```

## Network Configuration

### Using Custom Networks

```bash
# Create custom network
docker network create redis-runner-network

# Run Redis
docker run -d --name redis --network redis-runner-network redis:7-alpine

# Run redis-runner
docker run --rm --network redis-runner-network redis-runner/redis-runner redis -h redis -p 6379 -n 1000 -c 10
```

## Security Considerations

### User Permissions

```dockerfile
# Create non-root user
RUN addgroup -g 1001 -S runner &&\
    adduser -u 1001 -S runner -G runner

# Switch to non-root user
USER runner
```

### Read-only File System

```bash
# Use read-only file system (mount necessary directories)
docker run --rm --read-only \
  -v $(pwd)/reports:/reports \
  -v $(pwd)/config:/config \
  redis-runner/redis-runner redis -h redis -p 6379
```

## Monitoring and Logging

### Log Drivers

```bash
# Use json-file log driver
docker run --rm --log-driver json-file --log-opt max-size=10m redis-runner/redis-runner redis -h redis -p 6379

# Use syslog log driver
docker run --rm --log-driver syslog --log-opt syslog-address=tcp://localhost:514 redis-runner/redis-runner redis -h redis -p 6379
```

### Health Checks

```dockerfile
# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./redis-runner --version || exit 1
```

## Best Practices

1. **Use Tags**: Always use specific version tags instead of latest
2. **Minimize Images**: Use alpine base images to reduce image size
3. **Security**: Run containers as non-root users
4. **Configuration**: Manage configuration files through volume mounts
5. **Persistence**: Mount volumes to persist report data
6. **Networking**: Use custom networks to isolate services
7. **Resource Limits**: Set memory and CPU limits
8. **Monitoring**: Configure appropriate logging and monitoring