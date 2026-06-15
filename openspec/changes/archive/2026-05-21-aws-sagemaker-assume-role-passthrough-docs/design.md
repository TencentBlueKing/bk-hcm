## Context

HCM 已经新增了 AWS SageMaker AssumeRole 透传接口代码，覆盖 notebook instances、endpoints、endpoint configs、training jobs、processing jobs、transform jobs、apps、clusters、cluster nodes、training plans、training plan offerings 的 list / get。实现边界已经明确：HCM 只负责透传 AWS SageMaker 原生资源，不处理 GPU 统计、ML instance type 字典或跨资源聚合逻辑。

当前缺口不在代码本身，而在交付配套：
- 缺少面向调用方的 Markdown 接口文档
- 缺少蓝鲸 API 网关 YAML 开放接口注册
- 缺少 OpenSpec 变更工件，无法形成完整的需求/设计/任务追踪链路

这些工件都已有成熟先例：
- EC2 AssumeRole 接口文档
- CloudWatch AssumeRole 接口文档
- 既有 OpenSpec 变更归档
- `bk_apigw_resources_bk-hcm.yaml` 中已存在同类 AWS AssumeRole 开放接口注册方式

## Goals / Non-Goals

**Goals:**
- 为 18 个 SageMaker assume-role 接口补齐统一风格的 Markdown 文档
- 为 18 个接口补齐 API Gateway YAML 开放注册
- 为本次文档/配置交付补齐 OpenSpec proposal、design、specs、tasks 工件
- 在文档和 OpenSpec 中明确“透传-only”边界，避免误解为 GPU/instance-type 业务实现

**Non-Goals:**
- 不新增任何新的 SageMaker 业务接口
- 不修改现有 SageMaker handler/adaptor/client 的行为
- 不在 HCM 中新增 GPU 卡数、实例规格映射或资源聚合逻辑
- 不改造已有 AWS AssumeRole 基础能力

## Decisions

### 1. 文档按资源-动作粒度拆分，而不是合并成一份总文档

**决策**：每个 SageMaker 路由单独提供一个 Markdown 文档，命名与现有资源文档保持一致。

**理由**：
- 与现有 `docs/api-docs/web-server/docs/resource/` 目录风格一致
- 调用方按单个路由查阅更直接
- 后续若单个接口字段变化，变更范围最小

**备选方案**：
- 合并成一份 SageMaker 总文档 —— 被否决，因为不符合现有文档组织方式，且不利于 API 网关侧逐路由追踪

### 2. API 网关按开放接口逐条注册 18 个路由

**决策**：在 `bk_apigw_resources_bk-hcm.yaml` 中逐条增加 18 个 `/api/v1/cloud/vendors/aws/assume_role/sagemaker/...` 资源。

**理由**：
- 与已有 AssumeRole EC2 / CloudWatch 接口的注册方式一致
- 每个路由具备独立的 operationId、description、tag 和 backend path
- 方便后续单独授权、发布、排查问题

**备选方案**：
- 使用模糊匹配或子路径统一注册 —— 被否决，因为当前 YAML 已以显式资源为主，不应为 SageMaker 新开例外模式

### 3. OpenSpec 作为文档与接口配置变更的完整记录，而不是只记录代码变更

**决策**：为本次 SageMaker 文档和开放接口补齐独立 OpenSpec change。

**理由**：
- 本次需求明确要求补 `openspec-ff-change`
- 文档与 API 网关对外契约本身属于能力交付的一部分
- 可为后续归档提供完整上下文：为什么补文档、哪些接口对外、边界是什么

**备选方案**：
- 不补 OpenSpec，只补 docs/yaml —— 被否决，因为不满足用户明确要求，也会造成交付链断裂

## Risks / Trade-offs

- **文档示例字段与 AWS 实际返回可能存在未覆盖字段** → 在文档中明确说明 `data` 为 AWS 原始结构透传，示例只展示常用字段
- **API 网关 operationId 命名不一致风险** → 统一采用 `list_aws_assume_role_sagemaker_*` / `get_aws_assume_role_sagemaker_*` 模式
- **SageMaker 资源标签分类选择不一致** → 统一使用 `SageMaker` tag，避免将同一能力拆散到多个标签
- **后续接口字段变更导致文档过时** → 保持“一接口一文档”最小维护粒度，降低后续更新成本
