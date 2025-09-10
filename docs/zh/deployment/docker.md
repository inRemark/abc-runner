# Docker部署指南

[English](../en/deployment/docker.md) | [中文](docker.md)

## Docker镜像

abc-runner提供官方Docker镜像，可用于快速部署和测试。

## 拉取镜像

```bash
# 拉取最新版本
docker pull abc-runner/abc-runner:latest

# 拉取特定版本
docker pull abc-runner/abc-runner:v0.2.0
```

## 基本使用

### 运行Redis测试

```bash
# 运行基本的Redis测试
docker run --rm abc-runner/abc-runner redis -h host.docker.internal -p 6379 -n 1000 -c 10

# 使用自定义配置文件
docker run --rm -v $(pwd)/config:/config abc-runner/abc-runner redis --config /config/redis.yaml
```

### 运行HTTP测试

```bash
# 运行HTTP测试
docker run --rm abc-runner/abc-runner http --url http://host.docker.internal:8080 -n 1000 -c 10
```

### 运行Kafka测试

```bash
# 运行Kafka测试
docker run --rm abc-runner/abc-runner kafka --broker host.docker.internal:9092 --topic test -n 1000 -c 5
```

## Docker Compose

创建`docker-compose.yml`文件来编排测试环境：

```yaml
version: '3.8'

services:
  abc-runner:
    image: abc-runner/abc-runner:latest
    depends_on:
      - redis
      - kafka
    volumes:
      - ./config:/config
      - ./reports:/reports
    command: redis --config /config/redis.yaml

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  zookeeper:
    image: bitnami/zookeeper:latest
    ports:
      - "2181:2181"
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes

  kafka:
    image: bitnami/kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
      - ALLOW_PLAINTEXT_LISTENER=yes
```

### 运行测试环境

```bash
# 启动所有服务
docker-compose up -d

# 运行Redis测试
docker-compose run --rm abc-runner redis -h redis -p 6379 -n 1000 -c 10

# 运行Kafka测试
docker-compose run --rm abc-runner kafka --broker kafka:9092 --topic test -n 1000 -c 5

# 停止所有服务
docker-compose down
```

## 自定义镜像

### 构建自定义镜像

创建`Dockerfile`：

```dockerfile
FROM abc-runner/abc-runner:latest

# 复制自定义配置
COPY config/ /config/

# 设置工作目录
WORKDIR /app

# 设置默认命令
ENTRYPOINT ["abc-runner"]
CMD ["--help"]
```

构建镜像：

```bash
docker build -t my-abc-runner .
```

### 多阶段构建

```dockerfile
# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

# 安装依赖
RUN go mod tidy

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o abc-runner .

# 运行阶段
FROM alpine:latest

# 安装ca证书
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/abc-runner .

# 复制配置文件
COPY config/ /config/

# 设置端口
EXPOSE 8080

# 运行应用
CMD ["./abc-runner"]
```

## 持久化数据

### 报告数据

```bash
# 挂载报告目录
docker run --rm -v $(pwd)/reports:/reports abc-runner/abc-runner redis -h redis -p 6379
```

### 配置文件

```bash
# 挂载配置目录
docker run --rm -v $(pwd)/config:/config abc-runner/abc-runner redis --config /config/redis.yaml
```

## 环境变量

支持以下环境变量：

```bash
docker run --rm \
  -e REDIS_HOST=redis \
  -e REDIS_PORT=6379 \
  -e KAFKA_BROKERS=kafka:9092 \
  abc-runner/abc-runner redis -n 1000 -c 10
```

## 网络配置

### 使用自定义网络

```bash
# 创建自定义网络
docker network create abc-runner-network

# 运行Redis
docker run -d --name redis --network abc-runner-network redis:7-alpine

# 运行abc-runner
docker run --rm --network abc-runner-network abc-runner/abc-runner redis -h redis -p 6379 -n 1000 -c 10
```

## 安全考虑

### 用户权限

```dockerfile
# 创建非root用户
RUN addgroup -g 1001 -S runner &&\
    adduser -u 1001 -S runner -G runner

# 切换到非root用户
USER runner
```

### 只读文件系统

```bash
# 使用只读文件系统（挂载必要的目录）
docker run --rm --read-only \
  -v $(pwd)/reports:/reports \
  -v $(pwd)/config:/config \
  abc-runner/abc-runner redis -h redis -p 6379
```

## 监控和日志

### 日志驱动

```bash
# 使用json-file日志驱动
docker run --rm --log-driver json-file --log-opt max-size=10m abc-runner/abc-runner redis -h redis -p 6379

# 使用syslog日志驱动
docker run --rm --log-driver syslog --log-opt syslog-address=tcp://localhost:514 abc-runner/abc-runner redis -h redis -p 6379
```

### 健康检查

```dockerfile
# 添加健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./abc-runner --version || exit 1
```

## 最佳实践

1. **使用标签**: 始终使用特定版本标签而不是latest
2. **最小化镜像**: 使用alpine基础镜像减小镜像大小
3. **安全**: 使用非root用户运行容器
4. **配置**: 通过挂载卷管理配置文件
5. **持久化**: 挂载卷以持久化报告数据
6. **网络**: 使用自定义网络隔离服务
7. **资源限制**: 设置内存和CPU限制
8. **监控**: 配置适当的日志和监控