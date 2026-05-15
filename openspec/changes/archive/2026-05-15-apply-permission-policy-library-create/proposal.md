## Why

资源运营管理员需要将权限策略库"应用"到指定二级账号，即在云上创建 CAM 策略并在本地创建对应的 `permission_template` 记录。当前系统已有权限策略库和权限模板的 CRUD 能力，但缺少将两者串联的"应用"操作——调用云 API 创建策略、写入本地模板、记录审计。此外，该功能的核心逻辑需要为后续的 Biz 层审批接口和"应用更新"接口提供复用基础。

## What Changes

- 新增 TCloud CAM `CreatePolicy` adaptor 封装（`pkg/adaptor/types/account/` 类型定义 + `pkg/adaptor/tcloud/` 方法实现）
- 新增 hc-service 层 CAM 策略创建接口（`POST /vendors/tcloud/permission_templates/cam/create_policy`），供 cloud-server 调用
- 新增 hc-service client 方法（`pkg/client/hc-service/tcloud/`）
- 新增 `Apply` 审计 Action 枚举（`pkg/criteria/enumor/audit.go`）及 `ApplyOp` 操作常量（`pkg/api/data-service/audit/audit.go`）
- 新增 cloud-server 层 `ApplyPermissionPolicyLibraryCreate` handler（`POST /vendors/{vendor}/permission_policy_libraries/{id}/apply`）
- 新增 cloud-server 层 `applier.go` 公共逻辑，为后续 apply_update 和 biz 层接口提供复用
- 新增请求/响应 API Model（`pkg/api/cloud-server/`）
- 新增审计 build 逻辑（data-service audit），审计记录关联被应用的账号信息

## Capabilities

### New Capabilities
- `tcloud-cam-create-policy`: TCloud CAM CreatePolicy API adaptor 封装及 hc-service 暴露
- `apply-permission-policy-library-create`: cloud-server 层"应用权限策略库（创建）"接口，含公共 applier 逻辑、业务范围校验、审计集成

### Modified Capabilities
- `permission-policy-library-crud`: 新增 `Apply` 审计 Action 枚举，扩展审计 build 逻辑（含关联账号信息）

## Impact

- **新增代码**：`pkg/adaptor/types/account/` 类型定义、`pkg/adaptor/tcloud/` CAM 策略方法、`cmd/hc-service/service/permission-template/` CAM 策略 handler、`pkg/client/hc-service/tcloud/` client、`cmd/cloud-server/service/permission-policy-library/` applier + apply handler、`pkg/api/cloud-server/` 模型
- **修改代码**：`pkg/criteria/enumor/audit.go` 新增枚举、`pkg/api/data-service/audit/audit.go` 新增 ApplyOp 常量、`cmd/data-service/service/audit/cloud/` 审计 build 扩展、`pkg/adaptor/tcloud/interface.go` 接口扩展
- **依赖**：腾讯云 CAM SDK（`cam.CreatePolicy`），已在项目中引入
- **受影响系统**：hc-service（新增 API）、cloud-server（新增路由）、data-service（审计扩展）
