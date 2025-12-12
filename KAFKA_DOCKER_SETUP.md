# Kafka Docker 配置指南

## 允许程序自动创建 Topic

### 方法一：使用 Docker Compose（推荐）

使用项目根目录下的 `docker-compose.yml` 文件：

```bash
# 启动 Kafka 和 Zookeeper
docker-compose up -d

# 查看日志
docker-compose logs -f kafka

# 停止服务
docker-compose down
```

关键配置项：
- `KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"` - 允许自动创建 topic

### 方法二：直接使用 Docker 命令

#### 1. 启动 Zookeeper

```bash
docker run -d \
  --name zookeeper \
  -p 2181:2181 \
  -e ZOOKEEPER_CLIENT_PORT=2181 \
  -e ZOOKEEPER_TICK_TIME=2000 \
  confluentinc/cp-zookeeper:latest
```

#### 2. 启动 Kafka（允许自动创建 topic）

```bash
docker run -d \
  --name kafka \
  -p 9092:9092 \
  --link zookeeper:zookeeper \
  -e KAFKA_BROKER_ID=1 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
  -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
  -e KAFKA_NUM_PARTITIONS=1 \
  -e KAFKA_DEFAULT_REPLICATION_FACTOR=1 \
  apache/kafka:latest
```

**关键参数说明：**

- `KAFKA_AUTO_CREATE_TOPICS_ENABLE=true` - **这是最重要的配置**，允许 Kafka 在程序发送消息到不存在的 topic 时自动创建它
- `KAFKA_NUM_PARTITIONS=1` - 自动创建的 topic 的默认分区数
- `KAFKA_DEFAULT_REPLICATION_FACTOR=1` - 自动创建的 topic 的默认副本因子（单节点环境使用 1）

### 方法三：使用 KRaft 模式（Kafka 2.8+，无需 Zookeeper）

对于较新版本的 Kafka（2.8+），可以使用 KRaft 模式，无需 Zookeeper：

```bash
docker run -d \
  --name kafka \
  -p 9092:9092 \
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
```

## 查看 Kafka 版本号

有多种方法可以查看 Kafka 的版本号：

### 方法一：查看 Docker 镜像信息（推荐）

```bash
# 查看 Kafka 容器的镜像信息
docker inspect kafka | grep -i image

# 或者查看容器详细信息
docker ps --format "table {{.Names}}\t{{.Image}}" | grep kafka
```

### 方法二：进入容器查看版本

```bash
# 进入 Kafka 容器
docker exec -it kafka bash

# 查看 Kafka 版本（方法 1：查看 kafka-server-start.sh）
cat /opt/kafka/bin/kafka-server-start.sh | grep "kafka" | head -1

# 查看 Kafka 版本（方法 2：查看 jar 包）
ls -la /opt/kafka/libs/ | grep kafka_ | head -1

# 查看 Kafka 版本（方法 3：使用 kafka-broker-api-versions）
kafka-broker-api-versions --bootstrap-server localhost:9092 --version
```

### 方法三：从宿主机直接执行命令

```bash
# 直接执行命令查看版本信息
docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092 --version

# 或者查看 Kafka 日志中的版本信息
docker logs kafka | grep -i "kafka version" | head -1
```

### 方法四：查看 Kafka 日志

```bash
# 查看 Kafka 启动日志，通常会显示版本信息
docker logs kafka | grep -i version

# 或者查看完整的启动日志
docker logs kafka | head -20
```

**注意：** 如果使用的是 `apache/kafka:latest` 镜像，版本号会在容器启动时显示在日志中。

## 验证配置

### 1. 检查 Kafka 是否正常运行

```bash
# 进入 Kafka 容器
docker exec -it kafka bash

# 查看 topic 列表
kafka-topics.sh --bootstrap-server localhost:9092 --list

# 或者从宿主机执行
docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --list
```

### 2. 测试自动创建 topic

运行你的测试程序：

```bash
# 基本测试命令
go test -v ./behaviortree/runtime -run TestKafkaCreateTopic

# 禁用测试缓存（推荐，确保每次都是全新运行）
go test -v -count=1 ./behaviortree/runtime -run TestKafkaCreateTopic

# 或者先清除缓存再运行
go clean -testcache
go test -v ./behaviortree/runtime -run TestKafkaCreateTopic
```

**关于测试缓存：**
- Go 默认会缓存测试结果，如果测试代码和输入没有变化，会直接返回缓存结果
- 使用 `-count=1` 可以强制每次运行测试而不使用缓存
- 这对于需要实时验证 Kafka 状态的测试特别有用

如果配置正确，程序应该能够：
- 使用 `conn.CreateTopics()` API 创建 topic（如果 Kafka 版本支持）
- 或者通过发送消息自动创建 topic（当 `KAFKA_AUTO_CREATE_TOPICS_ENABLE=true` 时）

### 3. 手动创建 topic（可选）

如果需要手动创建 topic：

```bash
docker exec kafka kafka-topics.sh \
  --create \
  --topic test-topic \
  --bootstrap-server localhost:9092 \
  --partitions 3 \
  --replication-factor 1
```

## 单节点环境特殊说明

**重要：** `auto.create.topics.enable=true` 在单节点环境下**是有效的**，但必须满足以下条件：

1. **副本因子必须设置为 1**：
   - `KAFKA_DEFAULT_REPLICATION_FACTOR=1` - 自动创建 topic 的默认副本数
   - `KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1` - 内部 offset topic 的副本数
   - `KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1` - 事务日志的副本数

2. **为什么副本因子必须为 1？**
   - 单节点只有 1 个 broker
   - 如果 replication factor > 1，Kafka 无法在多个 broker 上创建副本
   - 这会导致自动创建 topic 失败

3. **如果 replication factor 设置错误会怎样？**
   - 自动创建 topic 会失败，错误信息通常包含 "replication factor" 相关提示
   - 需要手动创建 topic 或修正配置后重启 Kafka

## 常见问题

### 问题 1：程序无法创建 topic（单节点环境）

**可能原因：**
- `KAFKA_DEFAULT_REPLICATION_FACTOR` 设置大于 1
- `KAFKA_AUTO_CREATE_TOPICS_ENABLE` 未设置为 true
- Kafka 容器未重启以应用配置

**解决方案：**
- 确保所有 replication factor 相关配置都设置为 1
- 确保 `KAFKA_AUTO_CREATE_TOPICS_ENABLE=true` 已设置
- 重启 Kafka 容器：`docker restart kafka`
- 检查 Kafka 容器日志：`docker logs kafka | grep -i replication`
- 确认网络连接正常：`telnet localhost 9092`

### 问题 2：连接被拒绝

**解决方案：**
- 检查端口映射是否正确：`docker ps | grep kafka`
- 确认 `KAFKA_ADVERTISED_LISTENERS` 配置正确
- 如果从容器外部连接，使用 `localhost:9092`
- 如果从其他容器连接，使用容器名和内部端口

### 问题 3：自动创建的 topic 配置不符合预期

**解决方案：**
- 调整 `KAFKA_NUM_PARTITIONS` 设置默认分区数
- 调整 `KAFKA_DEFAULT_REPLICATION_FACTOR` 设置默认副本因子
- 或者在代码中使用 `CreateTopics` API 指定精确配置

## 生产环境建议

在生产环境中，建议：

1. **禁用自动创建 topic**：`KAFKA_AUTO_CREATE_TOPICS_ENABLE=false`
2. **使用管理工具预先创建 topic**，确保配置正确
3. **设置合适的副本因子**（至少 3 个副本）
4. **配置持久化存储**，避免数据丢失
5. **使用 Kafka 管理工具**（如 Kafka Manager、Confluent Control Center）进行监控

## 参考资源

- [Apache Kafka 官方文档](https://kafka.apache.org/documentation/)
- [Kafka Docker 镜像](https://hub.docker.com/r/apache/kafka)
- [Kafka 配置参数说明](https://kafka.apache.org/documentation/#configuration)
