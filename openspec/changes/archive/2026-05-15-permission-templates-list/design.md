## Context

- 已发布接口说明：`docs/api-docs/web-server/docs/biz/permission-template/list_permission_template.md`（`POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/list`）。
- 参考实现：`cmd/cloud-server/service/subaccount-secret/list.go` + data-service `ListSubAccountSecretJoinExt`（**三表**联查）；本需求中 **permission_template 列表仅在 data-service 做 permission_template ↔ sub_account 两表联查**，**account 不参与该 SQL**。**account 在 cloud-server 侧承担两件事**：（1）过滤阶段将云侧二级账号 ID 解析为 `account_id`；（2）**列表数据返回后**，将 `account` 表上的展示字段（如 `cloud_account_id`）**批量拼装**进 API 响应。
- 表定义：`pkg/dal/table/cloud/permission_template.go`、`pkg/dal/table/cloud/sub-account/sub_account.go`、`pkg/dal/table/cloud/account.go`。
- Filters 参考：`pkg/api/data-service/cloud/sub_account_secret.go` 中 `SubAccountSecretFilters`（数组长度、`Extension` 等校验风格）。

## Goals / Non-Goals

**Goals:**

- 实现与文档一致的请求/响应及分页语义（含 `page.count === true` 时仅返回 `count`，`details` 为空/null）。
- cloud-server：业务访问鉴权；vendor、`bk_biz_id` 校验；将 **二级账号云侧 ID** 条件转为本地 `account_id` 集合后再调 data-service；若无匹配 `account_id` 则短路返回空结果。
- data-service：提供 `ListPermissionTmplJoinExt`，DAO 层 `ListJoinSubAccount` 生成联表 SQL，支持文档中的 `cloud_sub_account_ids`（关联子账号）等过滤，并支持计算或聚合 **关联三级账号数**（按文档字段 `associated_sub_account_count`）。
- cloud-server：**在联表列表成功返回后**，对明细中的 **`account_id` 去重**，**一次或批量**拉取 `account` 记录（data-service 既有 account 查询能力或现有 logic），再 **map 回填** `cloud_account_id` 等字段；与「仅 count」模式对齐（无明细则跳过拼装）。
- **封装**：将「按 `cloud_id` + `type=resource` 解析 `account_id`」「按 `account_id` 集合批量取 account 并建索引 map」「将 join 行转为 API 明细」拆成 **小函数或 `logics/account` 包内可复用导出函数**，`list.go` 只编排 kit、鉴权、调用顺序；若与 `subaccount-secret` 等列表存在相同拼装模式，**抽取公共函数**而非复制粘贴。
- 定义 `PermissionTemplateFilters`，涵盖文档中的 `cloud_ids`、`names`（模糊）、`cloud_sub_account_ids`、`creator`、`reviser` 等，以及与厂商扩展相关的字段（如 TCloud `extension.cloud_type` 若作为过滤条件则进入 `Extension` JSON 过滤路径）。

**Non-Goals:**

- 不在本次变更中修改权限模板创建/更新/删除等其他接口。
- 不在 data-service 将 `account` 并入同一 SQL join（明确排除，以保持与密钥列表实现差异清晰）。

## Decisions

1. **Account 解析在 cloud-server**  
   - **做法**：当存在「二级账号云侧 ID」类条件时（对外文档字段为 `cloud_account_ids`；对内 filters/extension 可与 `SubAccountSecret` 类似使用 `cloud_main_account_ids` 映射至 `account.cloud_id`，以与现有 extension 命名一致——具体字段名在实现时与 `account` 表及已有 API 对齐），查询 `account`：`cloud_id IN (...)` 且 `type = resource`（及 biz/vendor/租户等现有约束），得到 `account_id` 列表。  
   - **理由**：用户明确要求 account 关联在 cloud-server 完成；减少 data-service SQL 复杂度并复用账号域已有查询模式。  
   - **备选**：三表 join（与 sub_account_secret 一致）— **否决**，与需求不符。

2. **data-service 仅 permission_template + sub_account 联表**  
   - **做法**：DAO `ListJoinSubAccount` 以 `permission_template` 为主表（或与现有 DAO 命名一致），JOIN `sub_account` 于「模板与三级账号关联」所需条件（具体 JOIN 条件以实现时现有 schema 为准：如关联表或 sub_account_id 字段；若当前 schema 为间接关联，则在设计中落地为实际外键/中间表查询）。  
   - **理由**：满足 `cloud_sub_account_ids` 与关联计数；避免 account 进 data-service。  
   - **备选**：单表 + 多次查询— 仅在性能不可接受时再评估。

3. **`PermissionTemplateFilters` 形态**  
   - **做法**：对齐 `SubAccountSecretFilters`：`[]string` 过滤 + `validate` 标签（`max=500` 等）、`Extension tabletypes.JsonField` 用于厂商扩展数组/枚举过滤。  
   - **理由**：统一 data-service 列表校验与构建 filter 的体验。

4. **空 account_id 短路**  
   - **做法**：只要本次请求**需要**按二级账号云侧 ID 过滤且解析结果为空，则返回 `count: 0` 与空 `details`（或文档约定的 null），**不**调用 data-service 列表。  
   - **理由**：与用户描述一致，避免无意义全表扫描。  
   - **注意**：若请求**未带**二级账号云侧条件，则不应因「未解析」而错误返回空（仅在有该条件时短路）。

5. **响应字段与 account 拼装**  
   - **做法**：联表结果携带 `permission_template.account_id`（及子账号侧字段）；**`cloud_account_id` 等来自 account 表的字段一律在 cloud-server 拼装**：对返回明细中的 `account_id` 去重后 **批量查询 account**（单次 list by ids 或等价批量接口），构建 `id -> account` map，再填充每条 API 明细。若过滤阶段已加载过同一批 account，**可复用同一 map**（避免重复 RPC），仅在 ID 集合不完全重叠时再补查。  
   - **理由**：data-service 不 join account；上层统一负责 account 投影，边界清晰。  
   - **备选**：在 data-service 三表 join 带出 `cloud_account_id`— **否决**（与既定两表联查方案冲突）。

6. **封装与公共函数抽取**  
   - **做法**：  
     - 将 **account 过滤解析**（`cloud_id` IN + `type=resource` + biz/vendor/租户）封装为可测函数，例如 `ResolveResourceAccountIDs(kt, bizID, vendor, cloudIDs) ([]string, error)`（实际命名与签名以实现时为准）。  
     - 将 **批量加载 account 并建 map** 封装为 `LoadAccountsByIDs(kt, ids) (map[string]*AccountCore, error)` 或与现有 `cmd/cloud-server/logics/account` 中已有函数合并/复用。  
     - 将 **join 行 -> 云服务器 API 结构体** 的转换放在 `list.go` 同包内的 `buildListDetail(...)` 或独立 `assemble.go`，入参包含 account map，**不在 handler 内写长段字段拷贝重复逻辑**。  
   - **理由**：便于单测、避免 N+1、后续其他 biz 列表若同样「data-service 不 join account」可复用。  
   - **原则**：优先 **扩展现有 logic**；若必须新建，保持包职责单一（account 相关不进 permission_template DAO）。  

## Risks / Trade-offs

- **[Risk] permission_template 与 sub_account 实际关联方式与假设不符** → **Mitigation**：apply 前阅读表结构与现有 DAO；在 `tasks.md` 中安排「确认关联模型」为首步。  
- **[Risk] `names` 模糊查询性能** → **Mitigation**：沿用项目既有 LIKE + 索引策略；必要时限制长度。  
- **[Risk] `associated_sub_account_count` 聚合开销** → **Mitigation**：在联表查询中使用子查询/COUNT 分组，与分页/count 模式区分（count-only 模式避免拉明细）。  
- **[Risk] account 批量查询与过滤解析重复打 data-service** → **Mitigation**：同一请求内合并 ID 集合、复用 map；封装在单一 logic 入口，避免 `list.go` 多处散落 RPC。  

## Migration Plan

- 纯新增 API 与 data-service 方法：无数据迁移。  
- 部署顺序：先 **data-service**（新 RPC/HTTP 路由），再 **cloud-server**（新路由），或同步发布；兼容旧客户端（仅新增路径）。  

## Open Questions

- `permission_template` 与 `sub_account` 的**精确关联路径**（直接外键 vs 关联表）需在实现时以代码库为准并在 spec/scenario 中写清。  
- 文档字段 `cloud_account_ids` 与 extension 内 `cloud_main_account_ids` 的**最终映射**（是否统一为同一语义）在编码时与 `account` 表 `cloud_id` 字段对齐并补充单测。  
