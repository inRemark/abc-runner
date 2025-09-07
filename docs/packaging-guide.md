# 打包管理使用指南

## 概述

本文档介绍了如何使用 Makefile 中的 `release` 目标来创建 redis-runner 的发布包。

## 打包流程

打包流程会执行以下操作：

1. 清理之前的构建产物
2. 安装/更新依赖
3. 为所有支持的平台构建二进制文件
4. 创建版本目录结构
5. 复制二进制文件、配置文件和文档到版本目录
6. 为每个平台创建压缩包

## 使用方法

### 创建默认版本的发布包

```bash
make release
```

此命令将创建 v0.2.0 版本的发布包。

### 创建指定版本的发布包

```bash
VERSION=1.0.0 make release
```

此命令将创建 v1.0.0 版本的发布包。

## 输出结构

打包完成后，将在 `releases/` 目录下生成以下内容：

```
releases/
├── v0.2.0/                    # 版本目录
│   ├── redis-runner-darwin-amd64     # macOS AMD64 二进制文件
│   ├── redis-runner-darwin-arm64     # macOS ARM64 二进制文件
│   ├── redis-runner-linux-amd64      # Linux AMD64 二进制文件
│   ├── redis-runner-linux-arm64      # Linux ARM64 二进制文件
│   ├── redis-runner-windows-amd64.exe # Windows AMD64 二进制文件
│   ├── config/                       # 配置文件目录
│   │   ├── redis.yaml
│   │   ├── http.yaml
│   │   └── kafka.yaml
│   ├── README.md                     # 项目说明
│   └── LICENSE                       # 许可证文件
├── redis-runner-v0.2.0-darwin-amd64.tar.gz     # 平台特定压缩包
├── redis-runner-v0.2.0-darwin-arm64.tar.gz
├── redis-runner-v0.2.0-linux-amd64.tar.gz
├── redis-runner-v0.2.0-linux-arm64.tar.gz
└── redis-runner-v0.2.0-windows-amd64.zip
```

## 使用发布包

用户可以下载对应平台的压缩包，解压后即可使用：

1. 下载对应平台的压缩包
2. 解压压缩包
3. 根据需要修改 `config/` 目录中的配置文件
4. 运行对应平台的二进制文件

### 示例（macOS）

```bash
# 下载并解压
tar -xzf redis-runner-v0.2.0-darwin-amd64.tar.gz

# 编辑配置文件（可选）
nano config/redis.yaml

# 运行
./redis-runner-darwin-amd64 redis --help
```

## 支持的平台

- macOS AMD64 (darwin/amd64)
- macOS ARM64 (darwin/arm64)
- Linux AMD64 (linux/amd64)
- Linux ARM64 (linux/arm64)
- Windows AMD64 (windows/amd64)

## 版本管理

项目使用语义化版本控制（SemVer）：
- 主版本号：不兼容的API修改
- 次版本号：向下兼容的功能性新增
- 修订号：向下兼容的问题修正