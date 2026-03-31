## Context

hc-service 是混合云管理平台中负责与云厂商 API 交互的服务层。当前已有子账号（SubAccount）管理的完整 CRUD 接口链路：adaptor 层封装腾讯云 SDK 调用 → hc-service handler 层提供 HTTP 接口。访问密钥管理需要沿用完全相同的分层架构和编码模式。

现有代码约定：
- adaptor 类型定义在 `pkg/adaptor/types/account/tcloud.go`
- adaptor 实现在 `pkg/adaptor/tcloud/account.go`，接口声明在 `interface.go`
- hc-service API 类型定义在 `pkg/api/hc-service/sub-account/tcloud.go`
- hc-service handler 实现在 `cmd/hc-service/service/account/` 目录下
- 路由注册在 `cmd/hc-service/service/account/service.go`

用户已创建空文件 `cmd/hc-service/service/account/secret.go` 用于放置 handler。

## Goals / Non-Goals

**Goals:**
- 实现腾讯云访问密钥的创建（CreateAccessKey）、删除（DeleteAccessKey）、更新（UpdateAccessKey）三个 hc-service 接口
- 完全复用现有架构分层模式，不引入新的抽象
- 复用已有的工具函数和类型（如 `converter`、`errf`、`validator` 等）

**Non-Goals:**
- 不涉及 data-service 层的密钥持久化存储
- 不涉及 cloud-server 层的业务编排或审批流程
- 不处理密钥的加密存储，仅透传云端 API 结果
- 不实现密钥列表查询（ListAccessKeys）接口

## Decisions

### 1. 三个接口均使用 AccountID + TargetUin 定位目标用户

腾讯云 CAM 的访问密钥 API 通过 `TargetUin` 指定操作目标子用户。hc-service 层需要 `AccountID` 来获取对应的云适配器客户端（含密钥信息），因此请求结构统一包含 `AccountID`（必填）和 `TargetUin`（必填）两个字段。

### 2. Handler 统一放在 secret.go

三个密钥接口属于同一个功能域（访问密钥管理），统一放在用户已创建的 `secret.go` 中，与子账号管理的 `sub_account.go` 并列。

### 3. 沿用现有 hc-service handler 模式

每个 handler 的流程为：DecodeInto → Validate → 获取 adaptor client → 调用 adaptor 方法 → 返回结果。与 `TCloudCreateSubAccount`、`TCloudDeleteSubAccount` 等完全一致。

### 4. UpdateAccessKey 语义为更新密钥状态（Active/Inactive）

腾讯云 UpdateAccessKey API 的核心功能是切换密钥的启用/禁用状态，请求参数为 `AccessKeyId` + `Status`（Active/Inactive），不涉及密钥内容的修改。

## Risks / Trade-offs

- **[密钥泄露风险]** CreateAccessKey 返回的 SecretAccessKey 仅在创建时可见，后续无法查询。hc-service 仅透传不存储，调用方需自行保管。→ 接口文档中需说明此特性。
- **[密钥数量限制]** 每个 CAM 用户最多支持两个 AccessKey。创建时可能因超限失败。→ 透传腾讯云错误码 `OperationDenied.AccessKeyOverLimit`。
