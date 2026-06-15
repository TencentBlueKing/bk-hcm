## MODIFIED Requirements

### Requirement: AwsWithAssumeRole 编排方法（支持 Role Chain）
`CloudAdaptorClient` SHALL 提供 `AwsWithAssumeRole` 方法，接收 `kt`、`rootAccountID`（根账号 HCM 内部 ID）、`cloudID`（成员账号 AWS Account ID，从 main_account 表获取）和 `roleChain`（角色名数组）参数，返回具备成员账号访问权限的 `*aws.Aws` client。该方法内部 SHALL 用 `rootAccountID` 调 `AwsRoot()` 获取根账号 AK/SK，然后按 roleChain 顺序执行链式 AssumeRole。`roleChain[0..n-2]` 的角色在管理账号中 AssumeRole，`roleChain[n-1]` 在目标成员账号（cloudID）中 AssumeRole。该能力 SHALL 同时支持 EC2、CloudWatch 和 SageMaker 透传场景，并保持 `Aws()` 和 `AwsRoot()` 的既有行为不变。仅涉及 AWS 云厂商。

#### Scenario: 单步 AssumeRole（roleChain 长度为 1）
- **GIVEN** rootAccountID 有效，cloudID 为目标成员账号 AWS Account ID，roleChain 为 `["gpu-readonly"]`
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 系统依次执行：AwsRoot 获取 AK/SK → 用成员账号 CloudID 拼接 Role ARN → AssumeRole → 构建并返回 Aws client

#### Scenario: 多步 Role Chain（roleChain 长度为 2）
- **GIVEN** rootAccountID 有效，cloudID 为目标成员账号 AWS Account ID，roleChain 为 `["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"]`
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 系统执行：AwsRoot 获取 AK/SK → 用**管理账号** CloudID 拼接第一个 Role ARN 并 AssumeRole → 用第一步的临时凭证 + **成员账号** CloudID 拼接第二个 Role ARN 并 AssumeRole → 构建并返回 Aws client

#### Scenario: SageMaker 透传复用 AssumeRole 链路
- **GIVEN** SageMaker notebook、endpoint、training job、processing job 或 transform job 透传接口收到有效的 `root_account_id`、`main_account_id`、`role_chain` 和 `region`
- **WHEN** hc-service 处理该请求
- **THEN** 系统通过 `main_account_id` 解析目标成员账号 CloudID，并复用 `AwsWithAssumeRole` 构建 SageMaker client 后访问目标成员账号资源

#### Scenario: rootAccountID 无效
- **GIVEN** rootAccountID 在 root_account 表中不存在
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 返回错误，错误信息包含 rootAccountID

#### Scenario: Role Chain 中间步骤失败
- **GIVEN** roleChain 中某个中间角色不存在或权限不足
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 返回 AssumeRole 失败错误，包含失败的 Role ARN
