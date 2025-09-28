# 多协议服务端辅助模块

本目录包含为 abc-runner 性能测试工具设计的多协议服务端辅助模块，用于提供标准化的测试环境。

## 架构概述

### 目录结构

```
servers/
├── README.md                   # 本文档
├── cmd/                        # 服务端可执行程序
│   ├── http-server/           # HTTP服务端程序
│   ├── tcp-server/            # TCP服务端程序
│   ├── udp-server/            # UDP服务端程序
│   ├── grpc-server/           # gRPC服务端程序
│   └── multi-server/          # 多协议统一启动器
├── internal/                   # 内部共享模块
│   ├── config/                # 统一配置管理
│   ├── monitoring/            # 监控和健康检查
│   ├── logging/               # 日志管理
│   └── common/                # 通用工具
├── pkg/                       # 公共服务端接口
│   ├── interfaces/            # 服务端接口定义
│   ├── http/                  # HTTP服务端模块
│   ├── tcp/                   # TCP服务端模块
│   ├── udp/                   # UDP服务端模块
│   └── grpc/                  # gRPC服务端模块
├── config/                    # 配置文件
│   ├── servers/               # 各协议服务端配置
│   └── examples/              # 配置示例
├── scripts/                   # 启动和管理脚本
│   ├── start-all.sh          # 启动所有服务端
│   ├── stop-all.sh           # 停止所有服务端
│   └── health-check.sh       # 健康检查脚本
├── docs/                      # 文档
│   ├── deployment.md         # 部署指南
│   ├── configuration.md      # 配置说明
│   └── api.md                # API文档
└── test/                      # 集成测试
    ├── integration/           # 集成测试
    └── performance/           # 性能测试
```

### 设计原则

1. **独立性**: 每个服务端模块独立运行，可单独启动
2. **简单性**: 专注于测试功能，避免复杂业务逻辑
3. **可配置性**: 支持灵活的配置选项
4. **监控性**: 提供服务运行状态监控
5. **扩展性**: 支持未来功能扩展

## 支持的协议

### HTTP服务端
- 支持 GET/POST/PUT/DELETE 等常用HTTP方法
- 支持 JSON/XML/文本响应格式
- 可配置响应延迟和状态码
- 支持负载模拟

### TCP服务端
- 支持多连接并发处理
- 提供回显服务功能
- 支持单向/双向通信模式
- 支持长连接 Keep-Alive

### UDP服务端
- 支持单播/组播/广播响应
- 支持数据包验证
- 可配置丢包率模拟
- 支持最大数据包大小限制

### gRPC服务端
- 支持一元调用、服务端流、客户端流、双向流
- 支持TLS安全传输
- 支持Token认证
- 符合gRPC标准协议

## 快速开始

### 启动所有服务端
```bash
cd servers
./scripts/start-all.sh
```

### 启动单个服务端
```bash
# HTTP服务端
./cmd/http-server/http-server --config config/servers/http-server.yaml

# TCP服务端  
./cmd/tcp-server/tcp-server --config config/servers/tcp-server.yaml

# UDP服务端
./cmd/udp-server/udp-server --config config/servers/udp-server.yaml

# gRPC服务端
./cmd/grpc-server/grpc-server --config config/servers/grpc-server.yaml
```

### 健康检查
```bash
./scripts/health-check.sh
```

## 配置说明

每个服务端都有独立的配置文件，支持以下配置类别：
- 网络配置（监听地址、端口）
- 协议配置（协议特定参数）
- 性能配置（并发数、缓冲区大小）
- 监控配置（健康检查、指标暴露）
- 日志配置（日志级别、输出位置）

详细配置说明请参考 [configuration.md](docs/configuration.md)

## 测试集成

这些服务端模块作为 abc-runner 的测试目标，为以下测试场景提供支持：
- 协议适配器功能测试
- 性能压力测试
- 网络连接测试
- 错误处理测试
- 并发能力测试

## 监控和日志

所有服务端提供统一的监控和日志接口：
- 健康检查端点
- 运行状态指标
- 结构化日志输出
- 性能统计信息

## 部署说明

详细的部署指南请参考 [deployment.md](docs/deployment.md)

## 开发指南

如需扩展或修改服务端功能，请参考现有代码结构和接口定义。