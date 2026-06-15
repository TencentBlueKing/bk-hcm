## Context

权限模板数据层（`permission-template-crud`）已完整实现，包含 DAO、DataService CRUD、Client 封装及 hc-service CAM Policy 接口。业务层（cloud-server）已有 `apply_permission_policy_library` 的审批流实现，通过 ITSM 进行权限策略库的应用操作。

当前缺失：用户无法通过业务侧 API 自主发起"创建云权限模板"的申请——即以自定义名称，基于策略库内容为指定账号创建 CAM 策略并记录本地模板。

设计约束：
- 必须走 ITSM 审批流（与现有 apply_permission_policy_library 行为一致）
- 仅支持 TCloud vendor（当前策略库和 CAM Policy 均只有 TCloud 实现）
- 单账号单次申请，返回单个审批单 ID
- 同一账号对同一策略库不能重复创建（审批前和 deliver 前均需校验）

## Goals / Non-Goals

**Goals:**
- 实现 `POST /bizs/{bk_biz_id}/vendors/{vendor}/applications/types/create_permission_template` 接口
- 建立独立的 `permission-template` handler 体系（base.go + create action），为后续 update/delete 预留扩展点
- 修改 `applier.go` 的 CAM Policy 创建/本地模板创建方法，支持自定义 `name` 和 `memo` 参数
- `applier.go` 现有调用方（tcloudApplyCreateForAccount）行为不变（传入 library.Name/Memo）

**Non-Goals:**
- 不实现 update/delete permission template 的审批流（本次仅 create）
- 不实现多账号批量申请（单次申请对应单账号）
- 不实现非 TCloud vendor 的创建

## Decisions

### 决策1：独立 handler 体系，不复用 permission-policy-library handler

**选择**：新建 `handlers/permission-template/` 目录，包含 `base.go` + `create/` 子目录，不在现有 `permission-policy-library/` handler 下扩展。

**理由**：`create_permission_template` 的 Content 结构含有 `name`/`memo` 字段，与 `ApplyPermPolicyLibContent` 不同。强行共用 Content 结构会引入可选字段，增加理解成本。独立体系与后续 update/delete 扩展自然对齐。

**备选**：在 `permission-policy-library/base.go` 的 action 注册表中加入新 action 类型。此方案导致 Content 结构臃肿，且权限模板操作与策略库操作的语义在未来可能进一步分化。

### 决策2：使用 `operate_permission_template` 作为 ApplicationType

**选择**：单个 ApplicationType 覆盖 create/update/delete 三类操作，通过 Content 内的 `Action` 字段区分。

**理由**：与 `apply_permission_policy_library` 的模式一致，在 `approve.go` 的 `getHandlerByApplication` 中只需一个 case，后续新增操作无需修改该 switch。

**备选**：每种操作独立 ApplicationType（如 `create_permission_template`、`update_permission_template`）。该方案导致 `approve.go` 随操作增多线性膨胀，不推荐。

### 决策3：修改 applier.go 函数签名，而非新增方法

**选择**：为 `TCloudCreateCAMPolicy` 和 `TCloudCreateLocalTemplate` 添加 `name string, memo *string` 入参。现有唯一调用方（`tcloudApplyCreateForAccount`，同文件内）改为传入 `library.Name, library.Memo`，行为不变。

**理由**：两个方法均为 `applier.go` 内部方法，无外部调用者，修改安全。新增方法会引入冗余，且未来扩展时需同步维护两套。

### 决策4：base.go 中使用 action 注册表模式

**选择**：`permission-template/base.go` 提供 `RegisterActionHandler` / `NewHandlerFromApplication` 注册表，`create/init.go` 通过 `init()` 注册。

**理由**：与现有 `permission-policy-library/base.go` 模式完全一致。`approve.go` 只需在 `getHandlerByApplication` 中加一个 case 调用 `NewHandlerFromApplication`，后续 update/delete handler 加入时 `approve.go` 无需任何修改。

## Risks / Trade-offs

- **风险：审批期间重复提交**：用户可能在审批中提交第二次创建申请（对同一账号同一库）。`CheckReq` 在创建时和 deliver 时均会执行 `CheckAccountApplied`，但审批中状态无法阻止第二次提交。→ 缓解：deliver 前的 `CheckReq` 是最终防线，若先完成交付则后单 deliver 失败并标记 DeliverError，保持数据一致性。

- **风险：CAM Policy 创建成功但本地模板写入失败**：与 `tcloudApplyCreateForAccount` 中的现有风险相同，`applier.go` 已有注释说明。→ 缓解：返回含 cloudPolicyID 的错误信息，便于人工补录。

- **Trade-off：单账号申请 vs 多账号批量**：本次设计为单账号，与 `apply_permission_policy_library`（多账号）不同。原因是创建模板场景中 `name` 字段是全局的，批量申请时不同账号使用同一 name 无意义；未来若有需求可在同一 ApplicationType 下扩展。
