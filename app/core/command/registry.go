package command

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// CommandRegistry 命令注册器
type CommandRegistry struct {
	handlers map[string]CommandHandler
	mutex    sync.RWMutex
}

// NewCommandRegistry 创建命令注册器
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		handlers: make(map[string]CommandHandler),
	}
}

// Register 注册命令处理器
func (r *CommandRegistry) Register(name string, handler CommandHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("command '%s' already registered", name)
	}

	r.handlers[name] = handler
	return nil
}

// Unregister 注销命令处理器
func (r *CommandRegistry) Unregister(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.handlers, name)
}

// Get 获取命令处理器
func (r *CommandRegistry) Get(name string) (CommandHandler, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	handler, exists := r.handlers[name]
	return handler, exists
}

// List 列出所有注册的命令
func (r *CommandRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	commands := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		commands = append(commands, name)
	}
	sort.Strings(commands)
	return commands
}

// ListEnhanced 列出增强版命令
func (r *CommandRegistry) ListEnhanced() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var commands []string
	for name, handler := range r.handlers {
		if handler.GetVersion() == Enhanced {
			commands = append(commands, name)
		}
	}
	sort.Strings(commands)
	return commands
}

// ListLegacy 列出传统版命令
func (r *CommandRegistry) ListLegacy() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var commands []string
	for name, handler := range r.handlers {
		if handler.GetVersion() == Legacy {
			commands = append(commands, name)
		}
	}
	sort.Strings(commands)
	return commands
}

// GetCommandsInfo 获取命令信息
func (r *CommandRegistry) GetCommandsInfo() map[string]CommandInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	info := make(map[string]CommandInfo)
	for name, handler := range r.handlers {
		info[name] = CommandInfo{
			Name:        handler.GetCommandName(),
			Description: handler.GetDescription(),
			Version:     handler.GetVersion(),
			Deprecated:  handler.IsDeprecated(),
			Usage:       handler.GetUsage(),
		}
	}
	return info
}

// CommandInfo 命令信息
type CommandInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     CommandVersion `json:"version"`
	Deprecated  bool           `json:"deprecated"`
	Usage       string         `json:"usage"`
}

// CommandRouter 命令路由器
type CommandRouter struct {
	registry       *CommandRegistry
	defaultHandler CommandHandler
}

// NewCommandRouter 创建命令路由器
func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		registry: NewCommandRegistry(),
	}
}

// SetDefaultHandler 设置默认处理器
func (r *CommandRouter) SetDefaultHandler(handler CommandHandler) {
	r.defaultHandler = handler
}

// RegisterCommand 注册命令
func (r *CommandRouter) RegisterCommand(name string, handler CommandHandler) error {
	return r.registry.Register(name, handler)
}

// Route 路由命令
func (r *CommandRouter) Route(ctx context.Context, command string, args []string) error {
	// 获取命令处理器
	handler, exists := r.registry.Get(command)
	if !exists {
		if r.defaultHandler != nil {
			// 使用默认处理器处理未知命令
			return r.defaultHandler.ExecuteCommand(ctx, append([]string{command}, args...))
		}
		return fmt.Errorf("unknown command: %s", command)
	}

	// 显示弃用警告
	if handler.IsDeprecated() {
		r.showDeprecationWarning(handler)
	}

	// 验证参数
	if err := handler.ValidateArgs(args); err != nil {
		return fmt.Errorf("invalid arguments for command '%s': %w", command, err)
	}

	// 执行命令
	return handler.ExecuteCommand(ctx, args)
}

// showDeprecationWarning 显示弃用警告
func (r *CommandRouter) showDeprecationWarning(handler CommandHandler) {
	fmt.Printf("⚠️  WARNING: Command '%s' is DEPRECATED.\n", handler.GetCommandName())
	
	// 尝试推荐增强版本
	enhancedName := r.findEnhancedVersion(handler.GetCommandName())
	if enhancedName != "" {
		fmt.Printf("   Please use '%s' instead.\n", enhancedName)
	}
	
	fmt.Printf("   Migration guide: https://docs.redis-runner.com/migration\n\n")
}

// findEnhancedVersion 查找增强版本
func (r *CommandRouter) findEnhancedVersion(legacyName string) string {
	// 尝试查找对应的增强版本
	enhancedName := legacyName + "-enhanced"
	if _, exists := r.registry.Get(enhancedName); exists {
		return enhancedName
	}

	// 查找相似的增强版本
	enhanced := r.registry.ListEnhanced()
	for _, name := range enhanced {
		if strings.Contains(name, legacyName) || strings.Contains(legacyName, strings.TrimSuffix(name, "-enhanced")) {
			return name
		}
	}

	return ""
}

// GetRegistry 获取注册器
func (r *CommandRouter) GetRegistry() *CommandRegistry {
	return r.registry
}

// ListCommands 列出所有命令
func (r *CommandRouter) ListCommands() []string {
	return r.registry.List()
}

// GetCommandInfo 获取命令信息
func (r *CommandRouter) GetCommandInfo(command string) (CommandInfo, bool) {
	handler, exists := r.registry.Get(command)
	if !exists {
		return CommandInfo{}, false
	}

	return CommandInfo{
		Name:        handler.GetCommandName(),
		Description: handler.GetDescription(),
		Version:     handler.GetVersion(),
		Deprecated:  handler.IsDeprecated(),
		Usage:       handler.GetUsage(),
	}, true
}

// GenerateHelp 生成帮助信息
func (r *CommandRouter) GenerateHelp() string {
	var help strings.Builder
	
	help.WriteString("Usage: redis-runner <command> [options]\n\n")
	
	// 增强版命令
	enhanced := r.registry.ListEnhanced()
	if len(enhanced) > 0 {
		help.WriteString("Enhanced Commands (Recommended):\n")
		for _, name := range enhanced {
			if handler, exists := r.registry.Get(name); exists {
				help.WriteString(fmt.Sprintf("  %-20s %s\n", name, handler.GetDescription()))
			}
		}
		help.WriteString("\n")
	}
	
	// 传统版命令
	legacy := r.registry.ListLegacy()
	if len(legacy) > 0 {
		help.WriteString("Legacy Commands (DEPRECATED):\n")
		for _, name := range legacy {
			if handler, exists := r.registry.Get(name); exists {
				help.WriteString(fmt.Sprintf("  %-20s ⚠️ DEPRECATED: %s\n", name, handler.GetDescription()))
			}
		}
		help.WriteString("\n")
	}
	
	help.WriteString("Global Options:\n")
	help.WriteString("  --help, -h       Show help information\n")
	help.WriteString("  --version, -v    Show version information\n")
	help.WriteString("  --config         Configuration file path\n\n")
	
	help.WriteString("Use \"redis-runner <command> --help\" for more information about a command.\n\n")
	help.WriteString("Migration Guide: https://docs.redis-runner.com/migration\n")
	
	return help.String()
}