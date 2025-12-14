# TimescaleDB Docker 使用指南

## 快速开始

### 1. 启动 TimescaleDB

```bash
./start_timescaledb.sh
```

脚本会自动：
- ✅ 检查 Docker 是否安装
- ✅ 创建数据目录（`timescaledb_data`）
- ✅ 启动 TimescaleDB 容器
- ✅ 等待数据库就绪
- ✅ 创建表结构和索引
- ✅ 创建触发器（用于实时通知）

### 2. 停止 TimescaleDB

```bash
./stop_timescaledb.sh
```

### 3. 查看日志

```bash
docker logs timescaledb
```

### 4. 进入数据库

```bash
docker exec -it timescaledb psql -U postgres -d behaviortree
```

## 连接信息

启动后，你会看到以下连接信息：

```
连接信息：
  主机: localhost
  端口: 5432
  数据库: behaviortree
  用户名: postgres
  密码: postgres

连接字符串：
  postgresql://postgres:postgres@localhost:5432/behaviortree
```

## 配置说明

### 默认配置

脚本中的默认配置：

```bash
CONTAINER_NAME="timescaledb"           # 容器名称
IMAGE_NAME="timescale/timescaledb:latest-pg16"  # 镜像名称
POSTGRES_USER="postgres"               # 数据库用户名
POSTGRES_PASSWORD="postgres"           # 数据库密码
POSTGRES_DB="behaviortree"            # 数据库名称
PORT="5432"                            # 端口
DATA_DIR="./timescaledb_data"         # 数据目录
```

### 修改配置

编辑 `start_timescaledb.sh` 文件，修改上述变量即可。

## 常用命令

### 查看容器状态

```bash
docker ps | grep timescaledb
```

### 查看日志

```bash
# 查看所有日志
docker logs timescaledb

# 实时查看日志
docker logs -f timescaledb
```

### 重启容器

```bash
docker restart timescaledb
```

### 删除容器（数据会保留）

```bash
docker stop timescaledb
docker rm timescaledb
```

### 删除容器和数据

```bash
docker stop timescaledb
docker rm timescaledb
rm -rf timescaledb_data
```

## 数据库操作

### 连接数据库

```bash
docker exec -it timescaledb psql -U postgres -d behaviortree
```

### 基本 SQL 操作

```sql
-- 查看所有表
\dt

-- 查看表结构
\d behavior_tree_events

-- 查看数据
SELECT * FROM behavior_tree_events LIMIT 10;

-- 查看数据量
SELECT COUNT(*) FROM behavior_tree_events;

-- 按服务器查询
SELECT * FROM behavior_tree_events
WHERE server_id = 'server-001'
ORDER BY time DESC
LIMIT 10;
```

### 测试插入数据

```sql
-- 插入测试数据
INSERT INTO behavior_tree_events 
(time, server_id, object_id, behavior_tree_id, event_type, data)
VALUES 
(NOW(), 'server-001', 12345, 67890, 'post_initialize', '{"test": "data"}');

-- 查询刚插入的数据
SELECT * FROM behavior_tree_events
WHERE server_id = 'server-001'
ORDER BY time DESC
LIMIT 1;
```

## 数据持久化

数据存储在 `timescaledb_data` 目录中，即使删除容器，数据也不会丢失。

### 备份数据

```bash
# 备份整个数据目录
tar -czf timescaledb_backup_$(date +%Y%m%d).tar.gz timescaledb_data
```

### 恢复数据

```bash
# 停止容器
docker stop timescaledb
docker rm timescaledb

# 恢复数据目录
tar -xzf timescaledb_backup_YYYYMMDD.tar.gz

# 重新启动
./start_timescaledb.sh
```

## 性能优化

### 调整 PostgreSQL 配置

编辑容器内的配置文件：

```bash
# 进入容器
docker exec -it timescaledb bash

# 编辑配置
vi /var/lib/postgresql/data/postgresql.conf

# 修改配置后重启
docker restart timescaledb
```

### 推荐配置（8GB 内存系统）

```conf
shared_buffers = 2GB
work_mem = 64MB
maintenance_work_mem = 512MB
effective_cache_size = 4GB
```

## 故障排查

### 容器无法启动

```bash
# 查看错误日志
docker logs timescaledb

# 检查端口是否被占用
lsof -i :5432

# 检查 Docker 是否运行
docker ps
```

### 数据库连接失败

```bash
# 检查容器是否运行
docker ps | grep timescaledb

# 检查数据库是否就绪
docker exec timescaledb pg_isready -U postgres
```

### 权限问题

```bash
# 确保脚本有执行权限
chmod +x start_timescaledb.sh
chmod +x stop_timescaledb.sh
```

## 升级 TimescaleDB

```bash
# 停止容器
./stop_timescaledb.sh

# 拉取最新镜像
docker pull timescale/timescaledb:latest-pg16

# 重新启动（数据会自动保留）
./start_timescaledb.sh
```

## 使用不同版本的 PostgreSQL

TimescaleDB 支持多个 PostgreSQL 版本：

```bash
# PostgreSQL 14
IMAGE_NAME="timescale/timescaledb:latest-pg14"

# PostgreSQL 15
IMAGE_NAME="timescale/timescaledb:latest-pg15"

# PostgreSQL 16（默认）
IMAGE_NAME="timescale/timescaledb:latest-pg16"
```

## 从 Go 代码连接

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

// 连接字符串
connStr := "postgres://postgres:postgres@localhost:5432/behaviortree?sslmode=disable"

// 连接数据库
db, err := sql.Open("postgres", connStr)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 测试连接
err = db.Ping()
if err != nil {
    log.Fatal(err)
}

fmt.Println("连接成功！")
```

## 总结

### 启动步骤

1. 运行 `./start_timescaledb.sh`
2. 等待数据库就绪（约 10-30 秒）
3. 使用连接信息连接数据库

### 数据安全

- ✅ 数据存储在 `timescaledb_data` 目录
- ✅ 删除容器不会删除数据
- ✅ 可以随时备份和恢复

### 性能

- ✅ 自动创建索引
- ✅ 自动分区（按时间）
- ✅ 支持实时通知（LISTEN/NOTIFY）

