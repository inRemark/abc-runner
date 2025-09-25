package metrics

import (
	"fmt"
	"sync"
	"time"
	
	"abc-runner/app/core/monitoring"
)

// KafkaMetricsCollector Kafka特定指标收集器
type KafkaMetricsCollector struct {
	*monitoring.EnhancedMetricsCollector
	
	// Kafka特定指标
	kafkaMetrics map[string]interface{}
	mutex        sync.RWMutex
	
	// 生产者指标
	producedMessages      int64
	producedBytes         int64
	produceLatencies      []time.Duration
	produceBatchSizes     []int
	produceErrors         map[string]int64
	
	// 消费者指标
	consumedMessages      int64
	consumedBytes         int64
	consumeLatencies      []time.Duration
	consumerLag          int64
	rebalanceCount       int64
	commitCount          int64
	commitErrors         int64
	
	// 分区指标
	partitionMetrics     map[string]*PartitionMetrics
	
	// 连接指标
	connectionCount      int32
	connectionErrors     int64
	reconnectionCount    int64
}

// PartitionMetrics 分区指标
type PartitionMetrics struct {
	Topic            string        `json:"topic"`
	Partition        int32         `json:"partition"`
	ProducedCount    int64         `json:"produced_count"`
	ConsumedCount    int64         `json:"consumed_count"`
	LastProduceTime  time.Time     `json:"last_produce_time"`
	LastConsumeTime  time.Time     `json:"last_consume_time"`
	LeaderChanges    int64         `json:"leader_changes"`
	HighWaterMark    int64         `json:"high_water_mark"`
	ConsumerOffset   int64         `json:"consumer_offset"`
	Lag              int64         `json:"lag"`
}

// NewKafkaMetricsCollector 创建Kafka指标收集器
func NewKafkaMetricsCollector() *KafkaMetricsCollector {
	return &KafkaMetricsCollector{
		EnhancedMetricsCollector: monitoring.NewEnhancedMetricsCollector(),
		kafkaMetrics:           make(map[string]interface{}),
		produceLatencies:       make([]time.Duration, 0),
		produceBatchSizes:      make([]int, 0),
		produceErrors:          make(map[string]int64),
		consumeLatencies:       make([]time.Duration, 0),
		partitionMetrics:       make(map[string]*PartitionMetrics),
	}
}

// RecordProduceOperation 记录生产操作
func (k *KafkaMetricsCollector) RecordProduceOperation(topic string, partition int32, messageSize int, batchSize int, duration time.Duration, success bool, err error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	// 更新基础指标
	if success {
		k.producedMessages++
		k.producedBytes += int64(messageSize)
		k.produceLatencies = append(k.produceLatencies, duration)
		k.produceBatchSizes = append(k.produceBatchSizes, batchSize)
	} else {
		if err != nil {
			errType := k.categorizeError(err)
			k.produceErrors[errType]++
		}
	}
	
	// 更新分区指标
	partitionKey := k.getPartitionKey(topic, partition)
	if pm, exists := k.partitionMetrics[partitionKey]; exists {
		if success {
			pm.ProducedCount++
			pm.LastProduceTime = time.Now()
		}
	} else {
		k.partitionMetrics[partitionKey] = &PartitionMetrics{
			Topic:           topic,
			Partition:       partition,
			ProducedCount:   1,
			LastProduceTime: time.Now(),
		}
	}
}

// RecordConsumeOperation 记录消费操作
func (k *KafkaMetricsCollector) RecordConsumeOperation(topic string, partition int32, messageSize int, offset int64, duration time.Duration, success bool, err error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	// 更新基础指标
	if success {
		k.consumedMessages++
		k.consumedBytes += int64(messageSize)
		k.consumeLatencies = append(k.consumeLatencies, duration)
	}
	
	// 更新分区指标
	partitionKey := k.getPartitionKey(topic, partition)
	if pm, exists := k.partitionMetrics[partitionKey]; exists {
		if success {
			pm.ConsumedCount++
			pm.LastConsumeTime = time.Now()
			pm.ConsumerOffset = offset
		}
	} else {
		k.partitionMetrics[partitionKey] = &PartitionMetrics{
			Topic:           topic,
			Partition:       partition,
			ConsumedCount:   1,
			LastConsumeTime: time.Now(),
			ConsumerOffset:  offset,
		}
	}
}

// RecordConsumerRebalance 记录消费者重平衡
func (k *KafkaMetricsCollector) RecordConsumerRebalance() {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	k.rebalanceCount++
}

// RecordOffsetCommit 记录偏移提交
func (k *KafkaMetricsCollector) RecordOffsetCommit(success bool) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	k.commitCount++
	if !success {
		k.commitErrors++
	}
}

// RecordConnectionEvent 记录连接事件
func (k *KafkaMetricsCollector) RecordConnectionEvent(eventType string) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	switch eventType {
	case "connect":
		k.connectionCount++
	case "disconnect":
		k.connectionCount--
	case "error":
		k.connectionErrors++
	case "reconnect":
		k.reconnectionCount++
	}
}

// UpdatePartitionLag 更新分区延迟
func (k *KafkaMetricsCollector) UpdatePartitionLag(topic string, partition int32, highWaterMark int64, consumerOffset int64) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	partitionKey := k.getPartitionKey(topic, partition)
	if pm, exists := k.partitionMetrics[partitionKey]; exists {
		pm.HighWaterMark = highWaterMark
		pm.ConsumerOffset = consumerOffset
		pm.Lag = highWaterMark - consumerOffset
	} else {
		k.partitionMetrics[partitionKey] = &PartitionMetrics{
			Topic:          topic,
			Partition:      partition,
			HighWaterMark:  highWaterMark,
			ConsumerOffset: consumerOffset,
			Lag:            highWaterMark - consumerOffset,
		}
	}
	
	// 更新全局延迟
	k.consumerLag = highWaterMark - consumerOffset
}

// GetKafkaSpecificMetrics 获取Kafka特定指标
func (k *KafkaMetricsCollector) GetKafkaSpecificMetrics() map[string]interface{} {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	
	metrics := make(map[string]interface{})
	
	// 生产者指标
	metrics["produced_messages"] = k.producedMessages
	metrics["produced_bytes"] = k.producedBytes
	metrics["produce_rate"] = k.calculateProduceRate()
	metrics["avg_produce_latency"] = k.calculateAvgProduceLatency()
	metrics["p95_produce_latency"] = k.calculateP95ProduceLatency()
	metrics["avg_batch_size"] = k.calculateAvgBatchSize()
	metrics["produce_errors"] = k.produceErrors
	
	// 消费者指标
	metrics["consumed_messages"] = k.consumedMessages
	metrics["consumed_bytes"] = k.consumedBytes
	metrics["consume_rate"] = k.calculateConsumeRate()
	metrics["avg_consume_latency"] = k.calculateAvgConsumeLatency()
	metrics["p95_consume_latency"] = k.calculateP95ConsumeLatency()
	metrics["consumer_lag"] = k.consumerLag
	metrics["rebalance_count"] = k.rebalanceCount
	metrics["commit_count"] = k.commitCount
	metrics["commit_errors"] = k.commitErrors
	metrics["commit_success_rate"] = k.calculateCommitSuccessRate()
	
	// 连接指标
	metrics["active_connections"] = k.connectionCount
	metrics["connection_errors"] = k.connectionErrors
	metrics["reconnection_count"] = k.reconnectionCount
	
	// 分区指标
	metrics["partitions"] = k.getPartitionMetricsSummary()
	
	return metrics
}

// getPartitionKey 获取分区键
func (k *KafkaMetricsCollector) getPartitionKey(topic string, partition int32) string {
	return fmt.Sprintf("%s-%d", topic, partition)
}

// categorizeError 分类错误
func (k *KafkaMetricsCollector) categorizeError(err error) string {
	if err == nil {
		return "unknown"
	}
	
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection"):
		return "connection"
	case contains(errStr, "broker"):
		return "broker"
	case contains(errStr, "topic"):
		return "topic"
	case contains(errStr, "partition"):
		return "partition"
	case contains(errStr, "offset"):
		return "offset"
	default:
		return "other"
	}
}

// calculateProduceRate 计算生产速率
func (k *KafkaMetricsCollector) calculateProduceRate() float64 {
	if len(k.produceLatencies) == 0 {
		return 0
	}
	
	// 简单实现：基于最近的操作计算速率
	recentCount := min(len(k.produceLatencies), 1000)
	if recentCount == 0 {
		return 0
	}
	
	var totalDuration time.Duration
	for i := len(k.produceLatencies) - recentCount; i < len(k.produceLatencies); i++ {
		totalDuration += k.produceLatencies[i]
	}
	
	if totalDuration.Seconds() > 0 {
		return float64(recentCount) / totalDuration.Seconds()
	}
	
	return 0
}

// calculateConsumeRate 计算消费速率
func (k *KafkaMetricsCollector) calculateConsumeRate() float64 {
	if len(k.consumeLatencies) == 0 {
		return 0
	}
	
	// 简单实现：基于最近的操作计算速率
	recentCount := min(len(k.consumeLatencies), 1000)
	if recentCount == 0 {
		return 0
	}
	
	var totalDuration time.Duration
	for i := len(k.consumeLatencies) - recentCount; i < len(k.consumeLatencies); i++ {
		totalDuration += k.consumeLatencies[i]
	}
	
	if totalDuration.Seconds() > 0 {
		return float64(recentCount) / totalDuration.Seconds()
	}
	
	return 0
}

// calculateAvgProduceLatency 计算平均生产延迟
func (k *KafkaMetricsCollector) calculateAvgProduceLatency() time.Duration {
	if len(k.produceLatencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range k.produceLatencies {
		total += latency
	}
	
	return total / time.Duration(len(k.produceLatencies))
}

// calculateAvgConsumeLatency 计算平均消费延迟
func (k *KafkaMetricsCollector) calculateAvgConsumeLatency() time.Duration {
	if len(k.consumeLatencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range k.consumeLatencies {
		total += latency
	}
	
	return total / time.Duration(len(k.consumeLatencies))
}

// calculateP95ProduceLatency 计算P95生产延迟
func (k *KafkaMetricsCollector) calculateP95ProduceLatency() time.Duration {
	if len(k.produceLatencies) == 0 {
		return 0
	}
	
	// 复制并排序
	latencies := make([]time.Duration, len(k.produceLatencies))
	copy(latencies, k.produceLatencies)
	sortDurations(latencies)
	
	index := int(float64(len(latencies)) * 0.95)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	
	return latencies[index]
}

// calculateP95ConsumeLatency 计算P95消费延迟
func (k *KafkaMetricsCollector) calculateP95ConsumeLatency() time.Duration {
	if len(k.consumeLatencies) == 0 {
		return 0
	}
	
	// 复制并排序
	latencies := make([]time.Duration, len(k.consumeLatencies))
	copy(latencies, k.consumeLatencies)
	sortDurations(latencies)
	
	index := int(float64(len(latencies)) * 0.95)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	
	return latencies[index]
}

// calculateAvgBatchSize 计算平均批处理大小
func (k *KafkaMetricsCollector) calculateAvgBatchSize() float64 {
	if len(k.produceBatchSizes) == 0 {
		return 0
	}
	
	var total int
	for _, size := range k.produceBatchSizes {
		total += size
	}
	
	return float64(total) / float64(len(k.produceBatchSizes))
}

// calculateCommitSuccessRate 计算提交成功率
func (k *KafkaMetricsCollector) calculateCommitSuccessRate() float64 {
	if k.commitCount == 0 {
		return 0
	}
	
	successCount := k.commitCount - k.commitErrors
	return float64(successCount) / float64(k.commitCount) * 100
}

// getPartitionMetricsSummary 获取分区指标摘要
func (k *KafkaMetricsCollector) getPartitionMetricsSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	
	totalPartitions := len(k.partitionMetrics)
	totalLag := int64(0)
	maxLag := int64(0)
	
	partitionDetails := make([]map[string]interface{}, 0, totalPartitions)
	
	for _, pm := range k.partitionMetrics {
		totalLag += pm.Lag
		if pm.Lag > maxLag {
			maxLag = pm.Lag
		}
		
		partitionDetails = append(partitionDetails, map[string]interface{}{
			"topic":             pm.Topic,
			"partition":         pm.Partition,
			"produced_count":    pm.ProducedCount,
			"consumed_count":    pm.ConsumedCount,
			"lag":              pm.Lag,
			"high_water_mark":  pm.HighWaterMark,
			"consumer_offset":  pm.ConsumerOffset,
		})
	}
	
	summary["total_partitions"] = totalPartitions
	summary["total_lag"] = totalLag
	summary["max_lag"] = maxLag
	if totalPartitions > 0 {
		summary["avg_lag"] = float64(totalLag) / float64(totalPartitions)
	} else {
		summary["avg_lag"] = 0
	}
	summary["details"] = partitionDetails
	
	return summary
}

// Reset 重置指标
func (k *KafkaMetricsCollector) Reset() {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	
	// 重置基础指标
	if k.EnhancedMetricsCollector != nil {
		k.EnhancedMetricsCollector.Reset()
	}
	
	// 重置Kafka特定指标
	k.producedMessages = 0
	k.producedBytes = 0
	k.produceLatencies = make([]time.Duration, 0)
	k.produceBatchSizes = make([]int, 0)
	k.produceErrors = make(map[string]int64)
	
	k.consumedMessages = 0
	k.consumedBytes = 0
	k.consumeLatencies = make([]time.Duration, 0)
	k.consumerLag = 0
	k.rebalanceCount = 0
	k.commitCount = 0
	k.commitErrors = 0
	
	k.partitionMetrics = make(map[string]*PartitionMetrics)
	k.connectionErrors = 0
	k.reconnectionCount = 0
}

// 辅助函数

// min 返回两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// indexOf 查找子字符串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// sortDurations 排序时间切片
func sortDurations(durations []time.Duration) {
	// 简单的冒泡排序，对于小数据集足够
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}