## ADDED Requirements

### Requirement: AssumeRole STS 调用
系统 SHALL 通过 AWS STS `AssumeRole` API 获取指定成员账号的临时凭证（AccessKeyId、SecretAccessKey、SessionToken），临时凭证有效期为 1 小时（AWS 默认值）。仅涉及 AWS 云厂商。

#### Scenario: 成功获取临时凭证
- **GIVEN** 管理账号 IAM User 拥有 `sts:AssumeRole` 权限
- **WHEN** 以 `root_account` 的 AK/SK 调用 AssumeRole，传入 Role ARN 和 SessionName
- **THEN** 返回包含 AccessKeyId、SecretAccessKey、SessionToken 和 Expiration 的临时凭证

#### Scenario: AssumeRole 失败
- **GIVEN** 目标成员账号未创建对应 IAM Role 或 Trust Policy 不允许
- **WHEN** 调用 AssumeRole
- **THEN** 返回错误，错误信息中包含 Role ARN 和原始 AWS 错误描述

### Requirement: Role ARN 自动拼接
系统 SHALL 根据下游平台透传的 `role_name` + 目标账号 CloudID（成员账号 AWS Account ID，从 `main_account` 表获取）+ `site`（国际站/中国站）自动拼接完整 Role ARN。不依赖数据库存储 Role ARN。仅涉及 AWS 云厂商。

#### Scenario: 国际站 ARN 拼接
- **GIVEN** site 为国际站，目标账号 CloudID 为 `123456789012`，role_name 为 `gpu-readonly`
- **WHEN** 系统拼接 Role ARN
- **THEN** 生成 `arn:aws:iam::123456789012:role/gpu-readonly`

#### Scenario: 中国站 ARN 拼接
- **GIVEN** site 为中国站，目标账号 CloudID 为 `123456789012`，role_name 为 `gpu-readonly`
- **WHEN** 系统拼接 Role ARN
- **THEN** 生成 `arn:aws-cn:iam::123456789012:role/gpu-readonly`

### Requirement: 临时凭证进程内缓存
系统 SHALL 缓存 AssumeRole 获取的临时凭证，缓存 key 为 `cloudAccountID + ":" + roleArn`（由 `AwsWithAssumeRole` 内部构建后传入缓存模块）。凭证在过期前 10 分钟自动刷新。缓存存储于进程内存，不引入 Redis 等外部依赖。Role Chain 场景下每一步 AssumeRole 独立缓存。仅涉及 AWS 云厂商。

#### Scenario: 缓存命中
- **GIVEN** 缓存中存在未过期（距过期 > 10 分钟）的临时凭证
- **WHEN** 相同 key 再次请求
- **THEN** 直接返回缓存凭证，不调用 STS API

#### Scenario: 缓存提前刷新
- **GIVEN** 缓存凭证距过期不足 10 分钟
- **WHEN** 请求该 key 的凭证
- **THEN** 调用 STS API 获取新凭证，更新缓存并返回新凭证

#### Scenario: 刷新失败降级
- **GIVEN** 缓存凭证距过期不足 10 分钟，且 STS API 调用失败
- **WHEN** 请求该 key 的凭证
- **THEN** 返回旧缓存凭证（仍在有效期内），记录 WARN 级别日志

#### Scenario: 缓存不存在或已过期
- **GIVEN** 缓存中无该 key 或凭证已过期
- **WHEN** 请求凭证
- **THEN** 调用 STS API 获取新凭证，写入缓存并返回

### Requirement: AwsWithAssumeRole 编排方法（支持 Role Chain）
`CloudAdaptorClient` SHALL 提供 `AwsWithAssumeRole` 方法，接收 `kt`、`rootAccountID`（根账号 HCM 内部 ID）、`cloudID`（成员账号 AWS Account ID，从 main_account 表获取）和 `roleChain`（角色名数组）参数，返回具备成员账号访问权限的 `*aws.Aws` client。该方法内部 SHALL 用 `rootAccountID` 调 `AwsRoot()` 获取根账号 AK/SK，然后按 roleChain 顺序执行链式 AssumeRole。`roleChain[0..n-2]` 的角色在管理账号中 AssumeRole，`roleChain[n-1]` 在目标成员账号（cloudID）中 AssumeRole。参考 GCP GPU monitoring 的 `GcpRoot` + `MainAccount` 模式。该方法 SHALL 不改动现有 `Aws()` 和 `AwsRoot()` 方法。仅涉及 AWS 云厂商。

#### Scenario: 单步 AssumeRole（roleChain 长度为 1）
- **GIVEN** rootAccountID 有效，cloudID 为目标成员账号 AWS Account ID，roleChain 为 `["gpu-readonly"]`
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 系统依次执行：AwsRoot 获取 AK/SK → 用成员账号 CloudID 拼接 Role ARN → AssumeRole → 构建并返回 Aws client

#### Scenario: 多步 Role Chain（roleChain 长度为 2）
- **GIVEN** rootAccountID 有效，cloudID 为目标成员账号 AWS Account ID，roleChain 为 `["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"]`
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 系统执行：AwsRoot 获取 AK/SK → 用**管理账号** CloudID 拼接第一个 Role ARN 并 AssumeRole → 用第一步的临时凭证 + **成员账号** CloudID 拼接第二个 Role ARN 并 AssumeRole → 构建并返回 Aws client

#### Scenario: rootAccountID 无效
- **GIVEN** rootAccountID 在 root_account 表中不存在
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 返回错误，错误信息包含 rootAccountID

#### Scenario: Role Chain 中间步骤失败
- **GIVEN** roleChain 中某个中间角色不存在或权限不足
- **WHEN** 调用 `AwsWithAssumeRole`
- **THEN** 返回 AssumeRole 失败错误，包含失败的 Role ARN

### Requirement: BaseSecret 支持 SessionToken
`BaseSecret` 结构体 SHALL 新增 `CloudSessionToken` 字段。`newClientSet` 初始化 AWS SDK credentials 时 SHALL 使用 `secret.CloudSessionToken` 替代硬编码的空字符串。其他云厂商的 `CloudSessionToken` 零值为空字符串，与当前行为等效，无任何影响。涉及所有云厂商（变更对 TCloud/HuaWei/GCP/Azure 零影响）。

#### Scenario: STS 临时凭证初始化
- **GIVEN** BaseSecret 的 CloudSessionToken 非空
- **WHEN** newClientSet 构建 credentials
- **THEN** 使用 AccessKeyId + SecretAccessKey + SessionToken 创建 StaticCredentials

#### Scenario: 非 STS 场景向后兼容
- **GIVEN** BaseSecret 的 CloudSessionToken 为空字符串（零值）
- **WHEN** newClientSet 构建 credentials
- **THEN** 行为与当前硬编码 `""` 完全一致，无任何变化
