# Capability: res-plan-demand-gpu-order-list

## Purpose

提供 GPU 需求提报主单的列表查询能力，支持资源视角（SCR）和业务视角（BIZ）两种查询模式，并在非 count 模式下附加子单聚合字段。

## Requirements

### Requirement: 资源视角 GPU 需求主单列表查询
系统 SHALL 提供资源视角（SCR）的 GPU 需求提报主单列表查询接口，接口路径为 `POST /api/v1/woa/plans/resources/gpu/demands/orders/list`，需持有平台-GPU需求（ZiYanResPlanGPUDemands Update）权限。

#### Scenario: 资源视角成功查询列表
- **WHEN** 拥有 ZiYanResPlanGPUDemands 权限的用户携带合法 filter + page 请求该接口
- **THEN** 系统返回主单列表，每条记录包含 id、bk_biz_id、op_product_id、op_product_name、template_id、status、remark、total_gpu_num、total_qpm_max、creator、reviser、created_at、updated_at

#### Scenario: 资源视角 count 模式
- **WHEN** 用户携带 `page.count=true` 请求该接口
- **THEN** 系统仅返回符合 filter 条件的主单总数，不返回 details，不查询子单

#### Scenario: 无权限访问
- **WHEN** 用户不持有 ZiYanResPlanGPUDemands 权限请求该接口
- **THEN** 系统返回 403 权限拒绝错误

### Requirement: 业务视角 GPU 需求主单列表查询
系统 SHALL 提供业务视角（BIZ）的 GPU 需求提报主单列表查询接口，接口路径为 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/demands/orders/list`，系统自动将路径参数 bk_biz_id 注入查询条件，确保用户只能查询自身有权访问的业务下的主单。

#### Scenario: 业务视角成功查询列表
- **WHEN** 用户对 bk_biz_id 有访问权限，携带合法 filter + page 请求该接口
- **THEN** 系统返回该业务下的主单列表，每条记录包含与资源视角相同的字段（含聚合字段）

#### Scenario: 业务视角越权访问
- **WHEN** 用户对路径中 bk_biz_id 无访问权限
- **THEN** 系统返回 403 权限拒绝错误，不返回任何主单数据

#### Scenario: 业务视角 count 模式
- **WHEN** 用户携带 `page.count=true` 请求业务视角接口
- **THEN** 系统仅返回该业务下符合 filter 条件的主单总数，不返回 details

### Requirement: 响应包含子单聚合字段
系统 SHALL 在列表查询响应（非 count 模式）中，为每条主单记录附加聚合字段：
- `total_gpu_num`：关联子单 gpu_num 字段的 SUM（不过滤子单状态）
- `total_qpm_max`：关联子单 qpm_max 字段的 SUM（不过滤子单状态）
若主单无关联子单，则两个字段值为 0。

#### Scenario: 主单有关联子单
- **WHEN** 主单 order-001 有 2 条子单，gpu_num 分别为 8 和 8，qpm_max 分别为 5000 和 5000
- **THEN** 响应中 order-001 的 total_gpu_num=16，total_qpm_max=10000

#### Scenario: 主单无关联子单
- **WHEN** 主单 order-002 没有任何关联子单
- **THEN** 响应中 order-002 的 total_gpu_num=0，total_qpm_max=0
