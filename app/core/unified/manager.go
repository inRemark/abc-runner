package unified

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"redis-runner/app/core/interfaces"
)

// unifiedCommandManager 统一命令管理器实现
type unifiedCommandManager struct {
	dispatcher    CommandDispatcher
	factory       ProtocolAdapterFactory
	aliasManager  AliasManager
	errorHandler  SmartErrorHandler
	analyzer      CommandAnalyzer
	protocols     map[string]interfaces.ProtocolAdapter
	defaultProtocol string
	mutex         sync.RWMutex
}

// NewUnifiedCommandManager 创建统一命令管理器
func NewUnifiedCommandManager() UnifiedCommandManager {
	manager := &unifiedCommandManager{
		protocols: make(map[string]interfaces.ProtocolAdapter),
		mutex:     sync.RWMutex{},
	}
	
	// 初始化子组件
	manager.factory = NewProtocolAdapterFactory()
	manager.aliasManager = NewAliasManager()
	manager.errorHandler = NewSmartErrorHandler()
	manager.analyzer = NewCommandAnalyzer()
	manager.dispatcher = NewCommandDispatcher(manager.factory, manager.aliasManager, manager.errorHandler, manager.analyzer)
	
	// 设置默认别名
	manager.setupDefaultAliases()
	
	return manager
}

// RegisterProtocol 注册协议适配器
func (m *unifiedCommandManager) RegisterProtocol(name string, adapter interfaces.ProtocolAdapter) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.protocols[name]; exists {
		return fmt.Errorf("protocol '%s' already registered", name)
	}
	
	m.protocols[name] = adapter
	log.Printf("Protocol registered: %s", name)
	return nil
}

// ExecuteCommand 执行命令
func (m *unifiedCommandManager) ExecuteCommand(ctx context.Context, command string, args []string) (*CommandResult, error) {
	// 解析命令和协议
	resolvedCommand, _ := m.aliasManager.ResolveAlias(command)
	if resolvedCommand != command {
		log.Printf("Command alias resolved: %s -> %s", command, resolvedCommand)
		command = resolvedCommand
	}
	
	// 构建命令请求
	request := &CommandRequest{
		Command:   command,
		Args:      args,
		RequestID: generateRequestID(),
		Timestamp: time.Now(),
	}
	
	// 分析命令以确定协议
	analysis := m.analyzer.AnalyzeCommand(command)
	if !analysis.IsKnown {
		suggestions := m.analyzer.FindSimilarCommands(command, 0.6)
		return &CommandResult{
			Success:     false,
			Command:     command,
			Error:       fmt.Errorf("unknown command: %s", command),
			Suggestions: suggestions,
			Timestamp:   time.Now(),
		}, nil
	}
	
	// 确定协议
	protocol := m.determineProtocol(command, analysis)
	request.Protocol = protocol
	
	// 获取协议适配器
	_, exists := m.protocols[protocol]
	if !exists {
		return &CommandResult{
			Success:   false,
			Command:   command,
			Protocol:  protocol,
			Error:     fmt.Errorf("protocol '%s' not available", protocol),
			Timestamp: time.Now(),
		}, nil
	}
	
	// 分发命令执行
	return m.dispatcher.Dispatch(ctx, request)
}

// ListCommands 列出所有可用命令
func (m *unifiedCommandManager) ListCommands() []CommandInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var commands []CommandInfo
	
	// 收集所有协议的命令信息
	for protocolName, adapter := range m.protocols {
		protocolCommands := m.getProtocolCommands(protocolName, adapter)
		commands = append(commands, protocolCommands...)
	}
	
	// 按名称排序
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	
	return commands
}

// GetCommandHelp 获取命令帮助信息
func (m *unifiedCommandManager) GetCommandHelp(command string) (string, error) {
	// 解析别名
	resolvedCommand, _ := m.aliasManager.ResolveAlias(command)
	
	// 分析命令
	analysis := m.analyzer.AnalyzeCommand(resolvedCommand)
	if !analysis.IsKnown {
		suggestions := m.analyzer.FindSimilarCommands(resolvedCommand, 0.6)
		help := fmt.Sprintf("Unknown command: %s\n", resolvedCommand)
		if len(suggestions) > 0 {
			help += fmt.Sprintf("Did you mean one of these?\n")
			for _, suggestion := range suggestions {
				help += fmt.Sprintf("  - %s\n", suggestion)
			}
		}
		return help, fmt.Errorf("unknown command: %s", resolvedCommand)
	}
	
	// 获取命令信息
	commands := m.ListCommands()
	for _, cmd := range commands {
		if cmd.Name == resolvedCommand || contains(cmd.Aliases, resolvedCommand) {
			return m.formatCommandHelp(&cmd), nil
		}
	}
	
	return "", fmt.Errorf("help not available for command: %s", resolvedCommand)
}

// SetDefaultProtocol 设置默认协议
func (m *unifiedCommandManager) SetDefaultProtocol(protocol string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.protocols[protocol]; !exists {
		return fmt.Errorf("protocol '%s' not registered", protocol)
	}
	
	m.defaultProtocol = protocol
	log.Printf("Default protocol set to: %s", protocol)
	return nil
}

// AddAlias 添加命令别名
func (m *unifiedCommandManager) AddAlias(alias string, target string) error {
	return m.aliasManager.AddAlias(alias, target)
}

// RemoveAlias 移除命令别名
func (m *unifiedCommandManager) RemoveAlias(alias string) error {
	return m.aliasManager.RemoveAlias(alias)
}

// IsAlias 检查是否为别名
func (m *unifiedCommandManager) IsAlias(command string) bool {
	return m.aliasManager.IsAlias(command)
}

// ResolveAlias 解析别名到实际命令
func (m *unifiedCommandManager) ResolveAlias(alias string) string {
	resolved, _ := m.aliasManager.ResolveAlias(alias)
	return resolved
}

// setupDefaultAliases 设置默认别名
func (m *unifiedCommandManager) setupDefaultAliases() {
	defaultAliases := map[string]string{
		"r":     "redis",
		"h":     "http", 
		"k":     "kafka",
		"redis": "redis-enhanced",  // 自动升级到增强版
		"http":  "http-enhanced",   // 自动升级到增强版
		"kafka": "kafka-enhanced",  // 自动升级到增强版
	}
	
	for alias, target := range defaultAliases {
		if err := m.aliasManager.AddAlias(alias, target); err != nil {
			log.Printf("Failed to add default alias %s -> %s: %v", alias, target, err)
		}
	}
}

// determineProtocol 确定命令对应的协议
func (m *unifiedCommandManager) determineProtocol(command string, analysis *CommandAnalysis) string {
	// 优先使用分析结果中的协议
	if analysis.Protocol != "" {
		return analysis.Protocol
	}
	
	// 基于命令名称推断协议
	if strings.Contains(command, "redis") {
		return "redis"
	}
	if strings.Contains(command, "http") {
		return "http"
	}
	if strings.Contains(command, "kafka") {
		return "kafka"
	}
	
	// 返回默认协议
	if m.defaultProtocol != "" {
		return m.defaultProtocol
	}
	
	// 最后回退策略
	return "redis"  // redis作为默认协议
}

// getProtocolCommands 获取协议的命令信息
func (m *unifiedCommandManager) getProtocolCommands(protocolName string, adapter interfaces.ProtocolAdapter) []CommandInfo {
	// 这里应该根据实际的适配器实现来获取命令信息
	// 暂时返回基础命令信息
	commands := []CommandInfo{
		{
			Name:        fmt.Sprintf("%s-enhanced", protocolName),
			Protocol:    protocolName,
			Description: fmt.Sprintf("%s enhanced performance testing", strings.Title(protocolName)),
			Usage:       fmt.Sprintf("redis-runner %s-enhanced [options]", protocolName),
			Version:     "enhanced",
			Deprecated:  false,
			Examples: []string{
				fmt.Sprintf("redis-runner %s-enhanced --config conf/%s.yaml", protocolName, protocolName),
				fmt.Sprintf("redis-runner %s-enhanced -n 1000 -c 10", protocolName),
			},
		},
	}
	
	// 添加传统版本（已弃用）
	commands = append(commands, CommandInfo{
		Name:        protocolName,
		Protocol:    protocolName,
		Description: fmt.Sprintf("%s performance testing (DEPRECATED)", strings.Title(protocolName)),
		Usage:       fmt.Sprintf("redis-runner %s [options]", protocolName),
		Version:     "legacy",
		Deprecated:  true,
		Replacement: fmt.Sprintf("%s-enhanced", protocolName),
		Examples: []string{
			fmt.Sprintf("redis-runner %s -h localhost -p 6379", protocolName),
		},
	})
	
	return commands
}

// formatCommandHelp 格式化命令帮助信息
func (m *unifiedCommandManager) formatCommandHelp(cmd *CommandInfo) string {
	var help strings.Builder
	
	help.WriteString(fmt.Sprintf("Command: %s\n", cmd.Name))
	help.WriteString(fmt.Sprintf("Protocol: %s\n", cmd.Protocol))
	help.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description))
	
	if cmd.Deprecated {
		help.WriteString(fmt.Sprintf("⚠️  DEPRECATED: This command is deprecated.\n"))
		if cmd.Replacement != "" {
			help.WriteString(fmt.Sprintf("   Use '%s' instead.\n", cmd.Replacement))
		}
	}
	
	help.WriteString(fmt.Sprintf("\nUsage: %s\n", cmd.Usage))
	
	if len(cmd.Aliases) > 0 {
		help.WriteString(fmt.Sprintf("Aliases: %s\n", strings.Join(cmd.Aliases, ", ")))
	}
	
	if len(cmd.Examples) > 0 {
		help.WriteString("\nExamples:\n")
		for _, example := range cmd.Examples {
			help.WriteString(fmt.Sprintf("  %s\n", example))
		}
	}
	
	if len(cmd.Flags) > 0 {
		help.WriteString("\nFlags:\n")
		for _, flag := range cmd.Flags {
			flagStr := fmt.Sprintf("  --%s", flag.Name)
			if flag.ShortName != "" {
				flagStr += fmt.Sprintf(", -%s", flag.ShortName)
			}
			flagStr += fmt.Sprintf(" (%s)", flag.Type)
			if flag.Required {
				flagStr += " [REQUIRED]"
			}
			help.WriteString(fmt.Sprintf("%s\n", flagStr))
			help.WriteString(fmt.Sprintf("      %s\n", flag.Description))
			if flag.DefaultValue != nil {
				help.WriteString(fmt.Sprintf("      Default: %v\n", flag.DefaultValue))
			}
			if flag.Deprecated {
				help.WriteString(fmt.Sprintf("      ⚠️  DEPRECATED: %s\n", flag.Migration))
			}
		}
	}
	
	return help.String()
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}