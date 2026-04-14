## Why

权限策略库（permission_policy_library）是一种可复用的云厂商权限策略定义，需要持久化存储并提供标准 CRUD 管理能力。目前缺少对应的数据层接口，上层服务无法通过 data-service 对策略库进行统一管理。

## What Changes

- 新增 `permission_policy_library` 表对应的 ORM 表结构定义
- 新增 DAL DAO 层接口与实现（Create、Update、List、BatchDelete）
- 新增 data-service HTTP handler，暴露 CRUD 接口
- 新增 SDK 客户端（tcloud vendor 写操作 + global 读/删操作）

## Capabilities

### New Capabilities

- `permission-policy-library-crud`: 权限策略库的增删改查能力，包括创建（含 vendor）、更新（自动维护 policy_hash 和 version）、批量删除、列表查询、单条获取

### Modified Capabilities

（无）

## Impact

- `pkg/dal/table/cloud/` — 新增表结构文件
- `pkg/dal/table/table.go` — 注册新表名常量
- `pkg/dal/dao/cloud/` — 新增 DAO 实现
- `pkg/dal/dao/dao.go` — 注册 DAO 到 Set 接口
- `pkg/dal/dao/types/` — 新增 ListDetails 类型
- `pkg/api/core/cloud/` — 新增 core 类型
- `pkg/api/data-service/cloud/` — 新增请求/响应结构
- `cmd/data-service/service/cloud/` — 新增 handler
- `cmd/data-service/service/service.go` — 注册新服务
- `pkg/client/data-service/global/` — 新增 global SDK 客户端
- `pkg/client/data-service/tcloud/` — 新增 tcloud vendor SDK 客户端
