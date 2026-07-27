---
name: domain-driven-design
description: Architecture guide for Hexagonal DDD, CQRS, and domain modeling. Triggers when starting projects, partitioning bounded contexts/aggregates, or when project context contains 'trigger/listener', 'application/command', 'domain/port', or 'domain/model' structures.
---

# Domain-Driven Design — 代码组织架构规范

严格遵循六边形架构 + 充血模型 + CQRS，聚焦代码组织与分层约束，适用于任何后端语言。

## 架构总览

```
trigger → application → domain ← infrastructure
```

| 层 | 职责 | 边界 |
|---|---|---|
| **trigger** | 入站适配器：HTTP / gRPC / Job / MQ | 接收外部请求，转换协议，委派给应用层 |
| **application** | 用例编排：Command / CommandHandler / QueryService | 不含业务逻辑，只做流程编排 |
| **domain** | 核心业务：聚合根、实体、值对象、领域服务、仓储端口（Port） | 零外部依赖，不引用任何框架 |
| **infrastructure** | 出站适配器：仓储实现、防腐层实现、DB/缓存/消息 | 实现 domain 端口，屏蔽技术细节 |

## 目录结构

```
{项目根目录}/
├── trigger/                          # 入站适配器层
│   ├── http/
│   │   ├── controller/               # HTTP 控制器
│   │   ├── req/                      # 请求 DTO
│   │   └── resp/                     # 响应 DTO
│   ├── rpc/                          # gRPC 服务实现
│   ├── job/                          # 定时任务触发器
│   └── listener/                     # 异步事件/消息监听器（替代旧 mq/ 命名）
│
├── application/                      # 应用层
│   ├── command/                      # 命令（写操作）定义
│   ├── commandHandler/               # 命令处理器
│   ├── queryService/                 # 查询服务（读操作）
│   ├── event/                        # 事件总线端口（仅 Publish）
│   └── port/                         # 读仓储端口（接口） + 读 DTO（CQRS 读模型）
│       ├── repo/                     # 读仓储端口 — 返回 DTO
│       └── thirdPartyApi/            # 纯技术支撑/外围服务端口（如文件存储、短信发送）
│
├── domain/                           # 领域层 — 零外部依赖
│   ├── event/                        # 领域事件契约（领域事件接口定义，叶子节点包）
│   ├── model/
│   │   ├── user/                     # 按业务概念（聚合根）独立建包
│   │   │   ├── user.go               # 聚合根
│   │   │   ├── user_role.go          # 属于 user 聚合的内部实体
│   │   │   ├── user_events.go        # user 产生的领域事件
│   │   │   ├── user_error.go         # user 专属的业务错误定义
│   │   │   └── user_test.go
│   │   ├── role/                     # role 聚合包
│   │   │   ├── role.go               # 聚合根
│   │   │   ├── role_permission.go    # 内部实体
│   │   │   └── role_test.go
│   │   └── order/                    # order 聚合包
│   │       ├── order.go              # 聚合根
│   │       ├── order_item.go         # 内部实体
│   │       ├── order_events.go       # 领域事件
│   │       └── order_error.go        # 业务错误
│   └── port/                         # 写仓储端口 + 核心业务依赖的第三方接口
│       ├── repo/                     # 写仓储端口（接口，仅返回领域对象，不包含读仓储）
│       └── thirdPartyApi/            # 核心业务依赖的外部能力端口（如支付网关、风控引擎）
│
└── infrastructure/                   # 基础设施层
    ├── persistence/                  # 持久化实现
    ├── messaging/                    # 事件总线实现（Kafka/RabbitMQ/Pulsar 等）
    ├── scheduler/                    # 定时任务/调度器实现
    └── adapter/
        └── thirdPartyApi/            # 第三方 API 适配器

pkg/                                  # 跨层公共工具
├── errors/                           # 错误定义
├── middleware/                        # 中间件
└── utils/                            # 通用工具
```

> **语言适配说明**：根目录按语言惯例替换。Go → `internal/`，Java → `src/main/java/`，TypeScript/Python → `src/`。`pkg/` 目录同样按需调整。

**文件放置是硬约束，不可放错。**

## 基础设施层组件划分与目录组织规约

### 组件分类原则

基础设施组件按职责分为两类：

- **中间件与技术子系统**：包含数据库持久化、消息队列/事件总线、定时任务调度、缓存与锁等。
- **第三方 API 适配器**：包含外部供应商/服务商的 HTTP/RPC 客户端。

### 目录结构规则

**规则 A：技术领域独立子包原则（平铺结构）**

允许且推荐将具有独立技术体系的中间件/系统直接作为 `infrastructure/` 的一级子包，无须强行嵌套进 `adapter/` 目录：

```
infrastructure/
├── persistence/         # 持久化实现
├── messaging/           # 消息/事件总线实现
├── scheduler/           # 定时任务/调度器实现
├── cache/               # 缓存实现
└── adapter/             # 外部第三方 API 适配器
    └── thirdPartyApi/
```

**规则 B：端口实现一致性定理**

判定一个包是否合规的唯一标准是其依赖方向，而非目录深度。

只要基础设施包实现了 `application/port/` 或 `domain/port/` 中定义的接口，且没有反向将基础设施的依赖暴露给内层，即属于合规的架构实现。

## 依赖规则（不可违反）

```
trigger ──→ application ──→ domain ←── infrastructure
```

| 层 | 可以依赖 | 禁止依赖 |
|---|---|---|
| **domain** | 标准库 + 语言内置库 + 无状态/无 IO 的算法库或纯数据结构工具 + 领域自身 | 框架、ORM、HTTP、日志库、任何涉及 IO 或网络请求的 SDK、DTO、应用层、基础设施层、业务工具包 |
| **infrastructure** | domain 端口 + application 读仓储端口 + ORM/DB/SDK | application 层的具体业务类 (如 CommandHandler, ApplicationService 等) |
| **application** | domain (含 domain/event) | infrastructure |
| **trigger** | application (含 Command + CommandHandler) | domain、infrastructure |

> **依赖单向倒置**：领域层是系统的最高核心，其他层（应用层、触发器层、基础设施层、通用工具包）均可以向内依赖领域层；领域层绝不能反向依赖应用层、基础设施层或业务相关的底层工具包。

### 领域标识符（ID）生成规则

领域层标识符生成遵循两条规则，保障聚合根的强封装性同时避免技术污染。

#### 规则一：聚合内部实体的 ID 控制权归属于聚合

聚合根负责维护其内部实体与值对象的完整生命周期。在聚合根内部创建子实体（如 `User.AssignRole()` 生成 `UserRole`）时，**允许且推荐由领域层直接生成唯一标识符**，无需强制要求外部传入，以保障聚合根的强封装性。

#### 规则二：ID 生成的技术抽象与解耦（二选一实施）

**选项 1：Seedwork 统一封装（推荐）**

若需要进一步收拢第三方库的依赖，禁止在各个业务聚合中零散引入工具库，必须统一收拢至领域种子包：

- 在 `domain/common`（或其他语言等效路径）中提供统一的值对象构造工具（如 `common.NewID()`）
- 具体的业务聚合仅依赖 `common` 包，不直接接触具体的底层 SDK
- 允许 `domain/common` 依赖无状态、无 IO 的算法库（如 UUID 生成库）

**选项 2：ID 生成器契约（针对复杂 ID 算法）**

若 ID 生成依赖复杂的外部状态或算法（如分布式 Snowflake、DB 自增）：

- 领域层仅定义 `IDGenerator` 接口契约
- 具体的算法实现交由基础设施层完成
- 在应用层编排时注入

#### 架构 Review 判定对照

| 场景 | 是否合规 | 判定依据 |
|---|---|---|
| 业务聚合直接引入 UUID 生成库 | 🟡 允许但不完美 | 属于无状态算法库，合规；但建议统一收拢至 `domain/common` |
| `domain/common` 引入 UUID 库并暴露 `NewID()` | 🟢 完全合规 | 符合 Seedwork 封装规范，兼顾纯洁度与开发效率 |
| 外部（Controller/App）生成所有子实体 ID 再传给领域层 | 🔴 不推荐 | 破坏了聚合根对内部实体的封装性与生命周期控制权 |
| 领域层引入 ORM 或 Web 框架 | ⛔ 绝对违规 | 引入了基础设施/框架污染，直接阻断构建 |

## CQRS 读写分离

### 写侧（Command Side）

```
HTTP Request → Controller → Command → CommandHandler
                                    → 仓储加载聚合
                                    → 聚合.业务方法()
                                    → 仓储保存聚合
                                    → 领域事件
```

- Command 命名体现业务语义：`CreateCaseCommand`、`CancelOrderCommand`、`AssignTaskCommand`
- 禁止技术命名：`SaveXxxRequest`、`UpdateXxxRequest`
- CommandHandler 只做编排，不含业务逻辑
- **CommandHandler 禁止包含读仓储操作**：CommandHandler 仅处理写操作，通过写仓储加载聚合、执行业务方法、持久化聚合。任何查询需求（包括写入前的校验性查询）都应通过写仓储的领域对象加载完成，严禁在 CommandHandler 中注入或调用读仓储
- 写操作通过**写仓储**加载和保存完整聚合

### 读侧（Query Side）

```
HTTP Request → Controller → QueryService → 读仓储查库 → 返回 DTO
```

- 读操作不经过领域层，QueryService 直接调用读仓储
- 读仓储返回 DTO（扁平结构），不返回聚合对象
- 读侧可以 JOIN 多表、跨聚合查询，无需考虑聚合完整性约束

### 仓储 CQRS 分离（关键）

**仓储端口拆分为写仓储和读仓储，分别定义在不同层**：

- **写仓储端口**定义在 `domain/port/repo/` — 返回领域对象
- **读仓储端口**定义在 `application/port/repo/` — 返回 DTO

```
domain/port/repo/
└── {entity}Repository              # 写仓储端口 — 返回领域对象

application/port/repo/
└── {entity}ReadRepository          # 读仓储端口 — 返回 DTO
```
> **文件命名按语言惯例**：Go → `{entity}_repo.go` / `{entity}_read_repo.go`，Java → `{entity}Repository.java` / `{entity}ReadRepository.java`，TypeScript → `{entity}.repository.ts` / `{entity}.read-repository.ts`

**写仓储端口** — 操作聚合/实体，返回领域对象：

```go
// domain/port/repo/user_repo.go
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByUsername(ctx context.Context, username string) (*User, error)
}
```

**读仓储端口** — 纯查询，返回 DTO，不涉及领域对象：

```go
// application/port/repo/user_read_repo.go
type UserReadRepository interface {
    FindByID(ctx context.Context, id string) (*UserDTO, error)
    List(ctx context.Context, offset, limit int, status int) ([]*UserDTO, int64, error)
    FindByUsername(ctx context.Context, username string) (*UserWithRolesDTO, error)
}
```

**读仓储 DTO 与接口定义在同一个包/模块内**，按语言惯例组织文件：
- Go/TypeScript/Python/Kotlin：同一文件内定义 DTO + 接口
- Java/C#：定义在同包下的独立文件（保证 **DTO 与读仓储端口在同一个包内**，不跨包引用）

```
# 语言无关原则：写仓储在领域层，读仓储在应用层
domain/port/repo/
├── user_repo.go           # 写仓储端口 — 返回领域对象

application/port/repo/
├── user_read_repo.go      # 读仓储端口 + DTO（Go/TS风格）
│                          #   或：同包下 UserDTO.java + UserReadRepository.java（Java风格）
```

**基础设施层同样拆分为两个实现**，目录结构镜像领域层和应用层：

```
infrastructure/persistence/repo/
├── user_repo.go           # 写仓储实现 — 映射领域对象 ↔ ORM 模型
└── user_read_repo.go      # 读仓储实现 — 直接查询返回 DTO
```

写仓储实现将领域实体映射为 ORM 模型后持久化，在查询时又将 ORM 模型还原为领域对象。读仓储实现直接查询并返回 DTO，不经过领域对象转换。

**设计意图**：
- 读写职责清晰，不混放——避免一个仓储端口里既有 `Save` 又有 `List` 分页查询
- 为后期读写分离部署（CQRS 异库）预留扩展点——读仓储可切换为从只读副本/ElasticSearch 查询
- 读仓储可自由使用 JOIN、聚合函数、分页，不受聚合完整性约束
- 写操作保证聚合一致性的同时，读操作可独立优化查询性能

## CQRS 补充约束

### 读仓储端口归属应用层

读仓储返回 DTO（扁平数据结构），不涉及领域模型的操作，因此读仓储端口应定义在**应用层**而非领域层：

- 领域层的 `port/repo/` 只存放写仓储端口（返回领域对象）
- 应用层的 `port/repo/` 存放读仓储端口（返回 DTO）

```
application/port/repo/
├── user_read_repo.go          # 读仓储端口 + DTO
└── order_read_repo.go         # 读仓储端口 + DTO
```

### 读 DTO 禁止持有领域对象引用

读 DTO 是纯数据载体，**不得包含任何指向领域对象的方法或字段**：

- 禁止在 DTO 上定义 `ToAggregate()` 或 `ToEntity()` 等转换方法
- DTO 与领域模型之间没有引用关系，完全是两个独立的数据结构
- CommandHandler 需要修改聚合状态时，必须通过写仓储的 `FindByID()` 加载完整聚合，不可通过读 DTO 还原聚合

### CommandHandler 返回应用层 DTO

CommandHandler 是应用层的编排器，其方法签名中的返回值类型必须是**应用层 DTO**，不得返回领域层类型：

- 输入：应用层定义的 Command 对象
- 输出：应用层定义的 DTO 或 Result 对象
- 领域对象仅存在于 `domain/` 内部，不跨过应用层边界暴露给外部

### 触发器层响应 DTO 是纯叶子包

触发器层的响应 DTO 包（如 `resp/`）必须是**零外部依赖的纯叶子节点**：

- 只包含纯结构体定义，不 import 任何业务包
- 对象转换函数（DTO ↔ 领域对象）必须在 Controller 层完成，不在响应 DTO 包中定义
- 转换函数放在 Controller 文件末尾，或独立的 `mapper/` 包中

### 缓存策略下沉到基础设施层

Cache-Aside 等缓存策略的实现细节应封装在基础设施层的仓储实现中，应用层 QueryService 不感知缓存的存在：

- QueryService 调用读仓储接口获取数据，从不过问数据来源（DB 还是缓存）
- 读仓储实现内部完成「先读缓存 → miss 回源 DB → 回写缓存」的完整流程
- 缓存预热、失效、刷新等操作全部在基础设施层处理
- 应用层对缓存仓储的依赖通过接口注入，不直接操作 Redis/Memcached 等客户端

### 基础设施层目录镜像领域层

基础设施层的实现目录结构应与领域层端口目录结构保持镜像对应：

```
domain/port/repo/              →  infrastructure/persistence/repo/
domain/port/anti-corruption/   →  infrastructure/persistence/anti-corruption/
```

这种镜像关系使开发者能直观地找到每个接口的实现，无需查阅额外的配置或文档。

## 聚合

- 整存整取：`repo.Save(aggr)` / `repo.Load(id)`，不单独持久化内部实体
- 写仓储接口的入参和返回值必须是领域模型/聚合根，绝不能接收或返回 DTO/DAO 结构体
- 外部只能引用聚合根 ID，禁止直接引用聚合内部实体
- 聚合内 ACID 事务，聚合间通过领域事件最终一致
- 聚合根及其内部实体按业务概念（聚合根）建包，放在 `domain/model/{aggregate}/` 下，不按 aggr/entity/valObj 等技术类型分包

### 模型层高内聚原则

- **通用模型基类统一收拢于 `domain/model/`**：聚合根基类（`AggregateRoot`）、实体基类（`Entity`）、值对象接口（`ValueObject`）等领域构件，必须统一归纳在 `domain/model/` 根目录或其子目录（如 `domain/model/base/` 或 `domain/model/aggregate/`）下，作为领域模型的统一基石
- **包名与概念语义对齐**：包路径必须表达代码的真正领域语义，`domain/event/` 仅允许存放事件接口定义、事件基类与具体领域事件结构体，不允许掺杂任何持久化、仓储或聚合根控制逻辑

## 实体 — 充血模型

- 实体必须包含业务行为方法，禁止仅有 getter/setter
- ❌ 反模式：`type X struct { ID, Status string }` + 外部 service 处理全部逻辑
- ✅ 正确：实体内部封装状态变更逻辑

```go
func (x *X) Submit() error {
    if x.Status != "draft" {
        return ErrInvalidStatus
    }
    x.Status = "submitted"
    x.Events = append(x.Events, NewXSubmittedEvent(x.ID))
    return nil
}
```

### 状态变更的同步约束

聚合根（及其实体）的每一个修改状态的方法，在完成业务逻辑后都必须同步更新时间戳：

- 聚合根上的修改方法 → 更新聚合根的 `UpdatedAt`
- 内部实体的修改方法 → 更新实体的时间戳 + 回溯更新聚合根的 `UpdatedAt`
- 聚合根统一管理时间戳一致性，确保外部观察到的修改时间始终是准确的

### 子实体操作必须通过聚合根

**禁止**外部直接操作聚合内的子实体集合。所有对内部实体的增删改都必须通过聚合根的公开方法：

- 聚合根方法负责维护业务不变量（如重复性检查、值唯一性约束、状态一致性）
- 外部代码只能调用聚合根的方法，不能直接访问 `aggregate.Items` 等内部列表
- 聚合根在操作子实体后同步更新自身状态（如计数、状态位、时间戳）

### 聚合根构造器

聚合根通过命名统一的构造器函数创建，禁止在外部直接 `new` 或 `&Aggr{}`：

- 构造器命名：`New{聚合名}(...)`
- 构造时完成字段校验（必填项、长度、格式）
- 构造时填充默认值（状态、时间戳、空切片）
- 校验失败返回错误，不创建无效对象

### 聚合根生命周期隔离

聚合根在内存中的实例化必须严格区分两种生命周期场景：

- **创生（Creation）**：代表新业务事实的发生，必须执行业务规则校验、分配新标识符、记录创建类领域事件
- **重建（Reconstitution）**：代表从持久化介质还原历史状态，必须跳过业务规则校验，且严禁触发任何领域事件

**绝对红线**：从数据库或缓存加载历史数据（Find/Load）时，绝对禁止触发创生类领域事件，避免产生"查询即误发通知/邮件"的幽灵事件（Ghost Events）。

### 聚合根重建函数

根据聚合根的特性与包结构，分类施策选择重建策略，无需死板地为每个聚合强制手写独立的重建函数：

**策略 A：独立重建函数（复杂聚合推荐）**

- 适用场景：聚合根构造时会产生领域事件，或聚合根属性为私有且仓储实现与领域模型处于不同物理包
- 命名：`Reconstitute{AggregateName}(...)`（如 `ReconstituteFileRecord(...)`）
- 调用方：仅限基础设施层（仓储实现）
- 内部行为：直接将持久化字段映射至聚合根的私有属性，禁止执行业务校验，禁止记录领域事件

**策略 B：同包直接映射（同包简化）**

- 适用场景：仓储 ORM 模型映射逻辑与领域模型处于同一物理包内
- 实现方式：利用同包访问权限直接给私有属性赋值（如 `po.toDomain()`），无需向外暴露公开的重建函数

**策略 C：构造函数复用（无事件聚合简化）**

- 适用场景：简单聚合根或配置类实体，其构造过程不产生任何领域事件
- 实现方式：允许仓储层直接复用 `NewXxx()` 构造方法或字面量，无需额外手写重建函数

### 聚合根私有属性封装保护

- 聚合根内部状态属性必须保持私有（不对外暴露），严禁为了 ORM 反射便利而将其公开
- 重建函数是基础设施层恢复聚合状态的唯一合法入口，在保障聚合根私有属性强封装的同时，避免向外暴露 Setter 方法或依赖反射破坏封装

### 子实体操作的幂等性

聚合根对子实体的**添加操作**必须是幂等的：

- 重复添加同一项：静默跳过，不报错
- 违反业务规则的添加（如值唯一性冲突）：返回业务错误
- 幂等检查通过唯一标识（ID 或唯一业务键）完成

Example:
```
assignItem(itemID):
    if already exists → return (no-op)
    append to list
    update aggregate timestamp
```

## 值对象

- 不可变，创建后不允许修改
- 通过属性值相等性判断，而非 ID
- 封装概念完整性（如 `Money{Amount, Currency}`、`Address{City, Street, Zip}`）

## 领域服务

- 无状态，放跨实体或跨值对象的业务逻辑
- 命名反映领域活动，禁止 "Manager"、"Handler"、"Util" 等无业务含义的词
- ✅ `OrderPricingService`、`OverdueCalculator`、`RiskAssessor`
- ❌ `OrderManager`、`OrderHandler`、`OrderUtil`

### 使用判例与充血优先原则

- **【充血优先原则】** 严禁滥用领域服务。优先将业务规则下沉至实体（Entity）与聚合根（Aggregate Root）内部（充血模型）。**能写在实体里的逻辑，绝对禁止剥离至领域服务**。
- **【准用场景（White List）】**
  1. **跨聚合协作（Cross-Aggregate Collaboration）**：如资金转账（协同两个 `Account` 聚合）、跨领域结算等。
  2. **依赖外部契约的领域校验**：如需调用仓储接口进行的"用户名唯一性校验"（`UserUniquenessChecker`）。
  3. **无主状态的纯领域计算/计算策略**：如复杂的动态计税引擎、折扣策略计算。

### 领域服务的三种典型场景

领域服务是实体的**补充**而非替代品，仅在以下场景中使用：

**1. 跨聚合的无状态协作**

当一个业务动作需要协同两个或多个不同的聚合，且该动作不属于其中任何一个聚合时：

- 典型例子：资金转账 — 扣钱属于账户 A 聚合，加钱属于账户 B 聚合，但"从 A 转账到 B"这个动作本身既不属于 A 也不属于 B
- 解决方案：创建领域服务协调多个聚合的状态变更

**2. 依赖外部/抽象能力的领域规则校验**

当核心业务规则校验需要依赖仓储数据或外部接口时：

- 典型例子：用户注册时的"用户名唯一性校验" — 实体在未创建时无法自行判断数据库是否重名
- 解决方案：定义领域服务，通过仓储接口完成校验

**3. 领域概念本身的纯计算/转换逻辑**

某些业务计算过程本身是一个独立的领域概念，但不属于某个特定实体：

- 典型例子：计税引擎、动态折扣计算策略

### 领域服务物理归属与包路径规约

领域服务不单独建全局目录，而是与所服务的聚合包同居或独立建子域包。

#### 规则 A：聚合亲和性物理同居（Standard Rule）

- **【标准规范】** 领域服务必须存放在其**所服务/主导的业务聚合包**内部（即 `domain/model/{domain_model}/`），实现业务概念的高内聚。
- 示例：`domain/model/user/user_uniqueness_checker.go` → `package user`

#### 规则 B：独立子域概念演进（Independent Sub-domain Rule）

- **【例外规范】** 当领域服务跨多个聚合且无明确主导方，且该服务本身已演化为一个**独立的业务子概念**（如"结算"`Settlement`）时：
- 允许在 `domain/model/` 下为其建立独立的业务子包（如 `domain/model/settlement/settlement_service.go` → `package settlement`）。

#### 规则 C：全局平铺禁令（Anti-Pattern Red Line）

- **【绝对禁止（Red Line）】** 严禁建立全局平铺的 `domain/service/` 目录。此类"按技术类型分包"的行为会导致包泛滥、高耦合及 import cycle 编译报错。

### 领域服务 vs 应用服务

| 维度 | 领域服务 (Domain Service) | 应用服务 (Application Service) |
|---|---|---|
| **所属层级** | **领域层** — `domain/model/{domain_model}/` | **应用层** — `application/commandHandler/` 或 `application/queryService/` |
| **核心职责** | 表达**纯粹的业务规则**（如计算手续费、转账规则） | 表达**系统用例与流程编排**（如事务控制、发邮件、权限拦截） |
| **对外暴露** | 不直接对外暴露，仅被应用层调用 | 对外暴露给 Controller / Trigger 入口 |
| **交互契约** | 入参/出参必须使用**领域模型**（实体、值对象、领域标识符） | 入参/出参使用 **Command / DTO / Query** |
| **感知范围** | **绝对禁止**感知 HTTP/RPC/DTO/框架概念 | 可感知 DTO，协调多个领域服务或基础设施层端口 |

> 应用服务像**指挥官/调度员**：接收请求，协调各部门但自己不干体力活。领域服务像**专家顾问**：专门解决某个极其复杂的业务难题，执行核心规则。

### 贫血模型警示

防止领域服务引发贫血模型是架构的关键约束：

1. **能写在实体里的，绝不写在领域服务里** — 如余额校验并扣款是账户实体自己的行为，必须写在实体内部
2. **只有涉及"多个实体协同"或"纯抽象计算"时，才剥离到领域服务中**
3. **领域服务是实体的补充，不是实体的替代品** — 如果发现实体只有 getter/setter 而所有逻辑都在领域服务里，说明架构已退化为贫血模型

### 架构 Review 判定对照表

| 场景代码描述 | 判定 | 规则依据与修正建议 |
|---|---|---|
| `User.UpdatePassword()` 内部修改密码并校验格式 | 🟢 **完全合规** | 属于实体自身的行为，写在实体内部，无需使用领域服务。 |
| `domain/model/user/user_uniqueness_checker.go` | 🟢 **完全合规** | 高内聚地放在 `user` 业务包下，属于合规的领域服务。 |
| `domain/service/user_service.go` 全局平铺 | ⛔ **违规** | 违反规则 C。应移入 `domain/model/user/` 下并归纳具体职责。 |
| 在领域服务内部开启 DB 事务 / 发送 HTTP 请求 | ⛔ **违规** | 污染领域层。事务控制与外部通信应退回应用层（Application Layer）。 |

## 领域事件

### 领域事件契约（Domain Event Contract）

- **概念定位**：领域事件是领域层的"一等公民"，代表业务中已发生的事实，是平行于静态模型的独立行为维度
- **文件路径**：`domain/event/event.go`（或其他语言等效路径）
- **设计约束**：
  - 该包必须是整个项目的最底层"叶子节点"，只定义通用的 `DomainEvent` 接口
  - **绝对不能**显式引入（import）任何具体的业务聚合包，以此彻底避免循环依赖
  - 具体的事件结构体（如 `OrderConfirmed`、`CaseAssigned`）可以放在各聚合包内，但统一实现 `domain/event` 包定义的接口
- **主体与产物分离原则**：
  - 聚合根是领域逻辑与状态变更的行为主体，领域事件只是状态变更的通知产物
  - **聚合根基类（AggregateRoot）严禁放入 `domain/event` 包**：聚合根基类的本质是模型（Model），绝不能为了引用方便而将其放入事件包中，否则会导致概念倒错（从"聚合产生事件"沦为"事件包含聚合"）
  - **依赖方向单向不可逆**：领域模型层可以依赖事件接口（聚合根内部维护 `[]DomainEvent`），但事件包必须保持纯净，绝对禁止反向依赖任何具体领域模型或聚合基类
  - **包内容纯净化**：`domain/event/` 仅允许存放事件接口定义（`DomainEvent`）、事件基类与具体领域事件结构体，严禁掺杂任何持久化逻辑、仓储操作或聚合根控制逻辑

```go
// domain/event/event.go — 叶子节点包，零业务依赖
package event

import "time"

type DomainEvent interface {
    EventName() string
    OccurredAt() time.Time
}
```

### 领域事件使用

- 跨聚合通信的唯一方式
- 聚合状态变更 → 产生领域事件 → CommandHandler 发布事件到 EventBus → trigger/listener/ 拾取 → 构造新的 Command 处理副作用
- 事件命名使用过去时：`OrderConfirmed`、`CaseAssigned`、`PaymentReceived`

#### 领域事件的产生时机

领域事件由**聚合根的业务方法自行触发**，CommandHandler 不手动创建或发布事件：

```
聚合根.业务方法()
  ├── 校验业务规则
  ├── 变更自身状态
  ├── 更新 UpdatedAt
  └── AddDomainEvent(NewXxxEvent(...))  ← 聚合根自行记录事件
```

- CommandHandler 调用聚合根的业务方法，方法内部调用 `AddDomainEvent()` 记录事件
- CommandHandler 不直接创建任何事件对象
- 仓储基类在事务提交后自动调用 `ClearDomainEvents()` 收集并发布所有未发布的事件

#### 事件发布时机

仓储基类封装**「事务内持久化 → 事务提交 → 自动发布领域事件」**的流程：

```
仓储.Save(聚合根)
  ├── 开启事务
  ├── 执行持久化 saveFn
  ├── 提交事务        ← 事务先提交
  └── ClearDomainEvents() → Publish...  ← 事件后发布
```

- 事件在事务**提交后**发布，避免部分失败导致事件重复或丢失
- 调用方只需传入持久化回调（`saveFn`），事件发布由基类统一处理
- 事务回滚时事件也不会被发布

### 事件总线接口（Event Bus Port）

- **概念定位**：事件总线是技术管道和流程编排工具，不包含纯粹的领域业务逻辑
- **文件路径**：`application/event/bus.go`
- **设计原则**：
  - **接口定义在应用层**：应用层定义 `EventBus` 接口（声明 `Publish` 方法），表达业务对消息传递基础设施的技术诉求。注意：EventBus 只负责**发布**事件，事件的**监听**由 `trigger/listener/` 层直接对接消息基础设施完成，不在 EventBus 中耦合 Subscribe 语义
  - **避免建立平级 shared 包**：在单体或单个微服务内部，不要建立与 application/domain 平级的 `shared` 包来放总线，这会导致依赖边界模糊和洋葱架构层级混乱。直接收拢在 `application/event` 下是最内聚、最安全的做法

```go
// application/event/bus.go — 应用层端口
package event

import (
    "context"
    domainEvent "your-project/internal/domain/event"
)

// EventBus 只负责发布事件，监听由 trigger/listener/ 层处理
type EventBus interface {
    Publish(ctx context.Context, events ...domainEvent.DomainEvent) error
}
```

### 异步流量触发器（Asynchronous Triggers）

- **概念定位**：负责对接外部异步消息源的入站适配器（Driving Adapter）
- **文件路径**：`trigger/listener/`
- **命名规范**：`trigger/listener/` 目录兼容 Kafka、RabbitMQ、Pulsar、EventBridge 等消息基础设施
- **职责约束**：`trigger` 层只做"翻译官"和防腐层（ACL），不做业务逻辑

#### 监听器命名规范（唯一强制模式）

**本项目强制采用"单一职责模式"**，不提供任何备选命名方案。

**核心思想**：一个 Listener 结构体只负责执行一件具体的、具有领域意义的后续副作用（Side Effect）。

**命名公式**：
```
[动词+宾语]On[事件名称]
```

**示例**：

| Listener 名称 | 职责说明 |
|---|---|
| `SendWelcomeEmailOnUserRegistered` | 用户注册后发送欢迎邮件 |
| `CreateDefaultProfileOnUserRegistered` | 用户注册后创建默认个人资料 |
| `InvalidateSessionOnUserDisabled` | 用户被禁用时强制清退 Session |
| `NotifyAssignerOnTaskOverdue` | 任务逾期时通知负责人 |
| `SyncToSearchEngineOnCaseUpdated` | 案件更新时同步搜索引擎 |

**三项命名军规（硬性约束）**：
1. **时态一致**：事件代表已发生的事实，必须使用过去时（`Registered`、`Disabled`、`Updated`），Listener 必须沿用
2. **行为先行**：动词（`Send`、`Create`、`Invalidate`）必须放在最前面，体现其行为本质
3. **严禁后缀冗余**：结构中既然已包含 `On+事件名`，**绝对不要**在末尾追加 `Listener` 或 `Handler`（❌ 严禁写成 `InvalidateSessionOnUserDisabledListener`）

#### 职责演进（命令驱动契约）

`trigger/listener/` 负责拦截消息、反序列化、防腐转换，并立即构造应用层 **Command** 调用 CommandHandler。所有写操作入口统一通过 Command 完成，保证可测试性。

```
外部消息 → trigger/listener/ 监听器
         → 反序列化 + 防腐转换
         → 构造 Command
         → 调用 application/commandHandler/ 处理
```

以下为符合规范的 Go 代码示例：

```go
// trigger/listener/send_welcome_email_on_user_registered.go
package listener

import (
    "context"
    "encoding/json"

    "your-project/internal/application/command"
    "your-project/internal/application/commandHandler"
)

// SendWelcomeEmailOnUserRegistered — 单一职责：用户注册后发送欢迎邮件
// 命名符合 [动词+宾语]On[事件名] 规范
type SendWelcomeEmailOnUserRegistered struct {
    handler commandHandler.SendWelcomeEmailCommandHandler
}

func NewSendWelcomeEmailOnUserRegistered(
    handler commandHandler.SendWelcomeEmailCommandHandler,
) *SendWelcomeEmailOnUserRegistered {
    return &SendWelcomeEmailOnUserRegistered{handler: handler}
}

// HandleMessage 反序列化领域事件消息，构造 Command 并委派给 CommandHandler
func (l *SendWelcomeEmailOnUserRegistered) HandleMessage(ctx context.Context, raw []byte) error {
    var event UserRegisteredEvent // 领域事件 DTO（反序列化载体，非领域实体）
    if err := json.Unmarshal(raw, &event); err != nil {
        return err // 序列化错误由基础设施层 ACK/NACK 策略处理
    }

    // 防腐转换：将外部消息结构转换为领域 Command
    cmd := command.SendWelcomeEmailCommand{
        UserID:   event.UserID,
        Email:    event.Email,
        UserName: event.UserName,
    }

    // 委派给应用层 CommandHandler，保证写入口统一
    return l.handler.Handle(ctx, cmd)
}
```

```go
// trigger/listener/invalidate_session_on_user_disabled.go
package listener

import (
    "context"
    "encoding/json"

    "your-project/internal/application/command"
    "your-project/internal/application/commandHandler"
)

// InvalidateSessionOnUserDisabled — 单一职责：用户被禁用时强制清退所有活跃 Session
type InvalidateSessionOnUserDisabled struct {
    handler commandHandler.InvalidateSessionCommandHandler
}

func NewInvalidateSessionOnUserDisabled(
    handler commandHandler.InvalidateSessionCommandHandler,
) *InvalidateSessionOnUserDisabled {
    return &InvalidateSessionOnUserDisabled{handler: handler}
}

func (l *InvalidateSessionOnUserDisabled) HandleMessage(ctx context.Context, raw []byte) error {
    var event UserDisabledEvent
    if err := json.Unmarshal(raw, &event); err != nil {
        return err
    }

    cmd := command.InvalidateSessionCommand{
        UserID: event.UserID,
        Reason: "account_disabled",
    }

    return l.handler.Handle(ctx, cmd)
}
```

**设计要点**：
- 每个 `trigger/listener/` 文件只包含一个 Listener 结构体，对应一个具体副作用
- Listener 通过构造函数注入所需的 CommandHandler，不直接依赖基础设施
- `HandleMessage` 方法的三步模式固定：**反序列化 → 构造 Command → 调用 Handler**
- 事件结构中只包含松耦合的 DTO 字段（如 `UserID`、`Email`），不引用任何领域实体

### 基础设施实现（Infrastructure Messaging）

- **概念定位**：具体的底层技术细节实现（Driven Adapter）
- **文件路径**：`infrastructure/messaging/`
- **设计原则**：
  - 严格遵循**依赖倒置原则（DIP）**
  - 诸如 `KafkaEventBus` 或 `RabbitMQEventBus` 等具体技术类必须存放在基础设施层，去 `import` 并实现应用层定义的 `EventBus` 接口
  - 所有关于序列化、连接池、ACK/NACK 控制、重试机制的技术脏活累活，必须在此层彻底隔离，严禁向上污染应用层和领域层

```go
// infrastructure/messaging/kafka_bus.go — 基础设施实现
package messaging

import (
    "context"
    appEvent "your-project/internal/application/event"
    domainEvent "your-project/internal/domain/event"
)

// KafkaEventBus 只负责发布事件，监听由 trigger/listener/ 层处理
type KafkaEventBus struct {
    producer *kafka.Producer
}

// 确保编译期检查实现了接口
var _ appEvent.EventBus = (*KafkaEventBus)(nil)

func (b *KafkaEventBus) Publish(ctx context.Context, events ...domainEvent.DomainEvent) error {
    // 序列化、发送、ACK 控制、重试 — 全部在此层隔离
}
```

## 端口与适配器命名规范

- **Port（端口）**：领域层或应用层定义的抽象接口，必须存放在 `port/` 目录下
- **Adapter（适配器）**：基础设施层的具体实现，必须统一存放在 `infrastructure/adapter/` 目录下
- **严禁**在 `domain/` 或 `application/` 层出现名为 `adapter` 的包

## 防腐层

- **第三方接口按业务属性归层**：
  - **核心业务依赖**（领域规则直接强依赖的外部能力，如支付网关、风控引擎）：端口定义在 `domain/port/thirdPartyApi/`
  - **纯技术支撑/外围服务**（仅供应用编排使用，如文件存储、短信发送）：端口定义在 `application/port/thirdPartyApi/`
- 实现（Adapter）统一在 `infrastructure/adapter/thirdPartyApi/`
- 必须将外部数据结构转换为领域语言，不让外部模型污染领域层
- 外部异常在防腐层内包装为领域异常

## 非领域数据降级

- 如果某项数据是纯 Append-Only（只追加）、无状态变更且无领域规则校验的技术日志/流水（如简单审计日志）：
  - **不要**强行定义为领域聚合根
  - 应将其降级为应用层技术服务接口，定义在 `application/port/thirdPartyApi/` 或基础设施层中，由基础设施层直接写入数据库

## 异常处理

- 领域层定义业务异常，向上抛，由应用层或触发器层统一处理
- 禁止在领域层处理 HTTP 响应、写日志、发通知

```go
var ErrOrderCannotBeCancelled = errors.New("已发货订单不可取消")
```

### 错误码按聚合分段

如果使用数值错误码，各聚合的错误码应按固定范围分段分配，避免不同聚合的错误码冲突：

```
聚合 A  错误码范围: 1000-1999
聚合 B  错误码范围: 2000-2999
聚合 C  错误码范围: 3000-3999
```

- 每个聚合定义一个常量或注释说明其错误码范围
- 同一范围内的码值按子实体或功能继续细分
- 错误码范围在聚合边界处统一规划，不交叉分配

## 项目初始化步骤

1. **理解业务域** — 识别核心域、支撑域、通用域
2. **划定限界上下文** — 每个上下文独立目录/模块
3. **定义通用语言** — 团队统一业务术语
4. **建模聚合** — 识别聚合根、实体、值对象的边界
5. **生成目录结构** — 按文件放置规则创建目录树
6. **编写领域层** — 先写实体行为 + 仓储接口（读写分离）
7. **实现基础设施** — 实现写仓储 + 读仓储 + 防腐层
8. **组装应用层** — CommandHandler + QueryService
9. **挂载触发器** — Controller / gRPC / Job
10. **注入依赖** — DI 容器配置

## 测试策略

| 层 | 测试类型 | 目标 |
|---|---|---|
| domain | 单元测试 | 覆盖所有业务规则和状态流转，覆盖率 > 90% |
| application | 单元测试 | mock 仓储接口，验证编排逻辑 |
| infrastructure | 集成测试 | 验证仓储实现与数据库交互正确性 |
| trigger | 端点测试 | 验证请求/响应格式和状态码 |

## 性能准则

- 读走 QueryService → 读仓储，直接查库返回 DTO，不加载聚合
- 跨聚合查询走读仓储，不通过写仓储
- 聚合加载按需，避免 N+1
- 领域事件异步投递，不阻塞主流程

## 场景示例

### 用户说："为新项目搭建 DDD 架构"

1. 提问：业务核心是什么？主要实体有哪些？状态如何流转？
2. 根据回答确定限界上下文
3. 按目录结构生成骨架
4. 先写聚合根和实体的业务行为
5. 对每个聚合，分别定义写仓储接口和读仓储接口
6. 实现基础设施层

### 用户说："这段代码不符合 DDD"

对照检查清单：
- [ ] 领域层是否引入了外部框架 import？
- [ ] 实体是否有业务行为方法（不只是 getter/setter）？
- [ ] 仓储接口是否拆分为写仓储和读仓储？
- [ ] 写仓储是否整存整取聚合？
- [ ] 读仓储是否返回 DTO 而非领域对象？
- [ ] 读操作是否走 QueryService 而非 CommandHandler？
- [ ] Command 命名是否体现业务语义？
- [ ] 跨聚合通信是否用了领域事件？
  - [ ] 领域事件接口是否定义在 `domain/event/` 叶子节点包，且不 import 任何业务包？
  - [ ] 聚合根基类（`AggregateRoot`）是否在 `domain/model/` 下，而非 `domain/event/` 下？
  - [ ] `domain/event/` 包是否仅包含事件接口/基类/事件结构体，不含持久化或聚合根控制逻辑？
  - [ ] 事件总线接口是否定义在 `application/event/`，且只包含 `Publish` 方法（无 `Subscribe`）？
  - [ ] 异步消息监听器是否使用 `trigger/listener/` 命名（而非 `trigger/mq/`）？
  - [ ] `trigger/listener/` 监听器是否只做反序列化 + 防腐转换 + 调用 CommandHandler，而非直接实现业务逻辑？
  - [ ] 监听器命名是否严格遵循 `[动词+宾语]On[事件名]` 单一职责模式？
  - [ ] 监听器命名是否以动词开头（`Send`、`Create`、`Invalidate`），而非以事件名开头？
  - [ ] 监听器命名末尾是否没有冗余的 `Listener`/`Handler` 后缀？
  - [ ] 一个监听器文件是否只包含一个监听器结构体，对应一个具体副作用？
  - [ ] 事件总线实现（KafkaEventBus 等）是否在 `infrastructure/messaging/` 中实现应用层接口，遵循 DIP？
- [ ] 防腐层是否转换了外部数据结构？
- [ ] 领域服务命名是否体现领域活动？
	- [ ] 领域服务是否放在 `domain/model/{聚合}/` 下，而非全局 `domain/service/`？
	- [ ] 领域服务是否无状态、入参/出参使用领域模型，且不包含事务/HTTP/框架概念？