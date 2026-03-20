## Why

`res_plan_demand_gpu_suborder` 已完成建表，但当前缺少对应的 data-service 数据访问接口，导致 GPU 需求提报子单无法通过统一的数据服务层进行创建、查询、更新和删除，也无法复用项目内现有的 DAO 与 client 调用模式。随着 GPU 需求提报能力落地，需要尽快补齐这一层，保证新表能够按既有资源规划数据模型接入系统。

## What Changes

- 在 `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/` 目录下新增 `res_plan_demand_gpu_suborder` 的表定义、字段描述和基础校验逻辑，与建表 SQL 保持一致。
- 在 `pkg/dal/table/table.go` 中补充 `res_plan_demand_gpu_suborder` 的表名常量，接入统一表名管理。
- 在 `pkg/dal/dao/dao.go` 中新增 `ResPlanDemandGpuSubOrder()` 访问入口，使 DAO 集合能够暴露该资源的标准操作能力。
- 在 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-suborder/` 目录下新增 GPU 需求提报子单的 CRUD 服务接口，沿用现有 `res_plan_demand` 的路由、请求处理和 DAO 调用模式。
- 在 `pkg/api/data-service/resource-plan/` 增加对应的请求、响应和列表查询数据结构，提供 client 侧统一交互接口。
- 在 `pkg/client/data-service/global/resource_plan.go` 中新增对应 CRUD client 封装，供上层服务通过统一 data-service client 调用 GPU 子单接口。
- 明确 GPU 需求提报子单在 data-service 层的标准访问方式，为后续业务层或其他服务接入该表提供稳定依赖。

## Capabilities

### New Capabilities
- `res-plan-demand-gpu-suborder-data-service`: 为 GPU 需求提报子单提供标准化的 data-service CRUD 能力，包括表结构映射、服务路由和 client 交互协议。

### Modified Capabilities
- 无

## Impact

- Affected code:
  - `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/`
  - `pkg/dal/table/table.go`
  - `pkg/dal/dao/dao.go`
  - `cmd/data-service/service/resource-plan/res-plan-demand-gpu-suborder/`
  - `pkg/api/data-service/resource-plan/`
  - `pkg/client/data-service/global/resource_plan.go`
- Affected systems:
  - data-service 资源层
  - DAL/DAO 数据访问层
  - 依赖 data-service client 访问资源规划数据的上层服务
- Dependencies:
  - 依赖现有资源规划模块的 DAO、REST handler 与校验框架
  - 参照 `res_plan_demand` 的现有实现模式，保持接口风格与代码组织一致
- Breaking changes:
  - 无
