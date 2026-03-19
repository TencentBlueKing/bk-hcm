# Spec: aws-cloudwatch-metrics

## Purpose

TBD — AWS CloudWatch 指标查询能力，提供 GetMetricData 时序数据查询和 ListMetrics 可用指标发现的透传接口，支持通过 AssumeRole 跨账号访问 CloudWatch。

## Requirements

### Requirement: CloudWatch client 支持
`clientSet` SHALL 新增 `cloudWatchClient(region string)` 方法，返回 `*cloudwatch.CloudWatch` client。`Aws` 结构体 SHALL 新增 `GetCloudWatchClient(region string)` 公开方法，与已有的 `GetEC2Client` 模式一致。仅涉及 AWS 云厂商。

#### Scenario: 创建 CloudWatch client
- **WHEN** 以有效凭证和 region 调用 `GetCloudWatchClient`
- **THEN** 返回可用的 CloudWatch client，不报错

#### Scenario: region 为空
- **WHEN** region 为空字符串
- **THEN** 使用 SDK 默认 region 行为（与 ec2Client 一致）

### Requirement: GetMetricData 指标时序查询
hc-service SHALL 提供 CloudWatch 指标时序数据查询接口，接收 `root_account_id`（根账号 HCM 内部 ID）、`main_account_id`（二级账号 HCM 内部 ID）、`role_chain`（角色名数组）、`region`、`metric_data_queries`（指标查询数组）、`start_time`、`end_time` 参数。系统通过 `root_account_id` 调 `AwsRoot()` 获取根账号 AK/SK，通过 `main_account_id` 查 `main_account` 表获取目标成员账号 CloudID，再通过 `AwsWithAssumeRole` 获取成员账号权限后，调用 CloudWatch `GetMetricData` API 并返回时序数据点。参考 GCP GPU monitoring 的入参模式。该接口为 AWS 数据透传，不持久化到本地数据库。仅涉及 AWS 云厂商。

#### Scenario: 查询单个 CPU 利用率指标
- **GIVEN** root_account_id、main_account_id、role_chain、region 均有效，metric_data_queries 包含 1 个查询（Namespace=AWS/EC2, MetricName=CPUUtilization, Dimensions=[InstanceId=i-xxx]）
- **WHEN** 调用指标查询接口
- **THEN** 返回指定时间范围内的 CPUUtilization 时序数据点（Timestamps + Values）

#### Scenario: 查询 GPU 利用率指标
- **GIVEN** 目标实例已部署 CloudWatch Agent 且配置了 nvidia_gpu 采集
- **WHEN** 以 Namespace=CWAgent, MetricName=nvidia_smi_utilization_gpu 查询
- **THEN** 返回 GPU 利用率时序数据点

#### Scenario: 单次查询多个指标
- **GIVEN** metric_data_queries 包含多个查询（如 CPUUtilization + GPU utilization + GPU memory）
- **WHEN** 调用指标查询接口
- **THEN** 返回每个 query 对应的时序数据，通过 query id 区分

#### Scenario: 指标不存在
- **GIVEN** 查询的指标在 CloudWatch 中不存在（如 Agent 未部署导致无 GPU 指标）
- **WHEN** 调用指标查询接口
- **THEN** 对应 query 返回空数据点列表，不报错

#### Scenario: main_account_id 无效
- **GIVEN** main_account_id 在 main_account 表中不存在
- **WHEN** 调用指标查询接口
- **THEN** 返回错误，提示该 main_account_id 无效

#### Scenario: AssumeRole 失败
- **GIVEN** Role Chain 中任一步失败
- **WHEN** 调用指标查询接口
- **THEN** 返回 AssumeRole 相关错误信息

### Requirement: ListMetrics 可用指标发现
hc-service SHALL 提供 CloudWatch 可用指标列表查询接口，接收 `root_account_id`、`main_account_id`、`role_chain`、`region`、`namespace`（可选）、`metric_name`（可选）、`dimensions`（可选过滤条件）参数。系统通过 `root_account_id` 调 `AwsRoot()` 获取根账号 AK/SK，通过 `main_account_id` 查 `main_account` 表获取目标成员账号 CloudID，再通过 `AwsWithAssumeRole` 获取权限后，调用 CloudWatch `ListMetrics` API 并返回匹配的指标列表。参考 GCP GPU monitoring 的入参模式。该接口为 AWS 数据透传。仅涉及 AWS 云厂商。

#### Scenario: 列出实例的所有 CWAgent 指标
- **GIVEN** 目标实例已部署 CloudWatch Agent
- **WHEN** 以 Namespace=CWAgent, Dimensions=[InstanceId=i-xxx] 查询
- **THEN** 返回该实例在 CWAgent 命名空间下的所有指标名列表（如 nvidia_smi_utilization_gpu、nvidia_smi_memory_used 等）

#### Scenario: 列出所有命名空间的指标
- **GIVEN** 不指定 namespace 过滤
- **WHEN** 调用可用指标列表接口
- **THEN** 返回该 region 下所有命名空间的指标（可能量大，下游应合理使用过滤条件）

#### Scenario: 无匹配指标
- **GIVEN** 查询条件无匹配（如 Agent 未部署导致 CWAgent 命名空间下无指标）
- **WHEN** 调用可用指标列表接口
- **THEN** 返回空列表，不报错

### Requirement: CloudWatch 接口 cloud-server 资源视角入口
cloud-server SHALL 提供 GetMetricData 和 ListMetrics 的资源视角入口 handler。入口 handler SHALL 执行鉴权后，通过微服务调用 hc-service 对应接口。路由注册于 cloud-server 的资源视角路由组。仅涉及 AWS 云厂商。

#### Scenario: cloud-server 路由注册
- **GIVEN** cloud-server 启动
- **WHEN** 初始化资源视角路由
- **THEN** `/vendors/aws/assume_role/cloudwatch/metric_data/get` 和 `/vendors/aws/assume_role/cloudwatch/metrics/list` 两个路由 SHALL 被注册

#### Scenario: 鉴权通过后转调 hc-service
- **GIVEN** 下游平台通过 API 网关调用 cloud-server CloudWatch 接口
- **WHEN** 鉴权通过
- **THEN** cloud-server handler 调用 hc-service 的对应接口并返回结果

### Requirement: CloudWatch 接口蓝鲸 API 网关注册
GetMetricData 和 ListMetrics 接口 SHALL 在蓝鲸 API 网关 YAML（`bk_apigw_resources_bk-hcm.yaml`）中注册为开放接口，供下游平台通过 API 网关调用。仅涉及 AWS 云厂商。

#### Scenario: API 网关注册
- **GIVEN** 部署 API 网关配置
- **WHEN** 加载 `bk_apigw_resources_bk-hcm.yaml`
- **THEN** GetMetricData 和 ListMetrics 两个接口 SHALL 出现在开放接口列表中

### Requirement: CloudWatch 接口文档
SHALL 按 HCM 接口文档规范（`docs/api-docs/web-server/docs/resource/`）编写 GetMetricData 和 ListMetrics 的接口文档，包含请求参数、响应结构、示例。仅涉及 AWS 云厂商。

#### Scenario: 接口文档完整
- **GIVEN** 开发完成 CloudWatch 透传接口
- **WHEN** 检查接口文档
- **THEN** `get_aws_assume_role_metric_data.md` 和 `list_aws_assume_role_metrics.md` 两个文档 SHALL 存在且包含请求参数、响应结构、示例
