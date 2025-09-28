package commands

import "abc-runner/app/core/interfaces"

// countSuccessful 统计成功操作数
func countSuccessful(results []*interfaces.OperationResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

// looksLikeHostname 检查字符串是否像主机名
func looksLikeHostname(s string) bool {
	if s == "" {
		return false
	}
	
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '.' || char == '-' || char == '_') {
			return false
		}
	}
	
	return true
}