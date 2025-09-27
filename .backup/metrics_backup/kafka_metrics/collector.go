package metrics

import (
	"sort"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
)

// MetricsCollector Kafka指标收集器 (参考Redis架构)
type MetricsCollector struct {
	// 基础指标
	totalOperations   int64
	successOperations int64
	failedOperations  int64
	readOperations    int64
	writeOperations   int64

	// 延迟指标
	totalLatency   time.Duration
	minLatency     time.Duration
	maxLatency     time.Duration
	latencyHistory []time.Duration

	// Kafka特定指标
	producerMetrics  map[string]*ProducerStat
	consumerMetrics  map[string]*ConsumerStat
	partitionMetrics map[string]*PartitionStat
	brokerMetrics    map[string]*BrokerStat
	topicMetrics     map[string]*TopicStat
	connectionStats  *ConnectionStat
	windowStats      *WindowStat

	// 错误统计
	errorStats map[string]int64

	mutex     sync.RWMutex
	startTime time.Time
}

// ProducerStat 生产者统计
type ProducerStat struct {
	ProducedMessages int64           `json:"produced_messages"`
	ProducedBytes    int64           `json:"produced_bytes"`
	BatchCount       int64           `json:"batch_count"`
	AvgBatchSize     float64         `json:"avg_batch_size"`
	ProduceLatencies []time.Duration `json:"produce_latencies"`
	AvgLatency       time.Duration   `json:"avg_latency"`
	FailureCount     int64           `json:"failure_count"`
	RetryCount       int64           `json:"retry_count"`
}

// ConsumerStat 消费者统计
type ConsumerStat struct {
	ConsumedMessages int64           `json:"consumed_messages"`
	ConsumedBytes    int64           `json:"consumed_bytes"`
	ConsumerLag      int64           `json:"consumer_lag"`
	ConsumeLatencies []time.Duration `json:"consume_latencies"`
	AvgLatency       time.Duration   `json:"avg_latency"`
	RebalanceCount   int64           `json:"rebalance_count"`
	CommitCount      int64           `json:"commit_count"`
	CommitErrors     int64           `json:"commit_errors"`
}

// PartitionStat 分区统计
type PartitionStat struct {
	Topic           string    `json:"topic"`
	Partition       int32     `json:"partition"`
	ProducedCount   int64     `json:"produced_count"`
	ConsumedCount   int64     `json:"consumed_count"`
	LastProduceTime time.Time `json:"last_produce_time"`
	LastConsumeTime time.Time `json:"last_consume_time"`
	HighWaterMark   int64     `json:"high_water_mark"`
	ConsumerOffset  int64     `json:"consumer_offset"`
	Lag             int64     `json:"lag"`
	LeaderChanges   int64     `json:"leader_changes"`
}

// BrokerStat 代理统计
type BrokerStat struct {
	BrokerID     int32         `json:"broker_id"`
	Host         string        `json:"host"`
	Port         int32         `json:"port"`
	Connections  int32         `json:"connections"`
	RequestCount int64         `json:"request_count"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount   int64         `json:"error_count"`
	LastSeen     time.Time     `json:"last_seen"`
}

// TopicStat 主题统计
type TopicStat struct {
	TopicName         string `json:"topic_name"`
	PartitionCount    int32  `json:"partition_count"`
	TotalProduced     int64  `json:"total_produced"`
	TotalConsumed     int64  `json:"total_consumed"`
	TotalLag          int64  `json:"total_lag"`
	ReplicationFactor int16  `json:"replication_factor"`
}

// ConnectionStat 连接统计
type ConnectionStat struct {
	TotalConnections  int64         `json:"total_connections"`
	ActiveConnections int64         `json:"active_connections"`
	FailedConnections int64         `json:"failed_connections"`
	ConnectionLatency time.Duration `json:"connection_latency"`
	ReconnectCount    int64         `json:"reconnect_count"`
}

// WindowStat 时间窗口统计
type WindowStat struct {
	WindowSize       time.Duration `json:"window_size"`
	CurrentWindow    int64         `json:"current_window"`
	WindowOperations []int64       `json:"window_operations"`
	RPS              float64       `json:"rps"`
	LastUpdate       time.Time     `json:"last_update"`
}

// BasicMetrics 基础指标
type BasicMetrics struct {
	TotalOperations   int64   `json:"total_operations"`
	SuccessOperations int64   `json:"success_operations"`
	FailedOperations  int64   `json:"failed_operations"`
	ReadOperations    int64   `json:"read_operations"`
	WriteOperations   int64   `json:"write_operations"`
	SuccessRate       float64 `json:"success_rate"`
	ReadWriteRatio    float64 `json:"read_write_ratio"`
	RPS               float64 `json:"rps"`
}

// LatencyMetrics 延迟指标
type LatencyMetrics struct {
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	P50Latency   time.Duration `json:"p50_latency"`
	P90Latency   time.Duration `json:"p90_latency"`
	P95Latency   time.Duration `json:"p95_latency"`
	P99Latency   time.Duration `json:"p99_latency"`
	TotalLatency time.Duration `json:"total_latency"`
}

// NewMetricsCollector 创建Kafka指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		producerMetrics:  make(map[string]*ProducerStat),
		consumerMetrics:  make(map[string]*ConsumerStat),
		partitionMetrics: make(map[string]*PartitionStat),
		brokerMetrics:    make(map[string]*BrokerStat),
		topicMetrics:     make(map[string]*TopicStat),
		connectionStats:  &ConnectionStat{},
		windowStats: &WindowStat{
			WindowSize:       time.Second,
			WindowOperations: make([]int64, 60),
			LastUpdate:       time.Now(),
		},
		errorStats:     make(map[string]int64),
		latencyHistory: make([]time.Duration, 0),
		startTime:      time.Now(),
		minLatency:     time.Duration(^uint64(0) >> 1),
		maxLatency:     0,
	}
}

// RecordOperation 实现核心接口（适配 interfaces.MetricsCollector）
func (kc *MetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	// 更新基础指标
	kc.totalOperations++
	if result.Success {
		kc.successOperations++
	} else {
		kc.failedOperations++
		if result.Error != nil {
			kc.errorStats[result.Error.Error()]++
		}
	}

	if result.IsRead {
		kc.readOperations++
	} else {
		kc.writeOperations++
	}

	// 更新延迟指标
	kc.totalLatency += result.Duration
	if result.Duration < kc.minLatency {
		kc.minLatency = result.Duration
	}
	if result.Duration > kc.maxLatency {
		kc.maxLatency = result.Duration
	}

	// 保存延迟历史（限制大小）
	kc.latencyHistory = append(kc.latencyHistory, result.Duration)
	if len(kc.latencyHistory) > 10000 { // 保留最近10000个样本
		kc.latencyHistory = kc.latencyHistory[1:]
	}

	// 更新时间窗口统计
	kc.updateWindowStats()

	// 处理Kafka特定元数据
	if result.Metadata != nil {
		kc.processKafkaMetadata(result)
	}
}

// GetMetrics 实现核心接口
func (kc *MetricsCollector) GetMetrics() *interfaces.Metrics {
	return kc.GetMetricsForCore()
}

// GetKafkaMetrics 获取Kafka特定指标
func (kc *MetricsCollector) GetKafkaMetrics() map[string]interface{} {
	kc.mutex.RLock()
	defer kc.mutex.RUnlock()

	duration := time.Since(kc.startTime)

	// 基础指标
	basicMetrics := &BasicMetrics{
		TotalOperations:   kc.totalOperations,
		SuccessOperations: kc.successOperations,
		FailedOperations:  kc.failedOperations,
		ReadOperations:    kc.readOperations,
		WriteOperations:   kc.writeOperations,
	}

	if kc.totalOperations > 0 {
		basicMetrics.SuccessRate = float64(kc.successOperations) / float64(kc.totalOperations) * 100
		basicMetrics.RPS = float64(kc.totalOperations) / duration.Seconds()
	}

	if kc.writeOperations > 0 {
		basicMetrics.ReadWriteRatio = float64(kc.readOperations) / float64(kc.writeOperations)
	}

	// 延迟指标
	latencyMetrics := kc.calculateLatencyMetrics()

	// 组装结果
	result := map[string]interface{}{
		"basic_metrics":     basicMetrics,
		"latency_metrics":   latencyMetrics,
		"producer_metrics":  kc.producerMetrics,
		"consumer_metrics":  kc.consumerMetrics,
		"partition_metrics": kc.partitionMetrics,
		"broker_metrics":    kc.brokerMetrics,
		"topic_metrics":     kc.topicMetrics,
		"connection_stats":  kc.connectionStats,
		"window_stats":      kc.windowStats,
		"error_stats":       kc.errorStats,
		"duration":          duration,
		"timestamp":         time.Now(),
	}

	return result
}
func (kc *MetricsCollector) GetMetricsForCore() *interfaces.Metrics {
	kc.mutex.RLock()
	defer kc.mutex.RUnlock()

	duration := time.Since(kc.startTime)

	metrics := &interfaces.Metrics{
		TotalOps:   kc.totalOperations,
		SuccessOps: kc.successOperations,
		FailedOps:  kc.failedOperations,
		ReadOps:    kc.readOperations,
		WriteOps:   kc.writeOperations,
		StartTime:  kc.startTime,
		EndTime:    time.Now(),
		Duration:   duration,
	}

	// 计算延迟指标
	if kc.totalOperations > 0 {
		metrics.AvgLatency = kc.totalLatency / time.Duration(kc.totalOperations)
		metrics.MinLatency = kc.minLatency
		metrics.MaxLatency = kc.maxLatency

		// 计算百分位数
		if len(kc.latencyHistory) > 0 {
			sortedLatencies := make([]time.Duration, len(kc.latencyHistory))
			copy(sortedLatencies, kc.latencyHistory)
			sort.Slice(sortedLatencies, func(i, j int) bool {
				return sortedLatencies[i] < sortedLatencies[j]
			})

			if len(sortedLatencies) > 0 {
				metrics.P90Latency = sortedLatencies[len(sortedLatencies)*90/100]
				metrics.P95Latency = sortedLatencies[len(sortedLatencies)*95/100]
				metrics.P99Latency = sortedLatencies[len(sortedLatencies)*99/100]
			}
		}
	}

	return metrics
}

// Export 实现核心接口（适配 interfaces.MetricsCollector）
func (kc *MetricsCollector) Export() map[string]interface{} {
	metrics := kc.GetMetricsForCore()
	duration := time.Since(kc.startTime)

	// 计算基本指标
	errorRate := float64(0)
	rps := float64(0)
	successRate := float64(0)

	if kc.totalOperations > 0 {
		errorRate = float64(kc.failedOperations) / float64(kc.totalOperations) * 100
		successRate = float64(kc.successOperations) / float64(kc.totalOperations) * 100
		rps = float64(kc.totalOperations) / duration.Seconds()
	}

	return map[string]interface{}{
		"total_ops":    kc.totalOperations,
		"success_ops":  kc.successOperations,
		"failed_ops":   kc.failedOperations,
		"read_ops":     kc.readOperations,
		"write_ops":    kc.writeOperations,
		"error_rate":   errorRate,
		"success_rate": successRate,
		"rps":          rps,
		"avg_latency":  metrics.AvgLatency.Nanoseconds(),
		"min_latency":  metrics.MinLatency.Nanoseconds(),
		"max_latency":  metrics.MaxLatency.Nanoseconds(),
		"p90_latency":  metrics.P90Latency.Nanoseconds(),
		"p95_latency":  metrics.P95Latency.Nanoseconds(),
		"p99_latency":  metrics.P99Latency.Nanoseconds(),
		"duration":     duration.Nanoseconds(),
		"start_time":   kc.startTime.Unix(),
		"end_time":     time.Now().Unix(),
	}
}

// Reset 重置指标
func (kc *MetricsCollector) Reset() {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	kc.totalOperations = 0
	kc.successOperations = 0
	kc.failedOperations = 0
	kc.readOperations = 0
	kc.writeOperations = 0
	kc.totalLatency = 0
	kc.minLatency = time.Duration(^uint64(0) >> 1)
	kc.maxLatency = 0
	kc.latencyHistory = kc.latencyHistory[:0]

	// 重置Kafka特定指标
	for key := range kc.producerMetrics {
		delete(kc.producerMetrics, key)
	}
	for key := range kc.consumerMetrics {
		delete(kc.consumerMetrics, key)
	}
	for key := range kc.partitionMetrics {
		delete(kc.partitionMetrics, key)
	}
	for key := range kc.brokerMetrics {
		delete(kc.brokerMetrics, key)
	}
	for key := range kc.topicMetrics {
		delete(kc.topicMetrics, key)
	}

	// 重置错误统计
	for errorMsg := range kc.errorStats {
		delete(kc.errorStats, errorMsg)
	}

	// 重置连接统计
	kc.connectionStats = &ConnectionStat{}

	// 重置时间窗口
	kc.windowStats = &WindowStat{
		WindowSize:       time.Second,
		WindowOperations: make([]int64, 60),
		LastUpdate:       time.Now(),
	}

	kc.startTime = time.Now()
}

// processKafkaMetadata 处理Kafka特定元数据
func (kc *MetricsCollector) processKafkaMetadata(result *interfaces.OperationResult) {
	metadata := result.Metadata

	// 处理生产者指标
	if operationType, exists := metadata["operation_type"]; exists {
		if opType, ok := operationType.(string); ok {
			switch opType {
			case "produce":
				kc.recordProducerOperation(result)
			case "consume":
				kc.recordConsumerOperation(result)
			}
		}
	}

	// 处理分区信息
	if topic, exists := metadata["topic"]; exists {
		if partition, exists := metadata["partition"]; exists {
			if t, ok := topic.(string); ok {
				if p, ok := partition.(int32); ok {
					kc.recordPartitionOperation(t, p, result)
				}
			}
		}
	}
}

// recordProducerOperation 记录生产者操作
func (kc *MetricsCollector) recordProducerOperation(result *interfaces.OperationResult) {
	clientID := "default"
	if cid, exists := result.Metadata["client_id"]; exists {
		if c, ok := cid.(string); ok {
			clientID = c
		}
	}

	if _, exists := kc.producerMetrics[clientID]; !exists {
		kc.producerMetrics[clientID] = &ProducerStat{
			ProduceLatencies: make([]time.Duration, 0),
		}
	}

	stat := kc.producerMetrics[clientID]

	if result.Success {
		stat.ProducedMessages++
		if messageSize, exists := result.Metadata["message_size"]; exists {
			if size, ok := messageSize.(int64); ok {
				stat.ProducedBytes += size
			}
		}

		stat.ProduceLatencies = append(stat.ProduceLatencies, result.Duration)
		if len(stat.ProduceLatencies) > 1000 {
			stat.ProduceLatencies = stat.ProduceLatencies[1:]
		}

		// 计算平均延迟
		if len(stat.ProduceLatencies) > 0 {
			var total time.Duration
			for _, latency := range stat.ProduceLatencies {
				total += latency
			}
			stat.AvgLatency = total / time.Duration(len(stat.ProduceLatencies))
		}
	} else {
		stat.FailureCount++
	}
}

// recordConsumerOperation 记录消费者操作
func (kc *MetricsCollector) recordConsumerOperation(result *interfaces.OperationResult) {
	clientID := "default"
	if cid, exists := result.Metadata["client_id"]; exists {
		if c, ok := cid.(string); ok {
			clientID = c
		}
	}

	if _, exists := kc.consumerMetrics[clientID]; !exists {
		kc.consumerMetrics[clientID] = &ConsumerStat{
			ConsumeLatencies: make([]time.Duration, 0),
		}
	}

	stat := kc.consumerMetrics[clientID]

	if result.Success {
		stat.ConsumedMessages++
		if messageSize, exists := result.Metadata["message_size"]; exists {
			if size, ok := messageSize.(int64); ok {
				stat.ConsumedBytes += size
			}
		}

		stat.ConsumeLatencies = append(stat.ConsumeLatencies, result.Duration)
		if len(stat.ConsumeLatencies) > 1000 {
			stat.ConsumeLatencies = stat.ConsumeLatencies[1:]
		}

		// 计算平均延迟
		if len(stat.ConsumeLatencies) > 0 {
			var total time.Duration
			for _, latency := range stat.ConsumeLatencies {
				total += latency
			}
			stat.AvgLatency = total / time.Duration(len(stat.ConsumeLatencies))
		}
	}
}

// recordPartitionOperation 记录分区操作
func (kc *MetricsCollector) recordPartitionOperation(topic string, partition int32, result *interfaces.OperationResult) {
	partitionKey := getPartitionKey(topic, partition)

	if _, exists := kc.partitionMetrics[partitionKey]; !exists {
		kc.partitionMetrics[partitionKey] = &PartitionStat{
			Topic:     topic,
			Partition: partition,
		}
	}

	stat := kc.partitionMetrics[partitionKey]

	if result.Success {
		if operationType, exists := result.Metadata["operation_type"]; exists {
			if opType, ok := operationType.(string); ok {
				switch opType {
				case "produce":
					stat.ProducedCount++
					stat.LastProduceTime = time.Now()
				case "consume":
					stat.ConsumedCount++
					stat.LastConsumeTime = time.Now()

					// 更新偏移信息
					if offset, exists := result.Metadata["offset"]; exists {
						if o, ok := offset.(int64); ok {
							stat.ConsumerOffset = o
						}
					}
					if hwm, exists := result.Metadata["high_water_mark"]; exists {
						if h, ok := hwm.(int64); ok {
							stat.HighWaterMark = h
							stat.Lag = h - stat.ConsumerOffset
						}
					}
				}
			}
		}
	}
}

// calculateLatencyMetrics 计算延迟指标
func (kc *MetricsCollector) calculateLatencyMetrics() *LatencyMetrics {
	latencyMetrics := &LatencyMetrics{
		MinLatency:   kc.minLatency,
		MaxLatency:   kc.maxLatency,
		TotalLatency: kc.totalLatency,
	}

	if kc.totalOperations > 0 {
		latencyMetrics.AvgLatency = kc.totalLatency / time.Duration(kc.totalOperations)
	}

	// 计算百分位数
	if len(kc.latencyHistory) > 0 {
		sortedLatencies := make([]time.Duration, len(kc.latencyHistory))
		copy(sortedLatencies, kc.latencyHistory)
		sort.Slice(sortedLatencies, func(i, j int) bool {
			return sortedLatencies[i] < sortedLatencies[j]
		})

		latencyMetrics.P50Latency = kc.getPercentile(sortedLatencies, 50)
		latencyMetrics.P90Latency = kc.getPercentile(sortedLatencies, 90)
		latencyMetrics.P95Latency = kc.getPercentile(sortedLatencies, 95)
		latencyMetrics.P99Latency = kc.getPercentile(sortedLatencies, 99)
	}

	return latencyMetrics
}

// getPercentile 获取百分位数
func (kc *MetricsCollector) getPercentile(sortedLatencies []time.Duration, percentile int) time.Duration {
	if len(sortedLatencies) == 0 {
		return 0
	}

	index := int(float64(len(sortedLatencies)) * float64(percentile) / 100.0)
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}
	if index < 0 {
		index = 0
	}

	return sortedLatencies[index]
}

// updateWindowStats 更新时间窗口统计
func (kc *MetricsCollector) updateWindowStats() {
	now := time.Now()
	timeSinceLastUpdate := now.Sub(kc.windowStats.LastUpdate)

	if timeSinceLastUpdate >= kc.windowStats.WindowSize {
		windowsToMove := int(timeSinceLastUpdate / kc.windowStats.WindowSize)
		for i := 0; i < windowsToMove && i < len(kc.windowStats.WindowOperations); i++ {
			copy(kc.windowStats.WindowOperations, kc.windowStats.WindowOperations[1:])
			kc.windowStats.WindowOperations[len(kc.windowStats.WindowOperations)-1] = 0
		}
		kc.windowStats.LastUpdate = now
	}

	currentIndex := len(kc.windowStats.WindowOperations) - 1
	kc.windowStats.WindowOperations[currentIndex]++

	totalOpsInWindow := int64(0)
	for _, ops := range kc.windowStats.WindowOperations {
		totalOpsInWindow += ops
	}
	windowDuration := time.Duration(len(kc.windowStats.WindowOperations)) * kc.windowStats.WindowSize
	kc.windowStats.RPS = float64(totalOpsInWindow) / windowDuration.Seconds()
}

// GetKafkaSpecificMetrics 获取Kafka特定指标
func (kc *MetricsCollector) GetKafkaSpecificMetrics() map[string]interface{} {
	kc.mutex.RLock()
	defer kc.mutex.RUnlock()

	return map[string]interface{}{
		"producers":   kc.producerMetrics,
		"consumers":   kc.consumerMetrics,
		"partitions":  kc.partitionMetrics,
		"brokers":     kc.brokerMetrics,
		"topics":      kc.topicMetrics,
		"connections": kc.connectionStats,
		"window":      kc.windowStats,
		"errors":      kc.errorStats,
	}
}

// getPartitionKey 获取分区键
func getPartitionKey(topic string, partition int32) string {
	return topic + "-" + string(rune(partition+'0'))
}
