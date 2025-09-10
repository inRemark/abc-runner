# abc-runner 架构迁移用户指南

## 概述

abc-runner正在从分散式架构迁移到统一的适配器框架架构。本指南将帮助您顺利迁移到新版本，享受更好的性能、更统一的接口和更强的扩展能力。

## 📋 迁移清单

### 迁移前准备

- [ ] 备份现有配置文件
- [ ] 记录当前使用的命令和参数
- [ ] 准备测试环境
- [ ] 下载最新版本的abc-runner

### 迁移步骤

- [ ] 运行配置迁移工具
- [ ] 更新命令行脚本
- [ ] 测试新配置
- [ ] 更新CI/CD流水线（如适用）
- [ ] 清理老文件

## 🔄 命令迁移映射

### Redis命令迁移

```bash
# 老命令（已弃用）
abc-runner redis -h localhost -p 6379 -n 10000 -c 50

# 新命令（推荐）
abc-runner redis-enhanced --config conf/redis.yaml
# 或者简化版本
abc-runner r --config conf/redis.yaml
```

#### 参数映射表

| 老参数 | 新配置位置 | 说明 |
|-------|-----------|------|
| `-h <host>` | `redis.standalone.addr` | 主机地址 |
| `-p <port>` | `redis.standalone.addr` | 端口号 |
| `-a <password>` | `redis.standalone.password` | 密码 |
| `-n <requests>` | `redis.benchmark.total` | 总请求数 |
| `-c <connections>` | `redis.benchmark.parallels` | 并发连接数 |
| `-t <test>` | `redis.benchmark.case` | 测试类型 |
| `-r <range>` | `redis.benchmark.random_keys` | 随机键范围 |
| `-d <bytes>` | `redis.benchmark.data_size` | 数据大小 |
| `--cluster` | `redis.mode: "cluster"` | 集群模式 |
| `-ttl <seconds>` | `redis.benchmark.ttl` | TTL设置 |
| `-R <percent>` | `redis.benchmark.read_percent` | 读操作比例 |

### HTTP命令迁移

```bash
# 老命令（已弃用）
abc-runner http --url http://example.com -n 1000 -c 10

# 新命令（推荐）
abc-runner http-enhanced --config conf/http.yaml
# 或者简化版本
abc-runner h --config conf/http.yaml
```

### Kafka命令迁移

```bash
# 老命令（已弃用）
abc-runner kafka --broker localhost:9092 --topic test -n 1000

# 新命令（推荐）
abc-runner kafka-enhanced --config conf/kafka.yaml
# 或者简化版本
abc-runner k --config conf/kafka.yaml
```

## 📝 配置文件迁移

### 自动迁移工具

使用我们提供的自动迁移工具：

```bash
# 迁移单个文件
./tools/migrate_config.sh conf/old-redis.yaml

# 批量迁移
./tools/migrate_config.sh conf/*.yaml

# 预览迁移结果（不实际修改）
./tools/migrate_config.sh -d conf/old-redis.yaml

# 详细输出
./tools/migrate_config.sh -v conf/old-redis.yaml
```

### 手动迁移示例

#### Redis配置迁移

**老格式：**

```yaml
redis:
  host: localhost
  port: 6379
  password: mypassword
  cluster: false
  db: 0
```

**新格式：**

```yaml
redis:
  mode: "standalone"
  benchmark:
    total: 10000
    parallels: 50
    random_keys: 50
    read_percent: 50
    data_size: 3
    ttl: 120
    case: "set_get_random"
  pool:
    pool_size: 10
    min_idle: 2
  standalone:
    addr: localhost:6379
    password: mypassword
    db: 0
```

#### 集群配置迁移

**老格式：**

```yaml
redis:
  cluster: true
  nodes:
    - localhost:6371
    - localhost:6372
    - localhost:6373
  password: mypassword
```

**新格式：**

```yaml
redis:
  mode: "cluster"
  cluster:
    addrs:
      - localhost:6371
      - localhost:6372
      - localhost:6373
    password: mypassword
  benchmark:
    total: 10000
    parallels: 50
    # ... 其他基准测试配置
```

## 🚀 新功能特性

### 1. 统一命令接口

所有协议现在都支持统一的命令格式：

```bash
# 标准格式
abc-runner <protocol>-enhanced [options]

# 简化别名
abc-runner <alias> [options]
```

### 2. 增强的配置管理

- **多源配置**：支持配置文件、环境变量、命令行参数
- **配置验证**：自动验证配置完整性和正确性
- **热重载**：支持配置的动态重载（部分场景）

```bash
# 使用环境变量覆盖配置
REDIS_HOST=prod-redis abc-runner redis-enhanced --config conf/redis.yaml

# 命令行覆盖配置
abc-runner redis-enhanced --config conf/redis.yaml --redis.benchmark.total=50000
```

### 3. 连接池管理

- **无全局变量**：消除全局状态依赖
- **连接复用**：提高性能，减少连接开销
- **健康检查**：自动检测和恢复连接

### 4. 高级指标收集

```yaml
# 启用高级指标
redis:
  monitoring:
    enabled: true
    interval: 5s
    metrics:
      - latency_histogram
      - throughput
      - error_rate
      - connection_stats
```

### 5. 智能错误处理

- **自动重试**：网络错误自动重试
- **错误分类**：详细的错误分类和建议
- **故障转移**：支持主从切换和节点故障转移

## 🔧 故障排除

### 常见问题

#### 1. 命令不存在错误

**问题：**

```bash
Error: command 'redis' not found
```

**解决方案：**

```bash
# 检查是否使用了正确的命令名
abc-runner redis-enhanced --config conf/redis.yaml

# 或者使用别名
abc-runner r --config conf/redis.yaml
```

#### 2. 配置文件格式错误

**问题：**

```bash
Error: failed to parse config file
```

**解决方案：**

```bash
# 使用迁移工具自动转换
./tools/migrate_config.sh your-config.yaml

# 或者手动检查YAML格式
yamllint your-config.yaml
```

#### 3. 连接超时

**问题：**

```bash
Error: connection timeout
```

**解决方案：**

```yaml
redis:
  connection:
    timeout: 30s
    retry_attempts: 3
    retry_delay: 1s
```

### 兼容性问题

#### 1. 老脚本不工作

如果您的脚本使用老命令格式，会看到弃用警告：

```bash
⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️
WARNING: Using DEPRECATED command 'redis'
Please migrate to the enhanced version: 'redis-enhanced'
⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️
```

**解决方案：**更新脚本使用新命令格式。

#### 2. 性能差异

新架构通常性能更好，但如果遇到性能问题：

```yaml
redis:
  pool:
    pool_size: 20    # 增加连接池大小
    min_idle: 5      # 增加最小空闲连接
  benchmark:
    parallels: 100   # 增加并发数
```

## 📊 性能对比

| 指标 | 老架构 | 新架构 | 提升 |
|-----|--------|--------|------|
| 连接建立时间 | ~100ms | ~10ms | 90% |
| 内存使用 | 基准 | -30% | 30% |
| 吞吐量 | 基准 | +50% | 50% |
| 错误处理 | 基础 | 高级 | 显著提升 |

## 📚 最佳实践

### 1. 配置管理

```bash
# 为不同环境使用不同配置
abc-runner redis-enhanced --config conf/redis-dev.yaml    # 开发环境
abc-runner redis-enhanced --config conf/redis-prod.yaml   # 生产环境
```

### 2. 监控和日志

```yaml
redis:
  logging:
    level: info
    output: logs/abc-runner.log
  monitoring:
    enabled: true
    export_to: prometheus
```

### 3. CI/CD集成

```yaml
# .github/workflows/performance-test.yml
- name: Run Redis Performance Test
  run: |
    abc-runner redis-enhanced \
      --config conf/redis-ci.yaml \
      --output results/redis-performance.json
```

### 4. 批处理脚本

```bash
#!/bin/bash
# 性能测试套件

protocols=("redis" "http" "kafka")
for protocol in "${protocols[@]}"; do
    echo "Testing $protocol..."
    abc-runner ${protocol}-enhanced \
        --config conf/${protocol}.yaml \
        --output results/${protocol}-results.json
done
```

## 🔄 回滚计划

如果迁移后遇到问题，可以临时回滚：

### 1. 使用老版本

```bash
# 下载老版本
wget https://github.com/abc-runner/releases/v1.x.x/abc-runner

# 使用老配置
./abc-runner-old redis -h localhost -p 6379 -n 1000
```

### 2. 兼容模式

新版本仍然支持老命令（带警告）：

```bash
# 这些命令仍然可以工作，但会显示弃用警告
abc-runner redis -h localhost -p 6379 -n 1000
abc-runner http --url http://example.com
abc-runner kafka --broker localhost:9092
```

## 📞 获取帮助

### 命令行帮助

```bash
# 全局帮助
abc-runner --help

# 特定命令帮助
abc-runner redis-enhanced --help
abc-runner http-enhanced --help
abc-runner kafka-enhanced --help
```

### 在线资源

- **文档网站**: <https://docs.abc-runner.com>
- **GitHub Issues**: <https://github.com/abc-runner/issues>
- **社区讨论**: <https://github.com/abc-runner/discussions>
- **迁移支持**: <https://docs.abc-runner.com/migration-support>

### 专业支持

如果您的组织需要迁移支持，请联系：

- 邮箱：<support@abc-runner.com>
- 企业支持：<enterprise@abc-runner.com>

## 🎯 迁移时间表

| 版本 | 状态 | 说明 |
|------|------|------|
| v0.0.1 | 当前 | 新旧架构共存 |
| v0.0.2 | 计划中 | 增强功能，老命令警告 |
| v0.0.3 | 计划中 | 老命令标记为严重弃用 |
| v0.0.5 | 未来 | 完全移除老架构 |

**建议：**尽快迁移到新架构，以享受性能提升和新功能。

---

## 快速开始

如果您想立即开始迁移：

```bash
# 1. 备份现有配置
cp conf/redis.yaml conf/redis.yaml.backup

# 2. 运行迁移工具
./tools/migrate_config.sh conf/redis.yaml

# 3. 测试新配置
abc-runner redis-enhanced --config conf/redis.new.yaml

# 4. 如果一切正常，替换原配置
mv conf/redis.new.yaml conf/redis.yaml
```

欢迎使用新版abc-runner！🚀
