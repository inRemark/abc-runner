package kafka

import (
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/utils"
)

// RegisterKafkaOperations 注册Kafka操作
func RegisterKafkaOperations(registry *utils.OperationRegistry) {
	// Kafka Producer 操作
	registry.Register("produce", &ProduceOperationFactory{})
	registry.Register("produce_message", &ProduceMessageOperationFactory{})
	registry.Register("produce_batch", &ProduceBatchOperationFactory{})
	
	// Kafka Consumer 操作
	registry.Register("consume", &ConsumeOperationFactory{})
	registry.Register("consume_message", &ConsumeMessageOperationFactory{})
	registry.Register("consume_batch", &ConsumeBatchOperationFactory{})
	
	// Kafka Admin 操作
	registry.Register("create_topic", &CreateTopicOperationFactory{})
	registry.Register("delete_topic", &DeleteTopicOperationFactory{})
	registry.Register("list_topics", &ListTopicsOperationFactory{})
	registry.Register("describe_consumer_groups", &DescribeConsumerGroupsOperationFactory{})
}

// Kafka操作工厂实现
type ProduceOperationFactory struct{}

func (f *ProduceOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "produce",
		Params: params,
	}, nil
}

func (f *ProduceOperationFactory) GetOperationType() string {
	return "produce"
}

func (f *ProduceOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type ProduceMessageOperationFactory struct{}

func (f *ProduceMessageOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "produce_message",
		Params: params,
	}, nil
}

func (f *ProduceMessageOperationFactory) GetOperationType() string {
	return "produce_message"
}

func (f *ProduceMessageOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type ProduceBatchOperationFactory struct{}

func (f *ProduceBatchOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "produce_batch",
		Params: params,
	}, nil
}

func (f *ProduceBatchOperationFactory) GetOperationType() string {
	return "produce_batch"
}

func (f *ProduceBatchOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type ConsumeOperationFactory struct{}

func (f *ConsumeOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "consume",
		Params: params,
	}, nil
}

func (f *ConsumeOperationFactory) GetOperationType() string {
	return "consume"
}

func (f *ConsumeOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type ConsumeMessageOperationFactory struct{}

func (f *ConsumeMessageOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "consume_message",
		Params: params,
	}, nil
}

func (f *ConsumeMessageOperationFactory) GetOperationType() string {
	return "consume_message"
}

func (f *ConsumeMessageOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type ConsumeBatchOperationFactory struct{}

func (f *ConsumeBatchOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "consume_batch",
		Params: params,
	}, nil
}

func (f *ConsumeBatchOperationFactory) GetOperationType() string {
	return "consume_batch"
}

func (f *ConsumeBatchOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type CreateTopicOperationFactory struct{}

func (f *CreateTopicOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "create_topic",
		Params: params,
	}, nil
}

func (f *CreateTopicOperationFactory) GetOperationType() string {
	return "create_topic"
}

func (f *CreateTopicOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type DeleteTopicOperationFactory struct{}

func (f *DeleteTopicOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "delete_topic",
		Params: params,
	}, nil
}

func (f *DeleteTopicOperationFactory) GetOperationType() string {
	return "delete_topic"
}

func (f *DeleteTopicOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type ListTopicsOperationFactory struct{}

func (f *ListTopicsOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "list_topics",
		Params: params,
	}, nil
}

func (f *ListTopicsOperationFactory) GetOperationType() string {
	return "list_topics"
}

func (f *ListTopicsOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type DescribeConsumerGroupsOperationFactory struct{}

func (f *DescribeConsumerGroupsOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	return interfaces.Operation{
		Type:   "describe_consumer_groups",
		Params: params,
	}, nil
}

func (f *DescribeConsumerGroupsOperationFactory) GetOperationType() string {
	return "describe_consumer_groups"
}

func (f *DescribeConsumerGroupsOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}