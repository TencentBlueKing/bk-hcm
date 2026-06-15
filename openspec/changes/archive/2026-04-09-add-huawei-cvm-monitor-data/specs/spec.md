# huawei-cvm-monitor-data

在统一云主机监控查询入口中新增华为云支持，允许调用方按厂商语义传参获取监控数据，并在统一基础结构上返回厂商扩展信息。

## ADDED Requirements

### Requirement: 统一入口支持华为云监控查询

系统 MUST 在资源视角接口 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 中支持 `vendor=huawei`。  
系统 MUST 保持现有腾讯云能力可用，并在同一入口下按 `vendor` 分流到对应云厂商实现。  
该能力属于 BlueKing 资源视角接口，MUST 继续复用现有 IAM 鉴权流程。

#### Scenario: 华为云查询请求被正确路由
- **GIVEN** 调用方具有目标实例的资源查看权限（IAM）
- **WHEN** 调用方请求 `POST /api/v1/cloud/vendors/huawei/cvms/monitor/data`
- **THEN** 系统 SHALL 进入华为云监控查询链路并返回查询结果

#### Scenario: 腾讯云查询行为保持兼容
- **GIVEN** 调用方请求 `POST /api/v1/cloud/vendors/tcloud/cvms/monitor/data`
- **WHEN** 请求参数合法
- **THEN** 系统 SHALL 沿用现有腾讯云逻辑处理且行为不回归

#### Scenario: 不支持的厂商请求被拒绝
- **WHEN** `vendor` 不属于已支持云厂商（Tencent Cloud/Huawei Cloud）
- **THEN** 系统 SHALL 返回 `InvalidParameter` 错误

### Requirement: 监控查询参数按厂商语义透传

系统 MUST 允许不同云厂商使用不同的监控查询参数语义，不得强制做跨云统一格式转换。  
对于华为云，系统 MUST 接受并透传符合华为云 CES 规则的参数（包括毫秒时间戳和厂商定义的 period 语义）。  
系统 MUST 支持华为云 `period=1` 的实时数据查询场景。

#### Scenario: 华为云毫秒时间戳参数透传
- **GIVEN** 请求 vendor 为 `huawei`
- **WHEN** 调用方传入符合华为云规范的毫秒时间戳参数
- **THEN** 系统 SHALL 直接按华为云语义下发查询，不进行秒/字符串等强制转换

#### Scenario: 华为云实时 period 查询
- **GIVEN** 请求 vendor 为 `huawei`
- **WHEN** 调用方传入 `period=1`
- **THEN** 系统 SHALL 受理并执行实时数据查询链路

#### Scenario: 不做跨云指标名映射
- **GIVEN** 请求 vendor 为 `huawei`
- **WHEN** 调用方传入华为云原生 `metric_name`
- **THEN** 系统 SHALL 直接使用该指标名查询，且不执行腾讯云到华为云的指标映射

### Requirement: 按账号和地域分组批量查询并聚合返回

系统 MUST 基于实例元数据中的 `account_id` 与 `region` 对目标实例分组，并按分组发起云上批量查询。  
系统 MUST 将查询结果与内部实例信息进行关联回填（`id/ip/region/instance_id`），并聚合输出。  
系统 SHOULD 在单次请求内尽可能减少对云厂商 API 的调用次数。

#### Scenario: 多账号多地域实例分组查询
- **GIVEN** 请求中包含来自多个 `account_id` 和 `region` 的实例
- **WHEN** 系统执行监控查询
- **THEN** 系统 SHALL 按 `(account_id, region)` 组合分别发起查询并合并结果

#### Scenario: 数据点与内部实例正确关联
- **GIVEN** 云厂商返回的数据点包含实例维度标识
- **WHEN** 系统完成响应组装
- **THEN** 每个数据点 SHALL 包含对应实例的 `id`、`ip`、`region` 与 `instance_id`

### Requirement: 响应结构支持基础字段与厂商扩展字段共存

系统 MUST 保持统一基础响应字段（`id/ip/region/instance_id/timestamps/values`）稳定可用。  
系统 MUST 允许在数据点中返回厂商扩展字段，用于承载基础字段无法完整表达的监控能力（例如华为云流量相关监控信息）。  
调用方在仅消费基础字段时 SHALL 保持兼容，在需要厂商特性时 MAY 读取扩展字段。

#### Scenario: 仅使用基础字段的调用方保持兼容
- **GIVEN** 调用方只解析基础响应字段
- **WHEN** 系统返回包含厂商扩展字段的响应
- **THEN** 调用方读取基础字段的行为 SHALL 不受影响

#### Scenario: 华为云扩展字段随数据点返回
- **GIVEN** 华为云监控查询结果包含厂商特有信息
- **WHEN** 系统构建响应数据点
- **THEN** 系统 SHALL 在对应数据点中返回可识别的华为云扩展字段
