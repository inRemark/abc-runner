# HTTP测试指南

[English](http.md) | [中文](http.zh.md)

本指南涵盖了abc-runner的HTTP特定功能和使用模式。

## 基本HTTP测试

### 简单GET请求

```bash
./abc-runner http --url http://localhost:8080 -n 10000 -c 50
```

### 带主体的POST请求

```bash
./abc-runner http --url http://api.example.com/users \
  --method POST --body '{"name":"test"}' \
  --content-type application/json -n 1000 -c 20
```

### 自定义头部

```bash
./abc-runner http --url http://api.example.com \
  --header "Authorization:Bearer token123" \
  --header "X-API-Key:secret" -n 1000
```

## HTTP方法

支持所有标准HTTP方法：

### GET

```bash
./abc-runner http --url http://localhost:8080/api/users -n 1000 -c 10
```

### POST

```bash
./abc-runner http --url http://localhost:8080/api/users \
  --method POST --body '{"name":"John"}' -n 1000 -c 10
```

### PUT

```bash
./abc-runner http --url http://localhost:8080/api/users/123 \
  --method PUT --body '{"name":"Jane"}' -n 1000 -c 10
```

### DELETE

```bash
./abc-runner http --url http://localhost:8080/api/users/123 \
  --method DELETE -n 1000 -c 10
```

## 基于时长的测试

除了固定请求数量外，您还可以运行特定时长的测试：

```bash
./abc-runner http --url http://localhost:8080 --duration 60s -c 100
```

## 配置文件示例

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

使用配置文件运行：

```bash
./abc-runner http --config http.yaml
```

## 高级功能

### 连接池

控制连接池行为：

```bash
./abc-runner http --url http://localhost:8080 -n 10000 -c 50 \
  --max-conns-per-host 100 --keep-alive 60s
```

### 请求定制

使用查询参数、头部和主体内容自定义请求：

```bash
./abc-runner http --url http://localhost:8080/api/search \
  --query "q=test&limit=10" \
  --header "Accept:application/json" \
  --header "User-Agent:abc-runner/1.0" \
  -n 1000 -c 20
```

### 响应验证

验证HTTP响应状态码：

```bash
./abc-runner http --url http://localhost:8080/health \
  --expected-status 200 -n 1000 -c 10
```

## 性能调优

### 并发控制

根据服务器容量调整并行连接数：

```bash
./abc-runner http --url http://localhost:8080 -n 100000 -c 200
```

### 连接复用

通过keep-alive设置优化连接复用：

```bash
./abc-runner http --url http://localhost:8080 -n 10000 -c 50 \
  --keep-alive 300s
```

### 超时配置

为端点设置适当的超时：

```bash
./abc-runner http --url http://localhost:8080 -n 1000 -c 10 \
  --timeout 10s
```