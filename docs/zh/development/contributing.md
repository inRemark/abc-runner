# 贡献指南

[English](contributing.md) | [中文](contributing.zh.md)

感谢您有兴趣为abc-runner做贡献！本文档提供了贡献项目的指南和最佳实践。

## 开始贡献

1. Fork仓库
2. 克隆您的fork：`git clone https://github.com/your-username/abc-runner.git`
3. 创建新分支：`git checkout -b feature/your-feature-name`
4. 进行您的修改
5. 提交更改：`git commit -am 'Add some feature'`
6. 推送到分支：`git push origin feature/your-feature-name`
7. 创建新的Pull Request

## 开发环境设置

### 先决条件

- Go 1.22或更高版本
- Redis服务器（用于测试）
- Kafka集群（用于测试）
- HTTP服务器（用于测试）

### 构建

```bash
# 克隆仓库
git clone https://github.com/your-org/abc-runner.git
cd abc-runner

# 构建二进制文件
go build -o abc-runner .
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -cover ./...

# 运行集成测试
go test -tags=integration ./...
```

## 代码风格和标准

### Go代码标准

- 遵循官方的[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用`gofmt`格式化代码
- 使用`golint`和`govet`检查问题
- 编写有意义的提交信息

### 文档

- 添加新功能时更新文档
- 为导出的函数和类型编写清晰简洁的注释
- 保持README和其他文档的更新

## Pull Request流程

1. 确保在构建结束前移除任何安装或构建依赖
2. 更新README.md，详细说明接口的更改，包括新的环境变量、暴露的端口、有用的文件位置和容器参数
3. 增加示例文件和README.md中的版本号，以反映此Pull Request所代表的新版本
4. 您的Pull Request将由维护者审查，他们可能会要求更改
5. 一旦获得批准，您的Pull Request将被合并

## 报告问题

### Bug报告

报告bug时，请包含以下内容：

1. 您使用的abc-runner版本
2. 操作系统和架构
3. 重现问题的确切步骤
4. 预期行为
5. 实际行为
6. 任何相关的日志或错误信息

### 功能请求

请求功能时，请包含以下内容：

1. 功能描述
2. 功能的使用场景
3. 潜在的实现方法（如果您有的话）
4. 功能的好处

## 架构指南

### 添加新协议

要添加对新协议的支持：

1. 在`app/adapters/`中创建新的适配器
2. 实现`ProtocolAdapter`接口
3. 在运行器中添加配置结构
4. 注册适配器
5. 添加命令行接口支持
6. 为新适配器编写测试
7. 记录新协议

### 扩展现有协议

扩展现有协议时：

1. 遵循现有的代码模式和约定
2. 尽可能保持向后兼容性
3. 为新功能添加全面的测试
4. 相应地更新文档

## 测试指南

### 单元测试

- 为所有新功能编写单元测试
- 目标是高代码覆盖率（>80%）
- 测试边界情况和错误条件
- 在适当时使用表驱动测试

### 集成测试

- 为协议适配器编写集成测试
- 尽可能针对真实服务进行测试
- 包含配置加载的测试
- 测试错误处理和恢复

### 性能测试

- 为性能关键代码包含基准测试
- 使用各种负载模式进行测试
- 测量内存分配和CPU使用率
- 比较更改前后的性能

## 行为准则

请注意，该项目发布了贡献者行为准则。参与此项目即表示您同意遵守其条款。