# Two-To API 后端工程说明

`backend` 是 Two-To 的 Go 后端 API 服务目录。当前技术栈为 Gin + GORM MySQL + zap。后端分层参考了本机代码库 `/Users/zhouxuhao/baidu/educloud/elderly-ai-care-agent_master` 的组织方式，保留其清晰的调用链和职责拆分：

```text
routers -> controllers -> models/page -> models/data -> models/dao
```

参考仓库中的典型链路是：

- `routers/router.go`：集中注册路由分组。
- `controllers/{module}`：处理 HTTP 入口，创建 request/response，然后交给 page 层。
- `models/page/{module}`：承载一次请求的业务编排、参数校验、权限校验、跨 data 调用。
- `models/data`：封装数据库读写、查询条件、事务和数据组合。
- `models/dao`：定义数据库表结构、枚举、`TableName` 等持久化模型。
- `library`：放鉴权、配置、外部服务 client、通用工具等基础能力。

Two-To 会沿用这个分层思想，但避免直接绑定参考仓库中的公司内部 servlet/resident 框架。项目早期优先使用清晰、轻量的 Go HTTP API 组织方式，同时保留 proto 作为请求参数和返回值的统一契约。

本项目会保留参考仓库中“请求参数和返回值由 proto 定义”的方式。`backend/proto/` 是后端接口契约目录，里面同时放 `.proto` 源文件和 Go 生成物 `*.pb.go`，controller 使用生成后的 `XxxRequest` 和 `XxxResponse`。

## 常用命令

```bash
go run .
go test ./...
GOOS=linux GOARCH=amd64 go test ./...
./scripts/gen-proto.sh
```

后端默认监听 `:8080`。如果 `TWO_TO_MYSQL_DSN` 为空，服务会跳过数据库初始化，便于先跑通 API 骨架和前后端联调。

## 目录结构

```text
backend/
  go.mod
  go.sum
  cmd/
    two-to-api/         # API 服务启动入口，装配配置、日志、数据库和路由
  routers/              # 路由注册与接口分组
  controllers/          # HTTP controller，负责请求解析、响应输出
  proto/                # 接口 proto 源文件和 Go 生成物
  models/
    page/               # 业务编排层，对应一次用户请求或业务用例
    data/               # 数据访问层，封装查询、写入、事务和数据组合
    dao/                # 数据库表结构、枚举、TableName、持久化模型
  library/              # 通用基础能力，例如配置、日志、数据库、响应、鉴权、外部 client
  middlewares/          # HTTP 中间件，例如 request id、请求日志、跨域、错误恢复
  conf/                 # 本地、测试、生产等环境配置
  migrations/           # 数据库迁移 SQL 或迁移脚本
  scripts/              # 本地工具脚本、一次性任务、维护脚本
```

## 分层职责

### cmd

服务启动入口，负责组装配置、数据库连接、HTTP server、路由和后台任务。

约束：

- 不写业务逻辑。
- 不直接操作数据库表。
- 只做启动、依赖装配和生命周期管理。

### routers

路由层负责注册接口路径、HTTP 方法、中间件和 controller 映射。

当前已注册：

```text
GET /health
GET /api/v1/ping
GET /api/v1/slogan
```

示例规划：

```text
/api/v1/assessments
/api/v1/breeds
/api/v1/pets
/api/v1/care-reminders
```

约束：

- 只声明路由，不写业务流程。
- 按业务模块分组，避免所有接口堆在一个超大文件中。

### controllers

Controller 是 HTTP 适配层，保持尽可能薄。

职责：

- 解析请求参数。
- 创建由 proto 生成的 request/response。
- 调用对应 page 层。
- 统一输出成功或失败响应。

约束：

- 不直接访问 `models/data` 或 `models/dao`。
- 不写复杂业务判断。
- 不拼装复杂返回数据。

Controller 典型形态：

```go
func List(ctx *gin.Context) {
    request := &proto.ExamplePetDetailRequest{}
    response := &proto.ExamplePetDetailResponse{}
    // 解析 request，调用 page 层，输出 response
}
```

### models/page

Page 层对应参考仓库里的 `models/page`，是一次请求的业务编排层。虽然名称叫 page，但在后端语义上更接近 use case/application service。

职责：

- 参数校验。
- 登录态与权限校验。
- 调用一个或多个 data 对象完成业务流程。
- 调用外部服务 client。
- 组装响应数据。

适合放在 page 层的例子：

- 完成一次宠物适配测评并生成推荐结果。
- 创建宠物档案并初始化默认提醒。
- 查询宠物档案详情，同时聚合健康记录和最近提醒。

### models/data

Data 层负责和数据源交互，封装数据库查询、写入、事务和数据组合。

职责：

- 定义查询条件结构，例如 `PetCond`、`BreedCond`。
- 封装 CRUD、分页、批量查询。
- 处理事务边界。
- 屏蔽底层 ORM 或 SQL 细节。

约束：

- 不读取 HTTP request。
- 不处理页面展示逻辑。
- 不做用户权限判断，权限由 page 层负责。

### models/dao

DAO 层只描述持久化模型和与数据库强相关的定义。

职责：

- 表结构 struct。
- `TableName`。
- 数据库枚举。
- 枚举合法性校验。

约束：

- 不写业务流程。
- 不写复杂查询。
- 不依赖 controller/page。

### library

通用基础能力层。

可放内容：

- 配置读取。
- 日志封装。
- GORM MySQL 初始化。
- 错误码。
- 鉴权工具。
- 时间工具。
- 第三方服务 client。

约束：

- 不放 Two-To 具体业务流程。
- 能被多个模块复用才下沉到这里。

## 调用方向

推荐调用方向：

```text
cmd
  -> routers
    -> middlewares
    -> controllers
      -> proto
      -> models/page
        -> models/data
          -> models/dao
        -> library
```

反向依赖禁止：

- `models/dao` 不依赖 `models/data`、`models/page` 或 `controllers`。
- `models/data` 不依赖 `models/page` 或 `controllers`。
- `controllers` 不直接依赖 `models/dao`。
- `library` 不依赖具体业务模块。

## 首期模块规划

| 模块 | routers/controllers | models/page | models/data | models/dao |
| --- | --- | --- | --- | --- |
| 适配测评 | `assessment` | 测评提交、结果生成、历史记录 | 测评题目、答案、结果查询 | assessment tables |
| 品种解析 | `breed` | 品种列表、详情、筛选 | 品种资料、标签、适配条件查询 | breed tables |
| 宠物档案 | `pet` | 创建档案、编辑档案、档案详情 | 宠物基础信息、成长记录查询 | pet tables |
| 健康照料 | `care` | 提醒创建、完成、列表 | 疫苗、驱虫、用药、提醒查询 | care tables |
| 用户体系 | `user` | 登录态、用户资料、偏好设置 | 用户与偏好数据读写 | user tables |

## 当前已落地链路

初始化阶段已经落地最小可运行链路：

```text
GET /api/v1/slogan
  -> routers.NewRouter
    -> controllers/slogan.Get
      -> models/page/slogan.Page.Handle
        -> proto.SloganResponse

GET /api/v1/ping
  -> routers.NewRouter
    -> controllers/ping.Ping
      -> models/page/ping.Page.Handle
        -> proto.PingResponse
```

健康检查链路：

```text
GET /health
  -> controllers/health.Check
    -> 可选检查 GORM MySQL 连接
```

## 与参考仓库的取舍

保留：

- 清晰的纵向调用链。
- controller 保持轻量。
- page 层承载业务编排。
- data 层封装数据访问。
- dao 层只表达数据库模型。
- library 放通用基础能力。

暂不照搬：

- 超大单文件路由。Two-To 初期按模块拆分路由文件。
- 与公司内部框架强绑定的 servlet/resident 服务。Two-To 会保持开源 GitHub 项目更容易启动和部署。

调整：

- proto 文件放在后端目录 `backend/proto/`，与参考代码库保持一致。
- 后续如果需要给前端生成 TypeScript 类型，生成物放在 `frontend/src/shared/proto/`。

## 新增接口开发流程

1. 在 `backend/proto/` 中定义 `XxxRequest` 和 `XxxResponse`。
2. 执行 `./backend/scripts/gen-proto.sh` 生成后端 Go 结构。
3. 在 `routers` 中注册接口路径。
4. 在 `controllers/{module}` 中创建薄 controller，并使用生成后的 proto request/response。
5. 在 `models/page/{module}` 中实现业务编排。
6. 在 `models/data` 中补充数据访问方法。
7. 在 `models/dao` 中补充表结构或枚举。
8. 补充接口文档和必要测试。

## 命名建议

- Go 后端业务模块目录优先使用小写单词，必要时使用下划线；不要使用短横线，避免和 package 命名习惯冲突。
- Controller 文件按动作命名，例如 `create.go`、`list.go`、`detail.go`。
- Page 结构按业务动作命名，例如 `PagePetCreate`、`PageBreedList`。
- Data 结构按领域对象命名，例如 `PetData`、`BreedData`。
- DAO 结构按表模型命名，例如 `PetProfile`、`BreedInfo`。
