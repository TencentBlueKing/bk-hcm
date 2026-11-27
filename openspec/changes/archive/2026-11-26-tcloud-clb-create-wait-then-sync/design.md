## Context

公有云购买 CLB 的现有链路（hc-service）：

1. `BatchCreateTCloudClb`（`cmd/hc-service/service/load-balancer/tcloud.go:113`）调用适配器 `CreateLoadBalancer`。
2. `CreateLoadBalancer`（`pkg/adaptor/tcloud/clb.go:164`）用 `poller.Poller` + `createClbPollingHandler` 轮询 `DescribeTaskStatus`，拿到 `SuccessCloudIDs`。
3. `createTCloudDBLoadBalancer` 用请求参数预写本地记录（含业务 `req.BkBizID`），失败也继续。
4. `lbSync`（`tcloud.go:335`）调用 `res-sync/tcloud` 的 `LoadBalancer()`：`common.Diff` 后，云上不返回、本地存在的记录进入删除清单；删除前再向云上确认一次，仍查不到就删除。

问题：

- **查询空窗**：`DescribeTaskStatus` 成功到 `DescribeLoadBalancers` 能查到，存在秒级延迟。第 4 步落在空窗里时，第 3 步预写的记录被删掉。
- **轮询超时空指针**：`createClbPollingHandler.Done` 在任务运行中返回 `(false, nil)`；`PollUntilDone` 超时后会再执行一次 Poll+Done，并把这个 `nil` 当结果返回（`pkg/adaptor/poller/poller.go:116-128`）。`CreateLoadBalancer` 第 205 行直接读 `result.SuccessCloudIDs`，空指针。

已核实的约束：

- 公有云同步新建记录时业务固定写"未分配"（`res-sync/tcloud/load_balancer.go:530`），之后的同步不再修改业务。公有云的业务只能来自购买请求。
- 交付成功后，cloud-server 只把云 ID 写入申请单（`cmd/cloud-server/service/application/handlers/load_balancer/tcloud/deliver.go:30-34`），不立即读本地记录。
- data-service 已有按记录批量更新 CLB 的接口 `TCloud.LoadBalancer.BatchUpdate`，空值字段不更新（`RearrangeSQLDataWithOption` 跳过空值）。
- 同步新建、更新记录时都不写备可用区（`convCloudToDBCreate` 只写公网 CLB 的主可用区，`isLBChange` 也不比较），原来备可用区只来自预写记录。但云上 `DescribeLoadBalancers` 返回了备可用区 `BackupZoneSet`（SDK 注释：可能返回 null），只是同步没有使用。
- `retry.RetryPolicy` 是"线性后随机"策略，首次 `Sleep()` 为 0。
- `TCloudListOption.CloudIDs` 最多 20 个（`constant.TCLBDescribeMax`）。
- 自研云适配器 `ZiyanAdpt` 内嵌 `tcloud.TCloud`，可直接复用等待函数。

## Goals / Non-Goals

**Goals:**

- 本地记录只在云上可见后由同步新建，从根上消除误删。
- 公有云与自研云的创建链路顺序一致：等待 → 同步 → 业务归属。
- 公有云的业务在同步后按请求补齐；备可用区由同步按云上数据写入。
- 异常（等待超时、补业务失败）显式暴露为交付失败，不让 CLB 静默落入"未分配"。
- 修复轮询超时空指针。

**Non-Goals:**

- 公有云的异步重试（本期不做）。
- 内部版合入与自研云改造（自研云子需求）。
- 修改 `res-sync/tcloud` 的删除核对逻辑；同步除备可用区外的其它字段行为。
- 主可用区 `zones` 的更新比较（同步更新路径仍不比较主可用区）。
- 其它云厂商、其它 CLB 子资源。
- 修改 `retry.RetryPolicy`、`poller` 本身的行为。

## Decisions

### D1：先等待、再同步，不再预写本地记录

被否决的方案是"保留预写记录，同步时传保留清单 `KeepCloudIDs` 不删本次记录"。它能解决误删，并且业务和记录一次写入；但公有云和自研云的顺序会不一致（自研云业务以云上标签为准，预写记录没有意义），还要改 `res-sync` 的删除逻辑。采用"先等待、再同步"后：

- 本地没有预写记录，同步的删除清单里不会出现本次云 ID，`res-sync` 的删除逻辑不用改。
- 记录的名称、可用区、VIP 等字段直接来自云上数据，不再用请求参数预填（预填值之后本来也会被同步按云上数据覆盖）。
- 代价：业务需要同步后再写一次，见 D4；备可用区原来只来自预写，改由同步写入，见 D7；等待超时时没有本地记录，见 D5。

`createTCloudDBLoadBalancer` 删除，不再调用。

### D2：等待函数放在 `pkg/adaptor/tcloud`，复用 `poller.Poller`

新增 `pkg/adaptor/tcloud/clb_wait.go`：

```go
// LoadBalancerLister 只包含查询 CLB 列表能力的适配器
type LoadBalancerLister interface {
	ListLoadBalancer(kt *kit.Kit, opt *typelb.TCloudListOption) ([]typelb.TCloudClb, error)
}

// WaitLoadBalancerVisible 等待 CLB 在云上可见，返回到达上限时仍查不到的云 ID，不返回错误
func WaitLoadBalancerVisible(kt *kit.Kit, lister LoadBalancerLister, region string, cloudIDs []string,
	opt *poller.PollUntilDoneOption) []string
```

- `opt` 为空时使用 `types.NewCLBVisibleWaitPollerOption()`；单测传 1 秒上限，避免跑满 30 秒。
- `clbVisibleWaitHandler` 实现 `poller.PollingHandler[LoadBalancerLister, map[string]bool, poller.BaseDoneResult]`：
  - `Poll`：按 20 个一批调用 `ListLoadBalancer`（`Page{Offset:0, Limit:20}`），返回"目标云 ID → 是否可见"。
  - `Done`：全部可见返回 `(true, result)`；否则 `(false, result)`，**result 始终非空**，未可见的放入 `UnknownCloudIDs`，避免超时分支返回空结果。
- `PollUntilDone` 报错（只在超时时最后一轮查询报错时出现）时，记 Warn，返回全部 `cloudIDs`。
- 有未可见 ID 时记 Warn（含云 ID、rid）。

不自己写循环，是为了复用 `poller` 已有的超时、重试计数和日志。用小接口，是因为单测只需 mock 一个方法，且 `ZiyanAdpt` 天然满足。

### D3：等待参数用常量

在 `pkg/criteria/constant/clb.go` 新增：

| 常量 | 值 | 含义 |
| --- | --- | --- |
| `TCLBVisibleWaitTimeoutSec` | 30 | 等待总上限（秒） |
| `TCLBVisibleWaitImmuneCount` | 4 | 线性递增的重试次数 |
| `TCLBVisibleWaitMinIntervalMS` | 500 | 线性步长 / 随机下限（毫秒） |
| `TCLBVisibleWaitMaxIntervalMS` | 2000 | 随机上限（毫秒） |

按 `RetryPolicy` 计算，等待间隔依次约为 0、500ms、1s、1.5s、2s，之后 1–2.5 秒随机；30 秒内约 18 轮，单次最长约 30 秒 + 2.5 秒 + 一次查询耗时。

### D4：同步后补业务

新增 `cmd/hc-service/service/load-balancer/tcloud_lb_fill.go`（`tcloud.go` 已超过 1100 行），`func (svc *clbSvc) fillTCloudLBBiz(kt *kit.Kit, req *protolb.TCloudLoadBalancerCreateReq, cloudIDs []string) error`：

1. 业务 ID 不大于 0 时直接返回。
2. 用 `Global.LoadBalancer.ListLoadBalancer` 按 `vendor=tcloud`、`account_id`、`region`、`cloud_id in cloudIDs` 查询，只取 `id`、`cloud_id`、`bk_biz_id`。
3. 按记录生成更新项（纯函数 `buildTCloudLBBizUpdates`）：只把 `constant.UnassignedBiz` 改为请求业务；已是请求业务的不动；其它值记 Warn，不覆盖。
4. 有更新项时调用 `TCloud.LoadBalancer.BatchUpdate`。
5. 第 2–4 步任一步失败直接返回错误，不重试，由调用方包装上云 ID，交付失败。

只改"未分配"的记录，是为了覆盖"周期同步抢先建了未分配记录"的情况，同时不覆盖已被人工或其它流程分配的业务。

替代方案是给 `SyncLBOption` 加字段，让同步新建记录时直接写请求业务，可以做到记录和业务一次写入。但它覆盖不了周期同步抢先建记录的情况，还要改 `res-sync`，不采用。

### D5：等待超时时交付失败

创建入口流程：

```go
missing := tcloud.WaitLoadBalancerVisible(kt, tcloudAdpt, req.Region, result.SuccessCloudIDs, nil)
if err := svc.lbSync(kt, tcloudAdpt, req.AccountID, req.Region, result.SuccessCloudIDs); err != nil {
	return nil, fmt.Errorf("lb instances have been created on cloud, but sync failed, cloud_ids: %v, err: %v", ...)
}
if err := svc.fillTCloudLBBiz(kt, req, result.SuccessCloudIDs); err != nil {
	return nil, fmt.Errorf("lb instances have been created on cloud, but fill biz failed, ...")
}
if len(missing) > 0 {
	return nil, fmt.Errorf("lb not visible on cloud after waiting, missing: %v, success: %v", missing,
		result.SuccessCloudIDs)
}
```

- 同步仍用全部 `SuccessCloudIDs`：本地没有这些记录，云上查不到的不会被删除，只是不新建。
- 已可见的部分先完成同步和补业务，再报错，避免它们也落入"未分配"。
- 查不到的云 ID 之后由周期同步拉回，业务为"未分配"，需人工分配；报错让申请单显示交付失败，错误中有云 ID 可追查。
- 公有云本期不做异步重试。

`lbSync` 签名不变（不再需要传核对选项）。更新 CLB、SNAT IP 等其它调用方不受影响。

### D6：轮询结果判空

`CreateLoadBalancer` 第 205 行改为 `if result == nil || len(result.SuccessCloudIDs) == 0`。结果为空时错误信息说明"轮询超时任务未结束"并带 RequestId。云上实例可能稍后创建成功，由周期同步拉回。

`DeleteLoadBalancer`、`CreateLoadBalancerSnatIps`、`DeleteLoadBalancerSnatIps` 也直接读 `result.SuccessCloudIDs`，不在本次范围内；实现时排查结论记录到 tasks，不顺带修改。

### D7：同步按云上数据写备可用区

`res-sync/tcloud/load_balancer.go` 新增 `getTCloudBackupZones(cloud)`：取 `BackupZoneSet` 中非空的 `Zone`，保持云上顺序；云上返回 null 或没有有效值时返回 nil。

- 新建：`convCloudToDBCreate` 写入 `BackupZones`。
- 更新：`convCloudToDBUpdate` 写入 `BackupZones`；`isLBChange` 在云上备可用区非空且与本地不同时判为有变化。
- 云上未返回时只补不清：更新项为空，data-service 不修改本地已有值。避免云上偶发返回 null 时误清。

data-service 配套改动：`LoadBalancerExtUpdateReq` 新增 `BackupZones []string`（`omitempty`），`batchUpdateLoadBalancer` 写入表字段 `backup_zones`。`RearrangeSQLDataWithOption` 跳过 nil 切片，未传时不更新。

放在同步而不是创建入口按请求补，是因为同步覆盖所有来源（平台购买、控制台创建、导入），且不用在入口维护一份与 `SlaveZoneID` 一致的推导条件。更新路径也纳入比较，是因为刚创建、状态仍为创建中的实例可能暂未返回备可用区，后续同步可补齐。

## Risks / Trade-offs

- **交付失败的情况变多** → 等待超时、补业务失败两种情况，从"静默成功但可能找不到"变为"交付失败带云 ID"。CLB 实际已创建并计费，申请人可能误以为没买到而重复购买；错误信息需明确写"实例已创建"。
- **等待超时后业务需人工分配** → 公有云不做异步重试，查不到的 CLB 由周期同步拉回为"未分配"。正常空窗为秒级，30 秒上限下预计极少发生；若线上出现，二期再评估异步重试。
- **记录字段来源变化** → 名称、可用区、备可用区、标签改为云上值（tasks 6.3 核对）；内网 CLB 原预写的请求可用区不再写入，云上内网 CLB 本身没有主可用区，影响小。
- **存量备可用区被批量补齐** → 周期同步会把云上有、本地缺失或不一致的备可用区写入，涉及一次面向存量的更新。只补不清，不会清空已有值；`BackupZoneSet` 在各类实例上的实际返回需在集成环境核对（tasks 6.5）。
- **交付变慢** → 正常多 1 轮查询；最坏多约 33 秒。免审直接交付在请求内同步执行，前端或网关超时待评审确认（父需求 Q-010）。
- **云 API 调用量增加** → 每轮每 20 个 ID 一次查询，最多约 18 轮。
- **公有云空窗未复现** → 用 mock 构造验证；正常路径只多一次查询。

## Migration Plan

无数据迁移脚本、无配置变更。部署 hc-service 即生效，存量备可用区由周期同步补齐；回滚直接回退版本，已补齐的备可用区保留，不影响旧版本行为。

## Open Questions

- Q-010（父需求，非阻塞）：免审直接交付路径下，前端或网关请求超时能否容纳最多约 33 秒的额外等待。
