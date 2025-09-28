# abc-runner 多协议服务端部署指南

## 概述

本文档提供 abc-runner 多协议服务端辅助模块的详细部署和使用指南。这些服务端模块为 abc-runner 性能测试工具提供标准化的测试目标。

## 系统要求

### 硬件要求

| 项目 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 1 核心 | 2+ 核心 |
| 内存 | 512MB | 1GB+ |
| 磁盘 | 100MB | 1GB |
| 网络 | 100Mbps | 1Gbps |

### 软件要求

- **操作系统**: Linux (Ubuntu 18.04+), macOS (10.14+), Windows (10+)
- **Go 版本**: 1.21 或更高版本
- **网络端口**: 确保以下端口可用：
  - HTTP: 8080 (默认)
  - TCP: 9090 (默认)
  - UDP: 9091 (默认)
  - gRPC: 50051 (默认)

## 快速开始

### 1. 构建服务端

```bash
# 进入服务端目录
cd abc-runner/servers

# 构建所有服务端
make build
# 或手动构建
go build -o bin/http-server ./cmd/http-server
go build -o bin/tcp-server ./cmd/tcp-server
go build -o bin/udp-server ./cmd/udp-server
go build -o bin/grpc-server ./cmd/grpc-server
go build -o bin/multi-server ./cmd/multi-server
```

### 2. 启动所有服务端

```bash
# 使用启动脚本
./scripts/start-all.sh

# 或使用多服务端启动器
./bin/multi-server
```

### 3. 验证服务端状态

```bash
# 使用健康检查脚本
./scripts/health-check.sh

# 或手动检查
curl http://localhost:8080/health    # HTTP服务端
curl http://localhost:50051/         # gRPC服务端
```

## 详细部署

### 单协议部署

#### HTTP服务端

```bash
# 启动HTTP服务端
./bin/http-server --host 0.0.0.0 --port 8080 --log-level info

# 测试HTTP服务端
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl -d '{"test":"data"}' http://localhost:8080/echo
```

**可用端点:**

- `/` - 服务信息
- `/health` - 健康检查
- `/metrics` - 指标信息
- `/echo` - 回显测试
- `/delay?delay=100ms` - 延迟测试
- `/status?code=404` - 状态码测试
- `/data?size=1024` - 数据大小测试

#### TCP服务端

```bash
# 启动TCP服务端
./bin/tcp-server --host 0.0.0.0 --port 9090 --log-level info

# 测试TCP服务端（需要TCP客户端工具）
echo "Hello TCP" | nc localhost 9090
```

**特性:**

- 回显服务器
- 长度前缀协议
- 连接池管理
- Keep-Alive支持
- 可配置超时

#### UDP服务端

```bash
# 启动UDP服务端
./bin/udp-server --host 0.0.0.0 --port 9091 --log-level info

# 测试UDP服务端
echo "Hello UDP" | nc -u localhost 9091
```

**特性:**

- 数据包回显
- 丢包模拟
- 多播支持
- 广播支持
- 统计信息

#### gRPC服务端

```bash
# 启动gRPC服务端
./bin/grpc-server --host 0.0.0.0 --port 50051 --log-level info

# 测试gRPC服务端
curl http://localhost:50051/                           # 服务信息
curl -X POST -d '{"message":"hello"}' http://localhost:50051/TestService/Echo
```

**服务方法:**

- `Echo` - 回显服务
- `ServerStream` - 服务端流
- `ClientStream` - 客户端流
- `BidirectionalStream` - 双向流
- `Health` - 健康检查

### 多协议部署

#### 使用多服务端启动器

```bash
# 启动所有协议服务端
./bin/multi-server

# 启动指定协议
./bin/multi-server --protocols http,tcp

# 自定义端口
./bin/multi-server --http-port 8888 --tcp-port 9999

# 不同主机
./bin/multi-server --host 0.0.0.0
```

#### 使用启动脚本

```bash
# 启动所有服务端
./scripts/start-all.sh

# 后台运行
./scripts/start-all.sh --daemon

# 指定协议
./scripts/start-all.sh --protocols http,grpc

# 查看状态
./scripts/start-all.sh --status

# 停止所有服务端
./scripts/stop-all.sh
```

## 配置管理

### 配置文件位置

```bash
config/
├── servers/
│   ├── http-server.yaml
│   ├── tcp-server.yaml
│   ├── udp-server.yaml
│   └── grpc-server.yaml
└── examples/
    └── custom-config.yaml
```

### HTTP服务端配置示例

```yaml
# config/servers/http-server.yaml
protocol: http
host: localhost
port: 8080

# 超时配置
read_timeout: 30s
write_timeout: 30s
idle_timeout: 60s
max_header_bytes: 1048576

# 响应配置
response:
  default_delay: 0ms
  default_status_code: 200
  default_size: 1024
  content_type: application/json

# CORS配置
cors:
  enabled: true
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE"]

# TLS配置
tls:
  enabled: false
  cert_file: ""
  key_file: ""
```

### TCP服务端配置示例

```yaml
# config/servers/tcp-server.yaml
protocol: tcp
host: localhost
port: 9090

# 连接配置
max_connections: 1000
connection_timeout: 30s
read_timeout: 30s
write_timeout: 30s
keep_alive: true

# 缓冲区配置
buffer_size: 4096
max_message_size: 65536

# 行为配置
echo_mode: true
response_delay: 0ms
log_connections: true
```

## 监控和日志

### 日志管理

#### 日志级别

- `debug` - 详细调试信息
- `info` - 一般信息（推荐）
- `warn` - 警告信息
- `error` - 错误信息

#### 日志格式

```json
{
  "timestamp": "2023-09-28T10:30:00Z",
  "level": "INFO",
  "message": "HTTP server started",
  "fields": {
    "protocol": "http",
    "address": "localhost:8080"
  }
}
```

### 指标监控

#### HTTP指标端点

```bash
curl http://localhost:8080/metrics
```

#### 关键指标

- `total_requests` - 总请求数
- `success_requests` - 成功请求数
- `failed_requests` - 失败请求数
- `active_connections` - 活跃连接数
- `uptime` - 运行时间
- `request_stats` - 请求统计（平均、最小、最大、百分位）

### 健康检查

#### 自动健康检查

```bash
# 使用健康检查脚本
./scripts/health-check.sh

# 自定义端口检查
./scripts/health-check.sh --http-port 8888 --tcp-port 9999
```

#### 手动健康检查

```bash
# HTTP健康检查
curl -f http://localhost:8080/health

# gRPC健康检查
curl -f http://localhost:50051/grpc.health.v1.Health/Check

# TCP连接测试
timeout 3 bash -c "</dev/tcp/localhost/9090"
```

## 生产环境部署

### 系统服务部署

#### systemd 服务文件示例

```ini
# /etc/systemd/system/abc-runner-servers.service
[Unit]
Description=abc-runner Multi-Protocol Test Servers
After=network.target

[Service]
Type=simple
User=abc-runner
Group=abc-runner
WorkingDirectory=/opt/abc-runner/servers
ExecStart=/opt/abc-runner/servers/bin/multi-server --host 0.0.0.0
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# 启用并启动服务
sudo systemctl enable abc-runner-servers
sudo systemctl start abc-runner-servers
sudo systemctl status abc-runner-servers
```

### Docker 部署

#### Dockerfile 示例

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/multi-server ./cmd/multi-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/multi-server .
COPY --from=builder /app/config ./config

EXPOSE 8080 9090 9091 50051

CMD ["./multi-server", "--host", "0.0.0.0"]
```

```bash
# 构建镜像
docker build -t abc-runner-servers .

# 运行容器
docker run -d --name abc-runner-servers \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 9091:9091 \
  -p 50051:50051 \
  abc-runner-servers
```

#### Docker Compose 示例

```yaml
# docker-compose.yml
version: '3.8'

services:
  abc-runner-servers:
    build: .
    ports:
      - "8080:8080"   # HTTP
      - "9090:9090"   # TCP
      - "9091:9091"   # UDP
      - "50051:50051" # gRPC
    environment:
      - LOG_LEVEL=info
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 性能调优

#### 系统级调优

```bash
# 增加文件描述符限制
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# 调整网络参数
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

#### 应用级调优

```yaml
# 高性能配置示例
tcp:
  max_connections: 10000
  buffer_size: 8192
  keep_alive: true
  no_delay: true

http:
  read_timeout: 10s
  write_timeout: 10s
  max_header_bytes: 2097152

udp:
  buffer_size: 8192
  max_packet_size: 65507
```

## 故障排查

### 常见问题

#### 1. 端口占用

```bash
# 检查端口占用
netstat -tlnp | grep :8080
lsof -i :8080

# 解决方案
pkill -f http-server
# 或更改端口
./bin/http-server --port 8081
```

#### 2. 权限问题

```bash
# 检查文件权限
ls -la bin/
chmod +x bin/*

# 检查端口权限（1024以下端口需要root）
sudo ./bin/http-server --port 80
```

#### 3. 内存不足

```bash
# 检查内存使用
free -h
ps aux | grep server

# 监控服务端内存
while true; do
  ps -o pid,rss,comm -p $(pgrep -f multi-server)
  sleep 5
done
```

#### 4. 网络连接问题

```bash
# 检查网络连接
ss -tlnp | grep :8080
curl -v http://localhost:8080/health

# 检查防火墙
ufw status
iptables -L
```

### 日志分析

#### 查看实时日志

```bash
# 如果使用后台运行
tail -f logs/multi.log

# 如果使用systemd
journalctl -u abc-runner-servers -f

# 过滤错误日志
grep "ERROR" logs/*.log
```

#### 性能分析

```bash
# 分析请求延迟
grep "duration" logs/http.log | awk '{print $NF}' | sort -n

# 统计请求数量
grep "HTTP request" logs/http.log | wc -l

# 分析错误率
grep -c "ERROR" logs/*.log
```

## 安全配置

### TLS/SSL 配置

#### 生成自签名证书

```bash
# 生成私钥
openssl genrsa -out server.key 2048

# 生成证书
openssl req -new -x509 -key server.key -out server.crt -days 365

# 配置HTTPS
./bin/http-server --config config/https.yaml
```

#### TLS配置文件

```yaml
# config/https.yaml
tls:
  enabled: true
  cert_file: "certs/server.crt"
  key_file: "certs/server.key"
```

### 防火墙配置

#### iptables 规则

```bash
# 允许特定端口
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
iptables -A INPUT -p tcp --dport 9090 -j ACCEPT
iptables -A INPUT -p udp --dport 9091 -j ACCEPT
iptables -A INPUT -p tcp --dport 50051 -j ACCEPT

# 保存规则
iptables-save > /etc/iptables/rules.v4
```

#### ufw 配置

```bash
ufw allow 8080/tcp
ufw allow 9090/tcp
ufw allow 9091/udp
ufw allow 50051/tcp
ufw enable
```

## 集成测试

### 自动化测试

```bash
# 运行集成测试
go test ./test/integration/... -v

# 运行基准测试
go test ./test/integration/... -bench=.

# 测试覆盖率
go test ./test/integration/... -cover
```

### 负载测试

#### HTTP负载测试

```bash
# 使用ab（Apache Bench）
ab -n 10000 -c 100 http://localhost:8080/health

# 使用wrk
wrk -t12 -c400 -d30s http://localhost:8080/health
```

#### 自定义测试脚本

```bash
#!/bin/bash
# load-test.sh

echo "开始负载测试..."

# HTTP测试
ab -n 1000 -c 10 http://localhost:8080/ > http_results.txt &

# gRPC测试
for i in {1..100}; do
  curl -s http://localhost:50051/TestService/Echo \
    -d '{"message":"test"}' > /dev/null &
done

wait
echo "负载测试完成"
```

## 维护指南

### 定期维护任务

#### 日志轮转

```bash
# 配置logrotate
cat > /etc/logrotate.d/abc-runner << EOF
/opt/abc-runner/servers/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        systemctl reload abc-runner-servers
    endscript
}
EOF
```

#### 性能监控脚本

```bash
#!/bin/bash
# monitor.sh

LOG_FILE="/var/log/abc-runner-monitor.log"

while true; do
    echo "$(date): Monitoring abc-runner servers" >> $LOG_FILE
    
    # 检查服务状态
    if ! curl -s http://localhost:8080/health > /dev/null; then
        echo "$(date): HTTP server health check failed" >> $LOG_FILE
        # 发送告警
        # send_alert "HTTP server down"
    fi
    
    # 检查内存使用
    MEM_USAGE=$(ps -o pid,rss -p $(pgrep -f multi-server) | awk 'NR>1{sum+=$2} END{print sum}')
    if [ $MEM_USAGE -gt 1048576 ]; then  # 1GB
        echo "$(date): High memory usage: ${MEM_USAGE}KB" >> $LOG_FILE
    fi
    
    sleep 60
done
```

### 备份和恢复

#### 配置备份

```bash
#!/bin/bash
# backup-config.sh

BACKUP_DIR="/backup/abc-runner/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# 备份配置文件
cp -r config/ $BACKUP_DIR/
cp -r scripts/ $BACKUP_DIR/

# 备份二进制文件
cp -r bin/ $BACKUP_DIR/

echo "Backup completed: $BACKUP_DIR"
```

#### 配置恢复

```bash
#!/bin/bash
# restore-config.sh

if [ -z "$1" ]; then
    echo "Usage: $0 <backup_date>"
    exit 1
fi

BACKUP_DIR="/backup/abc-runner/$1"

if [ ! -d "$BACKUP_DIR" ]; then
    echo "Backup directory not found: $BACKUP_DIR"
    exit 1
fi

# 停止服务
./scripts/stop-all.sh

# 恢复配置
cp -r $BACKUP_DIR/config/ ./
cp -r $BACKUP_DIR/scripts/ ./
cp -r $BACKUP_DIR/bin/ ./

# 启动服务
./scripts/start-all.sh

echo "Restore completed from: $BACKUP_DIR"
```

## 结语

本部署指南涵盖了 abc-runner 多协议服务端的完整部署和运维流程。通过遵循这些最佳实践，您可以确保服务端的稳定、高效运行，为 abc-runner 性能测试提供可靠的测试环境。

如有问题，请参考故障排查章节或查看项目文档获取更多帮助。
