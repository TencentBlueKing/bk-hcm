## ADDED Requirements

### Requirement: Table 结构体定义
系统 SHALL 在 `pkg/dal/table/resource-plan/res-plan-demand-gpu-order/` 目录下定义 `ResPlanDemandGpuOrderTable` 结构体，字段与 `res_plan_demand_gpu_order` 表一致，包含 `id`、`bk_biz_id`、`op_product_id`、`op_product_name`（string，对应数据库 varchar(64)，建表 SQL 中的 bigint 为笔误）、`template_id`、`status`、`remark`、`creator`、`reviser`、`created_at`、`updated_at`。结构体 SHALL 实现 `TableName()`、`InsertValidate()`、`UpdateValidate()` 方法。

#### Scenario: 插入校验
- **WHEN** 调用 `InsertValidate()` 时，`id`、`bk_biz_id`、`op_product_id`、`template_id`、`status`、`creator` 为空或非法
- **THEN** 返回对应错误信息

#### Scenario: 更新校验
- **WHEN** 调用 `UpdateValidate()` 时，`creator` 字段不为空，或 `reviser` 为空
- **THEN** 返回错误（creator 不可更新，reviser 不可为空）

### Requirement: DAO 接口与实现
系统 SHALL 在 `pkg/dal/dao/resource-plan/res_plan_demand_gpu_order.go` 中定义 `ResPlanDemandGpuOrderInterface` 接口，包含以下方法：
- `CreateWithTx(kt, tx, models) ([]string, error)`
- `UpdateWithTx(kt, tx, expr, model) error`
- `List(kt, opt) (*ResPlanDemandGpuOrderListResult, error)`
- `DeleteWithTx(kt, tx, expr) error`

并提供对应的 `ResPlanDemandGpuOrderDao` 结构体实现。

#### Scenario: 批量创建
- **WHEN** 调用 `CreateWithTx` 传入非空 models
- **THEN** 批量写入数据库并返回生成的 ID 列表

#### Scenario: 分页查询
- **WHEN** 调用 `List` 传入合法的 `ListOption`（含 filter、page、fields）
- **THEN** 返回满足条件的记录列表及总数

#### Scenario: 批量删除
- **WHEN** 调用 `DeleteWithTx` 传入合法的 filter 表达式
- **THEN** 删除满足条件的所有记录

### Requirement: DAO Set 注册
系统 SHALL 在 `pkg/dal/dao/dao.go` 的 `Set` 接口中添加 `ResPlanDemandGpuOrder() resplan.ResPlanDemandGpuOrderInterface` 方法，并在 `set` 结构体上实现该方法。

#### Scenario: DAO 可通过 Set 访问
- **WHEN** 通过 `dao.Set.ResPlanDemandGpuOrder()` 调用
- **THEN** 返回可用的 `ResPlanDemandGpuOrderInterface` 实例

### Requirement: 表名常量注册
系统 SHALL 在 `pkg/dal/table/table.go` 中添加 `ResPlanDemandGpuOrderTable Name = "res_plan_demand_gpu_order"` 常量，并在 `SupportedTables` 中注册。

#### Scenario: 表名有效
- **WHEN** DAO 使用该常量调用 ID 生成器
- **THEN** ID 生成器能正确识别该表名

### Requirement: API 协议层定义
系统 SHALL 在 `pkg/api/data-service/resource-plan/res_plan_demand_gpu_order.go` 中定义以下结构体（均须实现 `Validate()` 方法）：
- `ResPlanDemandGpuOrderBatchCreateReq`（包含 `Items []ResPlanDemandGpuOrderCreateReq`）
- `ResPlanDemandGpuOrderCreateReq`（字段对应表字段，id 除外）
- `ResPlanDemandGpuOrderBatchUpdateReq`（包含 `Items []ResPlanDemandGpuOrderUpdateReq`）
- `ResPlanDemandGpuOrderUpdateReq`（含 `id`，其余字段可选）
- `ResPlanDemandGpuOrderListReq`（内嵌 `core.ListReq`）
- `ResPlanDemandGpuOrderListResult`（alias `types.ListResult[table]`）

#### Scenario: 批量创建请求校验
- **WHEN** `Items` 为空或超过 100 条时调用 `Validate()`
- **THEN** 返回参数错误

### Requirement: status 枚举类型定义
系统 SHALL 在 `pkg/criteria/enumor/` 中定义 `ResPlanDemandGpuOrderStatus` 类型，枚举值为：`Init`（INIT）、`Pending`（PENDING）、`Done`（DONE）、`Reject`（REJECT）、`RejectAll`（REJECT_ALL）、`Terminate`（TERMINATE），并实现 `Validate()` 方法。

#### Scenario: 合法状态校验
- **WHEN** 传入合法枚举值调用 `Validate()`
- **THEN** 返回 nil

#### Scenario: 非法状态校验
- **WHEN** 传入不在枚举列表中的字符串调用 `Validate()`
- **THEN** 返回错误

### Requirement: data-service service 层接口
系统 SHALL 在 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-order/` 目录下提供以下 HTTP 接口：
- `POST /res_plans/res_plan_demand_gpu_orders/batch/create` → `BatchCreate`
- `PATCH /res_plans/res_plan_demand_gpu_orders/batch` → `BatchUpdate`
- `POST /res_plans/res_plan_demand_gpu_orders/list` → `List`
- `DELETE /res_plans/res_plan_demand_gpu_orders/batch` → `BatchDelete`

#### Scenario: 批量创建成功
- **WHEN** 请求体合法，调用 `BatchCreate`
- **THEN** 在事务中写入数据库，返回创建的 ID 列表

#### Scenario: 批量更新成功
- **WHEN** 请求体中包含合法的 `id` 列表和更新字段，调用 `BatchUpdate`
- **THEN** 在事务中更新对应记录，返回成功

#### Scenario: 分页查询成功
- **WHEN** 传入合法的 filter 和 page 参数，调用 `List`
- **THEN** 返回满足条件的记录列表

#### Scenario: 批量删除成功
- **WHEN** 传入合法的 ID 列表，调用 `BatchDelete`
- **THEN** 在事务中删除对应记录，返回成功

### Requirement: resource-plan 服务注册
系统 SHALL 在 `cmd/data-service/service/resource-plan/resource_plan.go` 的 `InitService` 函数中调用新服务包的 `InitService`，完成路由注册。

#### Scenario: 服务启动时路由可访问
- **WHEN** data-service 启动后向 `/res_plans/res_plan_demand_gpu_orders/list` 发起请求
- **THEN** 请求被正确路由到 `List` 处理函数
