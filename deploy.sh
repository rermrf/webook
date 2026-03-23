#!/bin/bash

# ============================================================
# WeBook 社区项目 - Docker 快速部署脚本
# ============================================================
# 使用方法:
#   ./deploy.sh              # 完整部署（基础设施 + 所有服务）
#   ./deploy.sh infra        # 仅部署基础设施
#   ./deploy.sh services     # 仅部署微服务（需先启动基础设施）
#   ./deploy.sh monitoring   # 部署监控服务
#   ./deploy.sh stop         # 停止所有服务
#   ./deploy.sh clean        # 清理所有容器和数据
#   ./deploy.sh logs [svc]   # 查看日志
#   ./deploy.sh status       # 查看服务状态
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.full.yaml"

# 打印带颜色的消息
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi

    log_success "Docker 环境检查通过"
}

# Docker Compose 命令封装
docker_compose() {
    if docker compose version &> /dev/null; then
        docker compose -f "$COMPOSE_FILE" "$@"
    else
        docker-compose -f "$COMPOSE_FILE" "$@"
    fi
}

# 等待服务健康
wait_for_service() {
    local service=$1
    local max_attempts=${2:-30}
    local attempt=1

    log_info "等待 $service 服务启动..."

    while [ $attempt -le $max_attempts ]; do
        if docker_compose ps "$service" | grep -q "healthy\|running"; then
            log_success "$service 服务已就绪"
            return 0
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "$service 服务启动超时"
    return 1
}

# 部署基础设施
deploy_infra() {
    log_info "正在部署基础设施服务..."

    # 创建必要的目录
    mkdir -p "$PROJECT_ROOT/deploy/mysql"

    # 复制初始化脚本
    if [ -f "$PROJECT_ROOT/deploy/mysql/init.sql" ]; then
        mkdir -p "$PROJECT_ROOT/script/mysql"
        cp "$PROJECT_ROOT/deploy/mysql/init.sql" "$PROJECT_ROOT/script/mysql/init.sql"
    fi

    # 启动基础设施
    docker_compose up -d mysql redis etcd mongo kafka elasticsearch

    # 等待关键服务就绪
    log_info "等待基础设施服务就绪..."
    sleep 10

    wait_for_service mysql 60
    wait_for_service redis 30
    wait_for_service etcd 30
    wait_for_service kafka 60

    log_success "基础设施部署完成"
}

# 构建服务镜像
build_services() {
    log_info "正在构建服务镜像..."

    local services=(
        "bff"
        "user"
        "article"
        "code"
        "comment"
        "follow"
        "interactive"
        "credit"
        "notification"
        "search"
        "tag"
        "feed"
        "ranking"
        "reward"
        "payment"
        "oauth2"
        "sms"
        "account"
        "cronjob"
    )

    for svc in "${services[@]}"; do
        log_info "构建 $svc 服务..."
        docker_compose build "$svc" || log_warn "构建 $svc 失败，跳过"
    done

    log_success "服务镜像构建完成"
}

# 部署微服务
deploy_services() {
    log_info "正在部署微服务..."

    # 核心服务（按依赖顺序启动）
    local core_services="user code sms oauth2"
    local app_services="article interactive comment follow tag feed ranking"
    local extra_services="credit notification search reward payment account cronjob"

    log_info "启动核心服务..."
    docker_compose up -d $core_services
    sleep 5

    log_info "启动应用服务..."
    docker_compose up -d $app_services
    sleep 5

    log_info "启动扩展服务..."
    docker_compose up -d $extra_services
    sleep 3

    log_info "启动 BFF 网关..."
    docker_compose up -d bff

    log_success "微服务部署完成"
}

# 部署监控服务
deploy_monitoring() {
    log_info "正在部署监控服务..."
    docker_compose --profile monitoring up -d
    log_success "监控服务部署完成"
    log_info "Prometheus: http://localhost:9091"
    log_info "Grafana: http://localhost:3000 (admin/admin)"
    log_info "Zipkin: http://localhost:9411"
}

# 完整部署
deploy_all() {
    log_info "============================================"
    log_info "开始完整部署 WeBook 社区项目"
    log_info "============================================"

    check_docker
    deploy_infra
    build_services
    deploy_services

    log_success "============================================"
    log_success "WeBook 部署完成！"
    log_success "============================================"
    log_info ""
    log_info "服务访问地址："
    log_info "  BFF API:      http://localhost:8081"
    log_info "  Prometheus:   http://localhost:8082/metrics"
    log_info ""
    log_info "基础设施："
    log_info "  MySQL:        localhost:13306 (root/root)"
    log_info "  Redis:        localhost:6379"
    log_info "  Etcd:         localhost:12379"
    log_info "  Kafka:        localhost:9094"
    log_info "  Elasticsearch: localhost:9200"
    log_info "  MongoDB:      localhost:27017 (root/example)"
    log_info ""
    log_info "使用 './deploy.sh logs bff' 查看 BFF 服务日志"
    log_info "使用 './deploy.sh status' 查看所有服务状态"
}

# 停止所有服务
stop_all() {
    log_info "正在停止所有服务..."
    docker_compose down
    log_success "所有服务已停止"
}

# 清理所有容器和数据
clean_all() {
    log_warn "此操作将删除所有容器和数据卷，是否继续？(y/N)"
    read -r confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        log_info "正在清理..."
        docker_compose down -v --rmi local
        log_success "清理完成"
    else
        log_info "取消清理"
    fi
}

# 查看日志
view_logs() {
    local service=$1
    if [ -z "$service" ]; then
        docker_compose logs -f --tail=100
    else
        docker_compose logs -f --tail=100 "$service"
    fi
}

# 查看服务状态
view_status() {
    log_info "服务状态："
    docker_compose ps
}

# 重启服务
restart_service() {
    local service=$1
    if [ -z "$service" ]; then
        log_error "请指定要重启的服务名"
        exit 1
    fi
    log_info "正在重启 $service..."
    docker_compose restart "$service"
    log_success "$service 已重启"
}

# 主函数
main() {
    cd "$PROJECT_ROOT"

    case "${1:-}" in
        infra)
            check_docker
            deploy_infra
            ;;
        services)
            check_docker
            build_services
            deploy_services
            ;;
        monitoring)
            check_docker
            deploy_monitoring
            ;;
        stop)
            stop_all
            ;;
        clean)
            clean_all
            ;;
        logs)
            view_logs "$2"
            ;;
        status)
            view_status
            ;;
        restart)
            restart_service "$2"
            ;;
        build)
            check_docker
            build_services
            ;;
        help|--help|-h)
            echo "WeBook Docker 部署脚本"
            echo ""
            echo "使用方法："
            echo "  ./deploy.sh              完整部署（基础设施 + 所有服务）"
            echo "  ./deploy.sh infra        仅部署基础设施"
            echo "  ./deploy.sh services     仅部署微服务"
            echo "  ./deploy.sh monitoring   部署监控服务"
            echo "  ./deploy.sh build        仅构建服务镜像"
            echo "  ./deploy.sh stop         停止所有服务"
            echo "  ./deploy.sh clean        清理所有容器和数据"
            echo "  ./deploy.sh logs [svc]   查看日志"
            echo "  ./deploy.sh status       查看服务状态"
            echo "  ./deploy.sh restart svc  重启指定服务"
            echo "  ./deploy.sh help         显示帮助信息"
            ;;
        *)
            deploy_all
            ;;
    esac
}

main "$@"
