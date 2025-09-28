# abc-runner

[English](README.md) | [中文](README.zh.md)

## 关于

一个用于Redis、HTTP和Kafka协议的统一性能测试工具。

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

### 构建

```bash
# Clone the repository
git clone https://github.com/your-org/abc-runner.git
cd abc-runner

# Build for current platform
make build

# Build for all supported platforms
make build-all

# Create release packages
make release

# Create release packages with specific version
VERSION=1.0.0 make release
```

For detailed information about the packaging process, see [Packaging Guide](docs/packaging-guide.md).

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

```bash
config/core.yaml
config.http.yaml
config/redis.yaml
config/kafka.yaml
config/tcp.yaml
config/udp.yaml
config/grpc.yaml
config/websocket.yaml
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

## 许可证

[Apache License 2.0](LICENSE)

## 贡献

欢迎贡献！请在提交PR之前阅读我们的贡献指南。

## 支持

如有问题和支持需求：

- 查看[迁移指南](docs/CHANGELOG.md)
- 查看命令帮助: `./abc-runner <command> --help`
- 提交issue报告bug或功能请求
