# 贡献指南

[English](../en/developer-guide/contributing.md) | [中文](contributing.md)

感谢您考虑为redis-runner项目做贡献！我们欢迎各种形式的贡献，包括代码、文档、bug报告和功能建议。

## 行为准则

请遵守我们的[行为准则](code-of-conduct.md)，确保所有参与者都能在一个开放和友好的环境中工作。

## 贡献方式

### 报告Bug

在报告bug时，请包含以下信息：

1. **版本信息**: redis-runner的版本
2. **环境信息**: 操作系统、Go版本等
3. **复现步骤**: 详细的复现步骤
4. **期望行为**: 您期望看到什么
5. **实际行为**: 实际发生了什么
6. **日志信息**: 相关的日志输出

### 提交功能建议

在提交功能建议时，请包含：

1. **问题描述**: 您想要解决的问题
2. **解决方案**: 您建议的解决方案
3. **替代方案**: 您考虑过的其他方案
4. **附加信息**: 任何相关的附加信息

### 代码贡献

#### 开发环境设置

1. **克隆仓库**:
   ```bash
   git clone https://github.com/your-org/redis-runner.git
   cd redis-runner
   ```

2. **安装依赖**:
   ```bash
   go mod tidy
   ```

3. **运行测试**:
   ```bash
   make test
   ```

#### 分支策略

我们使用GitFlow分支策略：

- **main**: 稳定版本分支
- **develop**: 开发分支
- **feature/***: 功能开发分支
- **hotfix/***: 紧急修复分支
- **release/***: 发布准备分支

#### 提交信息规范

请遵循以下提交信息规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**:
- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- perf: 性能优化
- test: 测试相关
- chore: 构建过程或辅助工具的变动

**示例**:
```
feat(redis): add support for Redis Streams

Implement Redis Streams operations including XADD, XREAD, and XLEN.
Add unit tests for all new operations.

Closes #123
```

#### 代码风格

1. **Go格式化**: 使用`go fmt`格式化代码
2. **命名规范**: 遵循Go命名约定
3. **注释**: 为导出的函数和类型添加注释
4. **错误处理**: 正确处理和返回错误
5. **测试**: 为新功能添加单元测试

#### Pull Request流程

1. **创建分支**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **开发和测试**:
   ```bash
   # 编写代码
   # 运行测试
   make test
   ```

3. **提交更改**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

4. **推送分支**:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **创建Pull Request**:
   - 在GitHub上创建PR
   - 填写PR描述
   - 关联相关issue

6. **代码审查**:
   - 等待审查反馈
   - 根据反馈进行修改
   - 重新推送更改

7. **合并**:
   - PR被批准后，维护者会合并到develop分支

### 文档贡献

#### 文档结构

文档按以下结构组织：

```
docs/
├── getting-started/     # 快速入门指南
├── user-guide/          # 用户使用手册
├── developer-guide/     # 开发者指南
├── architecture/        # 架构设计文档
├── configuration/       # 配置管理文档
├── deployment/          # 部署和运维文档
├── changelog/           # 变更日志
└── faq/                 # 常见问题解答
```

#### 文档格式

1. **Markdown**: 使用标准Markdown格式
2. **代码示例**: 提供可运行的代码示例
3. **链接**: 使用相对链接引用其他文档
4. **图片**: 将图片放在`docs/images/`目录下

### 测试贡献

#### 测试类型

1. **单元测试**: 测试单个函数或方法
2. **集成测试**: 测试组件间的交互
3. **性能测试**: 测试性能指标
4. **端到端测试**: 测试完整的工作流程

#### 测试工具

1. **Go测试**: 使用Go内置测试框架
2. **Testify**: 使用Testify断言库
3. **Mockery**: 生成mock对象

#### 测试覆盖率

目标测试覆盖率达到80%以上。

### 环境贡献

#### Docker

1. **Dockerfile**: 更新Docker镜像构建文件
2. **docker-compose**: 更新示例编排文件

#### Kubernetes

1. **Helm Chart**: 更新Helm Chart
2. **YAML文件**: 更新Kubernetes资源配置

## 发布流程

### 版本号规范

我们遵循[语义化版本控制](https://semver.org/lang/zh-CN/)规范：

- **主版本号**: 不兼容的API修改
- **次版本号**: 向后兼容的功能性新增
- **修订号**: 向后兼容的问题修正

### 发布步骤

1. **创建发布分支**:
   ```bash
   git checkout -b release/vX.Y.Z
   ```

2. **更新版本号**:
   - 更新`VERSION`文件
   - 更新`CHANGELOG.md`
   - 更新文档中的版本引用

3. **创建标签**:
   ```bash
   git tag -a vX.Y.Z -m "Release version X.Y.Z"
   ```

4. **推送标签**:
   ```bash
   git push origin vX.Y.Z
   ```

5. **创建GitHub Release**:
   - 在GitHub上创建Release
   - 上传预编译二进制文件
   - 发布到各平台

## 社区参与

### 讨论

- **GitHub Issues**: 技术讨论和问题跟踪
- **GitHub Discussions**: 一般性讨论和问答
- **Slack**: 实时交流（链接在README中）

### 会议

- **月度开发会议**: 讨论开发进展和计划
- **社区会议**: 用户反馈和需求收集

## 认可贡献者

我们会在README和贡献者列表中认可所有贡献者。

## 联系我们

如有任何问题，请通过以下方式联系我们：

- GitHub Issues: [https://github.com/your-org/redis-runner/issues](https://github.com/your-org/redis-runner/issues)
- Email: maintainers@redis-runner.org