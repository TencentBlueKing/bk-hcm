## ADDED Requirements

### Requirement: 业务下批量终止主单
接口路径：`POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/demands/orders/batch/terminate`
权限：业务访问（ListAuthorizedBiz 校验 bk_biz_id）
允许前置状态：`INIT`（待评审）或 `REJECT_ALL`（全部已驳回）
操作：主单状态 → `TERMINATE`，关联子单状态 → `TERMINATE`
单次最多 100 条，需记录审计

#### Scenario: 正常终止
- **WHEN** 用户提交合法的 order_ids（1~100 条），且所有主单均处于 INIT 或 REJECT_ALL 状态
- **THEN** 所有主单及其关联子单状态变更为 TERMINATE，返回 `{"code":0,"data":null}`

#### Scenario: 状态不合法
- **WHEN** 其中任意一条主单的状态不是 INIT 或 REJECT_ALL（如 PENDING）
- **THEN** 返回参数错误，所有主单不做任何变更

#### Scenario: 业务权限不足
- **WHEN** 当前用户对该 bk_biz_id 没有业务访问权限
- **THEN** 返回鉴权失败错误

---

### Requirement: 资源下批量改"评审中"状态
接口路径：`POST /api/v1/woa/plans/resources/gpu/demands/orders/batch/pending`
权限：平台-GPU需求（ZiYanResPlanGPUDemands + Update）
允许前置状态：`INIT`（待评审）
操作：主单状态 → `PENDING`，关联子单状态 → `PENDING`
单次最多 100 条，需记录审计

#### Scenario: 正常改评审中
- **WHEN** 用户提交合法的 order_ids（1~100 条），且所有主单均处于 INIT 状态
- **THEN** 所有主单及其关联子单状态变更为 PENDING，返回 `{"code":0,"data":null}`

#### Scenario: 状态不合法
- **WHEN** 其中任意一条主单的状态不是 INIT
- **THEN** 返回参数错误，所有主单不做任何变更

#### Scenario: 平台权限不足
- **WHEN** 当前用户没有平台-GPU需求权限
- **THEN** 返回鉴权失败错误

---

### Requirement: 资源下批量驳回主单
接口路径：`POST /api/v1/woa/plans/resources/gpu/demands/orders/batch/reject`
权限：平台-GPU需求（ZiYanResPlanGPUDemands + Update）
允许前置状态：`PENDING`（评审中）
操作：主单状态 → `REJECT_ALL`，关联子单状态 → `REJECT`
单次最多 100 条，需记录审计

#### Scenario: 正常驳回
- **WHEN** 用户提交合法的 order_ids（1~100 条），且所有主单均处于 PENDING 状态
- **THEN** 所有主单状态变更为 REJECT_ALL，所有关联子单状态变更为 REJECT，返回 `{"code":0,"data":null}`

#### Scenario: 状态不合法
- **WHEN** 其中任意一条主单的状态不是 PENDING
- **THEN** 返回参数错误，所有主单不做任何变更

---

### Requirement: 资源下批量终止主单
接口路径：`POST /api/v1/woa/plans/resources/gpu/demands/orders/batch/terminate`
权限：平台-GPU需求（ZiYanResPlanGPUDemands + Update）
允许前置状态：`PENDING`（评审中）
操作：主单状态 → `TERMINATE`，关联子单状态 → `TERMINATE`
单次最多 100 条，需记录审计

#### Scenario: 正常终止
- **WHEN** 用户提交合法的 order_ids（1~100 条），且所有主单均处于 PENDING 状态
- **THEN** 所有主单及其关联子单状态变更为 TERMINATE，返回 `{"code":0,"data":null}`

#### Scenario: 状态不合法
- **WHEN** 其中任意一条主单的状态不是 PENDING
- **THEN** 返回参数错误，所有主单不做任何变更

---

### Requirement: 公共状态变更逻辑
私有方法 `batchUpdateGpuOrderStatus(kt *kit.Kit, orderIDs []string, allowedFromStatuses, targetOrderStatus, targetSubOrderStatus)`

#### Scenario: 完整执行流程
- **WHEN** 调用方传入合法参数
- **THEN** 按顺序执行：查主单→校验前置状态→写审计（主单粒度）→更新主单状态→查子单→分批（100条/批）更新子单状态

#### Scenario: 子单分批更新
- **WHEN** 关联子单总数超过 100 条
- **THEN** 使用 `slice.Split` 将子单列表分批，每批不超过 100 条，逐批调用 BatchUpdateResPlanDemandGpuSubOrder
