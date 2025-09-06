package unified

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
)

// smartErrorHandler 智能错误处理器实现
type smartErrorHandler struct {
	suggestionRules map[string][]string
	mutex          sync.RWMutex
}

// NewSmartErrorHandler 创建智能错误处理器
func NewSmartErrorHandler() SmartErrorHandler {
	handler := &smartErrorHandler{
		suggestionRules: make(map[string][]string),
		mutex:          sync.RWMutex{},
	}
	
	// 添加默认建议规则
	handler.setupDefaultRules()
	
	return handler
}

// HandleError 处理错误并提供智能建议
func (h *smartErrorHandler) HandleError(err error, context *CommandRequest) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if err == nil {
		return nil
	}
	
	errorMsg := err.Error()
	
	// 检查是否为未知命令错误
	if strings.Contains(errorMsg, "unknown command") {
		suggestions := h.GetSuggestions(context.Command)
		if len(suggestions) > 0 {
			enhancedMsg := fmt.Sprintf("%s\n\nDid you mean one of these?\n", errorMsg)
			for _, suggestion := range suggestions {
				enhancedMsg += fmt.Sprintf("  - %s\n", suggestion)
			}
			return fmt.Errorf(enhancedMsg)
		}
	}
	
	// 检查是否为协议不支持错误
	if strings.Contains(errorMsg, "not supported") {
		enhancedMsg := fmt.Sprintf("%s\n\nSupported protocols: redis, http, kafka\n", errorMsg)
		enhancedMsg += "Use --help to see available commands.\n"
		return fmt.Errorf(enhancedMsg)
	}
	
	// 检查是否为配置错误
	if strings.Contains(errorMsg, "config") {
		enhancedMsg := fmt.Sprintf("%s\n\nPlease check your configuration file or use the default config:\n", errorMsg)
		enhancedMsg += "  redis-runner <command> --config conf/<protocol>.yaml\n"
		return fmt.Errorf(enhancedMsg)
	}
	
	// 返回原始错误（如果没有特殊处理）
	return err
}

// GetSuggestions 获取建议
func (h *smartErrorHandler) GetSuggestions(command string) []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	// 首先检查是否有特定的建议规则
	if suggestions, exists := h.suggestionRules[command]; exists {
		return suggestions
	}
	
	// 基于相似度生成建议
	var suggestions []string
	
	// 检查常见的拼写错误
	commonCommands := []string{
		"redis", "redis-enhanced",
		"http", "http-enhanced", 
		"kafka", "kafka-enhanced",
		"help", "version",
	}
	
	for _, cmd := range commonCommands {
		similarity := h.calculateStringSimilarity(command, cmd)
		if similarity > 0.6 { // 相似度阈值
			suggestions = append(suggestions, cmd)
		}
	}
	
	// 按相似度排序
	sort.Slice(suggestions, func(i, j int) bool {
		sim1 := h.calculateStringSimilarity(command, suggestions[i])
		sim2 := h.calculateStringSimilarity(command, suggestions[j])
		return sim1 > sim2
	})
	
	// 限制建议数量
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	
	return suggestions
}

// AddSuggestionRule 添加建议规则
func (h *smartErrorHandler) AddSuggestionRule(pattern string, suggestions []string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.suggestionRules[pattern] = suggestions
	log.Printf("Added suggestion rule for pattern '%s': %v", pattern, suggestions)
	return nil
}

// setupDefaultRules 设置默认建议规则
func (h *smartErrorHandler) setupDefaultRules() {
	defaultRules := map[string][]string{
		"rds":    {"redis", "redis-enhanced"},
		"redis-test": {"redis-enhanced"},
		"reds":   {"redis", "redis-enhanced"},
		"htp":    {"http", "http-enhanced"},
		"htpp":   {"http", "http-enhanced"},
		"web":    {"http", "http-enhanced"},
		"kfka":   {"kafka", "kafka-enhanced"},
		"kafka-test": {"kafka-enhanced"},
		"mq":     {"kafka", "kafka-enhanced"},
		"h":      {"help"},
		"v":      {"version"},
		"exit":   {"Use Ctrl+C to exit"},
		"quit":   {"Use Ctrl+C to exit"},
	}
	
	for pattern, suggestions := range defaultRules {
		h.suggestionRules[pattern] = suggestions
	}
}

// calculateStringSimilarity 计算字符串相似度 (使用编辑距离)
func (h *smartErrorHandler) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}
	
	// 使用编辑距离算法
	distance := h.levenshteinDistance(strings.ToLower(s1), strings.ToLower(s2))
	maxLen := float64(max(len(s1), len(s2)))
	
	return 1.0 - (float64(distance) / maxLen)
}

// levenshteinDistance 计算编辑距离
func (h *smartErrorHandler) levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	
	// 创建二维数组
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// 初始化第一行和第一列
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = minThree(
				matrix[i-1][j]+1,      // 删除
				matrix[i][j-1]+1,      // 插入
				matrix[i-1][j-1]+cost, // 替换
			)
		}
	}
	
	return matrix[len1][len2]
}

// commandAnalyzer 命令分析器实现
type commandAnalyzer struct {
	knownCommands       []string
	similarityThreshold float64
	mutex              sync.RWMutex
}

// NewCommandAnalyzer 创建命令分析器
func NewCommandAnalyzer() CommandAnalyzer {
	analyzer := &commandAnalyzer{
		similarityThreshold: 0.6,
		mutex:              sync.RWMutex{},
	}
	
	// 初始化已知命令列表
	analyzer.initializeKnownCommands()
	
	return analyzer
}

// AnalyzeCommand 分析命令
func (c *commandAnalyzer) AnalyzeCommand(input string) *CommandAnalysis {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	analysis := &CommandAnalysis{
		IsKnown:     c.IsKnownCommand(input),
		Suggestions: []string{},
		BestMatch:   "",
		Confidence:  0.0,
		Protocol:    "",
		Deprecated:  false,
		Migration:   "",
	}
	
	if analysis.IsKnown {
		analysis.BestMatch = input
		analysis.Confidence = 1.0
		analysis.Protocol = c.extractProtocol(input)
		analysis.Deprecated = c.isDeprecated(input)
		if analysis.Deprecated {
			analysis.Migration = c.getMigrationPath(input)
		}
	} else {
		// 查找相似命令
		analysis.Suggestions = c.FindSimilarCommands(input, c.similarityThreshold)
		if len(analysis.Suggestions) > 0 {
			analysis.BestMatch = analysis.Suggestions[0]
			analysis.Confidence = c.CalculateSimilarity(input, analysis.BestMatch)
		}
	}
	
	return analysis
}

// FindSimilarCommands 查找相似命令
func (c *commandAnalyzer) FindSimilarCommands(input string, threshold float64) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	type commandSimilarity struct {
		command    string
		similarity float64
	}
	
	var candidates []commandSimilarity
	
	for _, cmd := range c.knownCommands {
		similarity := c.CalculateSimilarity(input, cmd)
		if similarity >= threshold {
			candidates = append(candidates, commandSimilarity{
				command:    cmd,
				similarity: similarity,
			})
		}
	}
	
	// 按相似度降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].similarity > candidates[j].similarity
	})
	
	// 提取命令名称
	var result []string
	for _, candidate := range candidates {
		result = append(result, candidate.command)
	}
	
	// 限制结果数量
	if len(result) > 5 {
		result = result[:5]
	}
	
	return result
}

// CalculateSimilarity 计算相似度
func (c *commandAnalyzer) CalculateSimilarity(input string, target string) float64 {
	if input == target {
		return 1.0
	}
	
	// 使用编辑距离计算相似度
	distance := c.levenshteinDistance(strings.ToLower(input), strings.ToLower(target))
	maxLen := float64(max(len(input), len(target)))
	
	if maxLen == 0 {
		return 0.0
	}
	
	return 1.0 - (float64(distance) / maxLen)
}

// IsKnownCommand 检查是否为已知命令
func (c *commandAnalyzer) IsKnownCommand(command string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, known := range c.knownCommands {
		if known == command {
			return true
		}
	}
	
	return false
}

// initializeKnownCommands 初始化已知命令列表
func (c *commandAnalyzer) initializeKnownCommands() {
	c.knownCommands = []string{
		// 增强版命令
		"redis-enhanced",
		"http-enhanced",
		"kafka-enhanced",
		
		// 传统命令（已弃用）
		"redis",
		"http", 
		"kafka",
		
		// 系统命令
		"help",
		"version",
		
		// 短别名
		"r",
		"h",
		"k",
	}
}

// extractProtocol 提取协议名称
func (c *commandAnalyzer) extractProtocol(command string) string {
	if strings.Contains(command, "redis") {
		return "redis"
	}
	if strings.Contains(command, "http") {
		return "http"
	}
	if strings.Contains(command, "kafka") {
		return "kafka"
	}
	
	// 别名映射
	switch command {
	case "r":
		return "redis"
	case "h":
		return "http"
	case "k":
		return "kafka"
	}
	
	return ""
}

// isDeprecated 检查命令是否已弃用
func (c *commandAnalyzer) isDeprecated(command string) bool {
	deprecatedCommands := []string{"redis", "http", "kafka"}
	for _, deprecated := range deprecatedCommands {
		if command == deprecated {
			return true
		}
	}
	return false
}

// getMigrationPath 获取迁移路径
func (c *commandAnalyzer) getMigrationPath(command string) string {
	migrationMap := map[string]string{
		"redis": "Use 'redis-enhanced' instead",
		"http":  "Use 'http-enhanced' instead", 
		"kafka": "Use 'kafka-enhanced' instead",
	}
	
	if migration, exists := migrationMap[command]; exists {
		return migration
	}
	
	return ""
}

// levenshteinDistance 计算编辑距离（复用错误处理器中的实现）
func (c *commandAnalyzer) levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = minThree(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}
	
	return matrix[len1][len2]
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minThree(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}