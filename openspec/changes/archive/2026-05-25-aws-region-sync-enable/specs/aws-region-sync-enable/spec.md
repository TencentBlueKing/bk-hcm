## ADDED Requirements

### Requirement: aws_region 表新增 sync_enable 字段
系统 SHALL 在 `aws_region` 表中新增 `sync_enable` 字段，类型为 `TINYINT(1) NOT NULL`，默认值为 1（启用），用于控制该地域是否参与资源同步。迁移脚本为 `scripts/sql/9999_20260525_aws_region_sync_enable.sql`，不额外创建索引。

#### Scenario: 查询 aws_region 表结构
- **WHEN** 执行 `DESCRIBE aws_region`
- **THEN** 返回结果中包含 `sync_enable` 字段，类型为 `tinyint(1)`，默认值为 `1`

#### Scenario: 新增地域时默认启用同步
- **WHEN** 从云端同步到新地域并插入 `aws_region` 表
- **THEN** 新记录的 `sync_enable` 字段值为 `1`

---

### Requirement: Table/Core/DS 结构体扩展 SyncEnable 字段
`AwsRegionTable`（Table 层）SHALL 新增 `SyncEnable *bool` 字段，`AwsRegion`（Core 层）SHALL 新增 `SyncEnable bool` 字段。data-service 创建/更新请求中的 `SyncEnable` 为 `*bool`；创建时未传则写入 `true`。

#### Scenario: Table 结构体包含 SyncEnable 字段
- **WHEN** 查看 `pkg/dal/table/cloud/region/aws.go` 中的 `AwsRegionTable` 定义
- **THEN** 结构体包含 `SyncEnable *bool` 字段，db tag 为 `sync_enable`

#### Scenario: Core 结构体包含 SyncEnable 字段
- **WHEN** 查看 `pkg/api/core/cloud/region/aws.go` 中的 `AwsRegion` 定义
- **THEN** 结构体包含 `SyncEnable bool` 字段，json tag 为 `sync_enable`

---

### Requirement: 全量同步按 sync_enable 过滤
`ListRegion` SHALL 只按 `account_id` 查询 `aws_region`，再由 `parseSyncEnabledRegions` 在内存中分离启用和禁用地域。全量同步 SHALL 只处理 `sync_enable = true` 的地域，被禁用的地域不调用云厂商 API。

#### Scenario: 同步任务过滤禁用地域
- **GIVEN** `aws_region` 表中存在 region_id=`us-east-1`（sync_enable=true）和 region_id=`me-south-1`（sync_enable=false）
- **WHEN** 触发 AWS 全量资源同步，调用 `ListRegion`
- **THEN** 返回的待同步地域仅包含 `us-east-1`，不调用 `me-south-1` 相关云厂商 API

#### Scenario: 所有地域禁用时返回错误
- **GIVEN** `aws_region` 表中所有地域 `sync_enable = false`
- **WHEN** 调用 `ListRegion`
- **THEN** 返回 `all aws regions are disabled, disabled regions: [...]`，错误中包含被禁用的地域列表

#### Scenario: 地域表为空时返回错误
- **GIVEN** 该账号在 `aws_region` 中没有任何记录
- **WHEN** 调用 `ListRegion`
- **THEN** 返回 `aws region is empty`

---

### Requirement: 条件同步对已屏蔽地域报错
AWS 条件同步在显式传入 `regions` 时 SHALL NOT 按 `sync_enable` 过滤查询，以便区分地域不存在和已屏蔽。已屏蔽地域 SHALL 返回错误，不调用该地域的云厂商 API，也不静默跳过。未传 `regions` 时上游不使用该字段，SHALL 直接返回，不再查询地域。

#### Scenario: 显式同步已屏蔽地域
- **GIVEN** `me-south-1` 的 `sync_enable = false`
- **WHEN** 条件同步请求的 `regions` 包含 `me-south-1`
- **THEN** 返回 `some request regions sync is disabled`，且包含 `me-south-1`

#### Scenario: 显式同步不存在的地域
- **WHEN** 条件同步请求的 `regions` 包含库中不存在的 `region_id`
- **THEN** 返回 `some request regions don't exist`，且包含该地域

---

### Requirement: 同步任务输出启用与禁用地域日志
`ListRegion` 成功返回时 SHALL 记录启用地域数量和禁用地域列表。

#### Scenario: 日志记录禁用地域
- **GIVEN** `us-east-1`、`us-west-2` 启用同步，`me-south-1` 禁用同步
- **WHEN** 调用 `ListRegion`
- **THEN** 日志包含 `sync enabled region count` 和 `disabled regions`，禁用列表中包含 `me-south-1`

---

### Requirement: 地域同步状态管理接口
系统 SHALL 提供 `PATCH /api/v1/cloud/vendors/{vendor}/regions/sync_enable/batch` 接口，支持批量更新地域的 `sync_enable` 状态。请求体包含 `ids`（地域 ID 列表）和 `sync_enable`（目标状态）。请求结构体 `RegionBatchUpdateSyncEnableReq` 定义在 `pkg/api/cloud-server/region/request.go` 中，`Validate()` 方法 SHALL 限制 IDs 数量不超过 100。

#### Scenario: 批量禁用地域同步
- **GIVEN** 请求体为 `{"ids": ["xxx-id-1", "xxx-id-2"], "sync_enable": false}`
- **WHEN** 调用 `PATCH /api/v1/cloud/vendors/aws/regions/sync_enable/batch`
- **THEN** `aws_region` 表中对应 ID 的记录 `sync_enable` 更新为 `0`

#### Scenario: 批量启用地域同步
- **GIVEN** 请求体为 `{"ids": ["xxx-id-1"], "sync_enable": true}`
- **WHEN** 调用 `PATCH /api/v1/cloud/vendors/aws/regions/sync_enable/batch`
- **THEN** `aws_region` 表中对应 ID 的记录 `sync_enable` 更新为 `1`

#### Scenario: 空 IDs 请求返回错误
- **GIVEN** 请求体为 `{"ids": [], "sync_enable": false}`
- **WHEN** 调用 `Validate()`
- **THEN** 返回错误，错误信息包含 `ids`

#### Scenario: IDs 超过 100 个返回错误
- **GIVEN** 请求体中 `ids` 数组包含 101 个元素
- **WHEN** 调用 `Validate()`
- **THEN** 返回错误 `ids count should <= 100`

---

### Requirement: 存量数据迁移
系统 SHALL 提供数据迁移 SQL，为 `aws_region` 表所有存量记录将 `sync_enable` 设置为 `1`（通过字段默认值实现）。迁移完成后，不存在 `sync_enable IS NULL` 的记录。

#### Scenario: 迁移后存量记录 sync_enable 有值
- **WHEN** 执行迁移脚本后查询 `aws_region` 表
- **THEN** 所有记录的 `sync_enable` 字段均为 `1`，不存在 NULL 值

---

### Requirement: 单元测试覆盖
系统 SHALL 提供单元测试覆盖请求参数校验和同步过滤逻辑。所有单元测试 SHALL 通过 `go test ./...` 执行成功。

#### Scenario: 请求参数校验测试
- **WHEN** 执行 `go test ./pkg/api/cloud-server/region/...`
- **THEN** `TestRegionBatchUpdateSyncEnableReq_Validate_EmptyIDs`、`TestRegionBatchUpdateSyncEnableReq_Validate_NilIDs`、`TestRegionBatchUpdateSyncEnableReq_Validate_TooManyIDs`、`TestRegionBatchUpdateSyncEnableReq_Validate_NilSyncEnable`、`TestRegionBatchUpdateSyncEnableReq_Validate_ValidRequest`、`TestRegionBatchUpdateSyncEnableReq_Validate_MaxIDs`、`TestRegionBatchUpdateSyncEnableReq_Validate_SingleID` 通过

#### Scenario: 同步过滤逻辑测试
- **WHEN** 执行 `go test ./cmd/cloud-server/service/sync/aws/ -run TestListRegion_`
- **THEN** 以下测试用例全部通过：
  - `TestListRegion_OnlyEnabledRegions`：混合状态下只返回启用地域，并带回禁用地域列表
  - `TestListRegion_AllDisabled_ReturnsError`：全部禁用时返回 `all aws regions are disabled, disabled regions: [...]`
  - `TestListRegion_Empty_ReturnsError`：地域表为空时返回 `aws region is empty`
  - `TestListRegion_AllEnabled`：全部启用时禁用列表为空
