# UDP Performance Testing Guide

[English](udp.md) | [中文](../zh/user-guide/udp.md)

## Overview

The UDP adapter in abc-runner provides comprehensive performance testing capabilities for UDP connections, supporting datagram transmission testing, packet loss analysis, and network reliability assessment.

## Quick Start

### Basic UDP Test

```bash
# Simple UDP performance test
./abc-runner udp --host localhost --port 8080 -n 1000 -c 10

# Test with custom payload
./abc-runner udp --host localhost --port 8080 --payload "Hello UDP" -n 1000

# Test with configuration file
./abc-runner udp --config config/udp.yaml
```

## Configuration

### Connection Configuration

```yaml
udp:
  host: "localhost"
  port: 8080
  timeout: "30s"
  read_timeout: "5s"
  write_timeout: "5s"
  
  # Socket options
  socket:
    buffer_size: 65536
    read_buffer_size: 65536
    write_buffer_size: 65536
    
  # Packet options
  packet:
    max_size: 1472  # MTU - IP header - UDP header
    fragment_threshold: 1400
```

### Payload Configuration

```yaml
payloads:
  - type: "text"
    content: "Hello UDP Server"
    weight: 50

  - type: "binary"
    content_hex: "48656c6c6f20555350"  # "Hello UDP" in hex
    weight: 30

  - type: "random"
    size: 512
    weight: 20
```

### Test Patterns

```yaml
test_patterns:
  - name: "ping_test"
    send_data: "PING:${timestamp}"
    expect_response: true
    response_timeout: "1s"
    weight: 60

  - name: "broadcast_test"
    send_data: "BROADCAST:${random_string}"
    expect_response: false
    weight: 40
```

## Command Line Options

### Basic Options

```bash
./abc-runner udp [OPTIONS]

Options:
  --host string           UDP server host (default "localhost")
  --port int              UDP server port (default 8080)
  --payload string        Data payload to send
  --payload-size int      Random payload size in bytes
  --timeout duration      Operation timeout (default 30s)
  -n, --requests int      Total number of packets (default 1000)
  -c, --connections int   Number of concurrent senders (default 10)
  --config string         Configuration file path
```

### Advanced Options

```bash
  --packet-size int       Maximum packet size (default 1472)
  --buffer-size int       Socket buffer size (default 65536)
  --no-fragment           Don't allow packet fragmentation
  --broadcast             Enable broadcast mode
  --multicast string      Multicast group address
  --ttl int               Time-to-live for packets (default 64)
```

## Test Scenarios

### 1. Echo Server Testing

Test basic UDP echo functionality:

```bash
./abc-runner udp \
  --host localhost \
  --port 8080 \
  --payload "ECHO:Hello World" \
  -n 10000 -c 50
```

### 2. Packet Loss Testing

Test packet loss with high-frequency sending:

```bash
./abc-runner udp \
  --host localhost \
  --port 8080 \
  --payload-size 1024 \
  --duration 60s \
  -c 100
```

### 3. Throughput Testing

Test UDP throughput with large packets:

```bash
./abc-runner udp \
  --host localhost \
  --port 8080 \
  --packet-size 1472 \
  --payload-size 1400 \
  -n 5000 -c 25
```

### 4. Broadcast Testing

Test UDP broadcast functionality:

```bash
./abc-runner udp \
  --host 192.168.1.255 \
  --port 8080 \
  --broadcast \
  --payload "BROADCAST:Hello Network" \
  -n 1000 -c 5
```

### 5. Multicast Testing

Test UDP multicast functionality:

```bash
./abc-runner udp \
  --multicast 224.0.0.1 \
  --port 8080 \
  --payload "MULTICAST:Hello Group" \
  -n 1000 -c 10
```

## Payload Types

### Text Payloads

```yaml
payloads:
  - type: "text"
    content: "Hello UDP Server"
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
    size: 512
    pattern: "binary"  # binary, hex, alphanumeric
```

### Structured Payloads

```yaml
payloads:
  - type: "structured"
    format: "json"
    content: |
      {
        "id": "${sequence}",
        "timestamp": "${timestamp}",
        "data": "${random_string_32}"
      }
```

## Packet Configuration

### Size Optimization

```yaml
packet_config:
  # Optimize for network MTU
  max_packet_size: 1472  # Ethernet MTU (1500) - IP header (20) - UDP header (8)
  
  # Fragmentation settings
  allow_fragmentation: false
  fragment_threshold: 1400
  
  # Jumbo frames (if supported)
  jumbo_frames: false
  jumbo_packet_size: 9000
```

### Quality of Service

```yaml
qos:
  dscp: 0  # Differentiated Services Code Point
  tos: 0   # Type of Service
  priority: 0
```

## Monitoring and Metrics

### Built-in Metrics

The UDP adapter collects the following metrics:

- **Packet Metrics**: Packets sent/received, packet loss rate
- **Throughput Metrics**: Bytes sent/received per second
- **Latency Metrics**: Round-trip time (for echo tests)
- **Error Metrics**: Send errors, receive errors
- **Network Metrics**: Jitter, packet reordering

### Packet Loss Analysis

```yaml
packet_loss_analysis:
  enabled: true
  sequence_tracking: true
  duplicate_detection: true
  reorder_detection: true
  
  # Loss rate thresholds
  warning_threshold: 1.0   # 1% loss
  critical_threshold: 5.0  # 5% loss
```

### Jitter Analysis

```yaml
jitter_analysis:
  enabled: true
  measurement_window: "10s"
  percentiles: [50, 90, 95, 99]
```

## Network Configuration

### Broadcast Mode

```yaml
broadcast:
  enabled: true
  address: "192.168.1.255"
  interface: "eth0"  # Optional: specify interface
```

### Multicast Mode

```yaml
multicast:
  enabled: true
  group: "224.0.0.1"
  interface: "0.0.0.0"  # Listen on all interfaces
  ttl: 1  # Multicast TTL
```

### Socket Options

```yaml
socket_options:
  so_reuseaddr: true
  so_reuseport: true
  so_broadcast: true
  so_rcvbuf: 65536
  so_sndbuf: 65536
```

## Error Handling

### Common Error Types

1. **Network Errors**: Network unreachable, host unreachable
2. **Socket Errors**: Buffer overflow, socket creation errors
3. **Timeout Errors**: Send/receive timeouts
4. **Packet Errors**: Packet too large, fragmentation errors
5. **Configuration Errors**: Invalid addresses, port conflicts

### Error Recovery

```yaml
error_handling:
  retry_on_error: false  # UDP is fire-and-forget
  ignore_send_errors: false
  ignore_receive_errors: false
  
  # Specific error handling
  handle_icmp_errors: true
  handle_buffer_full: true
```

## Performance Tuning

### Buffer Optimization

```yaml
performance:
  # Increase buffer sizes for high throughput
  socket_buffer_size: 1048576  # 1MB
  
  # Batch processing
  batch_size: 100
  batch_timeout: "1ms"
  
  # CPU optimization
  cpu_affinity: [0, 1, 2, 3]
  use_kernel_bypass: false  # DPDK support (if available)
```

### Rate Limiting

```yaml
rate_limiting:
  enabled: true
  packets_per_second: 10000
  bytes_per_second: 10485760  # 10MB/s
  burst_size: 1000
```

## Best Practices

1. **Packet Size**: Keep packets under MTU size to avoid fragmentation
2. **Error Handling**: Don't rely on delivery guarantees
3. **Buffer Management**: Use appropriate buffer sizes
4. **Rate Control**: Implement rate limiting to avoid overwhelming networks
5. **Monitoring**: Monitor packet loss and jitter
6. **Network Tuning**: Optimize network stack parameters
7. **Testing Environment**: Test in realistic network conditions

## Troubleshooting

### Common Issues

1. **Packet Loss**: Check network capacity and buffer sizes
2. **High Jitter**: Investigate network congestion
3. **Send Errors**: Check socket buffer sizes and rate limits
4. **Receive Errors**: Verify server is listening and processing packets
5. **Firewall Issues**: Ensure UDP ports are open

### Debug Mode

Enable debug mode for detailed logging:

```bash
./abc-runner udp --host localhost --port 8080 --debug -n 100
```

### Network Diagnostics

```bash
# Check UDP connectivity
nc -u localhost 8080

# Monitor UDP traffic
tcpdump -i any udp port 8080

# Check socket statistics
ss -u -a | grep 8080

# Check packet loss
ping -c 100 hostname
```

## Reliability Testing

### Packet Loss Simulation

```yaml
reliability_testing:
  simulate_packet_loss: true
  loss_rate: 1.0  # 1% packet loss
  
  simulate_jitter: true
  jitter_range: "1ms-10ms"
  
  simulate_reordering: true
  reorder_rate: 0.1  # 0.1% packet reordering
```

### Stress Testing

```yaml
stress_testing:
  flood_test: true
  max_rate: 100000  # packets per second
  duration: "60s"
  
  buffer_overflow_test: true
  large_packet_test: true
  concurrent_senders: 100
```

## Examples

See the [configuration examples](../../config/examples/) directory for complete UDP configuration examples.