## Context

data-service 层已完整实现权限策略库的 CRUD（BatchCreate/BatchUpdate/BatchDelete/List），DAO、Table、Audit、Client 均已就绪。cloud-server 层目前仅暴露了 List 接口，Create/Update/Delete 尚未对外开放。本次仅聚焦 Create 接口。

现有代码依赖关系：
- DS 层 Create Handler：`cmd/data-service/service/cloud/permission-policy-library/create.go`
- DS TCloud Client：`pkg/client/data-service/tcloud/permission_policy_library.go` → `BatchCreate()`
- IAM 资源类型：`meta.PermissionPolicyLibrary` 已注册
- CS 层 service.go：已有 `svc` 结构体（含 `client` 和 `authorizer`）

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 层暴露单个创建语义的 `POST /vendors/{vendor}/permission_policy_libraries/create` 接口
- 实现完整的请求校验、IAM 鉴权、DS 调用链路
- CS 层请求模型 `bk_biz_ids` 标记为 required，要求调用方显式传递

**Non-Goals:**
- 不涉及 Update/Delete 接口的 CS 层实现
- 不涉及 `associated_account_count` 的实际计算逻辑
- 不修改 DS 层已有代码

## Decisions

### Decision 1：鉴权方式——直接调用 AuthorizeWithPerm

**选择：** 直接调用 `svc.authorizer.AuthorizeWithPerm(cts.Kit, ResourceAttribute{Type: PermissionPolicyLibrary, Action: Create})`，不走 `handler.ResOperateAuth`。

**理由：** 权限策略库不隶属于任何云账号（无 account_id 字段），`ResOperateAuth` 要求 `BasicInfos` 非空且依赖 `AccountID` 构建鉴权资源。参考 `CloudSelectionScheme` 的 Create 鉴权模式（`cmd/cloud-server/service/cloud-selection/scheme.go:88-97`），对于不绑定账号的平台级资源，直接基于资源类型+动作鉴权是项目已有的标准做法。

**备选：** 改造 `ResOperateAuth` 支持空 BasicInfos → 侵入性大，影响面广，收益不明显。

### Decision 2：单个创建 vs 批量创建的适配

**选择：** CS 层定义单个创建请求模型（flat JSON），Handler 内部包装为 DS 的 `PermissionPolicyLibraryBatchCreateReq`（数组长度为 1），返回 `{ "id": ids[0] }`。

**理由：** 接口文档定义的是单个创建语义。CS 层作为面向前端的 API 网关，保持简洁的单个创建接口；DS 层保持通用的批量接口供内部服务使用。这种"CS 单个 → DS 批量"的适配模式在项目中是常见做法。

### Decision 3：bk_biz_ids 在 CS 层 required

**选择：** CS 层请求模型 `BkBizIDs` 标记 `validate:"required"`，DS 层保持 `omitempty` 不变。

**理由：** CS 层作为面向前端的入口，要求调用方显式传递业务 ID 列表（即使为空数组 `[]`），表达"已关注业务归属"的意图。DS 层保持灵活以兼容内部批量调用场景。

## Risks / Trade-offs

- **[Risk] 鉴权粒度较粗** → 当前仅校验"用户是否有权对 PermissionPolicyLibrary 执行 Create"，不区分 vendor 维度。若未来需要按 vendor 细分权限，需扩展 IAM 资源模型。当前阶段 vendor 仅支持 tcloud，风险可控。
- **[Risk] CS 层 required vs DS 层 omitempty 不一致** → 内部服务直接调用 DS 时可以绕过 bk_biz_ids 校验。这是有意为之，DS 层服务间调用场景下的灵活性优先。
