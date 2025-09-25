package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"abc-runner/app/core/metrics"
)

// SystemMonitor 系统监控器
type SystemMonitor struct {
	config       *metrics.SystemConfig
	isRunning    bool
	stopChan     chan struct{}
	memStats     *runtime.MemStats
	snapshots    *metrics.RingBuffer[SystemSnapshot]
	listeners    []SystemListener
	mutex        sync.RWMutex
	lastSnapshot *SystemSnapshot
}

// SystemSnapshot 系统快照
type SystemSnapshot struct {
	Timestamp   time.Time              `json:"timestamp"`
	Memory      *MemorySnapshot        `json:"memory"`
	GC          *GCSnapshot            `json:"gc"`
	Goroutines  *GoroutineSnapshot     `json:"goroutines"`
	CPU         *CPUSnapshot           `json:"cpu"`
	Health      *HealthSnapshot        `json:"health"`
}

// MemorySnapshot 内存快照
type MemorySnapshot struct {
	Allocated    uint64  `json:"allocated"`     // 当前分配的内存
	TotalAlloc   uint64  `json:"total_alloc"`   // 累计分配的内存
	Sys          uint64  `json:"sys"`           // 从系统获得的内存
	HeapAlloc    uint64  `json:"heap_alloc"`    // 堆上分配的内存
	HeapSys      uint64  `json:"heap_sys"`      // 堆内存系统分配
	HeapInuse    uint64  `json:"heap_inuse"`    // 堆内存使用中
	StackInuse   uint64  `json:"stack_inuse"`   // 栈内存使用中
	UsagePercent float64 `json:"usage_percent"` // 内存使用百分比
}

// GCSnapshot GC快照
type GCSnapshot struct {
	NumGC         uint32        `json:"num_gc"`          // GC次数
	PauseTotal    time.Duration `json:"pause_total"`     // 总暂停时间
	PauseAvg      time.Duration `json:"pause_avg"`       // 平均暂停时间
	LastPause     time.Duration `json:"last_pause"`      // 最后一次暂停时间
	NextGC        uint64        `json:"next_gc"`         // 下次GC阈值
	GCCPUFraction float64       `json:"gc_cpu_fraction"` // GC CPU占用比例
}

// GoroutineSnapshot 协程快照
type GoroutineSnapshot struct {
	Count    int     `json:"count"`     // 当前协程数
	Peak     int     `json:"peak"`      // 峰值协程数
	Growth   int     `json:"growth"`    // 增长数
	GrowthRate float64 `json:"growth_rate"` // 增长率
}

// CPUSnapshot CPU快照
type CPUSnapshot struct {
	NumCPU    int     `json:"num_cpu"`    // CPU核心数
	NumCgoCall int64  `json:"num_cgo"`    // CGO调用次数
	Usage     float64 `json:"usage"`      // CPU使用率估算
}

// HealthSnapshot 健康快照
type HealthSnapshot struct {
	Status      string             `json:"status"`      // healthy/warning/critical
	Score       float64            `json:"score"`       // 健康分数 0-100
	Issues      []HealthIssue      `json:"issues"`      // 健康问题
	Trends      map[string]float64 `json:"trends"`      // 趋势数据
}

// HealthIssue 健康问题
type HealthIssue struct {
	Type        string    `json:"type"`        // 问题类型
	Severity    string    `json:"severity"`    // 严重程度
	Message     string    `json:"message"`     // 问题描述
	Metric      string    `json:"metric"`      // 相关指标
	Threshold   float64   `json:"threshold"`   // 阈值
	Current     float64   `json:"current"`     // 当前值
	FirstSeen   time.Time `json:"first_seen"`  // 首次发现时间
}

// SystemListener 系统监控监听器
type SystemListener interface {
	OnSystemSnapshot(snapshot *SystemSnapshot)
	OnHealthIssue(issue *HealthIssue)
	OnThresholdViolation(metric string, current, threshold float64)
}

// NewSystemMonitor 创建系统监控器
func NewSystemMonitor(config *metrics.SystemConfig) *SystemMonitor {
	if config == nil {
		config = &metrics.SystemConfig{
			MonitorInterval:   time.Second,
			SnapshotRetention: 100,
			Enabled:           true,
		}
	}

	return &SystemMonitor{
		config:    config,
		stopChan:  make(chan struct{}),
		memStats:  &runtime.MemStats{},
		snapshots: metrics.NewRingBuffer[SystemSnapshot](config.SnapshotRetention),
		listeners: make([]SystemListener, 0),
	}
}

// Start 启动监控
func (sm *SystemMonitor) Start(ctx context.Context) error {
	if !sm.config.Enabled {
		return nil
	}

	sm.mutex.Lock()
	if sm.isRunning {
		sm.mutex.Unlock()
		return nil
	}
	sm.isRunning = true
	sm.mutex.Unlock()

	go sm.monitorLoop(ctx)
	return nil
}

// Stop 停止监控
func (sm *SystemMonitor) Stop() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if !sm.isRunning {
		return nil
	}

	close(sm.stopChan)
	sm.isRunning = false
	return nil
}

// AddListener 添加监听器
func (sm *SystemMonitor) AddListener(listener SystemListener) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.listeners = append(sm.listeners, listener)
}

// GetLatestSnapshot 获取最新快照
func (sm *SystemMonitor) GetLatestSnapshot() *SystemSnapshot {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.lastSnapshot
}

// GetSnapshots 获取历史快照
func (sm *SystemMonitor) GetSnapshots() []SystemSnapshot {
	return sm.snapshots.ToSlice()
}

// GetMetrics 获取系统指标
func (sm *SystemMonitor) GetMetrics() metrics.SystemMetrics {
	if sm.lastSnapshot == nil {
		return metrics.SystemMetrics{}
	}

	return metrics.SystemMetrics{
		Memory: metrics.MemoryMetrics{
			Allocated:  sm.lastSnapshot.Memory.Allocated,
			TotalAlloc: sm.lastSnapshot.Memory.TotalAlloc,
			Sys:        sm.lastSnapshot.Memory.Sys,
			NumGC:      uint32(sm.lastSnapshot.GC.NumGC),
			Usage:      sm.lastSnapshot.Memory.UsagePercent,
		},
		GC: metrics.GCMetrics{
			NumGC:      sm.lastSnapshot.GC.NumGC,
			PauseTotal: sm.lastSnapshot.GC.PauseTotal,
			PauseAvg:   sm.lastSnapshot.GC.PauseAvg,
			LastPause:  sm.lastSnapshot.GC.LastPause,
		},
		Goroutine: metrics.GoroutineMetrics{
			Active: sm.lastSnapshot.Goroutines.Count,
			Peak:   sm.lastSnapshot.Goroutines.Peak,
		},
		CPU: metrics.CPUMetrics{
			Usage: sm.lastSnapshot.CPU.Usage,
			Cores: sm.lastSnapshot.CPU.NumCPU,
		},
	}
}

// monitorLoop 监控循环
func (sm *SystemMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(sm.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			snapshot := sm.collectSnapshot()
			sm.processSnapshot(snapshot)
		}
	}
}

// collectSnapshot 收集系统快照
func (sm *SystemMonitor) collectSnapshot() *SystemSnapshot {
	// 收集内存统计
	runtime.ReadMemStats(sm.memStats)
	
	// 收集协程信息
	numGoroutines := runtime.NumGoroutine()
	
	// 创建快照
	snapshot := &SystemSnapshot{
		Timestamp: time.Now(),
		Memory: &MemorySnapshot{
			Allocated:    sm.memStats.Alloc,
			TotalAlloc:   sm.memStats.TotalAlloc,
			Sys:          sm.memStats.Sys,
			HeapAlloc:    sm.memStats.HeapAlloc,
			HeapSys:      sm.memStats.HeapSys,
			HeapInuse:    sm.memStats.HeapInuse,
			StackInuse:   sm.memStats.StackInuse,
			UsagePercent: float64(sm.memStats.Alloc) / float64(sm.memStats.Sys) * 100,
		},
		GC: &GCSnapshot{
			NumGC:         sm.memStats.NumGC,
			PauseTotal:    time.Duration(sm.memStats.PauseTotalNs),
			NextGC:        sm.memStats.NextGC,
			GCCPUFraction: sm.memStats.GCCPUFraction,
		},
		Goroutines: &GoroutineSnapshot{
			Count: numGoroutines,
		},
		CPU: &CPUSnapshot{
			NumCPU:     runtime.NumCPU(),
			NumCgoCall: runtime.NumCgoCall(),
		},
	}

	// 计算GC指标
	if sm.memStats.NumGC > 0 {
		snapshot.GC.PauseAvg = time.Duration(sm.memStats.PauseTotalNs / uint64(sm.memStats.NumGC))
		if sm.memStats.NumGC > 0 {
			snapshot.GC.LastPause = time.Duration(sm.memStats.PauseNs[(sm.memStats.NumGC+255)%256])
		}
	}

	// 计算协程指标
	if sm.lastSnapshot != nil {
		snapshot.Goroutines.Peak = sm.lastSnapshot.Goroutines.Peak
		if numGoroutines > snapshot.Goroutines.Peak {
			snapshot.Goroutines.Peak = numGoroutines
		}
		snapshot.Goroutines.Growth = numGoroutines - sm.lastSnapshot.Goroutines.Count
		if sm.lastSnapshot.Goroutines.Count > 0 {
			snapshot.Goroutines.GrowthRate = float64(snapshot.Goroutines.Growth) / float64(sm.lastSnapshot.Goroutines.Count) * 100
		}
	} else {
		snapshot.Goroutines.Peak = numGoroutines
	}

	// 估算CPU使用率
	snapshot.CPU.Usage = sm.estimateCPUUsage()

	// 评估健康状况
	snapshot.Health = sm.assessHealth(snapshot)

	return snapshot
}

// processSnapshot 处理快照
func (sm *SystemMonitor) processSnapshot(snapshot *SystemSnapshot) {
	// 存储快照
	sm.snapshots.Push(*snapshot)
	
	sm.mutex.Lock()
	sm.lastSnapshot = snapshot
	sm.mutex.Unlock()

	// 通知监听器
	for _, listener := range sm.listeners {
		listener.OnSystemSnapshot(snapshot)
		
		// 检查健康问题
		for _, issue := range snapshot.Health.Issues {
			listener.OnHealthIssue(&issue)
		}
	}
}

// estimateCPUUsage 估算CPU使用率
func (sm *SystemMonitor) estimateCPUUsage() float64 {
	// 简化的CPU使用率估算
	// 实际项目中可能需要更精确的实现
	
	if sm.lastSnapshot == nil {
		return 0
	}

	// 基于协程数量和GC时间的简单估算
	goroutineUsage := float64(sm.lastSnapshot.Goroutines.Count) / float64(runtime.NumCPU()) * 10
	if goroutineUsage > 100 {
		goroutineUsage = 100
	}

	gcUsage := sm.memStats.GCCPUFraction * 100

	return (goroutineUsage + gcUsage) / 2
}

// assessHealth 评估健康状况
func (sm *SystemMonitor) assessHealth(snapshot *SystemSnapshot) *HealthSnapshot {
	health := &HealthSnapshot{
		Status: "healthy",
		Score:  100.0,
		Issues: make([]HealthIssue, 0),
		Trends: make(map[string]float64),
	}

	// 检查内存使用率
	if snapshot.Memory.UsagePercent > sm.config.HealthThresholds.MemoryUsage {
		severity := "warning"
		if snapshot.Memory.UsagePercent > sm.config.HealthThresholds.MemoryUsage*1.2 {
			severity = "critical"
			health.Status = "critical"
		} else if health.Status == "healthy" {
			health.Status = "warning"
		}

		health.Issues = append(health.Issues, HealthIssue{
			Type:      "memory",
			Severity:  severity,
			Message:   fmt.Sprintf("Memory usage %.1f%% exceeds threshold %.1f%%", snapshot.Memory.UsagePercent, sm.config.HealthThresholds.MemoryUsage),
			Metric:    "memory_usage",
			Threshold: sm.config.HealthThresholds.MemoryUsage,
			Current:   snapshot.Memory.UsagePercent,
			FirstSeen: snapshot.Timestamp,
		})

		health.Score -= 20
	}

	// 检查协程数量
	if float64(snapshot.Goroutines.Count) > float64(sm.config.HealthThresholds.GoroutineCount) {
		severity := "warning"
		if float64(snapshot.Goroutines.Count) > float64(sm.config.HealthThresholds.GoroutineCount)*1.5 {
			severity = "critical"
			health.Status = "critical"
		} else if health.Status == "healthy" {
			health.Status = "warning"
		}

		health.Issues = append(health.Issues, HealthIssue{
			Type:      "goroutines",
			Severity:  severity,
			Message:   fmt.Sprintf("Goroutine count %d exceeds threshold %d", snapshot.Goroutines.Count, sm.config.HealthThresholds.GoroutineCount),
			Metric:    "goroutine_count",
			Threshold: float64(sm.config.HealthThresholds.GoroutineCount),
			Current:   float64(snapshot.Goroutines.Count),
			FirstSeen: snapshot.Timestamp,
		})

		health.Score -= 15
	}

	// 检查GC频率
	if snapshot.GC.NumGC > sm.config.HealthThresholds.GCFrequency {
		health.Issues = append(health.Issues, HealthIssue{
			Type:      "gc",
			Severity:  "warning",
			Message:   fmt.Sprintf("GC frequency %d exceeds threshold %d", snapshot.GC.NumGC, sm.config.HealthThresholds.GCFrequency),
			Metric:    "gc_frequency",
			Threshold: float64(sm.config.HealthThresholds.GCFrequency),
			Current:   float64(snapshot.GC.NumGC),
			FirstSeen: snapshot.Timestamp,
		})

		if health.Status == "healthy" {
			health.Status = "warning"
		}
		health.Score -= 10
	}

	// 确保分数不低于0
	if health.Score < 0 {
		health.Score = 0
	}

	return health
}

// Reset 重置监控数据
func (sm *SystemMonitor) Reset() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.snapshots.Clear()
	sm.lastSnapshot = nil
}