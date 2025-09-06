package unified

import (
	"fmt"
	"log"
	"sync"
)

// aliasManager 别名管理器实现
type aliasManager struct {
	aliases map[string]string
	mutex   sync.RWMutex
}

// NewAliasManager 创建别名管理器
func NewAliasManager() AliasManager {
	return &aliasManager{
		aliases: make(map[string]string),
		mutex:   sync.RWMutex{},
	}
}

// AddAlias 添加别名
func (a *aliasManager) AddAlias(alias string, target string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 检查别名是否已存在
	if existingTarget, exists := a.aliases[alias]; exists {
		if existingTarget == target {
			// 相同的映射，不需要重复添加
			return nil
		}
		return fmt.Errorf("alias '%s' already exists, mapped to '%s'", alias, existingTarget)
	}

	// 防止循环引用
	if a.wouldCreateCycle(alias, target) {
		return fmt.Errorf("adding alias '%s' -> '%s' would create a cycle", alias, target)
	}

	a.aliases[alias] = target
	log.Printf("Added alias: %s -> %s", alias, target)
	return nil
}

// RemoveAlias 移除别名
func (a *aliasManager) RemoveAlias(alias string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.aliases[alias]; !exists {
		return fmt.Errorf("alias '%s' does not exist", alias)
	}

	delete(a.aliases, alias)
	log.Printf("Removed alias: %s", alias)
	return nil
}

// ResolveAlias 解析别名
func (a *aliasManager) ResolveAlias(alias string) (string, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// 跟踪已访问的别名以防止循环
	visited := make(map[string]bool)
	current := alias

	for {
		if visited[current] {
			// 检测到循环，返回原始别名
			log.Printf("Cycle detected in alias resolution for '%s'", alias)
			return alias, false
		}

		visited[current] = true

		if target, exists := a.aliases[current]; exists {
			current = target
		} else {
			// 没有找到进一步的别名，返回当前值
			return current, current != alias
		}

		// 防止无限循环（最多解析10层）
		if len(visited) > 10 {
			log.Printf("Too many alias resolutions for '%s', stopping", alias)
			return current, current != alias
		}
	}
}

// ListAliases 列出所有别名
func (a *aliasManager) ListAliases() map[string]string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	result := make(map[string]string)
	for alias, target := range a.aliases {
		result[alias] = target
	}

	return result
}

// IsAlias 检查是否为别名
func (a *aliasManager) IsAlias(command string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	_, exists := a.aliases[command]
	return exists
}

// LoadAliases 从配置加载别名
func (a *aliasManager) LoadAliases(aliases map[string]string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 验证所有别名是否会产生循环
	tempAliases := make(map[string]string)
	for k, v := range a.aliases {
		tempAliases[k] = v
	}

	for alias, target := range aliases {
		tempAliases[alias] = target
		if a.wouldCreateCycleInMap(tempAliases, alias, target) {
			return fmt.Errorf("loading aliases would create a cycle with '%s' -> '%s'", alias, target)
		}
	}

	// 如果没有循环，则添加所有别名
	for alias, target := range aliases {
		a.aliases[alias] = target
		log.Printf("Loaded alias: %s -> %s", alias, target)
	}

	return nil
}

// wouldCreateCycle 检查添加别名是否会创建循环
func (a *aliasManager) wouldCreateCycle(alias string, target string) bool {
	// 检查target是否最终解析回alias
	visited := make(map[string]bool)
	current := target

	for {
		if current == alias {
			return true
		}

		if visited[current] {
			// 检测到其他循环，但不涉及新的alias
			return false
		}

		visited[current] = true

		if nextTarget, exists := a.aliases[current]; exists {
			current = nextTarget
		} else {
			return false
		}

		// 防止无限循环
		if len(visited) > 10 {
			return false
		}
	}
}

// wouldCreateCycleInMap 检查在给定映射中是否会创建循环
func (a *aliasManager) wouldCreateCycleInMap(aliasMap map[string]string, alias string, target string) bool {
	visited := make(map[string]bool)
	current := target

	for {
		if current == alias {
			return true
		}

		if visited[current] {
			return false
		}

		visited[current] = true

		if nextTarget, exists := aliasMap[current]; exists {
			current = nextTarget
		} else {
			return false
		}

		if len(visited) > 10 {
			return false
		}
	}
}
