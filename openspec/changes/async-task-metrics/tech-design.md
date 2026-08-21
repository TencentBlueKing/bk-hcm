# 标准运维调用场景-细粒度监控补充 - 技术方案

## 1. 背景与目标

### 业务背景

目前标准运维平台执行任务流水线，调用HCM海垒平台的接口（主要是CLB负载均衡相关的接口），HCM内部收到请求后，跨多个微服务（cloud-server、task-server、hc-service）通过异步任务的方式执行业务侧的请求，在这个过程中目前缺少细粒度的监控。特别是按 `bk_biz_id` 业务维度的监控指标缺失，无法快速定位某个业务的异步任务执行状况。

### 技术目标

1. 补充从 cloud-server 入口创建异步任务的耗时监控
2. 补充异步任务执行过程的耗时监控（按 flow_name、action_name 维度）
3. 补充按 `bk_biz_id` 业务维度的失败数量统计
4. 补充任务状态变更过程监控
5. 复用现有 `pkg/metrics` 的 Prometheus 指标框架，参考 `HostRecycleSubSys` 的实现思路

### 关键约束

- 不改变现有异步任务框架的核心逻辑，仅在关键节点增加 metrics 打点
- 指标命名遵循项目已有的 `hcm_{subsystem}_{metric_name}` 规范
- 新增指标使用已有的 `AsyncSubSys` 子系统（`hcm_async_*`）
- `bk_biz_id` 信息当前不在异步任务框架的 Flow/Task 模型中，需通过 ShareData 或新增字段传递

## 2. 现有监控系统分析

### 2.1 现有指标体系

| 子系统 | 指标名 | 类型 | 标签 | 说明 |
|--------|--------|------|------|------|
| `async` | `hcm_async_task_init_queue_size` | Gauge | queue_name | 任务初始化队列大小 |
| `async` | `hcm_async_flow_type_running_num` | Gauge | flowType | 各类型运行中Flow数量 |
| `async` | `hcm_async_flow_type_exec_duration_seconds` | Histogram | flowType | 各类型Flow执行耗时 |
| `host_recycle` | `hcm_host_recycle_detect_step_cost_seconds` | Histogram | step_name | 回收预检步骤耗时 |
| `host_recycle` | `hcm_host_recycle_detect_step_err_count` | Counter | step_name | 回收预检步骤错误计数 |
| `host_recycle` | `hcm_host_recycle_order_state_cost_seconds` | Histogram | status, bk_biz_id | 回收单据状态流转耗时 |
| `host_recycle` | `hcm_host_recycle_order_state_err_count` | Counter | status, bk_biz_id | 回收单据状态流转错误计数 |
| `host_recycle` | `hcm_host_recycle_order_state_cost_since_commit_seconds` | Histogram | status, bk_biz_id | 提交到当前状态耗时 |

### 2.2 现有框架的不足

1. **缺少 bk_biz_id 维度**：异步任务框架中的 Flow/Task 模型没有 `bk_biz_id` 字段
2. **缺少创建阶段耗时**：从 cloud-server 接收请求到创建 Flow 的耗时无监控
3. **缺少任务级执行耗时**：只监控了 Flow 级耗时，缺少单个 Task/Action 的执行耗时
4. **缺少失败统计**：Flow/Task 成功或失败的 metrics 打点 TODO 尚未实现（见 `consumer/metrics.go:28` 注释）
5. **缺少状态变更过程追踪**：Flow 从 init→pending→scheduled→running→success/failed 的各阶段耗时无监控

### 2.3 参考实现：HostRecycleSubSys

HostRecycle 的 dispatcher metrics 实现思路值得借鉴：

```go
// dispatcher/metrics.go - 按 status + bk_biz_id 双维度
m.OrderStateCostSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
    Namespace: metrics.Namespace,
    Subsystem: metrics.HostRecycleSubSys,
    Name:      "order_state_cost_seconds",
    Help:      "the cost seconds of specific recycle order state",
    Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 10, 20, 30, 45, 90, ...},
}, []string{"status", "bk_biz_id"})

m.OrderStateErrCounter = prometheus.NewCounterVec(..., []string{"status", "bk_biz_id"})
```

## 3. 整体架构

### 3.1 调用链路分析

```
标准运维(SOPS) 
    │
    ▼
cloud-server (入口)
    │  1. 接收请求 /bizs/{bk_biz_id}/vendors/{vendor}/load_balancers/operations/{operation_type}/submit
    │  2. 创建 TaskManagement (含 bk_biz_id)
    │  3. 构建 CustomFlowTask 列表
    │  4. 调用 task-server CreateCustomFlow API
    │
    ▼
task-server (调度执行)
    │  1. Producer: 创建 Flow (init→pending) + Tasks
    │  2. Dispatcher: 分配 Worker (pending→scheduled)
    │  3. Scheduler: 调度执行 (scheduled→running)
    │  4. Executor: 执行 Task (pending→running→success/failed)
    │  5. WatchDog: 超时检测
    │
    ▼
hc-service (Action实际执行)
    │  1. task-server 的 Action 通过 gRPC/HTTP 调用 hc-service
    │  2. hc-service 调用云厂商 API
    │
    ▼
云厂商 API
```

### 3.2 监控打点位置

```
[A] cloud-server 入口 ──→ 创建Flow ──→ 返回FlowID
     │← metrics: flow_create_cost_seconds →│

[B] task-server Producer ──→ Flow创建完成
     │← metrics: flow_state_cost_seconds (init→pending) →│

[C] task-server Dispatcher ──→ 分配Worker
     │← metrics: flow_state_cost_seconds (pending→scheduled) →│

[D] task-server Scheduler ──→ 开始执行
     │← metrics: flow_state_cost_seconds (scheduled→running) →│

[E] task-server Executor ──→ Task执行完成
     │← metrics: task_exec_cost_seconds (per action_name) →│
     │← metrics: task_state_total (success/failed per action_name) →│

[F] Flow 执行完成
     │← metrics: flow_state_cost_since_create_seconds →│
     │← metrics: flow_state_total (success/failed/canceled) →│
```

## 4. 详细设计

### 4.1 bk_biz_id 传递方案

**方案选择**：通过 Flow 的 `ShareData` 传递 `bk_biz_id`。

**理由**：
- 不需要修改 Flow/Task 的数据库表结构
- ShareData 是 Flow 级别的共享数据，天然适合存储业务上下文信息
- cloud-server 在创建 Flow 时已经可以设置 ShareData
- 现有代码中已有通过 ShareData 传递数据的先例（如 `bpaas_sn`）

**实现**：

```go
// cloud-server 创建 Flow 时，将 bk_biz_id 写入 ShareData
shareData := tableasync.NewShareData()
shareData.Set("bk_biz_id", strconv.FormatInt(bkBizID, 10))

req := &apits.AddCustomFlowReq{
    Name:      flowName,
    ShareData: shareData,
    Tasks:     tasks,
}
```

**metrics 打点时获取**：

```go
// 从 Flow 的 ShareData 获取 bk_biz_id
func getBkBizIDFromFlow(flow *Flow) string {
    if flow.ShareData == nil {
        return "unknown"
    }
    if bizID, ok := flow.ShareData.Get("bk_biz_id"); ok {
        return bizID.(string)
    }
    return "unknown"
}
```

### 4.2 新增监控指标定义

#### 4.2.1 Flow 创建耗时（cloud-server 侧）

| 指标 | 类型 | 说明 |
|------|------|------|
| `hcm_async_flow_create_cost_seconds` | Histogram | 从 cloud-server 接收请求到创建 Flow 完成的耗时 |

**标签**：`flow_name`, `bk_biz_id`

**打点位置**：`cmd/cloud-server/service/load-balancer/` 和 `cmd/cloud-server/service/cvm/` 中调用 `CreateCustomFlow` 的前后

**Buckets**：`[]float64{0.05, 0.1, 0.2, 0.3, 0.5, 0.75, 1, 2, 3, 5, 10}`

#### 4.2.2 Flow 状态流转耗时（task-server 侧）

| 指标 | 类型 | 说明 |
|------|------|------|
| `hcm_async_flow_state_cost_seconds` | Histogram | Flow 从创建到指定状态的耗时 |
| `hcm_async_flow_state_cost_since_create_seconds` | Histogram | Flow 从创建（init）到指定终态的总耗时 |

**标签**：`state`, `flow_name`, `bk_biz_id`

**打点位置**：`pkg/async/consumer/` 中 `updateFlowState` / `updateFlowStateAndReason` 函数中

**Buckets**：`[]float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 10, 20, 30, 45, 90, 120, 180, 300, 600, 1800, 3600}`

**状态枚举**（来自 `pkg/criteria/enumor/async.go`）：

| FlowState | 值 | 说明 |
|-----------|------|------|
| FlowInit | init | 初始化（不参与调度） |
| FlowPending | pending | 待调度 |
| FlowScheduled | scheduled | 已调度 |
| FlowRunning | running | 执行中 |
| FlowCancel | canceled | 已取消 |
| FlowSuccess | success | 成功 |
| FlowFailed | failed | 失败 |

#### 4.2.3 Flow 终态统计

| 指标 | 类型 | 说明 |
|------|------|------|
| `hcm_async_flow_state_total` | Counter | Flow 到达指定终态的累计数量 |

**标签**：`state`（success/failed/canceled）, `flow_name`, `bk_biz_id`

**打点位置**：`pkg/async/consumer/scheduler.go` 中 `executeNext` 函数的 Flow 终态处理分支

#### 4.2.4 Task 执行耗时

| 指标 | 类型 | 说明 |
|------|------|------|
| `hcm_async_task_exec_cost_seconds` | Histogram | 单个 Task（Action）的执行耗时 |

**标签**：`action_name`, `flow_name`, `bk_biz_id`

**打点位置**：`pkg/async/consumer/executor.go` 中 `runTaskOnce` 函数，在 `act.Run()` 调用前后

**Buckets**：`[]float64{0.1, 0.2, 0.3, 0.5, 0.75, 1, 2, 3, 5, 10, 20, 30, 60, 120, 300, 600}`

#### 4.2.5 Task 终态统计

| 指标 | 类型 | 说明 |
|------|------|------|
| `hcm_async_task_state_total` | Counter | Task 到达指定终态的累计数量 |

**标签**：`state`（success/failed/canceled）, `action_name`, `flow_name`, `bk_biz_id`

**打点位置**：`pkg/async/consumer/executor.go` 中 `workerDo` 函数，在 Task 执行完成（成功或失败）后

**状态枚举**（来自 `pkg/criteria/enumor/async.go`）：

| TaskState | 值 | 说明 |
|-----------|------|------|
| TaskInit | init | 初始化 |
| TaskPending | pending | 待执行 |
| TaskRunning | running | 执行中 |
| TaskRollback | rollback | 回滚中 |
| TaskCancel | canceled | 已取消 |
| TaskSuccess | success | 成功 |
| TaskFailed | failed | 失败 |

#### 4.2.6 Flow 调度等待耗时

| 指标 | 类型 | 说明 |
|------|------|------|
| `hcm_async_flow_schedule_wait_seconds` | Histogram | Flow 从 pending 到 scheduled 的等待耗时 |

**标签**：`flow_name`

**打点位置**：`pkg/async/consumer/dispatcher.go` 中 Dispatcher 分配 Worker 时

**Buckets**：`[]float64{0.5, 1, 2, 5, 10, 20, 30, 60, 120, 300, 600}`

### 4.3 业务流程

#### 4.3.1 cloud-server 侧 Flow 创建监控

```
用户请求 → cloud-server 接收
    │
    ├─ 记录 start_time
    │
    ├─ 创建 TaskManagement（含 bk_biz_id）
    ├─ 构建 CustomFlowTask 列表
    ├─ ShareData.Set("bk_biz_id", bkBizID)
    ├─ 调用 task-server CreateCustomFlow API
    │
    ├─ 记录 end_time
    ├─ metrics: hcm_async_flow_create_cost_seconds.Observe(end-start)
    │           labels: {flow_name, bk_biz_id}
    │
    └─ 返回 FlowID
```

#### 4.3.2 task-server 侧 Flow 状态流转监控

```
Flow 状态变更 (updateFlowState / updateFlowStateAndReason)
    │
    ├─ 从 Flow.ShareData 获取 bk_biz_id
    ├─ 计算 Flow.CreatedAt 到当前的耗时
    │
    ├─ metrics: hcm_async_flow_state_cost_seconds.Observe(cost)
    │           labels: {state, flow_name, bk_biz_id}
    │
    ├─ 如果是终态 (success/failed/canceled):
    │   ├─ metrics: hcm_async_flow_state_cost_since_create_seconds.Observe(total_cost)
    │   │           labels: {state, flow_name, bk_biz_id}
    │   └─ metrics: hcm_async_flow_state_total.Inc()
    │               labels: {state, flow_name, bk_biz_id}
    │
    └─ 完成
```

#### 4.3.3 task-server 侧 Task 执行监控

```
Task 执行 (executor.runTaskOnce)
    │
    ├─ 记录 task_start_time
    ├─ act.Run() 执行
    ├─ 记录 task_end_time
    │
    ├─ metrics: hcm_async_task_exec_cost_seconds.Observe(end-start)
    │           labels: {action_name, flow_name, bk_biz_id}
    │
    ├─ metrics: hcm_async_task_state_total.Inc()
    │           labels: {state: success/failed, action_name, flow_name, bk_biz_id}
    │
    └─ 完成
```

### 4.4 代码修改清单

#### 4.4.1 `pkg/metrics/metric.go`

新增 SubSys 常量（可选，复用 AsyncSubSys）：

```go
// 已有 AsyncSubSys = "async"，无需新增
// 业务维度指标使用 AsyncSubSys 即可
```

#### 4.4.2 `pkg/async/consumer/metrics.go`

扩展 `metric` 结构体，新增指标定义：

```go
type metric struct {
    // 已有指标
    taskInitQueueSize  *prometheus.GaugeVec
    flowTypeRunningNum *prometheus.GaugeVec
    flowTypeExecTime   *prometheus.HistogramVec

    // 新增指标
    flowStateCostSec             *prometheus.HistogramVec  // Flow 状态流转耗时
    flowStateCostSinceCreateSec  *prometheus.HistogramVec  // Flow 从创建到终态耗时
    flowStateTotal               *prometheus.CounterVec    // Flow 终态统计
    taskExecCostSec              *prometheus.HistogramVec  // Task 执行耗时
    taskStateTotal               *prometheus.CounterVec    // Task 终态统计
    flowScheduleWaitSec          *prometheus.HistogramVec  // Flow 调度等待耗时
}
```

#### 4.4.3 `pkg/async/consumer/executor.go`

在 `runTaskOnce` 函数中添加 Task 执行耗时和终态统计打点。

在 `workerDo` 函数中添加 Task 失败统计打点。

#### 4.4.4 `pkg/async/consumer/scheduler.go`

在 `executeNext` 函数的 Flow 终态处理分支中添加 Flow 状态流转耗时和终态统计打点。

在 `runScheduledFlow` 函数中添加 Flow 从 scheduled 到 running 的状态流转耗时打点。

#### 4.4.5 `pkg/async/consumer/dispatcher.go`

在 Dispatcher 分配 Worker 时添加 Flow 调度等待耗时打点。

#### 4.4.6 `cmd/cloud-server/` 相关文件

在创建 CustomFlow 的业务代码中：
- 将 `bk_biz_id` 写入 `ShareData`
- 添加 Flow 创建耗时打点

涉及文件：
- `cmd/cloud-server/service/load-balancer/load_balancer.go`
- `cmd/cloud-server/logics/load-balancer/import_executor.go`
- `cmd/cloud-server/service/cvm/reset.go`
- `cmd/cloud-server/service/cvm/reboot_async.go`

#### 4.4.7 新增 `cmd/cloud-server/logics/async/metrics.go`

cloud-server 侧的 Flow 创建耗时指标：

```go
package async

import (
    "hcm/pkg/metrics"
    "github.com/prometheus/client_golang/prometheus"
)

var flowCreateCostSec *prometheus.HistogramVec

func InitMetrics(reg prometheus.Registerer) {
    flowCreateCostSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Namespace: metrics.Namespace,
        Subsystem: metrics.AsyncSubSys,
        Name:      "flow_create_cost_seconds",
        Help:      "The cost seconds to create an async flow from cloud-server",
        Buckets:   []float64{0.05, 0.1, 0.2, 0.3, 0.5, 0.75, 1, 2, 3, 5, 10},
    }, []string{"flow_name", "bk_biz_id"})
    reg.MustRegister(flowCreateCostSec)
}

func ReportFlowCreateCost(flowName string, bkBizID string, costSeconds float64) {
    flowCreateCostSec.WithLabelValues(flowName, bkBizID).Observe(costSeconds)
}
```

### 4.5 完整指标汇总

| 指标名 | 类型 | 标签 | Bucket范围 | 打点位置 | 说明 |
|--------|------|------|-----------|---------|------|
| `hcm_async_flow_create_cost_seconds` | Histogram | flow_name, bk_biz_id | 0.05s~10s | cloud-server | Flow创建耗时 |
| `hcm_async_flow_state_cost_seconds` | Histogram | state, flow_name, bk_biz_id | 0.1s~3600s | task-server scheduler | Flow状态流转耗时 |
| `hcm_async_flow_state_cost_since_create_seconds` | Histogram | state, flow_name, bk_biz_id | 0.1s~3600s | task-server scheduler | Flow从创建到终态总耗时 |
| `hcm_async_flow_state_total` | Counter | state, flow_name, bk_biz_id | - | task-server scheduler | Flow终态统计 |
| `hcm_async_task_exec_cost_seconds` | Histogram | action_name, flow_name, bk_biz_id | 0.1s~600s | task-server executor | Task执行耗时 |
| `hcm_async_task_state_total` | Counter | state, action_name, flow_name, bk_biz_id | - | task-server executor | Task终态统计 |
| `hcm_async_flow_schedule_wait_seconds` | Histogram | flow_name | 0.5s~600s | task-server dispatcher | Flow调度等待耗时 |

## 5. 非功能性设计

### 5.1 性能指标

- 新增 metrics 打点对业务逻辑的性能影响 < 1ms（Prometheus 客户端库的 Observe/Inc 操作为原子操作）
- `bk_biz_id` 标签的高基数问题：bk_biz_id 取值范围有限（企业内部业务数量通常 < 1000），不会造成指标爆炸

### 5.2 监控告警建议

| 告警规则 | 条件 | 级别 | 说明 |
|---------|------|------|------|
| Flow创建耗时过高 | `histogram_quantile(0.95, hcm_async_flow_create_cost_seconds) > 5` | Warning | cloud-server创建Flow耗时过长 |
| Flow执行失败率过高 | `rate(hcm_async_flow_state_total{state="failed"}[5m]) / rate(hcm_async_flow_state_total[5m]) > 0.1` | Critical | Flow失败率超过10% |
| Task执行失败率过高 | `rate(hcm_async_task_state_total{state="failed"}[5m]) / rate(hcm_async_task_state_total[5m]) > 0.1` | Critical | Task失败率超过10% |
| Flow调度等待过长 | `histogram_quantile(0.95, hcm_async_flow_schedule_wait_seconds) > 300` | Warning | Flow等待调度超过5分钟 |
| 特定业务Flow失败 | `rate(hcm_async_flow_state_total{state="failed", bk_biz_id="xxx"}[5m]) > 0` | Warning | 特定业务的Flow持续失败 |

### 5.3 Grafana 面板建议

1. **异步任务总览面板**：按 flow_name 展示各类型 Flow 的成功/失败/运行数量
2. **业务维度面板**：按 bk_biz_id 展示各业务的任务执行情况
3. **耗时分布面板**：展示 Flow 创建耗时、Task 执行耗时的 P50/P90/P95/P99
4. **状态流转面板**：展示 Flow 各状态之间的流转耗时分布

## 6. 风险与依赖

| 风险点 | 影响 | 缓解措施 |
|--------|-----|---------|
| bk_biz_id 标签高基数 | Prometheus 存储压力增大 | 限制 bk_biz_id 取值范围，异常值归为 "unknown" |
| ShareData 传递 bk_biz_id 依赖调用方设置 | 部分场景可能缺失 bk_biz_id | 缺失时使用 "unknown" 作为默认值，不影响指标采集 |
| metrics 打点增加代码复杂度 | 维护成本增加 | 封装统一的 metrics 报告函数，减少重复代码 |
| Flow 状态变更的耗时计算依赖 CreatedAt 字段 | 时钟偏差可能影响准确性 | 使用服务器本地时间戳记录，避免分布式时钟问题 |
