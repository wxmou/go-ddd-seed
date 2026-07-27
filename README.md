# go-ddd-seed

基于 **Go 语言** 的 DDD（领域驱动设计）+ CQRS（命令查询职责分离）脚手架项目，采用六边形架构（Hexagonal Architecture / Ports and Adapters），开箱即用地指导 AI Agent 遵循架构生成代码。

## 架构概览

```
┌──────────────────────────────────────────────────────────────┐
│                     Trigger Layer（入站适配器）                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐  │
│  │  HTTP 控制器   │  │  定时任务 Job  │  │  Event Listener   │  │
│  └──────┬───────┘  └──────┬───────┘  └────────────────────┘  │
├─────────┼──────────────────┼──────────────────────────────────┤
│               Application Layer（应用层 — 用例编排）            │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌────────────────────┐  │
│  │ CommandHandler│  │ QueryService  │  │   EventBus 端口     │  │
│  │ 命令处理器     │  │ 查询服务       │  │  事件总线接口        │  │
│  └──────┬───────┘  └──────┬───────┘  └────────────────────┘  │
│  ┌──────┴───────┐  ┌──────┴───────┐                            │
│  │  port/repo   │  │  port/repo   │                            │
│  │ 写仓储接口     │  │ 读仓储接口    │                            │
│  │ (应用层端口)   │  │ (DTO 返回)   │                            │
│  └──────┬───────┘  └─────────────┘                            │
├─────────┼──────────────────────────────────────────────────────┤
│               Domain Layer（领域层 — 核心业务逻辑）              │
│  ┌──────┴──────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │  Domain Model    │  │  port/repo  │  │  Domain Event     │  │
│  │  聚合根 / 实体    │  │ 领域仓储接口  │  │  领域事件          │  │
│  └─────────────────┘  └─────────────┘  └──────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│             Infrastructure Layer（基础设施层 — 出站适配器）     │
│  ┌──────────┐ ┌─────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │ GORM 实现  │ │  Redis   │ │  MinIO   │ │  Watermill MQ    │  │
│  │ 仓储适配器  │ │  缓存    │ │ 文件存储  │ │  消息中间件       │  │
│  └──────────┘ └─────────┘ └──────────┘ └──────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

### 关键设计原则

| 原则 | 说明 |
|------|------|
| **依赖倒置** | 领域层定义仓储接口，基础设施层实现；领域层不依赖任何外部框架 |
| **CQRS 分离** | 命令侧使用领域模型（聚合根 + 实体），查询侧使用 DTO，各走独立仓储接口 |
| **聚合根封装** | 所有业务操作通过聚合根方法完成，禁止直接修改内部状态 |
| **领域事件** | 内建 `AggregateRoot` 基类支持领域事件记录，事件总线异步发布 |
| **六边形架构** | 层间通过接口（Port）通信，适配器（Adapter）实现具体技术细节 |

## 特色功能 — AI 友好架构（Vibe Coding Ready）

本项目内置了 **Claude Code 的 DDD 编码约束 Skill**，让 AI Agent 在生成代码时自动遵循领域驱动设计的六边形架构规范，实现真正的 Vibe Coding 体验。

### 内置 Skill：`domain-driven-design`

当你在 Claude Code 中使用本项目的 DDD Skill 时，Agent 会自动获知：

- **六边形架构分层规则** — 严格区分 Trigger / Application / Domain / Infrastructure 层，不跨层引用
- **CQRS 约束** — 命令侧引用领域仓储，查询侧引用读仓储，CommandHandler 不可注入读仓储（认证流程特例除外）
- **聚合根封装** — 禁止直接修改聚合根字段，所有操作通过领域方法完成
- **依赖方向** — 领域层不依赖任何外部框架，基础设施层实现领域层接口
- **文件布局规范** — 每层包命名、文件组织方式、端口与适配器分离规则

### Vibe Coding 工作流

```
1. 描述业务需求 → "我想加一个会员等级功能"
2. Agent 自动识别需要创建：模型、仓储接口、命令/查询、处理器、控制器
3. 按 DDD 架构在正确位置生成代码，自动遵守依赖规则
4. 生成测试并验证架构合规性
```

### 触发方式

在 Claude Code 会话中提及以下关键词即可自动激活 Skill：

- `trigger/listener`、`application/command`、`domain/port`、`domain/model` 等架构路径
- 或直接描述 DDD 相关任务（"新增一个聚合根"、"添加一个领域事件"）

### 配套的工程项目 Skill

以下 Skills 随项目一同提供，共同构成完整的 AI 辅助开发体验：

| Skill | 说明 |
|-------|------|
| `domain-driven-design` | DDD 六边形架构 + CQRS 编码约束，指导 Agent 遵循架构生成代码 |
| `code-review` | 对当前代码变更进行正确性审查和简化建议 |
| `verify` | 运行应用并观察行为，验证代码变更是否生效 |
| `security-review` | 对挂起的变更完成安全审查 |

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | **Go 1.25** |
| Web 框架 | **Gin** |
| ORM | **GORM** |
| 依赖注入 | **Google Wire** |
| 数据库 | **PostgreSQL 16**（主库），**MySQL**（兼容支持） |
| 缓存 | **Redis 7** |
| 消息队列 | **Watermill** + Redis Stream |
| 文件存储 | **MinIO** / 本地文件系统（策略模式切换） |
| 认证鉴权 | **JWT**（Access + Refresh Token 双令牌） |
| 定时任务 | **Robfig Cron**（内嵌式调度器） |
| 日志 | **Zerolog** |
| 配置管理 | **Viper** |
| 参数校验 | **Validator v10**（i18n 中文错误提示） |
| 文档 | **Swagger**（自动生成 API 文档） |
| 测试 | Go 标准测试 + 表驱动测试 |
| 前端 | **Vue 3 + Element Plus + Vite + TypeScript** |

## 项目结构

```
go-ddd-seed/
├── cmd/
│   ├── api/                    # API 服务入口
│   │   ├── main.go             # 启动入口
│   │   ├── wire.go             # Wire 依赖注入声明
│   │   └── wire_gen.go         # Wire 自动生成代码
│   ├── worker/                 # Worker 服务入口（后台任务）
│   └── checkconn/              # 数据库连接检查工具
├── internal/
│   ├── trigger/                # 入站适配器（驱动层）
│   │   ├── http/               # HTTP 控制器 + 请求/响应定义
│   │   │   ├── controller/     # Gin 控制器（路由注册 & 请求处理）
│   │   │   ├── req/            # 请求参数结构体
│   │   │   └── resp/           # 响应结构体
│   │   └── job/                # 定时任务（Cron Job）
│   ├── application/            # 应用层（用例编排）
│   │   ├── command/            # 命令定义（Command 对象）
│   │   ├── commandHandler/     # 命令处理器（写操作）
│   │   ├── queryService/       # 查询服务（读操作）
│   │   ├── event/              # 事件总线端口（接口定义）
│   │   └── port/
│   │       ├── repo/           # 读仓储接口（返回 DTO）
│   │       └── thirdPartyApi/  # 第三方服务端口
│   ├── domain/                 # 领域层（核心业务逻辑）
│   │   ├── model/              # 领域模型（聚合根、实体）
│   │   │   ├── aggregate_root.go    # 聚合根基类（领域事件支持）
│   │   │   ├── user/           # 用户聚合（示例）
│   │   │   ├── role/           # 角色聚合（示例）
│   │   │   ├── permission/     # 权限聚合（示例）
│   │   │   ├── dict_type/      # 字典聚合（示例）
│   │   │   ├── kv_config/      # KV 配置聚合（示例）
│   │   │   └── file_record/    # 文件记录聚合（示例）
│   │   ├── event/              # 领域事件基类（事件接口定义）
│   │   ├── port/repo/          # 领域仓储接口（命令侧）
│   │   └── error.go            # 领域层错误定义
│   └── infrastructure/         # 基础设施层（出站适配器）
│       ├── adapter/
│       │   ├── repo/           # 仓储实现（GORM）
│       │   └── thirdPartyApi/
│       │       └── storage/    # 文件存储适配器（本地/MinIO）
│       ├── event/              # 事件总线实现（内存）
│       ├── messaging/          # 消息队列（Watermill + Redis Stream）
│       ├── scheduler/          # 定时任务调度器（Cron）
│       └── persistence/        # 持久化基础设施（DB/Redis 连接）
├── pkg/                        # 公共工具包（可被外部引用）
│   ├── auth/                   # JWT 认证（Token 签发/解析/验证）
│   ├── config/                 # 配置管理
│   ├── crypto/                 # 密码加密
│   ├── errors/                 # 应用层错误定义（含 HTTP 状态码映射）
│   ├── logger/                 # 日志封装
│   ├── middleware/             # Gin 中间件（认证/审计/CORS/日志/链路追踪）
│   ├── utils/                  # 响应工具函数
│   └── validator/              # 参数校验器（中文化错误信息）
├── configs/                    # 配置文件
├── docs/                       # Swagger 文档
├── docker-compose.yml          # 基础设施服务（PostgreSQL + Redis + MinIO）
└── Makefile                    # 统一构建入口
```

## 示例业务模块（Bounded Context）

以下为脚手架附带的示例模块，展示 DDD + CQRS 的最佳实践，可直接作为新业务模块的开发模板：

| 模块 | 聚合根 | 说明 |
|------|--------|------|
| **认证 (Auth)** | `User` | 注册/登录/登出/刷新令牌，JWT 双令牌机制 |
| **RBAC** | `Role`, `Permission` | 角色-权限管理，用户-角色分配 |
| **字典 (Dict)** | `DictType`, `DictEntry` | 字典类型与字典项管理 |
| **KV 配置** | `KVConfig` | Key-Value 配置管理 |
| **文件存储** | `FileRecord` | 文件上传/下载/记录，支持本地存储和 MinIO |
| **审计日志** | 无（纯记录） | 操作审计日志记录与查询 |

## 快速启动

### 前置条件

- Go 1.25+
- Node.js 18+
- Docker & Docker Compose

### 启动步骤

```bash
# 1. 启动基础设施（PostgreSQL + Redis + MinIO）
make docker-up

# 2. 启动后端（热重载，使用 Air）
make dev

# 3. 启动前端（热重载）
make dev-admin

# 4. 运行测试
make test

# 5. 生成 Wire 依赖注入代码
make wire
```

### API 文档

启动后端后访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

## CQRS 详解

### 命令侧（Command / Write Side）

```
HTTP POST /api/v1/auth/register
  → Controller 接收请求，校验参数
  → 构建 RegisterCommand 命令对象
  → CommandHandler.Register(ctx, cmd)
    → 调用领域仓储接口查询/加载聚合根
    → 调用聚合根方法执行业务逻辑（如 user.NewUser()）
    → 调用领域仓储接口保存聚合根
    → 返回 UserDTO 给 Controller
```

### 查询侧（Query / Read Side）

```
HTTP GET /api/v1/roles
  → Controller 接收请求
  → 调用 QueryService.List(ctx)
    → 调用读仓储接口（ReadRepository）直接查询 DTO
    → 返回 []*RoleDTO 给 Controller
```

### 关键区别

| 维度 | 命令侧 | 查询侧 |
|------|--------|--------|
| 模型 | 领域模型（聚合根 + 实体） | DTO（数据传输对象） |
| 仓储 | `domain/port/repo/` 接口 | `application/port/repo/` 接口 |
| 返回值 | 领域模型对象 | DTO（仅为读优化） |
| 业务逻辑 | 聚合根方法封装 | 无业务逻辑，只做查询组合 |
| 副作用 | 有（修改数据库状态） | 无（纯查询） |

## 领域事件

聚合根内嵌 `model.AggregateRoot` 基类即可记录领域事件：

```go
type User struct {
    model.AggregateRoot  // 继承领域事件记录能力
    // ... 字段
}

// 在业务方法中触发事件
func (u *User) ChangePassword(newHash string) {
    u.PasswordHash = newHash
    u.UpdatedAt = time.Now()
    u.AddDomainEvent(&event.PasswordChanged{
        BaseEvent: event.BaseEvent{Name: "user.password.changed", At: time.Now()},
        UserID:    u.ID,
    })
}
```

事件发布链路：`Domain Model → AggregateRoot → CommandHandler → EventBus → Listener`

## 依赖注入

项目使用 **Google Wire** 进行编译期依赖注入，所有依赖关系在 `wire.go` 中声明：

```go
// cmd/api/wire.go
func InitializeApp(cfg *config.Config) (*App, func(), error) {
    panic(wire.Build(
        appSet,           // 基础设施
        controllerSet,    // 控制器
        handlerSet,       // 命令处理器
        serviceSet,       // 查询服务
        repoSet,          // 仓储实现
    ))
}
```

修改依赖后运行 `make wire` 重新生成代码。

## 配置

配置文件位于 `configs/config.yaml`，支持以下环境变量覆盖：

| 环境变量 | 说明 |
|---------|------|
| `DB_HOST` | 数据库主机 |
| `DB_PORT` | 数据库端口 |
| `DB_USER` | 数据库用户名 |
| `DB_PASSWORD` | 数据库密码 |
| `DB_NAME` | 数据库名 |
| `REDIS_ADDR` | Redis 地址 |
| `JWT_SECRET` | JWT 密钥 |
| `STORAGE_DRIVER` | 存储驱动（local / minio） |

## 构建与部署

```bash
# 生产构建
make build

# 前端构建
make build-admin

# Docker 部署
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 自定义指南

### 1. 修改 Go Module 名称

```bash
# 将 go.mod 中的 module 名替换为你的项目名
go mod edit -module github.com/your-org/your-project

# 替换所有 import 路径
find . -type f -name '*.go' -exec sed -i 's|github.com/go-ddd-seed/go-ddd-seed|github.com/your-org/your-project|g' {} +
```

### 2. 新增业务模块

参考 `user` 聚合的完整结构，按以下步骤添加新模块：

1. `domain/model/{module}/` — 定义聚合根、实体、值对象
2. `domain/port/repo/` — 定义领域仓储接口
3. `application/command/` — 定义命令对象
4. `application/commandHandler/` — 实现命令处理器
5. `application/queryService/` — 实现查询服务
6. `application/port/repo/` — 定义读仓储接口（DTO）
7. `infrastructure/adapter/repo/` — 实现仓储适配器（GORM）
8. `trigger/http/controller/` — 实现 HTTP 控制器
9. `trigger/http/req/` 和 `trigger/http/resp/` — 定义请求/响应结构体
10. `cmd/api/wire.go` — 注册依赖注入