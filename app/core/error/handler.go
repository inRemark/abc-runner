package error

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"redis-runner/app/core/interfaces"
)

// ErrorHandler 错误处理器
type ErrorHandler struct {
	retryConfig      *RetryConfig
	circuitBreaker   *CircuitBreaker
	errorClassifier  *ErrorClassifier
	recoveryManager  *RecoveryManager
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	Jitter         bool
	RetryableErrors []ErrorType
}

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeConnection     ErrorType = "connection"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeTemporary      ErrorType = "temporary"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypePermanent      ErrorType = "permanent"
	ErrorTypeUnknown        ErrorType = "unknown"
)

// ErrorInfo 错误信息
type ErrorInfo struct {
	Type        ErrorType
	Message     string
	Retryable   bool
	Recoverable bool
	Severity    ErrorSeverity
	Metadata    map[string]interface{}
}

// ErrorSeverity 错误严重程度
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	state           CircuitState
	failureCount    int
	successCount    int
	failureThreshold int
	successThreshold int
	timeout         time.Duration
	lastFailTime    time.Time
	mutex           sync.RWMutex
}

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// ErrorClassifier 错误分类器
type ErrorClassifier struct {
	rules []ClassificationRule
}

// ClassificationRule 分类规则
type ClassificationRule func(error) *ErrorInfo

// RecoveryManager 恢复管理器
type RecoveryManager struct {
	strategies map[ErrorType]RecoveryStrategy
}

// RecoveryStrategy 恢复策略
type RecoveryStrategy func(context.Context, error, interfaces.ProtocolAdapter) error

// NewErrorHandler 创建错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		retryConfig:     NewDefaultRetryConfig(),
		circuitBreaker:  NewCircuitBreaker(),
		errorClassifier: NewErrorClassifier(),
		recoveryManager: NewRecoveryManager(),
	}
}

// NewDefaultRetryConfig 创建默认重试配置
func NewDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []ErrorType{
			ErrorTypeTimeout,
			ErrorTypeConnection,
			ErrorTypeNetwork,
			ErrorTypeTemporary,
			ErrorTypeRateLimit,
		},
	}
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: 5,
		successThreshold: 3,
		timeout:          30 * time.Second,
	}
}

// NewErrorClassifier 创建错误分类器
func NewErrorClassifier() *ErrorClassifier {
	classifier := &ErrorClassifier{
		rules: make([]ClassificationRule, 0),
	}
	
	// 添加默认分类规则
	classifier.AddRule(classifyTimeoutError)
	classifier.AddRule(classifyConnectionError)
	classifier.AddRule(classifyNetworkError)
	classifier.AddRule(classifyAuthenticationError)
	classifier.AddRule(classifyRateLimitError)
	
	return classifier
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager() *RecoveryManager {
	manager := &RecoveryManager{
		strategies: make(map[ErrorType]RecoveryStrategy),
	}
	
	// 注册默认恢复策略
	manager.RegisterStrategy(ErrorTypeConnection, reconnectionStrategy)
	manager.RegisterStrategy(ErrorTypeAuthentication, reauthenticationStrategy)
	manager.RegisterStrategy(ErrorTypeRateLimit, backoffStrategy)
	
	return manager
}

// HandleError 处理错误
func (h *ErrorHandler) HandleError(ctx context.Context, err error, adapter interfaces.ProtocolAdapter) *interfaces.OperationResult {
	if err == nil {
		return &interfaces.OperationResult{Success: true}
	}
	
	// 分类错误
	errorInfo := h.errorClassifier.Classify(err)
	
	// 检查熔断器状态
	if !h.circuitBreaker.CanExecute() {
		return &interfaces.OperationResult{
			Success: false,
			Error:   fmt.Errorf("circuit breaker is open: %w", err),
		}
	}
	
	// 尝试恢复
	if errorInfo.Recoverable {
		if recoveryErr := h.recoveryManager.Recover(ctx, err, adapter, errorInfo.Type); recoveryErr == nil {
			h.circuitBreaker.RecordSuccess()
			return &interfaces.OperationResult{Success: true}
		}
	}
	
	// 记录失败
	h.circuitBreaker.RecordFailure()
	
	return &interfaces.OperationResult{
		Success: false,
		Error:   err,
		Metadata: map[string]interface{}{
			"error_type":        string(errorInfo.Type),
			"error_severity":    errorInfo.Severity,
			"retryable":         errorInfo.Retryable,
			"recoverable":       errorInfo.Recoverable,
			"circuit_state":     h.circuitBreaker.GetState(),
		},
	}
}

// ExecuteWithRetry 带重试的执行
func (h *ErrorHandler) ExecuteWithRetry(ctx context.Context, operation interfaces.Operation, executor func(context.Context, interfaces.Operation) (*interfaces.OperationResult, error)) (*interfaces.OperationResult, error) {
	var lastErr error
	
	for attempt := 0; attempt <= h.retryConfig.MaxRetries; attempt++ {
		// 检查熔断器
		if !h.circuitBreaker.CanExecute() {
			return &interfaces.OperationResult{
				Success: false,
				Error:   fmt.Errorf("circuit breaker is open"),
			}, fmt.Errorf("circuit breaker is open")
		}
		
		// 执行操作
		result, err := executor(ctx, operation)
		
		if err == nil && result != nil && result.Success {
			h.circuitBreaker.RecordSuccess()
			return result, nil
		}
		
		lastErr = err
		if err == nil && result != nil {
			lastErr = result.Error
		}
		
		// 分类错误
		errorInfo := h.errorClassifier.Classify(lastErr)
		
		// 记录失败
		h.circuitBreaker.RecordFailure()
		
		// 检查是否可重试
		if !h.isRetryable(errorInfo) || attempt == h.retryConfig.MaxRetries {
			break
		}
		
		// 等待重试
		delay := h.calculateRetryDelay(attempt)
		select {
		case <-ctx.Done():
			return &interfaces.OperationResult{
				Success: false,
				Error:   ctx.Err(),
			}, ctx.Err()
		case <-time.After(delay):
			continue
		}
	}
	
	return &interfaces.OperationResult{
		Success: false,
		Error:   fmt.Errorf("operation failed after %d attempts: %w", h.retryConfig.MaxRetries, lastErr),
	}, lastErr
}

// isRetryable 检查错误是否可重试
func (h *ErrorHandler) isRetryable(errorInfo *ErrorInfo) bool {
	for _, retryableType := range h.retryConfig.RetryableErrors {
		if errorInfo.Type == retryableType {
			return true
		}
	}
	return false
}

// calculateRetryDelay 计算重试延迟
func (h *ErrorHandler) calculateRetryDelay(attempt int) time.Duration {
	delay := float64(h.retryConfig.InitialDelay) * math.Pow(h.retryConfig.BackoffFactor, float64(attempt))
	
	if delay > float64(h.retryConfig.MaxDelay) {
		delay = float64(h.retryConfig.MaxDelay)
	}
	
	// 添加抖动
	if h.retryConfig.Jitter {
		jitter := rand.Float64() * delay * 0.1 // 10%的抖动
		delay += jitter
	}
	
	return time.Duration(delay)
}

// CircuitBreaker methods

// CanExecute 检查是否可以执行
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Since(cb.lastFailTime) >= cb.timeout
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.successCount++
	cb.failureCount = 0
	
	if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
		cb.state = StateClosed
		cb.successCount = 0
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount++
	cb.successCount = 0
	cb.lastFailTime = time.Now()
	
	if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	} else if cb.state == StateHalfOpen {
		cb.state = StateOpen
	}
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// ErrorClassifier methods

// AddRule 添加分类规则
func (ec *ErrorClassifier) AddRule(rule ClassificationRule) {
	ec.rules = append(ec.rules, rule)
}

// Classify 分类错误
func (ec *ErrorClassifier) Classify(err error) *ErrorInfo {
	for _, rule := range ec.rules {
		if info := rule(err); info != nil {
			return info
		}
	}
	
	// 默认分类
	return &ErrorInfo{
		Type:        ErrorTypeUnknown,
		Message:     err.Error(),
		Retryable:   false,
		Recoverable: false,
		Severity:    SeverityMedium,
	}
}

// 分类规则实现

func classifyTimeoutError(err error) *ErrorInfo {
	if strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return &ErrorInfo{
			Type:        ErrorTypeTimeout,
			Message:     err.Error(),
			Retryable:   true,
			Recoverable: false,
			Severity:    SeverityMedium,
		}
	}
	return nil
}

func classifyConnectionError(err error) *ErrorInfo {
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "connect") {
		return &ErrorInfo{
			Type:        ErrorTypeConnection,
			Message:     err.Error(),
			Retryable:   true,
			Recoverable: true,
			Severity:    SeverityHigh,
		}
	}
	return nil
}

func classifyNetworkError(err error) *ErrorInfo {
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "broken pipe") {
		return &ErrorInfo{
			Type:        ErrorTypeNetwork,
			Message:     err.Error(),
			Retryable:   true,
			Recoverable: true,
			Severity:    SeverityHigh,
		}
	}
	return nil
}

func classifyAuthenticationError(err error) *ErrorInfo {
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "auth") || strings.Contains(errStr, "unauthorized") {
		return &ErrorInfo{
			Type:        ErrorTypeAuthentication,
			Message:     err.Error(),
			Retryable:   false,
			Recoverable: true,
			Severity:    SeverityCritical,
		}
	}
	return nil
}

func classifyRateLimitError(err error) *ErrorInfo {
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many") {
		return &ErrorInfo{
			Type:        ErrorTypeRateLimit,
			Message:     err.Error(),
			Retryable:   true,
			Recoverable: false,
			Severity:    SeverityMedium,
		}
	}
	return nil
}

// RecoveryManager methods

// RegisterStrategy 注册恢复策略
func (rm *RecoveryManager) RegisterStrategy(errorType ErrorType, strategy RecoveryStrategy) {
	rm.strategies[errorType] = strategy
}

// Recover 执行恢复
func (rm *RecoveryManager) Recover(ctx context.Context, err error, adapter interfaces.ProtocolAdapter, errorType ErrorType) error {
	strategy, exists := rm.strategies[errorType]
	if !exists {
		return fmt.Errorf("no recovery strategy for error type: %s", errorType)
	}
	
	return strategy(ctx, err, adapter)
}

// 恢复策略实现

func reconnectionStrategy(ctx context.Context, err error, adapter interfaces.ProtocolAdapter) error {
	// 尝试重新连接
	if adapter != nil {
		return adapter.HealthCheck(ctx)
	}
	return fmt.Errorf("adapter is nil")
}

func reauthenticationStrategy(ctx context.Context, err error, adapter interfaces.ProtocolAdapter) error {
	// 这里可以实现重新认证逻辑
	// 对于Redis，通常需要重新建立连接
	return reconnectionStrategy(ctx, err, adapter)
}

func backoffStrategy(ctx context.Context, err error, adapter interfaces.ProtocolAdapter) error {
	// 实现退避策略
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return nil
	}
}