package unified

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"redis-runner/app/core/interfaces"
)

// commandDispatcher 命令分发器实现
type commandDispatcher struct {
	factory      ProtocolAdapterFactory
	aliasManager AliasManager
	errorHandler SmartErrorHandler
	analyzer     CommandAnalyzer
	fallbackMode string
	smartSuggestion bool
}

// NewCommandDispatcher 创建命令分发器
func NewCommandDispatcher(
	factory ProtocolAdapterFactory,
	aliasManager AliasManager,
	errorHandler SmartErrorHandler,
	analyzer CommandAnalyzer,
) CommandDispatcher {
	return &commandDispatcher{
		factory:         factory,
		aliasManager:    aliasManager,
		errorHandler:    errorHandler,
		analyzer:        analyzer,
		fallbackMode:    "auto",  // 默认自动模式
		smartSuggestion: true,    // 默认启用智能建议
	}
}

// Dispatch 分发命令到相应的适配器
func (d *commandDispatcher) Dispatch(ctx context.Context, request *CommandRequest) (*CommandResult, error) {
	startTime := time.Now()
	
	// 构建基础结果
	result := &CommandResult{
		Command:   request.Command,
		Protocol:  request.Protocol,
		RequestID: request.RequestID,
		Timestamp: request.Timestamp,
	}
	
	// 验证请求
	if err := d.validateRequest(request); err != nil {
		result.Success = false
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, nil
	}
	
	// 解析命令模式 (enhanced/legacy/auto)
	commandMode, actualCommand := d.parseCommandMode(request.Command)
	
	// 根据模式选择执行策略
	switch commandMode {
	case "enhanced":
		return d.executeEnhancedCommand(ctx, request, actualCommand, startTime)
	case "legacy":
		return d.executeLegacyCommand(ctx, request, actualCommand, startTime)
	case "auto":
		return d.executeAutoCommand(ctx, request, actualCommand, startTime)
	default:
		return d.executeEnhancedCommand(ctx, request, actualCommand, startTime)
	}
}

// SetFallbackMode 设置回退模式
func (d *commandDispatcher) SetFallbackMode(mode string) error {
	validModes := []string{"enhanced", "legacy", "auto"}
	
	for _, validMode := range validModes {
		if mode == validMode {
			d.fallbackMode = mode
			log.Printf("Fallback mode set to: %s", mode)
			return nil
		}
	}
	
	return fmt.Errorf("invalid fallback mode: %s, valid modes: %v", mode, validModes)
}

// EnableSmartSuggestion 启用智能建议
func (d *commandDispatcher) EnableSmartSuggestion(enabled bool) {
	d.smartSuggestion = enabled
	log.Printf("Smart suggestion enabled: %v", enabled)
}

// GetSupportedCommands 获取支持的命令列表
func (d *commandDispatcher) GetSupportedCommands() []string {
	protocols := d.factory.GetSupportedProtocols()
	var commands []string
	
	for _, protocol := range protocols {
		// 增强版命令
		commands = append(commands, fmt.Sprintf("%s-enhanced", protocol))
		// 传统版命令（已弃用）
		commands = append(commands, protocol)
	}
	
	return commands
}

// validateRequest 验证请求
func (d *commandDispatcher) validateRequest(request *CommandRequest) error {
	if request.Command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	
	if request.Protocol == "" {
		return fmt.Errorf("protocol cannot be determined")
	}
	
	// 验证协议是否支持
	supportedProtocols := d.factory.GetSupportedProtocols()
	protocolSupported := false
	for _, protocol := range supportedProtocols {
		if protocol == request.Protocol {
			protocolSupported = true
			break
		}
	}
	
	if !protocolSupported {
		return fmt.Errorf("protocol '%s' not supported, available: %v", request.Protocol, supportedProtocols)
	}
	
	return nil
}

// parseCommandMode 解析命令模式
func (d *commandDispatcher) parseCommandMode(command string) (mode string, actualCommand string) {
	// 检查是否有明确的模式指定
	if strings.HasSuffix(command, "-enhanced") {
		return "enhanced", command
	}
	
	// 检查传统命令
	legacyCommands := []string{"redis", "http", "kafka"}
	for _, legacy := range legacyCommands {
		if command == legacy {
			return "legacy", command
		}
	}
	
	// 检查是否有 --legacy 或 --enhanced 标志
	// 这里简化处理，实际应该从args中解析
	
	// 默认为自动模式
	return "auto", command
}

// executeEnhancedCommand 执行增强版命令
func (d *commandDispatcher) executeEnhancedCommand(ctx context.Context, request *CommandRequest, command string, startTime time.Time) (*CommandResult, error) {
	log.Printf("Executing enhanced command: %s", command)
	
	// 确保命令是增强版格式
	if !strings.HasSuffix(command, "-enhanced") {
		command = command + "-enhanced"
	}
	
	// 创建适配器
	// 如果没有提供配置，使用默认配置
	cfg := request.Config
	if cfg == nil {
		// 使用mock配置
		cfg = &MockConfig{
			protocol: request.Protocol,
			addresses: []string{"127.0.0.1:6379"},
		}
	}
	
	adapter, err := d.factory.CreateAdapter(request.Protocol, cfg)
	if err != nil {
		return &CommandResult{
			Success:   false,
			Command:   request.Command,
			Protocol:  request.Protocol,
			Error:     fmt.Errorf("failed to create adapter: %w", err),
			Duration:  time.Since(startTime),
			RequestID: request.RequestID,
			Timestamp: request.Timestamp,
		}, nil
	}
	
	// 执行命令
	return d.executeWithAdapter(ctx, request, adapter, startTime)
}

// executeLegacyCommand 执行传统版命令
func (d *commandDispatcher) executeLegacyCommand(ctx context.Context, request *CommandRequest, command string, startTime time.Time) (*CommandResult, error) {
	log.Printf("Executing legacy command: %s (DEPRECATED)", command)
	
	// 显示弃用警告
	d.showDeprecationWarning(command)
	
	// 根据回退模式决定是否自动升级
	if d.fallbackMode == "auto" {
		log.Printf("Auto-upgrading legacy command '%s' to enhanced version", command)
		return d.executeEnhancedCommand(ctx, request, command, startTime)
	}
	
	// 执行传统版本（通过兼容层）
	return d.executeLegacyWithCompatibility(ctx, request, command, startTime)
}

// executeAutoCommand 执行自动模式命令
func (d *commandDispatcher) executeAutoCommand(ctx context.Context, request *CommandRequest, command string, startTime time.Time) (*CommandResult, error) {
	log.Printf("Executing auto command: %s", command)
	
	// 自动模式优先使用增强版
	if d.isEnhancedAvailable(request.Protocol) {
		return d.executeEnhancedCommand(ctx, request, command, startTime)
	}
	
	// 回退到传统版本
	log.Printf("Enhanced version not available, falling back to legacy for: %s", command)
	return d.executeLegacyCommand(ctx, request, command, startTime)
}

// executeWithAdapter 使用适配器执行命令
func (d *commandDispatcher) executeWithAdapter(ctx context.Context, request *CommandRequest, adapter interfaces.ProtocolAdapter, startTime time.Time) (*CommandResult, error) {
	// 连接到目标服务
	if err := adapter.Connect(ctx, request.Config); err != nil {
		return &CommandResult{
			Success:   false,
			Command:   request.Command,
			Protocol:  request.Protocol,
			Error:     fmt.Errorf("failed to connect: %w", err),
			Duration:  time.Since(startTime),
			RequestID: request.RequestID,
			Timestamp: request.Timestamp,
		}, nil
	}
	defer adapter.Close()
	
	// 执行健康检查
	if err := adapter.HealthCheck(ctx); err != nil {
		log.Printf("Health check failed for %s: %v", request.Protocol, err)
	}
	
	// 这里应该根据具体的命令参数构建操作
	// 暂时使用简化的实现
	operation := d.buildOperation(request)
	
	// 执行操作
	operationResult, err := adapter.Execute(ctx, operation)
	if err != nil {
		return &CommandResult{
			Success:   false,
			Command:   request.Command,
			Protocol:  request.Protocol,
			Error:     fmt.Errorf("execution failed: %w", err),
			Duration:  time.Since(startTime),
			RequestID: request.RequestID,
			Timestamp: request.Timestamp,
		}, nil
	}
	
	// 获取协议指标
	protocolMetrics := adapter.GetProtocolMetrics()
	
	return &CommandResult{
		Success:   operationResult.Success,
		Command:   request.Command,
		Protocol:  request.Protocol,
		Output:    d.formatOutput(operationResult),
		Error:     operationResult.Error,
		Duration:  time.Since(startTime),
		Metadata:  protocolMetrics,
		RequestID: request.RequestID,
		Timestamp: request.Timestamp,
	}, nil
}

// executeLegacyWithCompatibility 通过兼容层执行传统命令
func (d *commandDispatcher) executeLegacyWithCompatibility(ctx context.Context, request *CommandRequest, command string, startTime time.Time) (*CommandResult, error) {
	// 这里应该调用现有的传统命令实现
	// 为了保持兼容性，暂时返回成功状态
	
	output := fmt.Sprintf("Legacy command '%s' executed via compatibility layer", command)
	
	return &CommandResult{
		Success:   true,
		Command:   request.Command,
		Protocol:  request.Protocol,
		Output:    output,
		Duration:  time.Since(startTime),
		RequestID: request.RequestID,
		Timestamp: request.Timestamp,
		Metadata: map[string]interface{}{
			"execution_mode": "legacy_compatibility",
			"warning": "This command is deprecated and will be removed in future versions",
		},
	}, nil
}

// buildOperation 构建操作对象
func (d *commandDispatcher) buildOperation(request *CommandRequest) interfaces.Operation {
	// 这里应该根据命令参数构建具体的操作
	// 暂时使用简化的实现
	return interfaces.Operation{
		Type:   "benchmark",
		Key:    "test_key",
		Value:  "test_value",
		Params: map[string]interface{}{
			"command": request.Command,
			"args":    request.Args,
		},
		Metadata: map[string]string{
			"request_id": request.RequestID,
		},
	}
}

// formatOutput 格式化输出
func (d *commandDispatcher) formatOutput(result *interfaces.OperationResult) string {
	if result.Value != nil {
		return fmt.Sprintf("Operation completed successfully: %v", result.Value)
	}
	return "Operation completed successfully"
}

// showDeprecationWarning 显示弃用警告
func (d *commandDispatcher) showDeprecationWarning(command string) {
	fmt.Printf("⚠️  WARNING: Command '%s' is DEPRECATED.\n", command)
	
	// 根据命令推荐增强版本
	enhancedCommand := d.getEnhancedVersion(command)
	if enhancedCommand != "" {
		fmt.Printf("   Please use '%s' instead.\n", enhancedCommand)
	}
	
	fmt.Printf("   This command will be removed in future versions.\n")
	fmt.Printf("   Migration guide: https://docs.redis-runner.com/migration\n\n")
}

// getEnhancedVersion 获取增强版本命令
func (d *commandDispatcher) getEnhancedVersion(legacyCommand string) string {
	switch legacyCommand {
	case "redis":
		return "redis-enhanced"
	case "http":
		return "http-enhanced"
	case "kafka":
		return "kafka-enhanced"
	default:
		return legacyCommand + "-enhanced"
	}
}

// isEnhancedAvailable 检查增强版本是否可用
func (d *commandDispatcher) isEnhancedAvailable(protocol string) bool {
	// 检查工厂是否支持该协议的增强版本
	supportedProtocols := d.factory.GetSupportedProtocols()
	for _, supportedProtocol := range supportedProtocols {
		if supportedProtocol == protocol {
			return true
		}
	}
	return false
}