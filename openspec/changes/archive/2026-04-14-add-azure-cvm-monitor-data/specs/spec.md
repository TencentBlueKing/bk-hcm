# azure-cvm-monitor-data

## ADDED Requirements

### Requirement: Azure 云主机监控接口接入统一入口
系统 MUST 在既有接口 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 中支持 `vendor=azure`，并保持既有厂商请求与响应行为不变。Azure 请求必须复用统一 `data_points` 响应主结构，厂商差异信息通过 `extensions` 承载。

#### Scenario: Azure 请求命中监控入口
- **WHEN** 调用方以 `vendor=azure` 调用统一监控接口并提交合法请求参数
- **THEN** 系统返回 `data.data_points` 结构化监控结果，不新增顶层响应结构

#### Scenario: 非 Azure 厂商兼容性
- **WHEN** 调用方继续以 `vendor=tcloud|huawei|aws` 调用统一监控接口
- **THEN** 系统行为与本次变更前保持一致，不因 Azure 参数扩展产生行为变化

### Requirement: Azure 专有参数校验与透传
系统 MUST 支持 Azure 专有参数 `metric_namespace`、`aggregation`、`auto_adjust_timegrain`、`top`、`orderby`、`filter`、`result_type`。这些参数仅在 `vendor=azure` 下生效，并在 cloud-server 与 hc-service 两层完成参数校验和透传。

#### Scenario: Azure 专有参数合法
- **WHEN** 调用方在 `vendor=azure` 请求中传入合法的 Azure 专有参数组合
- **THEN** 系统将参数透传到 Azure 监控查询链路并返回查询结果

#### Scenario: 非 Azure 请求携带 Azure 专有参数
- **WHEN** 调用方在 `vendor=tcloud|huawei|aws` 请求中传入任意 Azure 专有参数
- **THEN** 系统返回参数错误，提示该参数仅支持 Azure

### Requirement: Azure 监控调用链路分层实现
系统 MUST 在 `cloud-server -> hc-service -> adaptor` 三层中实现 Azure 监控查询能力。cloud-server 层必须按 `(account_id, region)` 分组调用 hc-service；hc-service 层必须调用 adaptor 的 Azure 监控方法；adaptor 层必须封装 Azure Metrics List 查询并输出标准化时序结果。

#### Scenario: 多账号多地域分组查询
- **WHEN** 请求中的实例 ID 覆盖多个 `account_id` 与 `region`
- **THEN** cloud-server 按 `(account_id, region)` 分组分别调用 hc-service 并聚合返回

#### Scenario: adaptor 查询成功并标准化返回
- **WHEN** hc-service 调用 Azure adaptor 监控查询接口成功
- **THEN** adaptor 返回标准化数据点，包含 `dimensions`、`timestamps`、`values` 供上层转换

### Requirement: Azure 流量语义原生优先与兜底可观测
系统 MUST 采用“原生优先、兜底回退”语义策略：若 Azure 可直接提供内网流入/流出与外网流入/流出数据，则直接输出；若当前查询条件无法直接分离四象限，系统 MUST 执行兜底语义并在 `extensions` 明确语义来源与回退状态。

#### Scenario: 可直接分离四象限
- **WHEN** Azure 返回结果能够明确区分内网与外网入/出流量
- **THEN** 系统直接返回对应四象限数据，不执行 HCM 兜底映射

#### Scenario: 无法直接分离四象限
- **WHEN** Azure 当前查询维度仅返回总入流量/总出流量，无法直接区分内外网
- **THEN** 系统执行兜底语义并在 `extensions` 返回 `semantic_phase`、`traffic_scope`、`is_fallback`

### Requirement: 厂商扩展字段统一放入 extensions
系统 MUST 将 Azure 厂商扩展信息放入 `extensions`，且至少包含 `unit`、`cost`、`granularity`、`namespace`、`resource_region`。该规则必须与多云扩展字段策略一致，避免新增顶层字段破坏兼容性。

#### Scenario: Azure 返回扩展字段完整
- **WHEN** Azure 监控查询成功返回数据点
- **THEN** 每个数据点的 `extensions` 中至少包含 `unit`、`cost`、`granularity`、`namespace`、`resource_region`

#### Scenario: 厂商扩展字段不污染主结构
- **WHEN** 调用方解析统一监控响应
- **THEN** 厂商差异字段仅出现在 `extensions`，主结构字段集合保持稳定
