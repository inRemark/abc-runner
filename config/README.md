# 配置管理

## 目录结构

```
config/
├── templates/           # 配置模板文件
├── examples/            # 配置示例文件
├── production/          # 生产环境配置
├── development/         # 开发环境配置
└── README.md            # 配置说明文档
```

## 配置文件说明

### 模板文件 (templates/)

模板文件包含完整的配置选项和详细注释，适用于所有环境。使用时应复制到相应环境目录并根据需要进行修改。

### 示例文件 (examples/)

示例文件展示了常见使用场景的简化配置，适用于快速开始和测试。

### 环境配置

- **development/**: 开发环境配置
- **production/**: 生产环境配置

## 配置版本管理

所有配置文件都包含版本信息，便于跟踪变更历史。

## 使用说明

1. 选择合适的模板或示例文件
2. 复制到目标环境目录
3. 根据实际需求修改配置参数
4. 使用 `--config` 参数指定配置文件路径运行测试

```bash
# 使用配置文件运行Redis测试
./abc-runner redis --config config/production/redis.yaml

# 使用配置文件运行HTTP测试
./abc-runner http --config config/examples/http.yaml

# 使用配置文件运行Kafka测试
./abc-runner kafka --config config/templates/kafka.yaml
```