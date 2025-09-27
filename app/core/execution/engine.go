package execution

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// BenchmarkConfig 基准测试配置接口
type BenchmarkConfig interface {
	GetTotal() int              // 总操作数
	GetParallels() int          // 并发数
	GetDuration() time.Duration // 测试持续时间
	GetTimeout() time.Duration  // 操作超时时间
	GetRampUp() time.Duration   // 渐进加载时间
}

// Job 表示一个待执行的任务
type Job struct {
	ID        int                    // 任务ID
	Operation interfaces.Operation  // 操作定义
	Context   context.Context       // 执行上下文
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	TotalJobs     int64         // 总任务数
	CompletedJobs int64         // 完成任务数
	SuccessJobs   int64         // 成功任务数
	FailedJobs    int64         // 失败任务数
	TotalDuration time.Duration // 总执行时间
	StartTime     time.Time     // 开始时间
	EndTime       time.Time     // 结束时间
}

// OperationFactory 操作工厂接口
type OperationFactory interface {
	CreateOperation(jobID int, config BenchmarkConfig) interfaces.Operation
}

// ExecutionEngine 通用执行引擎
type ExecutionEngine struct {
	adapter          interfaces.ProtocolAdapter       // 协议适配器
	metricsCollector interfaces.DefaultMetricsCollector // 指标收集器
	operationFactory OperationFactory                   // 操作工厂
	
	// 状态管理
	isRunning        int32 // 原子操作标记
	mutex            sync.RWMutex
	
	// 监控数据
	totalJobs        int64 // 总任务数
	completedJobs    int64 // 完成任务数  
	successJobs      int64 // 成功任务数
	failedJobs       int64 // 失败任务数
	
	// 配置
	maxWorkers       int           // 最大工作协程数
	jobBufferSize    int           // 任务缓冲区大小
	resultBufferSize int           // 结果缓冲区大小
}

// NewExecutionEngine 创建新的执行引擎
func NewExecutionEngine(
	adapter interfaces.ProtocolAdapter,
	metricsCollector interfaces.DefaultMetricsCollector,
	operationFactory OperationFactory,
) *ExecutionEngine {
	return &ExecutionEngine{
		adapter:          adapter,
		metricsCollector: metricsCollector,
		operationFactory: operationFactory,
		maxWorkers:       100,  // 默认最大工作协程数
		jobBufferSize:    1000, // 默认任务缓冲区大小
		resultBufferSize: 1000, // 默认结果缓冲区大小
	}
}

// SetMaxWorkers 设置最大工作协程数
func (e *ExecutionEngine) SetMaxWorkers(maxWorkers int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if maxWorkers > 0 {
		e.maxWorkers = maxWorkers
	}
}

// SetBufferSizes 设置缓冲区大小
func (e *ExecutionEngine) SetBufferSizes(jobBufferSize, resultBufferSize int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if jobBufferSize > 0 {
		e.jobBufferSize = jobBufferSize
	}
	if resultBufferSize > 0 {
		e.resultBufferSize = resultBufferSize
	}
}

// RunBenchmark 运行基准测试
func (e *ExecutionEngine) RunBenchmark(ctx context.Context, config BenchmarkConfig) (*ExecutionResult, error) {
	// 检查是否已在运行
	if !atomic.CompareAndSwapInt32(&e.isRunning, 0, 1) {
		return nil, fmt.Errorf("execution engine is already running")
	}
	defer atomic.StoreInt32(&e.isRunning, 0)

	// 重置计数器
	atomic.StoreInt64(&e.totalJobs, 0)
	atomic.StoreInt64(&e.completedJobs, 0)
	atomic.StoreInt64(&e.successJobs, 0)
	atomic.StoreInt64(&e.failedJobs, 0)

	startTime := time.Now()
	
	// 确定工作协程数
	workerCount := config.GetParallels()
	if workerCount <= 0 {
		workerCount = 1
	}
	if workerCount > e.maxWorkers {
		workerCount = e.maxWorkers
	}

	// 创建通道
	jobChan := make(chan Job, e.jobBufferSize)
	resultChan := make(chan *interfaces.OperationResult, e.resultBufferSize)
	
	// 创建工作协程组
	var workerWG sync.WaitGroup
	
	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		workerWG.Add(1)
		go e.worker(ctx, &workerWG, jobChan, resultChan)
	}
	
	// 启动结果收集协程
	var resultWG sync.WaitGroup
	resultWG.Add(1)
	go e.resultCollector(&resultWG, resultChan)
	
	// 创建任务生成上下文（支持超时和持续时间）
	jobCtx := ctx
	if duration := config.GetDuration(); duration > 0 {
		var cancel context.CancelFunc
		jobCtx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}
	
	// 渐进加载
	if rampUp := config.GetRampUp(); rampUp > 0 {
		e.generateJobsWithRampUp(jobCtx, config, jobChan)
	} else {
		e.generateJobs(jobCtx, config, jobChan)
	}
	
	// 关闭任务通道
	close(jobChan)
	
	// 等待所有工作协程完成
	workerWG.Wait()
	
	// 关闭结果通道
	close(resultChan)
	
	// 等待结果收集完成
	resultWG.Wait()
	
	endTime := time.Now()
	
	// 构建执行结果
	result := &ExecutionResult{
		TotalJobs:     atomic.LoadInt64(&e.totalJobs),
		CompletedJobs: atomic.LoadInt64(&e.completedJobs),
		SuccessJobs:   atomic.LoadInt64(&e.successJobs),
		FailedJobs:    atomic.LoadInt64(&e.failedJobs),
		TotalDuration: endTime.Sub(startTime),
		StartTime:     startTime,
		EndTime:       endTime,
	}
	
	return result, nil
}

// worker 工作协程
func (e *ExecutionEngine) worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan Job, resultChan chan<- *interfaces.OperationResult) {
	defer wg.Done()
	
	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return // 任务通道已关闭
			}
			
			// 执行任务
			result := e.executeJob(job)
			
			// 发送结果
			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
			
			// 更新完成计数
			atomic.AddInt64(&e.completedJobs, 1)
			if result.Success {
				atomic.AddInt64(&e.successJobs, 1)
			} else {
				atomic.AddInt64(&e.failedJobs, 1)
			}
			
		case <-ctx.Done():
			return
		}
	}
}

// executeJob 执行单个任务
func (e *ExecutionEngine) executeJob(job Job) *interfaces.OperationResult {
	// 使用适配器执行操作
	result, err := e.adapter.Execute(job.Context, job.Operation)
	
	if err != nil {
		// 如果适配器返回错误，创建失败结果
		return &interfaces.OperationResult{
			Success:  false,
			Duration: 0,
			Error:    err,
			IsRead:   false, // 默认为写操作，具体可以从operation中获取
		}
	}
	
	if result == nil {
		// 如果结果为空，创建默认失败结果
		return &interfaces.OperationResult{
			Success:  false,
			Duration: 0,
			Error:    fmt.Errorf("adapter returned nil result"),
			IsRead:   false,
		}
	}
	
	return result
}

// resultCollector 结果收集协程
func (e *ExecutionEngine) resultCollector(wg *sync.WaitGroup, resultChan <-chan *interfaces.OperationResult) {
	defer wg.Done()
	
	for result := range resultChan {
		// 记录到指标收集器
		if e.metricsCollector != nil {
			e.metricsCollector.Record(result)
		}
	}
}

// generateJobs 生成任务（常规模式）
func (e *ExecutionEngine) generateJobs(ctx context.Context, config BenchmarkConfig, jobChan chan<- Job) {
	total := config.GetTotal()
	atomic.StoreInt64(&e.totalJobs, int64(total))
	
	for i := 0; i < total; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			// 创建操作
			operation := e.operationFactory.CreateOperation(i, config)
			
			// 创建任务
			job := Job{
				ID:        i,
				Operation: operation,
				Context:   ctx,
			}
			
			// 发送任务
			select {
			case jobChan <- job:
			case <-ctx.Done():
				return
			}
		}
	}
}

// generateJobsWithRampUp 生成任务（渐进加载模式）
func (e *ExecutionEngine) generateJobsWithRampUp(ctx context.Context, config BenchmarkConfig, jobChan chan<- Job) {
	total := config.GetTotal()
	rampUp := config.GetRampUp()
	atomic.StoreInt64(&e.totalJobs, int64(total))
	
	// 计算渐进间隔
	interval := rampUp / time.Duration(total)
	if interval < time.Microsecond {
		interval = time.Microsecond // 最小间隔
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for i := 0; i < total; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 创建操作
			operation := e.operationFactory.CreateOperation(i, config)
			
			// 创建任务
			job := Job{
				ID:        i,
				Operation: operation,
				Context:   ctx,
			}
			
			// 发送任务
			select {
			case jobChan <- job:
			case <-ctx.Done():
				return
			}
		}
	}
}

// IsRunning 检查是否正在运行
func (e *ExecutionEngine) IsRunning() bool {
	return atomic.LoadInt32(&e.isRunning) == 1
}

// GetStats 获取当前统计信息
func (e *ExecutionEngine) GetStats() (total, completed, success, failed int64) {
	return atomic.LoadInt64(&e.totalJobs),
		atomic.LoadInt64(&e.completedJobs),
		atomic.LoadInt64(&e.successJobs),
		atomic.LoadInt64(&e.failedJobs)
}