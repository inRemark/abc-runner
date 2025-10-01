# Configuration Management Guide

[English](configuration.md) | [中文](../zh/user-guide/configuration.md)

## Configuration File Structure

abc-runner uses YAML format configuration files, supporting configurations for seven protocols: Redis, HTTP, Kafka, gRPC, WebSocket, TCP, and UDP.

```
config/
├── templates/           # Configuration template files
├── examples/            # Configuration example files
├── production/          # Production environment configuration
├── development/         # Development environment configuration
├── core.yaml            # Core configuration for all protocols
├── redis.yaml           # Redis-specific configuration
├── http.yaml            # HTTP-specific configuration
├── kafka.yaml           # Kafka-specific configuration
├── grpc.yaml            # gRPC-specific configuration
├── websocket.yaml       # WebSocket-specific configuration
├── tcp.yaml             # TCP-specific configuration
├── udp.yaml             # UDP-specific configuration
└── README.md            # Configuration documentation
```

## Configuration Priority

Configuration is loaded in the following priority order:

1. **Command Line Arguments**: Highest priority
2. **Environment Variables**: Medium priority
3. **Configuration Files**: Lowest priority

## Common Configuration Options

### Benchmark Configuration

All protocols support the following benchmark configuration:

```yaml
benchmark:
  total: 10000              # Total requests/messages
  parallels: 50             # Concurrent connections
  duration: "60s"           # Test duration
  data_size: 1024           # Data size (bytes)
  read_percent: 50          # Read operation percentage
  random_keys: 1000         # Random key range
  timeout: "30s"            # Timeout
```

### Report Configuration

```yaml
reports:
  enabled: true
  formats: ["console", "json", "csv"]
  output_dir: "./reports"
  file_prefix: "benchmark"
  include_timestamp: true
  enable_console_report: true
```

## Redis Configuration

### Connection Configuration

```yaml
redis:
  mode: "standalone"        # standalone, sentinel, cluster
  standalone:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
  sentinel:
    master_name: "mymaster"
    addrs:
      - "127.0.0.1:26371"
    password: ""
    db: 0
  cluster:
    addrs:
      - "127.0.0.1:6371"
    password: ""
```

### Connection Pool Configuration

```yaml
pool:
  pool_size: 10
  min_idle: 2
```

## HTTP Configuration

### Connection Configuration

```yaml
http:
  connection:
    base_url: "http://example.com"
    timeout: "30s"
    keep_alive: "90s"
    max_idle_conns: 50
    max_conns_per_host: 20
```

### Request Templates

```yaml
requests:
  - method: "GET"
    path: "/api/users"
    headers:
      Accept: "application/json"
    weight: 100
```

### Authentication Configuration

```yaml
auth:
  type: "bearer"
  token: "your-token"
```

## Kafka Configuration

### Connection Configuration

```yaml
kafka:
  brokers:
    - "localhost:9092"
  client_id: "abc-runner-client"
```

### Producer Configuration

```yaml
producer:
  acks: "all"
  batch_size: 16384
  compression: "snappy"
  linger_ms: "5ms"
```

### Consumer Configuration

```yaml
consumer:
  group_id: "abc-runner-group"
  auto_offset_reset: "latest"
```

## gRPC Configuration

### Connection Configuration

```yaml
grpc:
  target: "localhost:9090"
  timeout: "30s"
  max_recv_msg_size: 4194304
  max_send_msg_size: 4194304
  
  # TLS Configuration
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
```

### Request Configuration

```yaml
requests:
  - method: "/example.Service/UnaryMethod"
    payload: |
      {
        "message": "test",
        "value": 123
      }
    metadata:
      authorization: "Bearer token123"
    weight: 100
```

## WebSocket Configuration

### Connection Configuration

```yaml
websocket:
  url: "ws://localhost:8080/ws"
  timeout: "30s"
  handshake_timeout: "10s"
  
  # Headers for handshake
  headers:
    Origin: "http://localhost:8080"
    Authorization: "Bearer token123"
```

### Message Configuration

```yaml
messages:
  - type: "text"
    content: "Hello WebSocket"
    weight: 70

  - type: "binary"
    content_base64: "SGVsbG8gV2ViU29ja2V0"
    weight: 30
```

## TCP Configuration

### Connection Configuration

```yaml
tcp:
  host: "localhost"
  port: 8080
  timeout: "30s"
  connect_timeout: "10s"
  
  # Socket options
  socket:
    keep_alive: true
    no_delay: true
    buffer_size: 4096
```

### Payload Configuration

```yaml
payloads:
  - type: "text"
    content: "Hello TCP Server"
    weight: 50

  - type: "binary"
    content_hex: "48656c6c6f20544350"
    weight: 30
```

## UDP Configuration

### Connection Configuration

```yaml
udp:
  host: "localhost"
  port: 8080
  timeout: "30s"
  
  # Packet options
  packet:
    max_size: 1472
    fragment_threshold: 1400
```

### Payload Configuration

```yaml
payloads:
  - type: "text"
    content: "Hello UDP Server"
    weight: 50

  - type: "random"
    size: 512
    weight: 30
```

## Environment Variables

The following environment variables are supported:

- `ABC_RUNNER_CONFIG`: Configuration file path
- `ABC_RUNNER_LOG_LEVEL`: Log level
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port
- `REDIS_PASSWORD`: Redis password
- `KAFKA_BROKERS`: Kafka broker list
- `HTTP_BASE_URL`: HTTP base URL
- `GRPC_TARGET`: gRPC server target
- `WEBSOCKET_URL`: WebSocket URL
- `TCP_HOST`: TCP server host
- `TCP_PORT`: TCP server port
- `UDP_HOST`: UDP server host
- `UDP_PORT`: UDP server port

## Configuration Validation

### Command Line Validation

```bash
# Validate configuration file
./abc-runner redis --config config/redis.yaml --validate
```

### Example Configuration Validation

```bash
# Validate using example configuration
./abc-runner redis --config config/examples/redis.yaml --dry-run
```

## Best Practices

1. **Environment Separation**: Maintain separate configuration files for different environments
2. **Version Control**: Include configuration files in version control
3. **Sensitive Information**: Use environment variables for sensitive information (such as passwords)
4. **Templates**: Use template files as configuration starting points
5. **Documentation**: Add comments for custom configurations
6. **Testing**: Test configurations before using in production
7. **Backup**: Regularly back up production environment configurations