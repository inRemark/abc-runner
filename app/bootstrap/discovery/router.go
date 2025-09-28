package discovery

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// CommandRouter 命令路由器
type CommandRouter struct {
	commands map[string]CommandHandler
	aliases  map[string]string
	builder  *AutoDIBuilder
}

// CommandHandler 命令处理器接口
type CommandHandler interface {
	Execute(ctx context.Context, args []string) error
	GetHelp() string
}

// NewCommandRouter 创建命令路由器
func NewCommandRouter(builder *AutoDIBuilder) *CommandRouter {
	return &CommandRouter{
		commands: make(map[string]CommandHandler),
		aliases:  make(map[string]string),
		builder:  builder,
	}
}

// AutoRegister 自动注册所有命令
func (r *CommandRouter) AutoRegister() error {
	log.Println("Auto-registering commands...")
	
	// 注册已发现的协议命令
	for protocolName := range r.builder.GetAllFactories() {
		if err := r.registerProtocolCommand(protocolName); err != nil {
			log.Printf("Warning: failed to register command for %s: %v", protocolName, err)
			continue
		}
	}
	
	log.Printf("Command auto-registration completed. Registered %d commands", len(r.commands))
	return nil
}

// registerProtocolCommand 注册协议命令
func (r *CommandRouter) registerProtocolCommand(protocol string) error {
	handlerName := protocol + "_handler"
	
	// 从builder获取命令处理器
	component, exists := r.builder.GetComponent(handlerName)
	if !exists {
		return fmt.Errorf("command handler not found: %s", handlerName)
	}
	
	handler, ok := component.(CommandHandler)
	if !ok {
		return fmt.Errorf("component is not a CommandHandler: %s", handlerName)
	}
	
	// 注册命令
	r.commands[protocol] = handler
	log.Printf("✅ Registered command: %s", protocol)
	
	// 注册常见别名
	r.registerCommonAliases(protocol)
	
	return nil
}

// registerCommonAliases 注册常见别名
func (r *CommandRouter) registerCommonAliases(protocol string) {
	var aliases []string
	
	switch strings.ToLower(protocol) {
	case "redis":
		aliases = []string{"r"}
	case "http":
		aliases = []string{"h"}
	case "kafka":
		aliases = []string{"k"}
	case "tcp":
		aliases = []string{"t"}
	case "udp":
		aliases = []string{"u"}
	case "grpc":
		aliases = []string{"g"}
	case "websocket":
		aliases = []string{"ws"}
	}
	
	for _, alias := range aliases {
		r.aliases[alias] = protocol
		log.Printf("✅ Registered alias: %s -> %s", alias, protocol)
	}
}

// Execute 执行命令
func (r *CommandRouter) Execute(ctx context.Context, command string, args []string) error {
	// 解析别名
	if target, exists := r.aliases[command]; exists {
		command = target
	}
	
	// 查找命令处理器
	handler, exists := r.commands[command]
	if !exists {
		return fmt.Errorf("unknown command: %s", command)
	}
	
	log.Printf("Executing command: %s with %d args", command, len(args))
	
	// 执行命令
	return handler.Execute(ctx, args)
}

// GetCommands 获取所有命令
func (r *CommandRouter) GetCommands() []string {
	var commands []string
	for cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// GetAliases 获取所有别名
func (r *CommandRouter) GetAliases() map[string]string {
	return r.aliases
}

// HasCommand 检查命令是否存在
func (r *CommandRouter) HasCommand(command string) bool {
	// 检查直接命令
	if _, exists := r.commands[command]; exists {
		return true
	}
	
	// 检查别名
	if target, exists := r.aliases[command]; exists {
		_, exists := r.commands[target]
		return exists
	}
	
	return false
}

// GetCommandHelp 获取命令帮助
func (r *CommandRouter) GetCommandHelp(command string) (string, error) {
	// 解析别名
	if target, exists := r.aliases[command]; exists {
		command = target
	}
	
	// 查找命令处理器
	handler, exists := r.commands[command]
	if !exists {
		return "", fmt.Errorf("unknown command: %s", command)
	}
	
	return handler.GetHelp(), nil
}

// ListCommands 列出所有可用命令
func (r *CommandRouter) ListCommands() string {
	var result strings.Builder
	
	result.WriteString("Available commands:\n")
	
	for command := range r.commands {
		result.WriteString(fmt.Sprintf("  %s", command))
		
		// 添加别名信息
		var aliases []string
		for alias, target := range r.aliases {
			if target == command {
				aliases = append(aliases, alias)
			}
		}
		
		if len(aliases) > 0 {
			result.WriteString(fmt.Sprintf(" (aliases: %s)", strings.Join(aliases, ", ")))
		}
		
		result.WriteString("\n")
	}
	
	return result.String()
}