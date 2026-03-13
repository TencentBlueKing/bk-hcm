## ADDED Requirements

### Requirement: GPU 需求子单表 SHALL 具备独立的数据表映射
系统 MUST 为 `res_plan_demand_gpu_suborder` 提供独立的数据表映射，表定义文件 MUST 放在 `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/` 子目录下，并与建表 SQL 中的字段、主键、索引相关字段语义保持一致。该映射 MUST 支持资源规划模块现有的数据校验和列描述机制，以便 DAO 与 data-service 可以统一访问该表。

#### Scenario: 新表映射接入资源规划表定义目录
- **GIVEN** `res_plan_demand_gpu_suborder` 已在数据库中完成建表
- **WHEN** 开发者为该表补齐 DAL 层定义
- **THEN** 系统 MUST 在 `pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder/` 下提供可被资源规划模块引用的表结构、列描述和基础校验定义

### Requirement: GPU 需求子单资源 SHALL 通过统一表名与 DAO 入口暴露
系统 MUST 在 `pkg/dal/table/table.go` 中定义 `res_plan_demand_gpu_suborder` 的表名常量并注册到统一表配置中，同时 MUST 在 `pkg/dal/dao/dao.go` 中提供 `ResPlanDemandGpuSubOrder()` 访问入口，使上层服务能够按现有 `dao.Set` 方式获取该资源的标准 DAO 能力。

#### Scenario: 上层服务通过 dao.Set 访问 GPU 子单资源
- **GIVEN** data-service 资源规划模块依赖 `dao.Set` 注入 DAO 能力
- **WHEN** service 初始化或处理 GPU 子单请求时需要访问该表
- **THEN** 系统 MUST 能通过 `ResPlanDemandGpuSubOrder()` 获取该资源的标准 DAO 实例，而无需直接依赖具体实现类型

### Requirement: GPU 需求子单 SHALL 提供独立的 data-service CRUD 接口
系统 MUST 为 GPU 需求子单提供独立的 data-service CRUD 接口，并将 service 文件组织在 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-suborder/` 子目录下。该资源 MUST 至少支持批量创建、批量更新、批量删除和列表查询，并沿用现有资源规划模块的 REST 路由风格与请求处理模式。

#### Scenario: 上层模块查询 GPU 子单列表
- **GIVEN** 调用方需要按资源规划模式查询 GPU 需求子单
- **WHEN** 调用方请求 GPU 子单的 list 接口
- **THEN** 系统 MUST 通过独立的 data-service 资源路由返回符合列表协议的数据结果

#### Scenario: 上层模块批量变更 GPU 子单
- **GIVEN** 调用方需要批量创建、更新或删除 GPU 需求子单
- **WHEN** 调用方访问对应的 batch create、batch patch 或 batch delete 接口
- **THEN** 系统 MUST 按现有资源规划资源的一致约定处理请求并完成对应 CRUD 操作

### Requirement: GPU 需求子单 SHALL 提供统一的 data-service API 协议与 client 封装
系统 MUST 在 `pkg/api/data-service/resource-plan/` 中定义 GPU 需求子单对应的 create、update、list 等交互协议，并 MUST 在 `pkg/client/data-service/global/resource_plan.go` 中提供对应的 CRUD client 方法。上层服务使用该资源时 MUST 通过统一的 data-service client 封装发起请求，而不是自行拼接 REST 请求。

#### Scenario: 上层服务通过 global client 调用 GPU 子单接口
- **GIVEN** 业务模块需要通过 data-service 访问 GPU 需求子单资源
- **WHEN** 业务模块调用资源规划 global client
- **THEN** 系统 MUST 提供与 GPU 子单 CRUD 接口一一对应的 client 方法，并通过统一请求封装完成调用
