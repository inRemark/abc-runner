package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// LegacyRedisConfig 老版本Redis配置结构
type LegacyRedisConfig struct {
	Host     string `yaml:"host,omitempty" json:"host,omitempty"`
	Port     int    `yaml:"port,omitempty" json:"port,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	DB       int    `yaml:"db,omitempty" json:"db,omitempty"`
	Cluster  bool   `yaml:"cluster,omitempty" json:"cluster,omitempty"`
}

// LegacyConfig 老版本配置结构
type LegacyConfig struct {
	Redis  *LegacyRedisConfig            `yaml:"redis,omitempty" json:"redis,omitempty"`
	HTTP   map[string]interface{}        `yaml:"http,omitempty" json:"http,omitempty"`
	Kafka  map[string]interface{}        `yaml:"kafka,omitempty" json:"kafka,omitempty"`
	Custom map[string]interface{}        `yaml:",inline" json:",inline"`
}

// NewRedisConfig 新版本Redis配置结构
type NewRedisConfig struct {
	Mode      string              `yaml:"mode" json:"mode"`
	Benchmark BenchmarkConfig     `yaml:"benchmark" json:"benchmark"`
	Pool      PoolConfig          `yaml:"pool" json:"pool"`
	Standalone StandaloneConfig   `yaml:"standalone" json:"standalone"`
	Sentinel  *SentinelConfig     `yaml:"sentinel,omitempty" json:"sentinel,omitempty"`
	Cluster   *ClusterConfig      `yaml:"cluster,omitempty" json:"cluster,omitempty"`
}

type BenchmarkConfig struct {
	Total       int    `yaml:"total" json:"total"`
	Parallels   int    `yaml:"parallels" json:"parallels"`
	RandomKeys  int    `yaml:"random_keys" json:"random_keys"`
	ReadPercent int    `yaml:"read_percent" json:"read_percent"`
	DataSize    int    `yaml:"data_size" json:"data_size"`
	TTL         int    `yaml:"ttl" json:"ttl"`
	Case        string `yaml:"case" json:"case"`
}

type PoolConfig struct {
	PoolSize int `yaml:"pool_size" json:"pool_size"`
	MinIdle  int `yaml:"min_idle" json:"min_idle"`
}

type StandaloneConfig struct {
	Addr     string `yaml:"addr" json:"addr"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

type SentinelConfig struct {
	MasterName string   `yaml:"master_name" json:"master_name"`
	Addrs      []string `yaml:"addrs" json:"addrs"`
	Password   string   `yaml:"password" json:"password"`
	DB         int      `yaml:"db" json:"db"`
}

type ClusterConfig struct {
	Addrs    []string `yaml:"addrs" json:"addrs"`
	Password string   `yaml:"password" json:"password"`
}

// NewConfig 新版本配置结构
type NewConfig struct {
	Redis *NewRedisConfig            `yaml:"redis,omitempty" json:"redis,omitempty"`
	HTTP  map[string]interface{}     `yaml:"http,omitempty" json:"http,omitempty"`
	Kafka map[string]interface{}     `yaml:"kafka,omitempty" json:"kafka,omitempty"`
}

// MigrationTool 配置迁移工具
type MigrationTool struct {
	inputFile     string
	outputFile    string
	protocol      string
	format        string // yaml or json
	backupEnabled bool
	verbose       bool
}

// NewMigrationTool 创建迁移工具
func NewMigrationTool() *MigrationTool {
	return &MigrationTool{
		format:        "yaml",
		backupEnabled: true,
		verbose:       false,
	}
}

// ParseFlags 解析命令行参数
func (m *MigrationTool) ParseFlags() {
	flag.StringVar(&m.inputFile, "input", "", "Input configuration file path")
	flag.StringVar(&m.outputFile, "output", "", "Output configuration file path (default: input file with .new suffix)")
	flag.StringVar(&m.protocol, "protocol", "auto", "Protocol to migrate (redis/http/kafka/auto)")
	flag.StringVar(&m.format, "format", "yaml", "Output format (yaml/json)")
	flag.BoolVar(&m.backupEnabled, "backup", true, "Create backup of original file")
	flag.BoolVar(&m.verbose, "verbose", false, "Verbose output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Configuration Migration Tool for Redis-Runner\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input old-redis.yaml -protocol redis\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input old-config.json -output new-config.yaml -format yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input legacy.yaml -protocol auto -verbose\n", os.Args[0])
	}

	flag.Parse()
}

// Migrate 执行迁移
func (m *MigrationTool) Migrate() error {
	// 验证输入参数
	if err := m.validateInputs(); err != nil {
		return err
	}

	// 读取原配置文件
	data, err := m.readConfigFile()
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析原配置
	legacyConfig, err := m.parseLegacyConfig(data)
	if err != nil {
		return fmt.Errorf("failed to parse legacy config: %w", err)
	}

	// 执行迁移转换
	newConfig, err := m.convertConfig(legacyConfig)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// 创建备份
	if m.backupEnabled {
		if err := m.createBackup(); err != nil {
			log.Printf("Warning: failed to create backup: %v", err)
		}
	}

	// 写入新配置文件
	if err := m.writeNewConfig(newConfig); err != nil {
		return fmt.Errorf("failed to write new config: %w", err)
	}

	m.printMigrationSummary()
	return nil
}

// validateInputs 验证输入参数
func (m *MigrationTool) validateInputs() error {
	if m.inputFile == "" {
		return fmt.Errorf("input file is required")
	}

	if _, err := os.Stat(m.inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", m.inputFile)
	}

	if m.outputFile == "" {
		ext := filepath.Ext(m.inputFile)
		base := strings.TrimSuffix(m.inputFile, ext)
		if m.format == "json" {
			m.outputFile = base + ".new.json"
		} else {
			m.outputFile = base + ".new.yaml"
		}
	}

	return nil
}

// readConfigFile 读取配置文件
func (m *MigrationTool) readConfigFile() ([]byte, error) {
	if m.verbose {
		log.Printf("Reading config file: %s", m.inputFile)
	}

	return ioutil.ReadFile(m.inputFile)
}

// parseLegacyConfig 解析老配置
func (m *MigrationTool) parseLegacyConfig(data []byte) (*LegacyConfig, error) {
	var config LegacyConfig

	ext := filepath.Ext(m.inputFile)
	switch strings.ToLower(ext) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// 尝试YAML解析
		if err := yaml.Unmarshal(data, &config); err != nil {
			// 如果YAML失败，尝试JSON
			if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse config (tried YAML and JSON): %w", err)
			}
		}
	}

	if m.verbose {
		log.Printf("Parsed legacy config successfully")
	}

	return &config, nil
}

// convertConfig 转换配置
func (m *MigrationTool) convertConfig(legacyConfig *LegacyConfig) (*NewConfig, error) {
	newConfig := &NewConfig{}

	// 转换Redis配置
	if legacyConfig.Redis != nil {
		redis, err := m.convertRedisConfig(legacyConfig.Redis)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Redis config: %w", err)
		}
		newConfig.Redis = redis
	}

	// HTTP和Kafka配置保持不变（已经是新格式）
	if legacyConfig.HTTP != nil {
		newConfig.HTTP = legacyConfig.HTTP
	}
	if legacyConfig.Kafka != nil {
		newConfig.Kafka = legacyConfig.Kafka
	}

	if m.verbose {
		log.Printf("Config conversion completed")
	}

	return newConfig, nil
}

// convertRedisConfig 转换Redis配置
func (m *MigrationTool) convertRedisConfig(legacy *LegacyRedisConfig) (*NewRedisConfig, error) {
	config := &NewRedisConfig{
		Mode: "standalone", // 默认为单机模式
		Benchmark: BenchmarkConfig{
			Total:       10000,
			Parallels:   50,
			RandomKeys:  50,
			ReadPercent: 50,
			DataSize:    3,
			TTL:         120,
			Case:        "set_get_random",
		},
		Pool: PoolConfig{
			PoolSize: 10,
			MinIdle:  2,
		},
	}

	// 构建地址
	addr := "127.0.0.1:6379" // 默认地址
	if legacy.Host != "" {
		if legacy.Port > 0 {
			addr = fmt.Sprintf("%s:%d", legacy.Host, legacy.Port)
		} else {
			addr = fmt.Sprintf("%s:6379", legacy.Host)
		}
	}

	if legacy.Cluster {
		// 集群模式
		config.Mode = "cluster"
		config.Cluster = &ClusterConfig{
			Addrs:    []string{addr},
			Password: legacy.Password,
		}
	} else {
		// 单机模式
		config.Standalone = StandaloneConfig{
			Addr:     addr,
			Password: legacy.Password,
			DB:       legacy.DB,
		}
	}

	return config, nil
}

// createBackup 创建备份
func (m *MigrationTool) createBackup() error {
	backupFile := m.inputFile + ".backup"
	
	if m.verbose {
		log.Printf("Creating backup: %s", backupFile)
	}

	input, err := ioutil.ReadFile(m.inputFile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(backupFile, input, 0644)
}

// writeNewConfig 写入新配置
func (m *MigrationTool) writeNewConfig(config *NewConfig) error {
	var data []byte
	var err error

	switch m.format {
	case "json":
		data, err = json.MarshalIndent(config, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported format: %s", m.format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if m.verbose {
		log.Printf("Writing new config to: %s", m.outputFile)
	}

	return ioutil.WriteFile(m.outputFile, data, 0644)
}

// printMigrationSummary 打印迁移摘要
func (m *MigrationTool) printMigrationSummary() {
	fmt.Printf("\n=== Configuration Migration Summary ===\n")
	fmt.Printf("Input File:  %s\n", m.inputFile)
	fmt.Printf("Output File: %s\n", m.outputFile)
	fmt.Printf("Format:      %s\n", m.format)
	
	if m.backupEnabled {
		fmt.Printf("Backup File: %s.backup\n", m.inputFile)
	}
	
	fmt.Printf("Status:      SUCCESS\n")
	fmt.Printf("=====================================\n\n")

	fmt.Printf("Next Steps:\n")
	fmt.Printf("1. Review the migrated configuration: %s\n", m.outputFile)
	fmt.Printf("2. Test the new configuration with: redis-runner <protocol>-enhanced --config %s\n", m.outputFile)
	fmt.Printf("3. Update your scripts to use the new configuration\n")
	fmt.Printf("4. Remove the old configuration and backup files when ready\n\n")

	fmt.Printf("Migration Guide: https://docs.redis-runner.com/migration\n")
}

func main() {
	tool := NewMigrationTool()
	tool.ParseFlags()

	if err := tool.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}