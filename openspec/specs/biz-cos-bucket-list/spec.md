# biz-cos-bucket-list

## Requirements

### Requirement: 业务视角查询 COS 存储桶列表（按标签过滤）
系统 SHALL 提供业务视角的 COS 存储桶查询接口 `POST /api/v1/cloud/bizs/{bk_biz_id}/cos/buckets/list`。查询时 MUST 根据业务 ID 生成标签过滤条件（`TagKey`/`TagValue`），确保仅返回当前业务下的存储桶，不能越权查看其他业务的资源。

#### Scenario: 成功查询当前业务的存储桶列表
- **WHEN** 业务运维人员提交查询请求，包含 `account_id`、`region`
- **THEN** 系统根据 `bk_biz_id` 生成二级业务标签（`TagKey` = "二级业务"，`TagValue` = "业务名_业务ID"），传入腾讯云 ListBuckets 接口进行过滤，仅返回匹配标签的存储桶列表

#### Scenario: 当前业务下无存储桶
- **WHEN** 腾讯云按标签过滤后无匹配存储桶
- **THEN** 系统返回空列表，不报错

#### Scenario: 支持分页查询
- **WHEN** 请求包含 `max_keys` 和 `marker` 参数
- **THEN** 系统将分页参数传递给腾讯云接口，返回分页结果（含 `next_marker`、`is_truncated`）

#### Scenario: 请求参数校验失败
- **WHEN** 请求缺少必填字段（如 `account_id`、`region`）
- **THEN** 系统返回 `InvalidParameter` 错误
