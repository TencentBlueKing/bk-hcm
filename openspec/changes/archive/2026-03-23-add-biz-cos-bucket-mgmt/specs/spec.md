# biz-cos-bucket-create

## ADDED Requirements

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

---

# biz-cos-bucket-delete

## ADDED Requirements

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

---

# biz-cos-bucket-list

## ADDED Requirements

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

---

# cos-appid-auto-concat

## ADDED Requirements

### Requirement: hc-service 自动拼接 BucketName 与 APPID
hc-service 的 `CreateTCloudZiyanCosBucket` MUST 在创建存储桶前检查 BucketName 是否已包含正确的 APPID 后缀。如果未包含，MUST 通过 CAM `GetUserAppId` 接口获取 APPID 并自动拼接。

#### Scenario: BucketName 已包含正确的 APPID 后缀
- **WHEN** 请求的 BucketName 为 `my-bucket-1250000000`，且通过 CAM 获取的 APPID 为 `1250000000`
- **THEN** 系统判断后缀匹配，不做额外处理，直接使用原始 BucketName 创建存储桶

#### Scenario: BucketName 未包含 APPID 后缀
- **WHEN** 请求的 BucketName 为 `my-bucket`，通过 CAM 获取的 APPID 为 `1250000000`
- **THEN** 系统自动拼接为 `my-bucket-1250000000`，使用拼接后的名称创建存储桶

#### Scenario: BucketName 包含错误的 APPID 后缀
- **WHEN** 请求的 BucketName 为 `my-bucket-9999999999`，但通过 CAM 获取的 APPID 为 `1250000000`
- **THEN** 系统判断后缀不匹配，自动拼接正确的 APPID，使用 `my-bucket-9999999999-1250000000` 创建存储桶

#### Scenario: 获取 APPID 失败
- **WHEN** 调用 CAM `GetUserAppId` 接口失败
- **THEN** 系统返回错误，不创建存储桶

---

# tcloud-account-info

## MODIFIED Requirements

### Requirement: GetAccountInfoBySecret 返回 AppId
`GetAccountInfoBySecret` 方法调用 CAM `GetUserAppId` 接口时，MUST 同时返回 `AppId` 字段（整数类型）。`TCloudInfoBySecret` 结构体 MUST 新增 `AppId int64` 字段。

#### Scenario: 成功获取包含 AppId 的账号信息
- **WHEN** 调用 `GetAccountInfoBySecret` 且 CAM 接口正常返回
- **THEN** 返回的 `TCloudInfoBySecret` 包含 `CloudMainAccountID`（OwnerUin）、`CloudSubAccountID`（Uin）以及 `AppId`（如 `1250000000`）

#### Scenario: CAM 接口返回的 AppId 为空
- **WHEN** 调用 CAM 接口成功但 `resp.Response.AppId` 为 nil
- **THEN** 系统返回错误，提示 AppId 为空
