package http

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// HttpStatusCodeTracker HTTP状态码追踪器
type HttpStatusCodeTracker struct {
	stats     map[int]*statusCodeData
	totalReqs int64
	mutex     sync.RWMutex
}

type statusCodeData struct {
	count        int64
	successCount int64
	totalLatency int64
	minLatency   int64
	maxLatency   int64
}

func NewHttpStatusCodeTracker() *HttpStatusCodeTracker {
	return &HttpStatusCodeTracker{
		stats: make(map[int]*statusCodeData),
	}
}

func (hst *HttpStatusCodeTracker) Record(statusCode int, result *interfaces.OperationResult) {
	hst.mutex.Lock()
	defer hst.mutex.Unlock()

	data, exists := hst.stats[statusCode]
	if !exists {
		data = &statusCodeData{
			minLatency: math.MaxInt64,
		}
		hst.stats[statusCode] = data
	}

	atomic.AddInt64(&data.count, 1)
	atomic.AddInt64(&hst.totalReqs, 1)

	if statusCode >= 200 && statusCode < 300 {
		atomic.AddInt64(&data.successCount, 1)
	}

	latencyNanos := result.Duration.Nanoseconds()
	atomic.AddInt64(&data.totalLatency, latencyNanos)

	// 更新延迟范围
	for {
		current := atomic.LoadInt64(&data.minLatency)
		if latencyNanos >= current || atomic.CompareAndSwapInt64(&data.minLatency, current, latencyNanos) {
			break
		}
	}
	for {
		current := atomic.LoadInt64(&data.maxLatency)
		if latencyNanos <= current || atomic.CompareAndSwapInt64(&data.maxLatency, current, latencyNanos) {
			break
		}
	}
}

func (hst *HttpStatusCodeTracker) GetStats() map[int]*HttpStatusCodeStat {
	hst.mutex.RLock()
	defer hst.mutex.RUnlock()

	stats := make(map[int]*HttpStatusCodeStat)
	totalReqs := atomic.LoadInt64(&hst.totalReqs)

	for code, data := range hst.stats {
		count := atomic.LoadInt64(&data.count)
		successCount := atomic.LoadInt64(&data.successCount)
		totalLatency := atomic.LoadInt64(&data.totalLatency)
		minLatency := atomic.LoadInt64(&data.minLatency)
		maxLatency := atomic.LoadInt64(&data.maxLatency)

		stat := &HttpStatusCodeStat{
			Count:        count,
			SuccessCount: successCount,
			TotalLatency: time.Duration(totalLatency),
			MinLatency:   time.Duration(minLatency),
			MaxLatency:   time.Duration(maxLatency),
		}

		if count > 0 {
			stat.AvgLatency = time.Duration(totalLatency / count)
		}
		if totalReqs > 0 {
			stat.Percentage = float64(count) / float64(totalReqs) * 100.0
		}

		stats[code] = stat
	}

	return stats
}

func (hst *HttpStatusCodeTracker) Reset() {
	hst.mutex.Lock()
	defer hst.mutex.Unlock()
	hst.stats = make(map[int]*statusCodeData)
	atomic.StoreInt64(&hst.totalReqs, 0)
}

// HttpMethodTracker HTTP方法追踪器
type HttpMethodTracker struct {
	stats     map[string]*methodData
	startTime time.Time
	mutex     sync.RWMutex
}

type methodData struct {
	count        int64
	successCount int64
	failureCount int64
	totalLatency int64
	minLatency   int64
	maxLatency   int64
}

func NewHttpMethodTracker() *HttpMethodTracker {
	return &HttpMethodTracker{
		stats:     make(map[string]*methodData),
		startTime: time.Now(),
	}
}

func (hmt *HttpMethodTracker) Record(method string, result *interfaces.OperationResult) {
	hmt.mutex.Lock()
	defer hmt.mutex.Unlock()

	data, exists := hmt.stats[method]
	if !exists {
		data = &methodData{
			minLatency: math.MaxInt64,
		}
		hmt.stats[method] = data
	}

	atomic.AddInt64(&data.count, 1)
	if result.Success {
		atomic.AddInt64(&data.successCount, 1)
	} else {
		atomic.AddInt64(&data.failureCount, 1)
	}

	latencyNanos := result.Duration.Nanoseconds()
	atomic.AddInt64(&data.totalLatency, latencyNanos)

	for {
		current := atomic.LoadInt64(&data.minLatency)
		if latencyNanos >= current || atomic.CompareAndSwapInt64(&data.minLatency, current, latencyNanos) {
			break
		}
	}
	for {
		current := atomic.LoadInt64(&data.maxLatency)
		if latencyNanos <= current || atomic.CompareAndSwapInt64(&data.maxLatency, current, latencyNanos) {
			break
		}
	}
}

func (hmt *HttpMethodTracker) GetStats() map[string]*HttpMethodStat {
	hmt.mutex.RLock()
	defer hmt.mutex.RUnlock()

	stats := make(map[string]*HttpMethodStat)
	duration := time.Since(hmt.startTime)

	for method, data := range hmt.stats {
		count := atomic.LoadInt64(&data.count)
		successCount := atomic.LoadInt64(&data.successCount)
		failureCount := atomic.LoadInt64(&data.failureCount)
		totalLatency := atomic.LoadInt64(&data.totalLatency)

		stat := &HttpMethodStat{
			Count:        count,
			SuccessCount: successCount,
			FailureCount: failureCount,
			TotalLatency: time.Duration(totalLatency),
			MinLatency:   time.Duration(atomic.LoadInt64(&data.minLatency)),
			MaxLatency:   time.Duration(atomic.LoadInt64(&data.maxLatency)),
		}

		if count > 0 {
			stat.AvgLatency = time.Duration(totalLatency / count)
			stat.SuccessRate = float64(successCount) / float64(count) * 100.0
		}
		if duration > 0 {
			stat.Throughput = float64(count) / duration.Seconds()
		}

		stats[method] = stat
	}

	return stats
}

func (hmt *HttpMethodTracker) Reset() {
	hmt.mutex.Lock()
	defer hmt.mutex.Unlock()
	hmt.stats = make(map[string]*methodData)
	hmt.startTime = time.Now()
}

// 简化的其他追踪器实现
type HttpURLTracker struct{ mutex sync.RWMutex; stats map[string]*HttpURLStat }
type HttpContentTypeTracker struct{ mutex sync.RWMutex; stats map[string]*HttpContentTypeStat }
type HttpNetworkTracker struct{ mutex sync.RWMutex; stats *HttpNetworkStat }
type HttpConnectionTracker struct{ mutex sync.RWMutex; stats *HttpConnectionStat }
type HttpSecurityTracker struct{ mutex sync.RWMutex; stats *HttpSecurityStat }
type HttpPerformanceTracker struct{ mutex sync.RWMutex; stats *HttpPerformanceStat }

func NewHttpURLTracker() *HttpURLTracker { 
	return &HttpURLTracker{stats: make(map[string]*HttpURLStat)} 
}
func NewHttpContentTypeTracker() *HttpContentTypeTracker { 
	return &HttpContentTypeTracker{stats: make(map[string]*HttpContentTypeStat)} 
}
func NewHttpNetworkTracker() *HttpNetworkTracker { 
	return &HttpNetworkTracker{stats: &HttpNetworkStat{}} 
}
func NewHttpConnectionTracker() *HttpConnectionTracker { 
	return &HttpConnectionTracker{stats: &HttpConnectionStat{}} 
}
func NewHttpSecurityTracker() *HttpSecurityTracker { 
	return &HttpSecurityTracker{stats: &HttpSecurityStat{}} 
}
func NewHttpPerformanceTracker() *HttpPerformanceTracker { 
	return &HttpPerformanceTracker{stats: &HttpPerformanceStat{}} 
}

func (t *HttpURLTracker) Record(url string, result *interfaces.OperationResult) {}
func (t *HttpURLTracker) GetStats() map[string]*HttpURLStat { return t.stats }
func (t *HttpURLTracker) Reset() { t.stats = make(map[string]*HttpURLStat) }

func (t *HttpContentTypeTracker) Record(contentType string, result *interfaces.OperationResult) {}
func (t *HttpContentTypeTracker) GetStats() map[string]*HttpContentTypeStat { return t.stats }
func (t *HttpContentTypeTracker) Reset() { t.stats = make(map[string]*HttpContentTypeStat) }

func (t *HttpNetworkTracker) Record(result *interfaces.OperationResult) {}
func (t *HttpNetworkTracker) GetStats() *HttpNetworkStat { return t.stats }
func (t *HttpNetworkTracker) Reset() { t.stats = &HttpNetworkStat{} }

func (t *HttpConnectionTracker) Record(result *interfaces.OperationResult) {}
func (t *HttpConnectionTracker) GetStats() *HttpConnectionStat { return t.stats }
func (t *HttpConnectionTracker) Reset() { t.stats = &HttpConnectionStat{} }

func (t *HttpSecurityTracker) Record(result *interfaces.OperationResult) {}
func (t *HttpSecurityTracker) GetStats() *HttpSecurityStat { return t.stats }
func (t *HttpSecurityTracker) Reset() { t.stats = &HttpSecurityStat{} }

func (t *HttpPerformanceTracker) Record(result *interfaces.OperationResult) {}
func (t *HttpPerformanceTracker) GetStats() *HttpPerformanceStat { return t.stats }
func (t *HttpPerformanceTracker) Reset() { t.stats = &HttpPerformanceStat{} }