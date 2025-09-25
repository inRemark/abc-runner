package utils

import (
	"log"
	"os"
	"path/filepath"
)

// ConfigDirName 配置目录名称常量
const ConfigDirName = ".abc-runner"

// StandardConfigPaths 标准配置文件查找路径
var StandardConfigPaths = map[string][]string{
	"redis": {
		"config/redis.yaml",
		"./config/redis.yaml",
		"../config/redis.yaml",
		filepath.Join(os.Getenv("HOME"), ConfigDirName, "redis.yaml"),
	},
	"http": {
		"config/http.yaml",
		"./config/http.yaml",
		"../config/http.yaml",
		filepath.Join(os.Getenv("HOME"), ConfigDirName, "http.yaml"),
	},
	"kafka": {
		"config/kafka.yaml",
		"./config/kafka.yaml",
		"../config/kafka.yaml",
		filepath.Join(os.Getenv("HOME"), ConfigDirName, "kafka.yaml"),
	},
}

// FindConfigFile 查找配置文件
func FindConfigFile(protocol string) string {
	// 获取相对于二进制文件的路径
	execPath, err := os.Executable()
	log.Printf("execPath Path at: %s", execPath)
	if err == nil {
		execDir := filepath.Dir(execPath)
		log.Printf("execDir Path at: %s", execPath)
		// 为协议添加相对于二进制文件的配置路径
		relativePath := filepath.Join(execDir, "config", protocol+".yaml")
		log.Printf("relativePath Path at: %s", execPath)
		if _, err := os.Stat(relativePath); err == nil {
			log.Printf("Found %s configuration file at: %s", protocol, relativePath)
			return relativePath
		}
	}

	// 使用标准路径查找
	configPaths, exists := StandardConfigPaths[protocol]
	if !exists {
		log.Printf("Unknown protocol: %s", protocol)
		return ""
	}

	log.Printf("Searching for %s configuration file...", protocol)
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			log.Printf("Found %s configuration file at: %s", protocol, path)
			return path
		} else {
			log.Printf("%s configuration file not found at: %s", protocol, path)
		}
	}

	log.Printf("No %s configuration file found", protocol)
	return ""
}

// FindCoreConfigFile 查找核心配置文件
func FindCoreConfigFile() string {
	// 相对于二进制文件的路径
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		relativePath := filepath.Join(execDir, "config", "core.yaml")
		if _, err := os.Stat(relativePath); err == nil {
			log.Printf("Found core configuration file at: %s", relativePath)
			return relativePath
		}
	}

	// 标准路径
	standardPaths := []string{
		"config/core.yaml",
		"./config/core.yaml",
		"../config/core.yaml",
		filepath.Join(os.Getenv("HOME"), ConfigDirName, "core.yaml"),
	}

	log.Println("Searching for core configuration file...")
	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			log.Printf("Found core configuration file at: %s", path)
			return path
		} else {
			log.Printf("Core configuration file not found at: %s", path)
		}
	}

	log.Println("No core configuration file found")
	return ""
}
