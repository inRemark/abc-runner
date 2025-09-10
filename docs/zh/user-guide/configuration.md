# 配置管理指南

[English](../en/user-guide/configuration.md) | [中文](configuration.md)

## 配置文件结构

abc-runner使用YAML格式的配置文件，支持三种协议的配置：

```
config/
├── templates/           # 配置模板文件
├── examples/            # 配置示例文件
├── production/          # 生产环境配置
├── development/         # 开发环境配置
└── README.md            # 配置说明文档
```

## 配置优先级

配置按以下优先级顺序加载：

1. **命令行参数**: 最高优先级
2. **环境变量**: 中等优先级
3. **配置文件**: 最低优先级

## 通用配置选项

### 基准测试配置

所有协议都支持以下基准测试配置：

```yaml
benchmark:
  total: 10000              # 总请求数/消息数
  parallels: 50             # 并发连接数
  duration: "60s"           # 测试持续时间
  data_size: 1024           # 数据大小(字节)
  read_percent: 50          # 读操作百分比
  random_keys: 1000         # 随机键范围
  timeout: "30s"            # 超时时间
```

### 报告配置

```yaml
reports:
  enabled: true
  formats: ["console", "json", "csv"]
  output_dir: "./reports"
  file_prefix: "benchmark"
  include_timestamp: true
  enable_console_report: true
```

## Redis配置

### 连接配置

```yaml
redis:
  mode: "standalone"        # standalone, sentinel, cluster
  standalone:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
  sentinel:
    master_name: "mymaster"
    addrs:
      - "127.0.0.1:26371"
    password: ""
    db: 0
  cluster:
    addrs:
      - "127.0.0.1:6371"
    password: ""
```

### 连接池配置

```yaml
pool:
  pool_size: 10
  min_idle: 2
```

## HTTP配置

### 连接配置

```yaml
http:
  connection:
    base_url: "http://example.com"
    timeout: "30s"
    keep_alive: "90s"
    max_idle_conns: 50
    max_conns_per_host: 20
```

### 请求模板

```yaml
requests:
  - method: "GET"
    path: "/api/users"
    headers:
      Accept: "application/json"
    weight: 100
```

### 认证配置

```yaml
auth:
  type: "bearer"
  token: "your-token"
```

## Kafka配置

### 连接配置

```yaml
kafka:
  brokers:
    - "localhost:9092"
  client_id: "abc-runner-client"
```

### 生产者配置

```yaml
producer:
  acks: "all"
  batch_size: 16384
  compression: "snappy"
  linger_ms: "5ms"
```

### 消费者配置

```yaml
consumer:
  group_id: "abc-runner-group"
  auto_offset_reset: "latest"
```

## 环境变量

支持以下环境变量：

- `ABC_RUNNER_CONFIG`: 配置文件路径
- `ABC_RUNNER_LOG_LEVEL`: 日志级别
- `REDIS_HOST`: Redis主机
- `REDIS_PORT`: Redis端口
- `REDIS_PASSWORD`: Redis密码
- `KAFKA_BROKERS`: Kafka broker列表

## 配置验证

### 命令行验证

```bash
# 验证配置文件
./abc-runner redis --config config/redis.yaml --validate
```

### 配置示例验证

```bash
# 使用示例配置验证
./abc-runner redis --config config/examples/redis.yaml --dry-run
```

## 最佳实践

1. **环境分离**: 为不同环境维护独立的配置文件
2. **版本控制**: 将配置文件纳入版本控制
3. **敏感信息**: 使用环境变量处理敏感信息（如密码）
4. **模板使用**: 使用模板文件作为配置起点
5. **文档化**: 为自定义配置添加注释说明
6. **测试**: 在生产环境使用前测试配置
7. **备份**: 定期备份生产环境配置