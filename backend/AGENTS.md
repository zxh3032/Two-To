# backend 开发协作规范

## 适用范围

本文件适用于 `backend/` 目录下的所有 Go 代码、proto 契约、配置、迁移脚本和维护脚本。后续使用大模型修改后端时，优先遵守本文档；如果更深层目录存在新的 `AGENTS.md`，以更深层目录文档为准。

## 项目概览

已确认：

- `backend` 是 Two-To 的 Go 后端 API 服务目录。
- Go module 是 `github.com/zxh3032/two-to/backend`，当前 `go.mod` 使用 Go `1.24.4`。
- 当前已引入 Gin、GORM MySQL、zap 和 `google.golang.org/protobuf`。
- 当前已有 `backend/proto/example.proto`、`backend/proto/system.proto` 与对应 Go 生成物。
- `go.work` 在仓库根目录使用 `./backend`，说明当前 monorepo 只把后端作为 Go workspace 成员。
- 当前已落地根目录 `main.go` 启动入口、Gin 路由、中间件、健康检查、ping controller/page 和 slogan controller/page。

推断：

- 后端会优先提供 Web API，支撑宠物适配测评、品种解析、宠物档案、健康照料、日常管理与成长记录等业务。
- 后端分层沿用 README 中声明的链路：`routers -> controllers -> models/page -> models/data -> models/dao`。

## 协作与需求确认规则

本项目是个人项目，很多产品想法、技术文档或代码调整会以一句话或一小段描述出现。大模型协作开发时，不能把这类描述直接当作完整规格执行。

要求：

- 执行任何后端代码修改前，必须先说明准备如何改、涉及哪些文件、接口或数据结构会如何变化、主要风险是什么，等用户明确同意后再落地。
- 如果用户已经明确说“直接改”“开整”“按这个做”等，可视为对当前已说明范围的同意；一旦发现新增范围、接口语义歧义、数据兼容风险或安全风险，需要再次停下来确认。
- 遇到业务语义、接口字段、错误码、权限边界、数据表设计、迁移策略、日志内容等不确定点时，先提出具体问题和可选方案，不要自行脑补后继续实现。
- 涉及需求相关内容时，需要主动和用户讨论目标用户、核心场景、边界条件、优先级和可能取舍；不要简单按字面照单执行。
- 用户不要求专业术语描述，沟通时优先使用清楚的中文解释方案、差异和影响。

## 整体架构

当前目录结构：

```text
backend/
  go.mod
  go.sum
  main.go               # API 服务启动入口，装配配置、日志、数据库和路由
  cmd/
    two-to-api/         # 历史预留目录，当前不再放服务入口
  routers/              # 路由注册与接口分组
  controllers/          # HTTP controller，负责请求解析、响应输出
  proto/                # 接口 proto 源文件和 Go 生成物
  models/
    page/               # 业务编排层，对应一次用户请求或业务用例
    data/               # 数据访问层，封装查询、写入、事务和数据组合
    dao/                # 数据库表结构、枚举、TableName、持久化模型
  library/              # 通用基础能力，例如配置、日志、鉴权、外部 client
  middlewares/          # HTTP 中间件，例如鉴权、日志、跨域、错误恢复
  conf/                 # 本地、测试、生产等环境配置
  migrations/           # 数据库迁移 SQL 或迁移脚本
  scripts/              # 本地工具脚本、一次性任务、维护脚本
```

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

- `models/dao` 不依赖 `models/data`、`models/page`、`controllers`。
- `models/data` 不依赖 `models/page`、`controllers`。
- `models/page` 不依赖 `controllers`。
- `controllers` 不直接依赖 `models/dao`。
- `library` 不依赖具体业务模块。

## 分层职责

### cmd

`main.go` 是服务启动入口。

职责：

- 加载配置。
- 初始化日志、数据库连接、外部 client。
- 创建 HTTP server。
- 注册路由和中间件。
- 处理启动、停止和资源释放。

约束：

- 不写业务逻辑。
- 不直接操作数据库表。
- 不拼装具体接口响应。

### routers

`routers` 负责注册接口路径、HTTP 方法、中间件和 controller 映射。

职责：

- 按业务模块组织路由分组。
- 保持路由路径、版本号、中间件顺序清晰。
- 将请求转发到 controller。

约束：

- 不写业务流程。
- 不解析复杂请求参数。
- 不直接调用 page/data/dao。

### middlewares

`middlewares` 负责 HTTP 横切逻辑。

适合放置：

- 请求日志。
- panic recovery。
- CORS。
- 鉴权和登录态解析。
- Trace ID 或 request ID 注入。

约束：

- 中间件只处理通用横切能力。
- 不把具体业务判断写入中间件。

### controllers

`controllers` 是 HTTP 适配层，保持薄 controller。

职责：

- 解析 query、path、body、header 等请求参数。
- 创建 `proto` 生成的 `XxxRequest` 和 `XxxResponse`。
- 调用对应 `models/page`。
- 将 page 层返回结果转换为统一 HTTP 响应。

约束：

- 不直接访问 `models/data` 或 `models/dao`。
- 不写复杂业务判断。
- 不拼装复杂聚合数据。
- 请求参数含义不清时，不在 controller 内临时猜测，应先补 proto 注释或待确认说明。

### proto

`proto` 是接口契约层。

已确认：

- 手写源文件是 `backend/proto/*.proto`。
- Go 生成物是 `backend/proto/*.pb.go`。
- `scripts/gen-proto.sh` 会先删除旧 `*.pb.go`，再通过 `protoc` 重新生成。
- 如果本机存在 `protoc-go-inject-tag`，脚本会注入 Go struct tag。

要求：

- 不手写 `*.pb.go`。
- 请求参数和返回值优先在 `.proto` 中定义。
- 字段注释要说明业务意义，尤其是 ID、状态、枚举、时间、分页和兼容字段。
- 修改 `.proto` 后必须重新生成 Go 文件，并同步考虑前端 `shared/proto` 的 TypeScript 类型。

### models/page

`models/page` 是业务编排层，语义上接近 use case 或 application service。

职责：

- 参数校验。
- 登录态和权限校验。
- 调用一个或多个 data 对象完成业务流程。
- 调用外部服务 client。
- 组装 proto response。
- 决定业务错误码和用户可理解的错误信息。

适合放置：

- 完成一次宠物适配测评并生成推荐结果。
- 创建宠物档案并初始化默认提醒。
- 查询宠物档案详情，并聚合健康记录和最近提醒。

约束：

- 不直接写 SQL 细节。
- 不处理 HTTP request/response。
- 不把通用工具随意塞入 page；多模块复用后再下沉到 `library`。

### models/data

`models/data` 是数据访问与数据组合层。

职责：

- 定义查询条件结构，例如 `PetCond`、`BreedCond`。
- 封装 CRUD、分页、批量查询。
- 处理事务边界。
- 屏蔽 ORM 或 SQL 细节。
- 必要时组合多个 DAO 查询结果。

约束：

- 不读取 HTTP request。
- 不处理页面展示逻辑。
- 不做用户权限判断，权限由 page 层负责。
- 不新增物理删除；删除语义优先遵循逻辑删除。

### models/dao

`models/dao` 是持久化模型层。

职责：

- 数据库表结构 struct。
- `TableName`。
- 数据库枚举。
- 枚举合法性校验。
- 与表字段强相关的常量。

约束：

- 不写业务流程。
- 不写复杂查询。
- 不依赖 controller/page/data。

### library

`library` 是基础能力层。

适合放置：

- 配置读取。
- 日志封装。
- 错误码。
- 鉴权工具。
- 时间工具。
- 第三方服务 client。
- 多模块复用且无具体业务归属的基础函数。

约束：

- 不放 Two-To 具体业务流程。
- 不为了少量重复逻辑过早抽 helper。

## 核心业务链路

HTTP 请求链路：

```text
HTTP 请求
  -> middlewares 注入 request id、日志、鉴权上下文
  -> routers 命中接口路由
  -> controllers 解析参数并创建 proto request/response
  -> models/page 执行业务编排、校验和权限判断
  -> models/data 执行数据库读写、事务和数据组合
  -> models/dao 映射数据库表结构与枚举
  -> models/page 组装 proto response
  -> controllers 输出统一 HTTP 响应
```

proto 变更链路：

```text
修改 backend/proto/*.proto
  -> 执行 backend/scripts/gen-proto.sh
    -> 更新 backend/proto/*.pb.go
      -> controller/page 使用新的请求与响应结构
      -> 前端后续生成 frontend/src/shared/proto/*.ts
```

当前现实：

- 已有 `/health`、`/api/v1/ping`、`/api/v1/slogan` 三个接口。
- `/api/v1/ping` 已经走通 `routers -> controllers -> models/page -> proto` 链路。
- `/api/v1/slogan` 已经走通 `routers -> controllers -> models/page -> proto` 链路，并返回项目口号“两两相逢，奔赴朝夕。”。
- 当前尚未落地真实业务 data/dao 代码。

## 数据与模型设计

已确认：

- 当前只有 proto 示例模型，还没有数据库表结构、迁移 SQL 或 DAO struct。
- `migrations/` 已预留数据库迁移目录。
- 已接入 GORM MySQL 初始化逻辑；`TWO_TO_MYSQL_DSN` 为空时跳过数据库连接，方便本地先启动 API。

约定：

- 接口层请求/响应使用 proto message 表达。
- 数据库表模型放在 `models/dao`。
- 查询条件、分页条件、写入参数可以放在 `models/data`，但字段含义必须有中文注释。
- page 层负责把 data 层结果转换为 proto response。
- 不直接把 DAO 模型暴露给 controller 或 HTTP 响应。
- 删除操作优先设计为逻辑删除，例如 `deleted_at`、`is_deleted`、`status` 等字段；不得新增物理删除作为默认实现。

## 注释规范

新增或修改代码时必须补充必要中文注释。

要求：

- 方法、函数、结构体、复杂分支需要说明作用。
- 请求参数和返回值需要说明业务意义。
- proto 字段要写清楚含义，尤其是枚举、状态、时间、ID、分页、兼容字段。
- DAO 字段要写清楚表字段业务含义。
- 复杂流程写“大步骤”注释，不做逐行翻译式注释。
- 生成物 `*.pb.go` 不手写注释；应修改 `.proto` 或生成脚本。

Go 导出标识建议：

- 导出函数、类型、常量保留 Go 常见注释风格，注释中包含标识符名称，同时用中文解释业务作用。
- 非导出函数如果承载关键流程，也要补中文注释。

## 日志与排查规范

日志是线上排查的重要依据。新增代码时，关键位置必须有日志。

必须记录：

- 服务启动、配置加载、依赖初始化失败。
- 请求入口的 method、path、request id、用户标识摘要、耗时、状态码。
- 参数校验失败的字段、业务错误码和错误摘要。
- page 层关键业务步骤开始、结束和失败。
- data 层数据库写入、事务失败、关键查询异常。
- 外部服务调用失败、超时、重试和降级。
- panic recovery 和不可预期错误。

要求：

- 日志字段结构化，至少包含模块名、动作名、request id、关键业务 ID、错误码、错误摘要。
- 不打印 token、密码、完整手机号、详细地址、敏感备注等隐私数据。
- 错误日志要保留原始 error，便于定位根因。
- controller 负责记录请求级上下文，page/data 负责记录业务级上下文，避免重复刷屏。

## 封装与模块拆分原则

不要过度封装。

具体要求：

- 不因为一两行逻辑重复就抽 helper。
- 不把需要连续阅读的业务流程拆成过多小函数。
- 只有在多处复用、职责清晰、能降低真实复杂度时，才放到 `library` 或公共 data 方法。
- 单个接口独有的校验和转换逻辑优先留在对应 page 或 controller 附近。
- data 层查询方法按真实业务查询组织，不为了形式统一制造空泛 repository。

## 增量修改与兼容要求

修改已有逻辑时，优先增量演进，不要直接删除。

要求：

- 旧逻辑需要替换时，先注释保留，并写明“待验证后删除”或“兼容旧流程”。
- 新旧逻辑差异要放在相邻位置，方便 review。
- 对接口字段、错误码、路由路径、数据库字段、枚举值的修改必须说明兼容影响。
- 涉及数据迁移时，先补迁移脚本和回滚说明，不在业务代码里临时修数据。
- 确认旧逻辑无调用、无兼容价值后，再单独提交清理。

## 新增接口开发流程

1. 在 `backend/proto/` 中定义 `XxxRequest` 和 `XxxResponse`，补齐字段中文注释。
2. 执行 `./backend/scripts/gen-proto.sh` 生成 Go 结构。
3. 在 `routers` 中注册接口路径和中间件。
4. 在 `controllers/{module}` 中创建薄 controller，解析请求并创建 proto request/response。
5. 在 `models/page/{module}` 中实现业务编排、参数校验、权限判断和响应组装。
6. 在 `models/data` 中补充数据访问、事务和组合查询。
7. 在 `models/dao` 中补充表结构、枚举和 `TableName`。
8. 补充日志、测试、接口文档和必要的迁移脚本。

## 校验要求

后端代码修改后至少执行 Linux 目标环境校验：

```bash
cd backend
GOOS=linux GOARCH=amd64 go test ./...
```

如果修改了 proto，还需要执行：

```bash
./backend/scripts/gen-proto.sh
```

如果新增了前端依赖的接口字段，需要同步考虑 `frontend/src/shared/proto` 的 TypeScript 生成物。

## 提交编号规则

后端提交编号独立计数，统一使用 `tutu-xxxx` 格式，`xxxx` 为四位数字。

要求：

- 从本次提交开始，如果历史提交中没有后端编号，则第一条后端提交使用 `tutu-0001`。
- 每次提交前，先查看历史提交中最近一次后端编号，再将编号加 1。
- 提交信息必须以编号开头，格式为 `tutu-0001 中文提交描述`。
- 不复用已经出现过的编号，不跳号，数字不足四位时左侧补 `0`。
- 如果一次改动同时包含前端和后端，优先拆成两次提交，前端提交使用 `taro-xxxx`，后端提交使用 `tutu-xxxx`，两边分别计数。

查看后端最近编号的推荐命令：

```bash
git log --format=%s -- backend | rg '^tutu-[0-9]{4}' | head -1
```

## 开发建议

新人读代码顺序：

1. 先看根目录 `README.md`，理解产品目标。
2. 再看 `backend/README.md` 和本文档，理解后端分层。
3. 接口契约先看 `backend/proto/*.proto`。
4. 业务实现落地后，按 `routers -> controllers -> models/page -> models/data -> models/dao` 顺序追链路。
5. 排查数据问题时，从 request id 日志、page 层业务日志、data 层查询/事务日志逐层定位。

新增模块建议：

- 模块名使用小写单词，必要时使用下划线，不使用短横线作为 Go package 名。
- controller 文件按动作命名，例如 `create.go`、`list.go`、`detail.go`。
- page 结构按业务动作命名，例如 `PagePetCreate`、`PageBreedList`。
- data 结构按领域对象命名，例如 `PetData`、`BreedData`。
- DAO 结构按表模型命名，例如 `PetProfile`、`BreedInfo`。

## 待确认事项

- 待确认：统一错误码和响应格式。
  当前观察：`library/response` 已有基础响应结构，但业务错误码分层尚未设计。
  不确定原因：当前还没有真实业务接口和错误场景。
  可能影响：前后端联调、错误展示、日志排查和兼容策略。

- 待确认：数据库迁移工具。
  当前观察：已接入 GORM MySQL，`migrations` 已预留，但还没有迁移工具和表结构。
  不确定原因：当前未进入具体业务表设计。
  可能影响：表结构版本管理、回滚策略和本地初始化数据方式。
