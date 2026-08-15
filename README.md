# 简易 HTTP 用户 CRUD 服务

使用 Go 标准库 `net/http` 实现的用户管理 REST 风格服务。数据仅保存在内存中，服务重启后会清空；项目没有第三方依赖。

## 目录结构

```text
simple-http-user-crud/
├── cmd/server/main.go            # 程序入口与路由注册
├── internal/handler/user.go      # HTTP 处理、JSON 编解码和参数校验
├── internal/handler/user_test.go # 接口单元测试
├── internal/model/user.go        # 用户模型与并发安全的内存仓储
├── go.mod
└── README.md
```

## 启动

需要 Go 1.22 或更高版本：

```bash
cd /Users/a1-6/Documents/go-project-2/simple-http-user-crud
go run ./cmd/server
```

服务默认监听 `http://localhost:8080`。

## 统一响应格式

成功响应：

```json
{"code":0,"message":"查询成功","data":{}}
```

失败响应的 `code` 为 HTTP 状态码，例如：

```json
{"code":400,"message":"邮箱格式无效"}
```

## 接口与 curl 测试命令

### 新增用户

```bash
curl -i -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"张三","email":"zhangsan@example.com","age":28}'
```

### 查询用户列表

```bash
curl -i http://localhost:8080/users
```

### 查询单个用户

```bash
curl -i http://localhost:8080/users/1
```

### 编辑用户

```bash
curl -i -X PUT http://localhost:8080/users/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"李四","email":"lisi@example.com","age":30}'
```

### 删除用户

```bash
curl -i -X DELETE http://localhost:8080/users/1
```

## 参数校验

- `name`：必填，最多 50 个字符。
- `email`：必填，必须是合法邮箱地址。
- `age`：0 到 150 的整数。
- 请求体只能是一个 JSON 对象，且不允许未定义字段。

## 运行测试

```bash
go test ./...
```
