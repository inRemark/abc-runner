# Kubernetes部署指南

[English](../en/deployment/kubernetes.md) | [中文](kubernetes.md)

## Helm Chart

redis-runner提供Helm Chart用于在Kubernetes集群中部署。

## 先决条件

- Kubernetes 1.16+
- Helm 3.0+

## 添加Helm仓库

```bash
helm repo add redis-runner https://redis-runner.github.io/helm-charts
helm repo update
```

## 安装Chart

### 基本安装

```bash
helm install my-redis-runner redis-runner/redis-runner
```

### 自定义安装

```bash
helm install my-redis-runner redis-runner/redis-runner \
  --set redis.host=redis-master \
  --set redis.port=6379 \
  --set replicaCount=3
```

## 配置值

### Redis配置

```yaml
redis:
  host: "redis-master"
  port: 6379
  password: ""
  database: 0
```

### HTTP配置

```yaml
http:
  url: "http://example.com"
  method: "GET"
```

### Kafka配置

```yaml
kafka:
  brokers: "kafka:9092"
  topic: "test-topic"
```

### 资源配置

```yaml
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### 持久化配置

```yaml
persistence:
  enabled: true
  size: 10Gi
  storageClass: "-"
```

## 使用ConfigMap

### 创建配置ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-runner-config
data:
  redis.yaml: |
    redis:
      mode: "standalone"
      benchmark:
        total: 10000
        parallels: 50
      standalone:
        addr: "redis-master:6379"
```

### 在Deployment中使用

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-runner
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis-runner
  template:
    metadata:
      labels:
        app: redis-runner
    spec:
      containers:
      - name: redis-runner
        image: redis-runner/redis-runner:latest
        args: ["redis", "--config", "/config/redis.yaml"]
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: redis-runner-config
```

## 批处理Job

### 一次性测试Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: redis-benchmark
spec:
  template:
    spec:
      containers:
      - name: redis-runner
        image: redis-runner/redis-runner:latest
        args: ["redis", "-h", "redis-master", "-p", "6379", "-n", "10000", "-c", "50"]
      restartPolicy: Never
  backoffLimit: 4
```

### 定时测试CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-benchmark-cron
spec:
  schedule: "0 */6 * * *"  # 每6小时运行一次
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: redis-runner
            image: redis-runner/redis-runner:latest
            args: ["redis", "-h", "redis-master", "-p", "6379", "-n", "10000", "-c", "50"]
          restartPolicy: OnFailure
```

## 服务监控

### Prometheus指标

```yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-runner-metrics
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
spec:
  selector:
    app: redis-runner
  ports:
  - name: metrics
    port: 8080
    targetPort: 8080
```

### Grafana仪表板

创建Grafana仪表板JSON文件来可视化redis-runner指标。

## 自动扩缩容

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: redis-runner-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: redis-runner
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
```

## 网络策略

### 限制网络访问

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: redis-runner-policy
spec:
  podSelector:
    matchLabels:
      app: redis-runner
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: monitoring
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: redis
```

## 安全配置

### Pod安全策略

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: redis-runner-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
  - ALL
  volumes:
  - configMap
  - emptyDir
  - projected
  - secret
  - downwardAPI
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: MustRunAs
    ranges:
    - min: 1
      max: 65535
  fsGroup:
    rule: MustRunAs
    ranges:
    - min: 1
      max: 65535
  readOnlyRootFilesystem: true
```

### RBAC配置

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: redis-runner
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: redis-runner-role
rules:
- apiGroups: [""]
  resources: ["pods", "configmaps"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: redis-runner-rolebinding
subjects:
- kind: ServiceAccount
  name: redis-runner
roleRef:
  kind: Role
  name: redis-runner-role
  apiGroup: rbac.authorization.k8s.io
```

## 故障排除

### 查看Pod日志

```bash
kubectl logs -f deployment/redis-runner
```

### 进入Pod调试

```bash
kubectl exec -it deployment/redis-runner -- sh
```

### 检查事件

```bash
kubectl get events --sort-by=.metadata.creationTimestamp
```

## 最佳实践

1. **资源限制**: 为Pod设置适当的资源请求和限制
2. **健康检查**: 配置liveness和readiness探针
3. **配置管理**: 使用ConfigMap和Secret管理配置
4. **安全**: 使用非root用户和只读文件系统
5. **监控**: 集成Prometheus和Grafana进行监控
6. **日志**: 配置适当的日志收集和分析
7. **备份**: 定期备份重要配置和数据
8. **更新策略**: 使用滚动更新策略减少停机时间