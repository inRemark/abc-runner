# 扩展redis-runner

[English](extending.md) | [中文](extending.zh.md)

本指南解释了如何扩展redis-runner以支持额外的协议或功能。

## 架构概述

redis-runner遵循基于适配器模式的模块化架构。每种协议（Redis、HTTP、Kafka）都实现为一个符合通用接口的独立适配器。

## 添加新协议

要添加对新协议的支持，您需要：

1. 创建新适配器
2. 实现ProtocolAdapter接口
3. 向运行器注册适配器
4. 添加命令行接口支持
5. 添加配置支持
6. 编写测试
7. 记录新协议

### 1. 创建新适配器

在`app/adapters/`下为您的协议创建新目录：

```bash
app/adapters/myprotocol/
├── adapter.go          # 主适配器实现
├── config/             # 配置结构
│   ├── config.go
│   ├── interfaces.go
│   └── loader.go
├── connection/         # 连接管理
│   ├── client.go
│   └── pool.go
├── metrics/            # 指标收集
│   ├── collector.go
│   └── reporter.go
├── operations/         # 协议操作
│   ├── factory.go
│   └── operations.go
└── test/               # 测试
    ├── adapter_test.go
    └── integration_test.go
```

### 2. 实现ProtocolAdapter接口

您的适配器必须实现`app/core/interfaces/adapter.go`中定义的`ProtocolAdapter`接口：

```go
type ProtocolAdapter interface {
    // 使用配置初始化适配器
    Init(config Config) error
    
    // 建立连接
    Connect() error
    
    // 执行单个操作
    Execute(operation Operation) (success bool, isRead bool, duration time.Duration, err error)
    
    // 关闭连接
    Close() error
    
    // 获取协议特定指标
    GetMetrics() map[string]interface{}
    
    // 健康检查
    HealthCheck() error
}
```

### 3. 注册适配器

通过在`app/commands/myprotocol.go`中的运行器适配器注册表中添加适配器来注册您的适配器：

```go
func init() {
    runner.RegisterAdapter("myprotocol", NewMyProtocolAdapter)
}
```

### 4. 添加命令行接口支持

在`app/commands/myprotocol.go`中创建新命令文件：

```go
package commands

import (
    "github.com/urfave/cli/v2"
    "your-project/app/core/runner"
)

func NewMyProtocolCommand() *cli.Command {
    return &cli.Command{
        Name:  "myprotocol",
        Usage: "Run myprotocol performance tests",
        Flags: []cli.Flag{
            // 在此处定义您的协议特定标志
        },
        Action: func(c *cli.Context) error {
            // 处理命令执行
            return runner.RunMyProtocolTest(c)
        },
    }
}
```

### 5. 添加配置支持

在`app/adapters/myprotocol/config/`中创建配置结构：

```go
type MyProtocolConfig struct {
    Host     string        `yaml:"host"`
    Port     int           `yaml:"port"`
    Timeout  time.Duration `yaml:"timeout"`
    // 添加其他配置字段
}

func (c *MyProtocolConfig) Validate() error {
    // 验证配置
    if c.Host == "" {
        return errors.New("host is required")
    }
    return nil
}
```

### 6. 编写测试

为您的适配器编写全面的测试：

- 每个组件的单元测试
- 针对真实服务的集成测试
- 性能关键代码的基准测试

### 7. 记录新协议

为您的新协议添加文档：

- 更新README.md，包含使用示例
- 在`docs/usage/myprotocol.md`中创建使用文档
- 添加配置示例

## 扩展现有协议

### 添加新操作

要向现有协议添加新操作：

1. 将操作添加到操作工厂
2. 实现操作逻辑
3. 如需要，更新配置
4. 编写测试
5. 记录新操作

### 添加配置选项

要添加新配置选项：

1. 向配置结构添加字段
2. 添加验证逻辑
3. 更新配置加载
4. 更新命令行标志
5. 更新文档

## 最佳实践

### 错误处理

- 使用描述性错误消息
- 使用`fmt.Errorf("context: %w", err)`包装带上下文的错误
- 优雅地处理超时和网络错误
- 在适当时提供有意义的错误代码

### 性能考虑

- 最小化内存分配
- 使用连接池
- 实现高效的序列化/反序列化
- 使用适当的并发模式
- 对性能关键代码进行性能分析

### 资源管理

- 始终关闭连接并清理资源
- 使用上下文进行取消和超时
- 实现正确的关闭程序
- 处理错误条件下的资源泄漏

### 测试

- 为所有公共函数编写测试
- 使用表驱动测试处理多个测试用例
- 模拟外部依赖
- 测试错误条件
- 使用集成测试进行端到端验证

## 示例实现

有关如何实现新协议的完整示例，请参考现有适配器：

- Redis适配器：`app/adapters/redis/`
- HTTP适配器：`app/adapters/http/`
- Kafka适配器：`app/adapters/kafka/`

这些实现展示了适配器设计的最佳实践，可以作为新协议的模板。