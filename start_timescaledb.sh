#!/bin/bash

# TimescaleDB Docker 启动脚本
# 用于启动 TimescaleDB 数据库容器

set -e

# 配置变量
CONTAINER_NAME="timescaledb"
IMAGE_NAME="timescale/timescaledb:latest-pg16"
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
POSTGRES_DB="behaviortree"
PORT="5432"
VOLUME_NAME="timescaledb_data"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    print_info "Docker 已安装"
}

# 删除已存在的容器
remove_container() {
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            print_info "停止并删除现有容器 ${CONTAINER_NAME}..."
            docker stop ${CONTAINER_NAME}
        else
            print_info "删除现有容器 ${CONTAINER_NAME}..."
        fi
        docker rm ${CONTAINER_NAME}
        print_info "容器已删除"
    else
        print_info "容器 ${CONTAINER_NAME} 不存在，将创建新容器"
    fi
}

# 创建或清理 Docker volume
create_volume() {
    if docker volume ls --format '{{.Name}}' | grep -q "^${VOLUME_NAME}$"; then
        print_info "Docker volume ${VOLUME_NAME} 已存在"
    else
        print_info "创建 Docker volume: ${VOLUME_NAME}"
        docker volume create ${VOLUME_NAME}
    fi
}

# 启动 TimescaleDB 容器
start_timescaledb() {
    print_info "启动 TimescaleDB 容器..."
    
    docker run -d \
        --name ${CONTAINER_NAME} \
        -e POSTGRES_USER=${POSTGRES_USER} \
        -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
        -e POSTGRES_DB=${POSTGRES_DB} \
        -p ${PORT}:5432 \
        -v ${VOLUME_NAME}:/var/lib/postgresql/data \
        ${IMAGE_NAME}
    
    if [ $? -eq 0 ]; then
        print_info "TimescaleDB 容器启动成功"
    else
        print_error "TimescaleDB 容器启动失败"
        exit 1
    fi
}

# 等待数据库就绪
wait_for_db() {
    print_info "等待数据库就绪..."
    max_attempts=60
    attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if docker exec ${CONTAINER_NAME} pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB} &> /dev/null; then
            # 再等待几秒确保数据库完全启动
            sleep 3
            # 测试实际连接
            if docker exec ${CONTAINER_NAME} psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "SELECT 1;" &> /dev/null; then
                print_info "数据库已就绪"
                return 0
            fi
        fi
        attempt=$((attempt + 1))
        echo -n "."
        sleep 1
    done
    
    echo ""
    print_error "数据库启动超时"
    return 1
}

# 初始化数据库（创建扩展和表）
init_database() {
    print_info "初始化数据库..."
    
    # 等待数据库完全就绪
    sleep 2
    
    # 创建 TimescaleDB 扩展（重试机制）
    for i in {1..5}; do
        if docker exec ${CONTAINER_NAME} psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "CREATE EXTENSION IF NOT EXISTS timescaledb;" 2>/dev/null; then
            break
        fi
        if [ $i -eq 5 ]; then
            print_error "创建 TimescaleDB 扩展失败"
            return 1
        fi
        print_warn "创建扩展失败，重试中... ($i/5)"
        sleep 2
    done
    
    # 创建表结构
    docker exec -i ${CONTAINER_NAME} psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} <<EOF
-- 创建行为树事件表
CREATE TABLE IF NOT EXISTS behavior_tree_events (
    time TIMESTAMPTZ NOT NULL,
    server_id TEXT NOT NULL,
    object_id BIGINT NOT NULL,
    behavior_tree_id BIGINT NOT NULL,
    event_type TEXT NOT NULL,
    behavior_tree_name TEXT,
    object_name TEXT,
    task_id BIGINT,
    execute_id BIGINT,
    stack_id BIGINT,
    status TEXT,
    data JSONB,
    PRIMARY KEY (time, server_id, object_id, behavior_tree_id)
);

-- 转换为超表（自动按时间分区）
SELECT create_hypertable('behavior_tree_events', 'time', if_not_exists => TRUE);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_server_time ON behavior_tree_events (server_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_object_time ON behavior_tree_events (server_id, object_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_tree_time ON behavior_tree_events (server_id, object_id, behavior_tree_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_event_type_time ON behavior_tree_events (event_type, time DESC);
CREATE INDEX IF NOT EXISTS idx_data_gin ON behavior_tree_events USING GIN (data);

-- 创建触发器函数（用于实时通知）
CREATE OR REPLACE FUNCTION notify_new_event()
RETURNS TRIGGER AS \$\$
BEGIN
    PERFORM pg_notify(
        'behavior_tree_events',
        json_build_object(
            'server_id', NEW.server_id,
            'object_id', NEW.object_id,
            'behavior_tree_id', NEW.behavior_tree_id,
            'event_type', NEW.event_type,
            'time', NEW.time
        )::text
    );
    RETURN NEW;
END;
\$\$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_notify_new_event ON behavior_tree_events;
CREATE TRIGGER trigger_notify_new_event
AFTER INSERT ON behavior_tree_events
FOR EACH ROW
EXECUTE FUNCTION notify_new_event();
EOF

    print_info "数据库初始化完成"
}

# 显示连接信息
show_connection_info() {
    echo ""
    print_info "=========================================="
    print_info "TimescaleDB 启动成功！"
    print_info "=========================================="
    echo ""
    echo "连接信息："
    echo "  主机: localhost"
    echo "  端口: ${PORT}"
    echo "  数据库: ${POSTGRES_DB}"
    echo "  用户名: ${POSTGRES_USER}"
    echo "  密码: ${POSTGRES_PASSWORD}"
    echo ""
    echo "连接字符串："
    echo "  postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${PORT}/${POSTGRES_DB}"
    echo ""
    echo "常用命令："
    echo "  查看日志: docker logs ${CONTAINER_NAME}"
    echo "  停止容器: docker stop ${CONTAINER_NAME}"
    echo "  启动容器: docker start ${CONTAINER_NAME}"
    echo "  删除容器: docker rm -f ${CONTAINER_NAME}"
    echo "  删除数据: docker volume rm ${VOLUME_NAME}"
    echo "  进入容器: docker exec -it ${CONTAINER_NAME} psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}"
    echo ""
}

# 主函数
main() {
    print_info "开始启动 TimescaleDB..."
    
    check_docker
    create_volume
    remove_container
    start_timescaledb
    
    if wait_for_db; then
        init_database
        show_connection_info
    else
        print_error "数据库启动失败，请检查日志: docker logs ${CONTAINER_NAME}"
        exit 1
    fi
}

# 运行主函数
main

