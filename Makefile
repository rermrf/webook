.PHONY: docker
docker:
    # 把上次编译的东西删除
	@rm webook || true
	# 运行一下 go mod tidy, 防止 go.sum 文件不对，编译失败
	@go mod tidy
	# 指定变成在 ARM 架构的 linux 操作系统上运行的可执行文件，
	# 名字叫做webook
	@GOOS=linux GOARCH=arm64 go build -tags=k8s -o webook
	# docker build
	@docker rmi -f rermrf/webook:v0.0.1
	@docker build -t rermrf/webook:v0.0.1 .

.PHONY: mock
mock:
	# 生成 mock 文件
	@mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user_mock.go
	@mockgen -source=internal/service/code.go -package=svcmocks -destination=internal/service/mocks/code_mock.go
	# 同步依赖
	@go mod tidy
