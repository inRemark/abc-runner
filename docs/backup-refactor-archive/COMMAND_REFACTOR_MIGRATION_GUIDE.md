# Redis Runner 命令重构迁移指南

## 概述

abc-runner 已完成破坏式升级，统一使用 `redis`、`http`、`kafka` 作为子命令名称，移除了增强命令（redis-enhanced、http-enhanced、kafka-enhanced）和传统命令的复杂兼容性架构。

## 主要变更

### 1. 命令简化

**旧版本命令:**

```bash
# 增强版命令
abc-runner redis-enhanced --config conf/redis.yaml
abc-runner http-enhanced --url http://localhost:8080
abc-runner kafka-enhanced --brokers localhost:9092

# 传统版命令（已移除）
abc-runner redis -h localhost -p 6379
abc-runner http --url http://localhost:8080
abc-runner kafka --brokers localhost:9092
```

**新版本命令:**

```bash
# 统一命令格式
abc-runner redis --config conf/redis.yaml
abc-runner http --url http://localhost:8080
abc-runner kafka --brokers localhost:9092

# 支持别名
abc-runner r -h localhost -p 6379 -n 1000 -c 10
abc-runner h --url http://localhost:8080 -n 1000 -c 50
abc-runner k --brokers localhost:9092 -n 1000 -c 5
```

### 2. 命令对照表

| 旧版本命令 | 新版本命令 | 别名 | 说明 |
|-----------|------------|------|------|
| `redis-enhanced` | `redis` | `r` | Redis性能测试 |
| `http-enhanced` | `http` | `h` | HTTP负载测试 |
| `kafka-enhanced` | `kafka` | `k` | Kafka性能测试 |
| `redis` (传统版) | `redis` | `r` | 合并到新版redis命令 |
| `http` (传统版) | `http` | `h` | 合并到新版http命令 |
| `kafka` (传统版) | `kafka` | `k` | 合并到新版kafka命令 |

### 3. 移除的功能

- **统一命令管理器**: 移除了复杂的统一管理架构
- **命令升级机制**: 移除了自动命令升级逻辑
- **复杂别名映射**: 简化为基础的 r/h/k 别名
- **兼容性警告**: 移除了弃用警告系统

## 迁移步骤

### 1. 更新脚本

使用以下命令自动更新您的脚本：

```bash
# 替换增强版命令
sed -i 's/redis-enhanced/redis/g' your_script.sh
sed -i 's/http-enhanced/http/g' your_script.sh
sed -i 's/kafka-enhanced/kafka/g' your_script.sh

# 检查并手动调整其他配置
```

### 2. 验证功能

```bash
# 测试Redis命令
abc-runner redis --help
abc-runner r -h localhost -p 6379 -n 10 -c 1

# 测试HTTP命令  
abc-runner http --help
abc-runner h --url http://httpbin.org/get -n 10 -c 1

# 测试Kafka命令
abc-runner kafka --help
abc-runner k --brokers localhost:9092 --topic test -n 10 -c 1
```

## 新功能特性

### 1. 简化的命令结构

- 统一的命令格式
- 清晰的参数命名
- 一致的帮助系统

### 2. 保留的核心功能

- 所有原有的性能测试功能
- 配置文件支持
- 详细的性能报告
- 多种测试模式

### 3. 改进的用户体验

- 更简洁的命令行接口
- 统一的错误处理
- 一致的输出格式

## 示例用法

### Redis 性能测试

```bash
# 基础测试
abc-runner redis -h 127.0.0.1 -p 6379 -n 10000 -c 50

# 使用配置文件
abc-runner redis --config conf/redis.yaml

# 自定义测试用例
abc-runner r -t set_get_random -n 50000 -c 100 --read-ratio 80
```

### HTTP 负载测试

```bash
# GET请求测试
abc-runner http --url http://localhost:8080/api/users -n 10000 -c 50

# POST请求测试
abc-runner h --url http://api.example.com/users --method POST \
  --body '{"name":"test","email":"test@example.com"}' \
  --content-type application/json -n 1000 -c 20

# 持续时间测试
abc-runner http --url https://api.example.com/health \
  --duration 60s -c 100
```

### Kafka 性能测试

```bash
# 生产者测试
abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

# 消费者测试
abc-runner k --broker localhost:9092 --topic test-topic \
  --test-type consume --group-id my-group -n 1000

# 混合测试
abc-runner kafka --brokers localhost:9092,localhost:9093 \
  --topic high-throughput --test-type produce_consume \
  --message-size 4096 --duration 60s -c 8
```

## 配置文件迁移

配置文件格式保持基本兼容，但建议检查以下项目：

### Redis 配置 (conf/redis.yaml)

```yaml
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

### HTTP 配置 (conf/http.yaml)

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
```

### Kafka 配置 (conf/kafka.yaml)

```yaml
protocol: kafka
brokers: ["localhost:9092"]
topic_configs:
  - name: "test-topic"
    partitions: 3

benchmark:
  total: 10000
  parallels: 5
  message_size: 1024
  test_type: "produce"
```

## 故障排除

### 常见问题

1. **命令不存在错误**

   ```bash
   unknown command: redis-enhanced
   ```

   **解决方案**: 使用 `redis` 替代 `redis-enhanced`

2. **配置文件错误**

   ```bash
   failed to load configuration
   ```

   **解决方案**: 检查配置文件格式，确保字段名正确

3. **连接问题**

   ```bash
   failed to connect to Redis/HTTP/Kafka
   ```

   **解决方案**: 检查目标服务是否运行，网络是否可达

### 获取帮助

```bash
# 查看全局帮助
abc-runner --help

# 查看具体命令帮助
abc-runner redis --help
abc-runner http --help
abc-runner kafka --help

# 查看版本信息
abc-runner --version
```

## 向后兼容性说明

⚠️ **重要提醒**: 这是一个破坏性升级，以下功能不再兼容：

- `redis-enhanced`、`http-enhanced`、`kafka-enhanced` 命令
- 统一命令管理器的API
- 复杂的别名映射规则
- 自动命令升级机制

所有现有脚本和配置需要按照本指南进行迁移。

## 技术支持

如果在迁移过程中遇到问题，请：

1. 检查本迁移指南
2. 查看命令帮助信息
3. 验证配置文件格式
4. 确认目标服务可用性

该重构显著简化了命令结构，提供了更一致和直观的用户体验。
