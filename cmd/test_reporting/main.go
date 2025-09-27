package main

import (
	"log"
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"abc-runner/app/reporting"
)

func main() {
	log.Println("Testing new structured reporting system...")

	// 创建模拟指标数据
	config := metrics.DefaultMetricsConfig()
	protocolData := map[string]interface{}{
		"application": "abc-runner",
		"version":     "3.0.0",
		"protocol":    "test",
	}

	baseCollector := metrics.NewBaseCollector(config, protocolData)

	// 模拟一些测试数据
	for i := 0; i < 100; i++ {
		result := &interfaces.OperationResult{
			Success:  i%10 != 0, // 90% 成功率
			Duration: time.Duration(i*10+50) * time.Millisecond,
			IsRead:   i%2 == 0,
		}
		baseCollector.Record(result)
	}

	// 获取快照
	snapshot := baseCollector.Snapshot()

	// 转换为结构化报告
	report := reporting.ConvertFromMetricsSnapshot(snapshot)

	// 创建报告生成器
	config2 := &reporting.RenderConfig{
		OutputFormats: []string{"console", "json"},
		OutputDir:     "./test-reports",
		FilePrefix:    "test_structured_report",
		Timestamp:     true,
	}

	generator := reporting.NewReportGenerator(config2)

	// 生成报告
	if err := generator.Generate(report); err != nil {
		log.Fatalf("Failed to generate report: %v", err)
	}

	log.Println("✅ Structured reporting system test completed successfully!")
}