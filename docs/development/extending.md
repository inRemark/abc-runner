# Extending redis-runner

This guide explains how to extend redis-runner to support additional protocols or features.

## Architecture Overview

redis-runner follows a modular architecture based on the Adapter pattern. Each protocol (Redis, HTTP, Kafka) is implemented as a separate adapter that conforms to a common interface.

## Adding a New Protocol

To add support for a new protocol, you need to:

1. Create a new adapter
2. Implement the ProtocolAdapter interface
3. Register the adapter with the runner
4. Add command-line interface support
5. Add configuration support
6. Write tests
7. Document the new protocol

### 1. Create a New Adapter

Create a new directory under `app/adapters/` for your protocol:

```bash
app/adapters/myprotocol/
├── adapter.go          # Main adapter implementation
├── config/             # Configuration structures
│   ├── config.go
│   ├── interfaces.go
│   └── loader.go
├── connection/         # Connection management
│   ├── client.go
│   └── pool.go
├── metrics/            # Metrics collection
│   ├── collector.go
│   └── reporter.go
├── operations/         # Protocol operations
│   ├── factory.go
│   └── operations.go
└── test/               # Tests
    ├── adapter_test.go
    └── integration_test.go
```

### 2. Implement the ProtocolAdapter Interface

Your adapter must implement the `ProtocolAdapter` interface defined in `app/core/interfaces/adapter.go`:

```go
type ProtocolAdapter interface {
    // Initialize the adapter with configuration
    Init(config Config) error
    
    // Establish connections
    Connect() error
    
    // Execute a single operation
    Execute(operation Operation) (success bool, isRead bool, duration time.Duration, err error)
    
    // Close connections
    Close() error
    
    // Get protocol-specific metrics
    GetMetrics() map[string]interface{}
    
    // Health check
    HealthCheck() error
}
```

### 3. Register the Adapter

Register your adapter in the runner by adding it to the adapter registry in `app/commands/myprotocol.go`:

```go
func init() {
    runner.RegisterAdapter("myprotocol", NewMyProtocolAdapter)
}
```

### 4. Add Command-Line Interface Support

Create a new command file in `app/commands/myprotocol.go`:

```go
package commands

import (
    "github.com/urfave/cli/v2"
    "your-project/app/core/runner"
)

func NewMyProtocolCommand() *cli.Command {
    return &cli.Command{
        Name:  "myprotocol",
        Usage: "Run myprotocol performance tests",
        Flags: []cli.Flag{
            // Define your protocol-specific flags here
        },
        Action: func(c *cli.Context) error {
            // Handle command execution
            return runner.RunMyProtocolTest(c)
        },
    }
}
```

### 5. Add Configuration Support

Create configuration structures in `app/adapters/myprotocol/config/`:

```go
type MyProtocolConfig struct {
    Host     string        `yaml:"host"`
    Port     int           `yaml:"port"`
    Timeout  time.Duration `yaml:"timeout"`
    // Add other configuration fields
}

func (c *MyProtocolConfig) Validate() error {
    // Validate configuration
    if c.Host == "" {
        return errors.New("host is required")
    }
    return nil
}
```

### 6. Write Tests

Write comprehensive tests for your adapter:

- Unit tests for each component
- Integration tests against a real service
- Benchmark tests for performance-critical code

### 7. Document the New Protocol

Add documentation for your new protocol:

- Update README.md with usage examples
- Create usage documentation in `docs/usage/myprotocol.md`
- Add configuration examples

## Extending Existing Protocols

### Adding New Operations

To add new operations to an existing protocol:

1. Add the operation to the operations factory
2. Implement the operation logic
3. Update configuration if needed
4. Write tests
5. Document the new operation

### Adding Configuration Options

To add new configuration options:

1. Add fields to the configuration struct
2. Add validation logic
3. Update configuration loading
4. Update command-line flags
5. Update documentation

## Best Practices

### Error Handling

- Use descriptive error messages
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Handle timeouts and network errors gracefully
- Provide meaningful error codes when appropriate

### Performance Considerations

- Minimize memory allocations
- Use connection pooling
- Implement efficient serialization/deserialization
- Use appropriate concurrency patterns
- Profile performance-critical code

### Resource Management

- Always close connections and clean up resources
- Use context for cancellation and timeouts
- Implement proper shutdown procedures
- Handle resource leaks in error conditions

### Testing

- Write tests for all public functions
- Use table-driven tests for multiple test cases
- Mock external dependencies
- Test error conditions
- Use integration tests for end-to-end validation

## Example Implementation

For a complete example of how to implement a new protocol, refer to the existing adapters:

- Redis adapter: `app/adapters/redis/`
- HTTP adapter: `app/adapters/http/`
- Kafka adapter: `app/adapters/kafka/`

These implementations demonstrate best practices for adapter design and can serve as templates for new protocols.
