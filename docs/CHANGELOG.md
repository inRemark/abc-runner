# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2025-10-01

### 🚀 New Features

- **Added comprehensive multi-protocol support**:
  - gRPC protocol adapter with connection pooling and streaming support
  - WebSocket protocol adapter for real-time communication testing
  - TCP protocol adapter for raw TCP connection performance testing
  - UDP protocol adapter for datagram transmission testing
- **Enhanced server mode capabilities**:
  - Multi-protocol server implementations for testing against live services
  - Unified server management framework with health checks
  - Configurable server endpoints and service discovery
- **Advanced architecture improvements**:
  - Dependency injection container for better modularity
  - Unified configuration management system across all protocols
  - Plugin-based adapter architecture for extensibility
  - Auto-discovery mechanism for protocol-specific configurations

### 🛠️ Improvements

- **Enhanced configuration system**:
  - Unified core configuration with protocol-specific extensions
  - Improved configuration validation and error reporting
  - Support for configuration file inheritance and overrides
- **Better resource management**:
  - Optimized connection pooling across all protocols
  - Improved memory management and garbage collection
  - Enhanced connection lifecycle management
- **Expanded metrics and monitoring**:
  - Advanced health checker with protocol-specific metrics
  - Enhanced metrics collection and storage
  - Improved reporting capabilities with structured output

### 🐛 Bug Fixes

- Fixed configuration loading issues with multiple protocol files
- Resolved connection pool leaks in long-running tests
- Fixed metrics aggregation for concurrent protocol testing
- Corrected error handling in adapter factory initialization

### ⚠️ Breaking Changes

- Configuration file structure updated to support multi-protocol architecture
- Command-line interface enhanced with new protocol-specific options
- Metrics output format standardized across all protocols

### 📚 Documentation

- Updated README with comprehensive protocol support information
- Added protocol-specific configuration examples
- Enhanced command reference with new protocol options
- Improved architecture documentation reflecting multi-protocol design

### 🔧 Internal Changes

- Refactored adapter architecture for better maintainability
- Implemented unified interfaces for all protocol adapters
- Enhanced error handling and logging throughout the codebase
- Improved test coverage for new protocol implementations

## [0.2.0] - 2025-09-08

### 🚀 New Features

- Added support for additional Redis operations:
  - Counter operations: INCR, DECR
  - List operations: LPUSH, RPUSH, LPOP, RPOP
  - Set operations: SADD, SREM, SMEMBERS, SISMEMBER
  - Sorted set operations: ZADD, ZREM, ZRANGE, ZRANK
  - Extended hash operations: HMSET, HMGET, HGETALL
  - Subscription operations: SUBSCRIBE, UNSUBSCRIBE
- Enhanced Redis operation factory with better extensibility
- Added comprehensive unit tests for all new operations
- **Added packaging and distribution management system**:
  - Automated cross-platform build process
  - Release package generation with platform-specific archives
  - Integrated configuration file distribution
  - Semantic versioning support

### 🛠️ Improvements

- Refactored Redis operations to support more data types
- Improved operation validation and error handling
- Enhanced documentation with examples for new operations
- Better code organization and modularity
- **Enhanced Makefile with improved release targets**:
  - `make release` now creates complete release packages
  - Platform-specific archive generation (tar.gz for Unix, zip for Windows)
  - Integrated documentation and license distribution

## [0.1.0] - 2025-09-07

### ⚠️ Breaking Changes

- Unified command structure for all protocols (Redis, HTTP, Kafka)
- Simplified configuration format
- Renamed commands:
  - `redis-enhanced` → `redis`
  - `http-enhanced` → `http`
  - `kafka-enhanced` → `kafka`

### 🚀 New Features

- Unified performance testing tool for Redis, HTTP, and Kafka protocols
- Enhanced Redis testing with support for cluster, sentinel, and standalone modes
- Comprehensive HTTP testing with custom headers and request bodies
- Kafka producer and consumer performance testing
- Configuration file support for all protocols
- Improved metrics collection and reporting

### 🛠️ Improvements

- Better connection pooling and resource management
- Enhanced error handling and logging
- Improved documentation and examples
- More comprehensive test coverage

### 📚 Documentation

- Updated README with comprehensive usage examples
- Added configuration file templates
- Improved command reference documentation