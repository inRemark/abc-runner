package metrics

import (
	"fmt"
	"strings"
	"time"

	"abc-runner/app/core/interfaces"
)

// HttpMetricsReporter HTTP指标报告器
type HttpMetricsReporter struct {
	collector *MetricsCollector
}

// NewHttpMetricsReporter 创建HTTP指标报告器
func NewHttpMetricsReporter(collector *MetricsCollector) *HttpMetricsReporter {
	return &HttpMetricsReporter{
		collector: collector,
	}
}

// GenerateReport 生成报告
func (r *HttpMetricsReporter) GenerateReport() string {
	var report strings.Builder

	// 基础指标
	baseMetrics := r.collector.GetMetrics()
	report.WriteString(r.formatBaseMetrics(baseMetrics))

	// HTTP特定指标
	httpMetrics := r.collector.GetHttpSpecificMetrics()
	report.WriteString(r.formatHttpMetrics(httpMetrics))

	return report.String()
}

// GenerateSimpleReport 生成简单报告
func (r *HttpMetricsReporter) GenerateSimpleReport() string {
	baseMetrics := r.collector.GetMetrics()

	return fmt.Sprintf(`HTTP Performance Report:
  Total Requests: %d
  Successful: %d (%.2f%%)
  Failed: %d (%.2f%%)
  RPS: %d
  Average Latency: %v
  P95 Latency: %v
  P99 Latency: %v
  Error Rate: %.2f%%
  Duration: %v
`,
		baseMetrics.TotalOps,
		baseMetrics.SuccessOps,
		float64(baseMetrics.SuccessOps)/float64(baseMetrics.TotalOps)*100,
		baseMetrics.FailedOps,
		baseMetrics.ErrorRate,
		baseMetrics.RPS,
		baseMetrics.AvgLatency,
		baseMetrics.P95Latency,
		baseMetrics.P99Latency,
		baseMetrics.ErrorRate,
		baseMetrics.Duration,
	)
}

// formatBaseMetrics 格式化基础指标
func (r *HttpMetricsReporter) formatBaseMetrics(metrics *interfaces.Metrics) string {
	var report strings.Builder

	report.WriteString("=== HTTP Performance Summary ===\n")
	report.WriteString(fmt.Sprintf("Test Duration: %v\n", metrics.Duration))
	report.WriteString(fmt.Sprintf("Start Time: %v\n", metrics.StartTime.Format(time.RFC3339)))
	report.WriteString(fmt.Sprintf("End Time: %v\n", metrics.EndTime.Format(time.RFC3339)))
	report.WriteString("\n")

	report.WriteString("=== Request Statistics ===\n")
	report.WriteString(fmt.Sprintf("Total Requests: %d\n", metrics.TotalOps))
	report.WriteString(fmt.Sprintf("Successful Requests: %d (%.2f%%)\n",
		metrics.SuccessOps, float64(metrics.SuccessOps)/float64(metrics.TotalOps)*100))
	report.WriteString(fmt.Sprintf("Failed Requests: %d (%.2f%%)\n",
		metrics.FailedOps, float64(metrics.FailedOps)/float64(metrics.TotalOps)*100))
	report.WriteString(fmt.Sprintf("Read Operations: %d\n", metrics.ReadOps))
	report.WriteString(fmt.Sprintf("Write Operations: %d\n", metrics.WriteOps))
	report.WriteString(fmt.Sprintf("Requests per Second: %d\n", metrics.RPS))
	report.WriteString("\n")

	report.WriteString("=== Latency Statistics ===\n")
	report.WriteString(fmt.Sprintf("Minimum Latency: %v\n", metrics.MinLatency))
	report.WriteString(fmt.Sprintf("Maximum Latency: %v\n", metrics.MaxLatency))
	report.WriteString(fmt.Sprintf("Average Latency: %v\n", metrics.AvgLatency))
	report.WriteString(fmt.Sprintf("P90 Latency: %v\n", metrics.P90Latency))
	report.WriteString(fmt.Sprintf("P95 Latency: %v\n", metrics.P95Latency))
	report.WriteString(fmt.Sprintf("P99 Latency: %v\n", metrics.P99Latency))
	report.WriteString(fmt.Sprintf("Error Rate: %.2f%%\n", metrics.ErrorRate))
	report.WriteString("\n")

	return report.String()
}

// formatHttpMetrics 格式化HTTP特定指标
func (r *HttpMetricsReporter) formatHttpMetrics(metrics map[string]interface{}) string {
	var report strings.Builder

	// 状态码分布
	if statusCodes, ok := metrics["status_codes"].(map[int]int64); ok && len(statusCodes) > 0 {
		report.WriteString("=== HTTP Status Code Distribution ===\n")
		for code, count := range statusCodes {
			report.WriteString(fmt.Sprintf("  %d: %d requests\n", code, count))
		}
		report.WriteString("\n")
	}

	// 请求方法统计
	if methods, ok := metrics["methods"].(map[string]int64); ok && len(methods) > 0 {
		report.WriteString("=== HTTP Method Distribution ===\n")
		for method, count := range methods {
			report.WriteString(fmt.Sprintf("  %s: %d requests\n", method, count))
		}
		report.WriteString("\n")
	}

	// 错误类型统计
	if errorTypes, ok := metrics["error_types"].(map[string]int64); ok && len(errorTypes) > 0 {
		report.WriteString("=== Error Type Distribution ===\n")
		for errorType, count := range errorTypes {
			report.WriteString(fmt.Sprintf("  %s: %d errors\n", errorType, count))
		}
		report.WriteString("\n")
	}

	// 响应大小统计
	if responseSizes, ok := metrics["response_sizes"].(map[string]interface{}); ok && responseSizes != nil {
		report.WriteString("=== Response Size Statistics ===\n")
		if min, ok := responseSizes["min"].(int64); ok {
			report.WriteString(fmt.Sprintf("  Minimum Size: %d bytes\n", min))
		}
		if max, ok := responseSizes["max"].(int64); ok {
			report.WriteString(fmt.Sprintf("  Maximum Size: %d bytes\n", max))
		}
		if avg, ok := responseSizes["avg"].(int64); ok {
			report.WriteString(fmt.Sprintf("  Average Size: %d bytes\n", avg))
		}
		if total, ok := responseSizes["total"].(int64); ok {
			report.WriteString(fmt.Sprintf("  Total Data: %d bytes (%.2f MB)\n", total, float64(total)/(1024*1024)))
		}
		report.WriteString("\n")
	}

	// TLS握手时间统计
	if tlsStats, ok := metrics["tls_handshake_times"].(map[string]interface{}); ok && tlsStats != nil {
		report.WriteString("=== TLS Handshake Time Statistics ===\n")
		if avg, ok := tlsStats["avg"].(int64); ok {
			report.WriteString(fmt.Sprintf("  Average TLS Handshake: %v\n", time.Duration(avg)))
		}
		if p95, ok := tlsStats["p95"].(int64); ok {
			report.WriteString(fmt.Sprintf("  P95 TLS Handshake: %v\n", time.Duration(p95)))
		}
		if p99, ok := tlsStats["p99"].(int64); ok {
			report.WriteString(fmt.Sprintf("  P99 TLS Handshake: %v\n", time.Duration(p99)))
		}
		report.WriteString("\n")
	}

	// 上传统计
	if uploadStats, ok := metrics["upload_stats"].(map[string]interface{}); ok && uploadStats != nil {
		report.WriteString("=== File Upload Statistics ===\n")
		if avgSpeed, ok := uploadStats["avg_upload_speed"].(float64); ok {
			report.WriteString(fmt.Sprintf("  Average Upload Speed: %.2f MB/s\n", avgSpeed))
		}
		if fileSizes, ok := uploadStats["file_sizes"].(map[string]interface{}); ok {
			if avg, ok := fileSizes["avg"].(int64); ok {
				report.WriteString(fmt.Sprintf("  Average File Size: %d bytes (%.2f MB)\n", avg, float64(avg)/(1024*1024)))
			}
			if total, ok := fileSizes["total"].(int64); ok {
				report.WriteString(fmt.Sprintf("  Total Upload Data: %d bytes (%.2f MB)\n", total, float64(total)/(1024*1024)))
			}
		}
		report.WriteString("\n")
	}

	// 网络时间统计
	if networkTimes, ok := metrics["network_times"].(map[string]interface{}); ok && len(networkTimes) > 0 {
		report.WriteString("=== Network Time Breakdown ===\n")
		if dns, ok := networkTimes["dns_resolve"].(map[string]interface{}); ok {
			if avg, ok := dns["avg"].(int64); ok {
				report.WriteString(fmt.Sprintf("  DNS Resolve Time: %v\n", time.Duration(avg)))
			}
		}
		if conn, ok := networkTimes["connection"].(map[string]interface{}); ok {
			if avg, ok := conn["avg"].(int64); ok {
				report.WriteString(fmt.Sprintf("  Connection Time: %v\n", time.Duration(avg)))
			}
		}
		if firstByte, ok := networkTimes["first_byte"].(map[string]interface{}); ok {
			if avg, ok := firstByte["avg"].(int64); ok {
				report.WriteString(fmt.Sprintf("  First Byte Time: %v\n", time.Duration(avg)))
			}
		}
		report.WriteString("\n")
	}

	// 计数器统计
	if redirectCount, ok := metrics["redirect_count"].(int64); ok && redirectCount > 0 {
		report.WriteString(fmt.Sprintf("Redirects: %d\n", redirectCount))
	}
	if timeoutCount, ok := metrics["timeout_count"].(int64); ok && timeoutCount > 0 {
		report.WriteString(fmt.Sprintf("Timeouts: %d\n", timeoutCount))
	}
	if connErrorCount, ok := metrics["connection_error_count"].(int64); ok && connErrorCount > 0 {
		report.WriteString(fmt.Sprintf("Connection Errors: %d\n", connErrorCount))
	}

	return report.String()
}

// GenerateJSONReport 生成JSON格式报告
func (r *HttpMetricsReporter) GenerateJSONReport() map[string]interface{} {
	result := make(map[string]interface{})

	// 基础指标
	baseMetrics := r.collector.GetMetrics()
	result["base_metrics"] = map[string]interface{}{
		"total_ops":   baseMetrics.TotalOps,
		"success_ops": baseMetrics.SuccessOps,
		"failed_ops":  baseMetrics.FailedOps,
		"read_ops":    baseMetrics.ReadOps,
		"write_ops":   baseMetrics.WriteOps,
		"rps":         baseMetrics.RPS,
		"avg_latency": baseMetrics.AvgLatency.Nanoseconds(),
		"min_latency": baseMetrics.MinLatency.Nanoseconds(),
		"max_latency": baseMetrics.MaxLatency.Nanoseconds(),
		"p90_latency": baseMetrics.P90Latency.Nanoseconds(),
		"p95_latency": baseMetrics.P95Latency.Nanoseconds(),
		"p99_latency": baseMetrics.P99Latency.Nanoseconds(),
		"error_rate":  baseMetrics.ErrorRate,
		"duration":    baseMetrics.Duration.Nanoseconds(),
		"start_time":  baseMetrics.StartTime.Unix(),
		"end_time":    baseMetrics.EndTime.Unix(),
	}

	// HTTP特定指标
	result["http_metrics"] = r.collector.GetHttpSpecificMetrics()

	return result
}

// GenerateCSVHeader 生成CSV头部
func (r *HttpMetricsReporter) GenerateCSVHeader() string {
	return "timestamp,total_ops,success_ops,failed_ops,rps,avg_latency_ms,p95_latency_ms,p99_latency_ms,error_rate"
}

// GenerateCSVRow 生成CSV行数据
func (r *HttpMetricsReporter) GenerateCSVRow() string {
	metrics := r.collector.GetMetrics()
	return fmt.Sprintf("%d,%d,%d,%d,%d,%.2f,%.2f,%.2f,%.2f",
		time.Now().Unix(),
		metrics.TotalOps,
		metrics.SuccessOps,
		metrics.FailedOps,
		metrics.RPS,
		float64(metrics.AvgLatency.Nanoseconds())/1000000, // 转换为毫秒
		float64(metrics.P95Latency.Nanoseconds())/1000000,
		float64(metrics.P99Latency.Nanoseconds())/1000000,
		metrics.ErrorRate,
	)
}
