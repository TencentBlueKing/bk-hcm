## ADDED Requirements

### Requirement: CLB submit 入口请求量和耗时指标
系统 MUST 在 CLB submit 接口 `/bizs/{bk_biz_id}/vendors/{vendor}/load_balancers/operations/{operation_type}/submit` 上报 `clb_submit_total` 和 `clb_submit_cost_seconds`。指标 label MUST 为 `bk_biz_id,vendor,operation_type`，统计范围 MUST 覆盖从 cloud-server 接收请求到返回任务创建结果的完整提交阶段。

#### Scenario: CLB submit 成功
- **WHEN** CLB submit 请求成功创建任务并返回 `TaskManagementID`
- **THEN** 系统按 `bk_biz_id,vendor,operation_type` 增加 `clb_submit_total` 并记录 `clb_submit_cost_seconds`

#### Scenario: CLB submit 不同操作隔离
- **WHEN** 同一业务和云厂商分别提交不同 `operation_type` 的 CLB 请求
- **THEN** 系统按不同 `operation_type` 分别统计请求量和耗时，不混合聚合

### Requirement: CLB submit 失败指标
系统 MUST 在 CLB submit 请求失败时上报 `clb_submit_fail_total`。指标 label MUST 为 `bk_biz_id,vendor,operation_type,err_type`，失败范围 MUST 覆盖请求解码、参数校验、鉴权、创建 executor、创建任务和更新任务详情等提交阶段错误。

#### Scenario: 参数或鉴权失败
- **WHEN** CLB submit 请求因参数校验或鉴权失败返回错误
- **THEN** 系统记录 `clb_submit_fail_total`，并将 `err_type` 归一化为 `invalid_param` 或 `auth`

#### Scenario: 任务创建失败
- **WHEN** CLB submit 请求在创建任务管理、Flow 或任务详情时失败
- **THEN** 系统记录 `clb_submit_fail_total`，并按有限 `err_type` 归类失败原因

### Requirement: CLB Flow 创建上下文传递
系统 MUST 在 CLB ImportExecutor 创建异步 Flow 时写入 `ShareData["bk_biz_id"]`、`ShareData["vendor"]`、`ShareData["operation_type"]`。这些字段 MUST 用于日志、任务上下文和必要时的关联排查，但不得默认作为通用 Flow/Task 指标标签。

#### Scenario: Flow 创建写入业务上下文
- **WHEN** CLB submit 创建任一 CLB 操作对应的异步 Flow
- **THEN** Flow 的 `ShareData` 中包含 `bk_biz_id`、`vendor` 和 `operation_type`

#### Scenario: 旧 Flow 缺少业务上下文
- **WHEN** 系统处理历史 Flow 或非 CLB Flow 且 `ShareData` 缺少这些字段
- **THEN** 任务执行不受影响，通用异步指标不因缺少业务上下文产生错误
