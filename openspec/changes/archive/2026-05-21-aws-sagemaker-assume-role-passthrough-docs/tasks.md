## 1. OpenSpec 工件

- [x] 1.1 创建 `openspec/changes/aws-sagemaker-assume-role-passthrough-docs/` 变更目录
- [x] 1.2 编写 `proposal.md`，说明 SageMaker 文档与开放接口配置补齐的动机和影响
- [x] 1.3 编写 `design.md`，说明文档拆分、API 网关注册和边界约束的设计决策
- [x] 1.4 编写 `specs/aws-sagemaker-assume-role-passthrough/spec.md`，定义文档与开放接口要求
- [x] 1.5 在 OpenSpec 中补充对 `aws-assume-role` 能力复用到 SageMaker 的要求说明

## 2. Markdown 接口文档

- [x] 2.1 新增 notebook instances 的 list / get 文档
- [x] 2.2 新增 endpoints 的 list / get 文档
- [x] 2.3 新增 endpoint configs 的 list / get 文档
- [x] 2.4 新增 training jobs 的 list / get 文档
- [x] 2.5 新增 processing jobs 的 list / get 文档
- [x] 2.6 新增 transform jobs 的 list / get 文档
- [x] 2.7 新增 apps 的 list / get 文档
- [x] 2.8 新增 clusters 的 list / get 文档
- [x] 2.9 新增 cluster nodes 的 list / get 文档
- [x] 2.10 新增 training plans 的 list / get 文档
- [x] 2.11 新增 training plan offerings 的 search 文档
- [x] 2.12 新增 inference components 的 list / get 文档
- [x] 2.13 新增 optimization jobs 的 list / get 文档
- [x] 2.14 新增 compute quotas 的 list / get 文档
- [x] 2.15 新增 reserved capacities 的 get 文档
- [x] 2.16 新增 reserved capacity ultra servers 的 list 文档

## 3. API 网关 YAML 配置

- [x] 3.1 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 notebook instances 的 list / get 开放接口
- [x] 3.2 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 endpoints 和 endpoint configs 的 list / get 开放接口
- [x] 3.3 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 training / processing / transform jobs 的 list / get 开放接口
- [x] 3.4 校验 YAML 中的 backend path、operationId 与 cloud-server 路由实现一致
- [x] 3.5 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 apps 的 list / get 开放接口
- [x] 3.6 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 clusters 和 cluster nodes 的 list / get 开放接口
- [x] 3.7 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 training plans 的 list / get 开放接口
- [x] 3.8 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 training plan offerings 的 search 开放接口
- [x] 3.9 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 inference components 的 list / get 开放接口
- [x] 3.10 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 optimization jobs 的 list / get 开放接口
- [x] 3.11 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 compute quotas 的 list / get 开放接口
- [x] 3.12 在 `bk_apigw_resources_bk-hcm.yaml` 中注册 reserved capacities 的 get 和 reserved capacity ultra servers 的 list 开放接口
