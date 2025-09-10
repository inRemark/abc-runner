# HTTP测试指南

[English](../en/user-guide/http.md) | [中文](http.md)

## 支持的HTTP方法

abc-runner支持所有标准HTTP方法：

- GET
- POST
- PUT
- PATCH
- DELETE
- HEAD
- OPTIONS

## 配置选项

### 命令行选项

```bash
# 基本选项
--url <url>           目标URL
--method <method>     HTTP方法 (默认: GET)
--body <body>         请求体
--content-type <type> 内容类型
--header <header>     自定义请求头 (可多次使用)

# 基准测试选项
-n <requests>         总请求数 (默认: 1000)
-c <connections>      并发连接数 (默认: 10)
--duration <time>     测试持续时间 (例如: 30s, 5m) - 覆盖-n
```

### 配置文件选项

HTTP测试支持复杂的配置文件，可以定义多个请求模板：

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

## 使用示例

### 基本GET请求测试

```bash
# 简单的GET请求测试
./abc-runner http --url http://httpbin.org/get -n 1000 -c 50

# 持续时间测试
./abc-runner http --url http://localhost:8080 --duration 60s -c 100
```

### POST请求测试

```bash
# 带请求体的POST请求
./abc-runner http --url http://httpbin.org/post \
  --method POST \
  --body '{"name":"test"}' \
  --content-type application/json \
  -n 1000 -c 20
```

### 自定义请求头

```bash
# 带自定义请求头的请求
./abc-runner http --url http://api.example.com \
  --header "Authorization:Bearer token123" \
  --header "X-API-Key:secret" \
  -n 1000
```

### 复杂场景测试

```bash
# 使用配置文件进行复杂测试
./abc-runner http --config config/examples/http-complex.yaml
```

## 请求权重

在配置文件中，您可以为不同的请求模板设置权重，以模拟真实的流量分布：

```yaml
requests:
  - method: "GET"
    path: "/api/users"
    weight: 70  # 70%的请求
    
  - method: "POST"
    path: "/api/users"
    weight: 20  # 20%的请求
    
  - method: "DELETE"
    path: "/api/users/1"
    weight: 10  # 10%的请求
```

## 认证支持

### Basic认证

```yaml
auth:
  type: "basic"
  username: "user"
  password: "password"
```

### Bearer Token认证

```yaml
auth:
  type: "bearer"
  token: "your-token-here"
```

## 文件上传测试

abc-runner支持文件上传测试：

```yaml
upload:
  enable: true
  files:
    - field: "document"
      path: "/path/to/files"
      pattern: "*.pdf"
```

## 结果解读

HTTP测试完成后，abc-runner会输出详细的性能报告：

- **RPS**: 每秒请求数
- **成功率**: 成功请求的百分比
- **状态码分布**: 不同HTTP状态码的分布
- **平均响应时间**: 平均请求处理时间
- **P90/P95/P99响应时间**: 90%/95%/99%请求的响应时间
- **最大响应时间**: 最大请求处理时间
- **吞吐量**: 数据传输速率

## 最佳实践

1. **预热**: 在正式测试前运行短时间的预热测试
2. **真实数据**: 使用接近生产环境的数据大小和结构
3. **网络**: 确保网络延迟不会影响测试结果
4. **并发连接**: 根据服务器性能调整并发连接数
5. **请求分布**: 使用权重配置模拟真实的请求分布
6. **监控**: 监控服务器资源使用情况