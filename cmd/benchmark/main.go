package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"abc-runner/app/adapters/redis"
	coreMetrics "abc-runner/app/core/metrics"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/reporting"
)

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	TestName        string        `json:"test_name"`
	Duration        time.Duration `json:"duration"`
	Operations      int64         `json:"operations"`
	OpsPerSecond    float64       `json:"ops_per_second"`
	AvgLatency      time.Duration `json:"avg_latency"`
	MemoryUsed      uint64        `json:"memory_used"`
	GoroutinesUsed  int           `json:"goroutines_used"`
	Success         bool          `json:"success"`
	ErrorMessage    string        `json:"error_message,omitempty"`
}

// PerformanceBenchmark 性能基准测试套件
type PerformanceBenchmark struct {
	results []BenchmarkResult
	mutex   sync.Mutex
}

func NewPerformanceBenchmark() *PerformanceBenchmark {
	return &PerformanceBenchmark{
		results: make([]BenchmarkResult, 0),
	}
}

func (pb *PerformanceBenchmark) AddResult(result BenchmarkResult) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()
	pb.results = append(pb.results, result)
}

func (pb *PerformanceBenchmark) GetResults() []BenchmarkResult {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()
	return append([]BenchmarkResult{}, pb.results...)
}

func main() {
	fmt.Println("开始执行指标系统性能测试和基准测试...")
	
	benchmark := NewPerformanceBenchmark()
	
	// 测试套件
	tests := []struct {
		name string
		test func() BenchmarkResult
	}{
		{"基础收集器性能测试", testBaseCollectorPerformance},
		{"环形缓冲区性能测试", testRingBufferPerformance},
		{"Redis收集器功能测试", testRedisCollectorFunctionality},
		{"内存使用稳定性测试", testMemoryStability},
		{"并发安全性测试", testConcurrencySafety},
		{"大数据量处理测试", testHighVolumeData},
		{"配置系统性能测试", testConfigSystemPerformance},
		{"报告生成性能测试", testReportGenerationPerformance},
	}

	for _, test := range tests {
		fmt.Printf("执行测试: %s\n", test.name)
		result := runBenchmarkTest(test.name, test.test)
		benchmark.AddResult(result)
		
		if result.Success {
			fmt.Printf("✅ %s - %.2f ops/sec, 平均延迟: %v\n", 
				test.name, result.OpsPerSecond, result.AvgLatency)
		} else {
			fmt.Printf("❌ %s - 失败: %s\n", test.name, result.ErrorMessage)
		}
	}

	// 生成性能报告
	generatePerformanceReport(benchmark.GetResults())
	
	fmt.Println("\n所有测试完成!")
}

func runBenchmarkTest(name string, testFunc func() BenchmarkResult) BenchmarkResult {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("测试 %s 发生panic: %v", name, r)
		}
	}()

	return testFunc()
}

// 基础收集器性能测试
func testBaseCollectorPerformance() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)
	goroutinesStart := runtime.NumGoroutine()

	config := coreMetrics.DefaultMetricsConfig()
	collector := coreMetrics.NewBaseCollector(config, map[string]interface{}{})
	defer collector.Stop()

	const numOps = 100000
	
	// 执行操作
	for i := 0; i < numOps; i++ {
		result := &interfaces.OperationResult{
			Success:  i%10 != 0, // 10%失败率
			Duration: time.Duration(i%100+1) * time.Microsecond,
			IsRead:   i%2 == 0,
		}
		collector.Record(result)
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	goroutinesEnd := runtime.NumGoroutine()

	snapshot := collector.Snapshot()
	
	return BenchmarkResult{
		TestName:       "基础收集器性能测试",
		Duration:       duration,
		Operations:     numOps,
		OpsPerSecond:   float64(numOps) / duration.Seconds(),
		AvgLatency:     snapshot.Core.Latency.Average,
		MemoryUsed:     memEnd.TotalAlloc - memStart.TotalAlloc,
		GoroutinesUsed: goroutinesEnd - goroutinesStart,
		Success:        snapshot.Core.Operations.Total == numOps,
	}
}

// 环形缓冲区性能测试
func testRingBufferPerformance() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	rb := coreMetrics.NewRingBuffer[int](10000)
	const numOps = 1000000

	// 执行操作
	for i := 0; i < numOps; i++ {
		rb.Push(i)
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	return BenchmarkResult{
		TestName:     "环形缓冲区性能测试",
		Duration:     duration,
		Operations:   numOps,
		OpsPerSecond: float64(numOps) / duration.Seconds(),
		MemoryUsed:   memEnd.TotalAlloc - memStart.TotalAlloc,
		Success:      rb.Size() <= rb.Capacity(),
	}
}

// Redis收集器功能测试
func testRedisCollectorFunctionality() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	config := coreMetrics.DefaultMetricsConfig()
	collector := redis.NewRedisCollector(config)
	defer collector.Stop()

	const numOps = 50000
	operations := []string{"GET", "SET", "HGET", "HSET", "LPUSH", "RPOP"}

	// 执行不同类型的Redis操作
	for i := 0; i < numOps; i++ {
		result := &interfaces.OperationResult{
			Success:  i%20 != 0, // 5%失败率
			Duration: time.Duration(i%50+1) * time.Microsecond,
			IsRead:   i%2 == 0,
			Metadata: map[string]interface{}{
				"operation_type": operations[i%len(operations)],
			},
		}
		collector.Record(result)
		
		// 模拟连接事件
		if i%100 == 0 {
			collector.RecordConnection(true, time.Duration(i%10+1)*time.Millisecond)
		}
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	redisMetrics := collector.GetRedisMetrics()
	snapshot := collector.Snapshot()

	return BenchmarkResult{
		TestName:     "Redis收集器功能测试",
		Duration:     duration,
		Operations:   numOps,
		OpsPerSecond: float64(numOps) / duration.Seconds(),
		AvgLatency:   snapshot.Core.Latency.Average,
		MemoryUsed:   memEnd.TotalAlloc - memStart.TotalAlloc,
		Success:      len(redisMetrics.Operations) > 0 && snapshot.Core.Operations.Total == numOps,
	}
}

// 内存使用稳定性测试
func testMemoryStability() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	config := coreMetrics.DefaultMetricsConfig()
	config.Latency.HistorySize = 1000 // 限制历史大小
	collector := coreMetrics.NewBaseCollector(config, map[string]interface{}{})
	defer collector.Stop()

	const numOps = 500000 // 大量操作测试内存稳定性
	var maxMemory uint64

	// 分批执行，监控内存使用
	batchSize := 10000
	for batch := 0; batch < numOps/batchSize; batch++ {
		for i := 0; i < batchSize; i++ {
			result := &interfaces.OperationResult{
				Success:  true,
				Duration: time.Duration(i%100+1) * time.Microsecond,
				IsRead:   i%2 == 0,
			}
			collector.Record(result)
		}

		// 检查内存使用
		var memCurrent runtime.MemStats
		runtime.ReadMemStats(&memCurrent)
		if memCurrent.Alloc > maxMemory {
			maxMemory = memCurrent.Alloc
		}

		// 强制GC以测试内存释放
		if batch%10 == 0 {
			runtime.GC()
		}
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	snapshot := collector.Snapshot()

	return BenchmarkResult{
		TestName:     "内存使用稳定性测试",
		Duration:     duration,
		Operations:   numOps,
		OpsPerSecond: float64(numOps) / duration.Seconds(),
		AvgLatency:   snapshot.Core.Latency.Average,
		MemoryUsed:   maxMemory - memStart.Alloc,
		Success:      snapshot.Core.Operations.Total == numOps && memEnd.Alloc < maxMemory*2,
	}
}

// 并发安全性测试
func testConcurrencySafety() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)
	goroutinesStart := runtime.NumGoroutine()

	config := coreMetrics.DefaultMetricsConfig()
	collector := coreMetrics.NewBaseCollector(config, map[string]interface{}{})
	defer collector.Stop()

	const numWorkers = 10
	const opsPerWorker = 10000
	const totalOps = numWorkers * opsPerWorker

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 启动多个并发worker
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for i := 0; i < opsPerWorker; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					result := &interfaces.OperationResult{
						Success:  i%10 != 0,
						Duration: time.Duration(i%100+1) * time.Microsecond,
						IsRead:   i%2 == 0,
						Metadata: map[string]interface{}{
							"worker_id": workerID,
						},
					}
					collector.Record(result)
				}
			}
		}(w)
	}

	wg.Wait()
	duration := time.Since(start)

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	goroutinesEnd := runtime.NumGoroutine()

	snapshot := collector.Snapshot()

	return BenchmarkResult{
		TestName:       "并发安全性测试",
		Duration:       duration,
		Operations:     totalOps,
		OpsPerSecond:   float64(totalOps) / duration.Seconds(),
		AvgLatency:     snapshot.Core.Latency.Average,
		MemoryUsed:     memEnd.TotalAlloc - memStart.TotalAlloc,
		GoroutinesUsed: goroutinesEnd - goroutinesStart,
		Success:        snapshot.Core.Operations.Total == totalOps,
	}
}

// 大数据量处理测试
func testHighVolumeData() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	config := coreMetrics.DefaultMetricsConfig()
	config.Latency.HistorySize = 50000 // 增大历史缓冲区
	collector := coreMetrics.NewBaseCollector(config, map[string]interface{}{})
	defer collector.Stop()

	const numOps = 1000000 // 100万操作

	// 模拟高频操作
	for i := 0; i < numOps; i++ {
		result := &interfaces.OperationResult{
			Success:  i%100 != 0, // 1%失败率
			Duration: time.Duration(i%1000+1) * time.Nanosecond,
			IsRead:   i%3 == 0,
		}
		collector.Record(result)

		// 每10万次操作检查一次
		if i%100000 == 0 && i > 0 {
			snapshot := collector.Snapshot()
			if snapshot.Core.Operations.Total != int64(i+1) {
				return BenchmarkResult{
					TestName:     "大数据量处理测试",
					Success:      false,
					ErrorMessage: fmt.Sprintf("操作计数不匹配: 期望 %d, 实际 %d", i+1, snapshot.Core.Operations.Total),
				}
			}
		}
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	snapshot := collector.Snapshot()

	return BenchmarkResult{
		TestName:     "大数据量处理测试",
		Duration:     duration,
		Operations:   numOps,
		OpsPerSecond: float64(numOps) / duration.Seconds(),
		AvgLatency:   snapshot.Core.Latency.Average,
		MemoryUsed:   memEnd.TotalAlloc - memStart.TotalAlloc,
		Success:      snapshot.Core.Operations.Total == numOps,
	}
}

// 配置系统性能测试
func testConfigSystemPerformance() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	tmpDir := "/tmp/abc_runner_test"
	configPath := tmpDir + "/test_config.yaml"
	
	cm := coreMetrics.NewConfigManager(configPath)
	
	const numOps = 1000

	// 测试配置操作
	for i := 0; i < numOps; i++ {
		err := cm.UpdateConfig(func(config *coreMetrics.MetricsConfig) error {
			config.Latency.HistorySize = 1000 + i
			config.System.MonitorInterval = time.Duration(i+1) * time.Millisecond
			return nil
		})
		if err != nil {
			return BenchmarkResult{
				TestName:     "配置系统性能测试",
				Success:      false,
				ErrorMessage: fmt.Sprintf("配置更新失败: %v", err),
			}
		}
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	config := cm.GetConfig()

	return BenchmarkResult{
		TestName:     "配置系统性能测试",
		Duration:     duration,
		Operations:   numOps,
		OpsPerSecond: float64(numOps) / duration.Seconds(),
		MemoryUsed:   memEnd.TotalAlloc - memStart.TotalAlloc,
		Success:      config.Latency.HistorySize == 1000+numOps-1,
	}
}

// 报告生成性能测试
func testReportGenerationPerformance() BenchmarkResult {
	start := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	// 准备测试数据
	config := coreMetrics.DefaultMetricsConfig()
	collector := coreMetrics.NewBaseCollector(config, map[string]interface{}{})
	defer collector.Stop()

	// 生成一些测试数据
	for i := 0; i < 10000; i++ {
		result := &interfaces.OperationResult{
			Success:  i%10 != 0,
			Duration: time.Duration(i%100+1) * time.Microsecond,
			IsRead:   i%2 == 0,
		}
		collector.Record(result)
	}

	snapshot := collector.Snapshot()

	// 测试报告生成
	reportConfig := reporting.DefaultReportConfig()
	reportConfig.Formats = []reporting.ReportFormat{
		reporting.FormatJSON,
		reporting.FormatCSV,
		reporting.FormatConsole,
	}
	
	generator := reporting.NewUniversalReportGenerator(reportConfig)
	
	const numReports = 100
	
	for i := 0; i < numReports; i++ {
		_, err := generator.Generate(snapshot)
		if err != nil {
			return BenchmarkResult{
				TestName:     "报告生成性能测试",
				Success:      false,
				ErrorMessage: fmt.Sprintf("报告生成失败: %v", err),
			}
		}
	}

	duration := time.Since(start)
	
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	return BenchmarkResult{
		TestName:     "报告生成性能测试",
		Duration:     duration,
		Operations:   numReports,
		OpsPerSecond: float64(numReports) / duration.Seconds(),
		MemoryUsed:   memEnd.TotalAlloc - memStart.TotalAlloc,
		Success:      true,
	}
}

// 生成性能报告
func generatePerformanceReport(results []BenchmarkResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("性能测试报告")
	fmt.Println(strings.Repeat("=", 80))
	
	totalTests := len(results)
	successTests := 0
	
	for _, result := range results {
		if result.Success {
			successTests++
		}
		
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		
		fmt.Printf("%s %s\n", status, result.TestName)
		fmt.Printf("   持续时间: %v\n", result.Duration)
		fmt.Printf("   操作数: %d\n", result.Operations)
		fmt.Printf("   吞吐量: %.2f ops/sec\n", result.OpsPerSecond)
		if result.AvgLatency > 0 {
			fmt.Printf("   平均延迟: %v\n", result.AvgLatency)
		}
		fmt.Printf("   内存使用: %.2f MB\n", float64(result.MemoryUsed)/1024/1024)
		if result.GoroutinesUsed > 0 {
			fmt.Printf("   协程使用: %d\n", result.GoroutinesUsed)
		}
		if !result.Success {
			fmt.Printf("   错误: %s\n", result.ErrorMessage)
		}
		fmt.Println()
	}
	
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("测试总结: %d/%d 通过 (%.1f%%)\n", 
		successTests, totalTests, float64(successTests)/float64(totalTests)*100)
	
	// 计算总体性能指标
	var totalOps int64
	var totalDuration time.Duration
	for _, result := range results {
		if result.Success {
			totalOps += result.Operations
			totalDuration += result.Duration
		}
	}
	
	if totalDuration > 0 {
		overallThroughput := float64(totalOps) / totalDuration.Seconds()
		fmt.Printf("总体吞吐量: %.2f ops/sec\n", overallThroughput)
	}
	
	fmt.Println(strings.Repeat("=", 80))
}