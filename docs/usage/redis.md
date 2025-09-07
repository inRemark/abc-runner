# Redis Testing Guide

This guide covers Redis-specific features and usage patterns for redis-runner.

## Redis Connection Modes

redis-runner supports three Redis deployment modes:

### 1. Standalone Mode (Default)

For single Redis instances:

```bash
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50
```

### 2. Cluster Mode

For Redis cluster deployments:

```bash
./redis-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50
```

### 3. Sentinel Mode

For Redis sentinel-managed instances:

```bash
./redis-runner redis --mode sentinel -h localhost -p 26379 -n 10000 -c 50
```

## Authentication

To test Redis instances with authentication:

```bash
./redis-runner redis -h localhost -p 6379 -a password -n 10000 -c 50
```

## Test Cases

redis-runner supports multiple Redis test cases:

### set_get_random

Mixed SET and GET operations with random keys:

```bash
./redis-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80
```

### set_only

SET operations only:

```bash
./redis-runner redis -t set_only -n 100000 -c 100
```

### get_only

GET operations only (requires pre-existing keys):

```bash
./redis-runner redis -t get_only -n 100000 -c 100
```

### incr

INCR operations on counter keys:

```bash
./redis-runner redis -t incr -n 50000 -c 100
```

### append

APPEND operations on string keys:

```bash
./redis-runner redis -t append -n 50000 -c 100
```

### lpush_lpop

LPUSH and LPOP operations on list keys:

```bash
./redis-runner redis -t lpush_lpop -n 10000 -c 50
```

## Key Generation Strategies

### Global Self-Increasing Keys

When `-r 0` (default), keys are globally self-increasing:

```bash
./redis-runner redis -n 100000 -c 100 -r 0
```

### Random Keys

When `-r > 0`, keys are randomly generated:

```bash
./redis-runner redis -n 100000 -c 100 -r 1000
```

## TTL Configuration

Set expiration time for keys:

```bash
./redis-runner redis -n 100000 -c 100 --ttl 300s
```

## Configuration File Example

```yaml
# redis.yaml
protocol: redis
connection:
  host: localhost
  port: 6379
  mode: standalone  # standalone, cluster, sentinel
  password: ""      # Optional password
  timeout: 30s

benchmark:
  total: 100000
  parallels: 50
  test_case: "set_get_random"
  data_size: 64
  read_ratio: 0.5
  key_range: 0        # 0 for self-increasing, >0 for random
  ttl: 0s             # 0 for no expiration, >0s for expiration
```

Run with configuration:

```bash
./redis-runner redis --config redis.yaml
```

## Performance Tuning

### Connection Pooling

Adjust the number of parallel connections based on your Redis server capacity:

```bash
./redis-runner redis -n 100000 -c 100  # 100 parallel connections
```

### Data Size

Control the size of data used in SET operations:

```bash
./redis-runner redis -n 100000 -c 50 -d 1024  # 1KB values
```

### Read Ratio

For mixed workloads, control the ratio of read to write operations:

```bash
./redis-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80  # 80% reads
```
