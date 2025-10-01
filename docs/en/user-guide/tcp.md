# TCP Performance Testing Guide

[English](tcp.md) | [中文](../zh/user-guide/tcp.md)

## Overview

The TCP adapter in abc-runner provides comprehensive performance testing capabilities for TCP connections, supporting custom payloads, connection patterns, and network performance analysis.

## Quick Start

### Basic TCP Test

```bash
# Simple TCP performance test
./abc-runner tcp --host localhost --port 8080 -n 1000 -c 10

# Test with custom payload
./abc-runner tcp --host localhost --port 8080 --payload "Hello TCP" -n 1000

# Test with configuration file
./abc-runner tcp --config config/tcp.yaml
```

## Configuration

### Connection Configuration

```yaml
tcp:
  host: "localhost"
  port: 8080
  timeout: "30s"
  connect_timeout: "10s"
  read_timeout: "30s"
  write_timeout: "30s"
  
  # Socket options
  socket:
    keep_alive: true
    keep_alive_period: "30s"
    no_delay: true
    buffer_size: 4096
    
  # TLS Configuration (for secure connections)
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
    insecure_skip_verify: false
```

### Payload Configuration

```yaml
payloads:
  - type: "text"
    content: "Hello TCP Server"
    weight: 50

  - type: "binary"
    content_hex: "48656c6c6f20544350"  # "Hello TCP" in hex
    weight: 30

  - type: "random" 
    size: 1024
    weight: 20
```

### Test Patterns

```yaml
test_patterns:
  - name: "echo_test"
    send_data: "PING"
    expect_response: true
    response_size: 4
    weight: 60

  - name: "throughput_test"
    send_data: "DATA:${random_1024}"
    expect_response: false
    weight: 40
```

## Command Line Options

### Basic Options

```bash
./abc-runner tcp [OPTIONS]

Options:
  --host string           TCP server host (default "localhost")
  --port int              TCP server port (default 8080)
  --payload string        Data payload to send
  --payload-size int      Random payload size in bytes
  --timeout duration      Connection timeout (default 30s)
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
  --keep-alive            Enable TCP keep-alive
  --no-delay              Enable TCP_NODELAY
  --buffer-size int       Socket buffer size (default 4096)
```

## Test Scenarios

### 1. Echo Server Testing

Test basic echo functionality:

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload "ECHO:Hello World" \
  -n 10000 -c 50
```

### 2. Throughput Testing

Test network throughput with large payloads:

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload-size 8192 \
  -n 5000 -c 25
```

### 3. Connection Stress Testing

Test connection establishment and teardown:

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload "CONNECT" \
  --duration 60s \
  -c 100
```

### 4. Protocol Testing

Test custom protocol implementations:

```bash
./abc-runner tcp \
  --host localhost \
  --port 8080 \
  --payload "GET /status HTTP/1.1\r\nHost: localhost\r\n\r\n" \
  -n 1000 -c 20
```

### 5. Secure TCP Testing

Test with TLS encryption:

```bash
./abc-runner tcp \
  --host localhost \
  --port 8443 \
  --tls \
  --cert client.crt \
  --key client.key \
  -n 1000 -c 10
```

## Payload Types

### Text Payloads

```yaml
payloads:
  - type: "text"
    content: "Hello TCP Server"
```

### Binary Payloads

```yaml
payloads:
  - type: "binary"
    content_hex: "deadbeef"
```

### Random Payloads

```yaml
payloads:
  - type: "random"
    size: 1024
    pattern: "alphanumeric"  # alphanumeric, binary, hex
```

### Template Payloads

```yaml
payloads:
  - type: "template"
    content: "ID:${sequence},DATA:${random_string_32}"
```

## Connection Patterns

### Single Connection Pattern

```yaml
connection_pattern:
  type: "single"
  reuse_connection: true
  max_requests_per_connection: 1000
```

### Connection Pool Pattern

```yaml
connection_pattern:
  type: "pool"
  pool_size: 50
  max_idle: 10
  idle_timeout: "300s"
```

### Per-Request Connection Pattern

```yaml
connection_pattern:
  type: "per_request"
  connect_timeout: "5s"
  close_after_request: true
```

## Monitoring and Metrics

### Built-in Metrics

The TCP adapter collects the following metrics:

- **Connection Metrics**: Connection success rate, connection time
- **Throughput Metrics**: Bytes sent/received per second
- **Latency Metrics**: Round-trip time, connect time
- **Error Metrics**: Connection errors, timeout errors
- **Network Metrics**: Packet loss, retransmissions

### Custom Metrics

```yaml
metrics:
  collection_interval: "1s"
  custom_metrics:
    - name: "tcp_connections_active"
      type: "gauge"
    - name: "tcp_bytes_total"
      type: "counter"
```

## Network Analysis

### Latency Analysis

```yaml
latency_analysis:
  enabled: true
  percentiles: [50, 90, 95, 99]
  histogram_buckets: [1, 5, 10, 25, 50, 100, 250, 500, 1000]
```

### Throughput Analysis

```yaml
throughput_analysis:
  enabled: true
  measurement_interval: "1s"
  bandwidth_limit: "100MB/s"
```

## Error Handling

### Common Error Types

1. **Connection Errors**: Network connectivity issues
2. **Timeout Errors**: Connection or operation timeouts
3. **Socket Errors**: Low-level socket errors
4. **Protocol Errors**: Application protocol errors
5. **Resource Errors**: System resource exhaustion

### Error Recovery

```yaml
error_handling:
  retry_on_connection_error: true
  max_retries: 3
  retry_delay: "1s"
  
  fail_fast_errors:
    - "connection_refused"
    - "host_unreachable"
```

## Socket Tuning

### Buffer Sizes

```yaml
socket_tuning:
  send_buffer_size: 65536
  recv_buffer_size: 65536
  write_buffer_size: 8192
  read_buffer_size: 8192
```

### TCP Options

```yaml
tcp_options:
  tcp_nodelay: true
  tcp_keepalive: true
  tcp_keepalive_time: 7200
  tcp_keepalive_interval: 75
  tcp_keepalive_probes: 9
```

## Best Practices

1. **Connection Management**: Use appropriate connection patterns
2. **Buffer Sizing**: Optimize buffer sizes for your use case
3. **Timeout Settings**: Set appropriate timeout values
4. **Error Handling**: Implement robust error recovery
5. **Resource Monitoring**: Monitor system resources
6. **Network Tuning**: Tune TCP parameters for performance
7. **Security**: Use TLS for sensitive data

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if TCP server is running and port is open
2. **Timeout Errors**: Increase timeout values or check network latency
3. **Socket Errors**: Check system limits and network configuration
4. **Performance Issues**: Optimize buffer sizes and connection patterns
5. **TLS Errors**: Verify certificate configurations

### Debug Mode

Enable debug mode for detailed logging:

```bash
./abc-runner tcp --host localhost --port 8080 --debug -n 100
```

### Network Diagnostics

```bash
# Check connectivity
telnet localhost 8080

# Check network statistics  
netstat -an | grep 8080

# Monitor network traffic
tcpdump -i any port 8080
```

## Examples

See the [configuration examples](../../config/examples/) directory for complete TCP configuration examples.