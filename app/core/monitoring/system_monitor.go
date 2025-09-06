package monitoring

import (
	"runtime"
	"sync"
	"time"
)

// SystemMonitor 系统资源监控器
type SystemMonitor struct {
	cpuUsage    float64
	memUsage    *MemoryUsage
	goRoutines  int
	gcStats     *GCStats
	mutex       sync.RWMutex
	stopChan    chan struct{}
	isRunning   bool
}

// MemoryUsage 内存使用情况
type MemoryUsage struct {
	Alloc      uint64  // 当前分配的内存
	TotalAlloc uint64  // 累计分配的内存
	Sys        uint64  // 系统内存
	NumGC      uint32  // GC次数
	HeapAlloc  uint64  // 堆内存分配
	HeapSys    uint64  // 堆系统内存
	HeapIdle   uint64  // 堆空闲内存
	HeapInuse  uint64  // 堆使用中内存
	StackInuse uint64  // 栈使用中内存
	StackSys   uint64  // 栈系统内存
}

// GCStats GC统计信息
type GCStats struct {
	NumGC      uint32        // GC次数
	PauseTotal time.Duration // 总暂停时间
	PauseNs    []uint64      // 暂停时间数组
	LastGC     time.Time     // 最后一次GC时间
}

// ConnectionPoolStats 连接池统计
type ConnectionPoolStats struct {
	TotalConns uint32
	IdleConns  uint32
	StaleConns uint32
	Hits       uint32
	Misses     uint32
	Timeouts   uint32
}

// NewSystemMonitor 创建系统监控器
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		memUsage: &MemoryUsage{},
		gcStats:  &GCStats{},
		stopChan: make(chan struct{}),
	}
}

// Start 启动监控
func (sm *SystemMonitor) Start(interval time.Duration) {
	sm.mutex.Lock()
	if sm.isRunning {
		sm.mutex.Unlock()
		return
	}
	sm.isRunning = true
	sm.mutex.Unlock()

	go sm.monitorLoop(interval)
}

// Stop 停止监控
func (sm *SystemMonitor) Stop() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if !sm.isRunning {
		return
	}
	
	sm.isRunning = false
	close(sm.stopChan)
}

// monitorLoop 监控循环
func (sm *SystemMonitor) monitorLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.collectMetrics()
		}
	}
}

// collectMetrics 收集指标
func (sm *SystemMonitor) collectMetrics() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 收集内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	sm.memUsage.Alloc = memStats.Alloc
	sm.memUsage.TotalAlloc = memStats.TotalAlloc
	sm.memUsage.Sys = memStats.Sys
	sm.memUsage.NumGC = memStats.NumGC
	sm.memUsage.HeapAlloc = memStats.HeapAlloc
	sm.memUsage.HeapSys = memStats.HeapSys
	sm.memUsage.HeapIdle = memStats.HeapIdle
	sm.memUsage.HeapInuse = memStats.HeapInuse
	sm.memUsage.StackInuse = memStats.StackInuse
	sm.memUsage.StackSys = memStats.StackSys

	// 收集GC统计
	sm.gcStats.NumGC = memStats.NumGC
	sm.gcStats.PauseTotal = time.Duration(memStats.PauseTotalNs)
	if len(memStats.PauseNs) > 0 {
		sm.gcStats.PauseNs = memStats.PauseNs[:]
	}
	if memStats.LastGC > 0 {
		sm.gcStats.LastGC = time.Unix(0, int64(memStats.LastGC))
	}

	// 收集goroutine数量
	sm.goRoutines = runtime.NumGoroutine()
}

// GetMemoryUsage 获取内存使用情况
func (sm *SystemMonitor) GetMemoryUsage() *MemoryUsage {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// 返回副本
	return &MemoryUsage{
		Alloc:      sm.memUsage.Alloc,
		TotalAlloc: sm.memUsage.TotalAlloc,
		Sys:        sm.memUsage.Sys,
		NumGC:      sm.memUsage.NumGC,
		HeapAlloc:  sm.memUsage.HeapAlloc,
		HeapSys:    sm.memUsage.HeapSys,
		HeapIdle:   sm.memUsage.HeapIdle,
		HeapInuse:  sm.memUsage.HeapInuse,
		StackInuse: sm.memUsage.StackInuse,
		StackSys:   sm.memUsage.StackSys,
	}
}

// GetGCStats 获取GC统计
func (sm *SystemMonitor) GetGCStats() *GCStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	pauseNs := make([]uint64, len(sm.gcStats.PauseNs))
	copy(pauseNs, sm.gcStats.PauseNs)
	
	return &GCStats{
		NumGC:      sm.gcStats.NumGC,
		PauseTotal: sm.gcStats.PauseTotal,
		PauseNs:    pauseNs,
		LastGC:     sm.gcStats.LastGC,
	}
}

// GetGoRoutineCount 获取goroutine数量
func (sm *SystemMonitor) GetGoRoutineCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.goRoutines
}

// GetSystemSnapshot 获取系统快照
func (sm *SystemMonitor) GetSystemSnapshot() *SystemSnapshot {
	return &SystemSnapshot{
		Timestamp:   time.Now(),
		Memory:      sm.GetMemoryUsage(),
		GC:          sm.GetGCStats(),
		GoRoutines:  sm.GetGoRoutineCount(),
	}
}

// SystemSnapshot 系统快照
type SystemSnapshot struct {
	Timestamp  time.Time    `json:"timestamp"`
	Memory     *MemoryUsage `json:"memory"`
	GC         *GCStats     `json:"gc"`
	GoRoutines int          `json:"goroutines"`
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	systemMonitor   *SystemMonitor
	snapshots       []*SystemSnapshot
	maxSnapshots    int
	mutex           sync.RWMutex
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(maxSnapshots int) *PerformanceMonitor {
	if maxSnapshots <= 0 {
		maxSnapshots = 100 // 默认保留100个快照
	}
	
	return &PerformanceMonitor{
		systemMonitor: NewSystemMonitor(),
		snapshots:     make([]*SystemSnapshot, 0, maxSnapshots),
		maxSnapshots:  maxSnapshots,
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start(interval time.Duration) {
	pm.systemMonitor.Start(interval)
	
	// 启动快照收集
	go pm.snapshotLoop(interval * 10) // 每10个间隔收集一次快照
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.systemMonitor.Stop()
}

// snapshotLoop 快照收集循环
func (pm *PerformanceMonitor) snapshotLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.addSnapshot(pm.systemMonitor.GetSystemSnapshot())
		}
	}
}

// addSnapshot 添加快照
func (pm *PerformanceMonitor) addSnapshot(snapshot *SystemSnapshot) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.snapshots = append(pm.snapshots, snapshot)
	
	// 如果超过最大数量，移除最老的快照
	if len(pm.snapshots) > pm.maxSnapshots {
		pm.snapshots = pm.snapshots[1:]
	}
}

// GetSnapshots 获取快照列表
func (pm *PerformanceMonitor) GetSnapshots() []*SystemSnapshot {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	snapshots := make([]*SystemSnapshot, len(pm.snapshots))
	copy(snapshots, pm.snapshots)
	return snapshots
}

// GetLatestSnapshot 获取最新快照
func (pm *PerformanceMonitor) GetLatestSnapshot() *SystemSnapshot {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	if len(pm.snapshots) == 0 {
		return pm.systemMonitor.GetSystemSnapshot()
	}
	
	return pm.snapshots[len(pm.snapshots)-1]
}

// GetSummary 获取性能摘要
func (pm *PerformanceMonitor) GetSummary() *PerformanceSummary {
	snapshots := pm.GetSnapshots()
	if len(snapshots) == 0 {
		return &PerformanceSummary{}
	}
	
	summary := &PerformanceSummary{
		Duration:    snapshots[len(snapshots)-1].Timestamp.Sub(snapshots[0].Timestamp),
		Snapshots:   len(snapshots),
	}
	
	// 计算平均值和峰值
	var totalMem, maxMem uint64
	var totalGoroutines, maxGoroutines int
	var totalGC uint32
	
	for _, snapshot := range snapshots {
		totalMem += snapshot.Memory.Alloc
		if snapshot.Memory.Alloc > maxMem {
			maxMem = snapshot.Memory.Alloc
		}
		
		totalGoroutines += snapshot.GoRoutines
		if snapshot.GoRoutines > maxGoroutines {
			maxGoroutines = snapshot.GoRoutines
		}
		
		if snapshot.GC.NumGC > totalGC {
			totalGC = snapshot.GC.NumGC
		}
	}
	
	summary.AvgMemoryUsage = totalMem / uint64(len(snapshots))
	summary.MaxMemoryUsage = maxMem
	summary.AvgGoRoutines = totalGoroutines / len(snapshots)
	summary.MaxGoRoutines = maxGoroutines
	summary.TotalGC = totalGC
	
	return summary
}

// PerformanceSummary 性能摘要
type PerformanceSummary struct {
	Duration        time.Duration `json:"duration"`
	Snapshots       int           `json:"snapshots"`
	AvgMemoryUsage  uint64        `json:"avg_memory_usage"`
	MaxMemoryUsage  uint64        `json:"max_memory_usage"`
	AvgGoRoutines   int           `json:"avg_goroutines"`
	MaxGoRoutines   int           `json:"max_goroutines"`
	TotalGC         uint32        `json:"total_gc"`
}