package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// Version 常量定义
const (
	ServerVersion = "1.0.0"
)

// GRPCServer 简化的gRPC服务端实现（基于HTTP/2）
type GRPCServer struct {
	*common.BaseServer

	config     *GRPCServerConfig
	httpServer *http.Server
	mux        *http.ServeMux
}

// NewGRPCServer 创建gRPC服务端
func NewGRPCServer(config *GRPCServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *GRPCServer {
	baseServer := common.NewBaseServer("grpc", config, logger, metricsCollector)

	server := &GRPCServer{
		BaseServer: baseServer,
		config:     config,
		mux:        http.NewServeMux(),
	}

	// 设置HTTP/2服务器
	server.httpServer = &http.Server{
		Addr:              config.GetAddress(),
		Handler:           server.buildHandler(),
		ReadTimeout:       config.ConnectionTimeout,
		WriteTimeout:      config.ConnectionTimeout,
		IdleTimeout:       config.ConnectionTimeout * 2,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 注册gRPC服务方法
	server.registerGRPCServices()

	return server
}

// Start 启动gRPC服务端
func (gs *GRPCServer) Start(ctx context.Context) error {
	if gs.IsRunning() {
		return fmt.Errorf("gRPC server is already running")
	}

	gs.LogInfo("Starting gRPC server", map[string]interface{}{
		"address":                gs.config.GetAddress(),
		"tls":                    gs.config.TLS.Enabled,
		"reflection":             gs.config.EnableReflection,
		"health_check":           gs.config.HealthCheck.Enabled,
		"max_concurrent_streams": gs.config.MaxConcurrentStreams,
	})

	go func() {
		var err error

		if gs.config.TLS.Enabled {
			err = gs.httpServer.ListenAndServeTLS(gs.config.TLS.CertFile, gs.config.TLS.KeyFile)
		} else {
			err = gs.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			gs.LogError("gRPC server error", err, map[string]interface{}{
				"address": gs.config.GetAddress(),
			})
		}
	}()

	gs.SetRunning(true)
	return nil
}

// Stop 停止gRPC服务端
func (gs *GRPCServer) Stop(ctx context.Context) error {
	if !gs.IsRunning() {
		return fmt.Errorf("gRPC server is not running")
	}

	gs.LogInfo("Stopping gRPC server", map[string]interface{}{
		"address": gs.config.GetAddress(),
	})

	// 优雅关闭HTTP服务器
	if err := gs.httpServer.Shutdown(ctx); err != nil {
		gs.LogError("Failed to gracefully shutdown gRPC server", err)
		return err
	}

	gs.SetRunning(false)
	return gs.Shutdown(ctx)
}

// buildHandler 构建HTTP处理器
func (gs *GRPCServer) buildHandler() http.Handler {
	// 应用中间件
	var handler http.Handler = gs.mux

	// 认证中间件
	if gs.config.Auth.Enabled {
		handler = gs.authMiddleware(handler)
	}

	// 日志中间件
	if gs.config.LogRequests {
		handler = gs.loggingMiddleware(handler)
	}

	// 指标中间件
	handler = gs.metricsMiddleware(handler)

	// gRPC Web支持中间件
	handler = gs.grpcWebMiddleware(handler)

	return handler
}

// registerGRPCServices 注册gRPC服务
func (gs *GRPCServer) registerGRPCServices() {
	// Echo服务
	gs.mux.HandleFunc("/TestService/Echo", gs.handleEcho)

	// 流服务
	gs.mux.HandleFunc("/TestService/ServerStream", gs.handleServerStream)
	gs.mux.HandleFunc("/TestService/ClientStream", gs.handleClientStream)
	gs.mux.HandleFunc("/TestService/BidirectionalStream", gs.handleBidirectionalStream)

	// 健康检查服务
	if gs.config.HealthCheck.Enabled {
		gs.mux.HandleFunc("/grpc.health.v1.Health/Check", gs.handleHealthCheck)
	}

	// 反射服务
	if gs.config.EnableReflection {
		gs.mux.HandleFunc("/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo", gs.handleReflection)
	}

	// 服务列表
	gs.mux.HandleFunc("/", gs.handleServiceList)
}

// handleEcho 处理Echo请求
func (gs *GRPCServer) handleEcho(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gs.sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解析请求
	var request map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &request); err != nil {
			// 如果不是JSON，使用原始数据
			request = map[string]interface{}{
				"message": string(body),
			}
		}
	}

	// 构建响应
	response := map[string]interface{}{
		"message":   request["message"],
		"timestamp": time.Now().Unix(),
		"echo":      true,
		"server":    "abc-runner gRPC test server",
	}

	// 发送响应
	gs.sendJSONResponse(w, response)

	// 记录指标
	duration := time.Since(start)
	gs.RecordRequest("echo", duration, true)

	if gs.config.LogRequests {
		gs.LogInfo("gRPC Echo request", map[string]interface{}{
			"method":      "Echo",
			"remote_addr": r.RemoteAddr,
			"duration":    duration.String(),
		})
	}
}

// handleServerStream 处理服务端流请求
func (gs *GRPCServer) handleServerStream(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 设置流响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// 发送多个响应
	for i := 0; i < 5; i++ {
		response := map[string]interface{}{
			"chunk_id":  i,
			"message":   fmt.Sprintf("Stream response %d", i),
			"timestamp": time.Now().Unix(),
			"is_final":  i == 4,
		}

		data, _ := json.Marshal(response)
		data = append(data, '\n')

		if _, err := w.Write(data); err != nil {
			break
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		time.Sleep(100 * time.Millisecond)
	}

	// 记录指标
	duration := time.Since(start)
	gs.RecordRequest("server_stream", duration, true)
}

// handleClientStream 处理客户端流请求
func (gs *GRPCServer) handleClientStream(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 读取流数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gs.sendError(w, "Failed to read stream", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 处理多行JSON数据
	lines := strings.Split(string(body), "\n")
	messages := make([]interface{}, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var msg interface{}
		if err := json.Unmarshal([]byte(line), &msg); err == nil {
			messages = append(messages, msg)
		}
	}

	// 构建响应
	response := map[string]interface{}{
		"received_count": len(messages),
		"messages":       messages,
		"timestamp":      time.Now().Unix(),
		"server":         "abc-runner gRPC test server",
	}

	// 发送响应
	gs.sendJSONResponse(w, response)

	// 记录指标
	duration := time.Since(start)
	gs.RecordRequest("client_stream", duration, true)
}

// handleBidirectionalStream 处理双向流请求
func (gs *GRPCServer) handleBidirectionalStream(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 设置流响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// 读取输入流
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()

	// 处理输入并发送响应流
	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		response := map[string]interface{}{
			"chunk_id":  i,
			"echo":      line,
			"timestamp": time.Now().Unix(),
			"is_final":  i == len(lines)-1,
		}

		data, _ := json.Marshal(response)
		data = append(data, '\n')

		if _, err := w.Write(data); err != nil {
			break
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		time.Sleep(50 * time.Millisecond)
	}

	// 记录指标
	duration := time.Since(start)
	gs.RecordRequest("bidirectional_stream", duration, true)
}

// handleHealthCheck 处理健康检查
func (gs *GRPCServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "SERVING",
		"timestamp": time.Now().Unix(),
		"server":    "abc-runner gRPC test server",
	}

	gs.sendJSONResponse(w, response)
}

// handleReflection 处理反射请求
func (gs *GRPCServer) handleReflection(w http.ResponseWriter, r *http.Request) {
	services := make([]map[string]interface{}, 0)

	for method, description := range ServiceMethods {
		services = append(services, map[string]interface{}{
			"method":      method,
			"description": description,
		})
	}

	response := map[string]interface{}{
		"services":  services,
		"timestamp": time.Now().Unix(),
		"server":    "abc-runner gRPC test server",
	}

	gs.sendJSONResponse(w, response)
}

// handleServiceList 处理服务列表请求
func (gs *GRPCServer) handleServiceList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":       "abc-runner gRPC Test Server",
		"version":    ServerVersion,
		"protocol":   "gRPC (HTTP/2)",
		"services":   ServiceMethods,
		"health":     gs.config.HealthCheck.Enabled,
		"reflection": gs.config.EnableReflection,
		"tls":        gs.config.TLS.Enabled,
		"auth":       gs.config.Auth.Enabled,
		"timestamp":  time.Now().Unix(),
	}

	gs.sendJSONResponse(w, response)
}

// Middleware implementations

// authMiddleware 认证中间件
func (gs *GRPCServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gs.config.Auth.RequireAuth {
			auth := r.Header.Get("Authorization")
			expectedToken := "Bearer " + gs.config.Auth.AuthToken

			if auth != expectedToken {
				gs.sendError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func (gs *GRPCServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		gs.LogInfo("gRPC request", map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"duration":    duration.String(),
		})
	})
}

// metricsMiddleware 指标中间件
func (gs *GRPCServer) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		gs.RecordRequest(r.URL.Path, duration, true)
	})
}

// grpcWebMiddleware gRPC Web支持中间件
func (gs *GRPCServer) grpcWebMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Grpc-Web")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 工具方法

// sendJSONResponse 发送JSON响应
func (gs *GRPCServer) sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		gs.LogError("Failed to encode JSON response", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendError 发送错误响应
func (gs *GRPCServer) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"error":     message,
		"status":    status,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// GetMetrics 获取gRPC服务端指标
func (gs *GRPCServer) GetMetrics() map[string]interface{} {
	baseMetrics := gs.BaseServer.GetMetrics()

	// 添加gRPC特定指标
	baseMetrics["tls"] = gs.config.TLS.Enabled
	baseMetrics["auth"] = gs.config.Auth.Enabled
	baseMetrics["reflection"] = gs.config.EnableReflection
	baseMetrics["health_check"] = gs.config.HealthCheck.Enabled
	baseMetrics["max_concurrent_streams"] = gs.config.MaxConcurrentStreams
	baseMetrics["services"] = ServiceMethods

	return baseMetrics
}

// GetGRPCConfig 获取gRPC配置
func (gs *GRPCServer) GetGRPCConfig() *GRPCServerConfig {
	return gs.config
}
