## Why

现有 CLB 申领提单接口（`/api/v1/cloud/vendors/{vendor}/applications/types/create_load_balancer`）的提单人固定从 `cts.Kit.User`（HTTP 请求头 `X-Bkapi-User-Name`）获取，不允许显式传递，以防止身份伪造。但在系统间调用场景下，调用方的 `Kit.User` 是系统虚拟用户（如 `bk_system`），而系统期望的提单人是实际的自然人用户。当前架构无法满足"系统代替指定用户提单"的需求。

## What Changes

- **新增系统提单路由**：新增 `POST /vendors/{vendor}/system/applications/types/create_load_balancer`，复用 `/system/` 路径前缀，后续所有 application 类型的系统提单接口可沿用此路径结构
- **请求体新增 `applicant` 字段**：系统提单接口允许调用方在请求体中显式指定提单人
- **提单人赋值逻辑变更**：系统提单接口中，创建申请单（`Applicant`）和 ITSM 工单（`Creator`）使用请求体传入的 `applicant`，而非 `cts.Kit.User`
- 除 `applicant` 赋值外，其余逻辑（参数校验、权限检查、ITSM 创建、申请单持久化）与原接口保持一致
- 支持的云厂商与原接口相同（TCloud、TCloudZiyan）
- 鉴权沿用原接口机制，上层网关额外进行系统调用方的访问拦截

## Capabilities

### New Capabilities
- `sys-create-load-balancer`: 系统提单创建 CLB 申领单的能力，允许系统调用方为任意指定用户提交负载均衡申领工单

### Modified Capabilities
<!-- 无需修改已有 capability，新接口独立于原接口 -->

## Impact

- **API 层**：cloud-server 新增一条路由注册，新增对应 Handler 方法
- **Proto 层**：新增系统提单请求结构体（包含 `applicant` 字段），或扩展现有 `CreateCommonReq`
- **接口文档**：需新增系统提单接口的 API 文档
- **影响范围有限**：不修改原接口逻辑，查看/取消/列表操作无需变更（`Applicant` 为实际提单人，权限校验行为符合预期）
- **无 DB 变更**：复用现有 `application` 表结构
