## Why

GPU 资源分析平台在获取到 AWS 实例列表（已由 `aws-assume-role-gpu-instance-data` 提供）后，还需要拉取实例级别的 **GPU 利用率、GPU 显存利用率、CPU 利用率**等监控指标数据，用于利用率分析和成本优化决策。AWS 的监控指标统一通过 CloudWatch `GetMetricData` API 查询，当前 HCM 完全没有对接 CloudWatch，需要新增该能力。

## What Changes

- **新增 CloudWatch client 支持**：在 `pkg/adaptor/aws/` 的 `clientSet` 中新增 `cloudWatchClient` 方法，引入 `github.com/aws/aws-sdk-go/service/cloudwatch` 依赖
- **新增 CloudWatch 指标查询能力**：在 `pkg/adaptor/aws/` 实现 `GetMetricData` 封装，支持按 Namespace + MetricName + Dimensions + 时间范围查询指标时序数据
- **新增 CloudWatch 可用指标列表查询能力**：封装 `ListMetrics` API，支持列出指定实例在 CloudWatch 中实际存在的指标（用于发现 Agent 采集了哪些 GPU 指标）
- **新增 CloudWatch 指标透传接口**：hc-service 实现指标查询 handler，入参为 `root_account_id` + `main_account_id` + `role_chain` + `region` + 指标参数，复用已有的 AssumeRole + Role Chain 链路
- **新增 CloudWatch 接口 cloud-server 入口与开放接口注册**：cloud-server 提供资源视角入口（含鉴权），在蓝鲸 API 网关 YAML 中注册为开放接口
- **新增接口文档**：按 HCM 接口文档规范编写指标查询和指标列表的接口文档

## Capabilities

### New Capabilities
- `aws-cloudwatch-metrics`: 基于 AssumeRole 的 CloudWatch 指标数据查询能力，包括 GetMetricData 时序数据查询、ListMetrics 可用指标发现，复用已有的 Role Chain 跨账号链路

### Modified Capabilities
- `aws-assume-role`: AssumeRole 角色权限需要新增 `cloudwatch:GetMetricData` 和 `cloudwatch:ListMetrics` 权限（仅前置依赖说明，不涉及 spec 变更）

## Impact

- **pkg/adaptor/aws/**：`clientSet` 新增 `cloudWatchClient` 方法；新增 `cloudwatch.go` 文件封装 GetMetricData 和 ListMetrics
- **pkg/adaptor/types/cloudwatch/**：新增 CloudWatch adaptor 层类型定义（Option、Result、Entity）
- **pkg/api/hc-service/instance-type/**：新增 `aws_cloudwatch.go`，定义 CloudWatch 指标查询和指标列表的 HTTP 请求/响应结构体
- **cmd/hc-service/service/**：新增 CloudWatch 指标查询和指标列表的 handler
- **pkg/client/hc-service/aws/**：新增 `GetMetricData` 和 `ListMetrics` client 方法
- **cmd/cloud-server/service/**：新增资源视角入口 handler
- **蓝鲸 API 网关**：在 `bk_apigw_resources_bk-hcm.yaml` 中注册 CloudWatch 开放接口
- **接口文档**：新增指标查询和指标列表接口文档
- **前置依赖**：AssumeRole 角色的 Permission Policy 需包含 `cloudwatch:GetMetricData` 和 `cloudwatch:ListMetrics` 权限；GPU 指标需要实例上已部署 CloudWatch Agent 并配置 nvidia_gpu 采集
