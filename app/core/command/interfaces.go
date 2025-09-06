package command

import (
	"context"

	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
)

// CommandHandler 统一命令处理器接口
type CommandHandler interface {
	// GetCommandName 获取命令名称
	GetCommandName() string

	// GetDescription 获取命令描述
	GetDescription() string

	// ExecuteCommand 执行命令
	ExecuteCommand(ctx context.Context, args []string) error

	// GetUsage 获取使用说明
	GetUsage() string

	// ValidateArgs 验证参数
	ValidateArgs(args []string) error

	// GetVersion 获取命令版本（enhanced/legacy）
	GetVersion() CommandVersion

	// IsDeprecated 是否已弃用
	IsDeprecated() bool
}

// CommandVersion 命令版本类型
type CommandVersion string

const (
	// Enhanced 增强版本
	Enhanced CommandVersion = "enhanced"
	// Legacy 传统版本
	Legacy CommandVersion = "legacy"
)

// BaseCommandHandler 基础命令处理器
type BaseCommandHandler struct {
	commandName   string
	description   string
	version       CommandVersion
	deprecated    bool
	adapter       interfaces.ProtocolAdapter
	configManager *config.ConfigManager
}

// NewBaseCommandHandler 创建基础命令处理器
func NewBaseCommandHandler(
	commandName string,
	description string,
	version CommandVersion,
	deprecated bool,
	adapter interfaces.ProtocolAdapter,
	configManager *config.ConfigManager,
) *BaseCommandHandler {
	return &BaseCommandHandler{
		commandName:   commandName,
		description:   description,
		version:       version,
		deprecated:    deprecated,
		adapter:       adapter,
		configManager: configManager,
	}
}

// GetCommandName 获取命令名称
func (b *BaseCommandHandler) GetCommandName() string {
	return b.commandName
}

// GetDescription 获取命令描述
func (b *BaseCommandHandler) GetDescription() string {
	return b.description
}

// GetUsage 获取使用说明（默认实现）
func (b *BaseCommandHandler) GetUsage() string {
	return "Usage: redis-runner " + b.commandName + " [options]"
}

// ValidateArgs 验证参数（默认实现）
func (b *BaseCommandHandler) ValidateArgs(args []string) error {
	// 基础验证，具体命令可以重写此方法
	return nil
}

// GetVersion 获取命令版本
func (b *BaseCommandHandler) GetVersion() CommandVersion {
	return b.version
}

// IsDeprecated 是否已弃用
func (b *BaseCommandHandler) IsDeprecated() bool {
	return b.deprecated
}

// GetAdapter 获取适配器
func (b *BaseCommandHandler) GetAdapter() interfaces.ProtocolAdapter {
	return b.adapter
}

// GetConfigManager 获取配置管理器
func (b *BaseCommandHandler) GetConfigManager() *config.ConfigManager {
	return b.configManager
}