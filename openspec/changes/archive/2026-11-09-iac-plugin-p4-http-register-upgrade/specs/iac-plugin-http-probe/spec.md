## ADDED Requirements

### Requirement: 地址校验
系统 SHALL 在注册、修改、探活时校验 `api_path`，必须同时满足：是以 `/` 开头的路径后缀（MUST NOT 含 scheme、host、userinfo、`..` 段、空白与控制字符）；拼接配置的网关环境地址前缀后，解析出的 scheme 与 host MUST 与前缀完全一致，即最终请求只能落在蓝鲸 API 网关环境内；探活通过且返回了 `plugin_version`。地址校验不通过 MUST 返回参数错误，且 MUST NOT 发出任何请求。

#### Scenario: 合法后缀
- **WHEN** `api_path` 为 `/api/bk-hcm-plugin-cmdb/prod`
- **THEN** 地址校验通过

#### Scenario: 带主机名
- **WHEN** `api_path` 为 `http://evil.example.com/x` 或 `//evil.example.com/x`
- **THEN** 返回参数错误，不发出请求

#### Scenario: 路径穿越
- **WHEN** `api_path` 含 `..` 段
- **THEN** 返回参数错误，不发出请求

#### Scenario: 拼接后 host 被改变
- **WHEN** 拼接结果解析出的 host 与配置前缀的 host 不同
- **THEN** 返回参数错误

### Requirement: 加密配置校验
系统 SHALL 校验 `crypto_type` 与 `crypto_config`：`crypto_type` MUST 在支持列表内（当前仅 `RSA`）；`RSA` 下 `crypto_config.public_key` MUST 为 PEM 编码的 SPKI 格式、可解析、类型为 RSA、长度不低于 3072 位。校验失败 MUST 返回参数错误并指出原因，且 MUST NOT 探活、MUST NOT 写库。

#### Scenario: 合法公钥
- **WHEN** 提供 3072 位 RSA 公钥
- **THEN** 校验通过

#### Scenario: 公钥长度不足
- **WHEN** 提供 2048 位 RSA 公钥
- **THEN** 返回参数错误，指出长度不低于 3072

#### Scenario: 非 RSA 公钥
- **WHEN** 提供 EC 公钥
- **THEN** 返回参数错误

#### Scenario: 不支持的加密类型
- **WHEN** `crypto_type` 为 `AES`
- **THEN** 返回参数错误

#### Scenario: 不能解析
- **WHEN** `public_key` 不是合法 PEM
- **THEN** 返回参数错误

### Requirement: 探活
hc-service SHALL 在进程内提供 http 插件探活能力（注册、升级、修改、巡检共用），MUST NOT 为此暴露路由。探活 MUST 先做地址校验；通过后请求 `{网关地址前缀}{api_path}/health`，请求 MUST 带 HCM 应用身份、使用固定超时（暂定 10 秒，可配置）、MUST NOT 跟随重定向、响应体不超过 1MB。网关 `iacGateway.endpoints` 未配置时探活 MUST 返回参数错误且不发出请求。统一响应中 `code` 为 0 且 `data.plugin_version` 非空才算通过；否则 MUST 返回业务处理异常，错误信息带上插件返回的 `message`。

#### Scenario: 探活成功
- **WHEN** 插件返回 `code=0` 且 `plugin_version=1.2.0`
- **THEN** 探活结果为 `1.2.0`

#### Scenario: 缺少版本
- **WHEN** 插件返回 `code=0` 但 `plugin_version` 为空
- **THEN** 返回业务处理异常

#### Scenario: 插件返回错误
- **WHEN** 插件返回 `code!=0` 且 `message=db down`
- **THEN** 返回业务处理异常，信息含 `db down`

#### Scenario: 重定向
- **WHEN** 插件地址返回 3xx
- **THEN** 不跟随，按失败处理

#### Scenario: 超时
- **WHEN** 插件在超时时间内未响应
- **THEN** 返回超时错误
