## Why

当前 `SystemMigrate` 只能注册一个 ITSM 流程模板，切换模板需要修改代码。随着业务演进，流程模板会持续新增（增量），
且 ITSM `/system/migrate/` 接口不幂等（重复调用会报错），需要一种机制来追踪每个租户已注册到哪个模板，
实现有序、增量、可恢复的批量注册。

## What Changes

- ITSM 包暴露有序模板列表 `MigrateTemplates`，`SystemMigrate` 支持指定模板内容进行注册
- `logics_admin.InitItsmProcess` 增加进度管理：从 `global_config` 读取上次注册进度，按顺序执行未注册的模板，每次成功后更新进度
- `global_config` 新增 ITSM 类型枚举，以 per-tenant 方式（每个租户一条记录）存储迁移进度
- 新增模板时只需在 `template.go` 添加模板常量 + `MigrateTemplates` 追加一行，无需修改其他文件

## Capabilities

### New Capabilities
- `itsm-batch-migrate`: ITSM 流程模板的有序批量注册与进度管理

### Modified Capabilities

## Impact

- `pkg/criteria/enumor/global_config.go`：新增 `GlobalConfigTypeITSM` 枚举类型和 key 常量
- `pkg/thirdparty/api-gateway/itsm/process_init.go`：暴露 `MigrateTemplates` 有序列表，改造 `SystemMigrate` 方法
- `pkg/thirdparty/api-gateway/itsm/itsm.go`：`Client` 接口签名调整
- `cmd/cloud-server/logics/admin/logics_admin.go`：`InitItsmProcess` 增加读取进度 → 循环注册 → 写进度逻辑
