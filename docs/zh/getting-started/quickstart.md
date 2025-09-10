# 快速开始

[English](../en/getting-started/quickstart.md) | [中文](quickstart.md)

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/your-org/abc-runner.git
cd abc-runner

# 构建
make build

# 或者直接使用Go构建
go build -o abc-runner .
```

### 使用预编译二进制文件

从[发布页面](https://github.com/your-org/abc-runner/releases)下载适合您系统的预编译二进制文件。

## 基本使用

### Redis性能测试

```bash
# 基本测试
./abc-runner redis -h localhost -p 6379 -n 10000 -c 50

# 使用配置文件
./abc-runner redis --config config/examples/redis.yaml
```

### HTTP负载测试

```bash
# 基本测试
./abc-runner http --url http://localhost:8080 -n 10000 -c 50

# 使用配置文件
./abc-runner http --config config/examples/http.yaml
```

### Kafka性能测试

```bash
# 基本测试
./abc-runner kafka --broker localhost:9092 --topic test -n 10000 -c 5

# 使用配置文件
./abc-runner kafka --config config/examples/kafka.yaml
```

## 使用别名

```bash
# 使用短别名进行快速测试
./abc-runner r -h localhost -p 6379 -n 1000 -c 10  # Redis
./abc-runner h --url http://httpbin.org/get -n 100  # HTTP
./abc-runner k --broker localhost:9092 -n 100      # Kafka
```

## 查看帮助

```bash
# 显示全局帮助
./abc-runner --help

# 显示特定命令帮助
./abc-runner redis --help
./abc-runner http --help
./abc-runner kafka --help
```