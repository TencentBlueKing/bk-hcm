## Why

业务侧需要按 `list_permission_template.md` 暴露「云权限模板列表」接口：支持多维度过滤、分页与仅统计条数模式，并在响应中返回策略库元数据、策略正文、关联三级账号数量及 TCloud `extension.cloud_type` 等字段。仅靠单表查询无法满足「按三级账号云上 ID 过滤」「关联子账号计数」等需求，且二级账号相关条件需在 cloud-server 结合 `account` 表解析后再下推到 `permission_template` 查询。

## What Changes

- 在 **cloud-server** 新增 `cmd/cloud-server/service/permission_templates/`（含 `service.go` 注册路由、`list.go` 实现列表），路径与文档一致：`POST .../bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/list`；鉴权与同类 biz 列表一致（业务访问）。
- **cloud-server 分层查询**：当请求包含与二级账号相关的条件时，先在 **account** 表按 `cloud_id`（对应请求/扩展中的云侧二级账号标识）且 `type = resource` 解析出本地 `account_id` 集合；若无匹配则直接返回空列表/零计数，不再调用 data-service 列表。
- 在 **data-service** 新增列表能力 **`ListPermissionTmplJoinExt`**：`permission_template` 与 **`sub_account`** 两表联查（DAO：`ListJoinSubAccount`），用于 `cloud_sub_account_ids` 等与子账号相关的过滤，以及返回联表所需字段；**不**在 data-service 联查 `account`（与 `sub_account_secret` 三表联查模式区分）。
- **cloud-server 结果拼装**：在拿到 data-service 联表结果后，由 **上层（cloud-server）** 根据行内 `account_id` **批量加载 account**（或复用前置解析阶段的缓存），拼装响应中的 **`cloud_account_id` 等 account 表字段**；禁止在循环中对每行单独打 data-service（避免 N+1），并通过 **合理封装与公共函数抽取**（例如放在 `cmd/cloud-server/logics/account` 或与现有 list 组装共用的小包内）保持 `list.go` 精简、可测。
- 在 **`pkg/api/data-service/cloud`** 定义 **`PermissionTemplateFilters`**（风格对齐 `SubAccountSecretFilters`：主键/状态类字段 + `Extension`），承载列表过滤；cloud-server 将 HTTP 请求字段映射为 filters（含 extension 中云侧 ID 数组等）。
- 补齐 **cloud-server / data-service API 类型、客户端、路由注册** 及与文档一致的响应字段（含 `associated_sub_account_count`、`policy_library_*`、`extension`）。

## Capabilities

### New Capabilities

- `biz-permission-templates-list`：业务作用域下云权限模板列表；account 条件在 cloud-server 解析为 `account_id`；data-service 对 `permission_template` 与 `sub_account` 联表列表；**列表返回后**在 cloud-server **批量拼装 account 展示字段**；行为与 `docs/api-docs/.../list_permission_template.md` 对齐。

### Modified Capabilities

- （无）`openspec/specs/` 中尚无权限模板列表基线 spec，本次仅新增能力 spec。

## Impact

- **cloud-server**：`cmd/cloud-server/service/permission_templates/`；**账号过滤解析 + 列表后 account 字段拼装**（封装至 `logics/account` 或与现有列表组装共用的辅助函数）；主服务路由挂载。
- **data-service**：`cmd/data-service/service/cloud/permission-template/`（或等价包名）及 handler 注册；DAO `pkg/dal/dao/cloud/permission-template/`（`ListJoinSubAccount`、`List`/`Count` 分页规则）。
- **共享类型**：`pkg/api/data-service/cloud`、`pkg/api/cloud-server`（若已有 proto 则扩展）、`pkg/client/data-service`。
- **数据表**：`permission_template`、`sub_account`（联表）；`account`（仅 cloud-server 侧前置解析）。
- **文档**：以现有 `list_permission_template.md` 为契约；若实现中发现字段命名与代码不一致，在 apply 阶段对齐代码或文档（本 proposal 不强制改文档）。
