## ADDED Requirements

### Requirement: 业务列表路由与鉴权

系统 SHALL 在 cloud-server 暴露与文档一致的接口：`POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/list`（或项目内等价路由注册方式，路径段与文档一致）。处理逻辑 SHALL 校验 `bk_biz_id`、`vendor`，并 SHALL 要求调用方对目标业务具备与其他 biz 云资源列表一致的业务访问权限；未授权时 SHALL 返回权限错误且不得返回业务数据。

#### Scenario: 已授权列表查询

- **WHEN** 调用方对已认证业务具备访问权限且请求参数合法
- **THEN** 系统 SHALL 返回符合过滤与分页语义的数据或仅返回总条数

#### Scenario: 未授权访问

- **WHEN** 调用方对 `bk_biz_id` 无业务访问权限
- **THEN** 系统 SHALL 拒绝请求且不泄露资源是否存在之外的敏感信息

### Requirement: 请求体与分页语义

系统 SHALL 接受并校验文档约定的输入字段，包括但不限于：`cloud_ids`、`names`、`cloud_sub_account_ids`、`cloud_account_ids`、`creator`、`reviser`、`page`。数组类过滤条件 SHALL 遵守最大长度（如 500，与项目同类接口一致）。`page` SHALL 支持 `count`：当 `count` 为 true 时 SHALL 仅返回总条数且 `details` 按文档为空数组或 null，且 SHALL 不应用分页偏移；当 `count` 为 false 时 SHALL 按 `start`/`limit` 返回明细且总条数字段按文档约定为 0 或另行一致实现。

#### Scenario: 仅统计条数

- **WHEN** `page.count` 为 true
- **THEN** 响应 SHALL 包含匹配过滤条件的 `count`，且 SHALL 不返回明细行（`details` 为空或 null）

#### Scenario: 明细分页

- **WHEN** `page.count` 为 false 且 `limit` 在允许范围内
- **THEN** 响应 SHALL 返回 `details` 切片及文档约定的分页行为

### Requirement: 二级账号条件在 cloud-server 解析

当请求包含与二级账号云侧标识相关的过滤（文档中的 `cloud_account_ids`，或经 cloud-server 映射后与 account 表 `cloud_id` 对应的等价条件）时，系统 SHALL 先在 **account** 表查询：`cloud_id` 属于给定集合且 `type` 为资源账号类型（`resource`），并 SHALL 得到本地 `account_id` 列表。若该条件存在且解析结果为空，系统 SHALL 直接返回空结果（`count` 为 0，明细为空），且 SHALL NOT 调用 data-service 的联表列表接口。

#### Scenario: 云侧二级账号 ID 无匹配

- **WHEN** 请求包含 `cloud_account_ids`（或非空等价条件）且 account 表中无匹配记录
- **THEN** 系统 SHALL 返回空列表或零计数且不触发 permission_template 联表查询

#### Scenario: 未传二级账号云侧条件

- **WHEN** 请求未包含任何二级账号云侧过滤条件
- **THEN** 系统 SHALL 不得仅因「未解析 account」而错误返回空结果；SHALL 按其余过滤条件继续查询

### Requirement: data-service 两表联表列表

data-service SHALL 提供 `ListPermissionTmplJoinExt`（名称以代码库一致为准），其数据访问 SHALL 通过 DAO `ListJoinSubAccount`（或等价命名）将 **permission_template** 与 **sub_account** 联表查询，以满足与子账号相关的过滤（如 `cloud_sub_account_ids`）及响应所需字段。该路径 SHALL NOT 在 SQL 中联接 **account** 表完成二级账号云侧 ID 过滤（该部分由 cloud-server 前置完成并以下推 `account_id` 方式约束 permission_template）。

#### Scenario: 按三级账号云上 ID 过滤

- **WHEN** 请求包含 `cloud_sub_account_ids` 且 vendor 有效
- **THEN** 返回的模板集合 SHALL 仅包含与这些云侧子账号存在关联关系的权限模板（关联语义与表结构一致）

### Requirement: 列表响应字段

当返回明细时，每条记录 SHALL 包含文档 `list_permission_template.md` 中列出的字段，包括：`id`、`cloud_id`、`name`、`vendor`、`account_id`、`cloud_account_id`、`policy_library_id`、`policy_library_name`、`policy_library_version`、`policy_library_sync_time`、`policy_document`、`memo`、`associated_sub_account_count`、`creator`、`reviser`、`created_at`、`updated_at`、`extension`。TCloud 下 `extension.cloud_type` SHALL 与文档枚举语义一致（自定义/预设策略类型）。

#### Scenario: TCloud 明细行形态

- **WHEN** vendor 为 tcloud 且返回一条模板
- **THEN** 响应 SHALL 包含 `extension.cloud_type` 及文档要求的策略库与策略正文字段

### Requirement: cloud-server 拼装 account 表字段

在 data-service 完成 `permission_template` 与 `sub_account` 联表列表后，cloud-server SHALL 根据明细中的 `account_id` **批量**获取 `account` 数据，并将文档要求的 **account 来源字段**（至少包含 `cloud_account_id`，及与文档一致的其他 account 投影）写入每条 API 明细。系统 SHALL NOT 在按行循环中对同一列表请求发起逐条 account 查询（禁止 N+1）。

#### Scenario: 多条模板同属一个二级账号

- **WHEN** `details` 中多条记录共享同一 `account_id`
- **THEN** 系统 SHALL 仅对该 `account_id` 查询或加载一次 account 信息（或等价批量接口一次覆盖），且每条明细的 `cloud_account_id` SHALL 正确一致

#### Scenario: 仅统计条数模式

- **WHEN** `page.count` 为 true 且无明细行
- **THEN** 系统 SHALL NOT 发起仅为拼装展示字段而进行的 account 批量加载

### Requirement: 封装与可复用组装逻辑

cloud-server 实现 SHALL 将 account 过滤解析、account 批量加载与 map 构建、以及联表结果到 API 响应结构的组装 **拆分为独立函数或集中在 `logics/account`（或与项目现有账号 logic 一致的位置）**，使 `list.go` 以编排为主。若与现有列表（如 sub_account_secret）存在相同拼装模式，实现 SHALL **抽取公共函数**而非复制实现。

#### Scenario: 组装逻辑可单测

- **WHEN** 对「account_id 列表 -> map -> 填充 `cloud_account_id`」进行单元测试
- **THEN** 测试 SHALL 能够不启动完整 HTTP 服务而仅调用封装后的纯函数或 logic 层函数（在仓库测试约定允许范围内）

### Requirement: Filters 类型与校验

系统 SHALL 在 `pkg/api/data-service/cloud` 定义 `PermissionTemplateFilters`，其风格 SHALL 与 `SubAccountSecretFilters` 对齐（含可选数组字段长度上限、`Extension` 字段用于厂商扩展过滤）。cloud-server SHALL 将 HTTP/API 请求映射为该 filters，并 SHALL 注入 vendor、biz、租户等现有全局作用域条件，与项目其他 cloud 列表一致。

#### Scenario: 超长过滤数组

- **WHEN** 任一数组过滤元素个数超过约定上限
- **THEN** 系统 SHALL 在访问数据库前返回参数错误

### Requirement: 实现一致性

Go 实现 SHALL 遵循项目错误与日志语言约定（英文）；SHALL 使用 `gofmt`/`goimports`；行宽 SHALL 遵守仓库规则（超过 120 列换行）。 SHALL 复用现有 kit、rest、errf、分页校验与 DAO list/count 模式。

#### Scenario: 与现有列表 DAO 一致

- **WHEN** 新增 DAO 联表列表方法
- **THEN** 错误处理、租户选项、count 与分页 SQL 结构 SHALL 与同目录其他资源 list 方法保持一致风格
