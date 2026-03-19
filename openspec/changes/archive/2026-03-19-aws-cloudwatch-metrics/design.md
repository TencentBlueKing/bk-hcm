## Context

HCM 已通过 `aws-assume-role-gpu-instance-data` 实现了 AssumeRole 跨账号访问和 GPU 实例数据透传接口。下游 GPU 资源分析平台现在需要进一步拉取实例级别的监控指标数据（GPU 利用率、CPU 利用率等），用于利用率分析和成本优化。

AWS 的所有监控指标（无论内置还是 Agent 采集）统一通过 CloudWatch `GetMetricData` API 查询。HCM 当前完全没有对接 CloudWatch。

关键背景：
- **CPU 利用率**：AWS 内置指标，命名空间 `AWS/EC2`，无需额外配置
- **GPU 利用率**：非内置，需要实例上部署 CloudWatch Agent + NVIDIA 驱动采集，指标在 `CWAgent` 命名空间下
- **查询接口统一**：无论 CPU 还是 GPU，都通过同一个 `GetMetricData` API 查询，区别仅在 Namespace 和 MetricName
- **跨账号链路完全复用**：已有的 AssumeRole + Role Chain 机制可直接用于创建 CloudWatch client

## Goals / Non-Goals

**Goals:**
- 新增 CloudWatch client 支持（`clientSet` 中新增 `cloudWatchClient`）
- 提供指标时序数据查询透传接口（`GetMetricData`），支持按 Namespace + MetricName + InstanceId + 时间范围查询
- 提供可用指标列表查询透传接口（`ListMetrics`），支持发现指定实例实际存在的指标
- 复用已有 AssumeRole + Role Chain 链路，入参模式与 GPU 实例数据接口一致（`root_account_id` + `main_account_id` + `role_chain` + `region` + 指标参数）
- 提供 cloud-server 资源视角入口、蓝鲸 API 网关注册、接口文档
- 非破坏性：不改动任何已有接口的行为

**Non-Goals:**
- 不做指标数据的本地持久化或聚合计算（纯透传）
- 不限定具体的指标名（透传下游传入的 Namespace + MetricName，不硬编码）
- 不实现告警规则或阈值判断
- 不做多 region 聚合（下游逐 region 调用）

## Decisions

### 1. 接口设计：通用透传 vs GPU 专用接口

**决策**：提供通用的 CloudWatch 指标透传接口，不限定 Namespace 和 MetricName。

**理由**：CloudWatch 的 `GetMetricData` 本身就是通用接口，CPU 和 GPU 指标只是参数不同。做成通用透传，下游平台可以灵活查询任意指标（CPU、GPU、网络、磁盘等），避免为每种指标单独建接口。

**否决方案**：为 GPU 指标做专用接口（硬编码 `CWAgent` 命名空间 + 固定指标名）—— 不灵活，GPU 指标名取决于 Agent 配置，可能因环境不同而不同。

### 2. GetMetricData 入参设计

**决策**：下游传入 `MetricDataQueries` 数组（与 AWS API 结构对齐），每个 query 包含 Namespace、MetricName、Dimensions、Stat、Period。另外传入 `start_time` 和 `end_time` 指定时间范围。

**理由**：`GetMetricData` 支持单次查询多个指标（最多 500 个），直接对齐 AWS API 结构可以最大化灵活性，下游可以一次请求拿到一个实例的 CPU + GPU + 显存等多个指标数据。

**否决方案**：每次只查一个指标 —— 浪费 API 调用次数，对下游不友好。

### 3. ListMetrics 的必要性

**决策**：提供 `ListMetrics` 透传接口。

**理由**：GPU 指标的 MetricName 取决于 CloudWatch Agent 配置，不同环境可能不同（如 `nvidia_smi_utilization_gpu` vs `utilization_gpu`）。下游平台可以先调 `ListMetrics` 发现某实例在 `CWAgent` 命名空间下实际有哪些指标，再用正确的指标名调 `GetMetricData`。这是一个"发现"步骤，避免指标名对不上导致查不到数据。

### 4. CloudWatch adaptor 封装

**决策**：在 `pkg/adaptor/aws/client.go` 的 `clientSet` 中新增 `cloudWatchClient` 方法（与 `ec2Client`、`stsClient` 平级），新增 `pkg/adaptor/aws/cloudwatch.go` 封装业务方法，新增 `pkg/adaptor/types/cloudwatch/aws.go` 定义 adaptor 层类型。

adaptor 对外暴露两个方法：

- `(a *Aws) GetMetricData(kt, opt) ([]*MetricDataResult, error)` — 封装 `cloudwatch.GetMetricData` SDK 调用，内部处理分页和跨页数据归并（同一 query Id 可能被拆分到多页）。返回的 `MetricDataResult` 包含完整字段（`ID`、`Label`、`StatusCode`、`Messages`、`Timestamps`、`Values`），参照 GCP monitoring 的做法，不丢弃 AWS 原始响应中的元数据。Timestamps 转为 Unix 秒时间戳（int64），Values 解引用为 float64。
- `(a *Aws) ListMetrics(kt, opt) ([]*cloudwatch.Metric, error)` — 封装 `cloudwatch.ListMetrics` SDK 调用，内部处理分页。直接返回 AWS SDK 原始 `cloudwatch.Metric` 对象，完全透传，不做类型转换。

hc-service handler 通过 `AwsWithAssumeRole` 获取 adaptor client 后，调用上述方法：

- **指标查询**：handler → `client.GetMetricData(kt, opt)` → adaptor 内部调用 `cloudwatch.GetMetricDataWithContext`。handler 将 adaptor 结果映射到 API 类型（`MetricDataResultItem`，含 `label`、`status_code`、`messages` 等完整字段）后返回。
- **指标列表**：handler → `client.ListMetrics(kt, opt)` → adaptor 内部调用 `cloudwatch.ListMetricsWithContext`。handler 直接返回 raw `cloudwatch.Metric` 对象，JSON 字段名为 PascalCase（AWS SDK Go v1 无 json tag）。

**透传策略说明**：
- **ListMetrics** 采用完全透传——adaptor 返回 raw SDK 类型，handler 原样返回，client 用 `json.RawMessage` 接收。与 `instances/list` 接口一致。
- **GetMetricData** 采用 GCP monitoring 模式——由于存在跨页归并逻辑（同一 Id 的数据可能分布在多个 page），无法直接返回 raw SDK 类型，因此定义 HCM 类型但保留 AWS 响应的所有字段，不丢弃任何信息。

**理由**：遵循已有 adaptor 封装模式（与 `ListInstanceType`、`ListCvm`、`ListRegion` 等平级）：

1. `clientSet` 已有 ec2、sts、athena、s3、organizations 等 client factory，CloudWatch 只是再加一个
2. adaptor 方法封装 SDK 调用细节（参数构造、分页、类型转换），handler 只依赖 `pkg/adaptor/types/cloudwatch/` 中的类型
3. 不暴露 `GetCloudWatchClient` 给 handler — 与 `GetEC2Client` 不同，CloudWatch 没有其他地方需要直接使用 SDK client，adaptor 方法已足够

## Risks / Trade-offs

| 风险 | 缓解措施 |
| --- | --- |
| GPU 指标不存在（Agent 未部署） | 返回空数据，不报错。下游可先调 ListMetrics 确认指标是否存在 |
| 指标名与预期不符（Agent 配置差异） | ListMetrics 接口可用于发现实际指标名，下游不硬编码 |
| AssumeRole 角色缺少 CloudWatch 权限 | 前置依赖：AWS 运维在角色 Permission Policy 中加入 `cloudwatch:GetMetricData` 和 `cloudwatch:ListMetrics` |
| GetMetricData 单次请求限制（500 指标） | 下游控制每次请求的 query 数量，HCM 透传不做额外限制 |
| CloudWatch 数据保留期限（1 分钟粒度仅 15 天，5 分钟粒度 63 天，1 小时粒度 455 天） | 下游需了解并根据业务需求选择合适的时间范围和粒度 |
