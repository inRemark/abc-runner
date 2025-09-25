package kafka

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// KafkaMetrics Kafka协议特定指标
type KafkaMetrics struct {
	// 生产者指标
	Producers map[string]*KafkaProducerStat `json:"producers"`
	
	// 消费者指标
	Consumers map[string]*KafkaConsumerStat `json:"consumers"`
	
	// 主题指标
	Topics map[string]*KafkaTopicStat `json:"topics"`
	
	// 分区指标
	Partitions map[string]*KafkaPartitionStat `json:"partitions"`
	
	// Broker指标
	Brokers map[string]*KafkaBrokerStat `json:"brokers"`
	
	// 连接指标
	Connection *KafkaConnectionStat `json:"connection"`
	
	// 性能指标
	Performance *KafkaPerformanceStat `json:"performance"`
}

// KafkaProducerStat Kafka生产者统计
type KafkaProducerStat struct {
	ProducedMessages  int64         `json:"produced_messages"`
	ProducedBytes     int64         `json:"produced_bytes"`
	BatchCount        int64         `json:"batch_count"`
	AvgBatchSize      float64       `json:"avg_batch_size"`
	AvgLatency        time.Duration `json:"avg_latency"`
	FailureCount      int64         `json:"failure_count"`
	RetryCount        int64         `json:"retry_count"`
	Throughput        float64       `json:"throughput"` // msgs/sec
}

// KafkaConsumerStat Kafka消费者统计
type KafkaConsumerStat struct {
	ConsumedMessages  int64         `json:"consumed_messages"`
	ConsumedBytes     int64         `json:"consumed_bytes"`
	AvgLatency        time.Duration `json:"avg_latency"`
	LagTotal          int64         `json:"lag_total"`
	OffsetCommits     int64         `json:"offset_commits"`
	Rebalances        int64         `json:"rebalances"`
	Throughput        float64       `json:"throughput"` // msgs/sec
}

// KafkaTopicStat Kafka主题统计
type KafkaTopicStat struct {
	MessageCount      int64   `json:"message_count"`
	BytesCount        int64   `json:"bytes_count"`
	PartitionCount    int32   `json:"partition_count"`
	ReplicationFactor int16   `json:"replication_factor"`
	ProduceRate       float64 `json:"produce_rate"`
	ConsumeRate       float64 `json:"consume_rate"`
}

// KafkaPartitionStat Kafka分区统计
type KafkaPartitionStat struct {
	Topic           string  `json:"topic"`
	Partition       int32   `json:"partition"`
	Offset          int64   `json:"offset"`
	HighWaterMark   int64   `json:"high_water_mark"`
	Lag             int64   `json:"lag"`
	Leader          int32   `json:"leader"`
	Replicas        []int32 `json:"replicas"`
	MessageRate     float64 `json:"message_rate"`
}

// KafkaBrokerStat Kafka Broker统计
type KafkaBrokerStat struct {
	BrokerID        int32   `json:"broker_id"`
	Host            string  `json:"host"`
	Port            int32   `json:"port"`
	ConnectionCount int64   `json:"connection_count"`
	RequestRate     float64 `json:"request_rate"`
	ResponseTime    time.Duration `json:"response_time"`
	IsController    bool    `json:"is_controller"`
}

// KafkaConnectionStat Kafka连接统计
type KafkaConnectionStat struct {
	ActiveConnections   int32         `json:"active_connections"`
	TotalConnections    int64         `json:"total_connections"`
	FailedConnections   int64         `json:"failed_connections"`
	AvgConnectionTime   time.Duration `json:"avg_connection_time"`
	MetadataRefreshes   int64         `json:"metadata_refreshes"`
}

// KafkaPerformanceStat Kafka性能统计
type KafkaPerformanceStat struct {
	MessagesPerSecond   float64       `json:"messages_per_second"`
	BytesPerSecond      float64       `json:"bytes_per_second"`
	AvgMessageSize      float64       `json:"avg_message_size"`
	CompressionRatio    float64       `json:"compression_ratio"`
	EndToEndLatency     time.Duration `json:"end_to_end_latency"`
	ProduceLatency      time.Duration `json:"produce_latency"`
	FetchLatency        time.Duration `json:"fetch_latency"`
}

// KafkaCollector Kafka指标收集器
type KafkaCollector struct {
	*metrics.BaseCollector[KafkaMetrics]
	
	// Kafka特定指标
	kafkaMetrics *KafkaMetrics
	mutex        sync.RWMutex
	
	// 追踪器
	producerTracker    *KafkaProducerTracker
	consumerTracker    *KafkaConsumerTracker
	topicTracker       *KafkaTopicTracker
	partitionTracker   *KafkaPartitionTracker
	brokerTracker      *KafkaBrokerTracker
	connectionTracker  *KafkaConnectionTracker
	performanceTracker *KafkaPerformanceTracker
	
	// 配置
	config *metrics.MetricsConfig
}

// NewKafkaCollector 创建Kafka指标收集器
func NewKafkaCollector(config *metrics.MetricsConfig) *KafkaCollector {
	if config == nil {
		config = metrics.DefaultMetricsConfig()
	}

	// 初始化Kafka指标
	kafkaMetrics := &KafkaMetrics{
		Producers:   make(map[string]*KafkaProducerStat),
		Consumers:   make(map[string]*KafkaConsumerStat),
		Topics:      make(map[string]*KafkaTopicStat),
		Partitions:  make(map[string]*KafkaPartitionStat),
		Brokers:     make(map[string]*KafkaBrokerStat),
		Connection:  &KafkaConnectionStat{},
		Performance: &KafkaPerformanceStat{},
	}

	// 创建基础收集器
	baseCollector := metrics.NewBaseCollector(config, *kafkaMetrics)

	collector := &KafkaCollector{
		BaseCollector:      baseCollector,
		kafkaMetrics:       kafkaMetrics,
		producerTracker:    NewKafkaProducerTracker(),
		consumerTracker:    NewKafkaConsumerTracker(),
		topicTracker:       NewKafkaTopicTracker(),
		partitionTracker:   NewKafkaPartitionTracker(),
		brokerTracker:      NewKafkaBrokerTracker(),
		connectionTracker:  NewKafkaConnectionTracker(),
		performanceTracker: NewKafkaPerformanceTracker(),
		config:            config,
	}

	return collector
}

// Record 记录操作结果
func (kc *KafkaCollector) Record(result *interfaces.OperationResult) {
	// 调用基础记录方法
	kc.BaseCollector.Record(result)

	// Kafka特定记录
	kc.recordKafkaOperation(result)
}

// recordKafkaOperation 记录Kafka特定操作
func (kc *KafkaCollector) recordKafkaOperation(result *interfaces.OperationResult) {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	metadata := result.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// 记录生产者指标
	if producerID, ok := metadata["producer_id"].(string); ok {
		kc.producerTracker.Record(producerID, result)
	}

	// 记录消费者指标
	if consumerID, ok := metadata["consumer_id"].(string); ok {
		kc.consumerTracker.Record(consumerID, result)
	}

	// 记录主题指标
	if topic, ok := metadata["topic"].(string); ok {
		kc.topicTracker.Record(topic, result)
	}

	// 记录分区指标
	if partition, ok := metadata["partition"].(int32); ok {
		if topic, ok := metadata["topic"].(string); ok {
			kc.partitionTracker.Record(topic, partition, result)
		}
	}

	// 记录性能指标
	kc.performanceTracker.Record(result)

	// 更新指标
	kc.updateKafkaMetrics()
}

// updateKafkaMetrics 更新Kafka指标
func (kc *KafkaCollector) updateKafkaMetrics() {
	kc.kafkaMetrics.Producers = kc.producerTracker.GetStats()
	kc.kafkaMetrics.Consumers = kc.consumerTracker.GetStats()
	kc.kafkaMetrics.Topics = kc.topicTracker.GetStats()
	kc.kafkaMetrics.Partitions = kc.partitionTracker.GetStats()
	kc.kafkaMetrics.Brokers = kc.brokerTracker.GetStats()
	kc.kafkaMetrics.Connection = kc.connectionTracker.GetStats()
	kc.kafkaMetrics.Performance = kc.performanceTracker.GetStats()

	// 更新基础收集器的协议数据
	kc.UpdateProtocolMetrics(*kc.kafkaMetrics)
}

// GetKafkaMetrics 获取Kafka特定指标
func (kc *KafkaCollector) GetKafkaMetrics() *KafkaMetrics {
	kc.mutex.RLock()
	defer kc.mutex.RUnlock()

	// 创建深拷贝
	metricsCopy := &KafkaMetrics{
		Producers:   make(map[string]*KafkaProducerStat),
		Consumers:   make(map[string]*KafkaConsumerStat),
		Topics:      make(map[string]*KafkaTopicStat),
		Partitions:  make(map[string]*KafkaPartitionStat),
		Brokers:     make(map[string]*KafkaBrokerStat),
		Connection:  &KafkaConnectionStat{},
		Performance: &KafkaPerformanceStat{},
	}

	// 深拷贝所有映射
	for id, stat := range kc.kafkaMetrics.Producers {
		statCopy := *stat
		metricsCopy.Producers[id] = &statCopy
	}
	for id, stat := range kc.kafkaMetrics.Consumers {
		statCopy := *stat
		metricsCopy.Consumers[id] = &statCopy
	}
	for name, stat := range kc.kafkaMetrics.Topics {
		statCopy := *stat
		metricsCopy.Topics[name] = &statCopy
	}
	for key, stat := range kc.kafkaMetrics.Partitions {
		statCopy := *stat
		metricsCopy.Partitions[key] = &statCopy
	}
	for id, stat := range kc.kafkaMetrics.Brokers {
		statCopy := *stat
		metricsCopy.Brokers[id] = &statCopy
	}

	*metricsCopy.Connection = *kc.kafkaMetrics.Connection
	*metricsCopy.Performance = *kc.kafkaMetrics.Performance

	return metricsCopy
}

// Export 导出指标
func (kc *KafkaCollector) Export() map[string]interface{} {
	baseMetrics := kc.BaseCollector.Snapshot()
	kafkaMetrics := kc.GetKafkaMetrics()

	result := make(map[string]interface{})
	result["core"] = baseMetrics.Core
	result["system"] = baseMetrics.System
	result["kafka"] = kafkaMetrics
	result["protocol"] = "kafka"
	result["timestamp"] = baseMetrics.Timestamp
	result["duration"] = baseMetrics.Core.Duration

	return result
}

// Reset 重置所有指标
func (kc *KafkaCollector) Reset() {
	kc.BaseCollector.Reset()

	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	kc.kafkaMetrics = &KafkaMetrics{
		Producers:   make(map[string]*KafkaProducerStat),
		Consumers:   make(map[string]*KafkaConsumerStat),
		Topics:      make(map[string]*KafkaTopicStat),
		Partitions:  make(map[string]*KafkaPartitionStat),
		Brokers:     make(map[string]*KafkaBrokerStat),
		Connection:  &KafkaConnectionStat{},
		Performance: &KafkaPerformanceStat{},
	}

	// 重置追踪器
	kc.producerTracker.Reset()
	kc.consumerTracker.Reset()
	kc.topicTracker.Reset()
	kc.partitionTracker.Reset()
	kc.brokerTracker.Reset()
	kc.connectionTracker.Reset()
	kc.performanceTracker.Reset()
}

// Validate 验证指标数据的有效性
func (km *KafkaMetrics) Validate() error {
	if km.Producers == nil {
		return fmt.Errorf("producers map is nil")
	}
	if km.Consumers == nil {
		return fmt.Errorf("consumers map is nil")
	}
	if km.Topics == nil {
		return fmt.Errorf("topics map is nil")
	}
	if km.Partitions == nil {
		return fmt.Errorf("partitions map is nil")
	}
	if km.Brokers == nil {
		return fmt.Errorf("brokers map is nil")
	}
	if km.Connection == nil {
		return fmt.Errorf("connection stats is nil")
	}
	if km.Performance == nil {
		return fmt.Errorf("performance stats is nil")
	}
	return nil
}

// String 返回字符串表示
func (km *KafkaMetrics) String() string {
	data, _ := json.MarshalIndent(km, "", "  ")
	return string(data)
}