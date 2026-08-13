# OpenClaw Agent 集成设计方案

> 版本: v2.1 | 日期: 2026-08-13 | 状态: 评审中（已基于系统现状核对）

> 修订记录（v2.1）：基于 sub2api 代码库现状完成对齐核对，明确「已存在 / 需新增」边界；统一错误类型为 `infraerrors`，修正 UsageLog 关联字段（当前无 `metadata` 字段）与成本字段名（`total_cost`）。详见「十九、现状核对与差距分析」。

---

## 目录

- [一、架构总览](#一架构总览)
- [二、数据流设计](#二数据流设计)
- [三、数据库模型](#三数据库模型)
- [四、API 设计](#四api-设计)
- [五、核心服务设计](#五核心服务设计)
- [六、前端设计](#六前端设计)
- [七、安全模型](#七安全模型)
- [八、LLM 计费链路](#八llm-计费链路)
- [九、模板来源管理与检索](#九模板来源管理与检索)
- [十、语言切换设计](#十语言切换设计)
- [十一、扩展模板来源](#十一扩展模板来源)
- [十二、管理端 Agent 监控](#十二管理端-agent-监控)
- [十三、用户端 Agent 监控](#十三用户端-agent-监控)
- [十四、错误处理与容错](#十四错误处理与容错)
- [十五、会话生命周期管理](#十五会话生命周期管理)
- [十六、并发与速率限制](#十六并发与速率限制)
- [十七、实施阶段](#十七实施阶段)
- [十八、关键技术决策](#十八关键技术决策)
- [十九、现状核对与差距分析](#十九现状核对与差距分析)
- [附录 A：统一配置参考](#附录-a统一配置参考)
- [附录 B：术语表](#附录-b术语表)

---

## 一、架构总览

### 1.1 系统分层

系统采用三层架构：

```
┌─────────────────────────────────────────────────┐
│                    前端层                         │
│  Vue 3 + vue-router + vue-i18n                   │
│  AgentListView / AgentChatView / AgentDashboard  │
└──────────────────────┬──────────────────────────┘
                       │ JWT Auth (HTTPS)
┌──────────────────────▼──────────────────────────┐
│                 业务层 (sub2api)                   │
│  Go + Gin + ent ORM + Redis + Google Wire DI     │
│  ├─ 用户认证 / Agent 可见性过滤                    │
│  ├─ Agent 会话管理（CRUD）                        │
│  ├─ 模板来源管理（导入/同步/更新检查）              │
│  ├─ 监控数据聚合（管理端 + 用户端）                │
│  └─ /v1 网关（API Key 鉴权 → 计费 → Token 池调度） │
└──────────────────────┬──────────────────────────┘
                       │ 内网 HTTP (127.0.0.1:9090)
┌──────────────────────▼──────────────────────────┐
│               执行层 (OpenClaw)                    │
│  Ubuntu 服务器，仅内网访问，不暴露公网              │
│  ├─ Agent 模板管理（SOUL.md + AGENTS.md + TOOLS.md）│
│  ├─ Agent 会话执行（Skill 调用 / LLM 调度）        │
│  └─ HTTP Gateway API（内网管理接口）               │
│       └─ LLM 调用 → http://127.0.0.1:8080/v1      │
└─────────────────────────────────────────────────┘
```

### 1.2 组件职责

| 组件 | 技术栈 | 职责 |
|------|--------|------|
| 前端 | Vue 3 + vue-router + vue-i18n | 用户界面、侧边栏导航、角色路由 |
| 后端 | Go + Gin + ent ORM (PostgreSQL) + Redis | 业务逻辑、鉴权、会话管理、模板管理 |
| /v1 网关 | Go + Gin | API Key 鉴权 → Group → Account pool → 上游 LLM |
| OpenClaw | Agent 执行内核 | Agent 模板执行、Skill 调用、LLM 调度 |
| 数据库 | PostgreSQL | 用户、API Key、Agent 模板、会话、用量日志 |
| 缓存 | Redis | 会话状态、速率限制、模板缓存 |

### 1.3 核心设计原则

1. **OpenClaw 不暴露公网**：`openclaw.base_url` 仅 `127.0.0.1`，Nginx 不代理 OpenClaw 端口
2. **业务层做所有鉴权**：用户认证、Agent 可见性过滤、会话归属校验均在业务层完成
3. **LLM 流量复用现有计费**：OpenClaw 调用 /v1 网关，走现有 API Key 鉴权 + Token 池调度 + 计费链路
4. **模板来源统一管理**：无论 GitHub、ClawHub、AgentPacks 还是 Hermes，统一走 AgentTemplate 表管理

---

## 二、数据流设计

### 2.1 请求生命周期

```
用户 → sub2api 前端 → sub2api 后端 (JWT)
  → OpenClaw Gateway (内网 HTTP)
    → OpenClaw 调用 LLM via sub2api /v1 (用户 API Key)
      → sub2api /v1 鉴权 + 计费 + Token 池调度
        → 上游 LLM

响应回流：
上游 LLM → sub2api /v1 → OpenClaw → sub2api 业务层 → 用户
```

### 2.2 会话创建流程

```
1. 用户选择 Agent 模板 + API Key
2. 前端 POST /api/v1/agents/sessions { agentId, apiKeyId, language? }
3. 业务层校验：
   - JWT 用户身份
   - agentId 在用户可见模板列表中
   - apiKeyId 归属当前用户且状态为 active
4. 业务层调用 OpenClaw API 创建会话：
   - 传入 agentId、llm_base_url、api_key、language
5. OpenClaw 返回 openclaw_session_id
6. 业务层写入 AgentSession 表，返回 session 信息给前端
7. 前端跳转到 AgentChatView
```

### 2.3 消息发送流程

```
1. 前端 POST /api/v1/agents/sessions/:id/messages (SSE 流式)
2. 业务层校验 session.user_id == current_user.id
3. 业务层调用 OpenClaw API 转发消息
4. OpenClaw 处理 Agent 逻辑（SOUL.md + Skills）
5. OpenClaw 调用 LLM → http://127.0.0.1:8080/v1 (用户 API Key)
6. sub2api /v1 完成鉴权、计费、调度
7. LLM 响应 → OpenClaw → 业务层 SSE 流式透传 → 前端
```

### 2.4 模板同步流程

```
模板生命周期：

[GitHub 仓库] ──git clone──→ [临时目录] ──cp──→ [OpenClaw /root/.openclaw/agents/]
                                                      │
[ClawHub 市场] ──download──→ [临时目录] ──解压──→      │
                                                      │
[管理员手写] ──直接编辑──→ ──────────────────────→     │
                                                      │
                                           [openclaw.json 注册]
                                                      │
                                           [重启 OpenClaw Gateway]
                                                      │
                                           [sub2api admin 触发同步]
                                                      │
                                           [AgentTemplate 表写入]
                                            ├─ source = github/clawhub/local/...
                                            ├─ source_version = commit/version/-
                                            ├─ tags = [...]
                                            └─ visibility = public/...
                                                      │
                                           [用户端 GET /api/v1/agents]
                                            ├─ 按 visibility 过滤
                                            ├─ 按 category/tags/source 筛选
                                            └─ 全文搜索匹配
```

---

## 三、数据库模型

### 3.1 AgentTemplate -- Agent 模板缓存表

从 OpenClaw 同步，在业务层缓存模板元数据，支持可见性过滤和来源管理。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int (PK) | 自增主键 |
| `agent_id` | string (unique) | OpenClaw 中的 agentId，如 `agent-report` |
| `name` | string | 显示名称 |
| `description` | string | 简介 |
| `category` | string | 分类：report / code / data / ops / content / finance / custom |
| `icon` | string | 图标 emoji 或标识 |
| `skills` | JSON | 可用 Skill 列表 |
| `sort_order` | int | 排序权重 |
| `status` | string | active / disabled |
| `visibility` | string | public / all_users / specific_groups |
| `allowed_group_ids` | JSON | specific_groups 模式下可见的分组 ID |
| `source` | string (enum) | 模板来源，见下方枚举定义 |
| `source_url` | string | 来源 URL（GitHub 仓库 / ClawHub 包页面等，local 为空） |
| `source_path` | string | 仓库内 agent 文件夹路径（local 为空） |
| `source_version` | string | 当前版本标识：GitHub 用 commit hash，ClawHub 用包版本号 |
| `source_author` | string | 模板作者 |
| `language` | string | 模板语言：zh / en / bilingual |
| `tags` | JSON (string[]) | 可搜索标签数组，如 `["IT", "运维", "中文"]` |
| `auto_update` | bool | 是否自动检查源更新（默认 false） |
| `update_available` | bool | 源有新版本时置 true |
| `last_source_check_at` | time.Time | 上次检查源更新的时间 |
| TimeMixin | | created_at / updated_at |
| SoftDeleteMixin | | deleted_at |

### 3.2 source 字段枚举（统一 8 种）

| 来源标识 | 来源名称 | 说明 | 更新检查方式 |
|----------|----------|------|-------------|
| `local` | 自定义手写 | 管理员直接编写 SOUL.md / AGENTS.md / TOOLS.md | 无需检查 |
| `github` | 社区 GitHub 仓库 | awesome-openclaw-agents、claw-agents 等 | git ls-remote 检查 commit |
| `clawhub` | ClawHub 市场 | hub.openclaw.ai 官方技能市场 | ClawHub API 检查版本 |
| `agentpacks` | AgentPacks 平台 | agentpacks.ai 多平台模板包市场 | AgentPacks API 检查版本 |
| `official` | OpenClaw 官方模板库 | openclaw-ai.dev 官方模板画廊 | 官方 API / GitHub 检查 |
| `dropspace` | DropSpace 社区市场 | dropspace.dev 社区模板市场 | DropSpace API 检查 |
| `hermes` | Hermes 生态兼容 | 从 Hermes 格式转换/导入的模板 | HermesAtlas 检查 |
| `custom_import` | 自定义文件导入 | 管理员上传 .skill/.zip 文件 | 无需检查 |

### 3.3 AgentSession -- 用户会话表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int (PK) | 自增主键 |
| `user_id` | int (edge → User) | 归属用户 |
| `agent_id` | string | 使用的 Agent 模板 ID |
| `openclaw_session_id` | string | OpenClaw 返回的会话 ID |
| `api_key_id` | int (edge → APIKey) | 绑定的 API Key（用于 LLM 计费） |
| `title` | string | 会话标题（首条消息自动生成或用户自定义） |
| `status` | string | active / idle / closed / error |
| `message_count` | int | 消息计数 |
| `last_message_at` | time.Time | 最后消息时间 |
| `expires_at` | time.Time | 会话过期时间（创建时间 + 24h） |
| `closed_at` | time.Time | 关闭时间 |
| `close_reason` | string | 关闭原因：user_closed / expired / admin_closed / error |
| TimeMixin + SoftDeleteMixin | | 标准混入 |

### 3.4 AgentTemplateI18n（可选，多语言变体）

| 字段 | 类型 | 说明 |
|------|------|------|
| `template_id` | int (edge → AgentTemplate) | 关联模板 |
| `locale` | string | 语言代码：zh / en |
| `display_name` | string | 该语言下的显示名称 |
| `description` | string | 该语言下的描述 |
| `soul_override` | text | 该语言的 SOUL.md 覆盖内容（可选） |

> **简化方案（推荐初期）**：不建多语言变体表，直接用 `AgentTemplate.language` 单字段标注，列表排序 + 角标即可满足需求。多语言变体表留作未来增强。

---

## 四、API 设计

### 4.1 用户端接口

新增路由组 `/api/v1/agents`，使用 JWT 认证中间件。

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/v1/agents` | 获取可见 Agent 模板列表（按 visibility + group 过滤） |
| GET | `/api/v1/agents/:agentId` | 获取 Agent 模板详情 |
| POST | `/api/v1/agents/sessions` | 创建会话（参数：agentId, apiKeyId, title?, language?） |
| GET | `/api/v1/agents/sessions` | 获取当前用户的会话列表 |
| GET | `/api/v1/agents/sessions/:id` | 获取会话详情（含历史消息代理） |
| POST | `/api/v1/agents/sessions/:id/messages` | 发送消息（SSE 流式透传） |
| DELETE | `/api/v1/agents/sessions/:id` | 关闭会话 |
| PATCH | `/api/v1/agents/sessions/:id` | 更新会话标题 |

#### 4.1.1 用户端监控接口

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/v1/agents/dashboard` | 用户 Agent 概览数据 |
| GET | `/api/v1/agents/sessions` | 我的会话列表（分页+筛选） |
| GET | `/api/v1/agents/sessions/:id` | 会话详情（含消息历史） |
| GET | `/api/v1/agents/sessions/:id/usage` | 单个会话 Token 消耗明细 |
| GET | `/api/v1/agents/usage/trends` | 我的使用趋势（参数：days=7/30） |

#### 4.1.2 筛选查询参数

`GET /api/v1/agents` 支持多维度筛选：

| 参数 | 类型 | 说明 |
|------|------|------|
| `category` | string | 分类筛选（不传=全部） |
| `tags` | string (逗号分隔) | 标签多选过滤（AND 关系） |
| `source` | string | 来源筛选（不传=全部） |
| `q` | string | 全文搜索（名称+描述+Skill） |
| `sort` | string | 排序：default / popular / newest |
| `language` | string | 语言过滤：zh / en / bilingual（不传=全部） |
| `page` | int | 分页页码 |

### 4.2 管理端接口

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/v1/admin/agents` | 管理端 Agent 模板列表 |
| POST | `/api/v1/admin/agents/sync` | 从 OpenClaw 同步模板列表 |
| PUT | `/api/v1/admin/agents/:id` | 更新模板可见性 / 排序 / 状态 |
| GET | `/api/v1/admin/agents/sessions` | 查看所有用户会话 |
| POST | `/api/v1/admin/agents/import/github` | 从 GitHub 仓库导入模板 |
| POST | `/api/v1/admin/agents/import/clawhub` | 从 ClawHub 下载或上传 .skill 导入 |
| POST | `/api/v1/admin/agents/import/agentpacks` | 从 AgentPacks 导入模板 |
| POST | `/api/v1/admin/agents/import/hermes` | 从 Hermes 导入并自动转换 |
| POST | `/api/v1/admin/agents/import/file` | 上传 .skill/.zip 文件导入 |
| POST | `/api/v1/admin/agents/:id/check-update` | 检查单个模板的源更新 |
| POST | `/api/v1/admin/agents/:id/pull-update` | 拉取源更新（需先 check-update） |
| GET | `/api/v1/admin/agents/categories` | 获取所有分类及模板数量 |
| GET | `/api/v1/admin/agents/tags` | 获取所有标签及使用数量 |

#### 管理端监控接口

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/v1/admin/agents/monitor/overview` | 总览仪表盘数据 |
| GET | `/api/v1/admin/agents/monitor/sessions` | 会话列表（分页+筛选） |
| GET | `/api/v1/admin/agents/monitor/sessions/:id` | 会话详情（含消息历史） |
| DELETE | `/api/v1/admin/agents/monitor/sessions/:id` | 强制关闭会话 |
| GET | `/api/v1/admin/agents/monitor/users` | 用户排行（分页+排序） |
| GET | `/api/v1/admin/agents/monitor/templates` | Agent 模板使用统计 |
| GET | `/api/v1/admin/agents/monitor/trends` | 趋势数据（参数：days=7/30） |
| GET | `/api/v1/admin/agents/monitor/export` | 导出监控报表（CSV） |

---

## 五、核心服务设计

### 5.1 OpenClawService

新增 `internal/service/openclaw_service.go`，负责 OpenClaw 网关通信和模板管理。

```
OpenClawService
├── SyncTemplates()                              — 从 OpenClaw 同步模板元数据到 DB
├── CreateSession(userId, agentId, apiKey)       — 调用 OpenClaw 创建会话
├── SendMessage(sessionId, message)              — 转发消息，SSE 流式透传
├── GetSessionHistory(sessionId)                 — 获取会话历史
├── CloseSession(sessionId)                      — 关闭会话
├── HealthCheck()                                — 检查 OpenClaw 状态
│
├── ── 模板来源管理 ──
├── ImportFromGitHub(repoURL, agentPath, agentId)    — 从 GitHub 导入
├── ImportFromClawHub(packageURL | file)              — 从 ClawHub 导入
├── ImportFromAgentPacks(packageURL)                  — 从 AgentPacks 导入
├── ImportFromHermes(projectURL, agentId)             — 从 Hermes 导入并转换
├── ImportFromFile(filePath, agentId)                 — 从上传文件导入
├── CheckGitHubUpdate(agentTemplate)                  — 检查 GitHub 源更新
├── CheckClawHubUpdate(agentTemplate)                 — 检查 ClawHub 源更新
├── CheckAgentPacksUpdate(agentTemplate)              — 检查 AgentPacks 源更新
├── CheckHermesUpdate(agentTemplate)                  — 检查 Hermes 源更新
├── PullUpdate(agentTemplate)                         — 拉取源更新
├── RestartOpenClawGateway()                          — 重启 OpenClaw 网关
└── GetOpenClawConfig()                               — 读取 openclaw.json
```

### 5.2 SessionCleanupService

新增 `internal/service/session_cleanup_service.go`，负责会话生命周期管理。

```
SessionCleanupService
├── RunCleanupCycle()        — 执行一次清理循环
├── MarkIdleSessions()       — 将超过 idle_timeout 的 active 会话标记为 idle
├── CloseExpiredSessions()   — 关闭超过 expires_at 的会话
├── SyncOpenClawSessions()   — 同步关闭 OpenClaw 侧的过期会话
└── StartScheduler()         — 启动定时清理任务（每 5 分钟）
```

### 5.3 RateLimitService

新增 `internal/service/rate_limit_service.go`，负责 Agent 并发和速率限制。

```
RateLimitService
├── CheckConcurrency(userId)           — 检查用户活跃会话数是否超限
├── CheckMessageRate(userId, sessionId) — 检查消息发送频率
├── RecordMessage(userId, sessionId)   — 记录消息发送（用于速率计算）
└── GetUserConcurrencyLimit(userId)    — 获取用户并发限制（复用 User.concurrency 或默认值）
```

---

## 六、前端设计

### 6.1 路由新增

| 路径 | 组件 | 认证 | 说明 |
|------|------|------|------|
| `/agents` | `user/AgentListView.vue` | requiresAuth | Agent 模板选择页 |
| `/agents/dashboard` | `user/AgentDashboardView.vue` | requiresAuth | 我的 Agent 监控面板 |
| `/agents/:sessionId` | `user/AgentChatView.vue` | requiresAuth | Agent 聊天页（交互/只读） |
| `/admin/agents` | `admin/AgentTemplatesView.vue` | admin | 模板管理（列表/导入/同步） |
| `/admin/agents/monitor` | `admin/AgentMonitorView.vue` | admin | 调用监控仪表盘 |

### 6.2 导航入口

- 首页顶部导航栏（HomeView.vue）：新增 "Agent" 链接
- 侧边栏（AppSidebar.vue）：`buildSelfNavItems` 中新增 Agent 菜单项

用户端侧边栏：

```
个人菜单 (buildSelfNavItems)
├── 仪表盘 → /dashboard
├── AI Agent → /agents
├── Agent 使用 → /agents/dashboard
├── API 密钥 → /keys
└── ...
```

管理端侧边栏：

```
AI Agent 管理 (折叠组)
├── 模板管理 → /admin/agents
└── 调用监控 → /admin/agents/monitor
```

### 6.3 页面布局

#### AgentListView.vue -- Agent 模板选择页

```
AgentListView.vue
├── 顶部搜索栏
│   ├── 全文搜索输入框（实时 debounce 300ms）
│   └── 高级筛选展开按钮
├── 分类 Tab 栏（横向滚动，单选）
├── 标签 Chip 区（多选，动态生成）
├── Agent 卡片网格（响应式，3列→2列→1列）
│   └── 每张卡片显示：
│       ├── 图标 + 来源角标
│       ├── 名称 + 语言标识（中/EN）
│       ├── 描述（截断 2 行）
│       ├── Skill 标签（最多显示 3 个）
│       └── "开始对话" 按钮
└── 空状态提示（无匹配结果时）
```

点击卡片 → 弹窗选择 API Key → 创建会话 → 跳转聊天页。

#### AgentChatView.vue -- Agent 聊天页

- 左侧：当前会话信息 + 历史会话列表
- 右侧：聊天界面（消息流 + 输入框）
- 底部：显示当前使用的 API Key 和余额

#### AgentDashboardView.vue -- 用户监控面板

**区域 1：概览卡片**

| 卡片 | 内容 | 数据来源 |
|------|------|----------|
| 活跃会话 | 当前 active 会话数 | AgentSession 表 |
| 总会话数 | 历史所有会话数 | AgentSession 表 |
| 本月消息数 | 当月发送消息总数 | AgentSession.message_count |
| 本月 Token 消耗 | 当月 Token 费用 | UsageLog 关联查询 |
| 当前余额 | 用户余额（链接到充值） | User.balance |
| 预估可用 | 按近 7 天日均消耗估算可用天数 | UsageLog 计算 |

**区域 2：我的 Agent 会话列表**

卡片式列表，可筛选（状态 / Agent / 时间范围）。每张会话卡片显示：

```
┌──────────────────────────────────────────┐
│ 财务分析助手              [活跃]  3 小时前  │
│ Agent: agent-finance                      │
│ 消息数: 12  |  Token: 15,230  |  $0.45   │
│ [继续对话]  [查看历史]  [关闭会话]        │
└──────────────────────────────────────────┘
```

**区域 3：使用趋势（简化图表）**

- 折线图：近 30 天每日 Token 消耗
- 柱状图：各 Agent 模板使用次数对比

**区域 4：会话详情页**

点击"查看历史"进入 AgentChatView.vue（只读模式），显示完整消息历史、Token 消耗明细、"继续对话"按钮。

#### 管理端 AgentTemplatesView.vue

模板列表表格：

| 列 | 说明 |
|----|------|
| Agent ID | 模板标识 |
| 名称 | 显示名称 |
| 来源 | 8 种来源标识（带图标） |
| 分类 | report / code / data / ... |
| 语言 | zh / en / bilingual |
| 状态 | active / disabled |
| 可见性 | public / all_users / specific_groups |
| 更新状态 | 最新 / 有更新（红色标记） |
| 操作 | 编辑 / 同步 / 更新 / 禁用 |

操作按钮：同步模板、导入 GitHub/ClawHub/AgentPacks/Hermes/文件、标记为自定义。

导入模板弹窗为多来源选项卡：

```
┌─────────────────────────────────────────────┐
│  导入 Agent 模板                              │
├─────────────────────────────────────────────┤
│  [GitHub] [ClawHub] [AgentPacks] [官方]      │
│  [DropSpace] [Hermes] [文件上传] [自定义]     │
├─────────────────────────────────────────────┤
│  ← 选中 GitHub 时显示：                       │
│  仓库 URL:  [___________________________]    │
│  Agent 路径: [___________________________]   │
│  目标 agentId: [_______________________]     │
│  ☑ 自动检查更新                               │
│  语言: [中文▼]  分类: [报告分析▼]             │
│  标签: [_______________] [+添加]              │
│                                              │
│  ← 选中 Hermes 时显示：                       │
│  Hermes 项目 URL: [_____________________]    │
│  ☑ 自动转换 Hermes → OpenClaw 格式           │
│  目标 agentId: [_______________________]     │
│                                              │
│  ← 选中文件上传时显示：                       │
│  [拖拽 .skill / .zip 文件到此处]              │
│  或 [点击选择文件]                            │
│  目标 agentId: [_______________________]     │
│                                              │
│           [取消]  [导入]                      │
└─────────────────────────────────────────────┘
```

#### 管理端 AgentMonitorView.vue

**Tab 1：总览仪表盘**

实时统计卡片：

| 指标 | 说明 | 数据来源 |
|------|------|----------|
| 活跃会话数 | 当前 status=active 的会话 | AgentSession 表 |
| 今日消息数 | 今日发送的消息总数 | AgentSession.message_count |
| 今日 Token 消耗 | 今日 Agent 产生的 Token 费用 | UsageLog 表（按 session 关联） |
| 活跃用户数 | 今日有 Agent 会话的用户数 | AgentSession 表 |
| 平均会话时长 | 会话从创建到最后消息的平均时间 | AgentSession 计算 |
| 错误率 | status=error 的会话占比 | AgentSession 表 |

图表区域：折线图（近 7/30 天每日会话数+消息数）、柱状图（Token 消耗）、饼图（分类占比）、排行榜（Top 10 最热门 Agent）。

**Tab 2：会话列表**

可筛选表格：

| 列 | 说明 | 筛选 |
|----|------|------|
| 会话 ID | 内部 ID | -- |
| 用户 | 邮箱 / 用户名 | 搜索 |
| Agent | 模板名称 | 下拉 |
| 状态 | active / idle / closed / error | 下拉 |
| 消息数 | 会话消息计数 | -- |
| Token 消耗 | 关联 UsageLog 汇总 | 排序 |
| 创建时间 | 会话创建时间 | 日期范围 |
| 最后消息 | last_message_at | 排序 |
| API Key | 使用的 Key 名称 | -- |
| 操作 | 查看详情 / 强制关闭 | -- |

**Tab 3：用户排行**

| 列 | 说明 |
|----|------|
| 排名 | 按消耗降序 |
| 用户 | 邮箱 / 用户名 |
| 会话总数 | 该用户创建的 Agent 会话数 |
| 活跃会话 | 当前 active 的会话数 |
| 总消息数 | 所有会话的消息总数 |
| 总 Token 消耗 | 所有会话的 Token 费用 |
| 常用 Agent | 最常使用的 Agent 模板 |
| 最后活跃 | 最后一次 Agent 会话时间 |

**Tab 4：Agent 模板统计**

| 列 | 说明 |
|----|------|
| Agent ID | 模板标识 |
| 名称 | 显示名称 |
| 来源 | 8 种来源标识 |
| 总会话数 | 使用该模板创建的会话总数 |
| 活跃会话 | 当前 active 的会话数 |
| 独立用户数 | 使用过该模板的不同用户数 |
| 总消息数 | 所有会话的消息总数 |
| 总 Token 消耗 | 所有会话的 Token 费用 |
| 平均会话消息数 | 消息数 / 会话数 |
| 更新状态 | 最新 / 有更新 |

### 6.4 i18n

新增 `agents` 命名空间（zh.ts / en.ts 同步），覆盖 Agent 列表、聊天、监控等所有 UI 文案。

---

## 七、安全模型

| 安全措施 | 实现方式 |
|----------|----------|
| OpenClaw 不暴露公网 | `openclaw.base_url` 仅 `127.0.0.1`，Nginx 不代理 OpenClaw 端口 |
| 用户鉴权 | 所有 API 走 JWT 中间件，与现有用户系统一致 |
| Agent 可见性过滤 | `AgentTemplate.visibility` + `allowed_group_ids` 控制用户可见范围 |
| 会话归属强制校验 | 每次 API 调用验证 `session.user_id == current_user.id` |
| agentId 授权校验 | 创建会话时验证 `agentId` 在用户可见模板列表中 |
| API Key 绑定 | 会话创建时绑定用户的一个 active API Key，OpenClaw 用它回调 /v1 |
| 流量审计 | 业务层记录每次 Agent 调用的 userId / agentId / sessionId / token 消耗 |
| 并发限制 | 每用户最大活跃 Agent 会话数，防止资源滥用 |
| 速率限制 | Agent 消息发送频率限制，防止恶意高频调用 |

---

## 八、LLM 计费链路

这是设计的核心优势 -- **OpenClaw 的 LLM 调用完全复用现有 /v1 网关**。

### 8.1 计费流程

1. 用户创建会话时，业务层读取用户的一个 active API Key
2. 调用 OpenClaw 创建会话时，传入 `llm_base_url` 和 `api_key`
3. OpenClaw 内部模型 provider 配置为 `http://127.0.0.1:8080/v1`
4. OpenClaw 每次调用 LLM 时，使用用户 API Key 访问 /v1
5. 现有 /v1 中间件自动完成：API Key 鉴权 → 余额检查 → Token 池调度 → 计费 → 用量记录

**无需修改 /v1 任何代码**，Agent 的 LLM 消耗与普通 API 调用一样计费。

### 8.2 Token 消耗关联方案

Agent 的 LLM 调用经过 /v1 网关，产生 UsageLog 记录，需要关联到 AgentSession 以便监控。

**解决方案**：在 OpenClaw 调用 /v1 时注入会话标识，由 /v1 网关写入 UsageLog 新增的关联列。

1. 业务层创建会话时写入 `AgentSession` 行，得到 `agent_session_id`（DB 自增主键）
2. 调用 OpenClaw 创建会话时，将 `agent_session_id` 作为会话级 metadata 传入
3. OpenClaw 调用 /v1 时，将 `agent_session_id` 放入请求（如自定义头 `X-Agent-Session-Id`）
4. sub2api /v1 网关解析该标识，写入 UsageLog 新增的 `agent_session_id` / `agent_template_id` 列
5. 管理端查询时，通过 `agent_session_id` 关联 UsageLog 和 AgentSession

**UsageLog 需新增字段**：当前 `backend/ent/schema/usage_log.go` 既无 `metadata` 字段，也无 Agent 关联字段，需通过新增迁移补齐。`UsageLog` 为只追加表（仅 `created_at`，不可更新/删除），新增列采用 `Optional().Nillable()`，历史行保持 NULL，不影响现有 /v1 计费与查询。

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_session_id` | int64 (optional) | 关联的 AgentSession 主键（仅 Agent 调用写入，普通 /v1 调用为 NULL） |
| `agent_template_id` | int64 (optional) | 关联的 AgentTemplate 主键 |

需为 `agent_session_id` 增加单列索引，以支撑会话维度的 Token 聚合查询。

查询示例：

```sql
-- 查询某会话的 Token 消耗（total_cost 为现有字段）
SELECT
    SUM(input_tokens) AS total_input,
    SUM(output_tokens) AS total_output,
    SUM(total_cost)   AS total_cost
FROM usage_logs
WHERE agent_session_id = ?;
```

---

## 九、模板来源管理与检索

### 9.1 三种基础来源

原始需求涉及三种 Agent 模板来源，全部通过 AgentTemplate 表统一管理：

1. **社区 GitHub 模板仓库** -- awesome-openclaw-agents、claw-agents 等现成模板
2. **ClawHub 在线市场**（hub.openclaw.ai）-- 浏览/下载 .skill 归档包
3. **自定义手写模板** -- 管理员直接编写 SOUL.md / AGENTS.md / TOOLS.md

三种来源的模板最终都要部署到 OpenClaw 服务器的 `/root/.openclaw/agents/` 目录，但业务层通过 `source` 字段追踪每个模板的来源、版本和更新状态。

### 9.2 来源一：社区 GitHub 模板仓库

**导入流程**：

1. 管理员在 sub2api 管理后台填写表单：仓库 URL、Agent 文件夹路径、目标 agentId、是否启用自动更新
2. sub2api 业务层执行导入（通过 OpenClaw 管理 API 或 SSH 脚本）：

```bash
git clone --depth 1 --sparse <repo_url> /tmp/openclaw-import
cd /tmp/openclaw-import
git sparse-checkout set <source_path>
cp -r <source_path>/* /root/.openclaw/agents/<agentId>/agent/
# 更新 openclaw.json 注册
# 重启 OpenClaw Gateway
```

3. 导入成功后，记录 `source=github`、`source_version=commit_hash` 到 DB
4. 同步模板元数据到 `AgentTemplate` 表

**更新策略 -- Watch 模式（推荐）**：

- `auto_update=true` 时，定时任务（每小时）执行 `git ls-remote` 检查仓库最新 commit
- 如果 commit hash 与 `source_version` 不同，设置 `update_available=true`
- 管理员在后台看到"有更新"标记，点击"拉取更新"按钮手动确认
- **不自动拉取**，因为模板变更可能改变 Agent 行为，需人工审查 changelog

### 9.3 来源二：ClawHub 在线市场

**导入流程**：

1. 管理员填写 ClawHub 包 URL 或上传 .skill 归档文件
2. sub2api 业务层执行导入：

```bash
# 方式 A：下载
wget -O /tmp/agent.skill <clawhub_download_url>
# 方式 B：管理员上传
# 接收上传的 .skill 文件

# 解压到 OpenClaw agents 目录
mkdir -p /root/.openclaw/agents/<agentId>/agent/
tar -xf /tmp/agent.skill -C /root/.openclaw/agents/<agentId>/agent/
# 更新 openclaw.json
# 重启 OpenClaw Gateway
```

3. 记录 `source=clawhub`、`source_version=package_version` 到 DB

**更新策略**：`auto_update=true` 时，定时任务调用 ClawHub API 查询包最新版本，版本号变化时设置 `update_available=true`，管理员手动确认下载更新。

### 9.4 来源三：自定义手写模板

**导入流程**：

1. 管理员直接在 OpenClaw 服务器上创建模板文件：

```
/root/.openclaw/agents/<agentId>/agent/
├── SOUL.md       -- 角色定义、性格、安全约束
├── AGENTS.md      -- 执行流程、Skill 调用逻辑
└── TOOLS.md       -- 可用工具白名单
```

2. 修改 `openclaw.json` 注册新 agentId
3. 重启 OpenClaw Gateway
4. 在 sub2api 管理后台点击"同步模板"，填写来源信息：`source=local`，设置名称、描述、分类、标签、语言、可见性

**更新策略**：无外部源，不需要版本检查。管理员修改模板文件后重启 OpenClaw，再在 sub2api 触发元数据同步。

### 9.5 更新策略总结

| 来源 | 更新方式 | 自动检查频率 | 更新动作 | 风险等级 |
|------|----------|-------------|----------|----------|
| GitHub 仓库 | Watch + 手动确认 | 每小时检查 commit | 管理员审查后拉取 | 中（行为可能变化） |
| ClawHub 市场 | 版本检查 + 手动确认 | 每天检查版本 | 管理员确认后下载 | 中（包内容可能变化） |
| AgentPacks 平台 | 版本检查 + 手动确认 | 每天检查版本 | 管理员确认后下载 | 中（包内容可能变化） |
| 官方模板库 | Watch + 手动确认 | 每小时检查 commit | 管理员审查后拉取 | 中（行为可能变化） |
| DropSpace 市场 | 版本检查 + 手动确认 | 每天检查版本 | 管理员确认后下载 | 中（包内容可能变化） |
| Hermes 兼容 | 版本检查 + 手动确认 | 每天检查版本 | 管理员确认后下载+转换 | 高（格式转换+行为变化） |
| 自定义手写 | 无需外部更新 | 不检查 | 直接编辑文件 | 低（管理员完全控制） |
| 自定义导入 | 无需外部更新 | 不检查 | 重新上传 | 低（管理员完全控制） |

**设计原则：不自动拉取外部模板更新。** 原因：

1. 模板变更可能改变 Agent 行为，影响业务稳定性
2. 社区模板大多英文，更新后可能覆盖已翻译的中文适配
3. SOUL.md 安全规则变更可能导致权限漏洞
4. 管理员需要审查 changelog 后决定是否更新

### 9.6 用户端检索设计

#### 可见性规则

三种来源的 Agent 对用户可见性完全一致，均通过 `AgentTemplate.visibility` + `allowed_group_ids` 控制：

- `public` -- 所有登录用户可见
- `all_users` -- 所有登录用户可见（与 public 区别在于可在特定时期隐藏）
- `specific_groups` -- 仅指定分组的用户可见

**来源不影响可见性** -- 无论 Agent 来自哪个来源，只要管理员将其状态设为 active 且可见性包含该用户，用户就能看到并使用。

#### 多维度检索

**1. 分类筛选（单选 tab）**

| 分类 | 说明 | 示例 Agent |
|------|------|-----------|
| 全部 | 默认，显示所有 | -- |
| 报告分析 | 报告生成、文档撰写 | agent-report |
| 代码开发 | 编码、审查、重构 | agent-code |
| 数据处理 | 数据分析、清洗、可视化 | agent-data |
| 运维管理 | 监控、部署、排障 | agent-ops |
| 内容创作 | 文案、营销、翻译 | agent-content |
| 财务分析 | 财报、预算、审计 | agent-finance |
| 自定义 | 未归入以上分类 | -- |

**2. 标签筛选（多选 chip）**：从所有模板的 `tags` 字段动态聚合，多选为 AND 关系。

**3. 来源筛选（单选 dropdown，可选）**：默认隐藏，用户可展开"高级筛选"选择来源。

**4. 全文搜索（输入框）**：实时搜索，匹配范围：Agent 名称（模糊）、描述（模糊）、标签（精确）、Skill 列表（模糊）。

**5. 排序选项**：默认排序（`sort_order`）/ 最热门（`AgentSession` 创建次数降序）/ 最新创建（`created_at` 降序）。

---

## 十、语言切换设计

### 10.1 问题分析

用户切换前端语言（中文/英文）时，Agent 列表页需要响应语言变化。涉及两个层面：

1. **UI 界面语言** -- 搜索框、分类标签、按钮等界面文案（已有 vue-i18n 支持，无需额外开发）
2. **Agent 模板语言** -- Agent 本身的 SOUL.md 语言决定了 Agent 回复的语言

### 10.2 语言切换策略

采用 **智能过滤 + 语言偏好传递** 双层方案。

#### 第一层：列表展示语言偏好

当用户切换 UI 语言时，Agent 列表页做以下调整：

1. **不隐藏不匹配的 Agent** -- 所有可见 Agent 仍然展示，避免用户错过功能
2. **语言排序优先** -- 匹配当前 UI 语言的 Agent 排在前面：
   - UI=中文 → 排序优先级：`zh` > `bilingual` > `en`
   - UI=English → 排序优先级：`en` > `bilingual` > `zh`
3. **语言角标显示** -- 每张 Agent 卡片显示语言标识：中文 / EN / 双语
4. **可选语言过滤** -- 高级筛选中提供语言下拉：全部语言 / 仅中文 / 仅英文 / 仅双语

#### 第二层：会话语言偏好传递

对于 `bilingual`（双语）类型的 Agent，创建会话时将用户当前 UI 语言传递给 OpenClaw：

```json
POST /api/v1/agents/sessions
{
  "agentId": "agent-report",
  "apiKeyId": 123,
  "language": "zh"
}
```

业务层调用 OpenClaw 创建会话时，在 SOUL.md 的 system prompt 中注入语言指令：

```
# 附加指令（业务层注入）
请使用 {language} 语言进行所有交互和输出。
```

**实现方式**：OpenClaw 支持 session-level 的 system prompt override，业务层在创建会话时通过 API 参数注入语言偏好，无需修改 SOUL.md 原文。

### 10.3 模板语言标注规范

管理员在导入/同步模板时，需设置 `language` 字段：

| 语言标识 | 判定标准 | 示例 |
|----------|----------|------|
| `zh` | SOUL.md 主要内容为中文 | 自写的中文模板 |
| `en` | SOUL.md 主要内容为英文 | 社区 GitHub 模板（大多英文） |
| `bilingual` | SOUL.md 中明确包含"respond in user's language"或中英双语指令 | 适配过的模板 |

---

## 十一、扩展模板来源

### 11.1 Hermes Agent 调研结论

**Hermes 是独立的 Agent 运行框架，不是 OpenClaw 的模板来源。**

| 维度 | OpenClaw | Hermes Agent |
|------|----------|-------------|
| 开发者 | OpenClaw 社区 | Nous Research |
| 发布时间 | 2025年 | 2026年2月 |
| 定位 | Agent 执行内核 | 自主学习型 Agent 框架 |
| 模板格式 | SOUL.md + AGENTS.md + TOOLS.md | SOUL.md + USER.md + Skills |
| 技能市场 | ClawHub（5,400+ skills） | HermesAtlas + 自学习技能 |
| 互通性 | 支持 Hermes 技能导入 | 支持 OpenClaw 技能导入 |

**关键发现**：两者模板格式相近（都基于 Markdown），且存在跨平台桥接工具（openclaw-hermes-watcher），AgentPacks 等平台同时为两者提供模板包。

### 11.2 Hermes 模板兼容方案

Hermes 和 OpenClaw 的模板都基于 Markdown 文件，但结构有差异：

| 文件 | OpenClaw | Hermes | 兼容处理 |
|------|----------|--------|----------|
| 角色定义 | SOUL.md | SOUL.md | 直接复用 |
| 执行流程 | AGENTS.md | AGENTS.md | 直接复用 |
| 工具配置 | TOOLS.md | USER.md 中的 Command allowlist | 转换：提取 allowlist → TOOLS.md |
| 用户配置 | 无 | USER.md | 忽略（业务层管理用户） |
| 技能 | Skill 目录 | Skills 目录 | 格式相同，直接复用 |

**导入 Hermes 模板时的自动转换**：

1. 管理员提供 Hermes 仓库 URL 或 .skill 包
2. 业务层下载并解析文件结构
3. 自动转换：
   - `SOUL.md` → 直接复用
   - `AGENTS.md` → 直接复用
   - `USER.md` 中的 `Command allowlist` → 提取为 `TOOLS.md`
   - `USER.md` 中的其他配置 → 忽略（API keys 等由业务层管理）
4. 部署到 OpenClaw agents 目录
5. 设置 `source=hermes`，记录原始来源 URL

### 11.3 来源对用户的透明度

**所有来源对用户透明** -- 用户不关心模板来源，只关心功能：

- 用户端 Agent 卡片**不显示来源标识**（无 GitHub/ClawHub/Hermes 角标）
- 用户端**不提供来源筛选**
- 用户端高级筛选仅保留：分类、标签、语言、搜索
- 来源信息仅出现在管理端模板管理页面

---

## 十二、管理端 Agent 监控

### 12.1 监控数据来源

所有监控数据均来自现有表 + 新增表，无需额外采集：

| 数据维度 | 数据来源 | 说明 |
|----------|----------|------|
| 会话记录 | `AgentSession` 表 | 会话数、状态、时长 |
| 消息记录 | OpenClaw API 回传 | 消息数、响应时间 |
| Token 消耗 | `UsageLog` 表（现有） | Agent 调用通过 /v1 产生 usage_log |
| 用户信息 | `User` 表（现有） | 用户名、邮箱、余额 |
| API Key | `APIKey` 表（现有） | 绑定的 Key、配额 |
| 错误信息 | 业务层日志 | OpenClaw 返回错误、超时等 |

### 12.2 实时会话监控（可选增强）

通过 WebSocket 实现实时会话监控：

- 管理端连接 `ws://api/v1/admin/agents/monitor/ws`
- 当任何用户创建/关闭会话、发送消息时，推送事件到管理端
- 管理端实时更新仪表盘和会话列表

**事件类型**：

```json
{
  "type": "session_created",
  "data": { "session_id": 123, "user_id": 456, "agent_id": "agent-report" }
}
{
  "type": "message_sent",
  "data": { "session_id": 123, "message_count": 5 }
}
{
  "type": "session_closed",
  "data": { "session_id": 123, "reason": "user_closed" }
}
{
  "type": "session_error",
  "data": { "session_id": 123, "error": "OpenClaw timeout" }
}
```

---

## 十三、用户端 Agent 监控

### 13.1 设计原则

**来源对用户完全透明** -- 用户不关心 Agent 来自 Hermes、OpenClaw 还是 ClawHub，只关心功能。因此：

- 用户端 Agent 列表**不显示来源标识**
- 用户端**不提供来源筛选**
- 用户端高级筛选仅保留：分类、标签、语言、搜索
- 来源信息仅出现在管理端模板管理页面

### 13.2 用户端 vs 管理端监控对比

| 维度 | 用户端 | 管理端 |
|------|--------|--------|
| 数据范围 | 仅自己的会话 | 全部用户会话 |
| 展示重点 | 我的消耗 / 余额 / 会话历史 | 平台总览 / 用户排行 / 模板统计 |
| 用户排行 | 无 | 有（按消耗排序） |
| 模板管理 | 无（只读列表） | 有（导入/同步/禁用） |
| 强制关闭 | 仅自己的会话 | 任意用户的会话 |
| 实时监控 | 无 | 有（WebSocket） |
| 导出报表 | 无 | 有（CSV） |
| 来源信息 | 不显示 | 显示 |
| Token 明细 | 自己的会话明细 | 全部会话明细 |

**安全约束**：所有用户端接口均通过 JWT 认证，且强制过滤 `user_id = current_user.id`，用户只能看到自己的数据。

---

## 十四、错误处理与容错

### 14.1 OpenClaw 不可用时的处理策略

当 OpenClaw 服务不可用时（宕机、重启、网络故障），业务层 MUST 保证不阻塞用户操作。

#### 14.1.1 前端体验

- **Agent 列表页**：OpenClaw 不可用时，Agent 列表仍可正常展示（数据来自 AgentTemplate 缓存表），但每个 Agent 卡片显示"服务暂不可用"标记，"开始对话"按钮置灰
- **Agent 聊天页**：如果会话已存在的 OpenClaw 会话被中断，显示"服务暂时不可用，请稍后重试"提示，保留消息历史
- **创建会话**：返回友好错误提示"Agent 服务暂时不可用，请稍后重试"，不阻塞页面

#### 14.1.2 后端处理

```
OpenClaw 不可用场景下的处理逻辑：

1. 模板同步任务失败 → 记录日志，保留上次同步成功的缓存数据
2. 创建会话请求 → 返回 503 Service Unavailable，附带友好错误信息
3. 发送消息请求 → 返回 503，会话状态不变，保留消息队列
4. 关闭会话请求 → 标记 DB 状态为 closed，待 OpenClaw 恢复后异步同步
```

### 14.2 超时重试机制

#### 14.2.1 重试策略

业务层对 OpenClaw API 调用采用**指数退避重试**策略：

| 参数 | 值 | 说明 |
|------|-----|------|
| 最大重试次数 | 3 | 同一请求最多重试 3 次 |
| 初始退避时间 | 1 秒 | 第一次重试前等待 |
| 退避倍数 | 2 | 每次重试等待时间翻倍：1s → 2s → 4s |
| 最大退避时间 | 10 秒 | 单次等待上限 |
| 总超时时间 | 30 秒 | 包含所有重试的总时间上限 |

#### 14.2.2 重试判定条件

仅在以下条件下重试：

- 网络超时（connection timeout）
- OpenClaw 返回 5xx 错误
- 连接被重置（connection reset）

以下情况**不重试**：

- OpenClaw 返回 4xx 错误（请求参数错误）
- 用户 API Key 鉴权失败
- 会话不存在（404）
- 请求被速率限制（429）

#### 14.2.3 实现伪代码

```go
func (s *OpenClawService) callWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
    var lastErr error
    backoff := 1 * time.Second

    for attempt := 0; attempt <= 3; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff):
                backoff *= 2
                if backoff > 10*time.Second {
                    backoff = 10 * time.Second
                }
            }
        }

        resp, err := s.httpClient.Do(req.Clone(ctx))
        if err != nil {
            if isRetryableError(err) {
                lastErr = err
                continue
            }
            return nil, err
        }

        if resp.StatusCode >= 500 {
            resp.Body.Close()
            lastErr = fmt.Errorf("openclaw returned %d", resp.StatusCode)
            continue
        }

        return resp, nil
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### 14.3 降级策略

#### 14.3.1 健康检查

业务层定时（每 30 秒）对 OpenClaw 执行健康检查：

```go
func (s *OpenClawService) HealthCheck() *HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := s.httpClient.Get(s.baseURL + "/health")
    if err != nil {
        return &HealthStatus{
            Healthy: false,
            Error:   err.Error(),
            CheckedAt: time.Now(),
        }
    }
    defer resp.Body.Close()

    return &HealthStatus{
        Healthy:    resp.StatusCode == 200,
        StatusCode: resp.StatusCode,
        CheckedAt:  time.Now(),
    }
}
```

#### 14.3.2 降级行为

当健康检查连续失败时：

| 连续失败次数 | 降级行为 |
|-------------|----------|
| 1 次 | 记录警告日志，不影响服务 |
| 3 次（约 90 秒） | 在 Agent 列表页顶部显示全局提示"Agent 服务暂时不稳定" |
| 5 次（约 150 秒） | 禁用新会话创建，现有会话可继续查看历史 |

当健康检查恢复时，自动恢复所有功能。

### 14.4 Agent 错误 Reason 定义

Agent 模块的错误 Reason 遵循现有系统的 `UPPER_SNAKE_CASE` 命名规范，与现有 Reason（如 `ACCOUNT_NOT_FOUND`、`OPS_REPO_UNAVAILABLE`）格式一致，以 `AGENT_` 前缀实现命名空间隔离。

**错误响应 JSON 结构**（复用现有 `Response` 结构体）：

```json
{
    "code": 503,
    "message": "OpenClaw 服务不可用",
    "reason": "AGENT_OPENCLAW_UNAVAILABLE"
}
```

其中 `code` 为 HTTP 状态码（`int32`），`reason` 为业务错误标识（`string`），与现有 `infraerrors.New()` / `response.ErrorFrom()` 体系完全兼容。

**Handler 层使用示例**：

```go
// 与现有错误处理方式完全一致
response.ErrorFrom(c, infraerrors.ServiceUnavailable(
    "AGENT_OPENCLAW_UNAVAILABLE",
    "OpenClaw 服务不可用，请稍后重试",
))
```

| Reason | HTTP 状态码 | 说明 | 前端处理 |
|--------|-----------|------|----------|
| `AGENT_OPENCLAW_UNAVAILABLE` | 503 | OpenClaw 服务不可用 | 显示"服务暂时不可用，请稍后重试" |
| `AGENT_OPENCLAW_TIMEOUT` | 504 | OpenClaw 请求超时 | 显示"请求超时，请重试" |
| `AGENT_SESSION_NOT_FOUND` | 404 | 会话不存在 | 返回列表页 |
| `AGENT_SESSION_NOT_OWNED` | 403 | 会话不属于当前用户 | 显示"无权访问此会话" |
| `AGENT_TEMPLATE_NOT_FOUND` | 404 | Agent 模板不存在 | 返回列表页 |
| `AGENT_TEMPLATE_DISABLED` | 403 | Agent 模板已禁用 | 显示"该 Agent 暂不可用" |
| `AGENT_CONCURRENCY_LIMIT` | 429 | 达到最大活跃会话数 | 显示"已达到最大同时会话数(X)，请关闭不用的会话后重试" |
| `AGENT_RATE_LIMIT` | 429 | 消息发送频率超限 | 显示"消息发送太频繁，请稍后再试"（含 Retry-After 头） |
| `AGENT_API_KEY_INVALID` | 400 | API Key 无效或已禁用 | 提示"请选择有效的 API Key" |
| `AGENT_SESSION_CLOSED` | 400 | 会话已关闭 | 切换为只读模式 |
| `AGENT_SESSION_EXPIRED` | 400 | 会话已过期 | 提示"会话已过期，请创建新会话" |
| `AGENT_IMPORT_FAILED` | 500 | 模板导入失败 | 显示具体错误原因 |
| `AGENT_SYNC_FAILED` | 500 | 模板同步失败 | 显示"同步失败，请稍后重试" |

---

## 十五、会话生命周期管理

### 15.1 会话状态机

Agent 会话遵循以下状态转换：

```
                    ┌──────────┐
         创建会话 →  │  active  │
                    └────┬─────┘
                         │
              ┌──────────┼──────────┐
              │          │          │
              ▼          ▼          ▼
         ┌────────┐ ┌────────┐ ┌────────┐
         │  idle  │ │ closed │ │ error  │
         └───┬────┘ └────────┘ └────────┘
             │
             │ (用户重新发送消息)
             ▼
         ┌────────┐
         │ active  │
         └────────┘
```

| 状态 | 说明 | 触发条件 |
|------|------|----------|
| `active` | 会话活跃中 | 创建会话 / 收到新消息 |
| `idle` | 会话空闲（超过 idle_timeout 无活动） | 超过 1 小时无消息 |
| `closed` | 会话已关闭 | 用户手动关闭 / 自动过期 / 管理员关闭 |
| `error` | 会话异常 | OpenClaw 返回不可恢复错误 |

### 15.2 会话存活时间

| 参数 | 值 | 说明 |
|------|-----|------|
| 默认存活时间 | 24 小时 | 自创建起 24 小时后自动过期 |
| 空闲超时 | 1 小时 | 无消息活动 1 小时后标记为 idle |
| 最大存活时间 | 7 天 | 创建后最多存活 7 天（可配置） |

### 15.3 定时清理任务

`SessionCleanupService` 每 5 分钟执行一次清理循环：

```
清理任务执行流程：

1. 查询所有 status = 'active' 且 last_message_at < now() - 1hour 的会话
   → 更新 status = 'idle'

2. 查询所有 status IN ('active', 'idle') 且 expires_at < now() 的会话
   → 更新 status = 'closed', close_reason = 'expired'
   → 异步调用 OpenClaw API 关闭对应的 OpenClaw 会话

3. 查询所有 status = 'closed' 且 closed_at < now() - 7days 的会话
   → 软删除（设置 deleted_at），释放数据库空间

4. 记录清理日志（清理的会话数量、各状态分布）
```

### 15.4 用户手动关闭 vs 自动过期关闭

| 对比维度 | 用户手动关闭 | 自动过期关闭 |
|----------|-------------|-------------|
| 触发方式 | 用户点击"关闭会话"按钮 / 调用 DELETE API | 定时任务检测 expires_at 超时 |
| close_reason | `user_closed` | `expired` |
| 用户通知 | 即时反馈"会话已关闭" | 下次进入会话时提示"会话已过期" |
| 消息历史 | 保留（可查看历史） | 保留 7 天（可查看历史） |
| 重新开启 | 不支持（需创建新会话） | 不支持（需创建新会话） |
| OpenClaw 同步 | 同步关闭 OpenClaw 侧会话 | 异步关闭 OpenClaw 侧会话 |

### 15.5 会话过期续期

用户每次发送消息时，自动续期会话：

```go
func (s *SessionService) ExtendSession(session *AgentSession) {
    now := time.Now()
    session.LastMessageAt = now
    session.ExpiresAt = now.Add(24 * time.Hour)

    // 如果会话处于 idle 状态，恢复为 active
    if session.Status == "idle" {
        session.Status = "active"
    }
}
```

### 15.6 数据库查询示例

```sql
-- 查询即将过期的会话（用于提前通知用户）
SELECT * FROM agent_sessions
WHERE status IN ('active', 'idle')
  AND expires_at < NOW() + INTERVAL '1 hour'
  AND expires_at > NOW();

-- 查询需要标记为 idle 的会话
SELECT * FROM agent_sessions
WHERE status = 'active'
  AND last_message_at < NOW() - INTERVAL '1 hour';

-- 查询需要关闭的过期会话
SELECT * FROM agent_sessions
WHERE status IN ('active', 'idle')
  AND expires_at < NOW();
```

---

## 十六、并发与速率限制

### 16.1 每用户最大活跃 Agent 会话数

#### 16.1.1 设计决策

| 参数 | 默认值 | 说明 |
|------|--------|------|
| 每用户最大活跃会话数 | 5 | 默认值，可复用 `User.concurrency` 字段或通过配置覆盖 |
| 计数范围 | 仅统计 status IN ('active', 'idle') 的会话 | 已关闭、已过期的会话不计入 |
| 超限响应 | 429 + 错误码 `AGENT_CONCURRENCY_LIMIT` | 前端提示用户关闭不用的会话 |

#### 16.1.2 实现方式

复用现有 `User.concurrency` 字段（`backend/ent/schema/user.go` 已存在，默认 5，与 `config.yaml` 的 `default.user_concurrency: 5` 一致），可用 `agent.max_active_sessions` 配置覆盖；未配置时回退到 `User.concurrency`。

```go
func (s *RateLimitService) CheckConcurrency(userId int) error {
    limit := s.getUserConcurrencyLimit(userId) // 从 User.concurrency 或配置读取

    activeCount, err := s.db.AgentSession.Query().
        Where(
            agentsession.UserIDEQ(userId),
            agentsession.StatusIn("active", "idle"),
        ).
        Count(ctx)

    if err != nil {
        return err
    }

    if activeCount >= limit {
        return infraerrors.TooManyRequests(
            "AGENT_CONCURRENCY_LIMIT",
            fmt.Sprintf("已达到最大同时会话数(%d)，请关闭不用的会话后重试", limit),
        )
    }

    return nil
}
```

### 16.2 Agent 消息发送频率限制

#### 16.2.1 设计决策

| 参数 | 默认值 | 说明 |
|------|--------|------|
| 每分钟最大消息数 | 10 | 每个会话每分钟最多发送 10 条消息 |
| 限流维度 | 按会话（session_id） | 不同会话独立计数 |
| 限流实现 | Redis 滑动窗口 | 使用 Redis Sorted Set 精确计数 |
| 超限响应 | 429 + 错误码 `AGENT_RATE_LIMIT` + `Retry-After` 头 | 前端显示等待时间 |

#### 16.2.2 Redis 滑动窗口实现

```go
func (s *RateLimitService) CheckMessageRate(userId int, sessionId int) error {
    ctx := context.Background()
    key := fmt.Sprintf("agent:rate:%d:%d", userId, sessionId)
    now := time.Now().UnixMilli()
    window := now - 60000 // 1 分钟窗口

    pipe := s.redis.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(window, 10))
    pipe.ZCard(ctx, key)
    pipe.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: now})
    pipe.Expire(ctx, key, 60*time.Second)

    cmds, err := pipe.Exec(ctx)
    if err != nil {
        return err
    }

    count := cmds[1].(*redis.IntCmd).Val()
    limit := s.config.Agent.MessageRateLimit // 默认 10

    if count >= int64(limit) {
        return infraerrors.TooManyRequests(
            "AGENT_RATE_LIMIT",
            fmt.Sprintf("消息发送太频繁，请稍后再试（每分钟最多 %d 条）", limit),
        ).WithMetadata(map[string]string{"retry_after_seconds": "60"})
    }

    return nil
}
```

> 说明：`Retry-After` 响应头由 Handler 层读取 `metadata.retry_after_seconds`（或配置）后通过 `c.Header("Retry-After", ...)` 单独设置；现有 `ApplicationError` 通过 `Metadata` 承载附加信息，不直接携带 HTTP 头。

### 16.3 资源预估

#### 16.3.1 假设条件

- 100 并发 Agent 会话
- 每个会话对应 1 个 OpenClaw 进程（或轻量级 goroutine）
- 每个 Agent 会话平均处理 10 轮对话
- 每轮对话涉及 2-5 次 LLM 调用（Agent 思考 + Skill 调用 + 最终回复）

#### 16.3.2 内存预估

| 组件 | 单会话内存 | 100 会话总内存 | 说明 |
|------|-----------|---------------|------|
| OpenClaw 进程 | ~50MB | ~5GB | Agent 执行内核 |
| sub2api 业务层 | ~5MB | ~500MB | Go gin 协程 + 连接池 |
| Redis 缓存 | ~1MB | ~100MB | 会话状态 + 速率限制数据 |
| **总计** | **~56MB** | **~5.6GB** | |

#### 16.3.3 CPU 预估

| 组件 | 单会话 CPU | 100 会话总 CPU | 说明 |
|------|-----------|---------------|------|
| OpenClaw 进程 | 0.1-0.3 核 | 10-30 核 | Agent 逻辑处理（非 LLM 推理） |
| sub2api 业务层 | 0.05 核 | 5 核 | HTTP 代理 + 业务逻辑 |
| LLM 推理 | 不占用本地 CPU | 0 | 调用远程 LLM API |
| **总计** | **0.15-0.35 核** | **15-35 核** | |

#### 16.3.4 建议配置

| 并发规模 | 建议服务器配置 | 预估支持用户数 |
|----------|---------------|---------------|
| 50 会话 | 4 核 / 8GB | ~200-500 用户 |
| 100 会话 | 8 核 / 16GB | ~500-1000 用户 |
| 200 会话 | 16 核 / 32GB | ~1000-2000 用户 |
| 500 会话 | 32 核 / 64GB | ~5000+ 用户 |

> **注意**：以上为 OpenClaw 进程资源预估，LLM 推理资源由上游 API 提供，不计入本地服务器。实际资源消耗取决于 Agent 模板的复杂度和 Skill 调用频率。

---

## 十七、实施阶段

| 阶段 | 内容 | 预估工作量 |
|------|------|-----------|
| **P1** | 后端：ent schema + OpenClawService + API 路由 + Wire 注入 | 3-4 天 |
| **P2** | 前端：Agent 列表页 + 聊天页 + 导航入口 + i18n | 2-3 天 |
| **P3** | 管理端：模板同步 + 可见性管理 + 会话查看 | 1-2 天 |
| **P4** | OpenClaw 部署：安装 + 模板创建 + 内网配置 + 联调 | 1-2 天 |
| **P5** | 会话生命周期：idle 检测 + 过期清理 + 状态机 | 1-2 天 |
| **P6** | 并发与速率限制：Redis 滑动窗口 + 并发检查 | 1 天 |
| **P7** | 错误处理与容错：重试机制 + 降级策略 + 错误码 | 1-2 天 |
| **P8** | 安全加固 + 测试 + 文档 | 1-2 天 |

---

## 十八、关键技术决策

**Q: 为什么不直接暴露 OpenClaw API 给前端？**

A: 安全要求。OpenClaw 没有用户鉴权能力，直接暴露等于任何人都能创建会话。业务层做鉴权转发，强制绑定 userId-session-agentId。

**Q: 为什么用用户的 API Key 而不是系统统一 Key？**

A: 计费隔离。每个用户的 Agent 消耗计入自己的余额，复用现有 /v1 计费逻辑，无需开发额外计费模块。用户可以在会话中看到自己的 Token 消耗。

**Q: OpenClaw 模板更新后如何同步？**

A: 管理员手动触发同步（`POST /admin/agents/sync`）或定时任务自动同步。同步只更新模板元数据（名称/描述/Skill），不影响已创建的会话。外部来源的模板更新需要管理员审查 changelog 后手动确认拉取，不自动更新。

**Q: 为什么设置 24 小时会话过期时间？**

A: 平衡用户体验和资源利用率。24 小时覆盖一个完整工作日，用户无需频繁创建会话；同时避免僵尸会话长期占用 OpenClaw 进程资源。用户每次发送消息时自动续期。

**Q: 为什么限制每用户最多 5 个活跃会话？**

A: 防止单个用户滥用资源。5 个会话足够覆盖日常使用场景（报告分析、代码开发、数据处理、财务分析、内容创作），同时控制 OpenClaw 进程数量。管理员可为特定用户调整此限制。

---

## 十九、现状核对与差距分析

本章基于 sub2api 代码库现状（`backend/ent/schema/`、`backend/internal/`、`backend/config.yaml`、`frontend/package.json`）逐项核对本方案，标注「已存在 / 需新增」，作为实施前的基线参考。

### 19.1 核对结论总表

| 设计项 | 现状位置 | 状态 | 说明 |
|--------|----------|------|------|
| 用户并发字段 | `backend/ent/schema/user.go` → `concurrency`（默认 5） | 已存在 | 16.1 可直接复用 |
| 默认并发配置 | `backend/config.yaml` → `default.user_concurrency: 5` | 已存在 | 与 `User.concurrency` 默认一致 |
| 现有限流配置 | `backend/config.yaml` → `rate_limit.requests_per_minute: 60`、`burst_size: 10` | 已存在 | 平台级 /v1 限流，非 Agent 会话级限流 |
| 错误类型体系 | `backend/internal/pkg/errors`（`New` / `ServiceUnavailable` / `TooManyRequests` / `ApplicationError`） | 已存在 | 14.4 复用 |
| 响应封装 | `backend/internal/pkg/response`（`Response` / `ErrorFrom` / `Paginated`） | 已存在 | 第四章接口复用 |
| 依赖注入 | `backend/cmd/server/wire_gen.go`（`ConcurrencyService` 等） | 已存在 | Agent 服务需新增注入 |
| 前端技术栈 | `frontend/package.json`（Vue 3.4 / vue-router 4 / pinia 2 / vue-i18n 9 / chart.js 4 / marked 17） | 已就绪 | 第六章页面可落地 |
| `openclaw:` / `agent:` 配置块 | `backend/config.yaml` | 需新增 | 附录 A 全量新增 |
| Agent 路由 | `backend/internal/server/routes/` | 需新增 | 现有仅 auth / user / admin / gateway / payment / common |
| Agent Handler | `backend/internal/handler/` | 需新增 | 现有无 agent 相关 handler |
| Agent 服务 | `backend/internal/service/` | 需新增 | OpenClawService / SessionCleanupService / RateLimitService |
| `AgentTemplate` / `AgentSession` / `AgentTemplateI18n` | `backend/ent/schema/` | 需新增 | 三张新表 + 迁移 |
| `UsageLog.agent_session_id` / `agent_template_id` | `backend/ent/schema/usage_log.go` | 需新增 | 当前无 `metadata` 字段，需新增两列（见 8.2） |
| 前端 Agent 页面 / 导航 / i18n | `frontend/src/views/`、`AppSidebar.vue`、`locales/` | 需新增 | 第六章全新页面与 `agents` 命名空间 |

### 19.2 关键差异说明

1. **错误类型已统一为 `ApplicationError`**：方案 16.1.2 / 16.2.2 原稿使用了不存在的 `AppError` 结构体，已改为 `infraerrors.TooManyRequests(reason, message)`，与 14.4 一致。`ApplicationError` 通过 `Metadata` 承载附加信息，不直接携带 HTTP 头（如 `Retry-After` 需在 Handler 层单独设置）。
2. **UsageLog 无 `metadata` 字段**：8.2 原稿「利用现有 metadata 字段」不成立，已改为「新增 `agent_session_id` / `agent_template_id` 两列」。`UsageLog` 为只追加表，新增列采用 `Optional().Nillable()`。
3. **并发限制默认值已存在**：`User.concurrency`（默认 5）与 `default.user_concurrency: 5` 均已在代码中，16.1 无需新建字段，`agent.max_active_sessions` 仅作可选的全局覆盖。
4. **会话级限流为新增能力**：现有 `rate_limit` 为平台/API Key 维度，16.2 的「按会话」滑动窗口限流需新增 Redis 实现。
5. **配置需扩展**：附录 A 的 `openclaw:` / `agent:` 为新增顶层块，需同步扩展 `internal/config` 配置结构体与 YAML 绑定。

### 19.3 实施基线

- 无需改动现有 /v1 网关计费链路，Agent LLM 消耗复用现有鉴权、计费、Token 池调度。
- 唯一对现有表的改动是 `UsageLog` 追加两列（向后兼容，NULL 默认）。
- 新增三张 ent 表、三个 service、一组 Agent 路由/Handler、前端页面与 i18n 命名空间。
- 实施阶段划分与工作量估算见第十七章，无变化。

---

## 附录 A：统一配置参考

> 说明：本节 `openclaw:` 与 `agent:` 为**新增配置块**。当前 `backend/config.yaml` 仅有 `server` / `database` / `redis` / `jwt` / `default` / `rate_limit` / `timezone` / `security` 八个顶层块。实施时需在 `internal/config` 配置结构体中新增对应字段并绑定 YAML 标签；`agent.max_active_sessions` 默认 5 与现有 `default.user_concurrency: 5`、`User.concurrency` 默认值保持一致。

### A.1 完整 openclaw 配置块

```yaml
openclaw:
  # 基础配置
  enabled: true
  base_url: "http://127.0.0.1:9090"          # OpenClaw Gateway 内网地址
  admin_token: "your-internal-token"           # OpenClaw 管理 API Token
  sync_interval: 300                            # 模板同步间隔（秒）
  llm_base_url: "http://127.0.0.1:8080/v1"     # 回调 sub2api 的 /v1 地址
  timeout: 600                                  # 请求超时（秒，Agent 可能长时间运行）
  idle_timeout: 3600                            # 会话空闲超时（秒，默认 1 小时）
  session_ttl: 86400                            # 会话存活时间（秒，默认 24 小时）
  max_session_ttl: 604800                       # 最大存活时间（秒，默认 7 天）

  # 健康检查
  health_check:
    interval: 30                                # 检查间隔（秒）
    degrade_after: 3                            # 连续失败 N 次后开始降级
    disable_after: 5                            # 连续失败 N 次后禁用新会话

  # 重试配置
  retry:
    max_retries: 3                              # 最大重试次数
    initial_backoff: 1                          # 初始退避时间（秒）
    backoff_multiplier: 2                       # 退避倍数
    max_backoff: 10                             # 最大退避时间（秒）
    total_timeout: 30                           # 总超时时间（秒）

  # 模板来源管理
  template_sources:
    github:
      enabled: true
      check_interval: 3600                      # 每小时检查一次更新（秒）
      default_repos:                            # 预置仓库（管理端可一键导入）
        - url: "https://github.com/glowElephant/awesome-openclaw-agents"
          name: "awesome-openclaw-agents"
          description: "社区 Agent 模板集合"
        - url: "https://github.com/partme-ai/claw-agents"
          name: "claw-agents"
          description: "企业向 Agent 模板"
    clawhub:
      enabled: true
      base_url: "https://hub.openclaw.ai"
      check_interval: 86400                     # 每天检查一次更新（秒）
    agentpacks:
      enabled: true
      base_url: "https://www.agentpacks.ai"
      check_interval: 86400
    official:
      enabled: true
      base_url: "https://openclaw-ai.dev"
      check_interval: 3600
    dropspace:
      enabled: true
      base_url: "https://dropspace.dev"
      check_interval: 86400
    hermes:
      enabled: true
      base_url: "https://hermesatlas.com"
      check_interval: 86400
      auto_convert: true                        # 自动转换 Hermes → OpenClaw 格式
    local:
      enabled: true                             # 自定义模板始终可用
    custom_import:
      enabled: true                             # 自定义文件导入始终可用

  # OpenClaw 服务器管理（用于远程导入模板）
  ssh:
    host: "127.0.0.1"
    port: 22
    user: "root"
    key_path: "/etc/sub2api/ssh_key"           # SSH 私钥路径
    agents_dir: "/root/.openclaw/agents"
    config_file: "/root/.openclaw/openclaw.json"
    restart_command: "systemctl restart openclaw"

# Agent 业务配置（独立于 openclaw 配置块）
agent:
  max_active_sessions: 5                        # 每用户最大活跃会话数
  message_rate_limit: 10                        # 每会话每分钟最大消息数
  cleanup_interval: 300                         # 会话清理任务间隔（秒）
```

### A.2 配置项说明

#### 基础配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `openclaw.enabled` | bool | true | 是否启用 Agent 功能 |
| `openclaw.base_url` | string | -- | OpenClaw Gateway 内网地址 |
| `openclaw.admin_token` | string | -- | OpenClaw 管理 API Token |
| `openclaw.sync_interval` | int | 300 | 模板同步间隔（秒） |
| `openclaw.llm_base_url` | string | -- | 回调 sub2api /v1 的地址 |
| `openclaw.timeout` | int | 600 | 请求超时（秒） |
| `openclaw.idle_timeout` | int | 3600 | 会话空闲超时（秒） |
| `openclaw.session_ttl` | int | 86400 | 会话存活时间（秒） |
| `openclaw.max_session_ttl` | int | 604800 | 最大存活时间（秒） |

#### 健康检查与降级

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `openclaw.health_check.interval` | int | 30 | 检查间隔（秒） |
| `openclaw.health_check.degrade_after` | int | 3 | 连续失败 N 次后开始降级 |
| `openclaw.health_check.disable_after` | int | 5 | 连续失败 N 次后禁用新会话 |

#### 重试配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `openclaw.retry.max_retries` | int | 3 | 最大重试次数 |
| `openclaw.retry.initial_backoff` | int | 1 | 初始退避时间（秒） |
| `openclaw.retry.backoff_multiplier` | int | 2 | 退避倍数 |
| `openclaw.retry.max_backoff` | int | 10 | 最大退避时间（秒） |
| `openclaw.retry.total_timeout` | int | 30 | 总超时时间（秒） |

#### 模板来源

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `openclaw.template_sources.<source>.enabled` | bool | true | 是否启用该来源 |
| `openclaw.template_sources.<source>.check_interval` | int | -- | 更新检查间隔（秒） |
| `openclaw.template_sources.<source>.base_url` | string | -- | 来源 API 地址 |
| `openclaw.template_sources.github.default_repos` | array | -- | 预置 GitHub 仓库列表 |
| `openclaw.template_sources.hermes.auto_convert` | bool | true | 是否自动转换 Hermes 格式 |

#### SSH 管理

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `openclaw.ssh.host` | string | 127.0.0.1 | OpenClaw 服务器地址 |
| `openclaw.ssh.port` | int | 22 | SSH 端口 |
| `openclaw.ssh.user` | string | root | SSH 用户 |
| `openclaw.ssh.key_path` | string | -- | SSH 私钥路径 |
| `openclaw.ssh.agents_dir` | string | /root/.openclaw/agents | Agent 模板目录 |
| `openclaw.ssh.config_file` | string | /root/.openclaw/openclaw.json | 配置文件路径 |
| `openclaw.ssh.restart_command` | string | -- | 重启命令 |

#### Agent 业务配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `agent.max_active_sessions` | int | 5 | 每用户最大活跃会话数 |
| `agent.message_rate_limit` | int | 10 | 每会话每分钟最大消息数 |
| `agent.cleanup_interval` | int | 300 | 会话清理任务间隔（秒） |

---

## 附录 B：术语表

| 术语 | 英文 | 定义 |
|------|------|------|
| OpenClaw | OpenClaw | Agent 执行内核，部署在 Ubuntu 服务器上，仅内网访问。负责 Agent 模板管理（SOUL.md + AGENTS.md + TOOLS.md）、会话执行、Skill 调用和 LLM 调度。通过 HTTP Gateway API 与 sub2api 业务层通信。 |
| sub2api | sub2api | 本项目，LLM API 网关 + 业务平台。负责用户认证、API Key 管理、Agent 会话管理、模板来源管理、计费、监控等所有业务逻辑。 |
| Hermes | Hermes | Nous Research 开发的自主学习型 Agent 框架（2026年2月发布）。与 OpenClaw 模板格式相近（都基于 Markdown），可通过格式转换导入到 OpenClaw 中使用。 |
| ClawHub | ClawHub | OpenClaw 官方技能市场（hub.openclaw.ai），提供 5,400+ skills 供下载。模板以 .skill 归档包形式分发。 |
| AgentPacks | AgentPacks | 多平台 Agent 模板包市场（agentpacks.ai），同时为 OpenClaw 和 Hermes 提供模板包。 |
| DropSpace | DropSpace | 社区 Agent 模板市场（dropspace.dev），提供社区贡献的 Agent 模板。 |
| SOUL.md | SOUL.md | Agent 角色定义文件，定义 Agent 的性格、行为准则、安全约束和回复风格。是 Agent 模板的核心配置文件。 |
| AGENTS.md | AGENTS.md | Agent 执行流程定义文件，定义 Agent 的 Skill 调用逻辑、工作流程和决策规则。 |
| TOOLS.md | TOOLS.md | Agent 可用工具白名单文件，定义 Agent 可以调用的系统命令、API 和外部工具列表。 |
| Agent 模板 | Agent Template | 包含 SOUL.md + AGENTS.md + TOOLS.md 的完整 Agent 定义。部署在 OpenClaw 的 `/root/.openclaw/agents/` 目录下，通过 `openclaw.json` 注册。 |
| Agent 会话 | Agent Session | 用户选中模板后创建的独立对话实例。每个会话绑定一个用户、一个 Agent 模板和一个 API Key，具有独立的状态和生命周期。 |
| API Key | API Key | 用户在 sub2api 平台创建的密钥，用于 LLM 调用鉴权和计费。Agent 会话创建时绑定用户的一个 active API Key，OpenClaw 使用该 Key 回调 /v1 网关。 |
| /v1 网关 | /v1 Gateway | sub2api 的 LLM API 代理端点，负责 API Key 鉴权 → 余额检查 → Token 池调度 → 计费 → 上游 LLM 转发。OpenClaw 的 LLM 调用通过此端点完成，实现计费复用。 |
| 会话状态机 | Session State Machine | Agent 会话的状态转换模型：active（活跃）→ idle（空闲）→ closed（关闭）/ error（错误）。由 SessionCleanupService 定时维护。 |
| 指数退避 | Exponential Backoff | 重试策略：每次重试前等待时间按倍数递增（1s → 2s → 4s），避免对故障服务造成冲击。 |
| 滑动窗口 | Sliding Window | 速率限制算法：使用 Redis Sorted Set 记录时间窗口内的请求数，精确控制消息发送频率。 |
| 降级策略 | Degradation Strategy | 当 OpenClaw 不可用时，业务层逐步限制功能（警告 → 禁用新会话），保证核心系统不受影响。 |
| agent_session_id | Agent Session ID | `AgentSession` 的 DB 自增主键，随 OpenClaw 的 LLM 调用注入 /v1 网关，写入 `UsageLog.agent_session_id` 列，用于关联 UsageLog 和 AgentSession，实现 Token 消耗追踪。 |