# Go HTTP 用户 CRUD 服务评测说明

## 项目类型

纯 Go 项目，使用 Go 标准库 `net/http`，没有前端工程和第三方 Go 依赖。

## 常用命令

```bash
# 编译全部包
go build ./...

# 运行单元测试
go test ./...

# 启动 HTTP 服务（默认监听 8080 端口）
go run ./cmd/server
```

## Docker 构建与验证

Docker 镜像保留完整 Go 1.22 工具链。项目没有外部模块依赖，因此不存在需要预下载的 `go.sum` 依赖。

```bash
chmod +x build_benzhi_docker.sh

# Apple Silicon 验证
./build_benzhi_docker.sh simple-http-user-crud linux/arm64

# x86_64 验证
./build_benzhi_docker.sh simple-http-user-crud linux/amd64

# 进入容器后验证
docker run -it simple-http-user-crud:latest
go build ./...
go test ./...
```

容器内的 `go build ./...` 不应出现任何模块下载信息。
