package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// EnhancedRunner 增强的运行引擎
type EnhancedRunner struct {
	adapter           interfaces.ProtocolAdapter
	config            interfaces.Config
	metricsCollector  interfaces.MetricsCollector
	keyGenerator      interfaces.KeyGenerator
	operationRegistry *utils.OperationRegistry
	progressTracker   *utils.ProgressTracker
	// 删除了重复的 retryConfig，由 adapter 内的 errorHandler 处理
}

// NewEnhancedRunner 创建增强运行引擎
func NewEnhancedRunner(
	adapter interfaces.ProtocolAdapter,
	config interfaces.Config,
	metricsCollector interfaces.MetricsCollector,
	keyGenerator interfaces.KeyGenerator,
	operationRegistry *utils.OperationRegistry,
) *EnhancedRunner {
	benchmarkConfig := config.GetBenchmark()
	return &EnhancedRunner{
		adapter:           adapter,
		config:            config,
		metricsCollector:  metricsCollector,
		keyGenerator:      keyGenerator,
		operationRegistry: operationRegistry,
		progressTracker:   utils.NewProgressTracker(int64(benchmarkConfig.GetTotal())),
		// 重试配置由 adapter 内的 errorHandler 处理
	}
}

// RunBenchmark 执行基准测试
func (r *EnhancedRunner) RunBenchmark(ctx context.Context) (*interfaces.Metrics, error) {
	benchmarkConfig := r.config.GetBenchmark()

	total := benchmarkConfig.GetTotal()
	parallels := benchmarkConfig.GetParallels()
	testCase := benchmarkConfig.GetTestCase()

	fmt.Printf("Starting benchmark: %s with %d requests using %d parallel connections\n",
		testCase, total, parallels)

	// 创建任务通道
	taskChan := make(chan struct{}, total)
	for i := 0; i < total; i++ {
		taskChan <- struct{}{}
	}
	close(taskChan)

	// 启动工作协程
	var wg sync.WaitGroup
	wg.Add(parallels)

	start := time.Now()
	for i := 0; i < parallels; i++ {
		go func(workerID int) {
			defer wg.Done()
			r.worker(ctx, workerID, taskChan, testCase)
		}(i)
	}

	// 等待所有任务完成
	wg.Wait()

	// 获取最终指标
	metrics := r.metricsCollector.GetMetrics()
	metrics.Duration = time.Since(start)

	fmt.Printf("\nBenchmark completed in %v\n", metrics.Duration)

	return metrics, nil
}

// worker 工作协程
func (r *EnhancedRunner) worker(ctx context.Context, workerID int, taskChan <-chan struct{}, testCase string) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-taskChan:
			if !ok {
				return // 任务队列已关闭
			}

			// 执行单个操作
			r.executeOperation(ctx, testCase)

			// 更新进度
			r.progressTracker.Update(1)
		}
	}
}

// executeOperation 执行单个操作
func (r *EnhancedRunner) executeOperation(ctx context.Context, testCase string) {
	benchmarkConfig := r.config.GetBenchmark()

	// 准备操作参数
	params := map[string]interface{}{
		"data_size":     benchmarkConfig.GetDataSize(),
		"ttl":           benchmarkConfig.GetTTL(),
		"read_percent":  benchmarkConfig.GetReadPercent(),
		"random_keys":   benchmarkConfig.GetRandomKeys(),
		"total":         benchmarkConfig.GetTotal(),
		"key_generator": r.keyGenerator,
	}

	// 创建操作
	operation, err := r.operationRegistry.CreateOperation(testCase, params)
	if err != nil {
		// 记录错误结果
		r.metricsCollector.RecordOperation(&interfaces.OperationResult{
			Success:  false,
			Duration: 0,
			IsRead:   false,
			Error:    fmt.Errorf("failed to create operation: %w", err),
		})
		return
	}

	// 执行操作
	start := time.Now()
	result, err := r.adapter.Execute(ctx, operation)
	duration := time.Since(start)

	if result == nil {
		result = &interfaces.OperationResult{}
	}

	result.Duration = duration
	if err != nil {
		result.Success = false
		result.Error = err
	} else {
		result.Success = true
	}

	// 记录操作结果
	r.metricsCollector.RecordOperation(result)
}
