## ADDED Requirements

### Requirement: 统一请求指标固定标签集合
系统 MUST 提供统一请求指标 `http_request_cost_seconds`、`http_request_total`、`http_request_fail_total`，并保证同名指标在所有打点组件中使用固定 label set。`http_request_cost_seconds` 和 `http_request_total` 的 label MUST 为 `component,endpoint,method,vendor`；`http_request_fail_total` 的 label MUST 为 `component,endpoint,method,vendor,err_type`。

#### Scenario: 请求指标标签集合一致
- **WHEN** `api-server`、`hc-service`、`data-service` 或 `adaptor` 上报 `http_request_*` 指标
- **THEN** 同名指标的 label 集合完全一致，不因组件不同而增减 label

#### Scenario: 失败指标包含错误类型
- **WHEN** 任一服务接口请求或云 API 调用被判定为失败
- **THEN** 系统在 `http_request_fail_total` 中按 `err_type` 记录失败次数

### Requirement: 服务端 HTTP 请求全量覆盖
系统 MUST 在 `api-server`、`hc-service`、`data-service` 的所有服务接口上报统一请求指标。服务端指标的 `component` MUST 分别为 `api-server`、`hc-service`、`data-service`，`endpoint` MUST 使用路由模板，`method` MUST 使用 HTTP 方法，`vendor` MUST 固定为 `none`。

#### Scenario: 服务接口请求完成
- **WHEN** `api-server`、`hc-service` 或 `data-service` 任一服务接口请求完成
- **THEN** 系统记录该请求的耗时和请求总量，且 `vendor` label 为 `none`

#### Scenario: 服务接口请求失败
- **WHEN** 服务接口因参数、鉴权、业务错误、下游错误或内部错误返回失败
- **THEN** 系统记录该接口的失败次数，并将错误归一化为有限 `err_type`

### Requirement: adaptor 云 API 调用接入统一请求指标
系统 MUST 在 adaptor 云厂商 API 调用完成后上报统一请求指标。adaptor 指标的 `component` MUST 固定为 `adaptor`，`endpoint` MUST 使用稳定云 API 名称，`method` MUST 固定为 `SDK` 或 `CALL`，`vendor` MUST 使用真实云厂商。

#### Scenario: 云 API 调用成功
- **WHEN** adaptor 调用云厂商 API 成功返回
- **THEN** 系统记录 `component=adaptor` 的调用耗时和请求总量，并按真实 `vendor` 和稳定 API 名称聚合

#### Scenario: 云 API 调用失败
- **WHEN** adaptor 调用云厂商 API 因超时、网络错误、云厂商错误或 HCM 内部错误失败
- **THEN** 系统记录 `http_request_fail_total`，并将失败原因归一化为有限 `err_type`

### Requirement: 禁止高基数标签
系统 MUST 禁止在新增 metrics label 中使用原始 URL、rid、任务 ID、Flow ID、实例 ID、监听器 ID、资源 ID、错误信息原文等高基数字段。需要定位单个请求或任务时，系统 MUST 通过日志、rid、任务 ID 或数据库记录关联排查。

#### Scenario: HTTP endpoint 标签生成
- **WHEN** 服务端请求路径包含业务 ID、资源 ID 或其它动态路径参数
- **THEN** `endpoint` label 使用路由模板，不使用请求原始 URL

#### Scenario: 错误信息记录
- **WHEN** 请求或云 API 调用失败并包含详细错误文本
- **THEN** metrics 仅记录有限 `err_type`，详细错误保留在日志中用于关联排查
