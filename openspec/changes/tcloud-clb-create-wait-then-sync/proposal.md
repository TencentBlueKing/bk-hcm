## Why

公有云（腾讯云）购买 CLB 时，hc-service 在创建任务成功后先写入本地记录，再立即调用 `res-sync/tcloud` 与云上核对。`DescribeTaskStatus` 返回成功到 `DescribeLoadBalancers` 能查到新实例之间，存在秒级延迟。核对时云上查不到，刚写入的本地记录就会被当作"云上已删除"清理，业务视角列表因此找不到刚买的 CLB。另外，创建任务轮询超时时 `CreateLoadBalancer` 会读取空结果，触发空指针异常。

本变更把创建链路改为"先等云上可见，再同步建记录，最后补业务"：本地不再提前写记录，也就不存在误删。自研云后续采用同样的顺序，两个云的创建链路保持一致。本变更是需求单 1069995598138578692 下公有云子需求 1069995598138704822 的实现载体（Wave 1），产出的公共等待函数供自研云子需求复用。

## What Changes

- 新增公共等待函数：创建任务成功后，按退避间隔分批查询 `DescribeLoadBalancers`，直到本次新建的 CLB 全部可见，或到达 30 秒上限；返回仍查不到的云 ID。
- 公有云创建入口 `BatchCreateTCloudClb` 改为：等待云上可见 → 同步（由同步按云上数据新建本地记录）→ 补业务。
- **删除**创建前写本地记录的步骤（`createTCloudDBLoadBalancer`）。
- 补业务：把本次云 ID 中业务为"未分配"的本地记录，改为请求中的业务；失败不重试，直接返回错误，交付失败，错误中带上云 ID。
- 同步备可用区：`res-sync/tcloud` 新建、更新本地记录时按云上 `BackupZoneSet` 写入备可用区，变更比较也纳入备可用区；云上未返回时不覆盖本地已有值。data-service 批量更新接口新增可选的备可用区字段。
- 等待超时：同步并补业务已可见的部分后，返回错误，交付失败，错误中带上仍查不到的云 ID。公有云本期不做异步重试。
- `pkg/adaptor/tcloud/clb.go` 的 `CreateLoadBalancer` 在读取轮询结果前判空，结果为空时返回含 RequestId 的错误。

`res-sync/tcloud` 的删除核对逻辑不改。不新增、不修改对外接口；data-service 内部批量更新接口新增可选字段，向后兼容。

行为变化：

- 等待超时、补业务失败两种情况下，交付由"成功"变为"失败"，错误信息中包含云 ID。
- 周期同步会把云上有、本地缺失或不一致的备可用区写入本地记录，存量记录（含控制台创建、导入的 CLB）在下一轮同步后补齐。

## Capabilities

### New Capabilities

- `tcloud-clb-create-wait-then-sync`：公有云 CLB 创建后等待云上可见、再同步建记录并补业务的创建链路。

### Modified Capabilities

（无。现有 spec 中没有覆盖 CLB 创建后同步的需求。）

## Impact

- **代码**
  - `pkg/adaptor/tcloud/`：新增等待函数及单测；`clb.go` 的 `CreateLoadBalancer` 判空。
  - `pkg/adaptor/types/poller_option.go`、`pkg/criteria/constant/clb.go`：等待参数与重试次数常量。
  - `cmd/hc-service/service/load-balancer/tcloud.go`：创建入口改造；删除 `createTCloudDBLoadBalancer`。
  - `cmd/hc-service/service/load-balancer/tcloud_lb_fill.go`：补业务逻辑及单测。
  - `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go`：新建、更新、变更比较纳入备可用区，及单测。
  - `pkg/api/data-service/cloud/load_balancer.go`、`cmd/data-service/service/cloud/load-balancer/update.go`：批量更新支持备可用区。
- **数据**：本地记录的名称、可用区、备可用区、标签等字段改由同步按云上数据写入，不再用请求参数预填。周期同步会批量补齐存量记录的备可用区。
- **耗时**：交付最多增加约 33 秒。审批流交付异步执行，不受影响；免审直接交付在用户请求内同步执行，前端或网关超时需在评审时确认（父需求 Q-010，非阻塞）。
- **云 API 调用量**：等待期间每轮每 20 个 ID 发 1 次查询。
- **依赖**：无新增外部依赖；补业务、同步备可用区复用 data-service 现有接口 `TCloud.LoadBalancer.BatchUpdate`（新增可选字段）。
- **下游**：内部版合入后，自研云子需求（1069995598138704782）复用等待函数，并改为同样的"先等待、再同步"顺序。
