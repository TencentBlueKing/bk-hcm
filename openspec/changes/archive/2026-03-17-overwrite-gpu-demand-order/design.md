## Context

GPU需求提报流程中，业务方上传Excel创建主单和子单后，主单进入评审流程。若评审方"整单驳回"，主单状态变为 `REJECT_ALL`，所有子单变为 `REJECT`。当前业务方只能终止旧单后重新创建，导致主单ID变更，不利于链路追踪。

现有代码中已具备：
- 创建主单和子单的完整链路（`CreateGpuDemandOrder`）
- Extension校验逻辑（`validateOrderExtensions`）
- 子单删除的data-service接口（`DeleteResPlanDemandGpuSubOrder`）
- 主单/子单状态校验模式（`validateGpuOrderStatuses`）

## Goals / Non-Goals

**Goals:**
- 提供覆盖上传接口，在保留主单ID的前提下替换已驳回子单
- 复用现有extension校验逻辑确保数据合法性
- 遵循现有状态机流转约束，只允许 `REJECT_ALL` 状态的主单进行覆盖
- 覆盖完成后主单回到 `INIT` 状态，子单为 `INIT` 状态，可重新进入评审流程

**Non-Goals:**
- 不支持部分覆盖（仅覆盖部分子单），每次覆盖都是全量替换
- 不修改主单的运营产品等属性，仅替换子单
- 不引入新的数据表或字段

## Decisions

### 1. 接口设计：业务视角的PATCH接口

**选择**: `PATCH /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/order/overwrite`

**理由**: 覆盖上传是对已有主单的局部更新（替换子单），语义上是PATCH。放在bizs路径下与CreateGpuDemandOrder保持一致，都是业务视角接口。

### 2. 请求体复用CreateGpuDemandOrderDetail

**选择**: 新增 `OverwriteGpuDemandOrderReq`，内含 `OrderID` 和 `Details []CreateGpuDemandOrderDetail`

**理由**: 覆盖上传的子单数据结构与创建时完全一致，复用detail类型避免重复定义。仅需额外的 `OrderID` 字段。

### 3. 删除后创建而非逐条更新

**选择**: 先删除所有旧子单，再批量创建新子单（delete-then-create）

**理由**: 
- 覆盖上传是全量替换，新旧子单数量和内容可能完全不同
- 逐条diff对比复杂度高，且业务上没有保留子单ID的需求
- 已有 `DeleteResPlanDemandGpuSubOrder` 和 `BatchCreateResPlanDemandGpuSubOrder` 接口

**替代方案**: 逐条更新子单——但由于子单数量可能增减，需要额外处理新增和删除的子单，实现更复杂。

### 4. 状态校验策略

**选择**: 
- 主单状态必须为 `REJECT_ALL`
- 子单状态必须全部为 `REJECT` 或 `TERMINATE`

**理由**: 只有全部驳回的单据才允许覆盖，避免覆盖正在评审中或已通过的子单。子单允许 `TERMINATE` 状态是因为业务方可能在驳回后先终止部分子单。

### 5. 覆盖后状态重置

**选择**: 主单状态重置为 `INIT`，新建子单状态为 `INIT`

**理由**: 覆盖后等同于重新提交，需要重新进入评审流程。

### 6. 逻辑层分工

**选择**: 
- Service层（`gpu_demand_excel.go`）处理HTTP解码、鉴权、biz归属校验
- 主单/子单状态校验在service层复用已有的 `validateGpuOrderStatuses` 模式
- Logic层（`gpu_demand_order.go`）处理extension校验、删除旧子单、创建新子单

**理由**: 与现有 `CreateGpuDemandOrder` 和 `BatchTerminateBizResPlanDemandGpuOrder` 保持一致的分层风格。

## Risks / Trade-offs

- **[非原子操作]** 删除旧子单成功但创建新子单失败时，会导致主单下无子单 → 缓解：创建失败时记录错误日志，前端可引导用户重新上传；后续可考虑引入事务或补偿机制
- **[状态竞争]** 覆盖操作执行期间若有其他操作（如终止）同时进行，可能导致状态不一致 → 缓解：前端操作互斥按钮，后端状态校验在操作前执行
- **[审计缺失]** 当前设计未写入审计日志 → 缓解：可在后续迭代中使用 `auditGpuOrderStatusUpdate` 模式补充审计记录
