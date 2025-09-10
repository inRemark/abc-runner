# Kubernetes Deployment Guide

[English](kubernetes.md) | [中文](../zh/deployment/kubernetes.md)

## Helm Chart

abc-runner provides Helm Charts for deployment in Kubernetes clusters.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.0+

## Adding Helm Repository

```bash
helm repo add abc-runner https://abc-runner.github.io/helm-charts
helm repo update
```

## Installing Charts

### Basic Installation

```bash
helm install my-abc-runner abc-runner/abc-runner
```

### Custom Installation

```bash
helm install my-abc-runner abc-runner/abc-runner \
  --set redis.host=redis-master \
  --set redis.port=6379 \
  --set replicaCount=3
```

## Configuration Values

### Redis Configuration

```yaml
redis:
  host: "redis-master"
  port: 6379
  password: ""
  database: 0
```

### HTTP Configuration

```yaml
http:
  url: "http://example.com"
  method: "GET"
```

### Kafka Configuration

```yaml
kafka:
  brokers: "kafka:9092"
  topic: "test-topic"
```

### Resource Configuration

```yaml
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Persistence Configuration

```yaml
persistence:
  enabled: true
  size: 10Gi
  storageClass: "-"
```

## Using ConfigMap

### Creating Configuration ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: abc-runner-config
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

### Using in Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: abc-runner
spec:
  replicas: 1
  selector:
    matchLabels:
      app: abc-runner
  template:
    metadata:
      labels:
        app: abc-runner
    spec:
      containers:
      - name: abc-runner
        image: abc-runner/abc-runner:latest
        args: ["redis", "--config", "/config/redis.yaml"]
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: abc-runner-config
```

## Batch Jobs

### One-time Test Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: redis-benchmark
spec:
  template:
    spec:
      containers:
      - name: abc-runner
        image: abc-runner/abc-runner:latest
        args: ["redis", "-h", "redis-master", "-p", "6379", "-n", "10000", "-c", "50"]
      restartPolicy: Never
  backoffLimit: 4
```

### Scheduled Test CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-benchmark-cron
spec:
  schedule: "0 */6 * * *"  # Run every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: abc-runner
            image: abc-runner/abc-runner:latest
            args: ["redis", "-h", "redis-master", "-p", "6379", "-n", "10000", "-c", "50"]
          restartPolicy: OnFailure
```

## Service Monitoring

### Prometheus Metrics

```yaml
apiVersion: v1
kind: Service
metadata:
  name: abc-runner-metrics
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
spec:
  selector:
    app: abc-runner
  ports:
  - name: metrics
    port: 8080
    targetPort: 8080
```

### Grafana Dashboard

Create Grafana dashboard JSON files to visualize abc-runner metrics.

## Auto Scaling

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: abc-runner-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: abc-runner
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

## Network Policies

### Restricting Network Access

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: abc-runner-policy
spec:
  podSelector:
    matchLabels:
      app: abc-runner
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

## Security Configuration

### Pod Security Policy

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: abc-runner-psp
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

### RBAC Configuration

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: abc-runner
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: abc-runner-role
rules:
- apiGroups: [""]
  resources: ["pods", "configmaps"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: abc-runner-rolebinding
subjects:
- kind: ServiceAccount
  name: abc-runner
roleRef:
  kind: Role
  name: abc-runner-role
  apiGroup: rbac.authorization.k8s.io
```

## Troubleshooting

### Viewing Pod Logs

```bash
kubectl logs -f deployment/abc-runner
```

### Entering Pod for Debugging

```bash
kubectl exec -it deployment/abc-runner -- sh
```

### Checking Events

```bash
kubectl get events --sort-by=.metadata.creationTimestamp
```

## Best Practices

1. **Resource Limits**: Set appropriate resource requests and limits for Pods
2. **Health Checks**: Configure liveness and readiness probes
3. **Configuration Management**: Use ConfigMaps and Secrets for configuration management
4. **Security**: Use non-root users and read-only file systems
5. **Monitoring**: Integrate with Prometheus and Grafana for monitoring
6. **Logging**: Configure appropriate log collection and analysis
7. **Backup**: Regularly back up important configurations and data
8. **Update Strategy**: Use rolling update strategies to minimize downtime