package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ConfigChange 配置变更记录
type ConfigChange struct {
	Timestamp   time.Time `json:"timestamp"`
	ConfigFile  string    `json:"config_file"`
	ChangeType  string    `json:"change_type"` // created, modified, deleted
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Version     string    `json:"version"`
}

// ConfigTracker 配置追踪器
type ConfigTracker struct {
	changes []ConfigChange
	logFile string
}

// NewConfigTracker 创建配置追踪器
func NewConfigTracker(logFile string) *ConfigTracker {
	return &ConfigTracker{
		changes: make([]ConfigChange, 0),
		logFile: logFile,
	}
}

// LogChange 记录配置变更
func (ct *ConfigTracker) LogChange(configFile, changeType, description, author, version string) {
	change := ConfigChange{
		Timestamp:   time.Now(),
		ConfigFile:  configFile,
		ChangeType:  changeType,
		Description: description,
		Author:      author,
		Version:     version,
	}

	ct.changes = append(ct.changes, change)

	// 保存到日志文件
	if err := ct.saveToFile(); err != nil {
		log.Printf("Warning: failed to save config change to log: %v", err)
	}

	fmt.Printf("Configuration change logged: %s - %s (%s)\n", configFile, changeType, description)
}

// saveToFile 保存变更记录到文件
func (ct *ConfigTracker) saveToFile() error {
	data, err := json.MarshalIndent(ct.changes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config changes: %w", err)
	}

	return ioutil.WriteFile(ct.logFile, data, 0644)
}

// LoadFromFile 从文件加载变更记录
func (ct *ConfigTracker) LoadFromFile() error {
	if _, err := os.Stat(ct.logFile); os.IsNotExist(err) {
		return nil // 文件不存在，返回空记录
	}

	data, err := ioutil.ReadFile(ct.logFile)
	if err != nil {
		return fmt.Errorf("failed to read config change log: %w", err)
	}

	return json.Unmarshal(data, &ct.changes)
}

// GetChangesByFile 获取指定文件的变更历史
func (ct *ConfigTracker) GetChangesByFile(configFile string) []ConfigChange {
	var changes []ConfigChange
	for _, change := range ct.changes {
		if change.ConfigFile == configFile {
			changes = append(changes, change)
		}
	}
	return changes
}

// GetAllChanges 获取所有变更记录
func (ct *ConfigTracker) GetAllChanges() []ConfigChange {
	return ct.changes
}

// PrintChangeHistory 打印变更历史
func (ct *ConfigTracker) PrintChangeHistory(configFile string) {
	changes := ct.GetChangesByFile(configFile)
	if len(changes) == 0 {
		fmt.Printf("No change history found for %s\n", configFile)
		return
	}

	fmt.Printf("Change history for %s:\n", configFile)
	fmt.Println("Timestamp           | Type      | Version | Description")
	fmt.Println("--------------------|-----------|---------|-------------")

	for _, change := range changes {
		fmt.Printf("%s | %-9s | %-7s | %s\n",
			change.Timestamp.Format("2006-01-02 15:04:05"),
			change.ChangeType,
			change.Version,
			change.Description)
	}
}

func main() {
	// 创建配置追踪器
	tracker := NewConfigTracker("config/config_changes.json")

	// 加载现有变更记录
	if err := tracker.LoadFromFile(); err != nil {
		log.Printf("Warning: failed to load config change history: %v", err)
	}

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: config_tracker <config_file> [change_type] [description] [author] [version]")
		fmt.Println("   or: config_tracker --history <config_file>")
		return
	}

	// 处理历史查询
	if os.Args[1] == "--history" && len(os.Args) > 2 {
		tracker.PrintChangeHistory(os.Args[2])
		return
	}

	// 记录配置变更
	configFile := os.Args[1]
	changeType := "modified"
	description := "Configuration updated"
	author := "unknown"
	version := "1.0.0"

	if len(os.Args) > 2 {
		changeType = os.Args[2]
	}
	if len(os.Args) > 3 {
		description = os.Args[3]
	}
	if len(os.Args) > 4 {
		author = os.Args[4]
	}
	if len(os.Args) > 5 {
		version = os.Args[5]
	}

	// 转换为相对路径
	if absPath, err := filepath.Abs(configFile); err == nil {
		if relPath, err := filepath.Rel(".", absPath); err == nil {
			configFile = relPath
		}
	}

	tracker.LogChange(configFile, changeType, description, author, version)
}
