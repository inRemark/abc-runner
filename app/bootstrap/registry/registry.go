package registry

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"
)

// ProtocolRegistry 协议注册中心
type ProtocolRegistry struct {
	adapters   map[string]ProtocolInfo
	commands   map[string]CommandInfo
	aliases    map[string]string
}

// ProtocolInfo 协议信息
type ProtocolInfo struct {
	Name        string
	AdapterType reflect.Type
	ConfigType  reflect.Type
	MetricsType reflect.Type
	Path        string
}

// CommandInfo 命令信息
type CommandInfo struct {
	Name        string
	Handler     interface{}
	Description string
	Aliases     []string
}

// NewProtocolRegistry 创建协议注册中心
func NewProtocolRegistry() *ProtocolRegistry {
	return &ProtocolRegistry{
		adapters: make(map[string]ProtocolInfo),
		commands: make(map[string]CommandInfo),
		aliases:  make(map[string]string),
	}
}

// DiscoverProtocols 发现协议
func (r *ProtocolRegistry) DiscoverProtocols(scanPaths []string) error {
	log.Println("Starting protocol discovery...")
	
	for _, path := range scanPaths {
		log.Printf("Scanning path: %s", path)
		
		// 解析glob模式
		matches, err := filepath.Glob(path)
		if err != nil {
			log.Printf("Warning: failed to scan path %s: %v", path, err)
			continue
		}
		
		for _, match := range matches {
			if err := r.scanProtocolDirectory(match); err != nil {
				log.Printf("Warning: failed to scan directory %s: %v", match, err)
			}
		}
	}
	
	log.Printf("Protocol discovery completed. Found %d protocols", len(r.adapters))
	return nil
}

// scanProtocolDirectory 扫描协议目录
func (r *ProtocolRegistry) scanProtocolDirectory(dirPath string) error {
	// 从路径中提取协议名称
	protocolName := filepath.Base(dirPath)
	
	// 检查是否包含必要的文件
	adapterFile := filepath.Join(dirPath, "adapter.go")
	
	// 检查文件是否存在
	if !r.fileExists(adapterFile) {
		return fmt.Errorf("adapter.go not found in %s", dirPath)
	}
	
	log.Printf("Found protocol: %s", protocolName)
	
	// 注册协议信息
	r.adapters[protocolName] = ProtocolInfo{
		Name: protocolName,
		Path: dirPath,
	}
	
	// 注册常见别名
	r.registerCommonAliases(protocolName)
	
	return nil
}

// registerCommonAliases 注册常见别名
func (r *ProtocolRegistry) registerCommonAliases(protocol string) {
	switch strings.ToLower(protocol) {
	case "redis":
		r.aliases["r"] = protocol
	case "http":
		r.aliases["h"] = protocol
	case "kafka":
		r.aliases["k"] = protocol
	}
}

// fileExists 检查文件是否存在
func (r *ProtocolRegistry) fileExists(path string) bool {
	// TODO: 实现文件存在检查
	// 暂时返回true，将在实际扫描中实现
	return true
}

// GetProtocols 获取所有协议
func (r *ProtocolRegistry) GetProtocols() map[string]ProtocolInfo {
	return r.adapters
}

// GetCommands 获取所有命令
func (r *ProtocolRegistry) GetCommands() map[string]CommandInfo {
	return r.commands
}

// GetAliases 获取所有别名
func (r *ProtocolRegistry) GetAliases() map[string]string {
	return r.aliases
}

// ResolveCommand 解析命令（包括别名）
func (r *ProtocolRegistry) ResolveCommand(command string) (string, bool) {
	// 首先检查是否是别名
	if target, exists := r.aliases[command]; exists {
		return target, true
	}
	
	// 检查是否是直接命令
	if _, exists := r.commands[command]; exists {
		return command, true
	}
	
	// 检查是否是协议名称
	if _, exists := r.adapters[command]; exists {
		return command, true
	}
	
	return "", false
}

// RegisterCommand 注册命令
func (r *ProtocolRegistry) RegisterCommand(name string, handler interface{}, description string, aliases ...string) {
	r.commands[name] = CommandInfo{
		Name:        name,
		Handler:     handler,
		Description: description,
		Aliases:     aliases,
	}
	
	// 注册别名
	for _, alias := range aliases {
		r.aliases[alias] = name
	}
	
	log.Printf("Registered command: %s (aliases: %v)", name, aliases)
}