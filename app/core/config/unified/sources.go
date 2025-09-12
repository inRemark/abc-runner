package unified

import (
	"fmt"
	"os"
	"strings"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// DefaultConfigSource 默认配置源
type DefaultConfigSource struct {
	configFactory func() interfaces.Config
}

// NewDefaultConfigSource 创建默认配置源
func NewDefaultConfigSource(configFactory func() interfaces.Config) *DefaultConfigSource {
	return &DefaultConfigSource{configFactory: configFactory}
}

// Load 加载配置
func (d *DefaultConfigSource) Load() (interfaces.Config, error) {
	return d.configFactory(), nil
}

// CanLoad 检查是否可以加载
func (d *DefaultConfigSource) CanLoad() bool {
	return true
}

// Priority 获取优先级
func (d *DefaultConfigSource) Priority() int {
	return 1
}

// YAMLConfigSource YAML文件配置源
type YAMLConfigSource struct {
	FilePath string
	parser   ConfigParser
}

// ConfigParser 配置解析器接口
type ConfigParser interface {
	Parse(data []byte) (interfaces.Config, error)
}

// NewYAMLConfigSource 创建YAML配置源
func NewYAMLConfigSource(filePath string, parser ConfigParser) *YAMLConfigSource {
	return &YAMLConfigSource{
		FilePath: filePath,
		parser:   parser,
	}
}

// Load 加载配置
func (y *YAMLConfigSource) Load() (interfaces.Config, error) {
	data, err := os.ReadFile(y.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", y.FilePath, err)
	}

	return y.parser.Parse(data)
}

// CanLoad 检查是否可以加载
func (y *YAMLConfigSource) CanLoad() bool {
	_, err := os.Stat(y.FilePath)
	return err == nil
}

// Priority 获取优先级
func (y *YAMLConfigSource) Priority() int {
	return 50
}

// EnvConfigSource 环境变量配置源
type EnvConfigSource struct {
	Prefix string
	mapper EnvVarMapper
	config interfaces.Config
}

// EnvVarMapper 环境变量映射器接口
type EnvVarMapper interface {
	MapEnvVarsToConfig(config interfaces.Config) error
	HasRelevantEnvVars() bool
}

// NewEnvConfigSource 创建环境变量配置源
func NewEnvConfigSource(prefix string, mapper EnvVarMapper) *EnvConfigSource {
	if prefix == "" {
		prefix = "ABC_RUNNER"
	}
	return &EnvConfigSource{
		Prefix: prefix,
		mapper: mapper,
	}
}

// Load 从环境变量加载配置
func (e *EnvConfigSource) Load() (interfaces.Config, error) {
	if e.config == nil {
		return nil, fmt.Errorf("no config provided for env source")
	}

	// 复制配置以避免修改原始配置
	config := e.config.Clone()

	// 映射环境变量到配置
	err := e.mapper.MapEnvVarsToConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to map env vars to config: %w", err)
	}

	return config, nil
}

// SetConfig 设置配置（用于链式加载）
func (e *EnvConfigSource) SetConfig(config interfaces.Config) {
	e.config = config
}

// CanLoad 检查是否可以加载
func (e *EnvConfigSource) CanLoad() bool {
	return e.mapper.HasRelevantEnvVars()
}

// Priority 获取优先级
func (e *EnvConfigSource) Priority() int {
	return 80
}

// ArgConfigSource 命令行参数配置源
type ArgConfigSource struct {
	Args   []string
	parser ArgParser
	config interfaces.Config
}

// ArgParser 命令行参数解析器接口
type ArgParser interface {
	ParseArgs(args []string, config interfaces.Config) error
}

// NewArgConfigSource 创建命令行参数配置源
func NewArgConfigSource(args []string, parser ArgParser) *ArgConfigSource {
	return &ArgConfigSource{
		Args:   args,
		parser: parser,
	}
}

// Load 从命令行参数加载配置
func (a *ArgConfigSource) Load() (interfaces.Config, error) {
	if a.config == nil {
		return nil, fmt.Errorf("no config provided for arg source")
	}

	// 复制配置以避免修改原始配置
	config := a.config.Clone()

	// 解析命令行参数到配置
	err := a.parser.ParseArgs(a.Args, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}

	return config, nil
}

// SetConfig 设置配置（用于链式加载）
func (a *ArgConfigSource) SetConfig(config interfaces.Config) {
	a.config = config
}

// CanLoad 检查是否可以加载
func (a *ArgConfigSource) CanLoad() bool {
	return len(a.Args) > 0
}

// Priority 获取优先级
func (a *ArgConfigSource) Priority() int {
	return 100
}

// ProtocolConfigFactory 协议配置工厂
type ProtocolConfigFactory struct {
	protocol string
}

// NewProtocolConfigFactory 创建协议配置工厂
func NewProtocolConfigFactory(protocol string) *ProtocolConfigFactory {
	return &ProtocolConfigFactory{protocol: protocol}
}

// CreateConfigSources 创建配置源列表
func (p *ProtocolConfigFactory) CreateConfigSources(configFile string, args []string,
	configFactory func() interfaces.Config,
	yamlParser ConfigParser,
	envMapper EnvVarMapper,
	argParser ArgParser) []ConfigSource {
	sources := make([]ConfigSource, 0)

	// 1. 默认配置源（最低优先级）
	defaultSource := NewDefaultConfigSource(configFactory)
	sources = append(sources, defaultSource)

	// 2. YAML配置文件源
	if configFile != "" {
		yamlSource := NewYAMLConfigSource(configFile, yamlParser)
		// 设置前一个源的配置作为基础
		sources = append(sources, yamlSource)
	} else {
		// 使用统一的配置文件查找机制
		foundPath := utils.FindConfigFile(p.protocol)
		if foundPath != "" {
			yamlSource := NewYAMLConfigSource(foundPath, yamlParser)
			sources = append(sources, yamlSource)
		}
	}

	// 3. 环境变量配置源
	envSource := NewEnvConfigSource(strings.ToUpper(p.protocol)+"_RUNNER", envMapper)
	sources = append(sources, envSource)

	// 4. 命令行参数配置源（最高优先级）
	if len(args) > 0 {
		argSource := NewArgConfigSource(args, argParser)
		sources = append(sources, argSource)
	}

	return sources
}
