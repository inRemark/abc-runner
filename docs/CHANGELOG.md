# Changelog

All notable changes to this project will be documented in this file.

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

### 🛠️ Improvements

- Refactored Redis operations to support more data types
- Improved operation validation and error handling
- Enhanced documentation with examples for new operations
- Better code organization and modularity

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