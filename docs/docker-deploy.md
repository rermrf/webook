# WeBook Docker 部署指南

## 快速开始

### Windows 用户

```powershell
# 完整部署
.\deploy.bat

# 仅部署基础设施（开发环境推荐）
.\deploy.bat infra

# 查看服务状态
.\deploy.bat status

# 停止所有服务
.\deploy.bat stop
```

### Linux/Mac 用户

```bash
# 添加执行权限
chmod +x deploy.sh

# 完整部署
./deploy.sh

# 仅部署基础设施
./deploy.sh infra

# 查看服务状态
./deploy.sh status

# 停止所有服务
./deploy.sh stop
```

## 部署模式

### 1. 开发模式（推荐）

只启动基础设施，本地运行微服务：

```bash
# 启动基础设施
docker-compose -f docker-compose.infra.yaml up -d

# 本地运行 BFF
cd bff && go run . --config ./config/dev.yaml

# 本地运行其他服务...
cd user && go run . --config ./config/dev.yaml
```

### 2. 完整部署模式

启动所有服务（基础设施 + 微服务）：

```bash
docker-compose -f docker-compose.full.yaml up -d
```

### 3. 单独服务模式

```bash
# 构建并启动指定服务
docker-compose -f docker-compose.full.yaml up -d --build bff user article
```

## 服务端口列表

### 基础设施

| 服务 | 端口 | 说明 |
|------|------|------|
| MySQL | 13306 | 用户名: root, 密码: root |
| Redis | 6379 | 无密码 |
| Etcd | 12379 | 无认证 |
| Kafka | 9094 | 外部访问端口 |
| Elasticsearch | 9200 | 无认证 |
| MongoDB | 27017 | 用户名: root, 密码: example |

### 微服务

| 服务 | gRPC 端口 | HTTP 端口 | 说明 |
|------|-----------|-----------|------|
| BFF | - | 8081 | HTTP 网关 |
| User | 8091 | - | 用户服务 |
| Article | 8092 | - | 文章服务 |
| SMS | 8093 | - | 短信服务 |
| Code | 8094 | - | 验证码服务 |
| Ranking | 8095 | - | 排行榜服务 |
| Reward | 8096 | - | 打赏服务 |
| Comment | 8097 | - | 评论服务 |
| Follow | 8098 | - | 关注服务 |
| Search | 8099 | - | 搜索服务 |
| Notification | 8100 | - | 通知服务 |
| Credit | 8101 | - | 积分服务 |
| Tag | 8102 | - | 标签服务 |
| Feed | 8103 | - | Feed 服务 |
| OAuth2 | 8105 | - | OAuth2 服务 |

### 监控服务（可选）

| 服务 | 端口 | 说明 |
|------|------|------|
| Prometheus | 9091 | 监控指标 |
| Grafana | 3000 | 可视化仪表盘 (admin/admin) |
| Zipkin | 9411 | 链路追踪 |

## 常用命令

```bash
# 查看所有服务状态
docker-compose -f docker-compose.full.yaml ps

# 查看服务日志
docker-compose -f docker-compose.full.yaml logs -f bff
docker-compose -f docker-compose.full.yaml logs -f --tail=100 user

# 重启单个服务
docker-compose -f docker-compose.full.yaml restart bff

# 重新构建并启动
docker-compose -f docker-compose.full.yaml up -d --build bff

# 停止并删除容器
docker-compose -f docker-compose.full.yaml down

# 停止并删除容器和数据卷
docker-compose -f docker-compose.full.yaml down -v

# 启动监控服务
docker-compose -f docker-compose.full.yaml --profile monitoring up -d
```

## 配置说明

### 配置文件位置

每个服务都有对应的 Docker 配置文件：

```
服务名/config/docker.yaml
```

配置文件中的关键配置：

- 数据库连接使用容器名 `mysql` 而非 `localhost`
- Redis 使用容器名 `redis`
- Etcd 使用容器名 `etcd`
- Kafka 使用容器名 `kafka`

### 环境变量

可以通过环境变量覆盖配置：

```yaml
services:
  bff:
    environment:
      - DB_DSN=root:root@tcp(mysql:3306)/webook
      - REDIS_ADDR=redis:6379
```

## 故障排查

### 1. 服务启动失败

```bash
# 查看服务日志
docker-compose -f docker-compose.full.yaml logs bff

# 检查容器状态
docker ps -a

# 进入容器调试
docker exec -it webook-bff sh
```

### 2. 数据库连接失败

确保 MySQL 已完全启动：

```bash
# 检查 MySQL 健康状态
docker-compose -f docker-compose.full.yaml ps mysql

# 手动连接测试
docker exec -it webook-mysql mysql -uroot -proot -e "SHOW DATABASES;"
```

### 3. 服务发现失败

检查 Etcd 是否正常：

```bash
# 检查 Etcd 状态
docker exec -it webook-etcd etcdctl endpoint health

# 查看注册的服务
docker exec -it webook-etcd etcdctl get --prefix /service
```

### 4. 网络问题

确保所有服务在同一网络：

```bash
# 查看网络
docker network ls

# 检查网络连接
docker network inspect webook_webook-net
```

## 生产环境建议

1. **使用外部数据库**：生产环境应使用托管数据库服务
2. **配置持久化存储**：使用云存储卷或 NFS
3. **启用 TLS**：配置 HTTPS 和 gRPC TLS
4. **资源限制**：为每个服务配置 CPU 和内存限制
5. **日志收集**：配置 ELK 或其他日志收集方案
6. **健康检查**：确保所有服务配置健康检查
7. **水平扩展**：使用 Kubernetes 或 Docker Swarm

## 目录结构

```
webook/
├── deploy/
│   ├── Dockerfile.base          # 基础构建镜像
│   ├── Dockerfile.service       # 通用服务镜像
│   ├── config/
│   │   └── prometheus.yaml      # Prometheus 配置
│   └── mysql/
│       └── init.sql             # 数据库初始化脚本
├── docker-compose.yaml          # 原有基础设施配置
├── docker-compose.infra.yaml    # 开发用基础设施配置
├── docker-compose.full.yaml     # 完整部署配置
├── deploy.sh                    # Linux/Mac 部署脚本
├── deploy.bat                   # Windows 部署脚本
└── */config/docker.yaml         # 各服务 Docker 配置
```
