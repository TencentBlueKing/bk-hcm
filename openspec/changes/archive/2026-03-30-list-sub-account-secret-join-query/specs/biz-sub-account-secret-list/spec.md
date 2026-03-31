## 新增需求（ADDED Requirements）

### Requirement: 业务列表接口与鉴权

系统必须提供 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/sub_account_secrets/join/list`，用于在业务作用域下列出三级账号密钥（联表查询）。处理函数必须校验 `bk_biz_id` 与 `vendor`，必须按与其他 biz 云 API 一致的方式要求业务级访问权限，且对未授权调用方的处理不得泄露资源是否存在。

#### Scenario: 已授权业务用户查询列表

- **WHEN** 已认证调用方对给定 `bk_biz_id` 具备业务访问权限，且传入合法的列表参数
- **THEN** 系统返回符合过滤条件且限定在该业务与 vendor 下的列表或统计数量

#### Scenario: 未授权调用方

- **WHEN** 调用方对 `bk_biz_id` 无业务访问权限
- **THEN** 系统返回鉴权错误，且不得返回任何密钥数据行

### Requirement: 请求校验

系统必须校验全部列表请求字段：`page`（含 `count` 语义、`count` 为 false 时的 `start`/`limit`、`limit` 最大 500）、可选数组字段最大长度 500（`account_ids`、`sub_account_ids`、`account_managers`、`sub_account_managers` 及厂商 extension 中的数组）、以及存在时的 `status`。非法组合必须以明确的参数错误拒绝。

#### Scenario: 仅统计条数

- **WHEN** `page.count` 为 true
- **THEN** 系统仅返回 `count`，`details` 按接口约定为空或 null，且不得应用分页偏移

#### Scenario: 过滤数组超长

- **WHEN** 任一过滤数组超过文档规定的最大长度
- **THEN** 系统在访问数据库之前拒绝该请求

### Requirement: 联表数据源与过滤

系统必须通过将 `sub_account_secret` 与 `sub_account`、`account` 联表来解析列表结果，使过滤与投影可使用：

- 表主键：`account_ids` → `account.id`，`sub_account_ids` → `sub_account.id`
- TCloud 云侧标识（extension）：`cloud_main_account_ids` → `account.extension.cloud_main_account_id`，`cloud_sub_account_ids` → `sub_account.extension.uin`，`cloud_secret_ids` → 密钥 extension 中云密钥 ID 字段
- 负责人过滤：`account_managers` 对应账号负责人，`sub_account_managers` 对应子账号负责人

连接条件必须为 `sub_account_secret.account_id = account.id` 且 `sub_account_secret.sub_account_id = sub_account.id`，vendor 与租户约束与现有 DAO 模式一致（查询层仍可按租户隔离；**响应体不得包含** `tenant_id`）。

#### Scenario: 按二级账号表 ID 过滤

- **WHEN** 传入 `account_ids`
- **THEN** 仅包含 `account_id` 属于给定 id 集合的密钥

#### Scenario: 按 TCloud 云三级账号 ID 过滤

- **WHEN** vendor 为 tcloud 且传入 `extension.cloud_sub_account_ids`
- **THEN** 仅包含关联子账号 extension 中 `uin` 与给定云侧 id 匹配的密钥

### Requirement: 明细模式下的响应形态

当 `page.count` 为 false 时，每条明细必须包含密钥基础字段、厂商 extension（tcloud：`cloud_secret_id`、`cloud_main_account_id`、`cloud_sub_account_id`、来自子账号 extension 的 `console_login`）、来自账号的 `account_managers`、`sub_account_managers`，JSON 字段名与已发布 API 一致（`sub_account_managers`、`account_managers`）；**不得**包含 `tenant_id` 字段。

#### Scenario: TCloud 明细行

- **WHEN** vendor 为 tcloud 且返回某条密钥记录
- **THEN** 响应包含 API 文档约定的 `console_login` 与各云侧 id 字段

### Requirement: 复用能力与代码质量

实现必须复用现有辅助逻辑（filter 构建、`ListOption` 校验、类似 `listBizSubAccountAuthRes` 的 biz 鉴权模式、ORM 租户选项、`pkg/dal/dao/cloud` 中既有 join SQL 风格）。厂商相关逻辑在行为分叉处必须在函数名或注释中带 `TCloud`（或对应 vendor）前缀。Go 源码行宽须遵守项目规则（超过 120 列需换行）。

#### Scenario: DAO 列表实现

- **WHEN** 在 DAO 中新增或扩展子账号密钥列表
- **THEN** 实现结构与其他云资源 DAO 的 list/join 方法一致（校验、count 与分页 SQL、错误日志）
