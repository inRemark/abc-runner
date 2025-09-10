# Installation Guide

[English](installation.md) | [中文](../zh/getting-started/installation.md)

## System Requirements

- Go 1.25 or higher (for building from source)
- Supported operating systems: Linux, macOS, Windows

## Installation Methods

### 1. Building from Source

#### Clone Repository

```bash
git clone https://github.com/your-org/abc-runner.git
cd abc-runner
```

#### Build Binary

```bash
# Use Makefile to build
make build

# Or build directly with Go
go build -o abc-runner .

# Build binaries for all platforms
make build-all
```

### 2. Using Pre-compiled Binaries

Download pre-compiled binaries suitable for your system from the [GitHub Releases page](https://github.com/your-org/abc-runner/releases).

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

Download the `abc-runner-windows-amd64.exe` file, rename it to `abc-runner.exe`, and add it to your system PATH.

### 3. Using Go Install

```bash
go install github.com/your-org/abc-runner@latest
```

## Verifying Installation

```bash
# Check version
abc-runner --version

# Show help
abc-runner --help
```

## Configuration Files

After installation, you can copy configuration templates to your working directory:

```bash
# Copy configuration templates
cp -r config/* conf/
```

Or use the Makefile command:

```bash
make config
```

## Environment Variables

abc-runner supports the following environment variables:

- `ABC_RUNNER_CONFIG`: Configuration file path
- `ABC_RUNNER_LOG_LEVEL`: Log level (debug, info, warn, error)

## Upgrading

To upgrade to a new version:

1. Backup your configuration files
2. Download the new version binaries or rebuild
3. Replace the old binaries
4. Update configuration files as needed