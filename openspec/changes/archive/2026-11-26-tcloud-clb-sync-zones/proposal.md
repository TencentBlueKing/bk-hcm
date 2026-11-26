# Proposal

## Why

在混合云管理平台（HCM）购买腾讯云公网负载均衡（CLB）时支持“随机选择可用区”（即创建请求不传 `zones` 参数，由腾讯云在地域下随机分配主可用区并在 `DescribeLoadBalancers` 的 `MasterZone.Zone` 中返回），且购买链路已改为“等待云上可见后由资源同步（`res-sync`）新建本地记录”。然而现有同步逻辑（`cmd/hc-service/logics/res-sync/tcloud/load_balancer.go`）仅在新建时对公网 CLB 直接写入 `[]string{cvt.PtrToVal(cloud.MasterZone.Zone)}`（未过滤空字符串，也未读取内网 CLB 的 `cloud.Zones`），而在差异比对（`isLBChange`）、同步更新（`convCloudToDBUpdate`）以及 `data-service` 批量更新接口（`LoadBalancerExtUpdateReq` / `batchUpdateLoadBalancer`）中完全缺失了 `Zones` 字段，导致实例创建初期云端尚未返回可用区或存量缺失可用区的实例在后续同步中永远无法回填可用区，详情页与列表页一直显示为 `--`。

## What Changes

- **统一提取云端可用区**：在 `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go` 中新增辅助函数 `getTCloudZones(cloud typeslb.TCloudClb) []string`：
  - 公网 CLB（`OPEN`）从 `cloud.MasterZone.Zone` 提取主可用区；
  - 内网 CLB（`INTERNAL`）从 `cloud.Zones` 提取可用区列表；
  - 统一过滤 `nil` 与空字符串 `""`，无有效可用区时返回 `nil`（防止写入 `[""]` 脏数据）。
- **同步新建、差异比对与更新回填可用区**：
  - `convCloudToDBCreate` 改用 `getTCloudZones(cloud)` 设置 `lb.Zones`；
  - `isLBChange` 在云端返回有效非空可用区（`len(zones) > 0`）且与本地 `db.Zones` 不一致时判定为实例发生变更；
  - `convCloudToDBUpdate` 写入 `Zones: getTCloudZones(cloud)`；当云端未返回有效可用区时为 `nil`，不覆盖或清空本地已有值（与备可用区 `BackupZones` 的“只补/只改不清”策略保持一致）。
- **`data-service` 批量更新接口支持 `Zones` 字段**：
  - `pkg/api/data-service/cloud/load_balancer.go` 的 `LoadBalancerExtUpdateReq[T]` 新增可选字段 `Zones []string`（`json:"zones,omitempty"`）；
  - `cmd/data-service/service/cloud/load-balancer/update.go` 的 `batchUpdateLoadBalancer` 将 `lb.Zones` 映射写入 `tablelb.LoadBalancerTable.Zones`（底层 `RearrangeSQLDataWithOption` 自动跳过 `nil` 切片，未传时不修改数据库原值）。
- **单元测试**：在 `cmd/hc-service/logics/res-sync/tcloud/load_balancer_test.go` 补充覆盖 `getTCloudZones`、`convCloudToDBCreate`、`convCloudToDBUpdate` 与 `isLBChange` 的可用区单元测试。

## Capabilities

### New Capabilities
- `tcloud-clb-sync-zones`: 腾讯云负载均衡（CLB）资源同步时按云上数据提取、新建、比对并回填更新可用区（`zones`）字段的能力。

### Modified Capabilities

（无。现有 `openspec/specs/` 中无覆盖 CLB 可用区同步的规范。）

## Impact

- **代码**：
  - `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go`：新增 `getTCloudZones`，并在新建、更新转换与变更比对中接入 `Zones`。
  - `cmd/hc-service/logics/res-sync/tcloud/load_balancer_test.go`：新增可用区提取、转换与变更检测单测。
  - `pkg/api/data-service/cloud/load_balancer.go`：`LoadBalancerExtUpdateReq` 新增可选字段 `Zones`。
  - `cmd/data-service/service/cloud/load-balancer/update.go`：`batchUpdateLoadBalancer` 写入 `Zones`。
- **数据与兼容性**：
  - 复用 `load_balancer` 表既有 `zones`（JSON 数组）列，无数据库表结构变更（DDL）。
  - `LoadBalancerExtUpdateReq.Zones` 为 `omitempty` 可选字段，`nil` 时跳过更新，完全向后兼容既有局部更新调用方（如 `fillTCloudLBBiz`、`updateLoadBalancerSyncTime`）。
  - 存量缺失可用区或含 `[""]` 的 CLB 记录将在下一轮周期同步或手动同步后自动回填修正。
- **下游**：开源版 `bk-hcm` 完成后，后续合入内部版时自研云（`tcloud_ziyan`）同步逻辑一并复用相同规则适配。
