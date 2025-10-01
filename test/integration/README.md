# abc-runner 集成测试指南

## 概述

此目录包含abc-runner项目的集成测试用例，用于验证不同协议适配器之间的协作能力和系统整体功能。

## 测试结构

### 主要测试文件

- `adapter_integration_test.go` - 多协议适配器集成测试
- `simple_integration_test.go` - 简单集成测试示例

### 测试覆盖范围

#### 协议适配器测试
- **Redis适配器** - 数据存储和缓存操作测试
- **HTTP适配器** - Web服务请求和响应测试  
- **TCP适配器** - 低层网络连接和数据传输测试

#### 核心功能测试
- **配置验证** - 各协议配置的正确性验证
- **指标收集** - 性能指标收集和统计验证
- **接口兼容性** - 适配器接口实现完整性检查

## 运行测试

### 运行所有集成测试
```bash
# 从项目根目录运行
make integration-test

# 或直接使用go test
go test -v ./test/integration/...
```

### 运行特定测试
```bash
# 运行适配器集成测试
go test -v ./test/integration/ -run TestAdapterIntegration

# 运行配置验证测试
go test -v ./test/integration/ -run TestConfigValidation

# 运行指标收集测试
go test -v ./test/integration/ -run TestMetricsIntegration
```

### 运行性能基准测试
```bash
# 运行适配器创建性能测试
go test -v ./test/integration/ -bench BenchmarkAdapterCreation

# 查看详细基准测试结果
go test -v ./test/integration/ -bench . -benchmem
```

## 测试环境要求

### 可选服务（测试会优雅处理服务不可用情况）

1. **Redis服务器** (可选)
   ```bash
   # 使用Docker启动Redis
   docker run -d -p 6379:6379 redis:latest
   
   # 或本地安装Redis
   brew install redis
   brew services start redis
   ```

2. **HTTP测试服务** (可选)
   - 默认使用 `https://httpbin.org` 进行HTTP测试
   - 可替换为本地HTTP服务器

3. **TCP测试服务** (可选)
   ```bash
   # 使用nc创建简单TCP服务器
   nc -l 8080
   
   # 或使用telnet测试
   telnet localhost 8080
   ```

## 测试场景

### 1. 连接测试
- 验证各协议适配器的连接建立能力
- 测试连接失败时的错误处理
- 检查连接资源的正确释放

### 2. 健康检查测试
- 验证适配器的健康状态监控
- 测试服务不可用时的状态报告
- 检查健康检查的响应时间

### 3. 配置验证测试
- 验证各协议配置参数的正确性
- 测试无效配置的错误处理
- 检查配置克隆和序列化功能

### 4. 指标收集测试
- 验证操作指标的准确记录
- 测试指标快照的完整性
- 检查性能统计的计算准确性

### 5. 接口兼容性测试
- 验证所有适配器实现统一接口
- 测试接口方法的返回值正确性
- 检查接口契约的遵守情况

## 测试输出示例

```
=== RUN   TestAdapterIntegration
=== RUN   TestAdapterIntegration/Redis_Adapter_Integration
    adapter_integration_test.go:49: Redis connection failed (expected if no server): dial tcp [::1]:6379: connect: connection refused
    adapter_integration_test.go:62: ✅ Redis adapter interface validation passed
    adapter_integration_test.go:70: ✅ Redis adapter integration test completed
=== RUN   TestAdapterIntegration/HTTP_Adapter_Integration
    adapter_integration_test.go:95: ✅ HTTP connection successful
    adapter_integration_test.go:102: ✅ HTTP health check passed
    adapter_integration_test.go:119: ✅ HTTP operation successful
    adapter_integration_test.go:125: ✅ HTTP adapter interface validation passed
    adapter_integration_test.go:127: ✅ HTTP adapter integration test completed
=== RUN   TestAdapterIntegration/TCP_Adapter_Integration
    adapter_integration_test.go:156: TCP connection failed (expected if no server): dial tcp [::1]:8080: connect: connection refused
    adapter_integration_test.go:169: ✅ TCP adapter interface validation passed
    adapter_integration_test.go:171: ✅ TCP adapter integration test completed
--- PASS: TestAdapterIntegration (2.45s)
```

## 扩展测试

### 添加新的协议测试
1. 在 `adapter_integration_test.go` 中添加新的测试函数
2. 按照现有模式实现适配器测试逻辑
3. 更新测试文档和运行说明

### 自定义测试配置
- 修改测试中的服务地址和端口
- 调整测试参数（并发数、操作数等）
- 添加特定场景的测试用例

## 最佳实践

1. **测试独立性** - 每个测试应该能独立运行，不依赖其他测试的状态
2. **资源清理** - 确保测试后正确释放连接和资源
3. **错误容忍** - 测试应该优雅处理外部服务不可用的情况
4. **日志记录** - 提供清晰的测试日志，便于问题定位
5. **性能监控** - 包含基准测试，监控性能回归

## 故障排除

### 常见问题

1. **连接失败**
   - 检查目标服务是否启动
   - 验证网络连通性和防火墙设置
   - 确认端口号和地址配置正确

2. **测试超时**
   - 调整测试上下文的超时时间
   - 检查网络延迟和服务响应时间
   - 减少测试并发数和操作数

3. **指标验证失败**
   - 确保指标收集器正确初始化
   - 检查操作结果的记录逻辑
   - 验证指标计算的准确性

### 调试技巧

- 使用 `-v` 选项查看详细测试输出
- 添加更多日志记录来跟踪测试执行
- 使用调试器逐步执行测试代码
- 检查测试环境的依赖服务状态