# Contribution Guide

[English](contributing.md) | [中文](../zh/developer-guide/contributing.md)

Thank you for considering contributing to the redis-runner project! We welcome all forms of contributions, including code, documentation, bug reports, and feature suggestions.

## Code of Conduct

Please adhere to our [Code of Conduct](code-of-conduct.md) to ensure that all participants can work in an open and friendly environment.

## Ways to Contribute

### Reporting Bugs

When reporting bugs, please include the following information:

1. **Version Information**: redis-runner version
2. **Environment Information**: Operating system, Go version, etc.
3. **Reproduction Steps**: Detailed steps to reproduce
4. **Expected Behavior**: What you expected to see
5. **Actual Behavior**: What actually happened
6. **Log Information**: Relevant log output

### Submitting Feature Suggestions

When submitting feature suggestions, please include:

1. **Problem Description**: The problem you want to solve
2. **Proposed Solution**: Your suggested solution
3. **Alternative Solutions**: Other solutions you've considered
4. **Additional Information**: Any relevant additional information

### Code Contributions

#### Development Environment Setup

1. **Clone Repository**:
   ```bash
   git clone https://github.com/your-org/redis-runner.git
   cd redis-runner
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run Tests**:
   ```bash
   make test
   ```

#### Branch Strategy

We use the GitFlow branch strategy:

- **main**: Stable version branch
- **develop**: Development branch
- **feature/***: Feature development branches
- **hotfix/***: Emergency fix branches
- **release/***: Release preparation branches

#### Commit Message Guidelines

Please follow these commit message guidelines:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- feat: New feature
- fix: Bug fix
- docs: Documentation update
- style: Code formatting adjustments
- refactor: Code refactoring
- perf: Performance optimization
- test: Test-related
- chore: Build process or auxiliary tool changes

**Example**:
```
feat(redis): add support for Redis Streams

Implement Redis Streams operations including XADD, XREAD, and XLEN.
Add unit tests for all new operations.

Closes #123
```

#### Code Style

1. **Go Formatting**: Use `go fmt` to format code
2. **Naming Conventions**: Follow Go naming conventions
3. **Comments**: Add comments to exported functions and types
4. **Error Handling**: Properly handle and return errors
5. **Testing**: Add unit tests for new features

#### Pull Request Process

1. **Create Branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Development and Testing**:
   ```bash
   # Write code
   # Run tests
   make test
   ```

3. **Commit Changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

4. **Push Branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create Pull Request**:
   - Create PR on GitHub
   - Fill in PR description
   - Link related issues

6. **Code Review**:
   - Wait for review feedback
   - Make modifications based on feedback
   - Push changes again

7. **Merge**:
   - After PR is approved, maintainers will merge to develop branch

### Documentation Contributions

#### Documentation Structure

Documentation is organized in the following structure:

```
docs/
├── getting-started/     # Quick start guide
├── user-guide/          # User manual
├── developer-guide/     # Developer guide
├── architecture/        # Architecture design documents
├── configuration/       # Configuration management documents
├── deployment/          # Deployment and operations documents
├── changelog/           # Change logs
└── faq/                 # Frequently asked questions
```

#### Documentation Format

1. **Markdown**: Use standard Markdown format
2. **Code Examples**: Provide runnable code examples
3. **Links**: Use relative links to reference other documents
4. **Images**: Place images in the `docs/images/` directory

### Test Contributions

#### Test Types

1. **Unit Tests**: Test individual functions or methods
2. **Integration Tests**: Test interactions between components
3. **Performance Tests**: Test performance metrics
4. **End-to-End Tests**: Test complete workflows

#### Test Tools

1. **Go Testing**: Use Go's built-in testing framework
2. **Testify**: Use Testify assertion library
3. **Mockery**: Generate mock objects

#### Test Coverage

Target test coverage rate is 80% or above.

### Environment Contributions

#### Docker

1. **Dockerfile**: Update Docker image build files
2. **docker-compose**: Update example orchestration files

#### Kubernetes

1. **Helm Chart**: Update Helm Charts
2. **YAML Files**: Update Kubernetes resource configurations

## Release Process

### Version Number Specification

We follow the [Semantic Versioning](https://semver.org/) specification:

- **Major Version**: Incompatible API modifications
- **Minor Version**: Backward-compatible functional additions
- **Patch Version**: Backward-compatible bug fixes

### Release Steps

1. **Create Release Branch**:
   ```bash
   git checkout -b release/vX.Y.Z
   ```

2. **Update Version Number**:
   - Update `VERSION` file
   - Update `CHANGELOG.md`
   - Update version references in documentation

3. **Create Tag**:
   ```bash
   git tag -a vX.Y.Z -m "Release version X.Y.Z"
   ```

4. **Push Tag**:
   ```bash
   git push origin vX.Y.Z
   ```

5. **Create GitHub Release**:
   - Create Release on GitHub
   - Upload pre-compiled binaries
   - Publish to various platforms

## Community Participation

### Discussion

- **GitHub Issues**: Technical discussions and issue tracking
- **GitHub Discussions**: General discussions and Q&A
- **Slack**: Real-time communication (link in README)

### Meetings

- **Monthly Development Meetings**: Discuss development progress and plans
- **Community Meetings**: User feedback and requirement collection

## Acknowledging Contributors

We will acknowledge all contributors in README and contributor lists.

## Contact Us

If you have any questions, please contact us through the following methods:

- GitHub Issues: [https://github.com/your-org/redis-runner/issues](https://github.com/your-org/redis-runner/issues)
- Email: maintainers@redis-runner.org