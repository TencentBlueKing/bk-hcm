## Context

当前 woa-server 的 GPU 需求相关业务逻辑（Excel导入预览、创建主单/子单、覆盖上传）全部实现在 `cmd/woa-server/logics/plan/gpu_demand_order.go` 的 `Controller` 结构体上，并通过 `plan.Logics` 接口暴露给 service 层调用。

然而 logics 层在项目中的定位是存放**跨模块公共逻辑**（如 fetcher、dispatcher、penalty 计算等），GPU 需求提报是独立的业务流程，不被其他模块复用，不应放在 logics 层。

service 层的 `service` 结构体已经持有 `client *client.ClientSet`，具备直接调用 data-service 的能力，无需绕道 logics 层。

## Goals / Non-Goals

**Goals:**

- 将 `ExcelImportGpuDemand`、`CreateGpuDemandOrder`、`OverwriteGpuDemandOrder` 三个方法及其所有辅助函数从 logics 层迁移到 service 层
- 从 `plan.Logics` 接口中移除这三个方法定义
- 删除 logics 层中 `gpu_demand_order.go` 文件（迁移完成后不保留空文件）
- 优化迁移后函数的命名和注释，使其更简洁清晰
- 保持对外 API 接口路径、请求/响应结构完全不变

**Non-Goals:**

- 不修改 data-service 层的接口或数据模型
- 不改变任何业务逻辑（校验、创建、删除、审计等流程不变）
- 不迁移 logics 层中其他非 GPU 相关的方法
- 不调整对外 REST API 路由

## Decisions

### 1. 将业务方法直接实现为 service 结构体的方法

**选择**: 将三个公开方法和所有辅助函数作为 `service` 结构体的方法（或包级函数），直接在 service 层内完成全部逻辑。

**理由**: service 结构体已持有 `client`、`dao`、`authorizer` 等依赖，迁移后方法可直接使用这些字段调用 data-service，无需引入新的依赖注入机制。相比创建新的中间层或 interface，直接放在 service 方法上最简洁。

**替代方案**: 在 service 层创建独立的 `GpuDemandService` 结构体 —— 增加了不必要的复杂度，当前 GPU 需求功能体量不大。

### 2. 拆分为独立文件 `gpu_demand_order.go`

**选择**: 在 `cmd/woa-server/service/plan/` 下新建 `gpu_demand_order.go` 存放迁移后的业务逻辑方法，现有 `gpu_demand_excel.go` 已存放 handler 函数，继续保留。

**理由**: handler（请求解析/鉴权/响应封装）和业务逻辑（校验/创建/删除）职责分离，文件命名与 logics 层保持一致，便于代码 review 时对照。

### 3. 保留无状态辅助函数为包级函数

**选择**: `validateOrderDetails`、`validateDetailFixedFields`、`validateDetailExtension`、`buildDetails`、`convertCellValue`、`convertEnumValue` 等不依赖 `Controller`/`service` 字段的函数，保持为包级函数（不绑定到 service 结构体）。

**理由**: 这些函数是纯逻辑函数，不需要访问实例状态，作为包级函数更清晰，也更容易单元测试。

### 4. handler 层调用方式从 planController 间接调用改为直接调用 service 内部方法

**选择**: `gpu_demand_excel.go` 中的 handler 函数原来通过 `s.planController.ExcelImportGpuDemand(...)` 调用，迁移后改为 `s.parseGpuDemandExcel(...)` 等直接调用。

**理由**: 业务逻辑已在 service 自身，无需再通过 interface 间接调用。

## Risks / Trade-offs

- **[风险] logics.Logics 接口变更可能影响 mock/测试** → 检查是否有其他代码 mock 了 `Logics` 接口中的这三个方法，如有需同步修改。实际上这三个方法是新增的，被 mock 的可能性极低。
- **[风险] 辅助函数依赖 `Controller` 字段** → 需仔细检查 `getLatestGpuTemplate`、`createGpuDemandOrder` 等方法的 `c.client` 引用，迁移后改为 `s.client`。
- **[权衡] 单文件体量** → 迁移后 `gpu_demand_order.go` 预计约 500 行，接近但未超过 800 行限制。如后续继续增长，可再拆分。
