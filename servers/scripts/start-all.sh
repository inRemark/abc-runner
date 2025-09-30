#!/bin/bash

# abc-runner 多协议服务端启动脚本
# 用于启动所有测试服务端

set -e

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$PROJECT_DIR/bin"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [[ "$DEBUG" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# 显示使用说明
show_help() {
    cat << EOF
abc-runner 多协议服务端启动脚本

用法:
    $0 [选项]

选项:
    -p, --protocols <list>    启动的协议 (all,http,tcp,udp,grpc,websocket) [默认: all]
    -H, --host <host>         监听主机 [默认: localhost]
    --http-port <port>        HTTP服务端口 [默认: 8080]
    --tcp-port <port>         TCP服务端口 [默认: 9090]
    --udp-port <port>         UDP服务端口 [默认: 9091]
    --grpc-port <port>        gRPC服务端口 [默认: 50051]
    --websocket-port <port>   WebSocket服务端口 [默认: 7070]
    -l, --log-level <level>   日志级别 (debug,info,warn,error) [默认: info]
    -d, --daemon              后台运行
    -s, --stop                停止所有服务端
    --status                  查看服务端状态
    --build                   构建所有服务端
    -h, --help                显示此帮助信息
    -v, --version             显示版本信息

示例:
    # 启动所有服务端
    $0

    # 只启动HTTP和WebSocket服务端
    $0 --protocols http,websocket

    # 在不同主机启动
    $0 --host 0.0.0.0

    # 后台运行
    $0 --daemon

    # 停止所有服务端
    $0 --stop

    # 查看状态
    $0 --status
EOF
}

# 默认配置
PROTOCOLS="all"
HOST="localhost"
HTTP_PORT=8080
TCP_PORT=9090
UDP_PORT=9091
GRPC_PORT=50051
WEBSOCKET_PORT=7070
LOG_LEVEL="info"
DAEMON=false
PID_DIR="$PROJECT_DIR/.pids"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--protocols)
            PROTOCOLS="$2"
            shift 2
            ;;
        -H|--host)
            HOST="$2"
            shift 2
            ;;
        --http-port)
            HTTP_PORT="$2"
            shift 2
            ;;
        --tcp-port)
            TCP_PORT="$2"
            shift 2
            ;;
        --udp-port)
            UDP_PORT="$2"
            shift 2
            ;;
        --grpc-port)
            GRPC_PORT="$2"
            shift 2
            ;;
        --websocket-port)
            WEBSOCKET_PORT="$2"
            shift 2
            ;;
        -l|--log-level)
            LOG_LEVEL="$2"
            shift 2
            ;;
        -d|--daemon)
            DAEMON=true
            shift
            ;;
        -s|--stop)
            stop_servers
            exit 0
            ;;
        --status)
            show_status
            exit 0
            ;;
        --build)
            build_servers
            exit 0
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            show_version
            exit 0
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查二进制文件
check_binaries() {
    log_info "检查二进制文件..."
    
    local missing=false
    for binary in http-server tcp-server udp-server grpc-server websocket-server multi-server; do
        if [[ ! -f "$BIN_DIR/$binary" ]]; then
            log_warn "二进制文件不存在: $binary"
            missing=true
        fi
    done
    
    if [[ "$missing" == "true" ]]; then
        log_info "构建缺失的二进制文件..."
        build_servers
    fi
}

# 构建服务端
build_servers() {
    log_info "构建所有服务端..."
    
    cd "$PROJECT_DIR"
    
    # 构建各个服务端
    for server in http-server tcp-server udp-server grpc-server websocket-server multi-server; do
        log_info "构建 $server..."
        if go build -o "bin/$server" "./cmd/$server"; then
            log_info "✅ $server 构建成功"
        else
            log_error "❌ $server 构建失败"
            exit 1
        fi
    done
    
    log_info "🎉 所有服务端构建完成"
}

# 创建PID目录
create_pid_dir() {
    if [[ ! -d "$PID_DIR" ]]; then
        mkdir -p "$PID_DIR"
        log_debug "创建PID目录: $PID_DIR"
    fi
}

# 启动单个服务端
start_single_server() {
    local server_name="$1"
    local server_binary="$2"
    local server_args="$3"
    
    log_info "启动 $server_name 服务端..."
    
    if [[ "$DAEMON" == "true" ]]; then
        # 后台运行
        nohup "$BIN_DIR/$server_binary" $server_args > "$PROJECT_DIR/logs/${server_name,,}.log" 2>&1 &
        local pid=$!
        echo $pid > "$PID_DIR/${server_name,,}.pid"
        log_info "✅ $server_name 服务端已启动 (PID: $pid)"
    else
        # 前台运行
        "$BIN_DIR/$server_binary" $server_args &
        local pid=$!
        echo $pid > "$PID_DIR/${server_name,,}.pid"
        log_debug "$server_name 服务端 PID: $pid"
    fi
}

# 启动服务端
start_servers() {
    log_info "启动多协议服务端..."
    log_info "协议: $PROTOCOLS"
    log_info "主机: $HOST"
    log_info "日志级别: $LOG_LEVEL"
    
    create_pid_dir
    
    # 创建日志目录
    mkdir -p "$PROJECT_DIR/logs"
    
    if [[ "$PROTOCOLS" == "all" ]]; then
        # 使用multi-server启动所有协议
        local args="--host $HOST --http-port $HTTP_PORT --tcp-port $TCP_PORT --udp-port $UDP_PORT --grpc-port $GRPC_PORT --websocket-port $WEBSOCKET_PORT --log-level $LOG_LEVEL"
        start_single_server "Multi" "multi-server" "$args"
    else
        # 单独启动指定协议
        IFS=',' read -ra PROTOCOL_LIST <<< "$PROTOCOLS"
        for protocol in "${PROTOCOL_LIST[@]}"; do
            protocol=$(echo "$protocol" | tr '[:upper:]' '[:lower:]' | xargs)
            case $protocol in
                http)
                    start_single_server "HTTP" "http-server" "--host $HOST --port $HTTP_PORT --log-level $LOG_LEVEL"
                    ;;
                tcp)
                    start_single_server "TCP" "tcp-server" "--host $HOST --port $TCP_PORT --log-level $LOG_LEVEL"
                    ;;
                udp)
                    start_single_server "UDP" "udp-server" "--host $HOST --port $UDP_PORT --log-level $LOG_LEVEL"
                    ;;
                grpc)
                    start_single_server "gRPC" "grpc-server" "--host $HOST --port $GRPC_PORT --log-level $LOG_LEVEL"
                    ;;
                websocket)
                    start_single_server "WebSocket" "websocket-server" "--host $HOST --port $WEBSOCKET_PORT --log-level $LOG_LEVEL"
                    ;;
                *)
                    log_warn "未知协议: $protocol"
                    ;;
            esac
        done
    fi
    
    if [[ "$DAEMON" == "false" ]]; then
        log_info "🚀 服务端启动完成，按 Ctrl+C 停止"
        # 等待所有后台进程
        wait
    else
        log_info "🚀 所有服务端已在后台启动"
        show_status
    fi
}

# 停止服务端
stop_servers() {
    log_info "停止所有服务端..."
    
    if [[ ! -d "$PID_DIR" ]]; then
        log_warn "没有找到运行的服务端"
        return
    fi
    
    local stopped=0
    for pid_file in "$PID_DIR"/*.pid; do
        if [[ -f "$pid_file" ]]; then
            local pid=$(cat "$pid_file")
            local service_name=$(basename "$pid_file" .pid)
            
            if kill -0 "$pid" 2>/dev/null; then
                log_info "停止 $service_name 服务端 (PID: $pid)..."
                kill -TERM "$pid"
                
                # 等待进程结束
                local count=0
                while kill -0 "$pid" 2>/dev/null && [[ $count -lt 30 ]]; do
                    sleep 1
                    ((count++))
                done
                
                if kill -0 "$pid" 2>/dev/null; then
                    log_warn "强制停止 $service_name 服务端..."
                    kill -KILL "$pid"
                fi
                
                log_info "✅ $service_name 服务端已停止"
                ((stopped++))
            else
                log_debug "$service_name 服务端未运行"
            fi
            
            rm -f "$pid_file"
        fi
    done
    
    if [[ $stopped -eq 0 ]]; then
        log_info "没有运行的服务端需要停止"
    else
        log_info "🛑 已停止 $stopped 个服务端"
    fi
}

# 显示状态
show_status() {
    log_info "服务端状态:"
    
    if [[ ! -d "$PID_DIR" ]]; then
        echo "  没有运行的服务端"
        return
    fi
    
    local running=0
    for pid_file in "$PID_DIR"/*.pid; do
        if [[ -f "$pid_file" ]]; then
            local pid=$(cat "$pid_file")
            local service_name=$(basename "$pid_file" .pid)
            
            if kill -0 "$pid" 2>/dev/null; then
                echo -e "  ✅ ${GREEN}$service_name${NC} (PID: $pid)"
                ((running++))
            else
                echo -e "  ❌ ${RED}$service_name${NC} (已停止)"
                rm -f "$pid_file"
            fi
        fi
    done
    
    if [[ $running -eq 0 ]]; then
        echo "  没有运行的服务端"
    else
        echo -e "\n  总计: ${GREEN}$running${NC} 个服务端正在运行"
    fi
}

# 显示版本
show_version() {
    echo "abc-runner 多协议服务端启动脚本"
    echo "版本: 1.0.0"
    echo "支持协议: HTTP, TCP, UDP, gRPC, WebSocket"
}

# 信号处理
cleanup() {
    log_info "接收到停止信号，正在停止服务端..."
    stop_servers
    exit 0
}

trap cleanup SIGINT SIGTERM

# 主程序
main() {
    log_info "abc-runner 多协议服务端启动脚本"
    
    check_binaries
    start_servers
}

# 执行主程序
main