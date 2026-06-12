# 后端 Proto 生成目录

本目录用于存放 Two-To 后端接口契约。

目录内同时包含手写的 `.proto` 源文件和由其生成的 Go `*.pb.go` 文件，保持和参考后端代码库一致。

约定：

- 手写 `*.proto`。
- 不手写 `*.pb.go`。
- 后端 controller 使用这里生成的 `XxxRequest` 和 `XxxResponse`。
- 源头只修改 `backend/proto/*.proto`。
- 生成命令：

```bash
./backend/scripts/gen-proto.sh
```

典型 controller 结构：

```go
func List(ctx *gin.Context) {
    request := &proto.ExamplePetDetailRequest{}
    response := &proto.ExamplePetDetailResponse{}
    // 解析 request，调用 page 层，输出 response
}
```
