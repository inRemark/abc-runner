# WebSocket Performance Testing Guide

[English](websocket.md) | [中文](../zh/user-guide/websocket.md)

## Overview

The WebSocket adapter in abc-runner provides comprehensive performance testing capabilities for WebSocket connections, supporting real-time communication testing, custom protocols, and message patterns.

## Quick Start

### Basic WebSocket Test

```bash
# Simple WebSocket performance test
./abc-runner websocket --url ws://localhost:8080/ws -n 1000 -c 10

# Test with custom messages
./abc-runner websocket --url ws://localhost:8080/ws --message "Hello WebSocket" -n 1000

# Test with configuration file
./abc-runner websocket --config config/websocket.yaml
```

## Configuration

### Connection Configuration

```yaml
websocket:
  url: "ws://localhost:8080/ws"
  timeout: "30s"
  handshake_timeout: "10s"
  read_buffer_size: 4096
  write_buffer_size: 4096
  
  # TLS Configuration (for wss://)
  tls:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
    insecure_skip_verify: false

  # Headers for handshake
  headers:
    Origin: "http://localhost:8080"
    Sec-WebSocket-Protocol: "chat"
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
    weight: 20

  - type: "ping"
    weight: 10
```

### Test Patterns

```yaml
test_patterns:
  - name: "echo_test"
    send_message: "ping"
    expect_response: true
    response_timeout: "5s"
    weight: 50

  - name: "broadcast_test"
    send_message: "broadcast:Hello everyone"
    expect_response: false
    weight: 30

  - name: "subscribe_test"
    send_message: "subscribe:channel1"
    expect_response: true
    keep_listening: true
    weight: 20
```

## Command Line Options

### Basic Options

```bash
./abc-runner websocket [OPTIONS]

Options:
  --url string            WebSocket URL (ws:// or wss://)
  --message string        Message to send
  --message-type string   Message type: text, binary, ping, pong (default "text")
  --timeout duration      Connection timeout (default 30s)
  -n, --requests int      Total number of messages (default 1000)
  -c, --connections int   Number of concurrent connections (default 10)
  --config string         Configuration file path
```

### Advanced Options

```bash
  --tls                   Enable TLS (wss://)
  --cert string           Client certificate file
  --key string            Client private key file
  --ca string             CA certificate file
  --insecure              Skip TLS certificate verification
  --header strings        Custom headers (key:value format)
  --subprotocol strings   WebSocket subprotocols
  --origin string         Origin header value
```

## Test Scenarios

### 1. Echo Testing

Test basic message echo functionality:

```bash
./abc-runner websocket \
  --url ws://localhost:8080/echo \
  --message "Hello Echo" \
  -n 10000 -c 50
```

### 2. Chat Application Testing

Test chat application with multiple connections:

```bash
./abc-runner websocket \
  --url ws://localhost:8080/chat \
  --header "Authorization:Bearer token123" \
  --message "Hello Chat Room" \
  -n 5000 -c 100
```

### 3. Real-time Data Streaming

Test real-time data streaming:

```bash
./abc-runner websocket \
  --url ws://localhost:8080/stream \
  --subprotocol "data-stream" \
  --message '{"type":"subscribe","channel":"market-data"}' \
  -n 1000 -c 20
```

### 4. Binary Data Testing

Test binary message transmission:

```bash
./abc-runner websocket \
  --url ws://localhost:8080/binary \
  --message-type binary \
  --message "$(echo -n 'Binary Data' | base64)" \
  -n 2000 -c 25
```

### 5. Secure WebSocket Testing

Test with TLS encryption:

```bash
./abc-runner websocket \
  --url wss://localhost:8443/secure \
  --tls \
  --cert client.crt \
  --key client.key \
  -n 1000 -c 10
```

## Message Types

### Text Messages

```yaml
messages:
  - type: "text"
    content: "Plain text message"
```

### Binary Messages

```yaml
messages:
  - type: "binary"
    content_base64: "SGVsbG8gV29ybGQ="  # Base64 encoded
```

### JSON Messages

```yaml
messages:
  - type: "text"
    content: |
      {
        "type": "message",
        "data": "Hello JSON",
        "timestamp": "2025-01-02T10:00:00Z"
      }
```

### Control Messages

```yaml
messages:
  - type: "ping"
    content: "ping data"
  
  - type: "pong"
    content: "pong data"
```

## Monitoring and Metrics

### Built-in Metrics

The WebSocket adapter collects the following metrics:

- **Connection Metrics**: Connection success rate, handshake time
- **Message Metrics**: Messages sent/received per second
- **Latency Metrics**: Round-trip time for echo messages
- **Error Metrics**: Connection errors, message errors
- **Throughput Metrics**: Bytes sent/received per second

### Real-time Monitoring

```yaml
monitoring:
  enabled: true
  interval: "1s"
  metrics:
    - connection_count
    - message_rate
    - error_rate
    - latency_p95
```

## Connection Management

### Connection Lifecycle

```yaml
connection:
  # Connection establishment
  connect_timeout: "10s"
  handshake_timeout: "5s"
  
  # Keep-alive settings
  ping_interval: "30s"
  pong_timeout: "10s"
  
  # Reconnection settings
  auto_reconnect: true
  reconnect_delay: "5s"
  max_reconnect_attempts: 3
```

### Connection Pooling

```yaml
pool:
  max_connections: 100
  idle_timeout: "300s"
  connection_lifetime: "3600s"
```

## Error Handling

### Common Error Types

1. **Connection Errors**: Network connectivity issues
2. **Handshake Errors**: HTTP upgrade failures
3. **Protocol Errors**: WebSocket protocol violations
4. **Message Errors**: Invalid message format
5. **Timeout Errors**: Operation timeouts

### Error Recovery

```yaml
error_handling:
  retry_on_failure: true
  max_retries: 3
  retry_delay: "1s"
  
  ignore_errors:
    - "close_normal"
    - "close_going_away"
```

## Best Practices

1. **Connection Management**: Use appropriate connection limits
2. **Message Size**: Consider message size and buffer limits
3. **Error Handling**: Implement proper reconnection logic
4. **Security**: Use WSS for production environments
5. **Monitoring**: Monitor connection health and message rates
6. **Resource Management**: Clean up connections properly
7. **Protocol Compliance**: Follow WebSocket protocol standards

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if WebSocket server is running
2. **Handshake Failed**: Verify URL and headers
3. **Protocol Error**: Check message format and protocol compliance
4. **Timeout Errors**: Increase timeout values
5. **TLS Errors**: Verify certificate configurations

### Debug Mode

Enable debug mode for detailed logging:

```bash
./abc-runner websocket --url ws://localhost:8080/ws --debug -n 100
```

## Examples

See the [configuration examples](../../config/examples/) directory for complete WebSocket configuration examples.