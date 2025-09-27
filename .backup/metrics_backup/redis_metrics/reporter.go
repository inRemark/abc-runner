package metrics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// ReportFormat 报告格式
type ReportFormat string

const (
	FormatJSON     ReportFormat = "json"
	FormatCSV      ReportFormat = "csv"
	FormatText     ReportFormat = "text"
	FormatConsole  ReportFormat = "console"
)

// MetricsReporter 指标报告器接口
type MetricsReporter interface {
	Report(metrics map[string]interface{}) error
	SetFormat(format ReportFormat)
	SetOutput(output string)
	GetFormat() ReportFormat
	GetOutput() string
}

// StandardMetricsReporter 标准指标报告器
type StandardMetricsReporter struct {
	format ReportFormat
	output string
	writer io.Writer
}

// NewMetricsReporter 创建指标报告器
func NewMetricsReporter(format ReportFormat, output string) *StandardMetricsReporter {
	reporter := &StandardMetricsReporter{
		format: format,
		output: output,
	}

	// 设置默认输出
	if output == "" || output == "console" || output == "stdout" {
		reporter.writer = os.Stdout
	} else {
		reporter.writer = nil // 将在报告时创建文件
	}

	return reporter
}

// SetFormat 设置格式
func (r *StandardMetricsReporter) SetFormat(format ReportFormat) {
	r.format = format
}

// SetOutput 设置输出
func (r *StandardMetricsReporter) SetOutput(output string) {
	r.output = output
	if output == "" || output == "console" || output == "stdout" {
		r.writer = os.Stdout
	} else {
		r.writer = nil
	}
}

// GetFormat 获取格式
func (r *StandardMetricsReporter) GetFormat() ReportFormat {
	return r.format
}

// GetOutput 获取输出
func (r *StandardMetricsReporter) GetOutput() string {
	return r.output
}

// Report 生成报告
func (r *StandardMetricsReporter) Report(metrics map[string]interface{}) error {
	var writer io.Writer
	var file *os.File
	var err error

	// 确定输出目标
	if r.writer != nil {
		writer = r.writer
	} else {
		// 创建输出文件
		file, err = r.createOutputFile()
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// 根据格式生成报告
	switch r.format {
	case FormatJSON:
		return r.reportJSON(writer, metrics)
	case FormatCSV:
		return r.reportCSV(writer, metrics)
	case FormatText:
		return r.reportText(writer, metrics)
	case FormatConsole:
		return r.reportConsole(writer, metrics)
	default:
		return fmt.Errorf("unsupported report format: %s", r.format)
	}
}

// createOutputFile 创建输出文件
func (r *StandardMetricsReporter) createOutputFile() (*os.File, error) {
	// 添加时间戳到文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := r.output
	
	// 如果文件名中没有扩展名，根据格式添加
	if !strings.Contains(filename, ".") {
		switch r.format {
		case FormatJSON:
			filename += "_" + timestamp + ".json"
		case FormatCSV:
			filename += "_" + timestamp + ".csv"
		case FormatText:
			filename += "_" + timestamp + ".txt"
		default:
			filename += "_" + timestamp + ".log"
		}
	} else {
		// 在扩展名前插入时间戳
		parts := strings.Split(filename, ".")
		if len(parts) >= 2 {
			parts[len(parts)-2] += "_" + timestamp
			filename = strings.Join(parts, ".")
		}
	}

	return os.Create(filename)
}

// reportJSON JSON格式报告
func (r *StandardMetricsReporter) reportJSON(writer io.Writer, metrics map[string]interface{}) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}

// reportCSV CSV格式报告
func (r *StandardMetricsReporter) reportCSV(writer io.Writer, metrics map[string]interface{}) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// 写入CSV标题
	headers := []string{"Metric", "Value", "Unit", "Category"}
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	// 解析并写入指标数据
	return r.writeMetricsToCSV(csvWriter, metrics, "")
}

// writeMetricsToCSV 将指标写入CSV
func (r *StandardMetricsReporter) writeMetricsToCSV(csvWriter *csv.Writer, data interface{}, prefix string) error {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			if err := r.writeMetricsToCSV(csvWriter, value, fullKey); err != nil {
				return err
			}
		}
	case *BasicMetrics:
		return r.writeBasicMetricsToCSV(csvWriter, v, prefix)
	case *LatencyMetrics:
		return r.writeLatencyMetricsToCSV(csvWriter, v, prefix)
	case *ConnectionStat:
		return r.writeConnectionStatsToCSV(csvWriter, v, prefix)
	case time.Duration:
		record := []string{prefix, v.String(), "duration", "timing"}
		return csvWriter.Write(record)
	case int64:
		record := []string{prefix, strconv.FormatInt(v, 10), "count", "counter"}
		return csvWriter.Write(record)
	case float64:
		record := []string{prefix, strconv.FormatFloat(v, 'f', 2, 64), "rate", "gauge"}
		return csvWriter.Write(record)
	case string:
		record := []string{prefix, v, "text", "info"}
		return csvWriter.Write(record)
	case time.Time:
		record := []string{prefix, v.Format("2006-01-02 15:04:05"), "timestamp", "time"}
		return csvWriter.Write(record)
	}
	return nil
}

// writeBasicMetricsToCSV 写入基础指标到CSV
func (r *StandardMetricsReporter) writeBasicMetricsToCSV(csvWriter *csv.Writer, metrics *BasicMetrics, prefix string) error {
	records := [][]string{
		{prefix + ".total_operations", strconv.FormatInt(metrics.TotalOperations, 10), "count", "basic"},
		{prefix + ".success_operations", strconv.FormatInt(metrics.SuccessOperations, 10), "count", "basic"},
		{prefix + ".failed_operations", strconv.FormatInt(metrics.FailedOperations, 10), "count", "basic"},
		{prefix + ".read_operations", strconv.FormatInt(metrics.ReadOperations, 10), "count", "basic"},
		{prefix + ".write_operations", strconv.FormatInt(metrics.WriteOperations, 10), "count", "basic"},
		{prefix + ".success_rate", strconv.FormatFloat(metrics.SuccessRate, 'f', 2, 64), "percentage", "basic"},
		{prefix + ".rps", strconv.FormatFloat(metrics.RPS, 'f', 2, 64), "rate", "basic"},
		{prefix + ".read_write_ratio", strconv.FormatFloat(metrics.ReadWriteRatio, 'f', 2, 64), "ratio", "basic"},
	}

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// writeLatencyMetricsToCSV 写入延迟指标到CSV
func (r *StandardMetricsReporter) writeLatencyMetricsToCSV(csvWriter *csv.Writer, metrics *LatencyMetrics, prefix string) error {
	records := [][]string{
		{prefix + ".min_latency", metrics.MinLatency.String(), "duration", "latency"},
		{prefix + ".max_latency", metrics.MaxLatency.String(), "duration", "latency"},
		{prefix + ".avg_latency", metrics.AvgLatency.String(), "duration", "latency"},
		{prefix + ".p50_latency", metrics.P50Latency.String(), "duration", "latency"},
		{prefix + ".p90_latency", metrics.P90Latency.String(), "duration", "latency"},
		{prefix + ".p95_latency", metrics.P95Latency.String(), "duration", "latency"},
		{prefix + ".p99_latency", metrics.P99Latency.String(), "duration", "latency"},
		{prefix + ".total_latency", metrics.TotalLatency.String(), "duration", "latency"},
	}

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// writeConnectionStatsToCSV 写入连接统计到CSV
func (r *StandardMetricsReporter) writeConnectionStatsToCSV(csvWriter *csv.Writer, stats *ConnectionStat, prefix string) error {
	records := [][]string{
		{prefix + ".total_connections", strconv.FormatInt(stats.TotalConnections, 10), "count", "connection"},
		{prefix + ".active_connections", strconv.FormatInt(stats.ActiveConnections, 10), "count", "connection"},
		{prefix + ".failed_connections", strconv.FormatInt(stats.FailedConnections, 10), "count", "connection"},
		{prefix + ".connection_latency", stats.ConnectionLatency.String(), "duration", "connection"},
		{prefix + ".reconnect_count", strconv.FormatInt(stats.ReconnectCount, 10), "count", "connection"},
	}

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// reportText 文本格式报告
func (r *StandardMetricsReporter) reportText(writer io.Writer, metrics map[string]interface{}) error {
	_, err := fmt.Fprintf(writer, "Redis Performance Metrics Report\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "Generated at: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}

	return r.writeMetricsAsText(writer, metrics, 0)
}

// writeMetricsAsText 将指标写为文本格式
func (r *StandardMetricsReporter) writeMetricsAsText(writer io.Writer, data interface{}, indent int) error {
	indentStr := strings.Repeat("  ", indent)

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			_, err := fmt.Fprintf(writer, "%s%s:\n", indentStr, strings.Title(key))
			if err != nil {
				return err
			}
			if err := r.writeMetricsAsText(writer, value, indent+1); err != nil {
				return err
			}
		}
	case *BasicMetrics:
		return r.writeBasicMetricsAsText(writer, v, indent)
	case *LatencyMetrics:
		return r.writeLatencyMetricsAsText(writer, v, indent)
	case *ConnectionStat:
		return r.writeConnectionStatsAsText(writer, v, indent)
	case time.Duration:
		_, err := fmt.Fprintf(writer, "%s%s\n", indentStr, v.String())
		return err
	case int64:
		_, err := fmt.Fprintf(writer, "%s%d\n", indentStr, v)
		return err
	case float64:
		_, err := fmt.Fprintf(writer, "%s%.2f\n", indentStr, v)
		return err
	case string:
		_, err := fmt.Fprintf(writer, "%s%s\n", indentStr, v)
		return err
	case time.Time:
		_, err := fmt.Fprintf(writer, "%s%s\n", indentStr, v.Format("2006-01-02 15:04:05"))
		return err
	}
	return nil
}

// writeBasicMetricsAsText 写入基础指标为文本
func (r *StandardMetricsReporter) writeBasicMetricsAsText(writer io.Writer, metrics *BasicMetrics, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	
	lines := []string{
		fmt.Sprintf("%sTotal Operations: %d", indentStr, metrics.TotalOperations),
		fmt.Sprintf("%sSuccess Operations: %d", indentStr, metrics.SuccessOperations),
		fmt.Sprintf("%sFailed Operations: %d", indentStr, metrics.FailedOperations),
		fmt.Sprintf("%sRead Operations: %d", indentStr, metrics.ReadOperations),
		fmt.Sprintf("%sWrite Operations: %d", indentStr, metrics.WriteOperations),
		fmt.Sprintf("%sSuccess Rate: %.2f%%", indentStr, metrics.SuccessRate),
		fmt.Sprintf("%sRPS: %.2f", indentStr, metrics.RPS),
		fmt.Sprintf("%sRead/Write Ratio: %.2f", indentStr, metrics.ReadWriteRatio),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return err
		}
	}
	return nil
}

// writeLatencyMetricsAsText 写入延迟指标为文本
func (r *StandardMetricsReporter) writeLatencyMetricsAsText(writer io.Writer, metrics *LatencyMetrics, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	
	lines := []string{
		fmt.Sprintf("%sMin Latency: %s", indentStr, metrics.MinLatency),
		fmt.Sprintf("%sMax Latency: %s", indentStr, metrics.MaxLatency),
		fmt.Sprintf("%sAvg Latency: %s", indentStr, metrics.AvgLatency),
		fmt.Sprintf("%sP50 Latency: %s", indentStr, metrics.P50Latency),
		fmt.Sprintf("%sP90 Latency: %s", indentStr, metrics.P90Latency),
		fmt.Sprintf("%sP95 Latency: %s", indentStr, metrics.P95Latency),
		fmt.Sprintf("%sP99 Latency: %s", indentStr, metrics.P99Latency),
		fmt.Sprintf("%sTotal Latency: %s", indentStr, metrics.TotalLatency),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return err
		}
	}
	return nil
}

// writeConnectionStatsAsText 写入连接统计为文本
func (r *StandardMetricsReporter) writeConnectionStatsAsText(writer io.Writer, stats *ConnectionStat, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	
	lines := []string{
		fmt.Sprintf("%sTotal Connections: %d", indentStr, stats.TotalConnections),
		fmt.Sprintf("%sActive Connections: %d", indentStr, stats.ActiveConnections),
		fmt.Sprintf("%sFailed Connections: %d", indentStr, stats.FailedConnections),
		fmt.Sprintf("%sConnection Latency: %s", indentStr, stats.ConnectionLatency),
		fmt.Sprintf("%sReconnect Count: %d", indentStr, stats.ReconnectCount),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return err
		}
	}
	return nil
}

// reportConsole 控制台格式报告
func (r *StandardMetricsReporter) reportConsole(writer io.Writer, metrics map[string]interface{}) error {
	// 控制台报告格式更加简洁和易读
	_, err := fmt.Fprintf(writer, "\n=== Redis Performance Metrics ===\n")
	if err != nil {
		return err
	}

	// 提取主要指标
	if basicMetrics, exists := metrics["basic_metrics"]; exists {
		if bm, ok := basicMetrics.(*BasicMetrics); ok {
			_, err = fmt.Fprintf(writer, "Operations: %d total, %d success (%.1f%%), %.1f ops/sec\n",
				bm.TotalOperations, bm.SuccessOperations, bm.SuccessRate, bm.RPS)
			if err != nil {
				return err
			}
		}
	}

	if latencyMetrics, exists := metrics["latency_metrics"]; exists {
		if lm, ok := latencyMetrics.(*LatencyMetrics); ok {
			_, err = fmt.Fprintf(writer, "Latency: avg=%s, p50=%s, p95=%s, p99=%s\n",
				lm.AvgLatency, lm.P50Latency, lm.P95Latency, lm.P99Latency)
			if err != nil {
				return err
			}
		}
	}

	if duration, exists := metrics["duration"]; exists {
		if d, ok := duration.(time.Duration); ok {
			_, err = fmt.Fprintf(writer, "Duration: %s\n", d)
			if err != nil {
				return err
			}
		}
	}

	_, err = fmt.Fprintf(writer, "==================================\n\n")
	return err
}

// MultiFormatReporter 多格式报告器
type MultiFormatReporter struct {
	reporters []MetricsReporter
}

// NewMultiFormatReporter 创建多格式报告器
func NewMultiFormatReporter() *MultiFormatReporter {
	return &MultiFormatReporter{
		reporters: make([]MetricsReporter, 0),
	}
}

// AddReporter 添加报告器
func (mr *MultiFormatReporter) AddReporter(reporter MetricsReporter) {
	mr.reporters = append(mr.reporters, reporter)
}

// Report 生成所有格式的报告
func (mr *MultiFormatReporter) Report(metrics map[string]interface{}) error {
	for _, reporter := range mr.reporters {
		if err := reporter.Report(metrics); err != nil {
			return fmt.Errorf("reporter %T failed: %w", reporter, err)
		}
	}
	return nil
}

// ReportBuilder 报告构建器
type ReportBuilder struct {
	collector *MetricsCollector
	reporters []MetricsReporter
}

// NewReportBuilder 创建报告构建器
func NewReportBuilder(collector *MetricsCollector) *ReportBuilder {
	return &ReportBuilder{
		collector: collector,
		reporters: make([]MetricsReporter, 0),
	}
}

// WithFormat 添加格式报告器
func (rb *ReportBuilder) WithFormat(format ReportFormat, output string) *ReportBuilder {
	reporter := NewMetricsReporter(format, output)
	rb.reporters = append(rb.reporters, reporter)
	return rb
}

// WithJSON 添加JSON报告器
func (rb *ReportBuilder) WithJSON(output string) *ReportBuilder {
	return rb.WithFormat(FormatJSON, output)
}

// WithCSV 添加CSV报告器
func (rb *ReportBuilder) WithCSV(output string) *ReportBuilder {
	return rb.WithFormat(FormatCSV, output)
}

// WithText 添加文本报告器
func (rb *ReportBuilder) WithText(output string) *ReportBuilder {
	return rb.WithFormat(FormatText, output)
}

// WithConsole 添加控制台报告器
func (rb *ReportBuilder) WithConsole() *ReportBuilder {
	return rb.WithFormat(FormatConsole, "console")
}

// Generate 生成报告
func (rb *ReportBuilder) Generate() error {
	metrics := rb.collector.GetMetrics()
	
	for _, reporter := range rb.reporters {
		if err := reporter.Report(metrics); err != nil {
			return fmt.Errorf("failed to generate report with %T: %w", reporter, err)
		}
	}
	
	return nil
}