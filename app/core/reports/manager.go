package reports

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"redis-runner/app/core/interfaces"
)

// ReportFormat 报告格式枚举
type ReportFormat string

const (
	FormatConsole ReportFormat = "console"
	FormatJSON    ReportFormat = "json"
	FormatCSV     ReportFormat = "csv"
	FormatText    ReportFormat = "text"
	FormatAll     ReportFormat = "all"
)

// ReportConfig 统一报告配置
type ReportConfig struct {
	Formats                []ReportFormat `json:"formats"`                  // 报告格式列表
	OutputDirectory        string         `json:"output_directory"`         // 输出目录
	FilePrefix             string         `json:"file_prefix"`              // 文件前缀
	EnableConsoleReport    bool           `json:"enable_console_report"`    // 启用控制台详细报告
	EnableProtocolMetrics  bool           `json:"enable_protocol_metrics"`  // 启用协议特定指标
	IncludeTimestamp       bool           `json:"include_timestamp"`        // 文件名包含时间戳
	OverwriteExisting      bool           `json:"overwrite_existing"`       // 覆盖已存在文件
}

// DefaultReportConfig 默认报告配置
func DefaultReportConfig() *ReportConfig {
	return &ReportConfig{
		Formats:                []ReportFormat{FormatConsole, FormatJSON},
		OutputDirectory:        "./reports",
		FilePrefix:             "benchmark",
		EnableConsoleReport:    true,
		EnableProtocolMetrics:  true,
		IncludeTimestamp:       true,
		OverwriteExisting:      false,
	}
}

// ReportManager 统一报告管理器
type ReportManager struct {
	config           *ReportConfig
	protocol         string
	metricsCollector interfaces.MetricsCollector
	protocolMetrics  map[string]interface{}
}

// NewReportManager 创建报告管理器
func NewReportManager(protocol string, collector interfaces.MetricsCollector, config *ReportConfig) *ReportManager {
	if config == nil {
		config = DefaultReportConfig()
	}
	
	return &ReportManager{
		config:           config,
		protocol:         protocol,
		metricsCollector: collector,
		protocolMetrics:  make(map[string]interface{}),
	}
}

// SetProtocolMetrics 设置协议特定指标
func (rm *ReportManager) SetProtocolMetrics(metrics map[string]interface{}) {
	rm.protocolMetrics = metrics
}

// GenerateReports 生成所有格式的报告
func (rm *ReportManager) GenerateReports() error {
	// 确保输出目录存在
	if err := rm.ensureOutputDirectory(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var errors []string

	// 生成每种格式的报告
	for _, format := range rm.config.Formats {
		switch format {
		case FormatConsole:
			if rm.config.EnableConsoleReport {
				rm.generateConsoleReport()
			}
		case FormatJSON:
			if err := rm.generateJSONReport(); err != nil {
				errors = append(errors, fmt.Sprintf("JSON report: %v", err))
			}
		case FormatCSV:
			if err := rm.generateCSVReport(); err != nil {
				errors = append(errors, fmt.Sprintf("CSV report: %v", err))
			}
		case FormatText:
			if err := rm.generateTextReport(); err != nil {
				errors = append(errors, fmt.Sprintf("Text report: %v", err))
			}
		case FormatAll:
			// 生成所有格式
			rm.config.Formats = []ReportFormat{FormatConsole, FormatJSON, FormatCSV, FormatText}
			return rm.GenerateReports()
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("report generation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// ensureOutputDirectory 确保输出目录存在
func (rm *ReportManager) ensureOutputDirectory() error {
	if rm.config.OutputDirectory == "" {
		return nil
	}
	return os.MkdirAll(rm.config.OutputDirectory, 0755)
}

// generateFilename 生成文件名
func (rm *ReportManager) generateFilename(extension string) string {
	var filename string
	
	if rm.config.FilePrefix != "" {
		filename = rm.config.FilePrefix
	} else {
		filename = fmt.Sprintf("%s_benchmark", rm.protocol)
	}
	
	if rm.config.IncludeTimestamp {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("%s_%s", filename, timestamp)
	}
	
	filename = fmt.Sprintf("%s.%s", filename, extension)
	
	if rm.config.OutputDirectory != "" {
		filename = filepath.Join(rm.config.OutputDirectory, filename)
	}
	
	return filename
}

// generateConsoleReport 生成控制台报告
func (rm *ReportManager) generateConsoleReport() {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("DETAILED %s PERFORMANCE REPORT\n", strings.ToUpper(rm.protocol))
	fmt.Println(strings.Repeat("-", 60))

	// 基础指标
	if rm.metricsCollector != nil {
		if exportedMetrics := rm.metricsCollector.Export(); exportedMetrics != nil {
			fmt.Println("\n=== Base Performance Metrics ===")
			rm.displayBaseMetrics(exportedMetrics)
		}
	}

	// 协议特定指标
	if rm.config.EnableProtocolMetrics && len(rm.protocolMetrics) > 0 {
		fmt.Printf("\n=== %s Specific Metrics ===\n", strings.ToUpper(rm.protocol))
		rm.displayProtocolMetrics(rm.protocolMetrics)
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Console Report Generation Completed")
	fmt.Println(strings.Repeat("-", 60))
}

// displayBaseMetrics 显示基础指标
func (rm *ReportManager) displayBaseMetrics(metrics map[string]interface{}) {
	// 显示主要性能指标
	if rps, exists := metrics["rps"]; exists {
		fmt.Printf("Requests per Second: %v\n", rps)
	}
	if totalOps, exists := metrics["total_ops"]; exists {
		fmt.Printf("Total Operations: %v\n", totalOps)
	}
	if successOps, exists := metrics["success_ops"]; exists {
		fmt.Printf("Successful Operations: %v\n", successOps)
	}
	if failedOps, exists := metrics["failed_ops"]; exists {
		fmt.Printf("Failed Operations: %v\n", failedOps)
	}
	if errorRate, exists := metrics["error_rate"]; exists {
		fmt.Printf("Error Rate: %.2f%%\n", errorRate)
	}

	// 延迟指标
	fmt.Println("\nLatency Metrics:")
	if avgLatency, exists := metrics["avg_latency"]; exists {
		if latency, ok := avgLatency.(int64); ok {
			fmt.Printf("  Average Latency: %.3f ms\n", float64(latency)/float64(time.Millisecond))
		}
	}
	if p95Latency, exists := metrics["p95_latency"]; exists {
		if latency, ok := p95Latency.(int64); ok {
			fmt.Printf("  P95 Latency: %.3f ms\n", float64(latency)/float64(time.Millisecond))
		}
	}
	if p99Latency, exists := metrics["p99_latency"]; exists {
		if latency, ok := p99Latency.(int64); ok {
			fmt.Printf("  P99 Latency: %.3f ms\n", float64(latency)/float64(time.Millisecond))
		}
	}
	if duration, exists := metrics["duration"]; exists {
		if d, ok := duration.(int64); ok {
			fmt.Printf("  Total Duration: %.3f seconds\n", float64(d)/float64(time.Second))
		}
	}
}

// displayProtocolMetrics 显示协议特定指标
func (rm *ReportManager) displayProtocolMetrics(metrics map[string]interface{}) {
	for key, value := range metrics {
		if subMetrics, ok := value.(map[string]interface{}); ok {
			fmt.Printf("\n%s:\n", strings.Title(key))
			for subKey, subValue := range subMetrics {
				fmt.Printf("  %s: %v\n", subKey, subValue)
			}
		} else {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

// generateJSONReport 生成JSON报告
func (rm *ReportManager) generateJSONReport() error {
	filename := rm.generateFilename("json")
	
	// 构建报告数据
	reportData := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"protocol":         rm.protocol,
		"base_metrics":     nil,
		"protocol_metrics": rm.protocolMetrics,
		"metadata": map[string]interface{}{
			"generated_by": "redis-runner",
			"version":      "3.0.0",
			"format":       "json",
		},
	}

	// 添加基础指标
	if rm.metricsCollector != nil {
		reportData["base_metrics"] = rm.metricsCollector.Export()
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report to %s: %w", filename, err)
	}

	fmt.Printf("%s JSON report saved to: %s\n", strings.ToUpper(rm.protocol), filename)
	return nil
}

// generateCSVReport 生成CSV报告
func (rm *ReportManager) generateCSVReport() error {
	filename := rm.generateFilename("csv")
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV report file %s: %w", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	header := []string{
		"timestamp", "protocol", "total_ops", "success_ops", "failed_ops", 
		"rps", "avg_latency_ms", "p95_latency_ms", "p99_latency_ms", "error_rate",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// 获取指标数据
	var metrics map[string]interface{}
	if rm.metricsCollector != nil {
		metrics = rm.metricsCollector.Export()
	}

	// 写入数据行
	record := []string{
		time.Now().Format(time.RFC3339),
		rm.protocol,
		rm.getMetricStringValue(metrics, "total_ops"),
		rm.getMetricStringValue(metrics, "success_ops"),
		rm.getMetricStringValue(metrics, "failed_ops"),
		rm.getMetricStringValue(metrics, "rps"),
		rm.getLatencyStringValue(metrics, "avg_latency"),
		rm.getLatencyStringValue(metrics, "p95_latency"),
		rm.getLatencyStringValue(metrics, "p99_latency"),
		rm.getMetricStringValue(metrics, "error_rate"),
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}

	fmt.Printf("%s CSV report saved to: %s\n", strings.ToUpper(rm.protocol), filename)
	return nil
}

// generateTextReport 生成文本报告
func (rm *ReportManager) generateTextReport() error {
	filename := rm.generateFilename("txt")
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create text report file %s: %w", filename, err)
	}
	defer file.Close()

	// 写入报告内容
	fmt.Fprintf(file, "%s Performance Test Report\n", strings.ToUpper(rm.protocol))
	fmt.Fprintf(file, "Generated at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 60))

	// 基础指标
	if rm.metricsCollector != nil {
		if metrics := rm.metricsCollector.Export(); metrics != nil {
			fmt.Fprintf(file, "Base Performance Metrics:\n")
			for key, value := range metrics {
				fmt.Fprintf(file, "  %s: %v\n", key, value)
			}
			fmt.Fprintf(file, "\n")
		}
	}

	// 协议特定指标
	if len(rm.protocolMetrics) > 0 {
		fmt.Fprintf(file, "%s Specific Metrics:\n", strings.ToUpper(rm.protocol))
		for key, value := range rm.protocolMetrics {
			fmt.Fprintf(file, "  %s: %v\n", key, value)
		}
	}

	fmt.Printf("%s Text report saved to: %s\n", strings.ToUpper(rm.protocol), filename)
	return nil
}

// getMetricStringValue 安全获取指标字符串值
func (rm *ReportManager) getMetricStringValue(metrics map[string]interface{}, key string) string {
	if metrics == nil {
		return "0"
	}
	if value, exists := metrics[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return "0"
}

// getLatencyStringValue 获取延迟值字符串（毫秒）
func (rm *ReportManager) getLatencyStringValue(metrics map[string]interface{}, key string) string {
	if metrics == nil {
		return "0.000"
	}
	if value, exists := metrics[key]; exists {
		if latency, ok := value.(int64); ok {
			return fmt.Sprintf("%.3f", float64(latency)/float64(time.Millisecond))
		}
	}
	return "0.000"
}

// ParseReportFormats 解析报告格式字符串
func ParseReportFormats(formatStr string) []ReportFormat {
	if formatStr == "" {
		return []ReportFormat{FormatConsole, FormatJSON}
	}

	var formats []ReportFormat
	parts := strings.Split(formatStr, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		switch part {
		case "console":
			formats = append(formats, FormatConsole)
		case "json":
			formats = append(formats, FormatJSON)
		case "csv":
			formats = append(formats, FormatCSV)
		case "text":
			formats = append(formats, FormatText)
		case "all":
			return []ReportFormat{FormatConsole, FormatJSON, FormatCSV, FormatText}
		}
	}

	if len(formats) == 0 {
		return []ReportFormat{FormatConsole, FormatJSON}
	}

	return formats
}