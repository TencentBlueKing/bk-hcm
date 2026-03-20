## Context

data-service 层已完成 `permission_policy_library` 的全部 CRUD（table → DAO → handler → SDK client），且 IAM 资源类型 `meta.PermissionPolicyLibrary` 和 auth-server 适配均已就位。当前缺失 cloud-server handler 层，web-server 纯代理 `/api/v1/cloud/*` 至 cloud-server，因此只需在 cloud-server 新增模块即可打通链路。

参考实现：`cert`（标准 List 模式）、`sub_account`（带 vendor 路径参数的 ListExt 模式）。

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 实现权限策略库 List 接口，支持 vendor 过滤和 IAM 鉴权
- 响应结构体预留 `associated_account_count` 字段

**Non-Goals:**
- 不实现 Create/Update/Delete handler（后续按需扩展）
- 不实现 `associated_account_count` 的实际计算逻辑
- 不实现 biz 域接口（`/bizs/{bk_biz_id}/...`）

## Decisions

### 1. 使用 ListResourceAuthRes 做鉴权

`PermissionPolicyLibrary` 在 auth-server 中映射到 `sys.CloudVendorConfig` 全局权限，`ListAuthInstWithFilter` 会返回 `IsAny=true`（有权限时）或空 IDs（无权限时），`"account_id"` 维度参数不会生效。使用 `ListResourceAuthRes` 可保持与 cert、sub_account 等资源一致的鉴权模式。

**替代方案**：直接调用 `Authorizer.Authorize` 做简单的全局权限检查 — 拒绝，因为破坏了 cloud-server 统一的 List 鉴权模式。

### 2. cloud-server 定义独立的响应结构体

在 `pkg/api/cloud-server/permission_policy_library.go` 定义 `PermissionPolicyLibraryResult`，内嵌 `BasePermissionPolicyLibrary` 并追加 `AssociatedAccountCount` 字段。这样 DS 返回的 base 模型可以直接包装，cloud-server 层拥有独立的扩展点。

参考：`pkg/api/cloud-server/disk/response.go` 中 `DiskResult` 包装 `BaseDisk` + 额外字段的模式。

### 3. vendor 过滤通过 tools.And 合并到 filter

从 URL 路径提取 vendor 后，使用 `tools.And(authExpr, tools.EqualExpression("vendor", vendor))` 合并到最终 filter。与 `sub_account/list.go` 中 `ListSubAccountExt` 的处理方式一致。

### 4. 请求类型使用 cloud-server 的 proto.ListReq

cloud-server 接收 `proto.ListReq`（Filter + Page），鉴权并追加 vendor filter 后，构造 `protocloud.PermissionPolicyLibraryListReq` 传给 DS SDK。

### 5. svc 结构体不包含 audit

List 是只读操作，不需要审计，svc 结构体只需 `client` 和 `authorizer` 两个字段。

### 6. handler 拆分为公开入口 + 私有实现

参考 `sub_account/list.go` 的 `ListSubAccount` / `listSubAccount` 模式，将核心逻辑抽到 `listPermissionPolicyLibrary(cts, authHandler)` 私有方法中，公开的 `ListPermissionPolicyLibrary` 入口固定传入 `handler.ListResourceAuthRes`。后续扩展 biz 域时只需新增入口方法传入 `handler.ListBizAuthRes`。

## Risks / Trade-offs

- **associated_account_count 暂为 0** → 前端需感知此字段当前无实际含义，后续建立关联表后需回来补充计算逻辑。
- **仅实现 List** → Create/Update/Delete 的 cloud-server handler 未来需单独补充，但不影响当前查询能力。
