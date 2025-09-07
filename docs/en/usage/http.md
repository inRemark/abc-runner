# HTTP Testing Guide

[English](http.md) | [中文](http.zh.md)

This guide covers HTTP-specific features and usage patterns for redis-runner.

## Basic HTTP Testing

### Simple GET Requests

```bash
./redis-runner http --url http://localhost:8080 -n 10000 -c 50
```

### POST Requests with Body

```bash
./redis-runner http --url http://api.example.com/users \
  --method POST --body '{"name":"test"}' \
  --content-type application/json -n 1000 -c 20
```

### Custom Headers

```bash
./redis-runner http --url http://api.example.com \
  --header "Authorization:Bearer token123" \
  --header "X-API-Key:secret" -n 1000
```

## HTTP Methods

Support for all standard HTTP methods:

### GET

```bash
./redis-runner http --url http://localhost:8080/api/users -n 1000 -c 10
```

### POST

```bash
./redis-runner http --url http://localhost:8080/api/users \
  --method POST --body '{"name":"John"}' -n 1000 -c 10
```

### PUT

```bash
./redis-runner http --url http://localhost:8080/api/users/123 \
  --method PUT --body '{"name":"Jane"}' -n 1000 -c 10
```

### DELETE

```bash
./redis-runner http --url http://localhost:8080/api/users/123 \
  --method DELETE -n 1000 -c 10
```

## Duration-Based Testing

Instead of a fixed number of requests, you can run tests for a specific duration:

```bash
./redis-runner http --url http://localhost:8080 --duration 60s -c 100
```

## Configuration File Example

```yaml
# http.yaml
protocol: http
connection:
  base_url: "http://localhost:8080"
  timeout: 30s
  max_conns_per_host: 50
  keep_alive: 30s

benchmark:
  total: 10000
  parallels: 50
  method: "GET"
  path: "/api/test"
  headers:
    "Content-Type": "application/json"
    "Authorization": "Bearer token"
  body: ""
```

Run with configuration:

```bash
./redis-runner http --config http.yaml
```

## Advanced Features

### Connection Pooling

Control connection pooling behavior:

```bash
./redis-runner http --url http://localhost:8080 -n 10000 -c 50 \
  --max-conns-per-host 100 --keep-alive 60s
```

### Request Customization

Customize requests with query parameters, headers, and body content:

```bash
./redis-runner http --url http://localhost:8080/api/search \
  --query "q=test&limit=10" \
  --header "Accept:application/json" \
  --header "User-Agent:redis-runner/1.0" \
  -n 1000 -c 20
```

### Response Validation

Validate HTTP response status codes:

```bash
./redis-runner http --url http://localhost:8080/health \
  --expected-status 200 -n 1000 -c 10
```

## Performance Tuning

### Concurrency Control

Adjust the number of parallel connections based on your server capacity:

```bash
./redis-runner http --url http://localhost:8080 -n 100000 -c 200
```

### Connection Reuse

Optimize connection reuse with keep-alive settings:

```bash
./redis-runner http --url http://localhost:8080 -n 10000 -c 50 \
  --keep-alive 300s
```

### Timeout Configuration

Set appropriate timeouts for your endpoints:

```bash
./redis-runner http --url http://localhost:8080 -n 1000 -c 10 \
  --timeout 10s
```
