# biz-cos-bucket-delete

## Requirements

### Requirement: 业务视角删除 COS 存储桶（含归属校验）
系统 SHALL 提供业务视角的 COS 存储桶删除接口 `DELETE /api/v1/cloud/bizs/{bk_biz_id}/cos/buckets/delete`。删除前 MUST 先按业务标签查询腾讯云的存储桶列表，校验目标存储桶属于当前业务后才允许删除，防止越权操作。

#### Scenario: 成功删除属于当前业务的存储桶
- **WHEN** 请求删除存储桶 `bucket-x`，系统按业务标签查询腾讯云返回的存储桶列表中包含 `bucket-x`
- **THEN** 系统确认存储桶属于当前业务，调用 hc-service 删除存储桶，返回成功

#### Scenario: 存储桶不属于当前业务，拒绝删除
- **WHEN** 请求删除存储桶 `bucket-y`，但按业务标签查询腾讯云返回的存储桶列表中不包含 `bucket-y`
- **THEN** 系统返回错误，提示存储桶不属于当前业务或不存在，不执行删除操作

#### Scenario: 请求参数校验失败
- **WHEN** 请求缺少必填字段（如 `account_id`、`region`、`name`）
- **THEN** 系统返回 `InvalidParameter` 错误
