# Redis测试指南

[English](redis.md) | [中文](redis.zh.md)

本指南涵盖了abc-runner的Redis特定功能和使用模式。

## Redis连接模式

abc-runner支持三种Redis部署模式：

### 1. 单机模式（默认）

用于单个Redis实例：

```bash
./abc-runner redis -h localhost -p 6379 -n 10000 -c 50
```

### 2. 集群模式

用于Redis集群部署：

```bash
./abc-runner redis --mode cluster -h localhost -p 6379 -n 10000 -c 50
```

### 3. 哨兵模式

用于Redis哨兵管理的实例：

```bash
./abc-runner redis --mode sentinel -h localhost -p 26379 -n 10000 -c 50
```

## 认证

测试带认证的Redis实例：

```bash
./abc-runner redis -h localhost -p 6379 -a password -n 10000 -c 50
```

## 测试用例

abc-runner支持多种Redis测试用例：

### set_get_random

混合SET和GET操作与随机键：

```bash
./abc-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80
```

### set_only

仅SET操作：

```bash
./abc-runner redis -t set_only -n 100000 -c 100
```

### get_only

仅GET操作（需要预先存在的键）：

```bash
./abc-runner redis -t get_only -n 100000 -c 100
```

### incr

对计数器键的INCR操作：

```bash
./abc-runner redis -t incr -n 50000 -c 100
```

### append

对字符串键的APPEND操作：

```bash
./abc-runner redis -t append -n 50000 -c 100
```

### lpush_lpop

对列表键的LPUSH和LPOP操作：

```bash
./abc-runner redis -t lpush_lpop -n 10000 -c 50
```

## 键生成策略

### 全局自增键

当`-r 0`（默认）时，键是全局自增的：

```bash
./abc-runner redis -n 100000 -c 100 -r 0
```

### 随机键

当`-r > 0`时，键是随机生成的：

```bash
./abc-runner redis -n 100000 -c 100 -r 1000
```

## TTL配置

为键设置过期时间：

```bash
./abc-runner redis -n 100000 -c 100 --ttl 300s
```

## 配置文件示例

```yaml
# redis.yaml
protocol: redis
connection:
  host: localhost
  port: 6379
  mode: standalone  # standalone, cluster, sentinel
  password: ""      # 可选密码
  timeout: 30s

benchmark:
  total: 100000
  parallels: 50
  test_case: "set_get_random"
  data_size: 64
  read_ratio: 0.5
  key_range: 0        # 0为自增，>0为随机
  ttl: 0s             # 0为不过期，>0s为过期
```

使用配置文件运行：

```bash
./abc-runner redis --config redis.yaml
```

## 性能调优

### 连接池

根据Redis服务器容量调整并行连接数：

```bash
./abc-runner redis -n 100000 -c 100  # 100个并行连接
```

### 数据大小

控制SET操作中使用的数据大小：

```bash
./abc-runner redis -n 100000 -c 50 -d 1024  # 1KB值
```

### 读取比例

对于混合工作负载，控制读写操作的比例：

```bash
./abc-runner redis -t set_get_random -n 100000 -c 100 --read-ratio 80  # 80%读取
```