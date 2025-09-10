# Redis测试指南

[English](../en/user-guide/redis.md) | [中文](redis.md)

## 支持的Redis模式

abc-runner支持以下Redis部署模式：

- **单机模式 (Standalone)**: 标准的单实例Redis
- **哨兵模式 (Sentinel)**: Redis Sentinel高可用配置
- **集群模式 (Cluster)**: Redis Cluster分布式配置

## 测试用例

abc-runner支持多种Redis操作测试：

- `set_get_random`: 随机SET/GET操作
- `set_only`: 仅SET操作
- `get_only`: 仅GET操作
- `del`: 删除操作
- `incr`: 计数器递增操作
- `decr`: 计数器递减操作
- `lpush`: 列表左推操作
- `rpush`: 列表右推操作
- `lpop`: 列表左弹出操作
- `rpop`: 列表右弹出操作
- `sadd`: 集合添加操作
- `smembers`: 集合成员获取操作
- `srem`: 集合移除操作
- `sismember`: 集合成员检查操作
- `zadd`: 有序集合添加操作
- `zrange`: 有序集合范围获取操作
- `zrem`: 有序集合移除操作
- `zrank`: 有序集合排名获取操作
- `hset`: 哈希设置操作
- `hget`: 哈希获取操作
- `hmset`: 哈希多字段设置操作
- `hmget`: 哈希多字段获取操作
- `hgetall`: 哈希获取所有字段操作
- `pub`: 发布操作
- `sub`: 订阅操作

## 配置选项

### 命令行选项

```bash
# 基本连接选项
-h <hostname>         Redis服务器主机名 (默认: 127.0.0.1)
-p <port>             Redis服务器端口 (默认: 6379)
-a <password>         Redis服务器密码
--mode <mode>         Redis模式: standalone/sentinel/cluster (默认: standalone)

# 基准测试选项
-n <requests>         总请求数 (默认: 1000)
-c <connections>      并发连接数 (默认: 10)
-t <test>             测试用例 (默认: set_get_random)
-d <size>             数据大小(字节) (默认: 64)
--duration <time>     测试持续时间 (例如: 30s, 5m) - 覆盖-n
--read-ratio <ratio>  读/写比例 (0-100, 默认: 50)
```

### 配置文件选项

在配置文件中，您可以指定更详细的选项：

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

## 使用示例

### 基本性能测试

```bash
# 简单的SET/GET测试
./abc-runner redis -h localhost -p 6379 -n 10000 -c 50

# 持续时间测试
./abc-runner redis -h localhost -p 6379 --duration 60s -c 100
```

### 集群模式测试

```bash
# 使用命令行参数
./abc-runner redis --mode cluster -h localhost -p 6371 -n 10000 -c 10

# 使用配置文件
./abc-runner redis --config config/examples/redis-cluster.yaml
```

### 自定义测试用例

```bash
# 计数器操作测试
./abc-runner redis -t incr -n 10000 -c 50 -d 10

# 列表操作测试
./abc-runner redis -t lpush_lpop -n 10000 -c 50 -d 64

# 集合操作测试
./abc-runner redis -t sadd_smembers -n 10000 -c 50 -d 32
```

### 哨兵模式测试

```bash
# 使用配置文件进行哨兵模式测试
./abc-runner redis --config config/examples/redis-sentinel.yaml
```

## 结果解读

测试完成后，abc-runner会输出详细的性能报告：

- **RPS**: 每秒请求数
- **成功率**: 成功请求的百分比
- **总操作数**: 执行的总操作数
- **读/写操作数**: 分别的读写操作数
- **平均延迟**: 平均响应时间
- **P90/P95/P99延迟**: 90%/95%/99%请求的响应时间
- **最大延迟**: 最大响应时间

## 最佳实践

1. **预热**: 在正式测试前运行短时间的预热测试
2. **资源监控**: 监控Redis服务器的CPU和内存使用情况
3. **网络**: 确保网络延迟不会影响测试结果
4. **数据大小**: 根据实际使用场景选择合适的数据大小
5. **并发连接**: 根据服务器性能调整并发连接数