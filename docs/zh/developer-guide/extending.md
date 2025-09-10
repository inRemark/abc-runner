# 扩展abc-runner

[English](../en/developer-guide/extending.md) | [中文](extending.md)

abc-runner设计为可扩展的架构，允许开发者添加新的协议支持、操作类型和功能模块。

## 架构概述

abc-runner采用插件化架构，核心组件包括：

1. **命令路由器**: 负责解析和路由命令
2. **协议适配器**: 为不同协议提供统一接口
3. **配置管理器**: 管理配置加载和验证
4. **操作注册表**: 管理可用的操作类型
5. **运行引擎**: 执行基准测试
6. **报告管理器**: 生成和输出测试报告

## 添加新协议

### 1. 创建协议适配器

在`app/adapters/`目录下创建新的协议适配器：

```go
// app/adapters/myprotocol/adapter.go
package myprotocol

import (
    "context"
    "abc-runner/app/core/interfaces"
)

type MyProtocolAdapter struct {
    // 适配器字段
}

func NewMyProtocolAdapter() *MyProtocolAdapter {
    return &MyProtocolAdapter{}
}

func (a *MyProtocolAdapter) Connect(ctx context.Context, config interfaces.Config) error {
    // 实现连接逻辑
    return nil
}

func (a *MyProtocolAdapter) Close() error {
    // 实现关闭逻辑
    return nil
}

func (a *MyProtocolAdapter) ExecuteOperation(ctx context.Context, op interfaces.Operation) (interface{}, error) {
    // 实现操作执行逻辑
    return nil, nil
}

func (a *MyProtocolAdapter) GetMetricsCollector() interfaces.MetricsCollector {
    // 返回指标收集器
    return nil
}
```

### 2. 实现配置管理

```go
// app/adapters/myprotocol/config/config.go
package config

import (
    "abc-runner/app/core/interfaces"
)

type MyProtocolConfig struct {
    // 配置字段
}

func (c *MyProtocolConfig) GetBenchmark() interfaces.BenchmarkConfig {
    // 实现基准测试配置接口
    return nil
}

func (c *MyProtocolConfig) GetConnection() interfaces.ConnectionConfig {
    // 实现连接配置接口
    return nil
}
```

### 3. 注册命令处理器

```go
// app/commands/myprotocol.go
package commands

import (
    "context"
    "abc-runner/app/adapters/myprotocol"
)

type MyProtocolCommandHandler struct {
    adapter *myprotocol.MyProtocolAdapter
}

func NewMyProtocolCommandHandler() *MyProtocolCommandHandler {
    return &MyProtocolCommandHandler{
        adapter: myprotocol.NewMyProtocolAdapter(),
    }
}

func (h *MyProtocolCommandHandler) Execute(ctx context.Context, args []string) error {
    // 实现命令执行逻辑
    return nil
}

func (h *MyProtocolCommandHandler) GetHelp() string {
    // 返回帮助信息
    return ""
}
```

### 4. 注册到命令路由器

在`main.go`中注册新的命令处理器：

```go
// 注册MyProtocol命令
myProtocolHandler := commands.NewMyProtocolCommandHandler()
commandRouter.RegisterCommand("myprotocol", myProtocolHandler)
commandRouter.RegisterAlias("mp", "myprotocol")
```

## 添加新操作类型

### 1. 创建操作工厂

```go
// app/adapters/redis/operations.go
type MyOperationFactory struct{}

func (f *MyOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
    // 创建操作实例
    return interfaces.Operation{}, nil
}

func (f *MyOperationFactory) GetOperationType() string {
    return "my_operation"
}

func (f *MyOperationFactory) ValidateParams(params map[string]interface{}) error {
    // 验证参数
    return nil
}
```

### 2. 注册操作工厂

```go
// 在RegisterRedisOperations函数中添加
registry.Register("my_operation", &MyOperationFactory{})
```

## 实现自定义报告格式

### 1. 创建报告生成器

```go
// app/core/reports/myreport.go
package reports

import (
    "abc-runner/app/core/interfaces"
)

type MyReportGenerator struct {
    metrics interfaces.MetricsCollector
}

func NewMyReportGenerator(metrics interfaces.MetricsCollector) *MyReportGenerator {
    return &MyReportGenerator{metrics: metrics}
}

func (r *MyReportGenerator) Generate() error {
    // 实现报告生成逻辑
    return nil
}
```

### 2. 集成到报告管理器

```go
// 在ReportManager中添加支持
func (rm *ReportManager) GenerateMyReport() error {
    generator := myreport.NewMyReportGenerator(rm.metricsCollector)
    return generator.Generate()
}
```

## 添加配置选项

### 1. 扩展配置结构

```go
// 在相应的配置结构中添加新字段
type ExtendedConfig struct {
    NewOption string `yaml:"new_option" json:"new_option"`
}
```

### 2. 实现配置验证

```go
func (c *ExtendedConfig) Validate() error {
    // 验证新选项
    if c.NewOption == "" {
        return fmt.Errorf("new_option is required")
    }
    return nil
}
```

## 实现自定义指标收集

### 1. 实现MetricsCollector接口

```go
type CustomMetricsCollector struct {
    // 自定义指标字段
}

func (c *CustomMetricsCollector) RecordOperation(start time.Time, err error) {
    // 实现指标记录逻辑
}

func (c *CustomMetricsCollector) Export() map[string]interface{} {
    // 导出指标数据
    return nil
}
```

### 2. 集成到适配器

```go
func (a *MyProtocolAdapter) GetMetricsCollector() interfaces.MetricsCollector {
    return &CustomMetricsCollector{}
}
```

## 最佳实践

### 代码组织

1. **模块化**: 将相关功能组织在同一个包中
2. **接口**: 使用接口定义契约，提高可测试性
3. **依赖注入**: 通过构造函数注入依赖
4. **错误处理**: 正确处理和返回错误

### 测试

1. **单元测试**: 为每个函数编写单元测试
2. **表驱动测试**: 使用表驱动方式测试多种情况
3. **mock**: 使用mock对象测试依赖
4. **集成测试**: 编写集成测试验证组件交互

### 文档

1. **代码注释**: 为导出的函数和类型添加注释
2. **使用示例**: 提供使用示例
3. **API文档**: 文档化公共API
4. **扩展指南**: 提供扩展开发指南

### 性能

1. **并发安全**: 确保并发访问的安全性
2. **内存管理**: 避免内存泄漏
3. **资源清理**: 及时释放资源
4. **性能测试**: 进行性能基准测试

## 调试和故障排除

### 日志记录

使用标准日志包记录调试信息：

```go
import "log"

log.Printf("Debug: %v", debugInfo)
```

### 性能分析

使用Go的pprof进行性能分析：

```go
import _ "net/http/pprof"

// 在main函数中启动pprof服务器
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

### 单元测试调试

使用Delve调试器进行调试：

```bash
dlv test ./app/adapters/myprotocol/
```

## 示例扩展

### 添加MongoDB支持

1. 创建MongoDB适配器
2. 实现连接和操作执行逻辑
3. 添加配置管理
4. 注册命令处理器
5. 编写测试用例

### 添加GraphQL支持

1. 创建GraphQL适配器
2. 实现查询和变更操作
3. 添加schema验证
4. 集成到命令系统
5. 提供使用示例

## 贡献扩展

如果您开发了有用的扩展：

1. **开源**: 将扩展开源到GitHub
2. **文档**: 提供详细的使用文档
3. **示例**: 提供使用示例
4. **测试**: 包含完整的测试套件
5. **贡献**: 考虑贡献到主项目