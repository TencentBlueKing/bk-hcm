## Why

GPU需求提报功能需要一张模版配置表 `res_plan_demand_gpu_template` 来管理Excel模版的结构定义。当前缺少从DAO层到data-service层的完整CRUD代码，导致上层服务无法对模版进行增删改查操作。

## What Changes

- 新增 `res_plan_demand_gpu_template` 表的table定义（结构体、列描述、校验逻辑）
- 新增该表的DAO层接口与实现（CreateWithTx、UpdateWithTx、List、DeleteWithTx）
- 新增data-service层的API类型定义（请求/响应结构体）
- 新增data-service层的HTTP handler（批量创建、列表查询、批量更新、删除）
- 在dao.go中注册新DAO、在resource_plan.go中注册新service

## Capabilities

### New Capabilities
- `gpu-template-crud`: GPU需求提报模版配置表的完整CRUD数据访问与服务层实现

### Modified Capabilities

## Impact

- `pkg/dal/table/table.go`：新增表名常量
- `pkg/dal/table/resource-plan/`：新增table定义目录
- `pkg/dal/dao/resource-plan/`：新增DAO文件
- `pkg/dal/dao/dao.go`：Set接口新增方法、set结构体新增实现
- `pkg/api/data-service/resource-plan/`：新增API类型定义
- `cmd/data-service/service/resource-plan/`：新增handler目录，修改resource_plan.go注册入口
