## Why

当前导出监听器的预检（`POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/export/pre_check`）对四层监听器、七层监听器、七层 URL 规则、四层 RS、七层 RS 各设了 5000 条的固定上限，且这些上限与勾选的负载均衡数量无关。

这导致一个实际问题：用户只勾选少量负载均衡（典型场景是导出一个大规格 LB 的全部配置）时，只要该 LB 下的监听器或 RS 超过 5000 条就无法导出，而这恰恰是最需要精确导出的场景。现有上限的设计初衷是防止「批量勾选大量 LB」造成超大导出，对「少量 LB 精确导出」属于误伤。

## What Changes

- 在导出监听器预检逻辑中新增按勾选负载均衡数量分流的判断：
  - 勾选的负载均衡数量 **≤ 5** 时，跳过所有数量类限制（四层/七层监听器数量、七层 URL 规则数量、四层/七层 RS 数量），预检直接通过。
  - 勾选的负载均衡数量 **> 5** 时，完全保持现有限制逻辑与错误提示不变。
- 归属正确性校验（请求中的 `lbl_ids` 是否确实属于对应 `lb_id`）在两种情况下都保留，不受数量分流影响。
- 请求参数层的数量类校验同样按勾选负载均衡数量分流：≤ 5 时跳过 `listeners` 数量、去重后 `lbl_ids` 总数、单个 `listeners` 元素的 `lbl_ids` 数量这三项限制；> 5 时保持现有限制值不变。`listeners` 非空与 `lb_id` 非空属于正确性校验，两种情况下都执行。
- 将导出阶段按监听器 ID 查询的三处未分批 `IN` 条件改造为分批查询，消除下游 data-service DAO 单个 `IN` 最多 10000 个元素的隐性上限。若不做该改造，参数层放开后监听器 ID 数量超过 10000 时会在 DAO 层报出用户无法理解的底层错误，放开限制等于没有落地。
- 新增负载均衡数量阈值常量，不使用魔数。
- 同步更新接口文档 `export_listener_pre_check.md`，补充分流规则说明，并修正文档中与代码不一致的参数长度描述。

本次变更不涉及请求/响应结构调整，不涉及数据库变更，无 **BREAKING** 变更：勾选数量 > 5 时行为与现状完全一致，≤ 5 时属于限制放宽（原先被拦截的请求现在会通过）。

## Capabilities

### New Capabilities

- `listener-export-precheck`: 导出监听器及其下属资源（URL 规则、RS）的预检规则，包含归属正确性校验、按勾选负载均衡数量分流的数量限制规则，以及预检结果的返回契约。

### Modified Capabilities

无。现有 `openspec/specs/` 中不存在覆盖监听器导出预检的能力，因此以新增能力承载。

## Impact

**受影响服务层**：cloud-server（Service Layer）与协议定义层 `pkg/api`。不涉及 data-service、hc-service 等 Resource Layer 的接口变更与 DAO 改动，仅调整 cloud-server 侧调用 data-service 的查询分批方式。

**受影响代码**：

- `pkg/api/cloud-server/load-balancer/load_balancer.go` — `ExportListenerReq.Validate` 与 `ExportListener.Validate` 的数量类校验按勾选负载均衡数量分流。
- `cmd/cloud-server/logics/load-balancer/export_listener_excel.go` — `listenerExporter.PreCheck` 增加数量限制分流判断；导出阶段三处按监听器 ID 查询改为分批。
- `pkg/criteria/constant/clb.go` — 新增负载均衡数量阈值常量。
- `docs/api-docs/web-server/docs/biz/load-balancer/export/export_listener_pre_check.md` 与 `export_listener.md` — 文档同步。

**受影响接口**（两者共用同一段预检逻辑，行为同步变化）：

- `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/export/pre_check`
- `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/export`（内部同样调用 `PreCheck`）

**不改动**：

- 前端无需改动，预检的 `pass` / `reason` 契约不变。
- 鉴权逻辑不变，仍为业务下负载均衡的 Update 权限校验。
- 导出数据组装与 Excel/zip 生成逻辑不变。

**已知风险**（本次不在方案内解决，仅记录）：导出链路是同步 HTTP 下载且数据全量驻留内存，放开限制后若 ≤ 5 个负载均衡下的数据量极大，存在内存占用升高与请求耗时变长的风险。是否需要总量兜底上限或异步化导出，作为后续独立议题跟进。
