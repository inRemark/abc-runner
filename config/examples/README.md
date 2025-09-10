# 配置示例说明

## Redis 配置示例 (redis.yaml)

简单的Redis单机模式配置示例，适用于快速开始测试。

### 适用场景

- 本地开发环境测试
- 简单的性能基准测试
- Redis基本功能验证

## HTTP 配置示例 (http.yaml)

基本的HTTP GET请求测试配置示例。

### 适用场景

- HTTP接口连通性测试
- 简单的负载测试
- API基本性能评估

## Kafka 配置示例 (kafka.yaml)

基本的Kafka生产者测试配置示例。

### 适用场景

- Kafka连通性测试
- 基本的消息生产性能测试
- Kafka环境验证

## 使用方法

直接使用示例配置文件运行测试，或根据实际需求进行修改：

```bash
# 直接使用示例配置
./abc-runner redis --config config/examples/redis.yaml

# 复制并修改后使用
cp config/examples/redis.yaml my-redis-config.yaml
# 编辑 my-redis-config.yaml
./abc-runner redis --config my-redis-config.yaml
```