# gRPC Performance Testing Guide

[English](grpc.md) | [中文](../zh/user-guide/grpc.md)

## Overview

The gRPC adapter in abc-runner provides comprehensive performance testing capabilities for gRPC services, supporting connection pooling, streaming operations, and custom metadata handling.

## Quick Start

### Basic gRPC Test

```bash
# Simple gRPC performance test
./abc-runner grpc --target localhost:9090 -n 1000 -c 10

# Test with custom method
./abc-runner grpc --target localhost:9090 --method "/example.Service/TestMethod" -n 1000

# Test with configuration file
./abc-runner grpc --config config/grpc.yaml
```

## Configuration

### Connection Configuration

```yaml
grpc:
  target: "localhost:9090"
  timeout: "30s"
  max_recv_msg_size: 4194304
  max_send_msg_size: 4194304
  enable_reflection: true
  
  # TLS Configuration
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
    insecure_skip_verify: false

  # Connection Pool
  pool:
    size: 10
    max_idle: 5
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
      x-custom-header: "value"
    weight: 100

  - method: "/example.Service/StreamMethod"
    stream_type: "client_stream"
    payload: |
      {
        "data": "streaming test"
      }
    weight: 50
```

## Command Line Options

### Basic Options

```bash
./abc-runner grpc [OPTIONS]

Options:
  --target string         gRPC server target (default "localhost:9090")
  --method string         gRPC method to call
  --payload string        Request payload (JSON format)
  --timeout duration      Request timeout (default 30s)
  -n, --requests int      Total number of requests (default 1000)
  -c, --connections int   Number of concurrent connections (default 10)
  --config string         Configuration file path
```

### Advanced Options

```bash
  --tls                   Enable TLS connection
  --cert string           Client certificate file
  --key string            Client private key file
  --ca string             CA certificate file
  --insecure              Skip TLS certificate verification
  --reflection            Enable server reflection
  --metadata strings      Custom metadata (key:value format)
```

## Test Scenarios

### 1. Unary RPC Testing

Test simple request-response RPC calls:

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/GetUser" \
  --payload '{"id": 123}' \
  -n 10000 -c 50
```

### 2. Server Streaming Testing

Test server-side streaming RPCs:

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/ListUsers" \
  --stream-type server_stream \
  -n 1000 -c 20
```

### 3. Client Streaming Testing

Test client-side streaming RPCs:

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/CreateUsers" \
  --stream-type client_stream \
  -n 500 -c 10
```

### 4. Bidirectional Streaming Testing

Test bidirectional streaming RPCs:

```bash  
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/ChatService" \
  --stream-type bidi_stream \
  -n 1000 -c 5
```

### 5. Load Testing with Authentication

Test with authentication metadata:

```bash
./abc-runner grpc \
  --target localhost:9090 \
  --method "/example.Service/SecureMethod" \
  --metadata "authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -n 5000 -c 25
```

## Monitoring and Metrics

### Built-in Metrics

The gRPC adapter collects the following metrics:

- **Response Time**: Request/response latency
- **Throughput**: Requests per second (RPS)
- **Error Rate**: Percentage of failed requests
- **Connection Metrics**: Active connections, connection errors
- **Stream Metrics**: Stream duration, message counts

### Custom Metrics

```yaml
metrics:
  custom_labels:
    service: "example-service"
    environment: "production"
  
  collection_interval: "1s"
  export_interval: "10s"
```

## Error Handling

### Common Error Types

1. **Connection Errors**: Network connectivity issues
2. **Authentication Errors**: Invalid credentials or expired tokens
3. **Method Errors**: Invalid method names or protobuf definitions
4. **Timeout Errors**: Request timeouts
5. **Resource Errors**: Server resource exhaustion

### Error Configuration

```yaml
error_handling:
  retry_attempts: 3
  retry_delay: "1s"
  ignore_errors:
    - "UNAVAILABLE"
    - "RESOURCE_EXHAUSTED"
```

## Best Practices

1. **Connection Management**: Use appropriate connection pool sizes
2. **TLS Configuration**: Enable TLS for production environments
3. **Payload Size**: Consider message size limits
4. **Streaming**: Use streaming for large data transfers
5. **Metadata**: Include necessary authentication and tracing metadata
6. **Error Handling**: Implement proper retry logic
7. **Monitoring**: Monitor key performance indicators

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if gRPC server is running
2. **Method Not Found**: Verify method names and service definitions
3. **Authentication Failed**: Check metadata and credentials
4. **Timeout Errors**: Increase timeout values or check server performance
5. **TLS Errors**: Verify certificate configurations

### Debug Mode

Enable debug mode for detailed logging:

```bash
./abc-runner grpc --target localhost:9090 --debug -n 100
```

## Examples

See the [configuration examples](../../config/examples/) directory for complete gRPC configuration examples.