#!/bin/bash

# Kafka Docker 启动脚本
# 容器名称: real-time-sync-kafka

CONTAINER_NAME="real-time-sync-kafka"
KAFKA_PORT=9092

echo "检查 Kafka 容器状态..."

# 检查容器是否存在
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "容器 ${CONTAINER_NAME} 已存在"
    
    # 检查容器是否正在运行
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo "正在停止容器 ${CONTAINER_NAME}..."
        docker stop ${CONTAINER_NAME}
    fi
    
    echo "正在删除容器 ${CONTAINER_NAME}..."
    docker rm ${CONTAINER_NAME}
    
    if [ $? -ne 0 ]; then
        echo "删除容器失败"
        exit 1
    fi
    echo "容器 ${CONTAINER_NAME} 已删除"
fi

echo "正在创建并启动新容器 ${CONTAINER_NAME}..."

# 使用 KRaft 模式启动 Kafka（无需 Zookeeper）
docker run -d \
    --name ${CONTAINER_NAME} \
    -p ${KAFKA_PORT}:9092 \
    -e KAFKA_NODE_ID=1 \
    -e KAFKA_PROCESS_ROLES=broker,controller \
    -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093 \
    -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
    -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
    -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
    -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
    -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
    -e KAFKA_NUM_PARTITIONS=1 \
    -e KAFKA_DEFAULT_REPLICATION_FACTOR=1 \
    apache/kafka:latest

if [ $? -eq 0 ]; then
    echo "容器 ${CONTAINER_NAME} 创建并启动成功"
    echo "等待 Kafka 就绪..."
    sleep 10
    echo "Kafka 地址: localhost:${KAFKA_PORT}"
    echo ""
    echo "查看日志: docker logs -f ${CONTAINER_NAME}"
    echo "停止容器: docker stop ${CONTAINER_NAME}"
    echo "删除容器: docker rm ${CONTAINER_NAME}"
else
    echo "容器创建失败"
    exit 1
fi

