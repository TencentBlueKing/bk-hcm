# Design: 子网条件同步（Subnet Conditional Sync）

## Overview

本提案复用现有 account 条件同步框架，将 `subnet` 纳入 `condSyncFuncMap` 支持的资源类型，并为四朵云分别实现 `CondSyncSubnet`。差异化处理集中在 azure / huawei 两类「子网依附于 VPC」的厂商：它们需要先定位到 VPC，再按 VPC 维度并发拉取子网，而不能像 aws / tcloud 那样直接按 region 拉取。

整体数据流：

```
account 条件同步 API (ResCondSyncReq / AzureResCondSyncReq)
        │  解析 regions / cloud_ids / tag_filters
        ▼
cloud-server CondSyncSubnet(vendor)
        │  构造 CondSyncParams，定位目标范围
        ▼
   ┌────┴─────┐  (azure/huawei) 先列出 VPC key
   │          │
   ▼          ▼
hc-service SyncSubnet   ──►  拉取指定子网写入本地 DB
```

## Architecture / Solution

条件同步框架已有统一抽象：

```go
// CondSyncParams 条件同步选项
type CondSyncParams struct {
    AccountID          string
    Regions            []string
    CloudIDs           []string
    TagFilters         core.MultiValueTagMap
    ResourceGroupNames []string
}

// CondSyncFunc sync resource by given condition
type CondSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error

func GetCondSyncFunc(res enumor.CloudResourceType) (syncFunc CondSyncFunc, ok bool)
```

`CondSyncSubnet` 通过 `cliSet.HCService().<Vendor>.Subnet.SyncSubnet(kt, option, nil)` 将同步请求下发给 `hc-service`。

## Detailed Design

### 1. aws / tcloud 实现

按 `Regions` 遍历，对每个 region 组装同步选项并发/串行调用 hc-service：

- aws：`&sync.AwsSyncSubnetOption{AccountID, Region, CloudIDs, TagFilters}` → `cliSet.HCService().Aws.Subnet.SyncSubnet`
- tcloud：`&sync.TCloudSyncSubnetOption{AccountID, Region, CloudIDs, TagFilters}` → `cliSet.HCService().TCloud.Subnet.SyncSubnet`

两者均直接按 region 维度拉取子网，逻辑与现有 `CondSyncSecurityGroup` 类似。

### 2. azure 实现（VPC 维度）

`CondSyncSubnet` 基于 `AzureResCondSyncReq`（含 `ResourceGroupNames` 与 `CloudIDs`）：

1. 若提供 `CloudIDs`：通过 `listAzureSubnetVpcKeysByCloudIDs` 在给定资源组内按 cloud_ids 列出子网，收集其 VPC 唯一键（`ResourceGroupName` + `VpcID`）。
2. 否则：通过 `listAzureSubnetVpcKeysByResourceGroupNames` 在给定资源组内列出全部 VPC，收集 VPC 唯一键。
3. 以 VPC 唯一键数量为并发度（`concurrency = len(vpcKeys)`）构造 pipeline，每个任务调用 `cliSet.HCService().Azure.Subnet.SyncSubnet(kt, &sync.AzureSyncSubnetOption{AccountID, ResourceGroupName, VpcID, CloudIDs}, nil)`。

### 3. huawei 实现（VPC 维度）

`CondSyncSubnet` 通过 `listHuaweiCondSyncSubnetVpcKeys` 定位 VPC 唯一键（`Region` + `VpcID`）：

1. 若提供 `CloudIDs`：跨 region 按 cloud_ids 列出子网，收集其 VPC 唯一键。
2. 否则：按 `Regions` 列出 VPC，收集 VPC 唯一键。
3. 同样以 VPC 唯一键数量为并发度构造 pipeline，每个任务调用 `cliSet.HCService().HuaWei.Subnet.SyncSubnet(kt, &sync.HuaWeiSyncSubnetOption{AccountID, Region, VpcID, CloudIDs}, nil)`。

### 4. hc-service 子网同步器 CloudIDs 支持

在 aws、tcloud 的 `SyncSubnet` 的 `Next()` 中新增 `CloudIDs` 分支：当 `s.cloudIDs != nil` 时，按 cloud_ids（配合单个 region）查询子网并返回结果，跳过原有的 region 全量分页逻辑；否则维持原分页行为。同时在 `pkg/api/hc-service/sync/sync.go` 注释中标注 `CloudIDs` 现已支持 `subnet`。

### 5. 入参校验

`ResCondSyncReq.Validate(needRegion bool)` 新增 `needRegion` 参数：

- `cloud_ids` 与 `tag_filters` 不可同时指定；
- 指定 `cloud_ids` 时 `regions` 必须为 1 个；
- 当 `needRegion == true`（subnet 场景）且 `regions` 为空时返回 `regions is required`。

azure 仍使用 `AzureResCondSyncReq.Validate`，约束 `resource_group_names` 必填、指定 `cloud_ids` 时资源组为 1 个。

## Data Flow

1. API 层解码请求 → 调用 `vendor.CondSyncFunc`（`GetCondSyncFunc(enumor.SubnetCloudResType)`）。
2. `CondSyncSubnet` 组装 `CondSyncParams`，对 aws/tcloud 按 region 直接同步，对 azure/huawei 先定位 VPC key。
3. 通过 `client.ClientSet` 调用 `hc-service` 的 `SyncSubnet` 接口，由 hc-service 拉取云上子网数据并落库。

## Risks & Mitigations

- **azure/huawei VPC 数量膨胀**：VPC 较多时会产生较多并发 `SyncSubnet` 调用。缓解：pipeline 并发度直接等于 `vpcKeys` 数量，且每个 VPC 内部的子网拉取由 hc-service 自带的分页/限流控制。
- **cloud_ids 与多 region 冲突**：入参层 `Validate` 已禁止「cloud_ids + regions>1」，从源头规避定位歧义。
- **region 缺失**：subnet 强依赖 region，`needRegion` 校验保证缺失时提前失败，避免下发无意义的同步请求。
