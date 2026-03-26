## Context

现有 CLB 申领提单接口从 `cts.Kit.User`（HTTP 请求头 `X-Bkapi-User-Name`）获取提单人，系统间调用时 `Kit.User` 是系统虚拟用户，无法为自然人提单。需要新增一个系统提单接口，允许通过请求体显式传递提单人。

**现有创建流程**：
1. `CreateForCreateLB` Handler → 解析 vendor、解码公共请求、权限检查、根据 vendor 创建对应 Handler
2. `create()` 通用方法 → `CheckReq` → `PrepareReq` → `createItsmTicket`（用 `Kit.User` 作 Creator）→ `createApplication`（用 `Kit.User` 作 Applicant）

## Goals / Non-Goals

**Goals:**
- 新增系统提单接口，支持显式指定提单人
- 最大程度复用现有代码，减少维护成本
- 路由结构支持后续扩展（其他 application 类型的系统提单）

**Non-Goals:**
- 不修改原有用户提单接口
- 不新增系统调用方鉴权逻辑（由上层网关负责）
- 不新增审计记录（保持与原接口一致）
- 不修改查看/取消/列表接口

## Decisions

### Decision 1: 采用显式参数传递 applicant

**选择**：为 `create()`、`createItsmTicket()`、`createApplication()` 增加 `applicant string` 参数，显式传递提单人身份。不覆盖 `cts.Kit.User`。

**理由**：
- 显式传参，语义清晰，不污染 Kit 共享状态
- 已有调用方传 `cts.Kit.User` 作为 applicant，行为完全不变
- 新系统提单接口传 `req.Applicant`，与普通接口走同一条代码路径
- 改动范围可控：3 个私有方法签名 + 7 处已有调用方机械添加参数

**备选方案**：
- 方案 B：覆盖 `cts.Kit.User = req.Applicant` 后调用 `create()` → 代码变更最小，但修改共享状态是副作用，且会导致权限检查对象错误（系统提单场景下应检查系统调用方权限，而非 applicant 权限）
- 方案 C：新建 `createWithApplicant()` 方法 → 与 `create()` 大量重复代码

### Decision 2: 权限检查使用系统调用方身份

**选择**：不覆盖 `cts.Kit.User`，`checkApplyResPermission` 检查的是系统调用方（如 `bk_system`）的权限，而非 applicant 的权限。

**理由**：
- 系统提单场景下，调用方（系统用户）与提单人（自然人）是不同实体，权限应检查调用方
- applicant 是被代提单的用户，不要求其本身拥有申请资源的权限
- 系统调用方的访问控制由上层网关 + 接口内权限检查双重保障

### Decision 3: 新增 SysCreateCommonReq 结构体

**选择**：新增 `SysCreateCommonReq` 结构体，内嵌 `CreateCommonReq` 并新增 `Applicant` 字段。

```go
type SysCreateCommonReq struct {
    CreateCommonReq `json:",inline"`
    Applicant       string `json:"applicant" validate:"required,min=1"`
}
```

**理由**：
- 内嵌复用现有字段（`Remark`），扩展性好
- `applicant` 使用 `required,min=1` 校验，确保非空
- 解码后可将 `&req.CreateCommonReq` 传递给 `create()` 方法，`req.Applicant` 作为显式参数传递

### Decision 4: 路由注册在 init.go 中使用 /system/ 前缀

**选择**：在 `init.go` 中注册新路由 `POST /vendors/{vendor}/system/applications/types/create_load_balancer`。

```go
h.Add("SysCreateForCreateLB", "POST",
    "/vendors/{vendor}/system/applications/types/create_load_balancer", svc.SysCreateForCreateLB)
```

**理由**：
- 与原路由通过 `/system/` 前缀区分，语义清晰
- 后续其他 application 类型的系统提单可复用此路径结构（如 `/vendors/{vendor}/system/applications/types/create_cvm`）

### Decision 5: 新 Handler 方法放在 create.go 中

**选择**：`SysCreateForCreateLB` 方法放在 `cmd/cloud-server/service/application/create.go` 中，紧跟 `CreateForCreateLB`。

**理由**：
- 与原方法逻辑高度相似，放在同一文件便于对照维护
- 不新建文件，保持目录结构简洁

## Risks / Trade-offs

- **[Trade-off] 修改已有方法签名** → `create()`、`createItsmTicket()`、`createApplication()` 增加 `applicant` 参数，已有 7 处调用方需同步添加 `cts.Kit.User` 参数。Mitigation：均为包内私有方法，改动机械且编译器会捕获遗漏。
- **[Risk] 系统调用方伪造 applicant** → 任何能访问此接口的调用方都可以为任意用户提单。Mitigation：由上层网关通过 AppCode 白名单等机制限制可调用方，接口本身不做额外校验。

## Open Questions

无。所有关键决策已在与需求方确认中达成一致。
