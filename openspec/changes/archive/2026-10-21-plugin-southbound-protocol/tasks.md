## 1. RSA-OAEP 加密与 launch_config 加密配置封装

- [x] 1.1 在 `pkg/cryptography/rsa_oaep.go` 中实现符合南向协议 §3.2.2 规范的 SPKI RSA 公钥解析（`-----BEGIN PUBLIC KEY-----`）、不低于 3072 位长度校验、明文长度上限检查（`keyBytes - 66`，3072 位为 318 字节）以及 RSA-OAEP（OAEP Hash 与 MGF1 Hash 均为 SHA-256，label 为空）+ 标准 Base64（RFC 4648 §4）编码
- [x] 1.2 在 `pkg/cryptography/rsa_oaep_test.go` 中编写单元测试，覆盖 3072/4096 位正常加解密还原、OAEP 每次加密密文不同、318/319 字节明文长度边界、低于 3072 位拒绝、非 SPKI/私钥/ECDSA/含尾部多余数据 PEM 拒绝等场景
- [x] 1.3 在 `pkg/thirdparty/api-gateway/plugin/crypto.go` 中实现 `CryptoType`（`RSA`）、`RSACryptoConfig`（`public_key`）、`ValidateCryptoConfig` 及 `LaunchConfigEncryptor` 接口封装（支持 `map[string]any` 与原始 UTF-8 JSON 明文加密为 `launch_config`）

## 2. 南向协议 Schema 模型、诊断结构与 11 个接口报文定义

- [x] 2.1 在 `pkg/thirdparty/api-gateway/plugin/schema.go` 中定义对齐 Terraform `providers schema -json` 的 `PluginSchema`（`protocol_version`、`plugin`、`schema_hash`、`capabilities`、`provider`、`resource_schemas`、`data_source_schemas`）、`Schema`、`Block`、`Attribute`、`NestedBlock`、`Capabilities`（9 个能力位）结构体，并实现顶层元数据、类型名 `<plugin.name>_` 前缀、属性修饰符（`required`/`optional`/`computed` 互斥与必选）、嵌套块（`single`/`list`/`set`/`map` 及 `min_items`/`max_items`）以及 `capabilities` 正反向声明约束的校验方法
- [x] 2.2 在 `pkg/thirdparty/api-gateway/plugin/types.go` 中定义对齐 Terraform `AttributePath.Step` 的 `PathStep`（`attribute`/`key`/`index` 严格三选一校验）、`Diagnostic`（含默认 `error` 级别与 `ValidateData.IsPassed()` 判定）、`TaskStatus`（`running`/`succeeded`/`failed`）、`Progress` 以及全部 11 个南向接口的 Request / Response Data 结构体与入参 `Validate()` 校验方法（含 `Apply`/`Plan` 的 `prior_state` 与 `proposed_state` 非同时为空校验、`List` 的 `limit` 1~500 范围校验）

## 3. 南向 HTTP 客户端与单元测试

- [x] 3.1 在 `pkg/thirdparty/api-gateway/plugin/client.go` 中定义 `Client` 接口，实现插件网关 `base_url` 解析（拆分 `Scheme://Host` 与网关环境 `basePath`），并基于 `apigateway.ApiGatewayCall` / `apigateway.ApiGatewayCallWithoutReq` 实现 11 个南向接口方法（`Health`、`GetSchema`、`Apply`、`GetResult`、`Stop`、`Plan`、`Validate`、`Import`、`Refresh`、`ReadData`、`List`）
- [x] 3.2 在 `pkg/thirdparty/api-gateway/plugin/client_test.go` 中编写单元测试，覆盖 `launch_config` 加密器、Schema/Capabilities 正反向约束校验、`Diagnostic`/`PathStep` 判定以及 11 个南向接口的 HTTP 调用（成功、异步 `running`、任务失败 `status=failed` 带 `data.message` 与部分落地 `state`、`Refresh` 对象不存在返回 `state: null`、网关接口非 0 错误拦截）
