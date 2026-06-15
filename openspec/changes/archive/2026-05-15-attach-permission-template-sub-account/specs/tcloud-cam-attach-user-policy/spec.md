## ADDED Requirements

### Requirement: TCloud CAM AttachUserPolicy adaptor 方法

系统 SHALL 在 `pkg/adaptor/tcloud/cam_policy.go` 中新增 `AttachUserPolicy` 方法，封装腾讯云 CAM `cam:AttachUserPolicy` API 调用。

`TCloudAttachUserPolicyOption` 类型定义在 `pkg/adaptor/types/account/` 包：
- `TargetUin`: uint64，必填，目标子用户 UIN
- `PolicyId`: uint64，必填，云上策略 ID

方法签名：`AttachUserPolicy(kt *kit.Kit, opt *TCloudAttachUserPolicyOption) error`

方法 SHALL 支持 adaptor 已有的限流重试机制（`SetRateLimitRetryWithRandomInterval`）。

#### Scenario: 成功绑定策略

- **WHEN** 传入有效的 `TargetUin` 和 `PolicyId`
- **THEN** 系统 SHALL 调用 `cam.NewAttachUserPolicyRequest()` 设置参数，通过 `CamServiceClient` 执行请求，返回 nil

#### Scenario: opt 为 nil

- **WHEN** `opt` 为 nil
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: TargetUin 为 0

- **WHEN** `TargetUin` 为 0
- **THEN** 系统 SHALL 在 Validate 阶段返回 InvalidParameter 错误

#### Scenario: PolicyId 为 0

- **WHEN** `PolicyId` 为 0
- **THEN** 系统 SHALL 在 Validate 阶段返回 InvalidParameter 错误

#### Scenario: 云 API 返回限流错误

- **WHEN** 腾讯云 CAM API 返回 `RequestLimitExceeded` 错误
- **THEN** 系统 SHALL 自动进行指数退避重试，最大重试次数与 adaptor 全局配置一致

#### Scenario: 云 API 返回策略不存在错误

- **WHEN** 腾讯云 CAM API 返回策略不存在错误
- **THEN** 系统 SHALL 透传云 API 错误信息

### Requirement: TCloud 接口扩展

系统 SHALL 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `AttachUserPolicy` 方法声明。

#### Scenario: 接口兼容

- **WHEN** 外部调用 `tcloud.TCloud` 接口
- **THEN** 接口 SHALL 包含 `AttachUserPolicy(kt *kit.Kit, opt *account.TCloudAttachUserPolicyOption) error` 方法

### Requirement: hc-service 暴露批量绑定权限策略接口

系统 SHALL 在 `cmd/hc-service/service/sub-account/` 中新增 handler，暴露 `POST /vendors/tcloud/sub_accounts/attach_user_policies` 接口。

请求模型 `TCloudAttachUserPoliciesReq`（`pkg/api/hc-service/sub-account/`）字段：
- `account_id`: string，必填，用于获取凭证
- `target_uin`: uint64，必填，目标子用户 UIN
- `policy_ids`: []uint64，必填，云上策略 ID 列表

响应模型 `TCloudAttachUserPoliciesResult` 字段：
- `success_count`: uint64，成功绑定的策略数量
- `failed_policy_ids`: []uint64，绑定失败的策略 ID 列表

handler 通过 `svc.ad.TCloud(kt, req.AccountID)` 获取带密钥的 adaptor 实例，遍历 `policy_ids` 逐个调用 `AttachUserPolicy`，收集成功/失败结果。

#### Scenario: 全部绑定成功

- **WHEN** 传入有效的 `account_id`、`target_uin` 和 `policy_ids` 列表，所有策略绑定成功
- **THEN** 返回 `TCloudAttachUserPoliciesResult{SuccessCount: len(policy_ids), FailedPolicyIds: []}`

#### Scenario: 部分绑定失败

- **WHEN** 传入 `policy_ids` 列表，部分策略绑定失败
- **THEN** 返回成功数量和失败的策略 ID 列表，不返回错误（继续处理其他策略）

#### Scenario: 账号密钥获取失败

- **WHEN** `account_id` 对应的账号不存在或密钥为空
- **THEN** 系统 SHALL 返回错误

#### Scenario: policy_ids 为空

- **WHEN** `policy_ids` 为空数组
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: hc-service client 封装

系统 SHALL 在 `pkg/client/hc-service/tcloud/account.go` 中新增 `AttachUserPolicies` 方法。

方法签名：`AttachUserPolicies(kt *kit.Kit, req *TCloudAttachUserPoliciesReq) (*TCloudAttachUserPoliciesResult, error)`

client 使用相对路径 `/sub_accounts/attach_user_policies` 发起 POST 请求（vendor 前缀由 client 基础 URL 配置注入）。

#### Scenario: cloud-server 通过 client 调用

- **WHEN** cloud-server 调用 `svc.client.HCService().TCloud.Account.AttachUserPolicies(kt, req)`
- **THEN** 请求 SHALL 正确路由到 hc-service 的 `POST /vendors/tcloud/sub_accounts/attach_user_policies`
