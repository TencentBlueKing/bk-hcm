## ADDED Requirements

### Requirement: SPKI RSA 公钥解析与位数校验
系统 SHALL 支持按照南向协议 §3.2.2 规范解析插件方提供的 PEM 编码 SPKI 格式（`-----BEGIN PUBLIC KEY-----`）RSA 公钥，并强制校验密钥位数不低于 3072 位，确保私钥仅由插件方保管、HCM 侧仅使用公钥加密。

#### Scenario: 合法的 3072 位及以上 SPKI RSA 公钥解析成功
- **WHEN** 传入标准 PEM 编码的 SPKI 格式 3072 位（或更高位数如 4096 位）RSA 公钥
- **THEN** 系统成功解析并返回 `*RSAOaep` 实例，且 `MaxPlaintextBytes()` 返回 `公钥字节数 - 66`（对 3072 位即 384 字节密钥返回 `318`，对 4096 位即 512 字节密钥返回 `446`）

#### Scenario: 非 SPKI 格式或含多余数据的 PEM 被拒绝
- **WHEN** 传入空字符串、非 PEM 格式、PKCS#1 格式（`-----BEGIN RSA PUBLIC KEY-----`）、私钥 PEM（`-----BEGIN PRIVATE KEY-----`）、非 RSA 公钥（如 ECDSA）或 PEM 块后附带多余非空白内容
- **THEN** 系统返回明确错误并拒绝构造加密器

#### Scenario: 低于 3072 位的 RSA 公钥被拒绝
- **WHEN** 传入低于 3072 位（如 2048 位）的 SPKI 格式 RSA 公钥
- **THEN** 系统返回密钥位数不足错误（要求至少 3072 位）

### Requirement: RSA-OAEP SHA-256 加密与标准 Base64 编码
系统 SHALL 使用 RSA-OAEP 算法（OAEP 哈希与 MGF1 哈希均为 SHA-256，label 为空）对 UTF-8 明文进行非对称加密，并将二进制密文按标准 Base64（RFC 4648 第 4 节）编码为字符串作为 `launch_config` 的值。

#### Scenario: 明文加密并可用对应私钥按协议参数解密还原
- **WHEN** 对不超过长度上限的 UTF-8 JSON 明文（如 `{"access_token":"...","bk_tenant_id":"default"}`）调用 `EncryptToBase64`
- **THEN** 返回标准 Base64 字符串（3072 位密钥加密结果固定为 512 个字符），同一份明文每次加密结果均不相同（OAEP 随机填充特性），且插件方使用对应 RSA 私钥先做标准 Base64 解码再按 RSA-OAEP（OAEP Hash SHA-256、MGF1 Hash SHA-256、label 为空）解密后与原始 UTF-8 JSON 明文完全一致

#### Scenario: 明文超过公钥允许长度上限时前置拦截报错
- **WHEN** 使用 3072 位公钥对 319 字节（超过 `384 - 66 = 318` 字节上限）的明文调用 `Encrypt` 或 `EncryptToBase64`
- **THEN** 系统前置拦截并返回包含当前字节数、最大允许字节数与密钥位数的明文长度超限错误

### Requirement: 插件加密配置校验与 launch_config 加密封装
系统 SHALL 支持校验插件注册时的加密配置（`crypto_type` 与 `crypto_config`），并提供 `LaunchConfigEncryptor` 接口将符合 Schema `provider` 定义的插件实例启动参数序列化为 UTF-8 JSON 后加密为 `launch_config` 字符串，同时屏蔽底层加密算法细节以便未来扩展其他 `crypto_type`。

#### Scenario: 合法的 RSA crypto_config 校验与 launch_config 加密成功
- **WHEN** `crypto_type` 为 `"RSA"` 且 `crypto_config` JSON 包含合法的 3072 位 SPKI `public_key`
- **THEN** `ValidateCryptoConfig` 校验通过，`NewLaunchConfigEncryptor` 成功构造 `LaunchConfigEncryptor` 实例，调用 `Encrypt(providerConfig)` 或 `EncryptJSON(rawJSON)` 可将启动参数加密为标准 Base64 `launch_config` 字符串

#### Scenario: 不支持的 crypto_type 或缺失/非法 public_key 被拒绝
- **WHEN** `crypto_type` 不是 `"RSA"`（或为空），或 `crypto_config` 为空、JSON 格式非法、缺少 `public_key` 字段，或 `public_key` 不满足 SPKI >= 3072 位 RSA 公钥要求
- **THEN** `ValidateCryptoConfig` 与 `NewLaunchConfigEncryptor` 返回具体的参数校验错误
