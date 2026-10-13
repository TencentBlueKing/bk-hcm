## Context

导出监听器的校验分布在两层。

**参数层** `pkg/api/cloud-server/load-balancer/load_balancer.go`：

- `ExportListenerReq.Validate()` — `listeners` 非空、`listeners` 数量 ≤ 5000、去重后 `lbl_ids` 总数 ≤ 5000
- `ExportListener.Validate()` — `lb_id` 非空、单个元素的 `lbl_ids` ≤ 100（复用 `constant.BatchOperationMaxLimit`）

**业务层** `cmd/cloud-server/logics/load-balancer/export_listener_excel.go` 的 `listenerExporter.PreCheck`：

```
PreCheck
  ├─ 1. checkClbListenerRel   请求中的 lbl_id 是否属于对应的 lb_id（正确性校验，已按 500 分批）
  ├─ 2. checkListenerCount    四层监听器 ≤ 5000 且 七层监听器 ≤ 5000
  ├─ ── only_export_listener == true 时在此 return nil ──
  ├─ 3. checkRuleCount        七层 URL 规则总数 ≤ 5000
  └─ 4. checkRsCount          四层 RS ≤ 5000 且 七层 RS ≤ 5000
```

各步之间是短路的 AND 关系，任一步超限即返回错误；`pre_check` 接口把该错误转成 `{pass: false, reason}`，`export` 接口则直接返回错误终止导出。限制值全部为 `pkg/criteria/constant/clb.go` 中的固定常量 5000。

前端有两个入口共用该接口：在负载均衡列表勾选 LB 时请求体为 `[{lb_id}, ...]`（不带 `lbl_ids`）；在监听器列表勾选监听器时按 `lb_id` 聚合并逐个枚举监听器 ID，请求体为 `[{lb_id, lbl_ids}]`。因此「单个负载均衡下勾选超过 100 个监听器导出」会被参数层的单元素 `lbl_ids ≤ 100` 直接拦下，业务层的分流放开对该场景无效——参数层必须一起处理。

**下游存在隐性上限**：`export_listener_excel.go` 中多处把 `lblIDs` / `lbIDs` 整个切片塞进 `tools.RuleIn` 而未分批，而 data-service 侧监听器、URL 规则、目标组关联、RS 四张表的 DAO 都以 `filter.MaxInLimit(constant.CLBTopoFindInLimit)` 限制单个 `IN` 最多 10000 个元素。当前参数层的 5000 上限恰好把这些查询护在 10000 以内；一旦参数层放开而查询不分批，监听器 ID 超过 10000 时会在 DAO 报 `invalid in operator's value, at most have 10000 elements`，失败点从可读的 `reason` 退化为底层错误。

## Goals / Non-Goals

**Goals:**

- 勾选负载均衡数量 ≤ 5 时，参数层与业务层的数量类限制全部不生效，导出不因数量被拒。
- 勾选负载均衡数量 > 5 时，两层的限制逻辑与错误提示与现状完全一致，不产生任何行为差异。
- 正确性校验（`listeners` 非空、`lb_id` 非空、监听器归属）在两种情况下都执行。
- 消除按监听器 ID 查询时的 10000 元素隐性上限，使「不做限制」在链路上真正成立。
- 阈值以命名常量承载，具备一键回退到旧行为的能力。

**Non-Goals:**

- 不引入总量兜底上限，不做限制值随勾选数量分档放大的梯度设计。
- 不改造导出的数据装载方式（全量内存装载）与同步下载模式。
- 不改造 data-service 的 DAO 与 `CLBTopoFindInLimit` 常量，分批在 cloud-server 侧完成。
- 不改动前端、鉴权逻辑与响应结构。

## Decisions

### 决策一：两层使用同一个判定口径 `len(GetAllLbIDs()) <= 阈值`

`ExportListenerReq` 上已有三种可用取值，两层统一选用 `GetAllLbIDs()`：

| 取值方式 | 含义 | 是否采用 |
|---|---|---|
| `len(r.Listeners)` | 请求体元素个数，同一 LB 可能被拆成多个元素而重复计数 | 否 |
| `GetPartLbAndLblIDs()` 的第一个返回值 | 仅包含未指定 `lbl_ids` 的 LB，会漏掉监听器维度导出的 LB | 否 |
| `GetAllLbIDs()` | 全部元素的 `lb_id` 去重后集合 | **是** |

`GetAllLbIDs()` 的语义正是「本次请求涉及的去重后负载均衡数量」，与需求中「勾选负载均衡个数」一致，且已被 `getLbs` 复用，无需新增方法。

参数层能够自行计算该值（`GetAllLbIDs()` 就定义在 `ExportListenerReq` 上，不依赖 client），因此分流可以在参数层落地，不需要把校验后移到业务层。

### 决策二：参数层按分流跳过数量类校验，保留正确性校验

`ExportListenerReq.Validate()` 与 `ExportListener.Validate()` 中的校验按性质拆成两类：

| 校验项 | 性质 | ≤ 阈值时 | > 阈值时 |
|---|---|---|---|
| `listeners` 非空 | 正确性 | 执行 | 执行 |
| `lb_id` 非空 | 正确性 | 执行 | 执行 |
| `listeners` 数量 ≤ 5000 | 数量 | 跳过 | 执行 |
| 去重后 `lbl_ids` 总数 ≤ 5000 | 数量 | 跳过 | 执行 |
| 单元素 `lbl_ids` ≤ 100 | 数量 | 跳过 | 执行 |

单元素 `lbl_ids` 的限制当前复用 `constant.BatchOperationMaxLimit`（通用批量操作上限 100）。`ExportListener.Validate()` 是元素级方法、拿不到整个请求的 LB 数量，因此该项的分流判断由 `ExportListenerReq.Validate()` 承担：命中分流时不再逐元素调用数量校验，只校验 `lb_id` 非空。

考虑过在 `ExportListener` 上增加参数传递是否跳过的入参，但会污染元素级校验方法的签名，不采用。

### 决策三：业务层分流判断放在归属校验之后、数量校验之前

```
PreCheck
  ├─ checkClbListenerRel              ← 始终执行
  ├─ skipCountLimit() → return nil    ← 新增
  ├─ checkListenerCount
  ├─ ── only_export_listener → return nil ──
  ├─ checkRuleCount
  └─ checkRsCount
```

**为什么放在归属校验之后**：归属校验是正确性判断而非数量限制，若一并跳过，非法 `lbl_id` 会在导出阶段的 `writeTCloudLayer4Listener` 等函数中触发 `can not get clb by lb id` 一类内部错误，对用户不可理解，且已经白做了全量数据查询。

**为什么不放在三个 check 函数内部各自判断**：同一条件重复三处，后续新增检查项容易漏改。

判断条件抽成 `listenerExporter` 的一个方法（例如 `skipCountLimit() bool`），只读取 `l.params`、不依赖 client，可被表驱动单测直接覆盖边界值；`PreCheck` 整体因依赖 `*client.ClientSet` 具体类型无法 mock，只能靠手动验证。

### 决策四：仅改造导出阶段的三处未分批查询，count 路径不动

按监听器 ID 查询的位置分为两类，只有导出阶段的会在分流命中后收到无界输入：

**需要分批（导出阶段，`Export` 链路）：**

| 位置 | 查询 | 改造方式 |
|---|---|---|
| `getTCloudListenersByProtocol` 的 `lblIDs` 分支 | `RuleIn("id", lblIDs)` | 外层套 `slice.Split` 循环，结果并入以 ID 为键的 map，天然去重 |
| `getTCloudRulesByRuleType` 的 `lblIDs` 分支 | `RuleIn("lbl_id", lblIDs)` | 同上 |
| `getTgLblRelClassifyProtocol` / `getTgLblRelRule` | `lb_id IN (...) OR lbl_id IN (...)` | 拆成 `lb_id` 维度与 `lbl_id` 维度两段独立分批查询，结果按关联记录 ID 去重后合并 |

**不需要分批（count 路径，`PreCheck` 链路）：** `getListenerCountRule`、`checkRuleCount`、`getRsCountRule`。理由是分流命中时这三个 count 校验整体被跳过、根本不会执行；未命中分流时参数层的 5000 上限仍然生效，`IN` 元素数不会超过 10000。该结论依赖「参数层与业务层使用同一阈值口径」这一前提，属于隐性耦合，需在代码注释中写明，避免后续单独调整某一层的阈值时踩坑。

**为什么第三处要拆成两段查询**：现状用单条 `OR` 表达式一次查完，同时命中两个条件的关联记录只会返回一次。若保持 `OR` 结构只对 `lblIDs` 分批，`lb_id` 条件会在每一批里重复命中，导致关联记录重复、进而 RS 在导出文件中重复。拆成两段后必须按关联记录 ID 去重才能保持与现状一致的结果。前端实际不会在一次请求中混合两种形态（要么全带 `lbl_ids`，要么全不带），但接口允许混合，不能依赖调用方约束。

**分批粒度**：沿用现有代码的 `core.DefaultMaxPageLimit`（500），与 `checkClbListenerRel`、`getLbs` 保持一致，且远低于 `CLBTopoFindInLimit`（10000），留足余量。

### 决策五：阈值常量化，并使其可作为回退开关

在 `pkg/criteria/constant/clb.go` 现有导出限制常量组中新增：

- 名称：`ExportSkipLimitLbCount`，值 `5`
- 注释：说明为「勾选负载均衡数量不超过该值时，导出不做数量限制」，并注明参数层与业务层共用该阈值

判断使用 `<=`，因此把该常量改为 `0` 即可让条件永不成立、完全恢复旧行为（请求至少包含一个 `lb_id`，`listeners` 非空校验已保证），无需回滚代码即可关闭该特性。

考虑过做成 `cc.CloudServer()` 配置项以便按环境调整，但当前导出相关限制全部为常量，单独把这一个做成配置会造成同类参数两套管理方式，本次保持一致，留作 Open Question。

### 决策六：新增日志记录实际数量

分流命中时以 Info 级别记录本次勾选的负载均衡数量与监听器 ID 数量，便于后续评估放开限制后的真实数据规模，为是否需要总量兜底提供依据。日志遵循项目规范，包含 `rid`。

## Risks / Trade-offs

**[放开限制后内存占用与耗时升高]** → 导出链路把选中范围内的监听器、URL 规则、RS 全量装载进内存，组装 Excel 行时还会再复制一份；`zip/excel.Operator` 会同时持有所有 xlsx 的 `excelize.File` 直到 `Save()`，而 `write()` 按 LB 每 5000 行切一个文件（`ExportClbOneFileRowLimit`）。同时导出是同步 HTTP 下载，受网关超时约束。本次不改造该链路，通过决策六的日志观察真实规模；若出现问题，可将阈值常量下调或置 0 快速止损。

**[分批后查询次数放大]** → 监听器 ID 数量很大时，每 500 个一批意味着查询次数线性增长（1 万个 ID 约 20 批，且监听器、URL 规则、关联记录三条链路各自都要分批），叠加原有的分页循环，总请求数可观，会进一步拉长导出耗时。分批粒度选 500 是与现有代码保持一致的结果，若实测偏慢可在不超过 `CLBTopoFindInLimit` 的前提下调大。

**[关联记录去重遗漏会导致 RS 重复导出]** → `getTgLblRelClassifyProtocol` 从单条 `OR` 查询拆成两段后，若漏掉按关联记录 ID 去重，混合传参场景下同一条关联会被计入两次，最终导出的 RS 出现重复行。属于本次改造中最容易出错的一处，需要在验证清单中专门覆盖「同时传 `lb_id` 与该 LB 下监听器 `lbl_ids`」的混合场景。

**[请求体大小无约束]** → 参数层放开后 `lbl_ids` 数量无上限，请求体可能达到数 MB。仓库内 Go 代码没有请求体大小限制，实际约束在网关/nginx 层，超限时的报错对用户不友好。本次不处理，记录为已知限制。

**[两层阈值口径存在隐性耦合]** → count 路径不分批的安全性依赖「参数层与业务层使用同一阈值」。若后续只调整其中一层，count 查询可能收到超过 10000 个 ID 而在 DAO 失败。通过共用同一常量与代码注释缓解。

**[限制随勾选数量非单调]** → 勾 5 个大 LB 可导出的数据量可能远超勾 6 个小 LB 被拦截的数据量，用户可能感到「选得少反而能导更多」。经确认本次接受该取舍，不做梯度设计。

**[无总量兜底]** → ≤ 5 个 LB 时理论上没有任何数量上限，极端数据量下会打到内存或超时才失败，且失败原因对用户不友好。作为后续独立议题跟进。

## Migration Plan

无数据库变更、无接口结构变更，属于业务逻辑与查询方式调整，随 cloud-server 常规发布即可，不需要数据迁移或前后端联合发布。

回滚策略分两级：优先将 `ExportSkipLimitLbCount` 改为 `0` 关闭特性（参数层与业务层同时恢复到变更前的限制，分批查询改造对原有行为无影响，可保留）；必要时回退代码提交。

## Open Questions

- 阈值 `5` 后续是否需要提升为配置项以便按环境调整？
- 是否需要为 ≤ 5 个 LB 的场景补一个很宽松的总量兜底上限（例如 5 万），以替代「完全无限制」？
- 分批粒度 500 在监听器 ID 数量上万时是否偏小、需要调大以减少请求次数？建议以实测数据决定。
