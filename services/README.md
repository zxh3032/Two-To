# Two-To 后端服务目录

`services/` 用于存放 Two-To 的后端服务。

当前规划：

- `api/`：Two-To 主 API 服务，使用 Go 实现，为 Web 前端提供业务接口。

后续如果拆出任务调度、AI 推荐、消息通知等独立服务，可以继续在 `services/` 下新增目录。
