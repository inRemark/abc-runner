# Component Documentation

## Core Components

### 1. Runner (`app/core/runner`)

The Enhanced Runner is the heart of the redis-runner tool. It coordinates the entire testing process:

- Manages test execution lifecycle
- Controls concurrency and parallelism
- Coordinates between different components
- Handles test scheduling and progress tracking

Key features:

- Configurable parallel execution
- Real-time progress monitoring
- Graceful shutdown handling
- Error propagation and handling

### 2. Configuration Manager (`app/core/config`)

The Configuration Manager handles all aspects of configuration:

- Loading from multiple sources (CLI, YAML, environment)
- Validation of configuration values
- Providing default values
- Managing configuration precedence

Key components:

- `ConfigManager`: Main configuration coordinator
- `MultiSourceLoader`: Loads config from multiple sources
- `ConfigSources`: Defines configuration sources and precedence

### 3. Report Manager (`app/core/reports`)

The Report Manager handles test result collection and presentation:

- Collects metrics from all protocol adapters
- Generates summary statistics
- Formats output for different report types
- Handles report export functionality

## Protocol Adapters

### 1. Redis Adapter (`app/adapters/redis`)

The Redis adapter provides comprehensive Redis testing capabilities:

#### Redis Connection Management

- Supports standalone, cluster, and sentinel modes
- Connection pooling for efficient resource usage
- Authentication and TLS support

#### Redis Operations

- Multiple test cases (SET/GET, INCR, LPUSH/LPOP, etc.)
- Configurable read/write ratios
- TTL support
- Key generation strategies

#### Redis Configuration

- Host, port, and authentication settings
- Mode selection (standalone/cluster/sentinel)
- Connection timeout and retry settings

### 2. HTTP Adapter (`app/adapters/http`)

The HTTP adapter enables HTTP endpoint performance testing:

#### HTTP Connection Management

- Connection pooling with keep-alive support
- Configurable connection limits
- Timeout and retry handling

#### HTTP Operations

- Support for GET, POST, PUT, DELETE methods
- Custom headers and request bodies
- Content-Type handling
- Response validation

#### HTTP Configuration

- Base URL and path settings
- Method and header configuration
- Request body templates
- Timeout and retry settings

### 3. Kafka Adapter (`app/adapters/kafka`)

The Kafka adapter provides Kafka producer and consumer testing:

#### Kafka Connection Management

- Broker connection management
- Producer and consumer pooling
- Topic management

#### Kafka Operations

- Producer performance testing
- Consumer lag testing
- Mixed produce/consume workloads
- Message size and compression configuration

#### Kafka Configuration

- Broker and topic settings
- Producer and consumer configurations
- Message format and serialization
- Security settings (TLS, SASL)

## Utility Components

### 1. Logging (`app/utils`)

The logging system provides structured logging capabilities:

- Configurable log levels
- Structured log output
- Performance-optimized logging

### 2. Command Parsing (`app/core/config`)

Command parsing handles CLI argument processing:

- Flag definition and validation
- Help text generation
- Argument type conversion

### 3. Metrics Collection (`app/core/monitoring`)

Metrics collection gathers performance data:

- Real-time metric collection
- Statistical analysis
- Performance profiling
