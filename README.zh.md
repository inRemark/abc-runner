# abc-runner

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

## 打包和分发

### 发布包

预构建的发布包可从[发布页面](https://github.com/your-org/abc-runner/releases)下载。每个发布包含：

- 针对macOS、Linux和Windows平台的二进制文件
- 配置文件模板
- 文档和许可证文件

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/your-org/abc-runner.git
cd abc-runner

# 为当前平台构建
make build

# 为所有支持的平台构建
make build-all

# 创建发布包
make release

# 创建指定版本的发布包
VERSION=1.0.0 make release
```

有关打包过程的详细信息，请参阅[打包指南](docs/packaging-guide.md)。

## 快速开始

### 安装

```bash
# 从源码构建
go build -o abc-runner .

# 或从发布页面下载预构建的二进制文件
```

### 基本用法

```bash
# 显示帮助
./abc-runner --help

# Redis性能测试
./abc-runner redis -h localhost -p 6379 -n 10000 -c 50

# HTTP负载测试
./abc-runner http --url http://localhost:8080 -n 10000 -c 50

# Kafka性能测试
./abc-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5
```

### 使用别名

```bash
# 快速测试的短别名
./abc-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./abc-runner h --url http://httpbin.org/get -n 100  # HTTP
./abc-runner k --broker localhost:9092 -n 100      # Kafka
```

## 命令参考

### 全局选项

```bash
./abc-runner --help                 # 显示帮助
./abc-runner --version              # 显示版本
```

### Redis命令

```bash
# 基本Redis测试
./abc-runner redis -h <host> -p <port> -n <requests> -c <connections>

# 带认证的Redis
./abc-runner redis -h localhost -p 6379 -a password -n 10000 -c 50

# Redis集群模式
./abc-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50

# 自定义测试用例和读取比例
./abc-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80

# 使用配置文件
./abc-runner redis --config config/redis.yaml

# 使用配置文件和核心配置
./abc-runner redis --config config/redis.yaml --core-config config/core.yaml
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
./abc-runner http --url http://localhost:8080 -n 10000 -c 50

# 带请求体的HTTP POST
./abc-runner http --url http://api.example.com/users \n  --method POST --body '{"name":"test"}' \n  --content-type application/json -n 1000 -c 20

# 基于持续时间的测试
./abc-runner http --url http://localhost:8080 --duration 60s -c 100

# 自定义头部
./abc-runner http --url http://api.example.com \n  --header "Authorization:Bearer token123" \n  --header "X-API-Key:secret" -n 1000

# 使用配置文件和核心配置
./abc-runner http --config config/http.yaml --core-config config/core.yaml
```

### Kafka命令

```bash
# 基本生产者测试
./abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# 消费者测试
./abc-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id my-group -n 1000

# 混合生产和消费测试
./abc-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic high-throughput --test-type produce_consume \n  --message-size 4096 --duration 60s -c 8

# 高性能测试与压缩
./abc-runner kafka --broker localhost:9092 --topic perf-test \n  --compression lz4 --acks all --batch-size 32768 -n 50000

# 使用配置文件和核心配置
./abc-runner kafka --config config/kafka.yaml --core-config config/core.yaml
```

## 配置文件

您可以使用YAML配置文件进行复杂设置：

### 核心配置 (config/core.yaml)

核心配置文件包含所有协议共享的通用设置：

```yaml
core:
  # 日志配置
  logging:
    level: "info"              # 日志级别: debug, info, warn, error
    format: "json"             # 日志格式: json, text
    output: "stdout"           # 输出目标: stdout, file, 或文件路径
    file_path: "./logs"        # 日志文件目录
    max_size: "100MB"          # 单个日志文件最大大小
    max_age: 7                 # 日志文件保留天数
    max_backups: 5             # 最大备份文件数
    compress: true             # 是否压缩旧日志文件

  # 报告配置
  reports:
    enabled: true              # 是否启用报告
    formats: ["console"]       # 报告格式: console, json, csv, text, all
    output_dir: "./reports"    # 报告输出目录
    file_prefix: "benchmark"   # 报告文件前缀
    include_timestamp: true    # 文件名包含时间戳
    enable_console_report: true # 启用控制台详细报告
    overwrite_existing: false  # 是否覆盖已存在文件

  # 监控配置
  monitoring:
    enabled: true              # 是否启用监控
    metrics_interval: "5s"     # 指标收集间隔
    prometheus:
      enabled: false           # 是否启用Prometheus导出
      port: 9090               # Prometheus导出端口
    statsd:
      enabled: false           # 是否启用StatsD导出
      host: "localhost:8125"   # StatsD服务器地址

  # 全局连接配置
  connection:
    timeout: "30s"             # 默认连接超时
    keep_alive: "30s"          # 连接保持时间
    max_idle_conns: 100        # 最大空闲连接数
    idle_conn_timeout: "90s"   # 空闲连接超时时间
```

### Redis配置 (config/redis.yaml)

```yaml
redis:
  mode: "standalone"    # 选项: standalone, sentinel, cluster
  benchmark:
    total: 10000              # 默认10000个请求
    parallels: 50             # 默认50个并行连接
    random_keys: 50           # 0:递增键, >0:随机键范围是[0, r]
    read_percent: 50          # 默认50%读取和50%写入
    data_size: 3              # 默认3字节
    ttl: 120                  # 默认120秒
    case: "set_get_random"    # 操作类型: set_get_random, set, get, del, pub, sub
  pool:
    pool_size: 10
    min_idle: 2
  standalone:
    addr: 127.0.0.1:6379
    password: "pwd@redis"
    db: 0
  sentinel:
    master_name: "mymaster"
    addrs:
      - "127.0.0.1:26371"
      - "127.0.0.1:26372"
      - "127.0.0.1:26373"
    password: "pwd@redis"
    db: 0
  cluster:
    addrs:
      - "127.0.0.1:6371"
      - "127.0.0.1:6372"
      - "127.0.0.1:6373"
    password: "pwd@redis"
```

### HTTP配置 (config/http.yaml)

```yaml
http:
  connection:
    base_url: "http://localhost:8080"
    timeout: 30s
    keep_alive: 90s
    max_idle_conns: 50
    max_conns_per_host: 20
    idle_conn_timeout: 90s
    disable_compression: false
  benchmark:
    total: 100000
    parallels: 50
    duration: "5m"
    ramp_up: "30s"
    data_size: 1024
    ttl: 0s
    read_percent: 70
    random_keys: 0
    test_case: "mixed_operations"
    timeout: 30s
  requests:
    - method: "GET"
      path: "/api/users"
      headers:
        Accept: "application/json"
      weight: 100
```

### Kafka配置 (config/kafka.yaml)

```yaml
kafka:
  brokers:
    - "localhost:9092"
  client_id: "abc-runner-kafka-client"
  version: "2.8.0"
  producer:
    acks: "all"
    retries: 3
    batch_size: 16384
    linger_ms: "5ms"
    compression: "snappy"
    idempotence: true
    max_in_flight: 5
    request_timeout: "30s"
    write_timeout: "10s"
    read_timeout: "10s"
  benchmark:
    default_topic: "benchmark-topic"
    message_size_range:
      min: 100
      max: 10240
    batch_sizes: [1, 10, 100, 1000]
    partition_strategy: "round_robin"
    total: 100000
    parallels: 50
    data_size: 1024
    ttl: 0
    read_percent: 50
    random_keys: 10000
    test_case: "produce"
    timeout: "30s"
```

## 文档

详细文档请参阅以下资源：

- [架构概述](docs/en/architecture/overview.md) - 系统架构和设计原则 | [架构概述](docs/zh/architecture/overview.md)
- [组件文档](docs/en/architecture/components.md) - 详细组件文档 | [组件详解](docs/zh/architecture/components.md)
- [快速入门指南](docs/en/getting-started/quickstart.md) - 快速开始 | [快速开始](docs/zh/getting-started/quickstart.md)
- [Redis测试指南](docs/en/user-guide/redis.md) - Redis特定功能和用法 | [Redis测试指南](docs/zh/user-guide/redis.md)
- [HTTP测试指南](docs/en/user-guide/http.md) - HTTP特定功能和用法 | [HTTP测试指南](docs/zh/user-guide/http.md)
- [Kafka测试指南](docs/en/user-guide/kafka.md) - Kafka特定功能和用法 | [Kafka测试指南](docs/zh/user-guide/kafka.md)
- [贡献指南](docs/en/developer-guide/contributing.md) - 贡献指南 | [贡献指南](docs/zh/developer-guide/contributing.md)
- [扩展abc-runner](docs/en/developer-guide/extending.md) - 如何扩展工具 | [扩展abc-runner](docs/zh/developer-guide/extending.md)

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
./abc-runner redis -h 127.0.0.1 -p 6379 -n 100000 -c 50

# 带认证的集群模式
./abc-runner redis --mode cluster -h localhost -p 6371 \n  -a "password" -n 100000 -c 10 -d 64 --read-ratio 50

# 自定义测试模式
./abc-runner redis -t incr -n 50000 -c 100  # 计数器操作
./abc-runner redis -t lpush_lpop -n 10000 -c 50  # 列表操作
```

### HTTP负载测试

```bash
# API端点测试
./abc-runner http --url http://api.example.com/health -n 10000 -c 100

# 带JSON负载的POST
./abc-runner http --url http://api.example.com/users \n  --method POST \n  --body '{"name":"John","email":"john@example.com"}' \n  --content-type "application/json" -n 1000 -c 20

# 带渐进的负载测试
./abc-runner http --url http://www.example.com \n  --duration 300s -c 200 --ramp-up 30s
```

### Kafka性能测试

```bash
# 生产者吞吐量测试
./abc-runner kafka --broker localhost:9092 --topic throughput-test \n  --message-size 1024 -n 100000 -c 10

# 消费者延迟测试
./abc-runner kafka --broker localhost:9092 --topic test-topic \n  --test-type consume --group-id perf-test-group -n 50000

# 端到端延迟测试
./abc-runner kafka --brokers localhost:9092,localhost:9093 \n  --topic latency-test --test-type produce_consume \n  --message-size 512 --duration 120s -c 5
```

## 许可证

[Apache License 2.0](LICENSE)

## 贡献

欢迎贡献！请在提交PR之前阅读我们的贡献指南。

## 支持

如有问题和支持需求：

- 查看[迁移指南](docs/CHANGELOG.md)
- 查看命令帮助: `./abc-runner <command> --help`
- 提交issue报告bug或功能请求

## 文档维护

本项目维护英文和中文两种语言的文档。有关维护多语言文档的指南，请参阅[文档翻译指南](docs/maintenance/document-translation-guide.md)。
