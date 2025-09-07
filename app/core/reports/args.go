package reports

import (
	"fmt"
	"strings"
)

// ReportArgs 报告相关命令行参数
type ReportArgs struct {
	ReportFormat          string // --report-format
	ReportDir             string // --report-dir
	ReportPrefix          string // --report-prefix
	NoConsoleReport       bool   // --no-console-report
	EnableProtocolMetrics bool   // --enable-protocol-metrics
	DisableReports        bool   // --disable-reports
}

// DefaultReportArgs 默认报告参数
func DefaultReportArgs() *ReportArgs {
	return &ReportArgs{
		ReportFormat:          "console,json",
		ReportDir:             "./reports",
		ReportPrefix:          "",
		NoConsoleReport:       false,
		EnableProtocolMetrics: true,
		DisableReports:        false,
	}
}

// ParseReportArgs 从命令行参数解析报告配置
func ParseReportArgs(args []string) (*ReportArgs, error) {
	reportArgs := DefaultReportArgs()
	
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--report-format":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--report-format requires a value")
			}
			reportArgs.ReportFormat = args[i+1]
			if err := validateReportFormat(reportArgs.ReportFormat); err != nil {
				return nil, fmt.Errorf("invalid report format: %w", err)
			}
			i++
			
		case "--report-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--report-dir requires a directory path")
			}
			reportArgs.ReportDir = args[i+1]
			i++
			
		case "--report-prefix":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--report-prefix requires a prefix string")
			}
			reportArgs.ReportPrefix = args[i+1]
			i++
			
		case "--no-console-report":
			reportArgs.NoConsoleReport = true
			
		case "--enable-protocol-metrics":
			reportArgs.EnableProtocolMetrics = true
			
		case "--disable-protocol-metrics":
			reportArgs.EnableProtocolMetrics = false
			
		case "--disable-reports":
			reportArgs.DisableReports = true
		}
	}
	
	return reportArgs, nil
}

// ToReportConfig 转换为ReportConfig
func (args *ReportArgs) ToReportConfig(protocol string) *ReportConfig {
	if args.DisableReports {
		return &ReportConfig{
			Formats:               []ReportFormat{},
			EnableConsoleReport:   false,
			EnableProtocolMetrics: false,
		}
	}

	config := DefaultReportConfig()
	
	// 设置输出目录
	if args.ReportDir != "" {
		config.OutputDirectory = args.ReportDir
	}
	
	// 设置文件前缀
	if args.ReportPrefix != "" {
		config.FilePrefix = args.ReportPrefix
	} else {
		config.FilePrefix = fmt.Sprintf("%s_benchmark", protocol)
	}
	
	// 设置报告格式
	config.Formats = ParseReportFormats(args.ReportFormat)
	
	// 设置控制台报告
	config.EnableConsoleReport = !args.NoConsoleReport
	
	// 设置协议指标
	config.EnableProtocolMetrics = args.EnableProtocolMetrics
	
	return config
}

// validateReportFormat 验证报告格式
func validateReportFormat(formatStr string) error {
	if formatStr == "" {
		return nil
	}
	
	parts := strings.Split(formatStr, ",")
	validFormats := map[string]bool{
		"console": true,
		"json":    true,
		"csv":     true,
		"text":    true,
		"all":     true,
	}
	
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if !validFormats[part] {
			return fmt.Errorf("invalid format '%s', valid formats: console, json, csv, text, all", part)
		}
	}
	
	return nil
}

// AddReportArgsToHelp 添加报告参数到帮助信息
func AddReportArgsToHelp(helpText string) string {
	reportHelp := `

Report Options:
  --report-format <formats>    Report formats: console,json,csv,text,all (default: console,json)
  --report-dir <directory>     Report output directory (default: ./reports)
  --report-prefix <prefix>     Report file prefix (default: <protocol>_benchmark)
  --no-console-report          Disable detailed console report
  --enable-protocol-metrics    Enable protocol-specific metrics (default: true)
  --disable-protocol-metrics   Disable protocol-specific metrics
  --disable-reports            Disable all report generation

Report Examples:
  # Generate JSON and CSV reports
  --report-format json,csv --report-dir ./test-reports

  # Generate all formats with custom prefix
  --report-format all --report-prefix my_test

  # Console only (no files)
  --report-format console --disable-protocol-metrics

  # Minimal reporting
  --no-console-report --report-format json`

	return helpText + reportHelp
}

// RemoveReportArgs 从命令行参数中移除报告参数
func RemoveReportArgs(args []string) []string {
	var filteredArgs []string
	reportFlags := map[string]bool{
		"--report-format":            true,
		"--report-dir":               true,
		"--report-prefix":            true,
		"--no-console-report":        true,
		"--enable-protocol-metrics":  true,
		"--disable-protocol-metrics": true,
		"--disable-reports":          true,
	}
	
	for i := 0; i < len(args); i++ {
		if reportFlags[args[i]] {
			// 跳过有值的参数
			if args[i] == "--report-format" || args[i] == "--report-dir" || args[i] == "--report-prefix" {
				i++ // 跳过下一个参数（值）
			}
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}
	
	return filteredArgs
}

// GetReportArguments 获取报告参数的命令行说明
func GetReportArguments() []string {
	return []string{
		"--report-format <formats>   Report output formats (console,json,csv,text,all)",
		"--report-dir <directory>    Output directory for report files",
		"--report-prefix <prefix>    Custom prefix for report filenames",
		"--no-console-report         Disable detailed console output",
		"--enable-protocol-metrics   Include protocol-specific metrics",
		"--disable-protocol-metrics  Exclude protocol-specific metrics",
		"--disable-reports           Completely disable report generation",
	}
}