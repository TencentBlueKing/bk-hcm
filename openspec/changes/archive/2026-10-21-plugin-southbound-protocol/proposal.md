## Why

海垒（HCM）通过插件机制接入外部系统，再由 DAG 统一编排执行。云厂商官方插件遵循有会话的 Terraform gRPC 插件协议（tfplugin6），要求使用 Go SDK 开发并按 Terraform 进程模型运行，接入门槛较高。为支持快速接入 CMDB、ITSM 等内部系统或用户自定义流程，HCM 定义了一套轻量、无状态的 HTTP 南向协议。

相较于 Terraform gRPC 插件协议，HTTP 南向协议的核心机制如下：
1. **无状态与每次请求携带加密启动参数**：不建立会话、不提供独立的 Configure 接口；HCM 每次发起 POST 请求时均在请求体中携带经插件 RSA 公钥（RSA-OAEP with SHA-256）加密后的实例启动参数 `launch_config`，由插件方用私钥无状态解密。
2. **支持长耗时异步执行与单次任务停止**：针对耗时数小时至数天的操作（如 ITSM 审批单），`Apply` 可立即返回 `status: "running"` 与 `apply_id`，由调用方通过 `GetResult` 轮询进度与结果，或通过 `Stop` 按 `apply_id` 请求停止单次执行（非回滚，已落地状态通过 `state` 交回）。
3. **网关化接入与按需能力声明**：每个插件在蓝鲸 API 网关（APIGW）上独立配置网关（开启应用认证、关闭用户认证），必须实现 `GET /health` 与 `GET /schema`，并在 Schema 的 `capabilities` 中声明已实现的其余接口（`apply`、`plan`、`refresh`、`read_data`、`async_apply`、`stop`、`validate`、`import`、`list`）。
4. **接口调用成败与业务任务状态解耦**：外层统一响应 `{ "result": bool, "code": int, "message": string, "data": ... }` 仅表示当次 HTTP 接口调用本身是否成功；`Apply` / `GetResult` / `Stop` 的任务成败通过 `data.status`（`running` / `succeeded` / `failed`）与 `data.message` 表达，`Validate` 的配置校验结论通过 `data.diagnostics` 表达。

本次变更为南向协议落地基础的 `launch_config` RSA-OAEP 加密能力、协议 Schema/Capabilities 与报文模型校验器，以及经蓝鲸网关（APIGW）调用的 11 个南向 HTTP 客户端接口。

## What Changes

- 新增 RSA-OAEP（SHA-256）公钥解析、校验与加密工具（`pkg/cryptography/rsa_oaep.go`）：
  - 支持解析 PEM 编码的 SPKI 格式（`-----BEGIN PUBLIC KEY-----`）RSA 公钥，强制校验密钥长度不低于 3072 位，拒绝 PKCS#1、私钥、非 RSA 密钥及含尾部多余数据的 PEM。
  - 实现 RSA-OAEP（OAEP Hash 与 MGF1 Hash 均为 SHA-256，label 为空）加密及标准 Base64（RFC 4648 §4）编码，前置校验明文长度不超过 `公钥字节数 - 66`（3072 位公钥上限为 318 字节）。
- 新增插件南向协议包（`pkg/thirdparty/api-gateway/plugin`）：
  - **加密配置与 `launch_config` 加密器（`crypto.go`）**：实现 `crypto_type`（一期支持 `RSA`）与 `crypto_config`（`public_key`）校验及 `LaunchConfigEncryptor` 封装，支持将 `provider` 启动参数序列化为 UTF-8 JSON 并加密为 Base64 `launch_config` 字符串。
  - **协议 Schema 模型与完备性校验（`schema.go`）**：内嵌复用 `github.com/hashicorp/terraform-json` 的 `tfjson.ProviderSchema`，补充 HCM 南向扩展字段定义 `PluginSchema`（`protocol_version`、`plugin`、`schema_hash`、`capabilities`），实现顶层元数据、顶层 `block` 非空、类型名 `<plugin.name>_` 前缀约束，以及 9 个 `capabilities` 能力位与 `resource_schemas` / `data_source_schemas` 的正反向声明约束校验（内部属性类型由 `tfjson` / `cty.Type` 反序列化自动保障）。
  - **公共诊断结构与 11 个接口报文模型（`types.go`）**：定义对齐 Terraform `AttributePath.Step` 的 `PathStep`（`attribute` / `key` / `index` 严格三选一），内嵌复用 `tfjson.Diagnostic`（及 `tfjson.DiagnosticSeverity`）并补充 `path` / `message` 字段定义 `Diagnostic` 及校验通过判定方法 `ValidateData.IsPassed()`；定义 `TaskStatus`（`running` / `succeeded` / `failed`）、`Progress` 以及 11 个南向接口的 Request / Response Data 结构体与入参 `Validate()` 方法。
  - **南向 HTTP 客户端实现（`client.go`）**：解析插件网关 `base_url`（拆分 `Scheme://Host` 与网关环境路径前缀），直接复用 `pkg/thirdparty/api-gateway` 的 `ApiGatewayCall` 与 `ApiGatewayCallWithoutReq` 实现全部 11 个南向 HTTP 接口（`Health`、`GetSchema`、`Apply`、`GetResult`、`Stop`、`Plan`、`Validate`、`Import`、`Refresh`、`ReadData`、`List`）。

## Capabilities

### New Capabilities
- `rsa-oaep-crypto`: 南向协议 `launch_config` 的 RSA-OAEP（SHA-256）SPKI 公钥解析校验（>= 3072 位）、明文长度上限检查（`keyBytes - 66`）与标准 Base64 加密，以及插件加密配置（`crypto_type` + `crypto_config`）校验与 `LaunchConfigEncryptor` 加密封装。
- `plugin-southbound-client`: 海垒插件南向协议完整数据模型、Schema 与 `capabilities` 声明约束校验、`Diagnostic`/`PathStep` 校验，以及基于蓝鲸 API 网关的 11 个南向 HTTP 接口客户端实现。

### Modified Capabilities

## Impact

- 新增 `pkg/cryptography/rsa_oaep.go` 及对应单元测试 `pkg/cryptography/rsa_oaep_test.go`。
- 新增 `pkg/thirdparty/api-gateway/plugin/` 包（`crypto.go`、`schema.go`、`types.go`、`client.go` 及单元测试 `client_test.go`）。
- 零侵入现有业务代码与配置文件，供后续插件注册与升级、Schema 同步、插件实例管理及 `hc-service` 北向/南向执行代理直接调用。
