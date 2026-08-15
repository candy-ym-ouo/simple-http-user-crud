# 用户列表包含空用户

## 复现步骤

1. 启动服务：

```bash
go run ./cmd/server
```

2. 在另一个终端依次执行：

```bash
curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{"name":"用户一","email":"one@example.com","age":20}'
curl -s http://localhost:8080/users
```

## 预期结果

列表只返回已创建的用户。

## 实际结果

响应中额外包含一条 ID 为 0、字段为空的用户。
