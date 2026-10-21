## Context

海垒（HCM）通过插件机制接入外部系统（如 CMDB、ITSM 等内部系统或用户自定义流程），再由上层 DAG 统一编排执行。除面向云厂商官方插件、有会话的 Terraform gRPC 插件协议（tfplugin6）外，HCM 定义了轻量的无状态 HTTP 南向协议。漂移检测、目标状态合成、启动参数解析均由 `hc-service` 完成，南向插件仅负责将外部系统调整为给定的目标状态并交回最新状态。

当前代码库背景：
1. `pkg/cryptography` 仅提供对称加解密 `AESGcm`（实现 `Crypto` 接口，其 `EncryptToBase64` 不返回 `error` 且要求具备私钥解密能力）。而南向协议要求 HCM 侧仅持有插件方在注册时提供的 RSA 公钥（SPKI PEM 格式，长度不低于 3072 位），私钥由插件方自行保管；每次调用南向 POST 接口前，HCM 按 RSA-OAEP（SHA-256 + MGF1-SHA-256，label 为空）加密 `provider` 启动参数明文 JSON，生成标准 Base64 编码的 `launch_config`。
2. `pkg/thirdparty/api-gateway/thirdparty.go` 已提供通用蓝鲸网关调用函数 `ApiGatewayCall` 和 `ApiGatewayCallWithoutReq`。南向协议所有接口响应已统一对齐为 `{ "result": true, "code": 0, "message": "", "data": {} }`，且将“HTTP/网关调用成败”与“业务任务成败/配置校验结论”解耦（任务失败体现在 `data.status="failed"` 与 `data.message`，配置不合法体现在 `data.diagnostics`），因此可直接复用现有网关调用层。

## Goals / Non-Goals

**Goals:**
- 在 `pkg/cryptography` 中提供符合南向协议 §3.2.2 规范的 `RSAOaep` 公钥解析、密钥位数校验（>= 3072 位）、明文长度上限检查（`公钥字节数 - 66`）及 RSA-OAEP + 标准 Base64 加密实现。
- 在 `pkg/thirdparty/api-gateway/plugin` 中提供 `crypto_type`（`RSA`）与 `crypto_config`（`public_key`）校验及 `LaunchConfigEncryptor` 封装，供南向请求加密及后续插件注册直接复用。
- 定义南向协议完整数据结构（基于 `github.com/hashicorp/terraform-json` 扩展 `PluginSchema` 与 `Diagnostic`，以及 `Capabilities`、`PathStep`、`TaskStatus`、`Progress` 和 11 个接口的 Req/Resp 模型）与参数/能力位校验逻辑。
- 实现南向协议 11 个 HTTP 接口客户端（`Health`、`GetSchema`、`Apply`、`GetResult`、`Stop`、`Plan`、`Validate`、`Import`、`Refresh`、`ReadData`、`List`），直接复用 `apigateway.ApiGatewayCall` 与 `apigateway.ApiGatewayCallWithoutReq`。

**Non-Goals:**
- 不包含插件注册表（`iac_plugin`）建表、DAO 层及 `cloud-server` 插件注册/升级/修改/启停管理接口（由插件注册方案实现）。
- 不包含 `iac_plugin_schema` 与 `iac_plugin_instance` 的存储与缓存（由 Schema 与实例管理方案实现）。
- 不修改现有 `pkg/thirdparty/api-gateway/thirdparty.go`。

## Protocol Specification Summary

### 1. 调用链路与蓝鲸网关（APIGW）约定
- **网关粒度**：每个插件对应一个蓝鲸网关（如 `bk-hcm-plugin-cmdb`），`base_url` 为网关环境地址（例如 `https://bkapi.example.com/api/bk-hcm-plugin-cmdb/prod`），南向接口路径直接拼接在 `base_url` 之后。
- **认证机制**：网关开启**应用认证**（对 HCM 的 `bk_app_code` 授权，请求头携带 `X-Bkapi-Authorization`），关闭**用户认证**（操作人身份由加密在 `launch_config` 中的用户态令牌或凭证确定）。

### 2. 11 个南向 HTTP 接口总览与能力位映射

| 接口方法 | HTTP 方法 + 路径 | 是否必须实现 | 对应 `capabilities` 位 | 协议职责说明 |
|---|---|---|---|---|
| `Health` | `GET /health` | 必须 | 无（默认必选） | 探活并返回 `plugin_version`；版本变化时触发 HCM 重新拉取 Schema |
| `GetSchema` | `GET /schema` | 必须 | 无（默认必选） | 返回完整 `PluginSchema`（含 `schema_hash`、`capabilities`、`provider`、`resource_schemas`、`data_source_schemas`） |
| `Apply` | `POST /apply` | `resource_schemas` 非空时必须 | `apply` | 根据 `prior_state` 与 `proposed_state` 执行资源创建、更新或删除；支持同步完成或返回 `running` 转异步 |
| `GetResult` | `POST /result` | `capabilities.async_apply=true` 时必须 | `async_apply` | 按 `apply_id` 查询异步 `Apply` 的当前执行状态、进度与已落地 `state` |
| `Stop` | `POST /stop` | 可选（须同时声明 `async_apply`） | `stop` | 按 `apply_id` 请求停止进行中的异步 `Apply`（**非回滚**，已落地变更通过 `state` 返回） |
| `Plan` | `POST /plan` | `resource_schemas` 非空时必须 | `plan` | 预览资源变更后的预计 `state` 及需先删后建的字段列表 `replace_paths`，不改动外部系统 |
| `Validate` | `POST /validate` | 可选 | `validate` | 在 HCM 完成 Schema 结构校验后，由插件校验业务规则并返回全量 `diagnostics` |
| `Import` | `POST /import` | `capabilities.import=true` 时必须 | `import` | 按外部系统对象标识 `import_id` 读取已有资源的完整 `state` 用于纳管 |
| `Refresh` | `POST /refresh` | `resource_schemas` 非空时必须 | `refresh` | 根据已保存的 `state`（提取定位字段）读取外部系统最新状态，用于漂移检测；对象已不存在时返回 `state: null` |
| `ReadData` | `POST /read_data` | `data_source_schemas` 非空时必须 | `read_data` | 按 `config` 过滤条件查询只读数据源并返回完整 `state` |
| `List` | `POST /list` | `capabilities.list=true` 时必须 | `list` | 基于 `limit`（`1~500`）与 `page_token` 分页拉取当前插件实例作用域下的存量资源列表（含 `import_id` 与 `state`） |

### 3. 统一响应包络与业务状态分离语义
所有接口统一返回 `{ "result": bool, "code": int, "message": string, "data": object | null }`：
- **接口调用级状态（`result` 与 `code`）**：`result: true` 且 `code: 0` 仅表示本次 HTTP 请求被插件正常处理并返回了业务响应；`result: false` 或 `code != 0`（如 `50300` 外部系统不可用、`40400` 导入/查询目标不存在）表示接口调用失败，此时 `data` 为 `null`，调用方不改动已保存的 Runtime 状态。
- **任务执行级状态（`Apply` / `GetResult` / `Stop` 的 `data.status`）**：
  - `running`：异步任务执行中，必须返回非空 `apply_id`；若有增量落地可返回当前实际 `state` 作为检查点。
  - `succeeded`：执行流程正常走完（包括审批单被驳回 `approve_result="rejected"`、部分交付等业务层面的正常终态），必须返回执行后的完整 `state`（删除成功时 `state` 为 `null`）。
  - `failed`：执行流程异常中断，此时外层仍为 `result: true, code: 0`，失败原因写入必填的 `data.message`。**部分落地要求**：若失败前外部系统已发生实际变更（例如主机已创建但转移模块失败，或审批单已创建但被管理员强制终止），插件必须在 `failed` 响应中返回外部系统的**实际 `state`**（含主键定位字段，如 `bk_host_id`），供下次重试作为 `prior_state`；若外部系统未发生任何变更则 `state` 为 `null`。
  - `progress`：可选进度信息（`total`、`succeeded`、`failed`、`description`），仅用于页面展示，不写入 Runtime、不参与漂移比对。
- **配置校验级状态（`Validate` 的 `data.diagnostics`）**：配置不合法不是接口调用失败，外层仍为 `result: true, code: 0`，所有问题列在 `data.diagnostics` 数组中；当且仅当请求未能完成（如依赖的外部服务不可用）时才返回 `result: false, code != 0`。

### 4. 公共请求字段与 `launch_config` 加密协议
除 `Health` 与 `GetSchema` 外，其余 9 个 `POST` 接口请求体均包含：
- `type_name`：操作的目标类型名，必须属于 Schema 中 `resource_schemas` 或 `data_source_schemas` 的 key，且以 `<plugin.name>_` 为前缀（如 `bkcmdb_host`、`itsm_ticket`）。
- `launch_config`：加密后的插件实例启动参数字符串。
  - **加密配置**：插件注册时指定 `crypto_type`（当前支持 `"RSA"`）与 `crypto_config`（`{"public_key": "..."}`）。
  - **公钥要求**：PEM 编码的 SPKI 格式（`-----BEGIN PUBLIC KEY-----`），RSA 密钥长度不低于 3072 位。
  - **加密算法与编码**：将符合 `provider` Schema 的 UTF-8 JSON 明文，使用 RSA-OAEP（OAEP Hash 为 SHA-256，MGF1 Hash 为 SHA-256，label 为空）加密为二进制密文，再按标准 Base64（RFC 4648 §4）编码为字符串（3072 位密钥固定生成 512 字符）。
  - **明文长度上限**：最大明文字节数为 `公钥字节数 - 66`（`keyBytes - 2*32 - 2`），3072 位（384 字节）公钥最多支持 **318 字节**明文。

### 5. Schema 结构与约束校验规则
Schema 直接内嵌复用 `github.com/hashicorp/terraform-json` 的 `tfjson.ProviderSchema`（内部属性类型表达式在 JSON 反序列化时由 `cty.Type` 自动解析校验），并扩展 HCM 顶层元数据：
- **顶层元数据与块非空**：`protocol_version`（必填，如 `"1.0"`）、`plugin.name`（必填，全局唯一且作为所有 `type_name` 的前缀）、`plugin.version`（必填，与 `/health` 的 `plugin_version` 一致）、`schema_hash`（必填，Schema 内容指纹）、`capabilities`、`provider`（`provider.block` 必填）、`resource_schemas`、`data_source_schemas`（各类型 `block` 必填）。
- **类型名前缀强约束**：`resource_schemas` 与 `data_source_schemas` 中的所有 key 必须以 `plugin.name + "_"` 开头。
- **`capabilities` 声明正反向约束**：
  1. `resource_schemas` 非空时，`apply`、`plan`、`refresh` 必须全部为 `true`；
  2. `resource_schemas` 为空时，所有资源类能力位（`apply`、`plan`、`refresh`、`async_apply`、`stop`、`import`、`list`）均不得为 `true`；
  3. `data_source_schemas` 非空时，`read_data` 必须为 `true`；`data_source_schemas` 为空时，`read_data` 不得为 `true`；
  4. `stop` 为 `true` 时，`async_apply` 必须同时为 `true`。

### 6. `Diagnostic` 与 `PathStep` 诊断路径规范
- **`PathStep` 三选一结构**：对齐 Terraform `AttributePath.Step`，每个 `PathStep` 有且仅有以下一个字段：
  - `attribute`（string，非空）：取当前对象的属性名或嵌套块名（必须使用 Schema 中的字段名，不得使用外部系统内部字段名）；
  - `key`（string）：取当前 map 类型字段的某个 key；
  - `index`（int，`>= 0`）：取当前 list 类型字段的某个下标（从 0 开始；`set` 类型无序，不加 `index`，路径只写到 `set` 字段本身）。
- **`Diagnostic` 结构与判定**：
  - `path`：从请求 `config` 根节点由外到内排列的 `PathStep` 数组；与具体字段无关的跨字段错误传空数组 `[]`。
  - `message`：必填的问题说明（不重复字段名）。
  - `severity`：可选，取值为 `"error"` 或 `"warning"`，省略时默认为 `"error"`。
  - **通过判定（`ValidateData.IsPassed()`）**：`diagnostics` 为空数组或仅包含 `severity: "warning"` 时判定为通过；包含任意 `severity: "error"`（含缺省）时判定为不通过。

### 7. 资源操作推导与分页约定
- **`Apply` 与 `Plan` 的操作推导**：根据 `prior_state` 与 `proposed_state` 是否为 `nil` 推导操作类型：
  - `prior_state == nil && proposed_state != nil` → **创建（Create）**
  - `prior_state != nil && proposed_state != nil` → **更新（Update）**
  - `prior_state != nil && proposed_state == nil` → **删除（Delete）**
  - `prior_state == nil && proposed_state == nil` → **非法请求**（客户端前置拦截报错）
- **`List` 分页约定**：请求 `limit` 必须在 `1~500` 范围内；首页 `page_token` 传空字符串 `""`，后续页原样回传上一页响应的 `next_page_token`，响应 `next_page_token` 为空字符串 `""` 表示已到最后一页。

## Decisions

### 1. 独立定义 `RSAOaep` 与 `LaunchConfigEncryptor`，不强行塞入对称 `cryptography.Crypto` 接口
- **原因**：现有 `cryptography.Crypto` 接口同时包含 `Encrypt` 与 `Decrypt`，且 `EncryptToBase64(plaintext string) string` 不返回 `error`。而南向 `launch_config` 在 HCM 侧只有公钥（只加密不解密），且 RSA-OAEP 加密存在明文超长（> `keyBytes - 66`）和随机数读取错误等失败分支，必须返回 `(string, error)`。
- **做法**：
  - 在 `pkg/cryptography/rsa_oaep.go` 中实现 `RSAOaep`，提供 `ParseRSAPublicKey(pubKeyPEM string)`、`NewRSAOaepFromPEM(pubKeyPEM string) (*RSAOaep, error)`、`MaxPlaintextBytes() int`、`Encrypt(plaintext []byte) ([]byte, error)`、`EncryptToBase64(plaintext []byte) (string, error)`。
  - 在 `pkg/thirdparty/api-gateway/plugin/crypto.go` 中定义 `LaunchConfigEncryptor` 接口与工厂函数 `NewLaunchConfigEncryptor(cryptoType CryptoType, cryptoConfig json.RawMessage)`，屏蔽具体加密算法细节，方便未来扩展新 `crypto_type`（对称或非对称）。

### 2. 直接复用 `apigateway.ApiGatewayCall` 与 `apigateway.ApiGatewayCallWithoutReq`
- **原因**：南向协议统一响应已规范为 `{ "result": true, "code": 0, "message": "", "data": {} }`，且任务失败（`status=failed`）与配置校验不通过时外层仍为 `result=true, code=0`，业务失败信息分别放入 `data.message` 与 `data.diagnostics`，已完全兼容 `thirdparty.go` 的判断逻辑（`if !resp.Result || resp.Code != 0`）。
- **做法**：
  - `GET /health`、`GET /schema` 使用 `apigateway.ApiGatewayCallWithoutReq`。
  - 其余 9 个 `POST` 接口在客户端方法内先执行 `req.Validate()`，再调用 `apigateway.ApiGatewayCall`。
  - 客户端构造时接收插件网关的 `baseURL`（解析为 `Scheme://Host` 交给 `apigateway.Discovery`，`Path` 交给 `rest.NewClient`）和 `*cc.ApiGateway` 认证配置。

### 3. `PathStep` 使用指针字段表达严格三选一语义
- **原因**：`PathStep` 有且仅有 `attribute`（string）、`key`（string）、`index`（int，从 0 开始）之一。由于 `0` 和空串在 Go 值类型下无法区分“未设置”还是“显式设置了 `0` 或空串”，使用 `*string` 和 `*int`（配合 `cvt.ValToPtr` / `cvt.PtrToVal`）可精确统计非 `nil` 字段个数必须等于 1，并支持对 `index: 0` 或 `key: ""` 的合法表达。

### 4. `PluginSchema.ProtocolVersion` 仅校验非空，不硬编码限制固定版本号字符串
- **原因**：南向协议后续演进遵循向后兼容原则（增量扩展字段与 `capabilities` 能力位），若在客户端校验中硬编码 `protocol_version == "1.0"`，会导致未来兼容的高版本插件（如 `"1.1"`）被旧版代码直接拒绝。因此 `PluginSchema.Validate()` 仅校验 `protocol_version` 非空，具体能力支持交由 `capabilities` 门禁控制。

## Risks / Trade-offs

- **[Risk] RSA-OAEP 明文长度受限于公钥长度（3072 位公钥最多 318 字节）** → **Mitigation**：在 `RSAOaep.Encrypt` 中前置校验 `len(plaintext) <= r.MaxPlaintextBytes()`，超限时立即返回包含当前字节数、最大允许字节数与密钥位数的明确错误，避免底层标准库返回模糊的 `crypto/rsa: message too long for RSA key size`。
- **[Risk] 插件 `base_url` 包含网关环境路径前缀（如 `https://bkapi.example.com/api/bk-hcm-plugin-cmdb/prod`）** → **Mitigation**：在 `plugin.NewClient` 中使用 `url.Parse(baseURL)` 拆分出 `Scheme://Host`（填入 `apigateway.Discovery.Servers`）与 `Path`（填入 `rest.NewClient(c, basePath)`），确保与 `rest.Request.WrapURL()` 拼接方式完全契合。
