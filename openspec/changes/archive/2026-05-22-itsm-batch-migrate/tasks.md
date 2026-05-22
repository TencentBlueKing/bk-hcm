## 1. 枚举常量定义

- [x] 1.1 在 `pkg/criteria/enumor/global_config.go` 中新增 `GlobalConfigTypeITSM GlobalConfigType = "itsm"` 枚举值
- [x] 1.2 在 `pkg/criteria/enumor/global_config.go` 中新增 `GlobalConfigKeyItsmMigrateVersion` 常量，值为 `"itsm_migrate_version"`

## 2. ITSM 包改造

- [x] 2.1 在 `pkg/thirdparty/api-gateway/itsm/process_init.go` 中定义 `MigrateTemplate` 结构体（Name、Content 字段）
- [x] 2.2 在 `pkg/thirdparty/api-gateway/itsm/process_init.go` 中定义 `MigrateTemplates` 有序列表，包含 `processInitTemplate` 和 `processInitTemplate20260520`
- [x] 2.3 修改 `pkg/thirdparty/api-gateway/itsm/itsm.go` 中 `Client` 接口的 `SystemMigrate` 方法签名，新增 `templateContent string` 参数
- [x] 2.4 修改 `pkg/thirdparty/api-gateway/itsm/process_init.go` 中 `SystemMigrate` 实现，使用传入的 `templateContent` 替代硬编码模板

## 3. 进度管理逻辑

- [x] 3.1 在 `cmd/cloud-server/logics/admin/logics_admin.go` 的 `InitItsmProcess` 中，添加从 global_config 读取当前租户 ITSM 迁移进度的逻辑
- [x] 3.2 实现根据进度定位 `MigrateTemplates` 起始索引的逻辑（无记录从 0 开始，有记录从 lastApplied+1 开始）
- [x] 3.3 实现循环调用 `SystemMigrate` 注册未执行模板的逻辑
- [x] 3.4 实现每次模板注册成功后更新/创建 global_config 进度记录的逻辑（首次 BatchCreate，后续 BatchUpdate）
