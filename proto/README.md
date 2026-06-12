# Two-To Proto 契约说明

`proto/` 是 Two-To 前后端共享的接口契约源。所有 HTTP 接口的请求参数、返回值、列表项、详情项等 DTO 都优先在这里定义，再生成到后端和前端使用。

## 设计目标

- 请求和响应结构由 proto 统一定义，避免前后端各写一份类型。
- 后端 controller 使用生成后的 Go struct 作为 request/response。
- 前端使用生成后的 TypeScript 类型作为接口入参和返回值类型。
- proto 文件和业务模块保持一致，方便按模块查找。

## 目录约定

```text
proto/
  common.proto        # 通用分页、空响应等基础结构
  user.proto          # 用户与偏好
  assessment.proto    # 适配性测评
  breed.proto         # 品种解析
  pet.proto           # 宠物档案
  care.proto          # 健康照料和日常提醒
```

## 生成目标

```text
services/api/proto/             # Go 生成物，后端使用
apps/web/src/shared/proto/      # TypeScript 生成物，前端使用
```

## 编写约定

- 每个接口都成对定义 `XxxRequest` 和 `XxxResponse`。
- 列表接口统一使用 `pn`、`rn`、`total`。
- proto 字段名使用 `snake_case`，JSON 字段名使用 lowerCamel。
- Go 侧如果使用 `encoding/json` 解析 proto 结构，生成后需要注入 json tag。
- 不在 proto 中表达数据库表结构；数据库模型仍放在后端 `models/dao`。
- 不在 proto 中表达前端页面状态；页面状态仍放在前端 feature/page 内部。

## 生成命令

项目提供了脚本：

```bash
./scripts/gen-proto.sh
```

当前脚本会生成 Go 结构到 `services/api/proto/`，并为 Go 结构注入 JSON tag。前端 TypeScript 生成器后续随着前端工程初始化一起接入。
