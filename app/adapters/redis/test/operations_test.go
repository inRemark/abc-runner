package test

import (
	"testing"

	"redis-runner/app/adapters/redis/operations"
)

func TestIncrOperation(t *testing.T) {
	// 创建INCR操作
	incrOp := operations.NewIncrOperation()
	
	// 验证操作类型
	if incrOp.GetType() != operations.OperationIncr {
		t.Errorf("Expected operation type 'incr', got '%s'", incrOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{
		Total: 100,
		RandomKeys: 10,
		DataSize: 10,
	}
	
	if err := incrOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestDecrOperation(t *testing.T) {
	// 创建DECR操作
	decrOp := operations.NewDecrOperation()
	
	// 验证操作类型
	if decrOp.GetType() != operations.OperationDecr {
		t.Errorf("Expected operation type 'decr', got '%s'", decrOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{
		Total: 100,
		RandomKeys: 10,
		DataSize: 10,
	}
	
	if err := decrOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestListOperations(t *testing.T) {
	// 测试LPUSH操作
	lpushOp := operations.NewListOperation(operations.OperationLPush)
	if lpushOp.GetType() != operations.OperationLPush {
		t.Errorf("Expected operation type 'lpush', got '%s'", lpushOp.GetType())
	}
	
	// 测试RPUSH操作
	rpushOp := operations.NewListOperation(operations.OperationRPush)
	if rpushOp.GetType() != operations.OperationRPush {
		t.Errorf("Expected operation type 'rpush', got '%s'", rpushOp.GetType())
	}
	
	// 测试LPOP操作
	lpopOp := operations.NewListOperation(operations.OperationLPop)
	if lpopOp.GetType() != operations.OperationLPop {
		t.Errorf("Expected operation type 'lpop', got '%s'", lpopOp.GetType())
	}
	
	// 测试RPOP操作
	rpopOp := operations.NewListOperation(operations.OperationRPop)
	if rpopOp.GetType() != operations.OperationRPop {
		t.Errorf("Expected operation type 'rpop', got '%s'", rpopOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{
		Total: 100,
		RandomKeys: 10,
		DataSize: 10,
	}
	
	// PUSH操作应该通过验证
	if err := lpushOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for lpush: %v", err)
	}
	
	// POP操作应该通过验证（不需要DataSize）
	params.DataSize = 0
	if err := lpopOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for lpop: %v", err)
	}
}

func TestSetOperations(t *testing.T) {
	// 测试SADD操作
	saddOp := operations.NewRedisSetOperation(operations.OperationSAdd)
	if saddOp.GetType() != operations.OperationSAdd {
		t.Errorf("Expected operation type 'sadd', got '%s'", saddOp.GetType())
	}
	
	// 测试SMEMBERS操作
	smembersOp := operations.NewRedisSetOperation(operations.OperationSMembers)
	if smembersOp.GetType() != operations.OperationSMembers {
		t.Errorf("Expected operation type 'smembers', got '%s'", smembersOp.GetType())
	}
	
	// 测试SREM操作
	sremOp := operations.NewRedisSetOperation(operations.OperationSRem)
	if sremOp.GetType() != operations.OperationSRem {
		t.Errorf("Expected operation type 'srem', got '%s'", sremOp.GetType())
	}
	
	// 测试SISMEMBER操作
	sismemberOp := operations.NewRedisSetOperation(operations.OperationSIsMember)
	if sismemberOp.GetType() != operations.OperationSIsMember {
		t.Errorf("Expected operation type 'sismember', got '%s'", sismemberOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{
		Total: 100,
		RandomKeys: 10,
		DataSize: 10,
	}
	
	// SADD操作应该通过验证
	if err := saddOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for sadd: %v", err)
	}
	
	// SMEMBERS操作应该通过验证（不需要DataSize）
	params.DataSize = 0
	if err := smembersOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for smembers: %v", err)
	}
}

func TestSortedSetOperations(t *testing.T) {
	// 测试ZADD操作
	zaddOp := operations.NewSortedSetOperation(operations.OperationZAdd)
	if zaddOp.GetType() != operations.OperationZAdd {
		t.Errorf("Expected operation type 'zadd', got '%s'", zaddOp.GetType())
	}
	
	// 测试ZRANGE操作
	zrangeOp := operations.NewSortedSetOperation(operations.OperationZRange)
	if zrangeOp.GetType() != operations.OperationZRange {
		t.Errorf("Expected operation type 'zrange', got '%s'", zrangeOp.GetType())
	}
	
	// 测试ZREM操作
	zremOp := operations.NewSortedSetOperation(operations.OperationZRem)
	if zremOp.GetType() != operations.OperationZRem {
		t.Errorf("Expected operation type 'zrem', got '%s'", zremOp.GetType())
	}
	
	// 测试ZRANK操作
	zrankOp := operations.NewSortedSetOperation(operations.OperationZRank)
	if zrankOp.GetType() != operations.OperationZRank {
		t.Errorf("Expected operation type 'zrank', got '%s'", zrankOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{
		Total: 100,
		RandomKeys: 10,
		DataSize: 10,
	}
	
	// ZADD操作应该通过验证
	if err := zaddOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for zadd: %v", err)
	}
	
	// ZRANGE操作应该通过验证（不需要DataSize）
	params.DataSize = 0
	if err := zrangeOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for zrange: %v", err)
	}
}

func TestHashOperations(t *testing.T) {
	// 测试HSET操作
	hsetOp := operations.NewHashOperation(operations.OperationHSet)
	if hsetOp.GetType() != operations.OperationHSet {
		t.Errorf("Expected operation type 'hset', got '%s'", hsetOp.GetType())
	}
	
	// 测试HMSET操作
	hmsetOp := operations.NewHashOperation(operations.OperationHMSet)
	if hmsetOp.GetType() != operations.OperationHMSet {
		t.Errorf("Expected operation type 'hmset', got '%s'", hmsetOp.GetType())
	}
	
	// 测试HMGET操作
	hmgetOp := operations.NewHashOperation(operations.OperationHMGet)
	if hmgetOp.GetType() != operations.OperationHMGet {
		t.Errorf("Expected operation type 'hmget', got '%s'", hmgetOp.GetType())
	}
	
	// 测试HGETALL操作
	hgetallOp := operations.NewHashOperation(operations.OperationHGetAll)
	if hgetallOp.GetType() != operations.OperationHGetAll {
		t.Errorf("Expected operation type 'hgetall', got '%s'", hgetallOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{
		Total: 100,
		RandomKeys: 10,
		DataSize: 10,
	}
	
	// HSET操作应该通过验证
	if err := hsetOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for hset: %v", err)
	}
	
	// HMSET操作应该通过验证
	if err := hmsetOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for hmset: %v", err)
	}
	
	// HMGET操作应该通过验证（不需要DataSize）
	params.DataSize = 0
	if err := hmgetOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for hmget: %v", err)
	}
}

func TestSubscribeOperations(t *testing.T) {
	// 测试SUBSCRIBE操作
	subOp := operations.NewSubscribeOperation("test_channel")
	if subOp.GetType() != operations.OperationSubscribe {
		t.Errorf("Expected operation type 'sub', got '%s'", subOp.GetType())
	}
	
	// 测试UNSUBSCRIBE操作
	unsubOp := operations.NewUnsubscribeOperation("test_channel")
	if unsubOp.GetType() != operations.OperationUnsubscribe {
		t.Errorf("Expected operation type 'unsub', got '%s'", unsubOp.GetType())
	}
	
	// 验证参数验证
	params := operations.OperationParams{}
	
	if err := subOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for subscribe: %v", err)
	}
	
	if err := unsubOp.Validate(params); err != nil {
		t.Errorf("Unexpected validation error for unsubscribe: %v", err)
	}
}

func TestOperationFactory(t *testing.T) {
	// 创建操作工厂
	factory := operations.NewOperationFactory()
	
	// 测试所有新操作的创建
	operationsToTest := []operations.OperationType{
		operations.OperationIncr,
		operations.OperationDecr,
		operations.OperationLPush,
		operations.OperationRPush,
		operations.OperationLPop,
		operations.OperationRPop,
		operations.OperationSAdd,
		operations.OperationSRem,
		operations.OperationSMembers,
		operations.OperationSIsMember,
		operations.OperationZAdd,
		operations.OperationZRem,
		operations.OperationZRange,
		operations.OperationZRank,
		operations.OperationHMSet,
		operations.OperationHMGet,
		operations.OperationHGetAll,
		operations.OperationSubscribe,
		operations.OperationUnsubscribe,
	}
	
	for _, opType := range operationsToTest {
		_, err := factory.Create(opType)
		if err != nil {
			t.Errorf("Failed to create operation '%s': %v", opType, err)
		}
	}
}