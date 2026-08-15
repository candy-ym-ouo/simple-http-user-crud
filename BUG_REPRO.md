# 用户删除后仍可查询

## 复现步骤

1. 启动服务：

```bash
go run ./cmd/server
```

2. 在另一个终端依次执行：

```bash
curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{"name":"用户一","email":"one@example.com","age":20}'
curl -X DELETE http://localhost:8080/users/1
curl -i http://localhost:8080/users/1
```

## 预期结果

删除成功后查询该用户应返回 `404 Not Found`。

## 实际结果

删除请求成功后，查询仍返回 `200 OK` 和用户数据。
