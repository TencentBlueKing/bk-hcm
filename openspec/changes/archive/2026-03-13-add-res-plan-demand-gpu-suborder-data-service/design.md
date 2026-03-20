## Context

当前 `res_plan_demand_gpu_suborder` 只有建表 SQL，尚未接入 BK-HCM 现有资源规划数据链路，因此缺少以下关键能力：

- DAL 层没有该表的结构体、列描述、校验逻辑和统一表名定义。
- DAO 集合没有暴露对应资源入口，上层 service 无法通过 `dao.Set` 访问该表。
- data-service 资源规划模块没有该资源的 CRUD 路由和请求处理逻辑。
- data-service API 与 global client 没有对应协议定义和调用封装。

项目内已经存在较成熟的资源规划实现模式，例如 `res_plan_demand`、`res_plan_sub_ticket`、`res_plan_demand_penalty_base`。本次设计目标不是引入新的访问模式，而是在最小改动范围内，把 `res_plan_demand_gpu_suborder` 接入现有分层架构，保证命名、目录组织、校验、DAO 暴露方式、REST 路由风格和 client 封装方式与同类资源保持一致。

该变更横跨 `pkg/dal/table`、`pkg/dal/dao`、`cmd/data-service/service`、`pkg/api/data-service`、`pkg/client/data-service/global` 多个模块，属于跨层接入型改动，适合先形成明确设计，降低后续实现时的命名偏差和接口不一致风险。

## Goals / Non-Goals

**Goals:**

- 为 `res_plan_demand_gpu_suborder` 提供完整的 data-service CRUD 支撑。
- 确保表结构映射与 `scripts/sql/9999_20260312_2206_res_plan_gpu_demads.sql` 保持一致。
- 按现有资源规划模块规范补齐表名常量、DAO 访问入口、API 协议定义、service 路由与 global client 封装。
- 让上层模块能够通过统一的 data-service client 调用 GPU 子单资源，而不直接依赖底层 DAO 或表结构。

**Non-Goals:**

- 不在本次变更中实现 `res_plan_demand_gpu_order` 或 `res_plan_demand_gpu_template` 的 data-service 能力。
- 不新增业务层聚合逻辑、联表查询逻辑或状态流转规则。
- 不调整现有 `res_plan_demand`、`res_plan_sub_ticket` 等资源的既有接口行为。
- 不引入新的通用框架、ORM 扩展或新的 client 调用机制。

## Decisions

### 1. 采用现有资源规划模块的标准接入模式

**Decision**

新增 `res_plan_demand_gpu_suborder` 时，整体实现方式直接对齐现有资源规划资源的模式，并保持一个资源对应一个子目录：

- `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/` 定义表结构、字段描述、`InsertValidate`、`UpdateValidate`
- `pkg/dal/table/table.go` 定义表名并注册到 `TableMap`
- `pkg/dal/dao/dao.go` 暴露 `ResPlanDemandGpuSubOrder()`
- `cmd/data-service/service/resource-plan/res-plan-demand-gpu-suborder/` 新建独立目录与 CRUD handler
- `pkg/api/data-service/resource-plan/` 定义 create/update/list 协议
- `pkg/client/data-service/global/resource_plan.go` 提供 CRUD client 封装

**Rationale**

这样可以最大程度复用项目已验证的组织方式，降低代码审查和后续维护成本，也能避免 GPU 子单资源成为资源规划模块中的“特例”实现。

**Alternatives considered**

- 直接把 GPU 子单逻辑并入现有 `res_plan_demand` service：可以少建部分文件，但会把两张不同职责的数据表耦合到同一资源入口，不利于后续独立扩展。
- 仅补 DAL 和 DAO，不提供 data-service client：短期可用，但不符合项目通过统一 client 访问 data-service 的惯例。

### 2. 表结构严格以建表 SQL 为源，字段类型和校验最小化映射

**Decision**

`res_plan_demand_gpu_suborder` 的 Go 表定义严格对应 SQL 中的字段：

- 标量字段保持 snake_case -> camelCase 的常规映射
- `comment`、`extension` 使用项目现有 JSON 字段处理方式
- `created_at`、`updated_at` 使用统一时间类型
- 插入校验仅校验 SQL 语义上明确要求的非空、长度和数值边界
- 更新校验遵循现有资源规划表“允许部分字段更新”的模式，避免把更新请求约束成全量提交

**Rationale**

当前需求是补齐 data-service 接口，而不是重定义业务模型。严格跟随建表 SQL 可以减少实现歧义，也方便后续与数据库字段排查一致性问题。

**Alternatives considered**

- 在 data-service 层新增更强的业务语义校验，例如状态机流转或 `order_id` 关联校验：这些规则可能属于更高层业务逻辑，目前没有明确需求，提前加入会扩大变更范围。

### 3. CRUD 能力按资源独立路由暴露，不复用已有需求单接口路径

**Decision**

为 GPU 子单单独定义一组 data-service 路由，沿用资源规划模块的既有 REST 风格：

- `list`
- `batch/create`
- `batch`

其中删除使用 `DELETE /batch`，更新使用 `PATCH /batch`，与现有资源规划资源的接口约定保持一致。

**Rationale**

独立路由更符合“一张资源表对应一组 data-service 资源接口”的现有习惯，也便于后续如需增加状态更新、批量 upsert、CAS 更新等能力时按资源继续扩展。

**Alternatives considered**

- 挂载到已有 `res_plan_demand` 路由下作为子资源：会弱化 `res_plan_demand_gpu_suborder` 作为独立数据实体的边界，且与当前 DAO/表设计不一致。

### 4. DAO 暴露方式沿用 `dao.Set` + 资源规划 DAO 实现的统一模式

**Decision**

在 `pkg/dal/dao/dao.go` 中新增 `ResPlanDemandGpuSubOrder()`，返回资源规划 DAO 接口实例，实例构造继续使用 `Orm`、`IDGen`、`Audit` 三元组合。

**Rationale**

资源规划相关 DAO 当前都通过 `dao.Set` 暴露，上层 service 初始化时依赖该统一入口。保持同样的依赖注入方式，能让新资源自然接入现有 capability/service 初始化流程。

**Alternatives considered**

- service 直接引用具体 DAO 实现：会破坏 `dao.Set` 抽象层，增加测试替换和后续重构成本。

### 5. 上层访问统一经由 `pkg/client/data-service/global/resource_plan.go`

**Decision**

除了 API 协议定义外，还需要在 `pkg/client/data-service/global/resource_plan.go` 新增 CRUD client 方法，内部统一调用 `pkg/client/common/request.go` 的标准请求封装。

**Rationale**

这符合仓库中 data-service client 的统一用法，能保证请求路径、泛型返回值、错误处理方式与现有 client 一致，避免上层重复拼接 REST 请求。

**Alternatives considered**

- 暂不提供 client，仅让调用方自行发请求：会造成使用姿势不统一，也违背当前仓库对 client 封装的要求。

## Risks / Trade-offs

- [Risk] SQL 中 `comment`、`extension` 为 JSON 字段，若选型与现有资源规划 JSON 字段实现不一致，可能导致序列化或查询行为不一致。
  Mitigation: 参考同仓库已有 JSON 字段表定义，优先复用相同类型与 tag 约定。

- [Risk] 新资源命名在表定义、DAO 方法、service 目录、API 协议、client 方法中需要保持一致，任何一处偏差都会导致调用链断裂。
  Mitigation: 统一采用 `ResPlanDemandGpuSubOrder` 作为 Go 命名基准，SQL 表名固定为 `res_plan_demand_gpu_suborder`。

- [Risk] 若直接照搬 `res_plan_demand`，可能引入不适用于 GPU 子单的字段校验或多余接口能力。
  Mitigation: 仅复用其分层模式和路由风格，不复用不相关字段和业务语义。

- [Trade-off] 先只提供通用 CRUD，短期内不能覆盖更复杂的业务查询或状态控制能力。
  Mitigation: 本次先建立稳定基础能力，后续如出现明确业务需求，再在独立资源接口上增量扩展。

## Migration Plan

1. 新增 `pkg/dal/table/table.go` 表名常量与 `TableMap` 注册项。
2. 在 `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/` 增加 `res_plan_demand_gpu_suborder` 表定义文件。
3. 在资源规划 DAO 模块中补充对应 DAO 接口/实现，并在 `pkg/dal/dao/dao.go` 暴露 `ResPlanDemandGpuSubOrder()`。
4. 在 `pkg/api/data-service/resource-plan/` 增加 CRUD 请求与响应结构。
5. 在 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-suborder/` 新增 service 目录和 CRUD handler，并接入初始化路由。
6. 在 `pkg/client/data-service/global/resource_plan.go` 增加 CRUD client 方法。
7. 通过编译和相关单测/静态检查验证调用链闭环。

本次为新增能力，无需数据迁移。若上线后发现接口实现问题，可通过回滚代码版本撤回未被调用的新接口；由于不修改现有表结构和既有接口，回滚影响面较小。

## Open Questions

- `comment` 与 `extension` 的 Go 字段类型应直接复用哪一种现有 JSON 包装类型，是否已有资源规划模块内的首选实践。
- 是否需要在首版同时提供 `batch upsert`、状态更新或按 `order_id` 的专用查询接口；若无明确调用场景，本次默认不包含。
