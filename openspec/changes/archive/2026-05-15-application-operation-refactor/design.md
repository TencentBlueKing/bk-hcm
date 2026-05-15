## Context

审批流程（`approval_process` 表）以 `application_type` 为键索引对应的 ITSM `service_id`（审批流程 ID），形成 1:1 的映射关系。`application` 表同样用 `type` 字段记录申请类型，前端通过 `type` 进行申请单分类展示。

当前已有改动（git 暂存区已修改）：
- `pkg/dal/table/application/application.go`：`ApplicationTable` 已加入 `operation` 字段，`InsertValidate` 已要求 `operation` 非空。
- `pkg/api/data-service/application.go`：`ApplicationCreateReq` 已加入 `operation` 字段（required）。
- `pkg/criteria/enumor/application.go`：`ApplicationOperation` 类型已声明，但常量尚未定义。

尚未完成：handler 接口未扩展 `GetOperation()`、各 handler 实现未更新构造函数、`create.go` 未传入 `operation`、存量数据缺少 `operation` 值、查询接口未暴露 `operation` 过滤能力。

## Goals / Non-Goals

**Goals:**
- 完善 `ApplicationOperation` 枚举常量，覆盖全量操作类型（向后兼容：现有 type 一一对应同值 operation）。
- 在 `ApplicationHandler` 接口增加 `GetOperation()` 方法，`BaseApplicationHandler` 内置默认实现，各 handler 在构造时传入对应 operation。
- `create.go` 在写库时填写 `operation` 字段；`bkBizIDs` 判断逻辑改用 `operation` 判断。
- 查询接口（`ApplicationResp` / `ApplicationListResult`）返回 `operation` 字段，允许按 `operation` 字段过滤。
- 提供存量数据迁移脚本，将已有记录的 `operation` 补填为对应的 `type` 值。

**Non-Goals:**
- 不修改 `approval_process` 表结构（`type` → `service_id` 映射关系维持不变，`type` 仍是审批流的粒度）。
- 不修改前端代码（接口语义调整，前端按需接入）。
- 不拆分现有审批流（不新增 `service_id`，不改变 ITSM 流程配置）。

## Decisions

### 1. `ApplicationOperation` 常量命名：使用 `Op` 前缀

**背景**：`ApplicationType` 常量（如 `AddAccount`、`CreateCvm`）与 `ApplicationOperation` 常量在同一 `enumor` 包中，Go 不允许同包内常量重名。

**决策**：`ApplicationOperation` 常量统一使用 `Op` 前缀命名（如 `OpAddAccount`、`OpCreateSubAccount`），字符串值则直接与已有 `ApplicationType` 字符串相同（如 `"add_account"`），新增细粒度操作用新字符串（如 `"create_sub_account"`）。

**为什么不用其他方案**：
- 方案 A（单独文件/包）：引入包路径变化，改动范围更大。
- 方案 B（不定义常量只用字符串字面量）：类型安全性差，散落于代码各处难以维护。
- 方案 C（`Op` 前缀）：最小改动，类型安全，可以用 `Validate()` 方法校验合法值。

### 2. `ApplicationHandler` 接口扩展：在 Base 层内置默认值

**背景**：有 15+ 个 handler，若全部改写 `GetOperation()`，工作量大且容易遗漏。

**决策**：
- `BaseApplicationHandler` 新增 `operation enumor.ApplicationOperation` 字段，`GetOperation()` 由 `BaseApplicationHandler` 提供默认实现（返回 `a.operation`）。
- `NewBaseApplicationHandler` 签名新增 `operation enumor.ApplicationOperation` 参数。
- 各具体 handler 的 `New...` 构造函数只需在调用 `NewBaseApplicationHandler` 时额外传入 operation 常量，其余方法无需修改。
- 对于 operation == type 的 handler（大多数），operation 传入对应的 `OpXxx` 常量即可。

**为什么不在接口层做 default**：Go 接口无默认实现，只能通过组合的 struct 实现，此方案已是最优路径。

### 3. `bkBizIDs` 判断改用 operation

**背景**：`create.go` 中按 `applicationType`（即 `type`）判断是否记录业务 ID。随着 `type` 变为粗粒度，用 `operation` 判断更准确，也能覆盖新增细粒度操作的判断。

**决策**：`createApplication` 函数改用 `handler.GetOperation()` 做判断，维护一个 operation 白名单集合（set），包含所有需要记录 `bkBizIDs` 的 operation 值。

```go
// 需要记录 bkBizIDs 的操作集合
var needBkBizIDsOps = map[enumor.ApplicationOperation]struct{}{
    enumor.OpCreateCvm:          {},
    enumor.OpCreateDisk:         {},
    enumor.OpCreateVpc:          {},
    enumor.OpCreateLoadBalancer: {},
    enumor.OpAddAccount:         {},
    // 新增细粒度操作按需添加
}
```

### 4. 查询接口：`ApplicationResp` 补充 `operation` 字段

**决策**：
- `ApplicationResp` 新增 `Operation string` 字段。
- `ListBizApplications` 接口不做额外限制，前端可在 filter 中传入 `operation` 字段做过滤（复用已有通用 filter 机制，无需专门改查询逻辑）。

### 5. 存量数据迁移：UPDATE 脚本

**决策**：提供 SQL 迁移脚本，对 `operation = ''` 的存量记录执行 `UPDATE application SET operation = type WHERE operation = ''`，保证历史数据向后兼容。迁移在部署新版本前执行。

## Risks / Trade-offs

- **[风险] 构造函数参数变更导致编译失败** → 所有调用 `NewBaseApplicationHandler` 的地方需同步更新，通过全量编译验证覆盖。
- **[风险] 存量数据 `operation` 为空** → `InsertValidate` 已要求 `operation` 非空，但存量记录不受插入校验约束；需在上线前执行迁移脚本，避免查询时返回空 operation。
- **[取舍] `approval_process` 仍以 `type` 为维度** → 多个细粒度 operation 共享同一 `type` 下的审批流，即共用同一 `service_id`；如需为某个 operation 单独配置审批流，需后续再调整 `approval_process` 表结构（本期 Non-Goal）。

## Migration Plan

1. 部署前：执行存量数据迁移 SQL：
   ```sql
   UPDATE application SET operation = `type` WHERE operation = '' OR operation IS NULL;
   ```
2. 部署新版本服务（cloud-server、data-service）。
3. 验证：抽样查询 `application` 表，确认 `operation` 字段均有值；通过 `ListBizApplications` 接口以 `operation` 为条件查询，验证返回正确。
4. 回滚策略：新字段向后兼容，若需回滚服务版本，旧版本只读 `type` 字段，`operation` 字段对旧版本无影响。

## Open Questions

- `operate_sub_account` 下的细粒度 operation（如 `create_sub_account`、`update_sub_account`）的全量枚举值，需由业务侧确认并补充到 `ApplicationOperation` 常量中。
- 其他存在"多操作共用一 type"场景的 type 是否还有，需排查后补充对应的 operation 常量。
