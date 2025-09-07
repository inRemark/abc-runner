# Contributing Guide

[English](contributing.md) | [中文](contributing.zh.md)

Thank you for your interest in contributing to redis-runner! This document provides guidelines and best practices for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/redis-runner.git`
3. Create a new branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Commit your changes: `git commit -am 'Add some feature'`
6. Push to the branch: `git push origin feature/your-feature-name`
7. Create a new Pull Request

## Development Environment Setup

### Prerequisites

- Go 1.22 or higher
- Redis server (for testing)
- Kafka cluster (for testing)
- HTTP server (for testing)

### Building

```bash
# Clone the repository
git clone https://github.com/your-org/redis-runner.git
cd redis-runner

# Build the binary
go build -o redis-runner .
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

## Code Style and Standards

### Go Code Standards

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use `golint` and `govet` to check for issues
- Write meaningful commit messages

### Documentation

- Update documentation when adding new features
- Write clear, concise comments for exported functions and types
- Keep README and other documentation up to date

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build
2. Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters
3. Increase the version numbers in any examples files and the README.md to the new version that this Pull Request would represent
4. Your Pull Request will be reviewed by maintainers, who may request changes
5. Once approved, your Pull Request will be merged

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. Version of redis-runner you are using
2. Operating system and architecture
3. Exact steps to reproduce the issue
4. Expected behavior
5. Actual behavior
6. Any relevant logs or error messages

### Feature Requests

When requesting features, please include:

1. Description of the feature
2. Use case for the feature
3. Potential implementation approach (if you have one)
4. Benefits of the feature

## Architecture Guidelines

### Adding New Protocols

To add support for a new protocol:

1. Create a new adapter in `app/adapters/`
2. Implement the `ProtocolAdapter` interface
3. Add configuration structures in `app/core/config/`
4. Register the adapter in the runner
5. Add command-line interface support
6. Write tests for the new adapter
7. Document the new protocol

### Extending Existing Protocols

When extending existing protocols:

1. Follow the existing code patterns and conventions
2. Maintain backward compatibility when possible
3. Add comprehensive tests for new functionality
4. Update documentation accordingly

## Testing Guidelines

### Unit Tests

- Write unit tests for all new functionality
- Aim for high code coverage (>80%)
- Test edge cases and error conditions
- Use table-driven tests where appropriate

### Integration Tests

- Write integration tests for protocol adapters
- Test against real services when possible
- Include tests for configuration loading
- Test error handling and recovery

### Performance Tests

- Include benchmarks for performance-critical code
- Test with various load patterns
- Measure memory allocation and CPU usage
- Compare performance before and after changes

## Code of Conduct

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.
