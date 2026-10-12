## 1. 常量与等待参数

- [x] 1.1 在 `pkg/criteria/constant/clb.go` 新增 `TCLBVisibleWaitTimeoutSec`、`TCLBVisibleWaitImmuneCount`、`TCLBVisibleWaitMinIntervalMS`、`TCLBVisibleWaitMaxIntervalMS`（design D3）
- [x] 1.2 在 `pkg/adaptor/types/poller_option.go` 新增 `NewCLBVisibleWaitPollerOption()`，使用 1.1 的常量

## 2. 公共等待函数

- [x] 2.1 新增 `pkg/adaptor/tcloud/clb_wait.go`：`LoadBalancerLister` 接口、`clbVisibleWaitHandler`（Poll 按 20 个分批；Done 结果始终非空，未可见 ID 放入 `UnknownCloudIDs`）、`WaitLoadBalancerVisible`（design D2）
- [x] 2.2 `WaitLoadBalancerVisible`：`opt` 为空用默认参数；`PollUntilDone` 报错时返回全部云 ID；有未可见 ID 时记 Warn（含云 ID、rid）；不返回错误
- [x] 2.3 新增 `pkg/adaptor/tcloud/clb_wait_test.go`，用 mock lister 和 1 秒上限覆盖：首轮全部可见只查 1 轮、第 2 轮可见提前结束、一直查不到返回该 ID 且耗时不超过上限 + 3 秒、25 个 ID 每轮 2 次查询且每次不超过 20 个、查询一直报错返回全部 ID
- [x] 2.4 加编译期断言，确认 `tcloud.TCloud` 满足 `LoadBalancerLister`

## 3. 轮询结果判空

- [x] 3.1 `pkg/adaptor/tcloud/clb.go` 的 `CreateLoadBalancer` 判空：`result == nil` 时返回带 RequestId、说明轮询超时任务未结束的错误（design D6）
- [x] 3.2 排查 `DeleteLoadBalancer`、`CreateLoadBalancerSnatIps`、`DeleteLoadBalancerSnatIps` 是否存在同样的空结果路径，结论记录到本文件末尾，不在本变更中修改

## 4. 补业务与同步备可用区

- [x] 4.1 新增 `cmd/hc-service/service/load-balancer/tcloud_lb_fill.go` 的 `fillTCloudLBBiz`：业务 ID 不大于 0 时跳过；按 vendor、账号、地域、云 ID 查本地记录；只把"未分配"记录改为请求业务；其它业务值记 Warn 不覆盖；通过 `TCloud.LoadBalancer.BatchUpdate` 写入；失败直接返回错误，不重试（design D4）
- [x] 4.2 把"生成更新项 / 需告警记录"抽为纯函数 `buildTCloudLBBizUpdates`，新增单测覆盖：全部未分配、部分已是其它业务、全部已是请求业务、记录为空
- [x] 4.3 data-service：`LoadBalancerExtUpdateReq` 新增可选字段 `BackupZones`，`batchUpdateLoadBalancer` 写入 `backup_zones`；未传时不更新
- [x] 4.4 `res-sync/tcloud/load_balancer.go` 新增 `getTCloudBackupZones`；`convCloudToDBCreate`、`convCloudToDBUpdate` 写入备可用区；`isLBChange` 在云上备可用区非空且与本地不同时判为有变化（design D7）
- [x] 4.5 新增 `res-sync/tcloud/load_balancer_test.go`，覆盖：云上返回 null、跳过空值、保持云上顺序；新建与更新写入备可用区、云上未返回时更新项为空；本地缺失或不同时判为有变化

## 5. 创建入口改造

- [x] 5.1 `BatchCreateTCloudClb` 改为：等待 → `lbSync`（全部 `SuccessCloudIDs`）→ 补业务；补业务失败返回含全部云 ID 的错误（design D5）
- [x] 5.2 等待返回非空时，在同步和补业务完成后返回错误，错误信息包含仍查不到的云 ID、全部成功云 ID，并写明"实例已在云上创建"
- [x] 5.3 删除 `createTCloudDBLoadBalancer` 及其不再使用的 import

## 6. 验证

- [x] 6.1 `go build ./...` 与 `go vet` 通过，改动文件 `gofmt`/`goimports` 无差异，行宽不超过 120 列
- [x] 6.2 `go test ./pkg/adaptor/tcloud/... ./cmd/hc-service/service/load-balancer/... ./cmd/hc-service/logics/res-sync/tcloud/...` 通过
- [x] 6.3 代码核对：`res-sync/tcloud` 的 `createLoadBalancer` 新建记录时写入了名称、可用区、备可用区、标签、VIP 等字段，与原预写字段相比没有缺失；有缺失则在本文件记录并评估
- [x] 6.4 代码核对：`git diff` 中 `cmd/hc-service/logics/res-sync/` 只有备可用区相关改动（D7），删除核对逻辑未改；`lbSync` 其它调用方未改动
- [ ] 6.5 联调环境用公有云账号、指定业务购买 CLB（单个与数量 3 各至少 1 次）：交付完成后立即查业务视角列表可见、业务正确，记录交付耗时对比修复前；公网主备、公网单可用区、内网各取一个实例，核对 `DescribeLoadBalancers` 返回的 `BackupZoneSet` 与本地 `backup_zones` 一致

## 实现记录

- 4.1：补业务逻辑放在新文件 `tcloud_lb_fill.go`，因 `tcloud.go` 已超过 1100 行；单测在同目录 `tcloud_lb_fill_test.go`。
- 4.3：`RearrangeSQLDataWithOption` 跳过 nil 切片，云上未返回备可用区时同步传 nil，不会清空已有值。
- 4.4：最初方案是创建入口按请求补备可用区；评审发现同步漏写了云上已返回的 `BackupZoneSet`，改为在同步中写入，删除按请求补的逻辑。
- 5.1：`lbSync` 失败的错误也包装上全部云 ID 和"实例已在云上创建"。交付失败时申请单只保存错误文本，不带云 ID 就无从追查。
- 6.1：`go vet ./pkg/adaptor/tcloud/` 的 3 条 unkeyed fields 告警在 `cvm.go`、`security_group.go`，与本变更无关。

## 3.2 排查结论

`DeleteLoadBalancer`（clb.go:488）、`CreateLoadBalancerSnatIps`（clb.go:732）、`DeleteLoadBalancerSnatIps`（clb.go:778）使用 `taskStatusDefaultPollingHandler`。它在任务运行中同样返回 `(false, nil)`，轮询 5 分钟超时后 `PollUntilDone` 返回空结果，随后读取 `result.SuccessCloudIDs` 会空指针。`listener.go` 中使用同一 handler 的 9 处调用存在同样问题。本变更不修改，建议另起变更统一修复（例如让 `taskStatusDefaultPollingHandler.Done` 始终返回非空结果）。

## 6.3 核对结论

`res-sync/tcloud` 的 `convCloudToDBCreate`（load_balancer.go:520）对比原预写字段：

| 字段 | 原预写 | 同步新建 | 结论 |
| --- | --- | --- | --- |
| 名称、标签、IP 版本、类型 | 请求值 | 云上值 | 一致或更准 |
| VIP、VPC、子网、状态等 | 无 | 云上值 | 更完整 |
| 可用区（公网） | 请求值 | 云上 `MasterZone` | 一致 |
| 可用区（内网） | 请求值 | 不写 | 云上内网 CLB 无主可用区，影响小 |
| 备可用区 | 请求值 | 云上 `BackupZoneSet`（新建与更新都写，4.4） | 一致，且覆盖控制台创建、导入的 CLB |

业务视角 CLB 详情页（`clb-detail/index.tsx:144`）展示 `backup_zones`。决策：同步按云上数据写备可用区（4.3、4.4），云上未返回时不覆盖本地已有值。
