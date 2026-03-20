## Why

GPU需求相关的三个业务函数（`ExcelImportGpuDemand`、`CreateGpuDemandOrder`、`OverwriteGpuDemandOrder`）当前实现在 `logics/plan` 层的 `Controller` 上，而 logics 层通常用于放置跨模块的公共逻辑。这三个函数是纯粹的 GPU 需求业务流程处理，应该放在 `service/plan` 层，使代码架构更加清晰、职责划分更加合理。

## What Changes

- 将 GPU 需求相关的三个方法实现从 `cmd/woa-server/logics/plan/gpu_demand_order.go` 迁移到 `cmd/woa-server/service/plan/` 下的新文件
- 从 `logics/plan.Logics` 接口中移除这三个方法定义
- 在 `service/plan/service` 中直接通过 `client.DataService()` 调用 data-service 接口，不再通过 logics 层中转
- service 层 handler 函数中去掉对 `planController` 的间接调用，改为直接调用 service 内部方法
- 优化函数命名和注释，使其更简洁明了地表达功能

## Capabilities

### New Capabilities

_无新增能力，仅做架构调整_

### Modified Capabilities

- `gpu-demand-excel-import`: 实现层从 logics 迁移到 service，接口行为不变
- `create-gpu-demand-order`: 实现层从 logics 迁移到 service，接口行为不变
- `overwrite-gpu-demand-order`: 实现层从 logics 迁移到 service，接口行为不变

## Impact

- **Service Layer**: `cmd/woa-server/service/plan/` — 新增 GPU 需求业务逻辑文件，service handler 改为直接调用内部方法
- **Logics Layer**: `cmd/woa-server/logics/plan/` — 移除 `Logics` 接口中的三个 GPU 方法定义，删除 `gpu_demand_order.go` 中的对应实现
- **API/协议**: 无变化，对外接口路径和请求/响应结构不变
- **依赖关系**: service 层需新增对 `client.DataService()` 和 `tools/excel` 包的直接依赖
