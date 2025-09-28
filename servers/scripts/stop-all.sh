#!/bin/bash

# abc-runner 多协议服务端停止脚本

set -e

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PID_DIR="$PROJECT_DIR/.pids"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 停止服务端
stop_servers() {
    log_info "停止所有abc-runner服务端..."
    
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
                while kill -0 "$pid" 2>/dev/null && [[ $count -lt 10 ]]; do
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
                log_warn "$service_name 服务端未运行"
            fi
            
            rm -f "$pid_file"
        fi
    done
    
    # 尝试通过进程名停止
    for process in http-server tcp-server udp-server grpc-server multi-server; do
        if pgrep -f "$process" > /dev/null; then
            log_info "通过进程名停止 $process..."
            pkill -TERM -f "$process" || true
            sleep 2
            pkill -KILL -f "$process" || true
            ((stopped++))
        fi
    done
    
    if [[ $stopped -eq 0 ]]; then
        log_info "没有运行的服务端需要停止"
    else
        log_info "🛑 已停止服务端进程"
    fi
    
    # 清理PID目录
    if [[ -d "$PID_DIR" ]]; then
        rm -rf "$PID_DIR"
    fi
}

# 主程序
main() {
    log_info "abc-runner 多协议服务端停止脚本"
    stop_servers
    log_info "✅ 停止操作完成"
}

# 执行主程序
main