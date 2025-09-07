# Redis Testing Guide

[English](redis.md) | [中文](../zh/user-guide/redis.md)

## Supported Redis Modes

redis-runner supports the following Redis deployment modes:

- **Standalone Mode**: Standard single-instance Redis
- **Sentinel Mode**: Redis Sentinel high availability configuration
- **Cluster Mode**: Redis Cluster distributed configuration

## Test Cases

redis-runner supports multiple Redis operation tests:

- `set_get_random`: Random SET/GET operations
- `set_only`: SET operations only
- `get_only`: GET operations only
- `del`: Delete operations
- `incr`: Counter increment operations
- `decr`: Counter decrement operations
- `lpush`: List left push operations
- `rpush`: List right push operations
- `lpop`: List left pop operations
- `rpop`: List right pop operations
- `sadd`: Set add operations
- `smembers`: Set members get operations
- `srem`: Set remove operations
- `sismember`: Set member check operations
- `zadd`: Sorted set add operations
- `zrange`: Sorted set range get operations
- `zrem`: Sorted set remove operations
- `zrank`: Sorted set rank get operations
- `hset`: Hash set operations
- `hget`: Hash get operations
- `hmset`: Hash multi-field set operations
- `hmget`: Hash multi-field get operations
- `hgetall`: Hash get all fields operations
- `pub`: Publish operations
- `sub`: Subscribe operations

## Configuration Options

### Command Line Options

```bash
# Basic connection options
-h <hostname>         Redis server hostname (default: 127.0.0.1)
-p <port>             Redis server port (default: 6379)
-a <password>         Redis server password
--mode <mode>         Redis mode: standalone/sentinel/cluster (default: standalone)

# Benchmark options
-n <requests>         Total requests (default: 1000)
-c <connections>      Concurrent connections (default: 10)
-t <test>             Test case (default: set_get_random)
-d <size>             Data size (bytes) (default: 64)
--duration <time>     Test duration (e.g., 30s, 5m) - overrides -n
--read-ratio <ratio>  Read/write ratio (0-100, default: 50)
```

### Configuration File Options

In configuration files, you can specify more detailed options:

```yaml
redis:
  mode: "standalone"
  benchmark:
    total: 10000
    parallels: 50
    random_keys: 50
    read_percent: 50
    data_size: 64
    ttl: 120
    case: "set_get_random"
  pool:
    pool_size: 10
    min_idle: 2
  standalone:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
```

## Usage Examples

### Basic Performance Testing

```bash
# Simple SET/GET test
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50

# Duration-based test
./redis-runner redis -h localhost -p 6379 --duration 60s -c 100
```

### Cluster Mode Testing

```bash
# Using command line arguments
./redis-runner redis --mode cluster -h localhost -p 6371 -n 10000 -c 10

# Using configuration file
./redis-runner redis --config config/examples/redis-cluster.yaml
```

### Custom Test Cases

```bash
# Counter operations test
./redis-runner redis -t incr -n 10000 -c 50 -d 10

# List operations test
./redis-runner redis -t lpush_lpop -n 10000 -c 50 -d 64

# Set operations test
./redis-runner redis -t sadd_smembers -n 10000 -c 50 -d 32
```

### Sentinel Mode Testing

```bash
# Sentinel mode testing using configuration file
./redis-runner redis --config config/examples/redis-sentinel.yaml
```

## Result Interpretation

After testing is completed, redis-runner will output detailed performance reports:

- **RPS**: Requests Per Second
- **Success Rate**: Percentage of successful requests
- **Total Operations**: Total operations executed
- **Read/Write Operations**: Separate read and write operation counts
- **Average Latency**: Average response time
- **P90/P95/P99 Latency**: Response time for 90%/95%/99% of requests
- **Maximum Latency**: Maximum response time

## Best Practices

1. **Warm-up**: Run short warm-up tests before formal testing
2. **Resource Monitoring**: Monitor Redis server CPU and memory usage
3. **Network**: Ensure network latency does not affect test results
4. **Data Size**: Choose appropriate data sizes based on actual usage scenarios
5. **Concurrent Connections**: Adjust concurrent connections based on server performance