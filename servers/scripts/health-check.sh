#!/bin/bash

# abc-runner 多协议服务端健康检查脚本

set -e

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认端口
HTTP_PORT=8080
TCP_PORT=9090
UDP_PORT=9091
GRPC_PORT=50051
HOST="localhost"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
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
        -h|--help)
            echo "用法: $0 [--host HOST] [--http-port PORT] [--tcp-port PORT] [--udp-port PORT] [--grpc-port PORT]"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 1
            ;;
    esac
done

# 检查HTTP服务端
check_http() {
    echo -n "检查HTTP服务端 ($HOST:$HTTP_PORT)... "
    
    if command -v curl >/dev/null 2>&1; then
        if curl -s -f "http://$HOST:$HTTP_PORT/health" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ 健康${NC}"
            return 0
        else
            echo -e "${RED}❌ 不健康${NC}"
            return 1
        fi
    else
        # 使用nc检查端口
        if timeout 3 bash -c "</dev/tcp/$HOST/$HTTP_PORT" >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠️  端口开放但无法验证健康状态${NC}"
            return 0
        else
            echo -e "${RED}❌ 端口未开放${NC}"
            return 1
        fi
    fi
}

# 检查TCP服务端
check_tcp() {
    echo -n "检查TCP服务端 ($HOST:$TCP_PORT)... "
    
    if timeout 3 bash -c "</dev/tcp/$HOST/$TCP_PORT" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ 健康${NC}"
        return 0
    else
        echo -e "${RED}❌ 不健康${NC}"
        return 1
    fi
}

# 检查UDP服务端
check_udp() {
    echo -n "检查UDP服务端 ($HOST:$UDP_PORT)... "
    
    # UDP检查比较复杂，这里只检查进程是否运行
    if pgrep -f "udp-server.*$UDP_PORT" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ 进程运行中${NC}"
        return 0
    else
        echo -e "${RED}❌ 进程未运行${NC}"
        return 1
    fi
}

# 检查gRPC服务端
check_grpc() {
    echo -n "检查gRPC服务端 ($HOST:$GRPC_PORT)... "
    
    if command -v curl >/dev/null 2>&1; then
        if curl -s -f "http://$HOST:$GRPC_PORT/" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ 健康${NC}"
            return 0
        else
            echo -e "${RED}❌ 不健康${NC}"
            return 1
        fi
    else
        # 使用nc检查端口
        if timeout 3 bash -c "</dev/tcp/$HOST/$GRPC_PORT" >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠️  端口开放但无法验证健康状态${NC}"
            return 0
        else
            echo -e "${RED}❌ 端口未开放${NC}"
            return 1
        fi
    fi
}

# 显示详细信息
show_details() {
    echo -e "\n${BLUE}=== 服务端详细信息 ===${NC}"
    
    # HTTP服务端详细信息
    if curl -s "http://$HOST:$HTTP_PORT/health" >/dev/null 2>&1; then
        echo -e "\n${GREEN}HTTP服务端:${NC}"
        echo "  地址: http://$HOST:$HTTP_PORT"
        echo "  健康检查: http://$HOST:$HTTP_PORT/health"
        echo "  指标: http://$HOST:$HTTP_PORT/metrics"
        if command -v curl >/dev/null 2>&1; then
            echo "  状态:"
            curl -s "http://$HOST:$HTTP_PORT/health" | head -3
        fi
    fi
    
    # gRPC服务端详细信息
    if curl -s "http://$HOST:$GRPC_PORT/" >/dev/null 2>&1; then
        echo -e "\n${GREEN}gRPC服务端:${NC}"
        echo "  地址: http://$HOST:$GRPC_PORT"
        echo "  服务信息: http://$HOST:$GRPC_PORT/"
        echo "  Echo服务: http://$HOST:$GRPC_PORT/TestService/Echo"
        if command -v curl >/dev/null 2>&1; then
            echo "  服务:"
            curl -s "http://$HOST:$GRPC_PORT/" | head -5
        fi
    fi
    
    # TCP和UDP服务端
    if pgrep -f "tcp-server.*$TCP_PORT" >/dev/null 2>&1; then
        echo -e "\n${GREEN}TCP服务端:${NC}"
        echo "  地址: tcp://$HOST:$TCP_PORT"
        echo "  类型: 回显服务器"
    fi
    
    if pgrep -f "udp-server.*$UDP_PORT" >/dev/null 2>&1; then
        echo -e "\n${GREEN}UDP服务端:${NC}"
        echo "  地址: udp://$HOST:$UDP_PORT"
        echo "  类型: 数据包回显服务器"
    fi
}

# 主检查函数
main_check() {
    echo -e "${BLUE}=== abc-runner 多协议服务端健康检查 ===${NC}"
    echo "检查主机: $HOST"
    echo ""
    
    local total=0
    local healthy=0
    
    # 检查各个服务端
    ((total++))
    if check_http; then ((healthy++)); fi
    
    ((total++))
    if check_tcp; then ((healthy++)); fi
    
    ((total++))
    if check_udp; then ((healthy++)); fi
    
    ((total++))
    if check_grpc; then ((healthy++)); fi
    
    echo ""
    echo -e "健康状态: ${GREEN}$healthy${NC}/$total 服务端健康"
    
    if [[ $healthy -eq $total ]]; then
        echo -e "${GREEN}✅ 所有服务端都健康${NC}"
        show_details
        exit 0
    elif [[ $healthy -eq 0 ]]; then
        echo -e "${RED}❌ 所有服务端都不健康${NC}"
        exit 1
    else
        echo -e "${YELLOW}⚠️  部分服务端不健康${NC}"
        show_details
        exit 1
    fi
}

# 执行主检查
main_check