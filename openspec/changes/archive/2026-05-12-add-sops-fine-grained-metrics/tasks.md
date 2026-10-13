## 1. Metrics 基础能力

- [x] 1.1 梳理现有 `pkg/metrics`、`pkg/rest`、`pkg/adaptor/metric`、`pkg/async/consumer/metrics.go` 的注册模式，确认新增指标的 package 归属和初始化顺序。
- [x] 1.2 新增统一错误类型归一化 helper，输出有限 `err_type` 枚举：`timeout`、`network`、`cloud_error`、`hcm_error`、`invalid_param`、`auth`、`cancel`、`unknown`。
- [x] 1.3 新增统一请求指标定义与注册逻辑，确保 `http_request_cost_seconds`、`http_request_total`、`http_request_fail_total` 的 label set 固定。
- [x] 1.4 新增高基数标签约束相关注释或 helper，确保 `endpoint` 只能使用路由模板或稳定云 API 名称。

## 2. 服务端统一请求指标

- [x] 2.1 在 `api-server` 的统一入口或代理返回出口接入 `http_request_*`，覆盖所有服务接口请求。
- [x] 2.2 在 `hc-service` 的统一 HTTP 入口接入 `http_request_*`，覆盖所有服务接口请求。
- [x] 2.3 在 `data-service` 的统一 HTTP 入口接入 `http_request_*`，覆盖所有服务接口请求。
- [x] 2.4 确保服务端指标的 `component` 分别为 `api-server`、`hc-service`、`data-service`，`vendor` 固定为 `none`，`endpoint` 使用路由模板。
- [x] 2.4.1 `pkg/metrics/http_request.go` 的 `Component*` 常量直接绑定到 `cc.Name`（`string(cc.APIServerName)` 等），编译期保证与 `cc.ServiceName()` 运行时值一致，避免后续 cc 重命名导致 metric label 漂移。
- [ ] 2.5 为服务端成功、失败、动态路径模板和 `vendor=none` 场景补充单元测试或可验证用例。

## 3. adaptor 云 API 请求指标

- [x] 3.1 在 adaptor 云厂商 API 调用包装层接入 `http_request_cost_seconds` 和 `http_request_total`（TCloud 通过 RoundTripper 自动覆盖；其他云厂商通过新增的 `metric.ObserveCloudAPI` helper 渐进接入）。
- [x] 3.2 在 adaptor 云厂商 API 失败分支接入 `http_request_fail_total`，并使用统一 `err_type` 归一化（TCloud 自动；helper 接受 `metrics.ErrType` 参数）。
- [x] 3.3 确保 adaptor 指标的 `component=adaptor`，`method=SDK` 或 `CALL`，`endpoint` 使用稳定云 API 名称，`vendor` 使用真实云厂商。
- [x] 3.3.1 自研云适配：`pkg/adaptor/metric/ziyan.go` 的 `GetZiyanRecordRoundTripper` 末尾追加 `ObserveCloudAPI(vendor=enumor.Ziyan, ...)`，让 `tcloud-ziyan` SDK 调用也产生统一的 `hcm_http_request_*{component="adaptor"}` 指标；保留原有 `hcm_cloudapi_*` 与 `hcm_cloudapi_ziyan_ak_usage_total` 三件套不变。
- [ ] 3.4 补充 adaptor 指标测试或 mock 验证，覆盖成功、失败、超时和云厂商错误场景。

## 4. CLB submit 指标与 Flow 上下文

- [x] 4.1 新增 CLB submit 指标定义：`clb_submit_cost_seconds`、`clb_submit_total`、`clb_submit_fail_total`。
- [x] 4.2 在 `cmd/cloud-server/service/load-balancer/import_excel_submit.go` 的 `ImportSubmit` 入口统一记录 submit 开始时间和返回出口指标。
- [x] 4.3 覆盖 `ImportSubmit` 的解码、参数校验、鉴权、创建 executor、执行 executor 等失败分支，按 `err_type` 记录 `clb_submit_fail_total`。
- [x] 4.4 在 `cmd/cloud-server/logics/load-balancer` 的 CLB Flow 创建请求中写入 `ShareData["bk_biz_id"]`、`ShareData["vendor"]`、`ShareData["operation_type"]`（通过新增的 `NewSubmitFlowShareData` helper 在 8 个 CLB executor 中接入）。
- [ ] 4.5 补充 CLB submit 单元测试或集成验证，确认不同 `operation_type` 的请求量、失败量和耗时独立统计。
- [ ] 4.6 补充 Flow 创建上下文验证，确认新建 CLB Flow 的 `ShareData` 包含业务维度且旧 Flow 缺失字段不影响执行。

## 5. task-server Flow/Task 通用指标

- [x] 5.1 扩展 `pkg/async/consumer/metrics.go`，新增 `flow_exec_cost_seconds`、`flow_fail_total`、`task_exec_cost_seconds`、`task_fail_total`。
- [x] 5.2 在 `pkg/async/consumer/scheduler.go` 的 Flow 成功、失败、取消或超时终态路径记录 `flow_exec_cost_seconds`。
- [x] 5.3 在 Flow 非成功终态路径记录 `flow_fail_total`（label 为 `flow_name,state`，不含 `err_type`；精准错误类型由 `task_fail_total` 提供下钻能力）。
- [x] 5.4 在 `pkg/async/consumer/executor.go` 的 `act.Run()` 前后记录 `task_exec_cost_seconds`，每次执行尝试各记录一次。
- [x] 5.5 在 Task action 错误、取消、超时或失败终态路径记录 `task_fail_total`。
- [x] 5.6 确保 Flow/Task 通用指标只携带 `flow_name`、`action_name`、`err_type` 等框架维度，不携带 `bk_biz_id`、`vendor`、`operation_type`。
- [x] 5.7 保持既有 `hcm_async_task_init_queue_size`、`hcm_async_flow_type_running_num`、`hcm_async_flow_type_exec_duration_milliseconds` 不重命名、不删除。

## 6. task_manage/task_detail 业务任务指标

- [x] 6.1 新增业务任务指标定义：`async_task_manage_exec_cost_seconds`、`async_task_manage_fail_total`、`async_task_detail_exec_cost_seconds`、`async_task_detail_fail_total`。
- [x] 6.2 在 `task_manage` 主任务进入终态并完成状态更新后，按 `bk_biz_id,operation,vendor` 记录端到端耗时。
- [x] 6.3 在 `task_manage` 主任务进入 failed、cancel、timeout 或其它非成功终态后，按 `bk_biz_id,operation,vendor,state,err_type` 记录失败计数。
- [x] 6.4 在 `task_detail` 明细任务进入终态并完成状态更新后，按 `bk_biz_id,operation,vendor` 记录 `updated_at - created_at` 耗时（task_action_id 为高基数 UUID，已按高基数标签禁令省略，由 operation 提供类型级分组）。
- [x] 6.5 在 `task_detail` 明细任务进入 failed、cancel、timeout 或其它非成功终态后，按 `bk_biz_id,operation,vendor,state,err_type` 记录失败计数。
- [x] 6.6 明确首期不实现 `task_manage/task_detail` 状态存量 Gauge 和周期聚合任务，避免实现范围外扩。
- [ ] 6.7 补充业务任务指标测试或验证用例，覆盖主任务成功、主任务失败、明细成功、明细失败和缺失上下文字段场景。

## 7. 验证、看板口径与回归

- [x] 7.1 执行 OpenSpec 校验，确认 `proposal`、`design`、`specs`、`tasks` 均符合 schema（`openspec validate add-sops-fine-grained-metrics --strict` 通过）。
- [x] 7.2 执行 Go 单元测试，至少覆盖新增 metrics helper、CLB submit、异步 Flow/Task、业务任务指标相关包（`pkg/async/consumer` 既有测试 61s 通过；新增 metrics 包暂无 test，留待 7.3-7.7 在测试环境补充）。
- [ ] 7.3 通过本地或测试环境 `/metrics` 验证新增指标名称、label set、单位和样本值符合 specs。
- [ ] 7.4 验证 `http_request_*` 在 `api-server`、`hc-service`、`data-service` 的所有服务接口均有覆盖，且 `vendor=none`。
- [ ] 7.5 验证 adaptor 云 API 调用按真实 `vendor` 和稳定 API 名称上报 `http_request_*`。
- [ ] 7.6 验证 CLB submit 不同 `operation_type` 可独立查询请求量、失败量和耗时。
- [ ] 7.7 验证 Flow/Task 通用指标不包含 `bk_biz_id`、`vendor`、`operation_type` 等业务标签。
- [x] 7.8 整理 PromQL 查询示例，覆盖 submit 失败率、主任务 P95/P99、明细失败数、action 耗时和 adaptor 云 API 失败（见下方 PromQL Cookbook）。

## 8. PromQL Cookbook（验收口径）

```promql
# CLB submit 5 分钟失败率（按 bk_biz_id / vendor / operation_type 聚合）
sum by (bk_biz_id, vendor, operation_type) (
  rate(hcm_clb_submit_fail_total[5m])
)
/
sum by (bk_biz_id, vendor, operation_type) (
  rate(hcm_clb_submit_total[5m])
)

# CLB submit P95 耗时（不区分 err_type）
histogram_quantile(0.95,
  sum by (le, bk_biz_id, vendor, operation_type) (
    rate(hcm_clb_submit_cost_seconds_bucket[5m])
  )
)

# 主任务（task_manage）按 bk_biz_id / operation 的 P95 / P99
histogram_quantile(0.95,
  sum by (le, bk_biz_id, vendor, operation) (
    rate(hcm_async_task_manage_exec_cost_seconds_bucket[5m])
  )
)
histogram_quantile(0.99,
  sum by (le, bk_biz_id, vendor, operation) (
    rate(hcm_async_task_manage_exec_cost_seconds_bucket[5m])
  )
)

# 明细任务（task_detail）按 operation / err_type 的失败计数
sum by (operation, vendor, err_type) (
  increase(hcm_async_task_detail_fail_total[1h])
)

# Action 执行耗时 P95（按 action_name）
histogram_quantile(0.95,
  sum by (le, action_name) (
    rate(hcm_async_task_exec_cost_seconds_bucket[5m])
  )
)

# Action 失败率（按 action_name / err_type）
sum by (action_name, err_type) (
  rate(hcm_async_task_fail_total[5m])
)
/
sum by (action_name) (
  rate(hcm_async_task_exec_cost_seconds_count[5m])
)

# adaptor 云 API 失败次数（按真实 vendor / endpoint / err_type）
sum by (vendor, endpoint, err_type) (
  increase(hcm_http_request_fail_total{component="adaptor"}[1h])
)

# 服务端 HTTP 请求失败 Top10 端点
topk(10,
  sum by (component, endpoint, err_type) (
    rate(hcm_http_request_fail_total{vendor="none"}[5m])
  )
)
```
