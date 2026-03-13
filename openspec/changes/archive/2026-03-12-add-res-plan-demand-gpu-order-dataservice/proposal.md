## Why

资源规划场景中需要对 GPU 需求提报的主单数据进行统一管理。当前缺少 `res_plan_demand_gpu_order` 表对应的 data-service 层 CRUD 接口，无法通过微服务调用对该表进行标准化的增删改查操作，需补齐该层能力。

## What Changes

- 在 `pkg/api/data-server/` 下新增 `res_plan_demand_gpu_order` 相关的请求/响应结构体
- 在 `pkg/dal/dao/` 下新增 `res_plan_demand_gpu_order` 的 DAO 层实现（批量创建、批量更新、分页查询、批量删除）
- 在 `cmd/data-service/service/resource-plan/res_plan_demand_gpu_order/` 下新增 service 层处理逻辑
- 在 `cmd/data-service/` 路由注册中添加对应接口路由

## Capabilities

### New Capabilities

- `res-plan-demand-gpu-order-dataservice`: data-service 层 `res_plan_demand_gpu_order` 表的批量创建、批量更新、分页查询、批量删除接口

### Modified Capabilities

（无）

## Impact

- **新增文件**：`pkg/api/data-server/cloud/res_plan_demand_gpu_order.go`、DAO 实现、service 实现
- **修改文件**：`cmd/data-service/` 路由注册入口（`resource-plan` 分组下新增路由）
- **数据库**：依赖 `res_plan_demand_gpu_order` 表已存在
- **依赖服务**：data-service 内部实现，不影响其他服务
