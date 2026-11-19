# plugin-resource-data-source Specification

## Purpose
定义 Terraform 插件执行器对资源（校验、状态升级、读取、计划、应用、导入）与数据源（校验、读取）方法的 RPC 语义、Diagnostics 错误与警告处理规则以及空状态判定规范。

## Requirements

### Requirement: Diagnostics 的 Error 使调用失败，Warning 只记日志
响应 Diagnostics 中任一条目的 severity 为 Error 时，该调用 MUST 失败。执行器 MUST NOT 因此拒绝之后另一次独立调用。severity 只有 Warning 时，调用结果 MUST NOT 标记为失败，执行器 MUST 把每条 Warning 的 summary 写入日志。空的属性路径表示问题不指向单个字段。

#### Scenario: 响应只有 Warning
- **WHEN** 插件响应的 Diagnostics 只含 severity 为 Warning
- **THEN** 调用结果不标记为失败，且该 Warning 的 summary 写入日志

#### Scenario: 响应含 Error
- **WHEN** 插件响应的 Diagnostics 含 severity 为 Error
- **THEN** 该调用失败，调用方仍能看到这些 Diagnostics

### Requirement: 类型名必须是对应 Schema 的 key
资源方法的 `TypeName` MUST 是 `resource_schemas` 的 key。数据源方法的 `TypeName` MUST 是 `data_source_schemas` 的 key。不满足时执行器 MUST 使该次调用失败，且插件侧该次 RPC 次数为 0。

#### Scenario: 资源类型名不在 Schema 中
- **WHEN** 调用方对 `ReadResource` 传入的类型名不是资源 Schema 的 key
- **THEN** 调用失败，插件侧 `ReadResource` 次数为 0

### Requirement: 资源校验或升级失败不拦住后续计划
`ValidateResourceTypeConfig` 或 `UpgradeResourceState` 的 Diagnostics 含 Error 时，该次调用 MUST 失败。执行器 MUST NOT 因此拒绝之后另一次 `PlanResourceChange`。升级成功时，解码后的状态 MUST 只出现在该次返回值里。执行器 MUST NOT 用它替换之后 `PlanResourceChange` 的 `PriorState`。请求中的升级前状态是 `RawState`，`JSON` 与 `Flatmap` 只填其中之一，执行器 MUST NOT 把它当作 `DynamicValue` 编码。

#### Scenario: 校验失败后请求计划
- **WHEN** `ValidateResourceTypeConfig` 的 Diagnostics 含 Error，调用方继续请求该类型的计划
- **THEN** 插件侧 `PlanResourceChange` 次数为 1

#### Scenario: 升级失败后请求计划
- **WHEN** 升级响应的 Diagnostics 含 Error，调用方请求该类型的计划
- **THEN** 插件侧 `PlanResourceChange` 次数为 1，且 `PriorState` 仍是调用方传入的状态

#### Scenario: 升级成功后计划仍使用调用方状态
- **WHEN** 升级响应不含 Error，调用方随后以自己的状态请求该类型的计划
- **THEN** 发往插件的 `PriorState` 是调用方传入的状态，升级结果只在升级返回值中

### Requirement: 计划只按空状态区分创建、更新和删除
`PlanResourceChange` MUST 只发出计划 RPC，MUST NOT 发出 `ApplyResourceChange`。`PriorState` 与 `ProposedNewState` 都为空时，执行器 MUST NOT 发出计划 RPC。`PriorState` 为空且 `ProposedNewState` 有值时，计划 MUST 按创建处理。两者都有值时，计划 MUST 按更新处理。`PriorState` 有值且 `ProposedNewState` 为空时，计划 MUST 按删除处理，返回的 `PlannedState` MUST 为空。`RequiresReplace` MUST 原样返回生成的属性路径，执行器 MUST NOT 把它改写成另一套结构，也 MUST NOT 在本次调用中删除或创建资源。空的状态 MUST 以空值交给插件，MUST NOT 编码成空对象。

#### Scenario: 计划不修改外部系统
- **WHEN** `PriorState` 有值且 `ProposedNewState` 有值，调用方只调用 `PlanResourceChange`
- **THEN** 插件侧只有计划 RPC，没有 `ApplyResourceChange`

#### Scenario: 新旧状态都为空
- **WHEN** `PriorState` 为空且 `ProposedNewState` 为空，调用方调用 `PlanResourceChange`
- **THEN** 插件侧计划 RPC 次数为 0

#### Scenario: 创建计划
- **WHEN** `PriorState` 为空且 `ProposedNewState` 有值，调用方调用 `PlanResourceChange`
- **THEN** 计划按创建处理

#### Scenario: 删除计划
- **WHEN** `PriorState` 有值且 `ProposedNewState` 为空，调用方调用 `PlanResourceChange`
- **THEN** 计划按删除处理，且返回的 `PlannedState` 为空

#### Scenario: RequiresReplace 只返回路径
- **WHEN** 计划响应的 `RequiresReplace` 包含字段 `name`
- **THEN** 返回路径包含 `name`，且本次调用的 `ApplyResourceChange` 次数为 0

### Requirement: 应用把插件返回的状态交给调用方
删除计划完成后，`ApplyResourceChange` 返回的 `NewState` MUST 为空。插件 Apply 返回的 `NewState` 非空且 Diagnostics 含 Error 时，调用 MUST 失败，调用方收到的 `NewState` MUST 与插件返回的字段一致。执行器 MUST NOT 在 Apply 失败时丢弃已返回的 `NewState`。

#### Scenario: 删除应用的状态为空
- **WHEN** 删除计划已完成且 `ApplyResourceChange` 返回
- **THEN** 调用方收到的 `NewState` 为空

#### Scenario: Apply 失败仍返回 NewState
- **WHEN** 插件 Apply 返回的 `NewState` 非空，且 Diagnostics 含 Error
- **THEN** 调用方收到失败，且收到的 `NewState` 与插件返回字段一致

### Requirement: 导入不创建资源也不再读取
`ImportResourceState` MUST NOT 在本方法期间调用 `ReadResource`，也 MUST NOT 创建外部资源。每一项返回资源 MUST 按其自身类型名对应的 Schema 解码。该类型名不在资源 Schema 中时，该次调用 MUST 失败，执行器 MUST NOT 改用请求类型的 Schema。外部标识指向已存在资源时，返回项 MUST 包含 `State`。外部标识在插件侧不存在时，Diagnostics MUST 含 Error，且外部系统新增资源数为 0。

#### Scenario: 导入已存在的资源
- **WHEN** 外部标识指向已存在资源并调用 `ImportResourceState`
- **THEN** 返回项包含 `State`，本方法期间 `ReadResource` 次数为 0，外部系统新增资源数为 0

#### Scenario: 返回类型不在 Schema 中
- **WHEN** 导入响应中某项的类型名不是资源 Schema 的 key
- **THEN** 该次调用失败，且该项状态没有按请求类型的 Schema 解码

#### Scenario: 导入不存在的标识
- **WHEN** 外部标识在插件侧不存在并调用 `ImportResourceState`
- **THEN** Diagnostics 含 Error，且外部系统新增资源数为 0

### Requirement: 数据源只经过校验和读取
`ValidateDataSourceConfig` 的 Diagnostics 含 Error 时，该次调用 MUST 失败。执行器 MUST NOT 因此拒绝之后另一次 `ReadDataSource`。`ReadDataSource` MUST 返回服从该数据源 Schema 的字段，且 MUST NOT 调用 `PlanResourceChange` 或 `ApplyResourceChange`。数据源路径 MUST NOT 经过计划和应用。

#### Scenario: 数据源校验失败后读取
- **WHEN** 数据源校验的 Diagnostics 含 Error，调用方请求读取
- **THEN** 插件侧 `ReadDataSource` 次数为 1

#### Scenario: 校验通过后读取数据源
- **WHEN** 查询条件通过校验并调用 `ReadDataSource`
- **THEN** 返回字段都存在于该数据源 Schema，且 `PlanResourceChange` 与 `ApplyResourceChange` 次数为 0

### Requirement: 不实现本期范围外的协议方法
执行器 MUST NOT 调用或向调用方暴露 `GetMetadata`、资源 identity、`MoveResourceState`、ephemeral 资源、`GetFunctions`、`CallFunction`、`async_apply` 和 `list`。

#### Scenario: 调用方只使用本期方法
- **WHEN** 调用方完成 Schema、配置、资源计划和数据源读取
- **THEN** 插件侧不出现 `GetMetadata`、`MoveResourceState`、ephemeral、function 或 list 的 RPC
