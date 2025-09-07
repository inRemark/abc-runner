# 快速入门指南

[English](quickstart.md) | [中文](quickstart.zh.md)

本指南将帮助您快速开始使用redis-runner。

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/your-org/redis-runner.git
cd redis-runner

# 构建二进制文件
go build -o redis-runner .

# 运行工具
./redis-runner --help
```

### 使用预构建的二进制文件

从[发布页面](https://github.com/your-org/redis-runner/releases)下载预构建的二进制文件。

## 基本用法

### Redis性能测试

```bash
# 基本Redis测试
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50

# 带认证的Redis
./redis-runner redis -h localhost -p 6379 -a password -n 10000 -c 50

# Redis集群模式
./redis-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50
```

### HTTP负载测试

```bash
# 基本HTTP GET测试
./redis-runner http --url http://localhost:8080 -n 10000 -c 50

# 带主体的HTTP POST
./redis-runner http --url http://api.example.com/users \
  --method POST --body '{"name":"test"}' \
  --content-type application/json -n 1000 -c 20
```

### Kafka性能测试

```bash
# 基本生产者测试
./redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# 消费者测试
./redis-runner kafka --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000
```

## 使用配置文件

您可以使用YAML配置文件进行复杂设置。更多详情请参见[配置文档](configuration.md)。

### Redis配置示例

```yaml
# redis.yaml
protocol: redis
connection:
  host: localhost
  port: 6379
  mode: standalone
  timeout: 30s

benchmark:
  total: 10000
  parallels: 50
  test_case: "set_get_random"
  data_size: 64
  read_ratio: 0.5
```

使用配置文件运行：

```bash
./redis-runner redis --config redis.yaml
```

## 命令别名

为了快速测试，您可以使用短别名：

```bash
# 快速测试的短别名
./redis-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./redis-runner h --url http://httpbin.org/get -n 100  # HTTP
./redis-runner k --broker localhost:9092 -n 100      # Kafka
```

## 查看帮助

```bash
# 显示通用帮助
./redis-runner --help

# 显示特定协议的帮助
./redis-runner redis --help
./redis-runner http --help
./redis-runner kafka --help
```