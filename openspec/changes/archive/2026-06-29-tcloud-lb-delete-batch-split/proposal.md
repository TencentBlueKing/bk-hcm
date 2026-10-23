## Why

腾讯云 `DeleteLoadBalancer` 接口限制单次调用最多传入 20 个 CloudID。当前 cloud-server 的 `buildLBDeletionTasks` 函数在按"账号+vendor+地域"分组生成删除任务时，将同组所有 LB ID 塞进一个 task 的 `IDs` 列表，未做上限分批。当某个 account+region 下有超过 20 个负载均衡需要删除时，task 执行透传到 hc-service 层触发 `Validate()` 校验报错（`"batch delete limit is 20"`），导致整批删除失败。

## What Changes

- 在 `cmd/cloud-server/service/load-balancer/delete.go` 的 `buildLBDeletionTasks` 函数中，对每个 account+region 分组的 IDs 再按 `constant.BatchListenerMaxLimit`（= 20）拆分为多个独立 task
- 新增 `hcm/pkg/criteria/constant` 包的 import（当前文件未引用该包）
- 保持 hc-service 的 `BatchDeleteLoadBalancerReq.Validate()` 不变，作为最后防线
- 保持 task-server 的 `DeleteLoadBalancerAction.Run` 透传逻辑不变

## Capabilities

### Modified Capabilities

- `lb-batch-delete`: 批量删除负载均衡的 task 生成逻辑从"每 account+region 一个 task"改为"每 account+region 每 20 个 LB ID 一个 task"，确保每个 task 的 IDs 数量不超过腾讯云接口限制。

## Impact

- **Files changed**:
  - `cmd/cloud-server/service/load-balancer/delete.go` — 修改 `buildLBDeletionTasks` 函数，增加按批次大小拆分 IDs 的逻辑；新增 `hcm/pkg/criteria/constant` import
- **Behavior**: 对 ≤ 20 个 LB 的删除场景无任何行为变化（仍生成 1 个 task）；对 > 20 个的场景从"失败"变为"多 task 并行删除成功"
- **Breaking changes**: None — 现有调用方无感知，task 数量可能增多但 flow 调度自动处理
