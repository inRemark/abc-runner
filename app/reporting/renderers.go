package reporting

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Renderer 渲染器接口
type Renderer interface {
	Render(report *StructuredReport) ([]byte, error)
	Format() string
	Extension() string
}

// RenderConfig 渲染配置
type RenderConfig struct {
	OutputFormats []string `json:"output_formats"`
	OutputDir     string   `json:"output_dir"`
	FilePrefix    string   `json:"file_prefix"`
	Timestamp     bool     `json:"timestamp"`
}

// DefaultRenderConfig 默认渲染配置
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		OutputFormats: []string{"console", "json"},
		OutputDir:     "./reports",
		FilePrefix:    "benchmark",
		Timestamp:     true,
	}
}

// ConsoleRenderer 控制台渲染器
type ConsoleRenderer struct{}

func NewConsoleRenderer() *ConsoleRenderer {
	return &ConsoleRenderer{}
}

func (c *ConsoleRenderer) Format() string {
	return "console"
}

func (c *ConsoleRenderer) Extension() string {
	return ""
}

func (c *ConsoleRenderer) Render(report *StructuredReport) ([]byte, error) {
	var buf bytes.Buffer
	
	// 报告头部
	buf.WriteString("\n" + strings.Repeat("=", 80) + "\n")
	buf.WriteString("             ABC-RUNNER 性能测试报告\n")
	buf.WriteString(strings.Repeat("=", 80) + "\n")
	
	// 执行摘要
	buf.WriteString("\n📊 执行摘要\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")
	buf.WriteString(fmt.Sprintf("性能评分: %d/100\n", report.Dashboard.PerformanceScore))
	buf.WriteString(fmt.Sprintf("系统状态: %s\n", c.formatStatus(report.Dashboard.StatusIndicator)))
	buf.WriteString(fmt.Sprintf("协议类型: %s\n", report.Context.TestConfiguration.Protocol))
	buf.WriteString(fmt.Sprintf("测试时长: %v\n", report.Context.TestConfiguration.TestDuration))
	
	// 核心指标
	buf.WriteString("\n⚡ 核心性能指标\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")
	ops := report.Metrics.CoreOperations
	buf.WriteString(fmt.Sprintf("总操作数: %d\n", ops.TotalOperations))
	buf.WriteString(fmt.Sprintf("成功操作: %d (%.2f%%)\n", ops.SuccessfulOps, ops.SuccessRate))
	buf.WriteString(fmt.Sprintf("失败操作: %d (%.2f%%)\n", ops.FailedOps, ops.ErrorRate))
	buf.WriteString(fmt.Sprintf("吞吐量: %.2f ops/sec\n", ops.OperationsPerSecond))
	
	// 延迟分析
	buf.WriteString("\n🚀 延迟分析\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")
	latency := report.Metrics.LatencyAnalysis
	buf.WriteString(fmt.Sprintf("平均延迟: %v\n", latency.AverageLatency))
	buf.WriteString(fmt.Sprintf("最小延迟: %v\n", latency.MinLatency))
	buf.WriteString(fmt.Sprintf("最大延迟: %v\n", latency.MaxLatency))
	buf.WriteString("延迟百分位:\n")
	buf.WriteString(fmt.Sprintf("  P50: %v\n", latency.Percentiles.P50))
	buf.WriteString(fmt.Sprintf("  P90: %v\n", latency.Percentiles.P90))
	buf.WriteString(fmt.Sprintf("  P95: %v\n", latency.Percentiles.P95))
	buf.WriteString(fmt.Sprintf("  P99: %v\n", latency.Percentiles.P99))
	
	// 系统健康状态
	buf.WriteString("\n💻 系统健康状态\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")
	system := report.System
	buf.WriteString(fmt.Sprintf("内存使用: %.2f%%\n", system.MemoryProfile.MemoryUsagePercent))
	buf.WriteString(fmt.Sprintf("活跃协程: %d\n", system.RuntimeMetrics.ActiveGoroutines))
	buf.WriteString(fmt.Sprintf("GC次数: %d\n", system.MemoryProfile.GCCount))
	
	// 关键洞察
	if len(report.Dashboard.KeyInsights) > 0 {
		buf.WriteString("\n💡 关键洞察\n")
		buf.WriteString(strings.Repeat("-", 40) + "\n")
		for _, insight := range report.Dashboard.KeyInsights {
			buf.WriteString(fmt.Sprintf("• %s: %s\n", insight.Title, insight.Description))
		}
	}
	
	// 优化建议
	if len(report.Dashboard.Recommendations) > 0 {
		buf.WriteString("\n🔧 优化建议\n")
		buf.WriteString(strings.Repeat("-", 40) + "\n")
		for _, rec := range report.Dashboard.Recommendations {
			buf.WriteString(fmt.Sprintf("• [%s] %s: %s\n", 
				strings.ToUpper(string(rec.Priority)), 
				rec.Category, 
				rec.Action))
		}
	}
	
	buf.WriteString("\n" + strings.Repeat("=", 80) + "\n")
	buf.WriteString(fmt.Sprintf("报告生成时间: %s\n", report.Context.ExecutionContext.GeneratedAt.Format("2006-01-02 15:04:05")))
	buf.WriteString(strings.Repeat("=", 80) + "\n")
	
	return buf.Bytes(), nil
}

func (c *ConsoleRenderer) formatStatus(status StatusLevel) string {
	switch status {
	case StatusGood:
		return "🟢 良好"
	case StatusWarning:
		return "🟡 警告"
	case StatusCritical:
		return "🔴 严重"
	default:
		return "⚪ 未知"
	}
}

// JSONRenderer JSON渲染器
type JSONRenderer struct{}

func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

func (j *JSONRenderer) Format() string {
	return "json"
}

func (j *JSONRenderer) Extension() string {
	return "json"
}

func (j *JSONRenderer) Render(report *StructuredReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// CSVRenderer CSV渲染器
type CSVRenderer struct{}

func NewCSVRenderer() *CSVRenderer {
	return &CSVRenderer{}
}

func (c *CSVRenderer) Format() string {
	return "csv"
}

func (c *CSVRenderer) Extension() string {
	return "csv"
}

func (c *CSVRenderer) Render(report *StructuredReport) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// 写入标题行
	headers := []string{
		"timestamp", "protocol", "performance_score", "status",
		"total_ops", "successful_ops", "failed_ops", "success_rate", "error_rate", "rps",
		"avg_latency_ms", "min_latency_ms", "max_latency_ms", 
		"p90_latency_ms", "p95_latency_ms", "p99_latency_ms",
		"memory_usage_percent", "active_goroutines", "gc_count",
	}
	
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}
	
	// 写入数据行
	record := []string{
		report.Context.ExecutionContext.GeneratedAt.Format(time.RFC3339),
		report.Context.TestConfiguration.Protocol,
		fmt.Sprintf("%d", report.Dashboard.PerformanceScore),
		string(report.Dashboard.StatusIndicator),
		fmt.Sprintf("%d", report.Metrics.CoreOperations.TotalOperations),
		fmt.Sprintf("%d", report.Metrics.CoreOperations.SuccessfulOps),
		fmt.Sprintf("%d", report.Metrics.CoreOperations.FailedOps),
		fmt.Sprintf("%.2f", report.Metrics.CoreOperations.SuccessRate),
		fmt.Sprintf("%.2f", report.Metrics.CoreOperations.ErrorRate),
		fmt.Sprintf("%.2f", report.Metrics.CoreOperations.OperationsPerSecond),
		fmt.Sprintf("%.3f", float64(report.Metrics.LatencyAnalysis.AverageLatency.Nanoseconds())/1000000),
		fmt.Sprintf("%.3f", float64(report.Metrics.LatencyAnalysis.MinLatency.Nanoseconds())/1000000),
		fmt.Sprintf("%.3f", float64(report.Metrics.LatencyAnalysis.MaxLatency.Nanoseconds())/1000000),
		fmt.Sprintf("%.3f", float64(report.Metrics.LatencyAnalysis.Percentiles.P90.Nanoseconds())/1000000),
		fmt.Sprintf("%.3f", float64(report.Metrics.LatencyAnalysis.Percentiles.P95.Nanoseconds())/1000000),
		fmt.Sprintf("%.3f", float64(report.Metrics.LatencyAnalysis.Percentiles.P99.Nanoseconds())/1000000),
		fmt.Sprintf("%.2f", report.System.MemoryProfile.MemoryUsagePercent),
		fmt.Sprintf("%d", report.System.RuntimeMetrics.ActiveGoroutines),
		fmt.Sprintf("%d", report.System.MemoryProfile.GCCount),
	}
	
	if err := writer.Write(record); err != nil {
		return nil, fmt.Errorf("failed to write CSV record: %w", err)
	}
	
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}
	
	return buf.Bytes(), nil
}

// HTMLRenderer HTML渲染器
type HTMLRenderer struct{}

func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

func (h *HTMLRenderer) Format() string {
	return "html"
}

func (h *HTMLRenderer) Extension() string {
	return "html"
}

func (h *HTMLRenderer) Render(report *StructuredReport) ([]byte, error) {
	tmpl := template.Must(template.New("report").Parse(htmlTemplate))
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report); err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}
	
	return buf.Bytes(), nil
}

// ReportGenerator 统一报告生成器
type ReportGenerator struct {
	config    *RenderConfig
	renderers map[string]Renderer
}

// NewReportGenerator 创建报告生成器
func NewReportGenerator(config *RenderConfig) *ReportGenerator {
	if config == nil {
		config = DefaultRenderConfig()
	}
	
	generator := &ReportGenerator{
		config:    config,
		renderers: make(map[string]Renderer),
	}
	
	// 注册内置渲染器
	generator.renderers["console"] = NewConsoleRenderer()
	generator.renderers["json"] = NewJSONRenderer()
	generator.renderers["csv"] = NewCSVRenderer()
	generator.renderers["html"] = NewHTMLRenderer()
	
	return generator
}

// Generate 生成所有格式的报告
func (g *ReportGenerator) Generate(report *StructuredReport) error {
	// 确保输出目录存在
	if g.config.OutputDir != "" {
		if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	
	for _, format := range g.config.OutputFormats {
		if err := g.renderFormat(report, format); err != nil {
			return fmt.Errorf("failed to render %s format: %w", format, err)
		}
	}
	
	return nil
}

// renderFormat 渲染指定格式
func (g *ReportGenerator) renderFormat(report *StructuredReport, format string) error {
	renderer, exists := g.renderers[format]
	if !exists {
		return fmt.Errorf("unsupported format: %s", format)
	}
	
	content, err := renderer.Render(report)
	if err != nil {
		return fmt.Errorf("rendering failed: %w", err)
	}
	
	if format == "console" {
		// 控制台输出直接打印
		fmt.Print(string(content))
		return nil
	}
	
	// 其他格式保存到文件
	filename := g.generateFilename(renderer)
	if err := g.writeToFile(filename, content); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}
	
	fmt.Printf("✅ %s report saved to: %s\n", strings.ToUpper(format), filename)
	return nil
}

// generateFilename 生成文件名
func (g *ReportGenerator) generateFilename(renderer Renderer) string {
	filename := g.config.FilePrefix
	
	if g.config.Timestamp {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("%s_%s", filename, timestamp)
	}
	
	filename = fmt.Sprintf("%s.%s", filename, renderer.Extension())
	
	if g.config.OutputDir != "" {
		filename = filepath.Join(g.config.OutputDir, filename)
	}
	
	return filename
}

// writeToFile 写入文件
func (g *ReportGenerator) writeToFile(filename string, content []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = file.Write(content)
	return err
}

// RegisterRenderer 注册自定义渲染器
func (g *ReportGenerator) RegisterRenderer(format string, renderer Renderer) {
	g.renderers[format] = renderer
}

// HTML模板
const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ABC-Runner 性能测试报告</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 8px 8px 0 0; }
        .header h1 { margin: 0; font-size: 2.5em; }
        .header .subtitle { opacity: 0.9; margin-top: 10px; }
        .content { padding: 30px; }
        .section { margin-bottom: 40px; }
        .section h2 { color: #333; border-bottom: 2px solid #667eea; padding-bottom: 10px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-top: 20px; }
        .metric-card { background: #f8f9fa; padding: 20px; border-radius: 6px; border-left: 4px solid #667eea; }
        .metric-value { font-size: 2em; font-weight: bold; color: #667eea; }
        .metric-label { color: #666; margin-top: 5px; }
        .status-good { color: #28a745; }
        .status-warning { color: #ffc107; }
        .status-critical { color: #dc3545; }
        .insights ul, .recommendations ul { list-style: none; padding: 0; }
        .insights li, .recommendations li { background: #f8f9fa; margin: 10px 0; padding: 15px; border-radius: 6px; border-left: 4px solid #17a2b8; }
        .footer { text-align: center; padding: 20px; color: #666; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ABC-Runner 性能测试报告</h1>
            <div class="subtitle">协议: {{.Context.TestConfiguration.Protocol}} | 生成时间: {{.Context.ExecutionContext.GeneratedAt.Format "2006-01-02 15:04:05"}}</div>
        </div>
        
        <div class="content">
            <div class="section">
                <h2>📊 执行摘要</h2>
                <div class="metrics-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.Dashboard.PerformanceScore}}/100</div>
                        <div class="metric-label">性能评分</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value status-{{.Dashboard.StatusIndicator}}">{{.Dashboard.StatusIndicator}}</div>
                        <div class="metric-label">系统状态</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Metrics.CoreOperations.TotalOperations}}</div>
                        <div class="metric-label">总操作数</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Metrics.CoreOperations.OperationsPerSecond}}</div>
                        <div class="metric-label">吞吐量 (ops/sec)</div>
                    </div>
                </div>
            </div>
            
            <div class="section">
                <h2>⚡ 核心性能指标</h2>
                <div class="metrics-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f%%" .Metrics.CoreOperations.SuccessRate}}</div>
                        <div class="metric-label">成功率</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f%%" .Metrics.CoreOperations.ErrorRate}}</div>
                        <div class="metric-label">错误率</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Metrics.LatencyAnalysis.AverageLatency}}</div>
                        <div class="metric-label">平均延迟</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Metrics.LatencyAnalysis.Percentiles.P99}}</div>
                        <div class="metric-label">P99延迟</div>
                    </div>
                </div>
            </div>
            
            {{if .Dashboard.KeyInsights}}
            <div class="section insights">
                <h2>💡 关键洞察</h2>
                <ul>
                    {{range .Dashboard.KeyInsights}}
                    <li><strong>{{.Title}}</strong>: {{.Description}}</li>
                    {{end}}
                </ul>
            </div>
            {{end}}
            
            {{if .Dashboard.Recommendations}}
            <div class="section recommendations">
                <h2>🔧 优化建议</h2>
                <ul>
                    {{range .Dashboard.Recommendations}}
                    <li><strong>[{{.Priority | upper}}] {{.Category}}</strong>: {{.Action}}</li>
                    {{end}}
                </ul>
            </div>
            {{end}}
        </div>
        
        <div class="footer">
            <p>由 ABC-Runner {{.Context.Environment.ABCRunnerVersion}} 生成 | 会话ID: {{.Context.ExecutionContext.UniqueSessionID}}</p>
        </div>
    </div>
</body>
</html>
`