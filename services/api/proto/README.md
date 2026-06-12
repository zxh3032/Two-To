# 后端 Proto 生成目录

本目录用于存放由根目录 `proto/` 生成的 Go 结构。

如果要看手写的接口源定义，请看仓库根目录 `proto/`。本目录下的 `*.pb.go` 是生成物。

约定：

- 不手写 `*.pb.go`。
- 后端 controller 使用这里生成的 `XxxRequest` 和 `XxxResponse`。
- 源头只修改根目录 `proto/*.proto`。
- 生成命令：

```bash
./scripts/gen-proto.sh
```

典型 controller 结构：

```go
func List(ctx *gin.Context) {
    request := &proto.ExamplePetDetailRequest{}
    response := &proto.ExamplePetDetailResponse{}
    // 解析 request，调用 page 层，输出 response
}
```
