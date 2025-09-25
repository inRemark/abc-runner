package reporting

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"abc-runner/app/core/metrics"
)

// ReportGenerator 新的报告生成器接口
type ReportGenerator interface {
	Generate(snapshot interface{}) (*Report, error)
	SetFormat(format ReportFormat)
	SetOutputPath(path string)
	Configure(config *ReportConfig) error
}

// ReportFormat 报告格式
type ReportFormat string

const (
	FormatJSON    ReportFormat = "json"
	FormatCSV     ReportFormat = "csv"
	FormatConsole ReportFormat = "console"
	FormatHTML    ReportFormat = "html"
	FormatMarkdown ReportFormat = "markdown"
)

// Report 报告结构
type Report struct {
	Format    ReportFormat    `json:"format"`
	Content   interface{}     `json:"content"`
	Metadata  ReportMetadata  `json:"metadata"`
	FilePath  string          `json:"file_path,omitempty"`
}

// ReportMetadata 报告元数据
type ReportMetadata struct {
	GeneratedAt time.Time `json:"generated_at"`
	Protocol    string    `json:"protocol"`
	Version     string    `json:"version"`
	Generator   string    `json:"generator"`
	Duration    time.Duration `json:"duration"`
}

// ReportConfig 报告配置
type ReportConfig struct {
	OutputDir       string        `json:"output_dir"`
	FilePrefix      string        `json:"file_prefix"`
	IncludeTimestamp bool         `json:"include_timestamp"`
	Formats         []ReportFormat `json:"formats"`
	Template        string        `json:"template,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty"`
}

// UniversalReportGenerator 通用报告生成器
type UniversalReportGenerator struct {
	config    *ReportConfig
	formatter map[ReportFormat]Formatter
}

// Formatter 格式化器接口
type Formatter interface {
	Format(data interface{}) ([]byte, error)
	FileExtension() string
	ContentType() string
}

// NewUniversalReportGenerator 创建通用报告生成器
func NewUniversalReportGenerator(config *ReportConfig) *UniversalReportGenerator {
	if config == nil {
		config = DefaultReportConfig()
	}

	generator := &UniversalReportGenerator{
		config:    config,
		formatter: make(map[ReportFormat]Formatter),
	}

	// 注册默认格式化器
	generator.formatter[FormatJSON] = &JSONFormatter{}
	generator.formatter[FormatCSV] = &CSVFormatter{}
	generator.formatter[FormatConsole] = &ConsoleFormatter{}
	generator.formatter[FormatHTML] = &HTMLFormatter{}
	generator.formatter[FormatMarkdown] = &MarkdownFormatter{}

	return generator
}

// Generate 生成报告
func (urg *UniversalReportGenerator) Generate(snapshot interface{}) (*Report, error) {
	reports := make([]*Report, 0, len(urg.config.Formats))

	for _, format := range urg.config.Formats {
		formatter, exists := urg.formatter[format]
		if !exists {
			return nil, fmt.Errorf("unsupported format: %s", format)
		}

		// 生成报告内容
		content, err := formatter.Format(snapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to format %s report: %w", format, err)
		}

		// 创建报告
		report := &Report{
			Format:  format,
			Content: content,
			Metadata: ReportMetadata{
				GeneratedAt: time.Now(),
				Protocol:    urg.extractProtocol(snapshot),
				Version:     "3.0.0",
				Generator:   "abc-runner-universal",
			},
		}

		// 保存到文件（除了console格式）
		if format != FormatConsole {
			filePath, err := urg.saveToFile(report, formatter)
			if err != nil {
				return nil, fmt.Errorf("failed to save %s report: %w", format, err)
			}
			report.FilePath = filePath
		} else {
			// Console格式直接输出
			fmt.Println(string(content))
		}

		reports = append(reports, report)
	}

	// 返回第一个报告作为主报告
	if len(reports) > 0 {
		return reports[0], nil
	}

	return nil, fmt.Errorf("no reports generated")
}

// SetFormat 设置单一格式
func (urg *UniversalReportGenerator) SetFormat(format ReportFormat) {
	urg.config.Formats = []ReportFormat{format}
}

// SetOutputPath 设置输出路径
func (urg *UniversalReportGenerator) SetOutputPath(path string) {
	urg.config.OutputDir = path
}

// Configure 配置报告生成器
func (urg *UniversalReportGenerator) Configure(config *ReportConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	urg.config = config
	return nil
}

// saveToFile 保存报告到文件
func (urg *UniversalReportGenerator) saveToFile(report *Report, formatter Formatter) (string, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(urg.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// 生成文件名
	fileName := urg.generateFileName(report.Format, formatter.FileExtension())
	filePath := filepath.Join(urg.config.OutputDir, fileName)

	// 写入文件
	var data []byte
	switch content := report.Content.(type) {
	case []byte:
		data = content
	case string:
		data = []byte(content)
	default:
		return "", fmt.Errorf("unsupported content type: %T", content)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// generateFileName 生成文件名
func (urg *UniversalReportGenerator) generateFileName(format ReportFormat, extension string) string {
	var parts []string
	
	if urg.config.FilePrefix != "" {
		parts = append(parts, urg.config.FilePrefix)
	} else {
		parts = append(parts, "metrics_report")
	}

	if urg.config.IncludeTimestamp {
		timestamp := time.Now().Format("20060102_150405")
		parts = append(parts, timestamp)
	}

	parts = append(parts, string(format))
	
	fileName := strings.Join(parts, "_")
	return fmt.Sprintf("%s.%s", fileName, extension)
}

// extractProtocol 从快照中提取协议类型
func (urg *UniversalReportGenerator) extractProtocol(snapshot interface{}) string {
	// 尝试从不同类型的快照中提取协议信息
	switch s := snapshot.(type) {
	case *metrics.MetricsSnapshot[interface{}]:
		// 尝试从协议数据中推断
		return fmt.Sprintf("%T", s.Protocol)
	case map[string]interface{}:
		if protocol, ok := s["protocol"].(string); ok {
			return protocol
		}
	}
	return "unknown"
}

// DefaultReportConfig 返回默认配置
func DefaultReportConfig() *ReportConfig {
	return &ReportConfig{
		OutputDir:        "./reports",
		FilePrefix:       "benchmark",
		IncludeTimestamp: true,
		Formats:          []ReportFormat{FormatJSON, FormatConsole},
		CustomFields:     make(map[string]interface{}),
	}
}

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

func (jf *JSONFormatter) Format(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (jf *JSONFormatter) FileExtension() string { return "json" }
func (jf *JSONFormatter) ContentType() string   { return "application/json" }

// CSVFormatter CSV格式化器
type CSVFormatter struct{}

func (cf *CSVFormatter) Format(data interface{}) ([]byte, error) {
	// 简化的CSV格式化实现
	// 实际项目中应该根据数据结构定制
	records := [][]string{
		{"metric", "value", "timestamp"},
		{"total_ops", "0", time.Now().Format(time.RFC3339)},
		{"success_rate", "0", time.Now().Format(time.RFC3339)},
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)
	
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	
	return []byte(buf.String()), nil
}

func (cf *CSVFormatter) FileExtension() string { return "csv" }
func (cf *CSVFormatter) ContentType() string   { return "text/csv" }

// ConsoleFormatter 控制台格式化器
type ConsoleFormatter struct{}

func (cf *ConsoleFormatter) Format(data interface{}) ([]byte, error) {
	var output strings.Builder
	
	output.WriteString("\n" + strings.Repeat("=", 60) + "\n")
	output.WriteString("PERFORMANCE METRICS REPORT\n")
	output.WriteString(strings.Repeat("=", 60) + "\n")
	output.WriteString(fmt.Sprintf("Generated at: %s\n", time.Now().Format(time.RFC3339)))
	output.WriteString("\nCore Metrics:\n")
	
	// 这里应该根据实际数据结构解析
	if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
		output.WriteString(string(jsonData))
	}
	
	output.WriteString("\n" + strings.Repeat("=", 60) + "\n")
	
	return []byte(output.String()), nil
}

func (cf *ConsoleFormatter) FileExtension() string { return "txt" }
func (cf *ConsoleFormatter) ContentType() string   { return "text/plain" }

// HTMLFormatter HTML格式化器
type HTMLFormatter struct{}

func (hf *HTMLFormatter) Format(data interface{}) ([]byte, error) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Performance Metrics Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; }
        .metrics { margin-top: 20px; }
        .metric-item { background-color: #fafafa; padding: 10px; margin: 5px 0; border-left: 4px solid #007cba; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Performance Metrics Report</h1>
        <p>Generated at: ` + time.Now().Format(time.RFC3339) + `</p>
    </div>
    <div class="metrics">
        <div class="metric-item">
            <strong>Total Operations:</strong> N/A
        </div>
        <div class="metric-item">
            <strong>Success Rate:</strong> N/A
        </div>
    </div>
</body>
</html>`
	
	return []byte(html), nil
}

func (hf *HTMLFormatter) FileExtension() string { return "html" }
func (hf *HTMLFormatter) ContentType() string   { return "text/html" }

// MarkdownFormatter Markdown格式化器
type MarkdownFormatter struct{}

func (mf *MarkdownFormatter) Format(data interface{}) ([]byte, error) {
	markdown := `# Performance Metrics Report

Generated at: ` + time.Now().Format(time.RFC3339) + `

## Core Metrics

| Metric | Value |
|--------|-------|
| Total Operations | N/A |
| Success Rate | N/A |
| Average Latency | N/A |

## Summary

This report contains performance metrics for the benchmark test.
`
	
	return []byte(markdown), nil
}

func (mf *MarkdownFormatter) FileExtension() string { return "md" }
func (mf *MarkdownFormatter) ContentType() string   { return "text/markdown" }