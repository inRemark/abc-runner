package kafka

import (
	"fmt"
	"time"
	
	"github.com/segmentio/kafka-go"
)

// Message Kafka消息结构
type Message struct {
	Key       string            `json:"key"`       // 消息键
	Value     string            `json:"value"`     // 消息值
	Headers   map[string]string `json:"headers"`   // 消息头
	Timestamp time.Time         `json:"timestamp"` // 时间戳
	Partition int32             `json:"partition"` // 分区
	Offset    int64             `json:"offset"`    // 偏移量
	Topic     string            `json:"topic"`     // 主题
}

// ProduceResult 生产结果
type ProduceResult struct {
	Partition int32         `json:"partition"` // 分区
	Offset    int64         `json:"offset"`    // 偏移量
	Timestamp time.Time     `json:"timestamp"` // 时间戳
	Duration  time.Duration `json:"duration"`  // 执行时长
}

// BatchResult 批量操作结果
type BatchResult struct {
	Results       []ProduceResult `json:"results"`        // 详细结果
	SuccessCount  int             `json:"success_count"`  // 成功数量
	FailureCount  int             `json:"failure_count"`  // 失败数量
	TotalDuration time.Duration   `json:"total_duration"` // 总耗时
}

// TopicConfig 主题配置
type TopicConfig struct {
	Name              string            `json:"name"`               // 主题名称
	NumPartitions     int               `json:"num_partitions"`     // 分区数
	ReplicationFactor int               `json:"replication_factor"` // 副本因子
	ConfigEntries     map[string]string `json:"config_entries"`     // 配置项
}

// TopicInfo 主题信息
type TopicInfo struct {
	Name       string             `json:"name"`       // 主题名称
	Partitions []PartitionInfo    `json:"partitions"` // 分区信息
	Config     map[string]string  `json:"config"`     // 配置
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	ID       int32   `json:"id"`       // 分区ID
	Leader   int32   `json:"leader"`   // Leader节点
	Replicas []int32 `json:"replicas"` // 副本节点
	ISR      []int32 `json:"isr"`      // 同步副本集
}

// ConsumerGroupInfo 消费者组信息
type ConsumerGroupInfo struct {
	GroupID      string                    `json:"group_id"`      // 组ID
	State        string                    `json:"state"`         // 状态
	Protocol     string                    `json:"protocol"`      // 协议
	ProtocolType string                    `json:"protocol_type"` // 协议类型
	Members      []ConsumerGroupMember     `json:"members"`       // 成员
	Coordinator  ConsumerGroupCoordinator  `json:"coordinator"`   // 协调器
}

// ConsumerGroupMember 消费者组成员
type ConsumerGroupMember struct {
	MemberID   string   `json:"member_id"`   // 成员ID
	ClientID   string   `json:"client_id"`   // 客户端ID
	ClientHost string   `json:"client_host"` // 客户端主机
	Assignment []string `json:"assignment"`  // 分配的主题
}

// ConsumerGroupCoordinator 消费者组协调器
type ConsumerGroupCoordinator struct {
	NodeID int32  `json:"node_id"` // 节点ID
	Host   string `json:"host"`    // 主机
	Port   int32  `json:"port"`    // 端口
}

// Transaction 事务接口
type Transaction interface {
	// Commit 提交事务
	Commit() error
	
	// Abort 回滚事务
	Abort() error
	
	// GetTransactionID 获取事务ID
	GetTransactionID() string
	
	// IsActive 检查事务是否活跃
	IsActive() bool
}

// KafkaTransaction Kafka事务实现
type KafkaTransaction struct {
	transactionID string
	active        bool
	writer        *kafka.Writer
}

// NewKafkaTransaction 创建Kafka事务
func NewKafkaTransaction(transactionID string, writer *kafka.Writer) *KafkaTransaction {
	return &KafkaTransaction{
		transactionID: transactionID,
		active:        true,
		writer:        writer,
	}
}

// Commit 提交事务
func (t *KafkaTransaction) Commit() error {
	if !t.active {
		return fmt.Errorf("transaction is not active")
	}
	
	// 实际的事务提交逻辑
	// 注意：kafka-go库目前对事务的支持有限，这里是概念性实现
	t.active = false
	return nil
}

// Abort 回滚事务
func (t *KafkaTransaction) Abort() error {
	if !t.active {
		return fmt.Errorf("transaction is not active")
	}
	
	// 实际的事务回滚逻辑
	t.active = false
	return nil
}

// GetTransactionID 获取事务ID
func (t *KafkaTransaction) GetTransactionID() string {
	return t.transactionID
}

// IsActive 检查事务是否活跃
func (t *KafkaTransaction) IsActive() bool {
	return t.active
}

// OperationType 操作类型枚举
type OperationType string

const (
	// 基础操作
	OperationProduce      OperationType = "produce"
	OperationProduceBatch OperationType = "produce_batch"
	OperationConsume      OperationType = "consume"
	OperationConsumeBatch OperationType = "consume_batch"
	
	// 事务操作
	OperationTransactionProduce OperationType = "transaction_produce"
	OperationExactlyOnce        OperationType = "exactly_once"
	
	// 管理操作
	OperationCreateTopic             OperationType = "create_topic"
	OperationDeleteTopic             OperationType = "delete_topic"
	OperationDescribeTopic           OperationType = "describe_topic"
	OperationListConsumerGroups      OperationType = "list_consumer_groups"
	OperationDescribeConsumerGroups  OperationType = "describe_consumer_groups"
	OperationResetOffset             OperationType = "reset_offset"
)

// IsReadOperation 判断是否为读操作
func (op OperationType) IsReadOperation() bool {
	readOperations := []OperationType{
		OperationConsume,
		OperationConsumeBatch,
		OperationDescribeTopic,
		OperationListConsumerGroups,
		OperationDescribeConsumerGroups,
	}
	
	for _, readOp := range readOperations {
		if op == readOp {
			return true
		}
	}
	return false
}

// IsWriteOperation 判断是否为写操作
func (op OperationType) IsWriteOperation() bool {
	return !op.IsReadOperation()
}

// IsManagementOperation 判断是否为管理操作
func (op OperationType) IsManagementOperation() bool {
	managementOperations := []OperationType{
		OperationCreateTopic,
		OperationDeleteTopic,
		OperationDescribeTopic,
		OperationListConsumerGroups,
		OperationDescribeConsumerGroups,
		OperationResetOffset,
	}
	
	for _, mgmtOp := range managementOperations {
		if op == mgmtOp {
			return true
		}
	}
	return false
}