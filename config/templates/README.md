# 配置模板说明

## Redis 配置模板 (redis.yaml)

包含Redis所有支持的配置选项，适用于单机、哨兵和集群模式。

### 主要配置项

- **mode**: Redis运行模式 (standalone/sentinel/cluster)
- **benchmark**: 基准测试参数
  - total: 总请求数
  - parallels: 并发连接数
  - random_keys: 随机键范围
  - read_percent: 读操作百分比
  - data_size: 数据大小(字节)
  - ttl: 键过期时间(秒)
  - case: 测试用例类型
- **pool**: 连接池配置
- **standalone/sentinel/cluster**: 不同模式的连接参数

## HTTP 配置模板 (http.yaml)

包含HTTP测试的所有配置选项，支持复杂的请求模板和认证配置。

### 主要配置项

- **connection**: 连接参数
- **auth**: 认证配置
- **requests**: 请求模板列表
- **benchmark**: 基准测试参数
- **upload**: 文件上传配置

## Kafka 配置模板 (kafka.yaml)

包含Kafka生产者和消费者的完整配置选项。

### 主要配置项

- **brokers**: Kafka broker地址列表
- **producer**: 生产者配置
- **consumer**: 消费者配置
- **security**: 安全配置(TLS/SASL)
- **benchmark**: 基准测试参数

## 使用方法

复制模板文件到目标目录并根据实际需求修改配置参数。