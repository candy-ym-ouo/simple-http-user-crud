# 保留完整 Go 工具链，供评测环境修改源码、编译和运行测试。
FROM golang:1.22

WORKDIR /app

# 本项目只使用 Go 标准库，没有 go.sum 和外部模块依赖。
COPY . .

# 构建一次以确认基础代码健康，并保留编译缓存。
RUN go build ./...

# 启动后进入 shell，便于评测环境执行 Go 命令。
CMD ["bash"]
