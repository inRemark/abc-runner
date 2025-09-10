# Packaging Management Guide

[English](packaging-guide.md) | [中文](packaging-guide.zh.md)

## Overview

This document describes how to use the `release` target in the Makefile to create abc-runner release packages.

## Packaging Process

The packaging process performs the following operations:

1. Clean up previous build artifacts
2. Install/update dependencies
3. Build binaries for all supported platforms
4. Create version directory structure
5. Copy binaries, configuration files, and documentation to the version directory
6. Create archives for each platform

## Usage

### Create Release Package with Default Version

```bash
make release
```

This command will create a release package for version v0.2.0.

### Create Release Package with Specified Version

```bash
VERSION=1.0.0 make release
```

This command will create a release package for version v1.0.0.

## Output Structure

After packaging is complete, the following content will be generated in the `releases/` directory:

```
releases/
├── v0.2.0/                    # Version directory
│   ├── abc-runner-darwin-amd64     # macOS AMD64 binary
│   ├── abc-runner-darwin-arm64     # macOS ARM64 binary
│   ├── abc-runner-linux-amd64      # Linux AMD64 binary
│   ├── abc-runner-linux-arm64      # Linux ARM64 binary
│   ├── abc-runner-windows-amd64.exe # Windows AMD64 binary
│   ├── config/                       # Configuration directory
│   │   ├── redis.yaml
│   │   ├── http.yaml
│   │   └── kafka.yaml
│   ├── README.md                     # Project README
│   └── LICENSE                       # License file
├── abc-runner-v0.2.0-darwin-amd64.tar.gz     # Platform-specific archive
├── abc-runner-v0.2.0-darwin-arm64.tar.gz
├── abc-runner-v0.2.0-linux-amd64.tar.gz
├── abc-runner-v0.2.0-linux-arm64.tar.gz
└── abc-runner-v0.2.0-windows-amd64.zip
```

## Using Release Packages

Users can download the archive for their platform and use it after extraction:

1. Download the archive for your platform
2. Extract the archive
3. Modify the configuration files in the `config/` directory as needed
4. Run the binary for your platform

### Example (macOS)

```bash
# Download and extract
tar -xzf abc-runner-v0.2.0-darwin-amd64.tar.gz

# Edit configuration files (optional)
nano config/redis.yaml

# Run
./abc-runner-darwin-amd64 redis --help
```

## Supported Platforms

- macOS AMD64 (darwin/amd64)
- macOS ARM64 (darwin/arm64)
- Linux AMD64 (linux/amd64)
- Linux ARM64 (linux/arm64)
- Windows AMD64 (windows/amd64)

## Version Management

The project uses Semantic Versioning (SemVer):
- Major version: Incompatible API changes
- Minor version: Backward compatible feature additions
- Patch version: Backward compatible bug fixes