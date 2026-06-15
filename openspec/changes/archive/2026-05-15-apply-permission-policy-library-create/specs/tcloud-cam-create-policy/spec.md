## ADDED Requirements

### Requirement: TCloud CAM CreatePolicy adaptor 方法

系统 SHALL 在 `pkg/adaptor/tcloud/` 中新增 `CreatePolicy` 方法，封装腾讯云 CAM `cam:CreatePolicy` API 调用。方法接收 `TCloudCreatePolicyOption`，返回 `TCloudCreatePolicyResult`（含云上策略 ID）。

`TCloudCreatePolicyOption` 和 `TCloudCreatePolicyResult` 类型定义在 `pkg/adaptor/types/account/` 包：
- `TCloudCreatePolicyOption` 字段：`Region`（string）、`PolicyName`（string，required）、`PolicyDocument`（string，required）、`Description`（string，可选）
- `TCloudCreatePolicyResult` 字段：`PolicyID`（uint64，云上策略 ID）

方法签名：`CreatePolicy(kt *kit.Kit, opt *account.TCloudCreatePolicyOption) (*account.TCloudCreatePolicyResult, error)`

#### Scenario: 成功创建 CAM 策略

- **WHEN** 传入有效的 `PolicyName` 和合法的 `PolicyDocument` JSON 字符串
- **THEN** 系统 SHALL 调用 `cam.NewCreatePolicyRequest()` 设置 `PolicyName`、`PolicyDocument`、`Description`（非空时设置），通过 `CamServiceClient` 执行请求，返回 `TCloudCreatePolicyResult{PolicyID: *resp.Response.PolicyId}`。`Region` 为空时使用 `constant.TCloudDefaultRegion`

#### Scenario: opt 为 nil

- **WHEN** `opt` 为 nil
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: PolicyName 为空

- **WHEN** `PolicyName` 为空字符串
- **THEN** 系统 SHALL 在 Validate 阶段返回 InvalidParameter 错误

#### Scenario: PolicyDocument 为空

- **WHEN** `PolicyDocument` 为空字符串
- **THEN** 系统 SHALL 在 Validate 阶段返回 InvalidParameter 错误

#### Scenario: 云 API 返回错误

- **WHEN** 腾讯云 CAM API 返回错误（如策略名重复、配额超限）
- **THEN** 系统 SHALL 透传云 API 错误信息

### Requirement: TCloud 接口扩展

系统 SHALL 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `CreatePolicy` 方法声明。

#### Scenario: 接口兼容

- **WHEN** 外部调用 `tcloud.TCloud` 接口
- **THEN** 接口 SHALL 包含 `CreatePolicy(kt *kit.Kit, opt *account.TCloudCreatePolicyOption) (*account.TCloudCreatePolicyResult, error)` 方法

### Requirement: hc-service 暴露 CAM 策略创建接口

系统 SHALL 在 `cmd/hc-service/service/permission-template/` 中新增 handler，暴露 `POST /vendors/tcloud/permission_templates/cam/create_policy` 接口。

请求模型 `CreateCAMPolicyReq`（`pkg/api/hc-service/permission-template/`）字段：`account_id`（string，必填）、`policy_name`（string，必填）、`policy_document`（string，必填）、`description`（string，可选）。

响应模型 `CreateCAMPolicyResult` 字段：`policy_id`（uint64，云上策略 ID）。

handler 通过 `svc.ad.TCloud(kt, req.AccountID)` 获取带密钥的 adaptor 实例，构造 `TCloudCreatePolicyOption` 调用 adaptor，将结果的 `PolicyID` 映射到响应的 `PolicyID` 返回。

#### Scenario: 成功创建

- **WHEN** 传入有效的 `account_id` 和策略参数
- **THEN** 系统 SHALL 使用该 account_id 获取 TCloud adaptor，调用 `CreatePolicy`，返回 `CreateCAMPolicyResult{PolicyID: result.PolicyID}`

#### Scenario: 账号密钥获取失败

- **WHEN** `account_id` 对应的账号不存在或密钥为空
- **THEN** 系统 SHALL 返回错误

#### Scenario: 云 API 调用失败

- **WHEN** CAM CreatePolicy API 返回错误
- **THEN** 系统 SHALL 返回包含云 API 错误信息的错误

### Requirement: hc-service client 封装

系统 SHALL 在 `pkg/client/hc-service/tcloud/` 中新增 `PermissionTemplateClient`，提供 `CreateCAMPolicy` 方法，供 cloud-server 调用 hc-service 的 CAM 策略创建接口。

方法签名：`CreateCAMPolicy(kt *kit.Kit, req *proto.CreateCAMPolicyReq) (*proto.CreateCAMPolicyResult, error)`

client 使用相对路径 `/permission_templates/cam/create_policy` 发起 POST 请求（vendor 前缀由 client 基础 URL 配置注入）。

#### Scenario: cloud-server 通过 client 调用

- **WHEN** cloud-server 调用 `svc.client.HCService().TCloud.PermissionTemplate.CreateCAMPolicy(kt, req)`
- **THEN** 请求 SHALL 正确路由到 hc-service 的 `POST /vendors/tcloud/permission_templates/cam/create_policy`
