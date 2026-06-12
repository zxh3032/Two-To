# Two-To Web 前端工程说明

`apps/web` 是 Two-To 的 Web 前端工程目录。前端优先使用 React 构建，工程分层采用“应用入口 + 页面 + 功能模块 + 领域实体 + 共享能力”的方式组织，目标是在项目早期保持足够清晰，后续功能变多时也能自然扩展。

## 目录结构

```text
apps/web/
  public/               # 静态资源，构建时原样输出
  src/
    app/                # 应用初始化、路由、全局 Provider、全局布局
    pages/              # 路由页面，负责组装 feature 和 entity 展示
    features/           # 具体业务功能，例如适配测评、品种检索、宠物档案编辑
    entities/           # 领域实体，例如 user、pet、breed、assessment
    shared/
      api/              # 请求客户端、接口封装、错误处理
      config/           # 前端环境配置、常量配置
      lib/              # 通用工具函数、hooks、格式化逻辑
      styles/           # 全局样式、主题变量、设计 token
      ui/               # 通用 UI 组件，例如 Button、Modal、FormItem
    assets/             # 图片、字体、图标等源码资产
  tests/                # 前端测试与测试工具
```

## 分层职责

### app

应用级入口，只处理全局装配：

- 创建 React 根节点。
- 注册路由。
- 注入全局 Provider，例如主题、请求缓存、登录态。
- 放置全局布局、错误边界、页面级 loading。

不在 `app/` 中写具体业务逻辑。

### pages

路由页面层，负责把一个页面需要的业务功能组合起来。

示例：

- `pages/home`
- `pages/assessment`
- `pages/breeds`
- `pages/pets`

页面层可以读取路由参数、设置页面标题、组织布局，但不直接写复杂业务规则。

### features

功能模块层，承载明确的用户动作或业务流程。

示例：

- `features/pet-matching`：宠物适配测评。
- `features/breed-guide`：品种解读和筛选。
- `features/pet-profile`：宠物档案创建和编辑。
- `features/care-reminder`：健康与日常提醒。

一个 feature 可以依赖 `entities` 和 `shared`，但不反向依赖页面。

### entities

领域实体层，沉淀稳定的业务对象、类型、展示组件和基础数据处理。

示例：

- `entities/pet`
- `entities/user`
- `entities/breed`
- `entities/assessment`

当多个功能都需要“宠物卡片”“品种标签”“用户养宠偏好”等概念时，应优先沉淀到 `entities`。

### shared

共享能力层，只放与具体业务无关或弱相关的基础能力。

- `shared/api`：HTTP client、API error、请求拦截。
- `shared/config`：环境变量读取、路由路径、常量。
- `shared/lib`：日期格式化、校验函数、通用 hooks。
- `shared/styles`：全局 CSS、主题变量。
- `shared/ui`：无业务语义的基础组件。

## 依赖方向

推荐依赖方向：

```text
app -> pages -> features -> entities -> shared
```

约束：

- `shared` 不依赖任何上层业务模块。
- `entities` 不依赖 `features` 或 `pages`。
- `features` 不直接依赖其他不相关 feature，跨功能复用时下沉到 `entities` 或 `shared`。
- `pages` 只做页面装配，不承担核心业务规则。

## 首期功能落点

| 功能 | 建议目录 |
| --- | --- |
| 首页与产品介绍 | `pages/home` |
| 适配性测评 | `features/pet-matching`、`entities/assessment` |
| 品种解析 | `features/breed-guide`、`entities/breed` |
| 宠物档案 | `features/pet-profile`、`entities/pet` |
| 健康照料 | `features/care-reminder`、`entities/pet` |

## 开发约定

- 新页面先放在 `pages`，页面内的业务动作逐步抽到 `features`。
- 新接口封装优先放在对应 feature 内；如果多个模块复用，再下沉到 `entities` 或 `shared/api`。
- UI 组件先就近放置；确认可复用且无业务语义后，再移动到 `shared/ui`。
- 类型定义优先贴近使用场景；跨模块复用的领域类型放到 `entities`。
- 样式优先组件内聚，主题色、间距、圆角、字体等基础变量放到 `shared/styles`。
