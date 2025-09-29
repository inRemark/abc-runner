package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// HTTPServer HTTP服务端实现
type HTTPServer struct {
	*common.BaseServer

	config     *HTTPServerConfig
	httpServer *http.Server
	mux        *http.ServeMux
	middleware []MiddlewareFunc

	// 统计信息
	requestCount int64
	mutex        sync.RWMutex
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(http.Handler) http.Handler

// NewHTTPServer 创建HTTP服务端
func NewHTTPServer(config *HTTPServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *HTTPServer {
	baseServer := common.NewBaseServer("http", config, logger, metricsCollector)

	server := &HTTPServer{
		BaseServer: baseServer,
		config:     config,
		mux:        http.NewServeMux(),
	}

	// 设置HTTP服务器
	server.httpServer = &http.Server{
		Addr:           config.GetAddress(),
		Handler:        server.buildHandler(),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// Start 启动HTTP服务端
func (hs *HTTPServer) Start(ctx context.Context) error {
	if hs.IsRunning() {
		return fmt.Errorf("HTTP server is already running")
	}

	hs.LogInfo("Starting HTTP server", map[string]interface{}{
		"address": hs.config.GetAddress(),
		"tls":     hs.config.TLS.Enabled,
	})

	// 设置监听器
	listener, err := net.Listen("tcp", hs.config.GetAddress())
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", hs.config.GetAddress(), err)
	}

	// 如果启用TLS
	if hs.config.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(hs.config.TLS.CertFile, hs.config.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		listener = tls.NewListener(listener, tlsConfig)
		hs.LogInfo("TLS enabled", map[string]interface{}{
			"cert_file": hs.config.TLS.CertFile,
			"key_file":  hs.config.TLS.KeyFile,
		})
	}

	// 启动服务器
	go func() {
		if err := hs.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			hs.LogError("HTTP server error", err)
		}
	}()

	hs.SetRunning(true)
	return nil
}

// Stop 停止HTTP服务端
func (hs *HTTPServer) Stop(ctx context.Context) error {
	if !hs.IsRunning() {
		return fmt.Errorf("HTTP server is not running")
	}

	hs.LogInfo("Stopping HTTP server", map[string]interface{}{
		"address": hs.config.GetAddress(),
	})

	// 创建关闭上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), hs.config.ShutdownTimeout)
	defer cancel()

	// 优雅关闭
	if err := hs.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	hs.SetRunning(false)
	return hs.Shutdown(ctx)
}

// buildHandler 构建请求处理器
func (hs *HTTPServer) buildHandler() http.Handler {
	return hs.mux
}

// registerRoutes 注册路由
func (hs *HTTPServer) registerRoutes() {
	// 注册配置中的路由
	for _, route := range hs.config.Routes {
		hs.mux.HandleFunc(route.Path, hs.createRouteHandler(route))
	}

	// 注册默认路由
	hs.mux.HandleFunc("/metrics", hs.handleMetrics)
	hs.mux.HandleFunc("/test/delay", hs.handleDelay)
	hs.mux.HandleFunc("/test/status", hs.handleStatus)
	hs.mux.HandleFunc("/test/data", hs.handleData)
	hs.mux.HandleFunc("/echo", hs.handleEcho)
}

// createRouteHandler 创建路由处理器
func (hs *HTTPServer) createRouteHandler(route RouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查HTTP方法
		if r.Method != route.Method && route.Method != "*" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 添加延迟
		if route.Delay > 0 {
			time.Sleep(route.Delay)
		}

		// 设置响应头
		w.Header().Set("Content-Type", route.ContentType)
		w.WriteHeader(route.StatusCode)

		// 生成响应体
		if route.Response != nil {
			// 如果响应包含timestamp字段，更新它
			if _, hasTimestamp := route.Response["timestamp"]; hasTimestamp {
				route.Response["timestamp"] = time.Now().Unix()
			}

			if jsonData, err := json.Marshal(route.Response); err == nil {
				w.Write(jsonData)
			} else {
				hs.LogError("Failed to marshal response", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}
	}
}

// handleMetrics 处理指标请求
func (hs *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := hs.GetMetrics()
	if jsonData, err := json.Marshal(metrics); err == nil {
		w.Write(jsonData)
	} else {
		hs.LogError("Failed to marshal metrics", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleDelay 处理延迟测试请求
func (hs *HTTPServer) handleDelay(w http.ResponseWriter, r *http.Request) {
	delayParam := r.URL.Query().Get("delay")
	if delayParam == "" {
		delayParam = "100ms"
	}

	delay, err := time.ParseDuration(delayParam)
	if err != nil {
		http.Error(w, "Invalid delay parameter", http.StatusBadRequest)
		return
	}

	time.Sleep(delay)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Delay test completed",
		"delay":   delay.String(),
		"time":    time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleStatus 处理状态码测试请求
func (hs *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	codeParam := r.URL.Query().Get("code")
	if codeParam == "" {
		codeParam = "200"
	}

	var statusCode int
	if _, err := fmt.Sscanf(codeParam, "%d", &statusCode); err != nil {
		http.Error(w, "Invalid status code parameter", http.StatusBadRequest)
		return
	}

	if statusCode < 100 || statusCode > 599 {
		http.Error(w, "Status code out of range", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status_code": statusCode,
		"message":     fmt.Sprintf("Status %d test", statusCode),
		"time":        time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleData 处理数据大小测试请求
func (hs *HTTPServer) handleData(w http.ResponseWriter, r *http.Request) {
	sizeParam := r.URL.Query().Get("size")
	if sizeParam == "" {
		sizeParam = "1024"
	}

	var size int
	if _, err := fmt.Sscanf(sizeParam, "%d", &size); err != nil {
		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
		return
	}

	if size < 0 || size > 10*1024*1024 { // 10MB limit
		http.Error(w, "Size out of range", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 生成指定大小的数据
	data := make([]byte, size)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}

	response := map[string]interface{}{
		"size":    size,
		"message": "Data size test",
		"data":    string(data),
		"time":    time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleEcho 处理回显请求
func (hs *HTTPServer) handleEcho(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = strings.Join(values, ", ")
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"query":       r.URL.RawQuery,
		"headers":     headers,
		"body":        string(body),
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"time":        time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetHTTPConfig 获取HTTP配置
func (hs *HTTPServer) GetHTTPConfig() *HTTPServerConfig {
	return hs.config
}

// GetRequestCount 获取请求计数
func (hs *HTTPServer) GetRequestCount() int64 {
	hs.mutex.RLock()
	defer hs.mutex.RUnlock()
	return hs.requestCount
}

// AddMiddleware 添加中间件
func (hs *HTTPServer) AddMiddleware(middleware MiddlewareFunc) {
	hs.middleware = append(hs.middleware, middleware)
}

// GetMetrics 获取HTTP服务端指标
func (hs *HTTPServer) GetMetrics() map[string]interface{} {
	baseMetrics := hs.BaseServer.GetMetrics()

	// 添加HTTP特定指标
	baseMetrics["request_count"] = hs.GetRequestCount()
	baseMetrics["read_timeout"] = hs.config.ReadTimeout.String()
	baseMetrics["write_timeout"] = hs.config.WriteTimeout.String()
	baseMetrics["idle_timeout"] = hs.config.IdleTimeout.String()
	baseMetrics["max_header_bytes"] = hs.config.MaxHeaderBytes
	baseMetrics["tls_enabled"] = hs.config.TLS.Enabled
	baseMetrics["cors_enabled"] = hs.config.CORS.Enabled

	// 路由信息
	routes := make([]map[string]interface{}, len(hs.config.Routes))
	for i, route := range hs.config.Routes {
		routes[i] = map[string]interface{}{
			"path":        route.Path,
			"method":      route.Method,
			"status_code": route.StatusCode,
			"delay":       route.Delay.String(),
		}
	}
	baseMetrics["routes"] = routes

	return baseMetrics
}
