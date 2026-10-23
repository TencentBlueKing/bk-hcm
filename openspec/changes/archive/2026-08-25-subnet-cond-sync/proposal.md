## Why

条件同步（Conditional Sync）允许针对单个云账号，按 `regions` / `cloud_ids` / `tag_filters` 等条件增量同步指定资源，避免每次都触发整账号全量同步，从而提升同步效率与可控性。

当前条件同步已支持 region、zone、安全组、负载均衡、子账号、权限模板等资源类型，但**子网（subnet）尚未纳入条件同步支持范围**。子网是网络规划的基础资源，业务上经常需要在调整 VPC / 网络后仅刷新子网数据，而非等待整账号全量同步。因此需要将 `subnet` 资源类型补充到条件同步体系中，覆盖 aws、azure、huawei、tcloud 四朵云，并对下游 `hc-service` 的子网同步能力做配套增强。

## What Changes

- **cloud-server 条件同步注册表**：在 aws/azure/huawei/tcloud 各自的 `condSyncFuncMap` 中新增 `enumor.SubnetCloudResType -> CondSyncSubnet`，并实现各厂商的 `CondSyncSubnet` 函数。
  - aws / tcloud：`CondSyncSubnet` 按 `regions` 遍历，结合 `cloud_ids` / `tag_filters` 选项调用 `hc-service` 的 `SyncSubnet` 接口。
  - azure / huawei：子网归属于 VPC，无法脱离 VPC 直接同步。`CondSyncSubnet` 先按条件（资源组 / 区域 或 `cloud_ids`）列出涉及的 VPC 唯一键（vpc key），再以 VPC 为粒度并发（pipeline）调用 `SyncSubnet` 完成子网同步。
- **hc-service 子网同步器**：为 aws、tcloud 的子网 `SyncSubnet` 的 `Next()` 增加 `CloudIDs` 分支，支持仅同步指定 `cloud_ids` 的子网（需配合单个 region）。
- **account API 入参校验**：`ResCondSyncReq.Validate` 新增 `needRegion` 参数，当资源类型需要 region（如 subnet）但未传 `regions` 时校验失败；并补充单测。
- **新增单测**：各厂商条件同步注册表对 `SubnetCloudResType` 的注册校验，以及 `ResCondSyncReq.Validate` 的用例。

## Impact

- 新增能力：四类云厂商（aws/azure/huawei/tcloud）支持 subnet 的条件同步。
- 受影响模块：`cmd/cloud-server/service/sync/{aws,azure,huawei,tcloud}`、`cmd/hc-service/service/sync/{aws,tcloud}`、`pkg/api/cloud-server/account`、`pkg/api/hc-service/sync` 及其单测。
- 向后兼容：均为增量新增逻辑，不改变既有资源类型的同步行为；`Validate` 新增参数仅影响新调用点。
- 行为约束：azure/huawei 子网条件同步必须以 VPC（resource_group+vpc / region+vpc）为粒度，调用方需保证 `regions` / `resource_group_names` 与 `cloud_ids` 的约束一致（例如指定 `cloud_ids` 时 region / resource_group 必须为单个）。
