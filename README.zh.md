# redis-runner

[English](README.md) | [中文](README.zh.md)

## 关于

一个用于Redis、HTTP和Kafka协议的统一性能测试工具。

⚠️ **重大变更通知**: 此版本 (v0.1.0) 引入了重大变更。请参阅[迁移指南](docs/CHANGELOG.md)了解升级说明。

## 特性

### Redis测试

- 支持Redis集群、哨兵和单机模式
- 多种测试用例：set_get_random, set_only, get_only, incr, decr, lpush, rpush, lpop, rpop, sadd, smembers, zadd, zrange, hset, hget, hmset, hmget, hgetall, pub, sub等
- 可配置的读写比例和TTL
- 全局自增键或随机键生成

### HTTP测试

- 支持GET、POST、PUT、DELETE方法
- 自定义头部和请求体
- 连接池和keep-alive
- 基于持续时间和请求数的测试

### Kafka测试

- 生产者和消费者性能测试
- 支持多个broker和主题
- 可配置的消息大小和压缩
- 混合生产和消费工作负载

## 快速开始

### 安装

```bash
# 从源码构建
go build -o redis-runner .

# 或从发布页面下载预构建的二进制文件
```

### 基本用法

```bash
# 显示帮助
./redis-runner --help

# Redis性能测试
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50

# HTTP负载测试
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# Kafka性能测试
./redis-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5
```

### 使用别名

```bash
# 快速测试的短别名
./redis-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./redis-runner h --url http://httpbin.org/get -n 100  # HTTP
./redis-runner k --broker localhost:9092 -n 100      # Kafka
```

## 命令参考

### 全局选项

```bash
./redis-runner --help                 # 显示帮助
./redis-runner --version              # 显示版本
```

### Redis命令

```bash
# 基本Redis测试
./redis-runner redis -h <host> -p <port> -n <requests> -c <connections>

# 带认证的Redis
./redis-runner redis -h localhost -p 6379 -a password -n 10000 -c 50

# Redis集群模式
./redis-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50

# 自定义测试用例和读取比例
./redis-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80

# 使用配置文件
./redis-runner redis --config conf/redis.yaml
```

支持的Redis测试用例 (`-t` 选项):
- `get`: 简单GET操作
- `set`: 简单SET操作
- `set_get_random`: 混合SET/GET操作，可配置读取比例
- `delete`: DEL操作
- `incr`: INCR操作 (递增计数器)
- `decr`: DECR操作 (递减计数器)
- `lpush`: LPUSH操作 (向列表左侧推入)
- `rpush`: RPUSH操作 (向列表右侧推入)
- `lpop`: LPOP操作 (从列表左侧弹出)
- `rpop`: RPOP操作 (从列表右侧弹出)
- `sadd`: SADD操作 (添加到集合)
- `smembers`: SMEMBERS操作 (获取集合所有成员)
- `srem`: SREM操作 (从集合移除)
- `sismember`: SISMEMBER操作 (检查集合成员)
- `zadd`: ZADD操作 (添加到有序集合)
- `zrange`: ZRANGE操作 (获取有序集合范围)
- `zrem`: ZREM操作 (从有序集合移除)
- `zrank`: ZRANK操作 (获取有序集合排名)
- `hset`: HSET操作 (设置哈希字段)
- `hget`: HGET操作 (获取哈希字段)
- `hmset`: HMSET操作 (设置多个哈希字段)
- `hmget`: HMGET操作 (获取多个哈希字段)
- `hgetall`: HGETALL操作 (获取所有哈希字段)
- `pub`: PUBLISH操作 (发布到频道)
- `sub`: SUBSCRIBE操作 (订阅频道)

### HTTP命令

```bash
# 基本HTTP GET测试
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# 带请求体的HTTP POST
./redis-runner http --url http://api.example.com/users \n  --method POST --body '{"name":"test"}' \n  --content-type application/json -n 1000 -c 20

# 基于持续时间的测试
./redis-runner http --url http://localhost:8080 --duration 60s -c 100

# 自定义头部
./redis-runner http --url http://api.example.com \n  --header "Authorization:Bearer token123" \n  --header "X-API-Key:secret" -n 1000
```

### Kafka命令

```bash
# 基本生产者测试
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# 消费者测试
./redis-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id my-group -n 1000

# 混合生产和消费测试
./redis-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic high-throughput --test-type produce_consume \n  --message-size 4096 --duration 60s -c 8

# 高性能测试与压缩
./redis-runner kafka --broker localhost:9092 --topic perf-test \n  --compression lz4 --acks all --batch-size 32768 -n 50000
```

## 配置文件

您可以使用YAML配置文件进行复杂设置：

### Redis配置 (conf/redis.yaml)

```yaml
protocol: redis
connection:
  host: localhost
  port: 6379
  mode: standalone  # standalone, cluster, sentinel
  timeout: 30s

benchmark:
  total: 10000
  parallels: 50
  test_case: "set_get_random"
  data_size: 64
  read_ratio: 0.5
```

### HTTP配置 (conf/http.yaml)

```yaml
protocol: http
connection:
  base_url: "http://localhost:8080"
  timeout: 30s
  max_conns_per_host: 50

benchmark:
  total: 10000
  parallels: 50
  method: "GET"
  path: "/api/test"
  headers:
    "Content-Type": "application/json"
    "Authorization": "Bearer token"
```

### Kafka配置 (conf/kafka.yaml)

```yaml
protocol: kafka
brokers: ["localhost:9092"]
topic_configs:
  - name: "test-topic"
    partitions: 3

producer:
  batch_size: 16384
  compression: "snappy"
  required_acks: 1

consumer:
  group_id: "test-group"
  auto_offset_reset: "earliest"

benchmark:
  total: 10000
  parallels: 5
  message_size: 1024
  test_type: "produce"
```

## 文档

详细文档请参阅以下资源：

- [架构概述](docs/architecture/overview.md) - 系统架构和设计原则
- [组件文档](docs/architecture/components.md) - 详细组件文档
- [快速入门指南](docs/usage/quickstart.md) - 快速开始
- [Redis测试指南](docs/usage/redis.md) - Redis特定功能和用法
- [HTTP测试指南](docs/usage/http.md) - HTTP特定功能和用法
- [Kafka测试指南](docs/usage/kafka.md) - Kafka特定功能和用法
- [贡献指南](docs/development/contributing.md) - 贡献指南
- [扩展redis-runner](docs/development/extending.md) - 如何扩展工具

## 从v0.0.x迁移

此版本引入了重大变更。主要变更：

- `redis-enhanced` → `redis`
- `http-enhanced` → `http`  
- `kafka-enhanced` → `kafka`
- 简化的命令结构
- 统一的配置格式

请参阅[迁移指南](docs/CHANGELOG.md)了解详细的升级说明。

## 示例

### Redis性能测试

```bash
# 基本性能测试
./redis-runner redis -h 127.0.0.1 -p 6379 -n 100000 -c 50

# 带认证的集群模式
./redis-runner redis --mode cluster -h localhost -p 6371 \n  -a "password" -n 100000 -c 10 -d 64 --read-ratio 50

# 自定义测试模式
./redis-runner redis -t incr -n 50000 -c 100  # 计数器操作
./redis-runner redis -t lpush_lpop -n 10000 -c 50  # 列表操作
```

### HTTP负载测试

```bash
# API端点测试
./redis-runner http --url http://api.example.com/health -n 10000 -c 100

# 带JSON负载的POST
./redis-runner http --url http://api.example.com/users \n  --method POST \n  --body '{"name":"John","email":"john@example.com"}' \n  --content-type "application/json" -n 1000 -c 20

# 带渐进的负载测试
./redis-runner http --url http://www.example.com \n  --duration 300s -c 200 --ramp-up 30s
```

### Kafka性能测试

```bash
# 生产者吞吐量测试
./redis-runner kafka --broker localhost:9092 --topic throughput-test \n  --message-size 1024 -n 100000 -c 10

# 消费者延迟测试
./redis-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id perf-test-group -n 50000

# 端到端延迟测试
./redis-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic latency-test --test-type produce_consume \n  --message-size 512 --duration 120s -c 5
```

## 许可证

[MIT](LICENSE)

## 贡献

欢迎贡献！请在提交PR之前阅读我们的贡献指南。

## 支持

如有问题和支持需求：

- 查看[迁移指南](docs/CHANGELOG.md)
- 查看命令帮助: `./redis-runner <command> --help`
- 提交issue报告bug或功能请求