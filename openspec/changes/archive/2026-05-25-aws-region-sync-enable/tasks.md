## 1. 数据库变更

- [x] 1.1 新建 SQL 迁移脚本 `scripts/sql/0047_20260525_1000_aws_region_sync_enable.sql`
  - 添加 `sync_enable` 字段（`tinyint(1) NOT NULL DEFAULT 1`）
  - 添加索引 `idx_sync_enable`

## 2. DAL 层改造

- [x] 2.1 `pkg/dal/table/cloud/region/aws.go`：`AwsRegionColumnDescriptor` 新增 `sync_enable` 列描述
- [x] 2.2 `pkg/dal/table/cloud/region/aws.go`：`AwsRegionTable` 结构体新增 `SyncEnable bool` 字段

## 3. Core Model 层改造

- [x] 3.1 `pkg/api/core/cloud/region/aws.go`：`AwsRegion` 结构体新增 `SyncEnable bool` 字段

## 4. Data-Service 层 API 改造

- [x] 4.1 `pkg/api/data-service/cloud/region/aws.go`：`AwsRegionBatchUpdate` 结构体新增 `SyncEnable *bool` 字段
- [x] 4.2 `pkg/api/data-service/cloud/region/aws.go`：`AwsRegionBatchCreate` 结构体新增 `SyncEnable *bool` 字段（默认 true）

## 5. Data-Service 层服务改造

- [x] 5.1 `cmd/data-service/service/cloud/region/aws.go`：`BatchUpdateAwsRegion` 支持更新 `sync_enable` 字段
- [x] 5.2 `cmd/data-service/service/cloud/region/aws.go`：`BatchCreateAwsRegion` 设置 `sync_enable` 默认值为 `true`
- [x] 5.3 `cmd/data-service/service/cloud/region/aws.go`：`convertAwsBaseRegion` 函数添加 `SyncEnable` 字段映射

## 6. Cloud-Server 层 API 改造

- [x] 6.1 `pkg/api/cloud-server/region/request.go`：新增 `RegionBatchUpdateSyncEnableReq` 请求结构体
- [x] 6.2 `pkg/api/cloud-server/region/request.go`：`Validate()` 方法限制 IDs 数量不超过 100

## 7. Cloud-Server 层服务改造

- [x] 7.1 `cmd/cloud-server/service/region/region.go`：新增 `BatchUpdateRegionSyncEnable` Handler
- [x] 7.2 `cmd/cloud-server/service/region/region.go`：新增 `batchUpdateAwsRegionSyncEnable` 实现方法
- [x] 7.3 `cmd/cloud-server/service/region/region.go`：`InitRegionService` 注册新路由
  ```go
  h.Add("BatchUpdateRegionSyncEnable", http.MethodPatch, "/vendors/{vendor}/regions/sync_enable/batch",
      svc.BatchUpdateRegionSyncEnable)
  ```

## 8. Data-Service Client 改造

- [x] 8.1 `pkg/client/data-service/aws/region.go`：`BatchUpdate` 方法已支持 `sync_enable` 字段（无需额外改动）

## 9. 同步逻辑改造

- [x] 9.1 `cmd/cloud-server/service/sync/aws/region.go`：`ListRegion` 函数添加 `sync_enable = true` 过滤条件
- [x] 9.2 `cmd/cloud-server/service/sync/aws/region.go`：添加日志输出启用同步的地域列表

## 10. 单元测试

- [x] 10.1 `pkg/api/cloud-server/region/request_test.go`：新建单元测试文件
- [x] 10.2 `TestRegionBatchUpdateSyncEnableReq_Validate`：验证请求参数校验逻辑
- [x] 10.3 `cmd/cloud-server/service/sync/aws/region_test.go`：新建同步逻辑单元测试文件
- [x] 10.4 `TestListRegion_FilterConditionValidation`：验证过滤条件包含 `account_id` 和 `sync_enable` 字段
- [x] 10.5 `TestListRegion_FilterWithDifferentAccountIDs`：验证不同 account_id 生成不同的过滤条件
- [x] 10.6 `TestListRegion_SyncEnableFilterValue`：验证 sync_enable 过滤值为 true
- [x] 10.7 `TestListRegion_FilterExpressionType`：验证过滤表达式类型
- [x] 10.8 `TestListRegion_FilterOperator`：验证过滤条件的操作符
