# Tasks

## 1. 注册子网条件同步能力（四朵云）

- [x] aws：`condSyncFuncMap` 新增 `SubnetCloudResType -> CondSyncSubnet`，实现 `CondSyncSubnet`（按 region 调用 `SyncSubnet`）
- [x] tcloud：同上，按 region 结合 `cloud_ids` / `tag_filters` 调用 `SyncSubnet`
- [x] azure：`condSyncFuncMap` 新增并基于 VPC key（resource_group + vpc）pipeline 实现 `CondSyncSubnet`
- [x] huawei：`condSyncFuncMap` 新增并基于 VPC key（region + vpc）pipeline 实现 `CondSyncSubnet`

## 2. 增强 hc-service 子网同步器

- [x] aws `SyncSubnet.Next()` 支持 `CloudIDs` 过滤分支
- [x] tcloud `SyncSubnet.Next()` 支持 `CloudIDs` 过滤分支
- [x] 更新 `pkg/api/hc-service/sync/sync.go` 注释，标注 `CloudIDs` 已支持 `subnet`

## 3. 账户条件同步入参校验

- [x] `ResCondSyncReq.Validate` 增加 `needRegion` 参数及 region 必填校验
- [x] 新增 `pkg/api/cloud-server/account/sync_test.go` 覆盖 `Validate` 各场景

## 4. 单测覆盖

- [x] 新增 `cmd/cloud-server/service/sync/cond_sync_subnet_test.go`，校验 aws/azure/huawei/tcloud 均注册 `SubnetCloudResType`
