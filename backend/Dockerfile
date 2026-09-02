# 使用官方 Go 镜像作为基础镜像
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建 HTTP 服务和独立定时任务服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taskd ./cmd/taskd

# 运行时镜像
FROM alpine:latest

# HTTPS 和 IANA 时区 Cron 运行依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/taskd .

# 复制配置目录
COPY --from=builder /app/config ./config

# 创建运行目录
RUN mkdir -p uploads logs

# 暴露端口
EXPOSE 8083

# 运行应用
CMD ["./main"]
