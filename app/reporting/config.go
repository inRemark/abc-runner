package reporting

// NewStandardReportConfig 创建标准报告配置
// 为所有protocol的性能测试提供统一的报告配置
func NewStandardReportConfig(protocolPrefix string) *RenderConfig {
	return &RenderConfig{
		OutputFormats: []string{"console", "json", "csv", "html"},
		OutputDir:     "./reports",
		FilePrefix:    protocolPrefix + "_performance",
		Timestamp:     true,
	}
}

// GetSupportedFormats 获取支持的报告格式列表
func GetSupportedFormats() []string {
	return []string{"console", "json", "csv", "html"}
}

// GetDefaultOutputDir 获取默认输出目录
func GetDefaultOutputDir() string {
	return "./reports"
}
