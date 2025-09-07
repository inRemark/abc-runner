# HTTP Testing Guide

[English](http.md) | [中文](../zh/user-guide/http.md)

## Supported HTTP Methods

redis-runner supports all standard HTTP methods:

- GET
- POST
- PUT
- PATCH
- DELETE
- HEAD
- OPTIONS

## Configuration Options

### Command Line Options

```bash
# Basic options
--url <url>           Target URL
--method <method>     HTTP method (default: GET)
--body <body>         Request body
--content-type <type> Content type
--header <header>     Custom request header (can be used multiple times)

# Benchmark options
-n <requests>         Total requests (default: 1000)
-c <connections>      Concurrent connections (default: 10)
--duration <time>     Test duration (e.g., 30s, 5m) - overrides -n
```

### Configuration File Options

HTTP testing supports complex configuration files that can define multiple request templates:

```yaml
http:
  connection:
    base_url: "http://example.com"
    timeout: 30s
  requests:
    - method: "GET"
      path: "/api/users"
      headers:
        Accept: "application/json"
      weight: 40
      
    - method: "POST"
      path: "/api/users"
      headers:
        Content-Type: "application/json"
      body:
        name: "test"
      weight: 60
```

## Usage Examples

### Basic GET Request Testing

```bash
# Simple GET request test
./redis-runner http --url http://httpbin.org/get -n 1000 -c 50

# Duration-based test
./redis-runner http --url http://localhost:8080 --duration 60s -c 100
```

### POST Request Testing

```bash
# POST request with body
./redis-runner http --url http://httpbin.org/post \
  --method POST \
  --body '{"name":"test"}' \
  --content-type application/json \
  -n 1000 -c 20
```

### Custom Request Headers

```bash
# Request with custom headers
./redis-runner http --url http://api.example.com \
  --header "Authorization:Bearer token123" \
  --header "X-API-Key:secret" \
  -n 1000
```

### Complex Scenario Testing

```bash
# Complex testing using configuration file
./redis-runner http --config config/examples/http-complex.yaml
```

## Request Weights

In configuration files, you can set weights for different request templates to simulate real traffic distribution:

```yaml
requests:
  - method: "GET"
    path: "/api/users"
    weight: 70  # 70% of requests
    
  - method: "POST"
    path: "/api/users"
    weight: 20  # 20% of requests
    
  - method: "DELETE"
    path: "/api/users/1"
    weight: 10  # 10% of requests
```

## Authentication Support

### Basic Authentication

```yaml
auth:
  type: "basic"
  username: "user"
  password: "password"
```

### Bearer Token Authentication

```yaml
auth:
  type: "bearer"
  token: "your-token-here"
```

## File Upload Testing

redis-runner supports file upload testing:

```yaml
upload:
  enable: true
  files:
    - field: "document"
      path: "/path/to/files"
      pattern: "*.pdf"
```

## Result Interpretation

After HTTP testing is completed, redis-runner will output detailed performance reports:

- **RPS**: Requests Per Second
- **Success Rate**: Percentage of successful requests
- **Status Code Distribution**: Distribution of different HTTP status codes
- **Average Response Time**: Average request processing time
- **P90/P95/P99 Response Time**: Response time for 90%/95%/99% of requests
- **Maximum Response Time**: Maximum request processing time
- **Throughput**: Data transfer rate

## Best Practices

1. **Warm-up**: Run short warm-up tests before formal testing
2. **Real Data**: Use data sizes and structures close to production environment
3. **Network**: Ensure network latency does not affect test results
4. **Concurrent Connections**: Adjust concurrent connections based on server performance
5. **Request Distribution**: Use weight configuration to simulate real request distribution
6. **Monitoring**: Monitor server resource usage