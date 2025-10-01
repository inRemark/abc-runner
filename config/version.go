package config

import (
	"fmt"
	"runtime"
	"time"
)

// 版本常量定义 - 统一管理所有版本号
const (
	// 主应用版本
	AppVersion = "0.2.0"
	AppName    = "abc-runner"

	// 服务组件版本
	GRPCServerVersion      = "1.0.0"
	WebSocketServerVersion = "1.0.0"

	// 第三方组件默认版本
	DefaultKafkaVersion = "2.8.0"

	// 内部版本标识
	AutoDiscoveryVersion  = "3.0.0"
	ConfigTemplateVersion = "0.2.0"
	ReportVersion         = "0.2.0"
)

// ConfigVersion 配置版本信息
type ConfigVersion struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
}

// NewConfigVersion 创建新的配置版本
func NewConfigVersion(version, description, author string) *ConfigVersion {
	return &ConfigVersion{
		Version:     version,
		CreatedAt:   time.Now(),
		Description: description,
		Author:      author,
	}
}

// String 返回版本信息字符串
func (cv *ConfigVersion) String() string {
	return fmt.Sprintf("Version: %s, Created: %s, Description: %s, Author: %s",
		cv.Version, cv.CreatedAt.Format("2006-01-02 15:04:05"), cv.Description, cv.Author)
}

// GetVersionInfo 获取配置版本信息
func GetVersionInfo() *ConfigVersion {
	return &ConfigVersion{
		Version:     AppVersion,
		CreatedAt:   time.Now(),
		Description: "Initial configuration version",
		Author:      AppName,
	}
}

// GetAppVersion 获取应用版本
func GetAppVersion() string {
	return AppVersion
}

// GetFullVersionString 获取完整版本字符串
func GetFullVersionString() string {
	return fmt.Sprintf("%s v%s", AppName, AppVersion)
}

// GetBuildInfo 获取构建信息
func GetBuildInfo() map[string]string {
	return map[string]string{
		"app_name":    AppName,
		"app_version": AppVersion,
		"go_version":  runtime.Version(),
		"build_time":  time.Now().Format("2006-01-02 15:04:05"),
	}
}
