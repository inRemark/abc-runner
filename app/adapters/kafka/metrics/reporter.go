package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MetricsReporter 指标报告器
type MetricsReporter struct {
	collector *KafkaMetricsCollector
	config    *ReporterConfig
}

// ReporterConfig 报告器配置
type ReporterConfig struct {
	OutputFormat   string        `json:"output_format"`   // json, csv, console
	OutputFile     string        `json:"output_file"`     // 输出文件路径
	ReportInterval time.Duration `json:"report_interval"` // 报告间隔
	EnableConsole  bool          `json:"enable_console"`  // 是否输出到控制台
	EnableFile     bool          `json:"enable_file"`     // 是否输出到文件
}

// NewMetricsReporter 创建指标报告器
func NewMetricsReporter(collector *KafkaMetricsCollector, config *ReporterConfig) *MetricsReporter {
	if config == nil {
		config = &ReporterConfig{
			OutputFormat:   "json",
			ReportInterval: 10 * time.Second,
			EnableConsole:  true,
			EnableFile:     false,
		}
	}
	
	return &MetricsReporter{
		collector: collector,
		config:    config,
	}
}

// GenerateReport 生成指标报告
func (r *MetricsReporter) GenerateReport() (*MetricsReport, error) {
	kafkaMetrics := r.collector.GetKafkaSpecificMetrics()
	baseMetrics := r.collector.Export()
	
	report := &MetricsReport{
		Timestamp:     time.Now(),
		KafkaMetrics:  kafkaMetrics,
		BaseMetrics:   baseMetrics,
		Summary:       r.generateSummary(kafkaMetrics, baseMetrics),
	}
	
	return report, nil
}

// PrintReport 打印报告到控制台
func (r *MetricsReporter) PrintReport(report *MetricsReport) error {
	if !r.config.EnableConsole {
		return nil
	}
	
	fmt.Printf("\n=== Kafka Performance Report ===\n")
	fmt.Printf("Timestamp: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Operations: %d\n", report.Summary.TotalOperations)
	fmt.Printf("Success Rate: %.2f%%\n", report.Summary.SuccessRate)
	fmt.Printf("Error Rate: %.2f%%\n", report.Summary.ErrorRate)
	
	return nil
}

// SaveReport 保存报告到文件
func (r *MetricsReporter) SaveReport(report *MetricsReport) error {
	if !r.config.EnableFile || r.config.OutputFile == "" {
		return nil
	}
	
	dir := filepath.Dir(r.config.OutputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	timestamp := report.Timestamp.Format("20060102_150405")
	ext := filepath.Ext(r.config.OutputFile)
	base := r.config.OutputFile[:len(r.config.OutputFile)-len(ext)]
	filename := fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}
	
	return nil
}

// generateSummary 生成摘要信息
func (r *MetricsReporter) generateSummary(kafkaMetrics, baseMetrics map[string]interface{}) *MetricsSummary {
	summary := &MetricsSummary{
		ProtocolSpecific: make(map[string]interface{}),
	}
	
	// 基础指标摘要
	if totalOps, ok := baseMetrics["total_ops"].(int64); ok {
		summary.TotalOperations = totalOps
	}
	if successOps, ok := baseMetrics["success_ops"].(int64); ok {
		summary.SuccessOperations = successOps
	}
	if errorRate, ok := baseMetrics["error_rate"].(float64); ok {
		summary.ErrorRate = errorRate
	}
	
	if summary.TotalOperations > 0 {
		summary.SuccessRate = float64(summary.SuccessOperations) / float64(summary.TotalOperations) * 100
	}
	
	return summary
}

// MetricsReport 指标报告结构
type MetricsReport struct {
	Timestamp     time.Time              `json:"timestamp"`
	KafkaMetrics  map[string]interface{} `json:"kafka_metrics"`
	BaseMetrics   map[string]interface{} `json:"base_metrics"`
	Summary       *MetricsSummary        `json:"summary"`
}

// MetricsSummary 指标摘要
type MetricsSummary struct {
	TotalOperations    int64                          `json:"total_operations"`
	SuccessOperations  int64                          `json:"success_operations"`
	FailedOperations   int64                          `json:"failed_operations"`
	SuccessRate        float64                        `json:"success_rate"`
	ErrorRate          float64                        `json:"error_rate"`
	AvgLatency         time.Duration                  `json:"avg_latency"`
	P95Latency         time.Duration                  `json:"p95_latency"`
	P99Latency         time.Duration                  `json:"p99_latency"`
	ProtocolSpecific   map[string]interface{}         `json:"protocol_specific"`
}