# TCP协议测试操作手册

## 概述

abc-runner TCP协议测试模块提供了全面的TCP性能测试功能，支持多种测试场景和灵活的配置选项。本手册提供了详细的命令示例和最佳实践。

## 快速开始

### 基本语法

```bash
abc-runner tcp [选项]
```

### 最简单的测试

```bash
# 使用默认配置连接到localhost:9090进行echo测试
abc-runner tcp

# 显示帮助信息
abc-runner tcp --help
```

## 连接配置

### 基本连接选项

```bash
# 指定服务器地址和端口
abc-runner tcp --host 192.168.1.100 --port 8080

# 使用域名连接
abc-runner tcp --host example.com --port 9090

# 连接到本地不同端口
abc-runner tcp -h localhost -p 8888
```

### 高级连接配置

```bash
# 启用TCP Keep-Alive
abc-runner tcp --host server.example.com --port 9090 --keep-alive

# 禁用Nagle算法（降低小包延迟）
abc-runner tcp --host 10.0.0.1 --port 9090 --no-delay

# 组合多个连接选项
abc-runner tcp -h 192.168.1.50 -p 9090 --keep-alive --no-delay
```

## 测试类型与场景

### 1. Echo测试（回显测试）

最常用的TCP测试类型，发送数据并验证服务器返回相同数据。

```bash
# 基本echo测试
abc-runner tcp --test-case echo_test

# 指定数据包大小的echo测试
abc-runner tcp --test-case echo_test --data-size 2048

# 高并发echo测试
abc-runner tcp --test-case echo_test -c 50 -n 10000

# 长时间echo测试
abc-runner tcp --test-case echo_test --duration 300s -c 20
```

### 2. 仅发送测试（Send Only）

只发送数据，不等待响应，适合测试服务器接收能力。

```bash
# 基本发送测试
abc-runner tcp --test-case send_only

# 大数据包发送测试
abc-runner tcp --test-case send_only --data-size 8192 -n 5000

# 高速发送测试
abc-runner tcp --test-case send_only -c 100 -n 50000 --data-size 1024

# 持续发送测试
abc-runner tcp --test-case send_only --duration 600s -c 30
```

### 3. 仅接收测试（Receive Only）

只接收服务器推送的数据，适合测试服务器推送能力。

```bash
# 基本接收测试
abc-runner tcp --test-case receive_only

# 长时间接收测试
abc-runner tcp --test-case receive_only --duration 180s -c 10

# 多连接接收测试
abc-runner tcp --test-case receive_only -c 25 -n 1000
```

### 4. 双向传输测试（Bidirectional）

同时进行发送和接收，测试全双工通信能力。

```bash
# 基本双向测试
abc-runner tcp --test-case bidirectional

# 大数据双向测试
abc-runner tcp --test-case bidirectional --data-size 4096 -c 20

# 高并发双向测试
abc-runner tcp --test-case bidirectional -c 50 -n 20000

# 长时间双向测试
abc-runner tcp --test-case bidirectional --duration 900s -c 15
```

## 性能调优参数

### 并发连接数调优

```bash
# 低并发（适合延迟敏感测试）
abc-runner tcp -c 5 -n 1000 --test-case echo_test

# 中等并发（平衡性能测试）
abc-runner tcp -c 25 -n 10000 --test-case echo_test

# 高并发（压力测试）
abc-runner tcp -c 100 -n 100000 --test-case echo_test

# 极高并发（极限测试）
abc-runner tcp -c 500 -n 500000 --test-case send_only
```

### 数据包大小调优

```bash
# 小包测试（适合低延迟场景）
abc-runner tcp --data-size 64 -c 50 -n 50000

# 标准包测试
abc-runner tcp --data-size 1024 -c 25 -n 10000

# 大包测试（适合吞吐量场景）
abc-runner tcp --data-size 8192 -c 10 -n 5000

# 巨包测试（网络极限测试）
abc-runner tcp --data-size 65536 -c 5 -n 1000
```

### 测试持续时间控制

```bash
# 短时间快速测试（30秒）
abc-runner tcp --duration 30s -c 20 --test-case echo_test

# 中等时间稳定性测试（5分钟）
abc-runner tcp --duration 300s -c 15 --test-case bidirectional

# 长时间耐久性测试（30分钟）
abc-runner tcp --duration 1800s -c 10 --test-case send_only

# 超长时间可靠性测试（2小时）
abc-runner tcp --duration 7200s -c 5 --test-case echo_test
```

## 实际应用场景

### 1. 服务器性能基准测试

```bash
# 测试服务器基本响应能力
abc-runner tcp -h prod-server.example.com -p 9090 \
  --test-case echo_test -c 10 -n 10000 --data-size 1024

# 测试服务器极限吞吐量
abc-runner tcp -h prod-server.example.com -p 9090 \
  --test-case send_only -c 100 -n 100000 --data-size 8192

# 测试服务器稳定性
abc-runner tcp -h prod-server.example.com -p 9090 \
  --test-case bidirectional --duration 3600s -c 20 --data-size 2048
```

### 2. 网络延迟和质量测试

```bash
# 低延迟网络测试
abc-runner tcp -h local-server -p 9090 \
  --test-case echo_test --data-size 64 -c 1 -n 10000

# 跨地域网络测试
abc-runner tcp -h remote-server.example.com -p 9090 \
  --test-case echo_test --data-size 1024 -c 5 -n 5000

# 网络抖动测试
abc-runner tcp -h test-server -p 9090 \
  --test-case echo_test --duration 600s -c 10 --data-size 512
```

### 3. 负载均衡器测试

```bash
# 测试负载均衡器分发能力
abc-runner tcp -h lb.example.com -p 9090 \
  --test-case echo_test -c 50 -n 50000 --data-size 1024

# 测试负载均衡器连接保持
abc-runner tcp -h lb.example.com -p 9090 \
  --test-case bidirectional --duration 1800s -c 25 --keep-alive
```

### 4. 容器和K8s环境测试

```bash
# 测试容器内服务
abc-runner tcp -h tcp-service.default.svc.cluster.local -p 9090 \
  --test-case echo_test -c 20 -n 20000

# 测试NodePort服务
abc-runner tcp -h k8s-node.example.com -p 30090 \
  --test-case send_only -c 30 -n 30000 --data-size 2048
```

## 性能调优建议

### 根据测试目标选择参数

```bash
# 延迟优化测试
abc-runner tcp --test-case echo_test --data-size 64 -c 1 -n 10000 --no-delay

# 吞吐量优化测试  
abc-runner tcp --test-case send_only --data-size 8192 -c 50 -n 100000

# 连接数优化测试
abc-runner tcp --test-case echo_test -c 200 -n 200000 --data-size 1024

# 稳定性测试
abc-runner tcp --test-case bidirectional --duration 3600s -c 15 --keep-alive
```

### 渐进式压力测试

```bash
# 第一阶段：基线测试
abc-runner tcp --test-case echo_test -c 5 -n 5000 --data-size 1024

# 第二阶段：中等压力
abc-runner tcp --test-case echo_test -c 25 -n 25000 --data-size 1024

# 第三阶段：高压力
abc-runner tcp --test-case echo_test -c 50 -n 50000 --data-size 1024

# 第四阶段：极限压力
abc-runner tcp --test-case echo_test -c 100 -n 100000 --data-size 1024
```

## 故障排查

### 连接问题诊断

```bash
# 基本连通性测试
abc-runner tcp -h target-server -p 9090 -c 1 -n 1 --test-case echo_test

# 超时问题诊断
abc-runner tcp -h slow-server -p 9090 -c 1 -n 100 --test-case echo_test

# 端口可用性测试
abc-runner tcp -h server -p 8080 --test-case send_only -c 1 -n 1
abc-runner tcp -h server -p 8081 --test-case send_only -c 1 -n 1
abc-runner tcp -h server -p 8082 --test-case send_only -c 1 -n 1
```

### 性能瓶颈定位

```bash
# CPU瓶颈测试（小包高频）
abc-runner tcp --test-case echo_test --data-size 64 -c 100 -n 1000000

# 带宽瓶颈测试（大包）
abc-runner tcp --test-case send_only --data-size 65536 -c 10 -n 10000

# 连接数瓶颈测试
abc-runner tcp --test-case echo_test -c 500 -n 100000 --data-size 1024

# 内存瓶颈测试
abc-runner tcp --test-case bidirectional --data-size 32768 -c 50 -n 50000
```

## 监控和分析

### 关键指标观察

运行测试时，重点关注以下输出指标：

```bash
# 运行示例命令并观察输出
abc-runner tcp --test-case echo_test -c 20 -n 10000 --data-size 2048

# 输出示例:
# ✅ Concurrent TCP test completed
#    Test Case: echo_test
#    Total Jobs: 10000
#    Completed: 10000
#    Success: 9998
#    Failed: 2
#    Duration: 15.2s
#    Success Rate: 99.98%
#    Throughput: 657.89 ops/sec
```

### 性能基准参考

```bash
# 局域网环境基准（期望值）
# - 延迟: < 1ms
# - 吞吐量: > 10000 ops/sec
# - 成功率: > 99.9%
abc-runner tcp -h local-server -p 9090 --test-case echo_test -c 10 -n 10000

# 广域网环境基准（期望值）
# - 延迟: < 100ms  
# - 吞吐量: > 1000 ops/sec
# - 成功率: > 99.5%
abc-runner tcp -h remote-server.example.com -p 9090 --test-case echo_test -c 5 -n 5000
```

## 配置文件使用

创建配置文件 `tcp-test.yaml`：

```yaml
tcp:
  connection:
    address: "test-server.example.com"
    port: 9090
    timeout: "30s"
    keep_alive: true
    keep_alive_period: "30s"
    pool:
      pool_size: 20
      min_idle: 5
      max_idle: 15
      idle_timeout: "300s"
      connection_timeout: "30s"

  benchmark:
    total: 50000
    parallels: 25
    data_size: 2048
    duration: "300s"
    read_percent: 80
    random_keys: 1000
    test_case: "echo_test"

  tcp_specific:
    connection_mode: "persistent"
    no_delay: true
    buffer_size: 8192
    linger_timeout: -1
    reuse_address: true
```

使用配置文件：

```bash
# 使用配置文件运行测试
abc-runner tcp --config tcp-test.yaml

# 配置文件 + 命令行参数覆盖
abc-runner tcp --config tcp-test.yaml --test-case bidirectional -c 50
```

## 最佳实践

### 1. 测试前准备

```bash
# 确认目标服务器状态
abc-runner tcp -h target-server -p 9090 -c 1 -n 1 --test-case echo_test

# 预热测试（避免冷启动影响）
abc-runner tcp -h target-server -p 9090 -c 5 -n 1000 --test-case echo_test
```

### 2. 分阶段测试策略

```bash
# 阶段1：功能验证
abc-runner tcp --test-case echo_test -c 1 -n 100

# 阶段2：性能基线
abc-runner tcp --test-case echo_test -c 10 -n 5000

# 阶段3：压力测试
abc-runner tcp --test-case echo_test -c 50 -n 50000

# 阶段4：极限测试
abc-runner tcp --test-case echo_test -c 100 -n 100000
```

### 3. 不同场景的推荐配置

```bash
# 微服务健康检查模拟
abc-runner tcp --test-case echo_test -c 5 -n 1000 --data-size 256

# API网关压力测试
abc-runner tcp --test-case bidirectional -c 100 -n 100000 --data-size 2048

# 数据库连接池测试
abc-runner tcp --test-case echo_test -c 200 --duration 1800s --keep-alive

# 消息队列性能测试
abc-runner tcp --test-case send_only -c 50 -n 1000000 --data-size 4096
```

## 注意事项

1. **服务器准备**: 确保目标服务器在指定端口运行TCP回显服务
2. **防火墙设置**: 确认网络防火墙允许TCP连接
3. **系统资源**: 高并发测试时注意客户端系统的文件描述符限制
4. **网络环境**: 测试结果会受网络质量影响，建议多次测试取平均值
5. **服务器负载**: 避免在生产高峰期进行大规模压力测试

## 常见错误及解决方案

### 连接被拒绝

```bash
# 错误现象：connection refused
# 解决方案：检查服务器状态和端口
netstat -tlnp | grep 9090
abc-runner tcp -h localhost -p 9090 -c 1 -n 1
```

### 连接超时

```bash
# 错误现象：connection timeout
# 解决方案：检查网络连通性和防火墙
ping target-server
telnet target-server 9090
```

### 性能不达预期

```bash
# 解决方案：逐步调优参数
abc-runner tcp --test-case echo_test -c 1 -n 1000    # 基线测试
abc-runner tcp --test-case echo_test -c 5 -n 1000    # 增加并发
abc-runner tcp --test-case echo_test -c 10 -n 1000   # 继续增加
```

通过本手册的详细示例，您可以根据不同的测试需求选择合适的命令参数，有效地进行TCP协议性能测试和故障诊断。
