## ADDED Requirements

### Requirement: aws_region 表新增 sync_enable 字段
系统 SHALL 在 `aws_region` 表中新增 `sync_enable` 字段，类型为 `TINYINT(1)`，默认值为 1（启用），用于控制该地域是否参与资源同步。字段 SHALL 有对应索引以提升查询效率。

#### Scenario: 查询 aws_region 表结构
- **WHEN** 执行 `DESCRIBE aws_region`
- **THEN** 返回结果中包含 `sync_enable` 字段，类型为 `tinyint(1)`，默认值为 `1`

#### Scenario: 新增地域时默认启用同步
- **WHEN** 从云端同步到新地域并插入 `aws_region` 表
- **THEN** 新记录的 `sync_enable` 字段值为 `1`

---

### Requirement: Table/Core/DS 结构体扩展 SyncEnable 字段
`AwsRegionTable`（Table 层）、`AwsRegion`（Core 层）以及 data-service 请求/响应结构体 SHALL 新增 `SyncEnable` 字段，类型为 `bool`，用于承载同步开关状态。

#### Scenario: Table 结构体包含 SyncEnable 字段
- **WHEN** 查看 `pkg/dal/table/cloud/region/aws.go` 中的 `AwsRegionTable` 定义
- **THEN** 结构体包含 `SyncEnable bool` 字段，db tag 为 `sync_enable`

#### Scenario: Core 结构体包含 SyncEnable 字段
- **WHEN** 查看 `pkg/api/core/cloud/region/aws.go` 中的 `AwsRegion` 定义
- **THEN** 结构体包含 `SyncEnable bool` 字段，json tag 为 `sync_enable`

---

### Requirement: 地域同步时按 sync_enable 过滤
资源同步任务在查询待同步地域时，SHALL 仅返回 `sync_enable = 1` 的地域记录。被禁用（`sync_enable = 0`）的地域 SHALL 被跳过，不调用云厂商 API。

#### Scenario: 同步任务过滤禁用地域
- **GIVEN** `aws_region` 表中存在 region_id=`us-east-1`（sync_enable=1）和 region_id=`ap-east-1`（sync_enable=0）
- **WHEN** 触发 AWS 资源同步任务
- **THEN** 同步逻辑仅处理 `us-east-1`，不调用 `ap-east-1` 相关云厂商 API

#### Scenario: 所有地域禁用时返回错误
- **GIVEN** `aws_region` 表中所有地域 `sync_enable = 0`
- **WHEN** 触发 AWS 资源同步任务，调用 `ListRegion` 函数
- **THEN** 返回 `aws region is empty` 错误

---

### Requirement: 同步任务输出启用地域日志
同步任务执行时，SHALL 在日志中记录启用同步的地域列表，便于运维人员排查问题。

#### Scenario: 日志记录启用同步的地域
- **GIVEN** `aws_region` 表中 `us-east-1` 和 `us-west-2` 的 `sync_enable = 1`
- **WHEN** 触发 AWS 资源同步任务
- **THEN** 日志中包含启用同步的地域列表信息

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
- **WHEN** 调用 `PATCH /api/v1/cloud/vendors/aws/regions/sync_enable/batch`
- **THEN** 返回 400 错误，提示 `ids` 不能为空

#### Scenario: IDs 超过 100 个返回错误
- **GIVEN** 请求体中 `ids` 数组包含 101 个元素
- **WHEN** 调用 `PATCH /api/v1/cloud/vendors/aws/regions/sync_enable/batch`
- **THEN** 返回 400 错误，提示 `ids` 数量超过限制

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
- **THEN** `TestRegionBatchUpdateSyncEnableReq_Validate` 测试通过，覆盖空 IDs、超过 100 个 IDs 等校验场景

#### Scenario: 同步过滤逻辑测试
- **WHEN** 执行 `go test ./cmd/cloud-server/service/sync/aws/...`
- **THEN** 以下测试用例全部通过：
  - `TestListRegion_OnlyEnabledRegions`：验证 `sync_enable = true` 的地域正常返回
  - `TestListRegion_FilterDisabledRegions`：验证 `sync_enable = false` 的地域被过滤
  - `TestListRegion_AllDisabled_ReturnsError`：验证所有地域禁用时返回 `aws region is empty` 错误
  - `TestListRegion_MixedEnableStatus`：验证混合启用/禁用状态下仅返回启用的地域
  - `TestListRegion_FilterConditionValidation`：验证过滤条件包含 `account_id` 和 `sync_enable` 字段
