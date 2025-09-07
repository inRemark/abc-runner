# Extending redis-runner

[English](extending.md) | [中文](../zh/developer-guide/extending.md)

redis-runner is designed with an extensible architecture that allows developers to add new protocol support, operation types, and functional modules.

## Architecture Overview

redis-runner adopts a plugin architecture with core components including:

1. **Command Router**: Responsible for parsing and routing commands
2. **Protocol Adapters**: Provide unified interfaces for different protocols
3. **Configuration Manager**: Manages configuration loading and validation
4. **Operation Registry**: Manages available operation types
5. **Execution Engine**: Executes benchmark tests
6. **Report Manager**: Generates and outputs test reports

## Adding New Protocols

### 1. Create Protocol Adapter

Create a new protocol adapter in the `app/adapters/` directory:

```go
// app/adapters/myprotocol/adapter.go
package myprotocol

import (
    "context"
    "redis-runner/app/core/interfaces"
)

type MyProtocolAdapter struct {
    // Adapter fields
}

func NewMyProtocolAdapter() *MyProtocolAdapter {
    return &MyProtocolAdapter{}
}

func (a *MyProtocolAdapter) Connect(ctx context.Context, config interfaces.Config) error {
    // Implement connection logic
    return nil
}

func (a *MyProtocolAdapter) Close() error {
    // Implement close logic
    return nil
}

func (a *MyProtocolAdapter) ExecuteOperation(ctx context.Context, op interfaces.Operation) (interface{}, error) {
    // Implement operation execution logic
    return nil, nil
}

func (a *MyProtocolAdapter) GetMetricsCollector() interfaces.MetricsCollector {
    // Return metrics collector
    return nil
}
```

### 2. Implement Configuration Management

```go
// app/adapters/myprotocol/config/config.go
package config

import (
    "redis-runner/app/core/interfaces"
)

type MyProtocolConfig struct {
    // Configuration fields
}

func (c *MyProtocolConfig) GetBenchmark() interfaces.BenchmarkConfig {
    // Implement benchmark configuration interface
    return nil
}

func (c *MyProtocolConfig) GetConnection() interfaces.ConnectionConfig {
    // Implement connection configuration interface
    return nil
}
```

### 3. Register Command Handler

```go
// app/commands/myprotocol.go
package commands

import (
    "context"
    "redis-runner/app/adapters/myprotocol"
)

type MyProtocolCommandHandler struct {
    adapter *myprotocol.MyProtocolAdapter
}

func NewMyProtocolCommandHandler() *MyProtocolCommandHandler {
    return &MyProtocolCommandHandler{
        adapter: myprotocol.NewMyProtocolAdapter(),
    }
}

func (h *MyProtocolCommandHandler) Execute(ctx context.Context, args []string) error {
    // Implement command execution logic
    return nil
}

func (h *MyProtocolCommandHandler) GetHelp() string {
    // Return help information
    return ""
}
```

### 4. Register with Command Router

Register the new command handler in `main.go`:

```go
// Register MyProtocol command
myProtocolHandler := commands.NewMyProtocolCommandHandler()
commandRouter.RegisterCommand("myprotocol", myProtocolHandler)
commandRouter.RegisterAlias("mp", "myprotocol")
```

## Adding New Operation Types

### 1. Create Operation Factory

```go
// app/adapters/redis/operations.go
type MyOperationFactory struct{}

func (f *MyOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
    // Create operation instance
    return interfaces.Operation{}, nil
}

func (f *MyOperationFactory) GetOperationType() string {
    return "my_operation"
}

func (f *MyOperationFactory) ValidateParams(params map[string]interface{}) error {
    // Validate parameters
    return nil
}
```

### 2. Register Operation Factory

```go
// In RegisterRedisOperations function
registry.Register("my_operation", &MyOperationFactory{})
```

## Implementing Custom Report Formats

### 1. Create Report Generator

```go
// app/core/reports/myreport.go
package reports

import (
    "redis-runner/app/core/interfaces"
)

type MyReportGenerator struct {
    metrics interfaces.MetricsCollector
}

func NewMyReportGenerator(metrics interfaces.MetricsCollector) *MyReportGenerator {
    return &MyReportGenerator{metrics: metrics}
}

func (r *MyReportGenerator) Generate() error {
    // Implement report generation logic
    return nil
}
```

### 2. Integrate with Report Manager

```go
// In ReportManager, add support
func (rm *ReportManager) GenerateMyReport() error {
    generator := myreport.NewMyReportGenerator(rm.metricsCollector)
    return generator.Generate()
}
```

## Adding Configuration Options

### 1. Extend Configuration Structure

```go
// Add new fields to the corresponding configuration structure
type ExtendedConfig struct {
    NewOption string `yaml:"new_option" json:"new_option"`
}
```

### 2. Implement Configuration Validation

```go
func (c *ExtendedConfig) Validate() error {
    // Validate new options
    if c.NewOption == "" {
        return fmt.Errorf("new_option is required")
    }
    return nil
}
```

## Implementing Custom Metrics Collection

### 1. Implement MetricsCollector Interface

```go
type CustomMetricsCollector struct {
    // Custom metrics fields
}

func (c *CustomMetricsCollector) RecordOperation(start time.Time, err error) {
    // Implement metrics recording logic
}

func (c *CustomMetricsCollector) Export() map[string]interface{} {
    // Export metrics data
    return nil
}
```

### 2. Integrate with Adapter

```go
func (a *MyProtocolAdapter) GetMetricsCollector() interfaces.MetricsCollector {
    return &CustomMetricsCollector{}
}
```

## Best Practices

### Code Organization

1. **Modularity**: Organize related functionality in the same package
2. **Interfaces**: Use interfaces to define contracts, improving testability
3. **Dependency Injection**: Inject dependencies through constructors
4. **Error Handling**: Properly handle and return errors

### Testing

1. **Unit Tests**: Write unit tests for each function
2. **Table-Driven Tests**: Use table-driven approach to test multiple cases
3. **Mocking**: Use mock objects to test dependencies
4. **Integration Tests**: Write integration tests to verify component interactions

### Documentation

1. **Code Comments**: Add comments to exported functions and types
2. **Usage Examples**: Provide usage examples
3. **API Documentation**: Document public APIs
4. **Extension Guide**: Provide extension development guide

### Performance

1. **Concurrency Safety**: Ensure safety of concurrent access
2. **Memory Management**: Avoid memory leaks
3. **Resource Cleanup**: Release resources in a timely manner
4. **Performance Testing**: Conduct performance benchmark testing

## Debugging and Troubleshooting

### Logging

Use the standard log package to record debugging information:

```go
import "log"

log.Printf("Debug: %v", debugInfo)
```

### Performance Analysis

Use Go's pprof for performance analysis:

```go
import _ "net/http/pprof"

// In main function, start pprof server
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

### Unit Test Debugging

Use Delve debugger for debugging:

```bash
dlv test ./app/adapters/myprotocol/
```

## Example Extensions

### Adding MongoDB Support

1. Create MongoDB adapter
2. Implement connection and operation execution logic
3. Add configuration management
4. Register command handler
5. Write test cases

### Adding GraphQL Support

1. Create GraphQL adapter
2. Implement query and mutation operations
3. Add schema validation
4. Integrate with command system
5. Provide usage examples

## Contributing Extensions

If you develop useful extensions:

1. **Open Source**: Open source the extension to GitHub
2. **Documentation**: Provide detailed usage documentation
3. **Examples**: Provide usage examples
4. **Testing**: Include comprehensive test suite
5. **Contribution**: Consider contributing to the main project