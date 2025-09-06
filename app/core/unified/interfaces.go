package unified

import (
	"context"
	"time"

	"redis-runner/app/core/interfaces"
)

// UnifiedCommandManager 统一命令管理器接口
type UnifiedCommandManager interface {
	// RegisterProtocol 注册协议适配器
	RegisterProtocol(name string, adapter interfaces.ProtocolAdapter) error
	
	// ExecuteCommand 执行命令
	ExecuteCommand(ctx context.Context, command string, args []string) (*CommandResult, error)
	
	// ListCommands 列出所有可用命令
	ListCommands() []CommandInfo
	
	// GetCommandHelp 获取命令帮助信息
	GetCommandHelp(command string) (string, error)
	
	// SetDefaultProtocol 设置默认协议
	SetDefaultProtocol(protocol string) error
	
	// AddAlias 添加命令别名
	AddAlias(alias string, target string) error
	
	// RemoveAlias 移除命令别名
	RemoveAlias(alias string) error
	
	// IsAlias 检查是否为别名
	IsAlias(command string) bool
	
	// ResolveAlias 解析别名到实际命令
	ResolveAlias(alias string) string
}

// CommandDispatcher 命令分发器接口
type CommandDispatcher interface {
	// Dispatch 分发命令到相应的适配器
	Dispatch(ctx context.Context, request *CommandRequest) (*CommandResult, error)
	
	// SetFallbackMode 设置回退模式 (enhanced | legacy | auto)
	SetFallbackMode(mode string) error
	
	// EnableSmartSuggestion 启用智能建议
	EnableSmartSuggestion(enabled bool)
	
	// GetSupportedCommands 获取支持的命令列表
	GetSupportedCommands() []string
}

// ProtocolAdapterFactory 协议适配器工厂接口
type ProtocolAdapterFactory interface {
	// CreateAdapter 创建协议适配器
	CreateAdapter(protocol string, config interfaces.Config) (interfaces.ProtocolAdapter, error)
	
	// RegisterAdapter 注册适配器创建函数
	RegisterAdapter(protocol string, creator AdapterCreator) error
	
	// GetSupportedProtocols 获取支持的协议列表
	GetSupportedProtocols() []string
	
	// ValidateConfig 验证协议配置
	ValidateConfig(protocol string, config interfaces.Config) error
}

// AdapterCreator 适配器创建函数类型
type AdapterCreator func(config interfaces.Config) (interfaces.ProtocolAdapter, error)

// CommandRequest 命令请求
type CommandRequest struct {
	Command    string            `json:"command"`
	Protocol   string            `json:"protocol"`
	Args       []string          `json:"args"`
	Flags      map[string]string `json:"flags"`
	Config     interfaces.Config `json:"config"`
	Context    map[string]string `json:"context"`
	RequestID  string            `json:"request_id"`
	Timestamp  time.Time         `json:"timestamp"`
}

// CommandResult 命令执行结果
type CommandResult struct {
	Success      bool                           `json:"success"`
	Protocol     string                         `json:"protocol"`
	Command      string                         `json:"command"`
	Output       string                         `json:"output"`
	Error        error                          `json:"error"`
	Metrics      *interfaces.Metrics            `json:"metrics"`
	Duration     time.Duration                  `json:"duration"`
	Metadata     map[string]interface{}         `json:"metadata"`
	RequestID    string                         `json:"request_id"`
	Timestamp    time.Time                      `json:"timestamp"`
	Suggestions  []string                       `json:"suggestions,omitempty"`
}

// CommandInfo 命令信息
type CommandInfo struct {
	Name         string   `json:"name"`
	Protocol     string   `json:"protocol"`
	Description  string   `json:"description"`
	Usage        string   `json:"usage"`
	Aliases      []string `json:"aliases"`
	Version      string   `json:"version"`
	Deprecated   bool     `json:"deprecated"`
	Replacement  string   `json:"replacement,omitempty"`
	Examples     []string `json:"examples"`
	Flags        []FlagInfo `json:"flags"`
}

// FlagInfo 标志信息
type FlagInfo struct {
	Name         string      `json:"name"`
	ShortName    string      `json:"short_name,omitempty"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Deprecated   bool        `json:"deprecated"`
	Migration    string      `json:"migration,omitempty"`
}

// Args 统一参数结构
type Args struct {
	Raw        []string          `json:"raw"`
	Parsed     map[string]interface{} `json:"parsed"`
	Positional []string          `json:"positional"`
	Flags      map[string]string `json:"flags"`
	Protocol   string            `json:"protocol"`
	Command    string            `json:"command"`
}

// Result 统一结果结构
type Result struct {
	Success     bool                           `json:"success"`
	Metrics     *interfaces.Metrics            `json:"metrics"`
	Output      string                         `json:"output"`
	Error       error                          `json:"error"`
	Duration    time.Duration                  `json:"duration"`
	Protocol    string                         `json:"protocol"`
	Metadata    map[string]interface{}         `json:"metadata"`
}

// SmartErrorHandler 智能错误处理器接口
type SmartErrorHandler interface {
	// HandleError 处理错误并提供智能建议
	HandleError(err error, context *CommandRequest) error
	
	// GetSuggestions 获取建议
	GetSuggestions(command string) []string
	
	// AddSuggestionRule 添加建议规则
	AddSuggestionRule(pattern string, suggestions []string) error
}

// CommandAnalyzer 命令分析器接口
type CommandAnalyzer interface {
	// AnalyzeCommand 分析命令
	AnalyzeCommand(input string) *CommandAnalysis
	
	// FindSimilarCommands 查找相似命令
	FindSimilarCommands(input string, threshold float64) []string
	
	// CalculateSimilarity 计算相似度
	CalculateSimilarity(input string, target string) float64
	
	// IsKnownCommand 检查是否为已知命令
	IsKnownCommand(command string) bool
}

// CommandAnalysis 命令分析结果
type CommandAnalysis struct {
	IsKnown     bool     `json:"is_known"`
	Suggestions []string `json:"suggestions"`
	BestMatch   string   `json:"best_match"`
	Confidence  float64  `json:"confidence"`
	Protocol    string   `json:"protocol,omitempty"`
	Deprecated  bool     `json:"deprecated"`
	Migration   string   `json:"migration,omitempty"`
}

// AliasManager 别名管理器接口
type AliasManager interface {
	// AddAlias 添加别名
	AddAlias(alias string, target string) error
	
	// RemoveAlias 移除别名
	RemoveAlias(alias string) error
	
	// ResolveAlias 解析别名
	ResolveAlias(alias string) (string, bool)
	
	// ListAliases 列出所有别名
	ListAliases() map[string]string
	
	// IsAlias 检查是否为别名
	IsAlias(command string) bool
	
	// LoadAliases 从配置加载别名
	LoadAliases(aliases map[string]string) error
}