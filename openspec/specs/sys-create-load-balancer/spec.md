# sys-create-load-balancer

系统提单创建 CLB 申领单的能力，允许系统调用方为任意指定用户提交负载均衡申领工单。

## Purpose

TBD

## Requirements

### Requirement: 系统提单接口路由

系统 SHALL 提供独立的系统提单路由 `POST /vendors/{vendor}/system/applications/types/create_load_balancer`，与现有用户提单路由分离。路由 MUST 注册在 cloud-server 的 application 模块下，复用 `/system/` 路径前缀以便后续扩展其他类型的系统提单接口。

#### Scenario: 正常访问系统提单路由
- **WHEN** 系统调用方发送 `POST /api/v1/cloud/vendors/tcloud/system/applications/types/create_load_balancer` 请求
- **THEN** 系统 SHALL 路由到系统提单 Handler 进行处理

#### Scenario: 不支持的云厂商
- **WHEN** 系统调用方发送请求且 `vendor` 路径参数不是 `tcloud` 或 `tcloud_ziyan`
- **THEN** 系统 SHALL 返回 `InvalidParameter` 错误

### Requirement: 请求体包含 applicant 字段

系统提单接口的请求体 MUST 包含 `applicant` 字段（字符串类型），用于指定实际提单人。该字段 MUST 为必填且不能为空字符串。其余请求参数（`remark`、CLB 配置参数等）与原接口保持一致。

#### Scenario: 正常传递 applicant
- **WHEN** 请求体中 `applicant` 为合法的用户标识（如 `"zhangsan"`）
- **THEN** 系统 SHALL 接受请求并使用该值作为提单人

#### Scenario: 缺少 applicant 字段
- **WHEN** 请求体中未包含 `applicant` 字段
- **THEN** 系统 SHALL 返回 `InvalidParameter` 错误，提示 applicant 为必填字段

#### Scenario: applicant 为空字符串
- **WHEN** 请求体中 `applicant` 为空字符串 `""`
- **THEN** 系统 SHALL 返回 `InvalidParameter` 错误

### Requirement: 申请单使用指定提单人

创建申请单（Application）时，`Applicant` 字段 MUST 使用请求体传入的 `applicant` 值，而非 `cts.Kit.User`。申请单的其他字段（`SN`、`Source`、`Type`、`Status`、`BkBizIDs`、`Content`、`DeliveryDetail`、`Memo`）的赋值逻辑 SHALL 与原接口保持一致。

#### Scenario: 申请单记录正确的提单人
- **WHEN** 系统用户 A 调用系统提单接口，`applicant` 指定为用户 B
- **THEN** 创建的申请单记录中 `Applicant` SHALL 为用户 B
- **THEN** 用户 B 能通过查看接口查看该申请单（无需额外权限）
- **THEN** 用户 B 能通过取消接口取消该申请单

### Requirement: ITSM 工单使用指定提单人

调用 ITSM 创建工单时，`Creator` 字段 MUST 使用请求体传入的 `applicant` 值，而非 `cts.Kit.User`。ITSM 工单的其他参数（`ServiceID`、`CallbackURL`、`Title`、`ContentDisplay`、`VariableApprovers`）的生成逻辑 SHALL 与原接口保持一致。

#### Scenario: ITSM 工单创建人为指定提单人
- **WHEN** 系统用户 A 调用系统提单接口，`applicant` 指定为用户 B
- **THEN** ITSM 工单的 `Creator` SHALL 为用户 B
- **THEN** 用户 B 能在 ITSM 中看到自己创建的工单

### Requirement: 鉴权与权限检查

系统提单接口 SHALL 沿用原接口的权限检查逻辑（`checkApplyResPermission`），验证负载均衡申请资源的权限。权限检查的对象是系统调用方（`cts.Kit.User`），而非请求体中的 `applicant`。applicant 不要求拥有申请资源的权限。上层网关负责对系统调用方进行额外的访问拦截。

#### Scenario: 权限检查通过
- **WHEN** 系统调用方具有负载均衡申请资源权限
- **THEN** 系统 SHALL 继续执行提单流程

#### Scenario: 权限检查不通过
- **WHEN** 系统调用方不具有负载均衡申请资源权限
- **THEN** 系统 SHALL 返回 `PermissionDenied` 错误

#### Scenario: applicant 无权限不影响提单
- **WHEN** 系统调用方具有权限，但 `applicant` 指定的用户不具有负载均衡申请资源权限
- **THEN** 系统 SHALL 正常创建申请单，不因 applicant 无权限而拒绝

### Requirement: 支持 TCloud 和 TCloudZiyan 厂商

系统提单接口 MUST 支持与原接口相同的云厂商（TCloud 和 TCloudZiyan）。每个厂商的 CLB 创建参数校验、预处理、ITSM 表单渲染等逻辑 SHALL 复用现有的 Handler 实现（`ApplicationOfCreateTCloudLB`、`ApplicationOfCreateZiyanLB`）。

#### Scenario: TCloud 厂商提单
- **WHEN** `vendor` 路径参数为 `tcloud`，请求体包含合法的 `TCloudLoadBalancerCreateReq` 参数
- **THEN** 系统 SHALL 使用 TCloud Handler 处理并创建申请单

#### Scenario: TCloudZiyan 厂商提单
- **WHEN** `vendor` 路径参数为 `tcloud_ziyan`，请求体包含合法的 `TCloudZiyanLoadBalancerCreateReq` 参数
- **THEN** 系统 SHALL 使用 TCloudZiyan Handler 处理并创建申请单
