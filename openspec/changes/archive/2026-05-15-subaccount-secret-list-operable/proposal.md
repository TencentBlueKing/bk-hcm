## Why

业务侧「子账号密钥」列表接口目前只返回联表后的密钥与账号信息，前端无法像「业务子账号」列表那样区分当前业务是否可对该密钥执行后续操作。需要与子账号扩展列表一致的 **operable** 语义：按密钥关联的 `account_id` 与账号所属 `bk_biz_id` 对比当前请求的 `bk_biz_id`。

## What Changes

- 在 cloud-server 层 `ListSubAccountSecret` 的返回中，为每条明细增加布尔字段 **operable**（业务是否有权操作该密钥）。
- 在 `pkg/api/cloud-server/sub-account-secret` 中新增响应组合结构体 `BizSubAccountSecretJoinExtDetail`（Biz 前缀与业务子账号扩展列表命名一致），在现有 join 明细形态基础上嵌入 **operable**，不污染 data-service / core 的纯数据模型。
- 列表处理在拿到 data-service 的 join 结果后，批量解析 `account_id`，复用 `cmd/cloud-server/logics/account` 中已有的账号信息与 operable 映射构建逻辑（与 `ListBizSubAccountExt` / `convertBizSubAccountExtList` 一致：`account.BkBizID == 当前 bk_biz_id` 则为 true，缺账号为 false）。

## Capabilities

### New Capabilities

（无；行为扩展归入既有「业务子账号密钥列表」能力。）

### Modified Capabilities

- `biz-sub-account-secret-list`: 在业务 join 列表的明细响应中增加 **operable** 字段及实现约束（复用账号层 operable 计算）。

## Impact

- **代码**: `cmd/cloud-server/service/subaccount-secret/list.go`（调用 data-service 后封装转换）、`pkg/api/cloud-server/sub-account-secret/sub_account_secret.go`（新结构体与列表结果类型，若需）。
- **API**: 业务子账号密钥 join 列表响应 JSON 增加 `operable`；需同步 **web-server API 文档**（若仓库内已有对应 md）。
- **依赖**: 无新外部依赖；依赖现有 `logicaccount.BatchBuildOperableAndNameMap` / `BuildOperableMapByAccountMap` 与 Global Account 列表查询。
