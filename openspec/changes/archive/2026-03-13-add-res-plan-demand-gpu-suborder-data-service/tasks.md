## 1. DAL 表定义接入

- [x] 1.1 在 `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/` 新建 `res_plan_demand_gpu_suborder` 表定义文件，并按 SQL 补齐字段、列描述和时间字段映射
- [x] 1.2 为 `res_plan_demand_gpu_suborder` 实现基础 `InsertValidate` 与 `UpdateValidate`，覆盖必填字段、长度限制和基础数值校验
- [x] 1.3 在 `pkg/dal/table/table.go` 中新增 `ResPlanDemandGpuSubOrderTable` 常量并注册到 `TableMap`

## 2. DAO 能力接入

- [x] 2.1 在资源规划 DAO 模块中新增 GPU 子单对应的 DAO 接口与实现，复用现有 `Orm`、`IDGen`、`Audit` 模式
- [x] 2.2 在 `pkg/dal/dao/dao.go` 的 `Set` 接口和 `set` 实现中新增 `ResPlanDemandGpuSubOrder()` 访问入口

## 3. data-service 协议定义

- [x] 3.1 在 `pkg/api/data-service/resource-plan/` 中新增 GPU 子单的 create 请求、update 请求和 list 请求结构
- [x] 3.2 在 `pkg/api/data-service/resource-plan/` 中补齐 GPU 子单列表返回结果和必要的校验逻辑

## 4. data-service 资源接口实现

- [x] 4.1 在 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-suborder/` 新建 service 目录及初始化文件，注册 list、batch create、batch update、batch delete 路由
- [x] 4.2 实现 GPU 子单批量创建逻辑，完成请求校验、DAO 调用和返回结果封装
- [x] 4.3 实现 GPU 子单列表查询逻辑，完成查询条件处理和列表结果返回
- [x] 4.4 实现 GPU 子单批量更新与批量删除逻辑，保持与现有资源规划资源一致的处理模式

## 5. client 封装与调用闭环

- [x] 5.1 在 `pkg/client/data-service/global/resource_plan.go` 中新增 GPU 子单的 list、batch create、batch update、batch delete client 方法
- [x] 5.2 确认新增 client 方法统一通过 `pkg/client/common/request.go` 封装调用新增 data-service 路由

## 6. 校验与收尾

- [x] 6.1 对照 `res_plan_demand` 实现检查命名、目录结构、路由风格和协议风格是否一致
- [x] 6.2 运行相关编译或静态检查，确认新增表定义、DAO、service、API、client 调用链可以正常通过验证
