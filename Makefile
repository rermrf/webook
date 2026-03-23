# ============================================================
# WeBook 项目 Makefile
# ============================================================

.PHONY: help docker build grpc grpc_mock

# 默认目标
help:
	@echo "WeBook 项目构建命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make build              构建所有服务"
	@echo "  make build-bff          构建 BFF 服务"
	@echo "  make run-bff            运行 BFF 服务"
	@echo "  make grpc               生成 Proto 代码"
	@echo ""
	@echo "Docker 命令:"
	@echo "  make docker-infra       启动基础设施"
	@echo "  make docker-all         完整部署所有服务"
	@echo "  make docker-stop        停止所有容器"
	@echo "  make docker-clean       清理容器和数据"
	@echo "  make docker-logs        查看日志"
	@echo "  make docker-status      查看服务状态"
	@echo ""

# ============================================================
# 原有构建命令
# ============================================================

docker:
	# 把上次编译的东西删除
	@rm webook || true
	# 运行一下 go mod tidy, 防止 go.sum 文件不对，编译失败
	@go mod tidy
	# 指定变成在 ARM 架构的 linux 操作系统上运行的可执行文件
	@GOOS=linux GOARCH=arm64 go build -tags=k8s -o webook
	# docker build
	@docker rmi -f rermrf/webook:v0.0.1
	@docker build -t rermrf/webook:v0.0.1 .

# 所有服务列表
APPS := user article interactive sms code search bff follow comment tag feed ranking credit notification reward payment oauth2 account cronjob

.PHONY: build $(APPS)
build: $(APPS)

$(APPS): %:
	@echo "Building $@..."
	@rm ./cmd/$@ 2>/dev/null || true
	@GOOS=linux GOARCH=arm64 go build -o ./cmd/$@ ./$@

# 构建单个服务 (AMD64)
build-amd64-%:
	@echo "Building $* for AMD64..."
	@GOOS=linux GOARCH=amd64 go build -o ./bin/$* ./$*

grpc:
	@buf generate ./api/proto

grpc_mock:
	@mockgen -source=./api/proto/gen/payment/v1/payment_grpc.pb.go -package=pmtmocks -destination=./api/proto/gen/payment/v1/mocks/payment_grpc.mock.go
	@mockgen -source=./api/proto/gen/follow/v1/follow_grpc.pb.go -package=followMocks -destination=./api/proto/gen/follow/v1/mocks/follow_grpc.mock.go

# ============================================================
# Docker Compose 命令
# ============================================================

.PHONY: docker-infra docker-all docker-stop docker-clean docker-logs docker-status docker-monitoring

# 启动基础设施
docker-infra:
	@docker-compose -f docker-compose.infra.yaml up -d
	@echo ""
	@echo "基础设施已启动:"
	@echo "  MySQL:         localhost:13306 (root/root)"
	@echo "  Redis:         localhost:6379"
	@echo "  Etcd:          localhost:12379"
	@echo "  Kafka:         localhost:9094"
	@echo "  Elasticsearch: localhost:9200"
	@echo "  MongoDB:       localhost:27017 (root/example)"

# 完整部署
docker-all:
	@docker-compose -f docker-compose.full.yaml up -d --build
	@echo ""
	@echo "所有服务已启动"
	@echo "  BFF API: http://localhost:8081"

# 停止所有容器
docker-stop:
	@docker-compose -f docker-compose.full.yaml down 2>/dev/null || true
	@docker-compose -f docker-compose.infra.yaml down 2>/dev/null || true
	@echo "所有容器已停止"

# 清理容器和数据
docker-clean:
	@docker-compose -f docker-compose.full.yaml down -v --rmi local 2>/dev/null || true
	@docker-compose -f docker-compose.infra.yaml down -v --rmi local 2>/dev/null || true
	@echo "已清理所有容器和数据"

# 查看日志
docker-logs:
	@docker-compose -f docker-compose.full.yaml logs -f --tail=100

# 查看指定服务日志
docker-logs-%:
	@docker-compose -f docker-compose.full.yaml logs -f --tail=100 $*

# 服务状态
docker-status:
	@docker-compose -f docker-compose.full.yaml ps

# 启动监控服务
docker-monitoring:
	@docker-compose -f docker-compose.full.yaml --profile monitoring up -d
	@echo ""
	@echo "监控服务已启动:"
	@echo "  Prometheus: http://localhost:9091"
	@echo "  Grafana:    http://localhost:3000 (admin/admin)"
	@echo "  Zipkin:     http://localhost:9411"

# ============================================================
# 开发辅助命令
# ============================================================

.PHONY: run-bff run-user run-article

# 运行 BFF
run-bff:
	@cd bff && go run . --config ./config/dev.yaml

# 运行用户服务
run-user:
	@cd user && go run . --config ./config/dev.yaml

# 运行文章服务
run-article:
	@cd article && go run . --config ./config/dev.yaml

# 通用运行命令
run-%:
	@cd $* && go run . --config ./config/dev.yaml

# 格式化代码
.PHONY: fmt
fmt:
	@gofmt -w .

# 清理构建产物
.PHONY: clean
clean:
	@rm -rf bin/ cmd/
	@echo "已清理构建产物"
