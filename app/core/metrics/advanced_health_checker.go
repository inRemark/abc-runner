package metrics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// HealthChecker 高性能健康检查器
type HealthChecker interface {
	Check(ctx context.Context, sysMetrics interfaces.SystemMetrics) *HealthCheckResult
	IsHealthy() bool
	GetLastCheck() *HealthCheckResult
	SetThresholds(thresholds HealthThresholds)
}

// AdvancedHealthChecker 增强版健康检查器
type AdvancedHealthChecker struct {
	thresholds    HealthThresholds
	lastResult    *HealthCheckResult
	checkCount    int64
	failureCount  int64
	mutex         sync.RWMutex
	alertHandlers []AlertHandler
	circuitBreaker *CircuitBreaker
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(thresholds HealthThresholds) HealthChecker {
	return &AdvancedHealthChecker{
		thresholds:     thresholds,
		alertHandlers:  make([]AlertHandler, 0),
		circuitBreaker: NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 5,
			ResetTimeout:     30 * time.Second,
			MaxRequests:      10,
		}),
	}
}

// Check 执行健康检查
func (hc *AdvancedHealthChecker) Check(ctx context.Context, sysMetrics interfaces.SystemMetrics) *HealthCheckResult {
	atomic.AddInt64(&hc.checkCount, 1)
	
	// 检查熔断器状态
	if hc.circuitBreaker.GetState() == CircuitBreakerOpen {
		return &HealthCheckResult{
			Timestamp: time.Now(),
			Overall:   HealthCritical,
			Message:   "熔断器开启，系统不健康",
			Issues: []HealthIssue{{
				Type:     "circuit_breaker",
				Severity: HealthCritical,
				Message:  "系统熔断器处于开启状态",
			}},
		}
	}
	
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Overall:   HealthGood,
		Issues:    make([]HealthIssue, 0),
	}
	
	// 检查内存使用率
	if sysMetrics.MemoryUsage.InUse > 0 {
		usagePercent := float64(sysMetrics.MemoryUsage.InUse) / float64(sysMetrics.MemoryUsage.Sys) * 100
		if usagePercent > hc.thresholds.MemoryUsage {
			issue := HealthIssue{
				Type:     "memory",
				Severity: hc.calculateSeverity(usagePercent, hc.thresholds.MemoryUsage),
				Message:  fmt.Sprintf("内存使用率过高: %.2f%%", usagePercent),
				Value:    usagePercent,
				Threshold: hc.thresholds.MemoryUsage,
			}
			result.Issues = append(result.Issues, issue)
			if issue.Severity > result.Overall {
				result.Overall = issue.Severity
			}
		}
	}
	
	// 检查协程数量
	goroutineCount := float64(sysMetrics.GoroutineCount)
	if goroutineCount > float64(hc.thresholds.GoroutineCount) {
		issue := HealthIssue{
			Type:     "goroutine",
			Severity: hc.calculateSeverity(goroutineCount, float64(hc.thresholds.GoroutineCount)),
			Message:  fmt.Sprintf("协程数量过高: %.0f", goroutineCount),
			Value:    goroutineCount,
			Threshold: float64(hc.thresholds.GoroutineCount),
		}
		result.Issues = append(result.Issues, issue)
		if issue.Severity > result.Overall {
			result.Overall = issue.Severity
		}
	}
	
	// 检查GC频率
	gcFreq := float64(sysMetrics.GCStats.NumGC)
	if gcFreq > float64(hc.thresholds.GCFrequency) {
		issue := HealthIssue{
			Type:     "gc",
			Severity: hc.calculateSeverity(gcFreq, float64(hc.thresholds.GCFrequency)),
			Message:  fmt.Sprintf("GC频率过高: %.0f", gcFreq),
			Value:    gcFreq,
			Threshold: float64(hc.thresholds.GCFrequency),
		}
		result.Issues = append(result.Issues, issue)
		if issue.Severity > result.Overall {
			result.Overall = issue.Severity
		}
	}
	
	// 检查CPU使用率
	if sysMetrics.CPUUsage.UsagePercent > hc.thresholds.CPUUsage {
		issue := HealthIssue{
			Type:     "cpu",
			Severity: hc.calculateSeverity(sysMetrics.CPUUsage.UsagePercent, hc.thresholds.CPUUsage),
			Message:  fmt.Sprintf("CPU使用率过高: %.2f%%", sysMetrics.CPUUsage.UsagePercent),
			Value:    sysMetrics.CPUUsage.UsagePercent,
			Threshold: hc.thresholds.CPUUsage,
		}
		result.Issues = append(result.Issues, issue)
		if issue.Severity > result.Overall {
			result.Overall = issue.Severity
		}
	}
	
	// 记录检查结果
	hc.mutex.Lock()
	hc.lastResult = result
	if result.Overall > HealthGood {
		atomic.AddInt64(&hc.failureCount, 1)
		hc.circuitBreaker.RecordFailure()
		// 触发告警
		hc.triggerAlerts(result)
	} else {
		hc.circuitBreaker.RecordSuccess()
	}
	hc.mutex.Unlock()
	
	return result
}

// IsHealthy 检查是否健康
func (hc *AdvancedHealthChecker) IsHealthy() bool {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.lastResult != nil && hc.lastResult.Overall <= HealthWarning
}

// GetLastCheck 获取最后一次检查结果
func (hc *AdvancedHealthChecker) GetLastCheck() *HealthCheckResult {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.lastResult
}

// SetThresholds 设置阈值
func (hc *AdvancedHealthChecker) SetThresholds(thresholds HealthThresholds) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.thresholds = thresholds
}

// calculateSeverity 计算严重程度
func (hc *AdvancedHealthChecker) calculateSeverity(value, threshold float64) HealthStatus {
	ratio := value / threshold
	if ratio >= 2.0 {
		return HealthCritical
	} else if ratio >= 1.5 {
		return HealthError
	} else if ratio >= 1.0 {
		return HealthWarning
	}
	return HealthGood
}

// triggerAlerts 触发告警
func (hc *AdvancedHealthChecker) triggerAlerts(result *HealthCheckResult) {
	for _, handler := range hc.alertHandlers {
		go func(h AlertHandler) {
			h.HandleAlert(result)
		}(handler)
	}
}

// AddAlertHandler 添加告警处理器
func (hc *AdvancedHealthChecker) AddAlertHandler(handler AlertHandler) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.alertHandlers = append(hc.alertHandlers, handler)
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Timestamp time.Time     `json:"timestamp"`
	Overall   HealthStatus  `json:"overall"`
	Message   string        `json:"message"`
	Issues    []HealthIssue `json:"issues"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

// HealthIssue 健康问题
type HealthIssue struct {
	Type      string       `json:"type"`
	Severity  HealthStatus `json:"severity"`
	Message   string       `json:"message"`
	Value     float64      `json:"value"`
	Threshold float64      `json:"threshold"`
	Timestamp time.Time    `json:"timestamp"`
}

// HealthStatus 健康状态
type HealthStatus int

const (
	HealthGood HealthStatus = iota
	HealthWarning
	HealthError
	HealthCritical
)

func (hs HealthStatus) String() string {
	switch hs {
	case HealthGood:
		return "good"
	case HealthWarning:
		return "warning"
	case HealthError:
		return "error"
	case HealthCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// AlertHandler 告警处理器接口
type AlertHandler interface {
	HandleAlert(result *HealthCheckResult)
}

// LogAlertHandler 日志告警处理器
type LogAlertHandler struct{}

func (lah *LogAlertHandler) HandleAlert(result *HealthCheckResult) {
	// 在实际项目中，这里可以使用日志库输出
	fmt.Printf("[HEALTH ALERT] %s: %s\n", result.Overall.String(), result.Message)
	for _, issue := range result.Issues {
		fmt.Printf("  - %s: %s (%.2f > %.2f)\n", issue.Type, issue.Message, issue.Value, issue.Threshold)
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config       CircuitBreakerConfig
	state        CircuitBreakerState
	failureCount int64
	lastFailure  time.Time
	mutex        sync.RWMutex
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           // 失败阈值
	ResetTimeout     time.Duration // 重置超时
	MaxRequests      int           // 半开状态下的最大请求数
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount++
	cb.lastFailure = time.Now()
	
	if cb.failureCount >= int64(cb.config.FailureThreshold) {
		cb.state = CircuitBreakerOpen
	}
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	// 检查是否可以从开启转为半开
	if cb.state == CircuitBreakerOpen && time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
		cb.state = CircuitBreakerHalfOpen
		cb.failureCount = 0
	}
	
	return cb.state
}

// HealthAggregator 健康状态聚合器
type HealthAggregator struct {
	checkers map[string]HealthChecker
	mutex    sync.RWMutex
}

// NewHealthAggregator 创建健康聚合器
func NewHealthAggregator() *HealthAggregator {
	return &HealthAggregator{
		checkers: make(map[string]HealthChecker),
	}
}

// AddChecker 添加检查器
func (ha *HealthAggregator) AddChecker(name string, checker HealthChecker) {
	ha.mutex.Lock()
	defer ha.mutex.Unlock()
	ha.checkers[name] = checker
}

// CheckAll 检查所有组件
func (ha *HealthAggregator) CheckAll(ctx context.Context, sysMetrics interfaces.SystemMetrics) map[string]*HealthCheckResult {
	ha.mutex.RLock()
	defer ha.mutex.RUnlock()
	
	results := make(map[string]*HealthCheckResult)
	for name, checker := range ha.checkers {
		results[name] = checker.Check(ctx, sysMetrics)
	}
	
	return results
}

// IsOverallHealthy 检查整体健康状态
func (ha *HealthAggregator) IsOverallHealthy() bool {
	ha.mutex.RLock()
	defer ha.mutex.RUnlock()
	
	for _, checker := range ha.checkers {
		if !checker.IsHealthy() {
			return false
		}
	}
	
	return true
}