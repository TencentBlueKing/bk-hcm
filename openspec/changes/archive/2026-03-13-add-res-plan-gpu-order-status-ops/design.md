## Context

GPU需求提报主单（`res_plan_demand_gpu_order`）和子单（`res_plan_demand_gpu_suborder`）的 CRUD 接口已在 data-service 层实现，data-service client 也已封装好对应方法。当前缺少的是 woa-server 层的状态变更业务接口，以及审计记录。

状态流转规则：
- 主单：INIT（待评审）→ PENDING（评审中）→ REJECT_ALL（全部已驳回）/ TERMINATE（已终止）
- 业务下终止：允许 INIT 或 REJECT_ALL 状态的主单终止
- 资源下操作：INIT → PENDING（评审中），PENDING → REJECT_ALL（驳回），PENDING → TERMINATE（终止）
- 子单状态与主单联动：主单终止时子单同为 TERMINATE；主单改评审中时子单同为 PENDING；整单驳回时主单为 REJECT_ALL，子单为 REJECT

## Goals / Non-Goals

**Goals:**
- 在 woa-server 实现 4 个状态变更接口（3 资源下 + 1 业务下）
- 抽取公共状态变更逻辑，避免重复代码
- 主单和子单状态联动更新
- 记录审计日志

**Non-Goals:**
- 不修改 data-service 层（直接复用现有 BatchUpdate/List 接口）
- 不涉及子单的单条状态变更（另有独立接口）
- 不涉及子单编辑逻辑

## Decisions

### 公共逻辑下沉到私有方法（不在 data-service）

4 个接口核心步骤完全相同：查主单→校验前置状态→写审计→更新主单→查子单→分批更新子单。将这部分逻辑抽取为 woa-server 内的私有方法 `batchUpdateGpuOrderStatus(kt *kit.Kit, ...)`，接受 `kit.Kit` 而非 `rest.Contexts`，与 HTTP 层解耦。

各 handler 只负责权限鉴权，其余交给公共方法：

```
handler (权限鉴权) → batchUpdateGpuOrderStatus(kt, orderIDs, 前置状态[], 主单目标状态, 子单目标状态)
```

### 子单分批更新

子单 BatchUpdate 的 max 限制为 100。单次请求最多 100 个主单，每主单可能有多个子单，合计可能超过 100。需使用 `slice.Split` 对子单列表分批（每批 100 条）调用更新接口。

### 审计时序：先写审计再更新数据

与项目内 `record.go` 的惯例一致，先调用 `CloudResourceUpdateAudit` 写审计，再调用 data-service 更新状态。审计记录按主单粒度（不记录子单）。

### 状态前置校验在 woa-server

在 woa-server 中 List 查出主单后立即做状态校验（快速失败），不将校验逻辑下沉到 data-service，保持 data-service 的通用性。

## Risks / Trade-offs

- **非原子性**：woa-server 先更新主单、再更新子单，两次调用之间若 data-service 出现故障可能导致主子单状态不一致。鉴于该场景下数据一致性要求不属于强事务级别，且现有 woa-server 普遍采用此模式，可接受。
- **子单查询量**：如果子单数量极多（极端情况下 100 主单 × N 子单），List 查询可能需要多次分页。当前方案假设子单总量在合理范围内（单次 List 可覆盖），若后续有大量子单场景需要补充分页逻辑。
