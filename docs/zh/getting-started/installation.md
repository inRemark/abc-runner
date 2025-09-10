# 安装指南

[English](../en/getting-started/installation.md) | [中文](installation.md)

## 系统要求

- Go 1.25或更高版本（从源码构建）
- 支持的操作系统：Linux、macOS、Windows

## 安装方式

### 1. 从源码构建

#### 克隆仓库

```bash
git clone https://github.com/your-org/abc-runner.git
cd abc-runner
```

#### 构建二进制文件

```bash
# 使用Makefile构建
make build

# 或者直接使用Go构建
go build -o abc-runner .

# 构建所有平台的二进制文件
make build-all
```

### 2. 使用预编译二进制文件

从[GitHub发布页面](https://github.com/your-org/abc-runner/releases)下载适合您系统的预编译二进制文件。

#### Linux

```bash
wget https://github.com/your-org/abc-runner/releases/download/v0.2.0/abc-runner-linux-amd64
chmod +x abc-runner-linux-amd64
sudo mv abc-runner-linux-amd64 /usr/local/bin/abc-runner
```

#### macOS

```bash
wget https://github.com/your-org/abc-runner/releases/download/v0.2.0/abc-runner-darwin-amd64
chmod +x abc-runner-darwin-amd64
sudo mv abc-runner-darwin-amd64 /usr/local/bin/abc-runner
```

#### Windows

下载`abc-runner-windows-amd64.exe`文件并将其重命名为`abc-runner.exe`，然后将其添加到系统PATH中。

### 3. 使用Go安装

```bash
go install github.com/your-org/abc-runner@latest
```

## 验证安装

```bash
# 检查版本
abc-runner --version

# 显示帮助
abc-runner --help
```

## 配置文件

安装完成后，您可以复制配置模板到您的工作目录：

```bash
# 复制配置模板
cp -r config/* conf/
```

或者使用Makefile命令：

```bash
make config
```

## 环境变量

abc-runner支持以下环境变量：

- `ABC_RUNNER_CONFIG`: 配置文件路径
- `ABC_RUNNER_LOG_LEVEL`: 日志级别 (debug, info, warn, error)

## 升级

要升级到新版本：

1. 备份您的配置文件
2. 下载新版本的二进制文件或重新构建
3. 替换旧的二进制文件
4. 根据需要更新配置文件