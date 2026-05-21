## ADDED Requirements

### Requirement: SageMaker AssumeRole 接口文档
HCM SHALL 按现有 AWS AssumeRole 文档规范，在 `docs/api-docs/web-server/docs/resource/` 下为 SageMaker assume-role 透传接口提供独立 Markdown 文档。文档范围 SHALL 覆盖 notebook instances、endpoints、endpoint configs、training jobs、processing jobs、transform jobs、apps、clusters、cluster nodes、training plans、training plan offerings、training plans 的 list / get，training plan offerings 的 search，以及 inference components、optimization jobs、compute quotas、reserved capacities、reserved capacity ultra servers 等 GPU 统计关键来源接口共 29 个接口。每份文档 SHALL 包含接口描述、URL、请求参数、调用示例、响应参数说明，并明确 `data` 为 AWS SageMaker 原始结构透传。仅涉及 AWS 云厂商。

#### Scenario: 文档文件完整存在
- **WHEN** 检查 `docs/api-docs/web-server/docs/resource/` 目录
- **THEN** 可以找到 29 个 SageMaker assume-role 接口文档文件，分别对应 GPU 统计关键资源的 list / get/search 路由

#### Scenario: 文档声明透传边界
- **WHEN** 调用方阅读任一 SageMaker 接口文档
- **THEN** 文档明确说明返回体中的 `data` 是 AWS SageMaker 原始结构透传，HCM 不在该接口中做 GPU 或实例规格业务推导

### Requirement: SageMaker AssumeRole 开放接口注册
SageMaker assume-role 透传接口 SHALL 在蓝鲸 API 网关 YAML（`docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml`）中逐条注册为开放接口。注册范围 SHALL 覆盖 `/api/v1/cloud/vendors/aws/assume_role/sagemaker/...` 下的 29 个 list / get/search 路由，并为每个路由配置独立的 operationId、description、tag、backend path 和 timeout。仅涉及 AWS 云厂商。

#### Scenario: API 网关资源完整注册
- **WHEN** 检查 `bk_apigw_resources_bk-hcm.yaml`
- **THEN** 29 个 SageMaker assume-role 路由均已出现，并指向对应的 `/api/v1/cloud/vendors/aws/assume_role/sagemaker/...` backend path

#### Scenario: 路由注册与实现一致
- **WHEN** 对照 YAML 中的 SageMaker resource path 和 cloud-server 实际注册路由
- **THEN** 两者路径一一对应，不存在缺失、拼写错误或动作不一致
