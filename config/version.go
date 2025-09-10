package config

import (
	"fmt"
	"time"
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
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		Description: "Initial configuration version",
		Author:      "abc-runner",
	}
}
