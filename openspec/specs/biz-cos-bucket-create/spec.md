# biz-cos-bucket-create

## Requirements

### Requirement: 业务视角创建 COS 存储桶
系统 SHALL 提供业务视角的 COS 存储桶创建接口 `POST /api/v1/cloud/bizs/{bk_biz_id}/cos/buckets/create`。接口 MUST 根据业务 ID、负责人、备份负责人生成自研云标签，通过 `XCosTagging` 注入到腾讯云创建请求中。仅支持 TCloudZiyan 厂商。

#### Scenario: 成功创建带业务标签的存储桶
- **WHEN** 业务运维人员提交创建请求，包含 `account_id`、`region`、`name`、`manager`、`bak_manager`
- **THEN** 系统根据 `bk_biz_id`、`manager`、`bak_manager` 调用 `GenTagsForBizsWithManager` 生成业务标签（运营产品、一级业务、二级业务、运营部门、负责人、备份负责人），将标签转为 `XCosTagging` 格式注入请求，调用 hc-service 创建存储桶，返回成功

#### Scenario: 请求参数校验失败
- **WHEN** 请求缺少必填字段（如 `account_id`、`region`、`name`、`manager`、`bak_manager`）
- **THEN** 系统返回 `InvalidParameter` 错误，不调用腾讯云接口

#### Scenario: 账号厂商不支持
- **WHEN** 请求的 `account_id` 对应的厂商不是 `TCloudZiyan`
- **THEN** 系统返回错误，提示厂商不支持

#### Scenario: 业务标签生成失败
- **WHEN** 通过业务 ID 查询 CMDB/global_config 获取业务元数据失败
- **THEN** 系统返回错误，不调用腾讯云创建接口
