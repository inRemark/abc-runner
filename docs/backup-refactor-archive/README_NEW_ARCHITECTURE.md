# Redis Runner - 重构后的架构指南

## 项目概述

Redis Runner 是一个高性能的 Redis 基准测试工具，经过全面重构后采用了模块化、可扩展的架构设计。新架构消除了全局变量依赖，支持多种配置源，具备完善的错误处理和重试机制。

## 新架构特性

### 🏗️ 模块化架构

- **协议适配器**：统一的协议接口，支持 Redis、HTTP、Kafka 等
- **配置管理**：多源配置加载（命令行、环境变量、YAML文件）
- **连接管理**：无全局变量的连接池管理
- **操作注册**：可扩展的操作类型注册机制
- **错误处理**：带重试和熔断器的错误处理

### 📊 增强监控

- **性能指标**：RPS、延迟分布、成功率等
- **系统监控**：内存使用、GC统计、goroutine数量
- **协议指标**：连接池状态、特定操作统计
- **实时进度**：带ETA的进度显示

### 🔧 配置方式

支持多种配置方式，按优先级排序：

1. 命令行参数（最高优先级）
2. 环境变量
3. YAML配置文件（最低优先级）

## 快速开始

### 使用命令行参数

```bash
# 基本用法
./redis-runner redis -h localhost -p 6379 -n 10000 -c 50 -t set_get_random

# 集群模式
./redis-runner redis --cluster -h localhost -p 6371 -a password -n 50000 -c 100 -d 64 -R 70 -ttl 300 -t set_get_random

# 完整参数示例
./redis-runner redis \
  -h 127.0.0.1 \
  -p 6379 \
  -a "mypassword" \
  -n 100000 \
  -c 50 \
  -d 64 \
  -r 1000 \
  -R 80 \
  -ttl 120 \
  -db 0 \
  -t set_get_random
```

### 使用配置文件

创建 `conf/redis.yaml` 文件：

```yaml
redis:
  mode: "standalone"    # standalone, sentinel, cluster
  benchmark:
    total: 10000              # 总请求数
    parallels: 50             # 并发连接数
    random_keys: 50           # 随机键范围 (0表示递增键)
    read_percent: 50          # 读操作百分比
    data_size: 3              # 数据大小(字节)
    ttl: 120                  # TTL(秒)
    case: "set_get_random"    # 测试用例
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

然后运行：

```bash
./redis-runner redis --config
```

### 使用环境变量

```bash
export REDIS_RUNNER_MODE=cluster
export REDIS_RUNNER_TOTAL=50000
export REDIS_RUNNER_PARALLELS=100
export REDIS_RUNNER_ADDRS="127.0.0.1:6371,127.0.0.1:6372,127.0.0.1:6373"
export REDIS_RUNNER_PASSWORD="mypassword"
export REDIS_RUNNER_CASE=set_get_random

./redis-runner redis
```

## 支持的操作类型

| 操作类型 | 说明 | 示例 |
|---------|------|------|
| `get` | GET操作 | 从已生成键中随机读取 |
| `set` | SET操作 | 写入键值对 |
| `del` | DELETE操作 | 删除键 |
| `set_get_random` | 混合读写 | 根据读写比例执行操作 |
| `hget` | Hash GET | Hash表读取 |
| `hset` | Hash SET | Hash表写入 |
| `pub` | 发布消息 | 发布到Redis频道 |
| `sub` | 订阅消息 | 订阅Redis频道 |

## 参数说明

### 连接参数

- `-h, --host`: Redis服务器地址 (默认: 127.0.0.1)
- `-p, --port`: Redis服务器端口 (默认: 6379)
- `-a, --auth`: Redis密码
- `--cluster`: 启用集群模式
- `-db`: 数据库编号 (默认: 0, 集群模式下忽略)

### 测试参数

- `-n`: 总请求数 (默认: 100000)
- `-c`: 并发连接数 (默认: 50)
- `-d`: 数据大小(字节) (默认: 3)
- `-r`: 随机键范围 (0表示递增键)
- `-R`: 读操作百分比 (默认: 50)
- `-ttl`: 键的TTL(秒) (默认: 120)
- `-t`: 测试用例类型

### 配置参数

- `--config`: 使用配置文件模式

## 输出示例

```bash
============================================================
REDIS BENCHMARK RESULTS
============================================================
Test Case: set_get_random
Total Requests: 10000
Parallel Connections: 50
RPS: 15432
Success Rate: 99.98%

------------------------------------------------------------
Avg Latency: 3.245 ms
P95 Latency: 8.234 ms
P99 Latency: 15.678 ms

============================================================
BENCHMARK COMPLETED
============================================================
```

## 新架构开发指南

### 添加新的操作类型

1. 实现 `OperationFactory` 接口：

```go
type MyOperationFactory struct{}

func (f *MyOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
    // 创建操作逻辑
    return interfaces.Operation{
        Type: "my_operation",
        Key:  "test_key",
        // ... 其他字段
    }, nil
}

func (f *MyOperationFactory) GetOperationType() string {
    return "my_operation"
}

func (f *MyOperationFactory) ValidateParams(params map[string]interface{}) error {
    // 参数验证逻辑
    return nil
}
```

2. 在适配器中添加执行逻辑：

```go
func (r *RedisAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
    switch operation.Type {
    case "my_operation":
        return r.executeMyOperation(ctx, operation)
    // ... 其他case
    }
}
```

3. 注册操作工厂：

```go
registry.Register("my_operation", &MyOperationFactory{})
```

### 自定义错误处理

```go
// 添加自定义错误分类规则
errorHandler.errorClassifier.AddRule(func(err error) *ErrorInfo {
    if strings.Contains(err.Error(), "my_custom_error") {
        return &ErrorInfo{
            Type:        "custom",
            Retryable:   true,
            Recoverable: false,
            Severity:    SeverityHigh,
        }
    }
    return nil
})

// 添加自定义恢复策略
errorHandler.recoveryManager.RegisterStrategy("custom", func(ctx context.Context, err error, adapter interfaces.ProtocolAdapter) error {
    // 自定义恢复逻辑
    return nil
})
```

### 扩展性能监控

```go
// 记录自定义指标
metricsCollector.RecordProtocolMetric("custom_metric", value)

// 获取增强指标
enhancedMetrics := metricsCollector.GetEnhancedMetrics()
systemHealth := metricsCollector.GetSystemHealth()
```

## 性能调优建议

### 连接池配置

```yaml
pool:
  pool_size: 100        # 根据并发数调整
  min_idle: 10          # 保持最小连接
  max_idle: 50          # 最大空闲连接
  idle_timeout: 300s    # 空闲超时
  connection_timeout: 30s # 连接超时
```

### 系统资源优化

- **内存**: 监控堆内存使用，避免频繁GC
- **Goroutine**: 控制并发数，避免goroutine泄露
- **网络**: 合理设置连接池大小和超时时间

### 测试场景优化

- **数据大小**: 根据实际业务调整 `-d` 参数
- **读写比例**: 使用 `-R` 参数模拟真实负载
- **键分布**: 使用 `-r` 参数控制键的分布模式

## 故障排除

### 常见问题

1. **连接失败**

   ```bash
   Error: failed to connect to Redis: dial tcp: connection refused
   ```

   - 检查Redis服务是否运行
   - 验证地址和端口配置
   - 检查防火墙设置

2. **认证失败**

   ```bash
   Error: authentication failed
   ```

   - 验证密码配置
   - 检查Redis AUTH配置

3. **性能异常**
   - 查看系统资源使用情况
   - 检查网络延迟
   - 调整并发数和连接池配置

### 调试模式

设置环境变量启用详细日志：

```bash
export REDIS_RUNNER_DEBUG=true
export REDIS_RUNNER_LOG_LEVEL=debug
```

## 更新日志

### v2.0.0 (重构版本)

- ✅ 全新模块化架构
- ✅ 消除全局变量依赖
- ✅ 多源配置支持
- ✅ 增强错误处理和重试机制
- ✅ 完善的性能监控
- ✅ 可扩展的操作注册机制
- ✅ 系统资源监控
- ✅ 单元测试覆盖

### 与v1.0的对比

| 特性 | v1.0 | v2.0 |
|------|------|------|
| 架构 | 单体式 | 模块化 |
| 全局变量 | 是 | 否 |
| 配置方式 | 单一 | 多源 |
| 错误处理 | 基础 | 增强 |
| 监控指标 | 基础 | 全面 |
| 扩展性 | 有限 | 高度可扩展 |
| 测试覆盖 | 无 | 全面 |

## 贡献指南

欢迎提交Issue和Pull Request！

### 开发环境

- Go 1.22.6+
- Redis 6.0+

### 提交要求

- 代码通过所有测试
- 添加相应的单元测试
- 更新相关文档

## 许可证

MIT License
