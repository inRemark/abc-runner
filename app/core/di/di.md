# DI

## 架构

DI容器纯净：

✅ 不依赖任何具体适配器实现
✅ 只依赖抽象接口和核心组件
✅ 移除了所有硬编码的具体类型

运行时依赖注入：

✅ 在 main.go 应用层创建具体实现
✅ 通过 container.Provide() 运行时注册
✅ 通过 container.Invoke() 真正使用DI

架构清晰：

✅ 核心层：纯净的DI容器，只管理抽象依赖
✅ 应用层：决定具体实现，负责运行时注入
✅ 适配器层：独立的具体实现，不被DI容器直接依赖

```bash
应用层 (main.go)
├── 创建具体实现 (CustomAdapterFactory, EnhancedMetricsCollector)
├── 运行时注册到DI容器 (container.Provide)
└── 使用DI获取依赖 (container.Invoke)

核心DI层 (di/container.go)  
├── 只依赖抽象接口 (interfaces.*)
├── 管理抽象依赖关系
└── 不知道具体实现

适配器层 (adapters/*)
├── 独立的具体实现
└── 不被DI容器直接引用
```
