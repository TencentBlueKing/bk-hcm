## Why

资源同步任务在执行过程中会遍历账号下所有地域，逐个调用云厂商 API 获取资源信息。当部分地域因账号权限受限、服务未开通或云厂商侧异常导致接口调用失败时，整个同步任务被迫中断，影响其他正常地域的资源数据更新。

现状痛点：
- 部分 AWS 地域（如 opt-in regions）需要手动启用后才能访问，未启用的账号调用时返回 `AuthFailure`；
- 部分地域因合规或业务原因未开通服务，调用失败后阻断整个同步流程；
- 运维人员只能通过修改代码或手动跳过异常地域，缺乏灵活的管理手段。

## What Changes

- **新增 `sync_enable` 字段**：在 `aws_region` 表中新增 `sync_enable` 字段（TINYINT(1)），标识该地域是否参与资源同步。默认启用（sync_enable=1）。
- **同步过滤逻辑改造**：AWS 地域同步时，仅查询并处理 `sync_enable=1` 的地域，跳过被禁用的地域。
- **新增「地域同步状态管理」接口**：支持按账号/地域批量启用或禁用同步状态，避免直接操作数据库。
- **同步日志增强**：在同步任务执行时记录被过滤的地域信息，便于问题排查。

## Capabilities

### New Capabilities

- `aws-region-sync-enable`：AWS 地域同步开关功能，包含数据库字段扩展、同步过滤逻辑、管理接口。

### Modified Capabilities

- `aws-region-sync`：现有 AWS 地域同步逻辑增加 `sync_enable` 过滤条件。

## Impact

- **数据层**：`aws_region` 表新增 `sync_enable` 列（需 DDL 变更）。
- **Table 结构体**：`pkg/dal/table/cloud/region/aws.go` 中 `AwsRegionTable` 新增 `SyncEnable` 字段。
- **Core/DS 结构体**：`pkg/api/core/cloud/region/aws.go` 和 `pkg/api/data-service/cloud/region/aws.go` 新增 `SyncEnable` 字段。
- **同步逻辑**：`cmd/hc-service/logics/res-sync/aws/region.go` 增加同步状态过滤与日志输出。
- **新增接口**：`data-service` 新增地域同步状态管理接口（启用/禁用）。
