## 上下文

- **接口契约**：`list_sub_account_secret.md` 定义过滤、分页与响应字段；负责人等来自 `account` 联表；**本实现不在响应中返回 `tenant_id`**（联表仍可用 `account` 做过滤与负责人）。TCloud 扩展相关字段来自 `sub_account`（`uin` 即云侧三级账号 ID，`console_login`）。
- **现状**：data-service 的 `ListSubAccountSecretWithExtension` 构造 `filter.Expression` 后调用 `SubAccountSecretDao.List`，仅查询 `sub_account_secret` 单表。cloud-server `subaccount-secret` 目前只注册了创建；`list.go` 为空。
- **联表参考**：`AccountBizRelDao.ListJoinAccount` 展示了 `LEFT JOIN`、`FieldsNamedExpr` 与表别名用法。子账号、账号 DAO 已为 filter 校验注册扩展字段类型（如子账号的 `extension.uin`、`extension.console_login`；账号列表的主账号 ID 等字段）。

## 目标 / 非目标

**目标：**

- 端到端实现业务列表：cloud-server（鉴权 + 请求）→ data-service → DAO 联表查询，返回约定字段（**不含**响应中的 `tenant_id`）。
- 将过滤语义集中在 data-service/DAO，cloud-server 保持薄层（解码、鉴权、调客户端）。
- 参数校验与业务 + 租户安全策略与现有云资源保持一致。

**非目标：**

- 除 tcloud 外其他厂商的业务列表支持（若路由已要求通用 vendor 分支，可对非 tcloud 返回不支持，待 spec 扩展）。
- 修改 `sub_account_secret`、`sub_account`、`account` 的物理表结构。
- 密钥的批量创建/更新/删除行为（本变更不覆盖）。

## 技术决策

1. **DAO 新方法 vs 在 service 里写大块 SQL**  
   **决策**：在 `SubAccountSecretDao` 上增加专用方法（如 `ListJoinAccountSubAccount` 或 `ListBiz`），用别名 `sass`、`sa`、`acc`（与仓库惯例一致）拼 `SELECT`，基于扩展后的 `ListOption` 规则集生成 `whereExpr`，映射到新的结果结构体（嵌入密钥行 + 反规范化字段：负责人、响应所需的 extension 片段；**不**映射 `tenant_id` 到 API 输出）。  
   **理由**：SQL 与列校验集中维护；与 `ListJoinAccount` 模式一致。  
   **备选**：service 多次查询 — 因负责人交集、云 ID 等过滤较复杂而放弃。

2. **过滤表达式构造**  
   **决策**：在 data-service 用专用列表请求 DTO 构造 `filter.Expression`（优先，贴合文档 JSON），或明确 API 字段到 filter 路径的映射（`account_id` 的 `in` 规则、`sub_account_id`、extension 与负责人数组的 JSON 路径）。复用 `tools` 包（`EqualExpression`、`And`、必要时 `RuleJSONContains`）。  
   **理由**：复用现有 `List` 的 filter → SQL 管线。

3. **cloud-server 鉴权**  
   **决策**：对齐 `ListBizSubAccountExt`：解析 `bk_biz_id`，对子账号密钥列表使用合适的业务资源动作（与其他 biz 子账号密钥操作同族，若已注册则如 `meta.SubAccountSecret` + `meta.Find`），将业务作用域并入 filter（`bk_biz_ids` 在子账号或账号表上取决于业务绑定方式 — 与当前密钥归属方式一致，很可能为子账号 `bk_biz_ids` 含业务 ID）。  
   **理由**：与 `listBizSubAccountAuthRes` 的体验与 IAM 一致。

4. **代码中的 TCloud 命名**  
   **决策**：联表逻辑尽量与厂商无关；仅在 extension 字段映射存在差异的函数名或注释中使用 `tcloud`/`TCloud`（例如 API 的 `cloud_sub_account_id` 对应表字段 `uin`）。

5. **响应 DTO**  
   **决策**：在 core/cloud-server 侧新增或扩展列表结果类型，包含 `AccountManagers`、`SubAccountManagers`（**不含** `TenantID`），与现有 `SubAccountSecret[T]` 并存或使用 biz 专用包装，避免破坏通用 list 调用方。

## 风险与权衡

- **[风险] 过滤 SQL 复杂** → **缓解**：对 WHERE 生成做单测或 sqlite/mysql 集成测试；`columnTypes` 与 join 别名在同一处维护。
- **[风险] 大租户性能** → **缓解**：保证 `account_id`、`sub_account_id`、`vendor` 等索引；分页上限与文档一致（500）。ORM 租户注入仍按现有 DAO 行为作用于查询，与是否返回 `tenant_id` 无关。
- **[风险] 文档表格与 JSON 字段名不一致**（如 `account_manager` vs `account_managers`）→ **缓解**：实现与同文档中的 JSON 示例对齐（复数字段名）。

## 迁移与发布

- data-service 与 cloud-server 同步或先发 data-service；新路由为增量能力；表结构不变则无需 DB 迁移。
- 回滚：去掉路由注册与 DAO 新方法；不涉及数据迁移。

## 待确认问题

- 本路由「业务访问」对应的 IAM Action 精确取值 — 与 `pkg/iam` 中 `SubAccountSecret` 列表定义核对。
- 业务作用域应仅通过 `sub_account.bk_biz_ids`、`account.bk_biz_id` 还是两者组合过滤 — 与 `CreateBizSubAccountSecret` 如何绑定业务对齐确认。
