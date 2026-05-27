## Why

AWS SageMaker assume-role 透传接口已经完成代码实现，但当前还缺少对外可交付的配套文档与开放接口配置，导致下游项目无法按统一规范接入和发布。现在需要把 SageMaker 接口的接口文档、API 网关 YAML，以及 OpenSpec 变更工件一并补齐，和既有 AWS EC2 / CloudWatch assume-role 能力保持一致。

## What Changes

- 新增 SageMaker assume-role 透传接口的 OpenSpec 变更工件，覆盖 proposal、design、specs、tasks
- 新增 18 个 SageMaker assume-role 接口的 Markdown 文档，覆盖 notebook instances、endpoints、endpoint configs、training jobs、processing jobs、transform jobs、apps、clusters、cluster nodes、training plans、training plan offerings 的 list / get
- 在 `docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml` 中注册 18 个 SageMaker assume-role 开放接口
- 文档中明确 HCM 的职责边界：仅提供 SageMaker 原生资源透传接口，不承担 GPU / ml instance type 业务逻辑

## Capabilities

### New Capabilities
- `aws-sagemaker-assume-role-passthrough`: AWS SageMaker AssumeRole 透传接口的资源发现、详情查询、开放接口注册与配套文档能力

### Modified Capabilities
- `aws-assume-role`: 复用既有 AssumeRole 能力到 SageMaker 资源透传场景，补充 SageMaker 使用边界与配套发布工件

## Impact

- `docs/api-docs/web-server/docs/resource/`：新增 SageMaker assume-role 接口文档
- `docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml`：新增 18 个开放接口注册
- `openspec/changes/aws-sagemaker-assume-role-passthrough-docs/`：新增 proposal、design、specs、tasks
- 下游调用方可通过 API 网关直接接入 SageMaker assume-role 透传接口
