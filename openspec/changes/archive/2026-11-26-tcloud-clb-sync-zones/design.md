# Design

## Context

动机见 `proposal.md - Why`，行为契约见 `specs/tcloud-clb-sync-zones/spec.md`。

当前腾讯云 CLB 资源同步（`cmd/hc-service/logics/res-sync/tcloud/load_balancer.go`）与数据服务更新链路（`cmd/data-service/service/cloud/load-balancer/update.go`）现状：

1. **云端 SDK 字段形态（`tencentcloud-sdk-go/tencentcloud/clb/v20180317.LoadBalancer`）**：
   - `MasterZone *ZoneInfo`（含 `Zone *string`）：公网 CLB 主可用区（包括单可用区、主备可用区，以及创建时未指定 `ZoneId` 由云端随机分配的主可用区）；可能返回 `nil` 或空字符串 `""`。
  - `Zones []*string`：私有网络内网 CLB 就近接入模式下规则所落在的可用区列表；可能返回 `nil`。
   - `BackupZoneSet []*ZoneInfo`：备可用区列表，已通过 `getTCloudBackupZones` 接入新建、更新转换与变更比对。
2. **现有同步代码缺口**：
   - `convCloudToDBCreate`（第 579–581 行）仅写：
     ```go
     if typeslb.TCloudLoadBalancerType(lb.LoadBalancerType) == typeslb.OpenLoadBalancerType && cloud.MasterZone != nil {
         lb.Zones = []string{cvt.PtrToVal(cloud.MasterZone.Zone)}
     }
     ```
     当 `cloud.MasterZone.Zone` 为 `nil` 或 `""` 时，会产出 `[""]`；且内网 CLB 未读取 `cloud.Zones`。
   - `convCloudToDBUpdate`（第 666–704 行）与 `isLBChange`（第 706–775 行）只处理了 `BackupZones`，完全没有处理 `Zones`。
   - `pkg/api/data-service/cloud/load_balancer.go` 的 `LoadBalancerExtUpdateReq` 与 `cmd/data-service/service/cloud/load-balancer/update.go` 的 `batchUpdateLoadBalancer` 仅包含 `BackupZones`，不包含 `Zones`。

## Goals / Non-Goals

**Goals:**
- 统一公网（`OPEN`）与内网（`INTERNAL`）CLB 的云端可用区提取逻辑，并严格过滤空指针与空字符串。
- 在 CLB 同步的新建、变更比对与更新链路中完整支持 `Zones` 字段回填与修正。
- 保持更新链路“只补/只改不清”语义，与 `BackupZones` 行为一致，不影响其他局部更新调用方。

**Non-Goals:**
- 不通过关联子网（`Subnet`）反查回填内网 CLB 可用区（仅以云端 CLB 实例自身返回的字段为准）。
- 不修改 CLB 创建入口参数透传逻辑（创建时随机可用区即不传 `ZoneId`，由同步从云端回填）。
- 本变更仅覆盖开源版 `bk-hcm` 公有云（`tcloud`）；内部版自研云（`tcloud_ziyan`）在后续合入内部版时按相同模式适配。

## Decisions

### D1：新增 `getTCloudZones(cloud typeslb.TCloudClb) []string` 统一提取并过滤空值

在 `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go` 中新增辅助函数 `getTCloudZones`：

```go
// getTCloudZones 返回云上实例的主可用区或可用区列表，云上未返回有效值时为 nil
func getTCloudZones(cloud typeslb.TCloudClb) []string {
	switch typeslb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
	case typeslb.OpenLoadBalancerType:
		if cloud.MasterZone == nil || cvt.PtrToVal(cloud.MasterZone.Zone) == "" {
			return nil
		}
		return []string{cvt.PtrToVal(cloud.MasterZone.Zone)}
	case typeslb.InternalLoadBalancerType:
		var zones []string
		for _, zone := range cloud.Zones {
			if cvt.PtrToVal(zone) == "" {
				continue
			}
			zones = append(zones, cvt.PtrToVal(zone))
		}
		return zones
	default:
		return nil
	}
}
```

- **理由**：将公网 `MasterZone.Zone` 与内网 `cloud.Zones` 收口到同一纯函数中，既消除了 `MasterZone.Zone` 为空字符串时生成 `[""]` 的缺陷，又让 `convCloudToDBCreate`、`convCloudToDBUpdate` 与 `isLBChange` 三处复用完全一致的取值口径，且后续合入内部版自研云（`ziyan`）时可直接对齐。
- **备选方案对比**：
  - *仅处理公网 `MasterZone.Zone`*：无法覆盖内网就近接入模式下云端返回的 `cloud.Zones`，且与内部版 `ziyan` 分叉。
  - *内网为空时查子网可用区兜底*：偏离云端 CLB 实例原始字段语义，且子网变更或跨可用区场景下易产生歧义，不采用。

### D2：同步更新采用“只补/只改不清”策略（与 `BackupZones` 对齐）

- **新建（`convCloudToDBCreate`）**：`lb.Zones = getTCloudZones(cloud)`。
- **变更比对（`isLBChange`）**：
  ```go
  if zones := getTCloudZones(cloud); len(zones) > 0 && !slices.Equal(db.Zones, zones) {
      return true
  }
  ```
- **更新转换（`convCloudToDBUpdate`）**：设置 `Zones: getTCloudZones(cloud)`。
- **理由**：当云端返回非空有效可用区且与本地 `db.Zones` 不同（例如本地为空、`[""]` 或可用区发生迁移）时触发更新；当云端未返回有效可用区（`getTCloudZones(cloud)` 为 `nil`）时，`isLBChange` 不判为变更，`convCloudToDBUpdate` 中 `Zones` 为 `nil`，避免因云端偶发未返回字段而清空本地已有可用区数据。
- **备选方案对比**：
  - *云端为空时强制清空本地 `db.Zones`*：由于底层 DAO `RearrangeSQLDataWithOption` 默认跳过 `nil` / 零值字段，若要清空需额外引入特殊清空机制，且在云端偶发返回 `null` 时有丢失已落库可用区的风险，不采用。

### D3：`data-service` 批量更新接口新增可选 `Zones` 字段

- `pkg/api/data-service/cloud/load_balancer.go` 的 `LoadBalancerExtUpdateReq[T]` 新增：
  ```go
  // Zones 可用区，为空时不更新
  Zones []string `json:"zones,omitempty"`
  ```
- `cmd/data-service/service/cloud/load-balancer/update.go` 的 `batchUpdateLoadBalancer` 在构造 `tablelb.LoadBalancerTable` 时写入 `Zones: lb.Zones`。
- **理由**：`tablelb.LoadBalancerTable` 已存在 `Zones tabletype.StringSlice` 列（`db:"zones"`），DAO 层 `UpdateByIDWithTx` 通过 `utils.RearrangeSQLDataWithOption` 生成 SQL 时会自动跳过 `nil` 切片。因此仅更新 `BkBizID`（`fillTCloudLBBiz`）或 `SyncTime`（`updateLoadBalancerSyncTime`）的调用方无需改动，天然保持向后兼容。

## Risks / Trade-offs

- **[存量缺失可用区的实例将在首次同步时批量更新]** → 周期同步首次扫描到历史 `zones` 为空或 `[""]` 的公网 CLB 时，会触发一次 DB 更新回填真实可用区；单批次受限于同步分页（20 条/批），数据库写入压力可控。
- **[云端创建初期 `MasterZone.Zone` 暂未就绪]** → 若购买完成瞬间首次同步时云端 `MasterZone.Zone` 仍为空，首次新建记录 `zones` 为空（而非 `[""]`），待后续同步（手动同步或周期同步）云端返回非空 `MasterZone.Zone` 后会自动检测差异并回填。

## Migration Plan

无需执行 SQL 数据迁移脚本或配置变更。升级部署 `data-service` 与 `hc-service` 后，新购买及存量 CLB 实例将在同步时自动回填 `zones` 字段；如需回滚，直接回退服务镜像即可，已回填的 `zones` 数据合法且不影响旧版本运行。
