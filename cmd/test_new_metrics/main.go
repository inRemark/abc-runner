package main

import (
	"fmt"
	"time"

	"abc-runner/app/adapters/http"
	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/redis"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"abc-runner/app/core/reporting"
)

func main() {
	fmt.Println("🚀 测试新的指标系统架构...")

	// 测试基础收集器
	testBaseCollector()

	// 测试Redis收集器
	testRedisCollector()

	// 测试HTTP收集器
	testHttpCollector()

	// 测试Kafka收集器
	testKafkaCollector()

	// 测试报告生成
	testReportGeneration()

	fmt.Println("✅ 所有测试完成，新架构工作正常！")
}

func testBaseCollector() {
	fmt.Println("\n📊 测试基础收集器...")

	config := metrics.DefaultMetricsConfig()
	config.Latency.HistorySize = 100

	collector := metrics.NewBaseCollector(config, map[string]interface{}{
		"test_protocol": "demo",
	})
	defer collector.Stop()

	// 模拟一些操作
	for i := 0; i < 50; i++ {
		result := &interfaces.OperationResult{
			Success:  i%10 != 0, // 10%失败率
			Duration: time.Duration(i+1) * time.Millisecond,
			IsRead:   i%2 == 0,
			Metadata: map[string]interface{}{
				"operation_id": i,
			},
		}
		collector.Record(result)
	}

	snapshot := collector.Snapshot()
	fmt.Printf("  - 总操作数: %d\n", snapshot.Core.Operations.Total)
	fmt.Printf("  - 成功率: %.1f%%\n", snapshot.Core.Operations.Rate)
	fmt.Printf("  - 平均延迟: %v\n", snapshot.Core.Latency.Average)
	fmt.Printf("  - P95延迟: %v\n", snapshot.Core.Latency.P95)
	fmt.Printf("  - 吞吐量: %.2f ops/sec\n", snapshot.Core.Throughput.RPS)

	// 测试健康状况
	health := collector.GetHealthStatus()
	fmt.Printf("  - 健康状态: %s\n", health.Status)
}

func testRedisCollector() {
	fmt.Println("\n🔴 测试Redis收集器...")

	config := metrics.DefaultMetricsConfig()
	collector := redis.NewRedisCollector(config)
	defer collector.Stop()

	// 模拟Redis操作
	operations := []string{"GET", "SET", "HGET", "HSET", "LPUSH", "RPOP"}
	for i := 0; i < 30; i++ {
		result := &interfaces.OperationResult{
			Success:  i%15 != 0, // 6.7%失败率
			Duration: time.Duration(i%20+1) * time.Millisecond,
			IsRead:   i%2 == 0,
			Metadata: map[string]interface{}{
				"operation_type": operations[i%len(operations)],
			},
		}
		collector.Record(result)

		// 模拟连接事件
		if i%10 == 0 {
			collector.RecordConnection(true, time.Duration(i%5+1)*time.Millisecond)
		}
	}

	redisMetrics := collector.GetRedisMetrics()
	fmt.Printf("  - Redis操作类型数: %d\n", len(redisMetrics.Operations))
	fmt.Printf("  - 连接统计可用: %t\n", redisMetrics.Connection != nil)
	fmt.Printf("  - 性能统计可用: %t\n", redisMetrics.Performance != nil)

	summary := collector.GetSummary()
	fmt.Printf("  - 协议: %s\n", summary["protocol"])
	fmt.Printf("  - QPS: %.2f\n", summary["qps"])
}

func testHttpCollector() {
	fmt.Println("\n🌐 测试HTTP收集器...")

	config := metrics.DefaultMetricsConfig()
	collector := http.NewHttpCollector(config)
	defer collector.Stop()

	// 模拟HTTP操作
	statusCodes := []int{200, 201, 404, 500}
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	for i := 0; i < 40; i++ {
		result := &interfaces.OperationResult{
			Success:  statusCodes[i%len(statusCodes)] < 400,
			Duration: time.Duration(i%30+1) * time.Millisecond,
			IsRead:   methods[i%len(methods)] == "GET",
			Metadata: map[string]interface{}{
				"status_code":  statusCodes[i%len(statusCodes)],
				"method":       methods[i%len(methods)],
				"url":          fmt.Sprintf("/api/v1/resource/%d", i%5),
				"content_type": "application/json",
			},
		}
		collector.Record(result)
	}

	httpMetrics := collector.GetHttpMetrics()
	fmt.Printf("  - HTTP状态码类型数: %d\n", len(httpMetrics.StatusCodes))
	fmt.Printf("  - HTTP方法类型数: %d\n", len(httpMetrics.Methods))
	fmt.Printf("  - URL路径数: %d\n", len(httpMetrics.URLs))

	summary := collector.GetSummary()
	fmt.Printf("  - 协议: %s\n", summary["protocol"])
	fmt.Printf("  - 请求每秒: %.2f\n", summary["requests_per_sec"])
}

func testKafkaCollector() {
	fmt.Println("\n📨 测试Kafka收集器...")

	config := metrics.DefaultMetricsConfig()
	collector := kafka.NewKafkaCollector(config)
	defer collector.Stop()

	// 模拟Kafka操作
	topics := []string{"user-events", "order-events", "system-logs"}
	for i := 0; i < 25; i++ {
		result := &interfaces.OperationResult{
			Success:  i%20 != 0, // 5%失败率
			Duration: time.Duration(i%25+1) * time.Millisecond,
			IsRead:   i%3 == 0, // 消费者操作
			Metadata: map[string]interface{}{
				"topic":       topics[i%len(topics)],
				"partition":   int32(i % 3),
				"producer_id": fmt.Sprintf("producer-%d", i%2),
				"consumer_id": fmt.Sprintf("consumer-%d", i%2),
			},
		}
		collector.Record(result)
	}

	kafkaMetrics := collector.GetKafkaMetrics()
	fmt.Printf("  - Kafka生产者数: %d\n", len(kafkaMetrics.Producers))
	fmt.Printf("  - Kafka消费者数: %d\n", len(kafkaMetrics.Consumers))
	fmt.Printf("  - Kafka主题数: %d\n", len(kafkaMetrics.Topics))

	snapshot := collector.Snapshot()
	fmt.Printf("  - 协议: kafka\n")
	fmt.Printf("  - 消息每秒: %.2f\n", snapshot.Core.Throughput.RPS)
}

func testReportGeneration() {
	fmt.Println("\n📋 测试报告生成...")

	// 创建一个收集器并添加数据
	config := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(config, map[string]interface{}{
		"test_data": "report_generation",
	})
	defer collector.Stop()

	// 添加一些测试数据
	for i := 0; i < 10; i++ {
		result := &interfaces.OperationResult{
			Success:  true,
			Duration: time.Duration(i+1) * time.Millisecond,
			IsRead:   i%2 == 0,
		}
		collector.Record(result)
	}

	snapshot := collector.Snapshot()

	// 配置报告生成器
	reportConfig := reporting.DefaultReportConfig()
	reportConfig.Formats = []reporting.ReportFormat{
		reporting.FormatJSON,
		reporting.FormatConsole,
	}
	reportConfig.OutputDir = "/tmp/abc_runner_reports"

	generator := reporting.NewUniversalReportGenerator(reportConfig)

	// 生成报告
	report, err := generator.Generate(snapshot)
	if err != nil {
		fmt.Printf("  ❌ 报告生成失败: %v\n", err)
		return
	}

	fmt.Printf("  - 报告格式: %s\n", report.Format)
	fmt.Printf("  - 生成时间: %s\n", report.Metadata.GeneratedAt.Format("15:04:05"))
	if report.FilePath != "" {
		fmt.Printf("  - 保存路径: %s\n", report.FilePath)
	}
	fmt.Printf("  - 协议类型: %s\n", report.Metadata.Protocol)
}
