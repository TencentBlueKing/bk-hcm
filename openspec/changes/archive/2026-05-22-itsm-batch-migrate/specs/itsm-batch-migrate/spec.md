## ADDED Requirements

### Requirement: 有序模板列表定义
系统 SHALL 在 `itsm` 包（`pkg/thirdparty/api-gateway/itsm/process_init.go`）中定义
`MigrateTemplate` 结构体（包含 `Name string` 和 `Content string` 字段）
以及 `MigrateTemplates []MigrateTemplate` 有序列表，按注册顺序排列所有流程模板。

#### Scenario: 模板列表包含所有已有模板
- **WHEN** 系统启动时读取 `MigrateTemplates`
- **THEN** 列表 SHALL 按顺序包含 `processInitTemplate` 和 `processInitTemplate20260520`

#### Scenario: 新增模板只需追加
- **WHEN** 需要新增 ITSM 流程模板时
- **THEN** 开发者只需在 `template.go` 添加模板常量，并在 `MigrateTemplates` 末尾追加一行

### Requirement: SystemMigrate 支持指定模板内容
`itsm.Client` 接口的 `SystemMigrate` 方法 SHALL 接受 `templateContent string` 参数，
使用指定的模板内容调用 ITSM `/system/migrate/` 接口。

#### Scenario: 使用指定模板注册
- **WHEN** 调用 `SystemMigrate(kt, systemID, templateContent)` 且 templateContent 为有效的流程模板
- **THEN** 系统 SHALL 渲染模板中的 `systemID` 和 `tenantID` 变量，并上传至 ITSM

#### Scenario: ITSM 返回错误
- **WHEN** ITSM API 返回非成功状态码
- **THEN** 系统 SHALL 返回包含错误码和消息的 error

### Requirement: global_config ITSM 枚举定义
系统 SHALL 在 `pkg/criteria/enumor/global_config.go` 中新增：
- `GlobalConfigTypeITSM GlobalConfigType = "itsm"`
- `GlobalConfigKeyItsmMigrateVersionPrefix` 常量（类型 `GlobalConfigKeyITSM`），值为 `"itsm_migrate_version"`
  实际 config_key 格式为 `itsm_migrate_version_{tenantID}`，由调用方通过 `fmt.Sprintf` 拼接

#### Scenario: 枚举值可用于 global_config 查询
- **WHEN** 使用 `GlobalConfigTypeITSM` 和拼接 tenantID 的 config_key 查询 global_config
- **THEN** 系统 SHALL 返回该租户的 ITSM 迁移进度记录（如存在）

### Requirement: 新增 ApplicationType 与 WorkflowKey
系统 SHALL 在 `pkg/criteria/enumor/application.go` 中新增以下审批流程映射：
- `OperateSubAccount` → `SubAccountWorkflow`（值 `"sub_account"`）
- `ApplyPermissionPolicyLibrary` → `PermissionPolicyLibraryWorkflow`（值 `"permission_policy_library"`）
- `OperatePermissionTemplate` → `PermissionTemplateWorkflow`（值 `"permission_template"`）

#### Scenario: 新增流程类型在 ApplicationWorkflow 中注册
- **WHEN** 查询 `ApplicationWorkflow` map
- **THEN** 以上三组映射 SHALL 存在于 map 中

### Requirement: 按进度批量注册流程模板（migrateItsmTemplates）
`logics_admin` SHALL 新增私有方法 `migrateItsmTemplates(kt, systemID)`，在 `InitItsmProcess` 开头调用，实现以下流程：
1. 拼接 config_key（格式 `itsm_migrate_version_{tenantID}`）
2. 调用 `getItsmMigrateProgress` 从 `global_config` 读取进度，返回最后完成的模板名称和记录 ID
3. 在 `MigrateTemplates` 中定位起始位置（无记录则从头开始）
4. 按顺序调用 `SystemMigrate(kt, systemID, tmpl.Content)` 注册每个未执行的模板
5. 每个模板注册成功后调用 `saveItsmMigrateProgress` 更新 global_config 中的进度

#### Scenario: 首次初始化（无进度记录）
- **WHEN** global_config 中不存在该租户的 ITSM 迁移记录
- **THEN** `getItsmMigrateProgress` 返回空字符串和空 ID，系统 SHALL 从 `MigrateTemplates[0]` 开始逐个注册

#### Scenario: 部分已完成（存在进度记录）
- **WHEN** global_config 记录该租户最后完成的模板为 `processInitTemplate`
- **THEN** 系统 SHALL 跳过 `processInitTemplate`，从 `processInitTemplate20260520` 开始继续注册

#### Scenario: 全部已完成
- **WHEN** global_config 记录该租户最后完成的模板为列表中最后一个
- **THEN** `startIdx >= len(MigrateTemplates)`，系统 SHALL 不调用任何 ITSM API，直接返回 nil

#### Scenario: 注册中途失败
- **WHEN** 某个模板调用 ITSM API 失败
- **THEN** 系统 SHALL 停止后续模板注册并返回错误，已完成的模板进度已保存，下次重试时从失败的模板继续

### Requirement: ApprovalProcess 增量创建
`InitItsmProcess` 在模板注册完成后，SHALL 增量创建本地 ApprovalProcess 记录：
1. 查询该租户已有的 ApprovalProcess 列表，构建 `processMap`（key 为 `WorkflowKey`）
2. 遍历 `ApplicationWorkflow`，跳过 `processMap` 中已存在的 WorkflowKey
3. 仅对不存在的 WorkflowKey 创建新的 ApprovalProcess 记录
4. 如果所有 WorkflowKey 均已存在（`len(createItems) == 0`），直接返回 nil

#### Scenario: 首次初始化（无已有 ApprovalProcess）
- **WHEN** 该租户没有任何 ApprovalProcess 记录
- **THEN** 系统 SHALL 为 `ApplicationWorkflow` 中所有条目创建对应的 ApprovalProcess

#### Scenario: 增量新增（部分 ApprovalProcess 已存在）
- **WHEN** 该租户已有部分 ApprovalProcess 记录（如旧版本创建的），新增了新的 ApplicationType
- **THEN** 系统 SHALL 仅创建缺失的 ApprovalProcess 记录，不重复创建已存在的

#### Scenario: 全部已存在
- **WHEN** 所有 WorkflowKey 均已有对应的 ApprovalProcess 记录
- **THEN** 系统 SHALL 不执行任何创建操作，直接返回 nil

### Requirement: 进度记录持久化
每次模板注册成功后，`saveItsmMigrateProgress` SHALL 立即将该模板名称写入 global_config：
- 根据 `configID` 是否为空判断是首次写入还是更新
- 首次写入（`configID == ""`）使用 `BatchCreate`
- 后续更新（`configID != ""`）使用 `BatchUpdate`

#### Scenario: 首次创建进度记录
- **WHEN** `configID` 为空字符串，且模板注册成功
- **THEN** 系统 SHALL 调用 `BatchCreate` 创建一条新的 global_config 记录（config_type="itsm", config_key, config_value=模板名称）

#### Scenario: 更新已有进度记录
- **WHEN** `configID` 不为空，且下一个模板注册成功
- **THEN** 系统 SHALL 调用 `BatchUpdate` 更新该记录的 config_value 为最新完成的模板名称

### Requirement: 进度读取与解析（getItsmMigrateProgress）
`getItsmMigrateProgress` SHALL 从 global_config 查询进度记录，返回 `(templateName, configID, error)`：
- 查询条件：`config_type = "itsm"` AND `config_key = configKey`
- 如果记录不存在，返回空字符串和空 ID
- 如果记录存在，使用 `json.Unmarshal` 解析 `ConfigValue` 得到模板名称，同时返回记录 ID
