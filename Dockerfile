# 基础镜像
FROM ubuntu:20.04

# 将编译好的应用复制到镜像中，放到工作目录 /app 下
COPY webook /app/webook

# 设置工作目录
WORKDIR /app

# CMD 是执行命令
# 最佳
ENTRYPOINT [ "/app/webook" ]


