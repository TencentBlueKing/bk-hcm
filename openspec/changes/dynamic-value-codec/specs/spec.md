# dynamic-value-codec

## ADDED Requirements

### Requirement: 按 Schema 把调用方对象编码为 msgpack DynamicValue
系统 MUST 提供满足 `executor.ValueCodec` 的 `Encode`。入参 MUST 是调用方对象和一份 `*tfplugin5.Schema`。成功时 MUST 返回只写入 `Msgpack` 的 `*tfplugin5.DynamicValue`。系统 MUST 把该 Schema 的顶层块转成对象类型：属性的 `Type` 字节按 TF JSON 解析，嵌套块按 `NestingMode` 并入同一对象。`SINGLE` 与 `GROUP` MUST 对应对象，`LIST` MUST 对应 list，`SET` MUST 对应 set，`MAP` MUST 对应 map。调用方侧的对象和 map MUST 使用 `map[string]any`，list 和 set MUST 使用 `[]any`。`min_items`、`max_items` 和 Schema 的 `version` MUST NOT 改变编码结果。

#### Scenario: 可选字段未设置
- **WHEN** 入参对象的字段类型与 Schema 一致，且一个可选属性既没有出现在 map 中，值也不是未知值哨兵
- **THEN** 系统返回 `DynamicValue`，该属性在带类型的值中为 null

#### Scenario: 嵌套 list 与 set 使用各自的块类型
- **WHEN** Schema 含有一个 `LIST` 嵌套块和一个 `SET` 嵌套块，调用方分别传入 `[]any`
- **THEN** 两段编码分别使用 list 和 set 的类型，调用方形状都是 `[]any`

#### Scenario: 嵌套模式无效
- **WHEN** 某个嵌套块的 `NestingMode` 为 `INVALID`
- **THEN** 系统不返回 `DynamicValue`，并返回错误

### Requirement: 缺省值为 null，未知值只来自哨兵
键不存在或值为 `nil` 时，系统 MUST 把该字段编码为其所在类型的 null。`required` 字段缺失时 MUST 同样编码为 null，MUST NOT 在编解码中另报缺少必填。系统 MUST 导出固定哨兵表示未知值。调用方传入该哨兵时，系统 MUST 把对应字段编码为未知值，MUST NOT 编成 null 或空字符串。系统 MUST NOT 因为属性带有 `computed` 就把缺省字段改成未知值。

#### Scenario: 显式 nil 与缺省相同
- **WHEN** 同一可选属性在一份对象里缺省，在另一份对象里值为 `nil`
- **THEN** 两份编码里该属性都是 null

#### Scenario: Computed 字段当前未知
- **WHEN** 一个 `computed` 属性的当前值是未知值哨兵
- **THEN** 该属性编码为未知值，不是 null，也不是空字符串

#### Scenario: Computed 字段缺省
- **WHEN** 一个 `computed` 属性没有出现在调用方对象中
- **THEN** 该属性编码为 null，不是未知值

### Requirement: 未声明字段和类型不匹配在构造前失败
系统 MUST 在构造会 panic 的带类型值之前完成校验。Schema 未声明的字段 MUST 使本次编码失败，且 MUST NOT 返回 `DynamicValue`。错误文本 MUST 包含从根开始的属性路径。list 或 set 元素上的错误 MUST 包含下标，例如 `network[0].name`。类型与 Schema 不符时 MUST 返回错误，MUST NOT 把该值交给会 panic 的构造函数。错误文本 MUST NOT 包含字段值。属性名与嵌套块名冲突，或属性类型为 `dynamic`，或类型字节无法解析时，系统 MUST 失败并带上属性路径。

#### Scenario: 未声明字段
- **WHEN** 入参对象包含 Schema 未声明的字段 `extra`
- **THEN** 系统不返回 `DynamicValue`，错误信息包含 `extra`，且不包含该字段的值

#### Scenario: 元素类型不匹配
- **WHEN** list 元素的类型与 Schema 不符
- **THEN** 系统不返回 `DynamicValue`，错误信息包含该元素的路径

#### Scenario: dynamic 类型
- **WHEN** 某个属性的 TF JSON 类型是 `dynamic`
- **THEN** 系统不返回 `DynamicValue`，错误信息包含该属性路径

### Requirement: 按同一份 Schema 把 DynamicValue 解回调用方对象
`Decode` MUST 使用调用方传入的同一份 Schema 确定类型，MUST NOT 从 msgpack 字节推断 List 与 Set、Object 与 Map 的差别。系统 MUST 只读取 `Msgpack`。Schema 为空、顶层块为空，或只有 `Json` 而没有 `Msgpack` 时，系统 MUST 返回错误，MUST NOT 产出调用方对象。null MUST 解为 `nil`，未知值 MUST 解为同一个哨兵。string 与 bool MUST 保持原类型。number MUST 解为 `*big.Float`。list、set 和 tuple MUST 解为 `[]any`。object 与 map MUST 解为 `map[string]any`。

#### Scenario: 同一 Schema 往返
- **WHEN** 一个只含已知字段的对象按一份 Schema 编码后再用同一份 Schema 解码
- **THEN** 已知字段与编码前一致；number 以 `*big.Float` 的数值相等为准

#### Scenario: 用另一种集合 Schema 解码
- **WHEN** 一段由 set 编出的 msgpack 改用元素类型相同的 list Schema 解码
- **THEN** 系统按 list 返回 `[]any`，不从字节恢复 set

#### Scenario: 缺少 Schema
- **WHEN** 解码时 Schema 或顶层块为空
- **THEN** 系统不产出调用方对象

#### Scenario: 只有 Json
- **WHEN** `DynamicValue` 的 `Msgpack` 为空且 `Json` 非空
- **THEN** 系统返回错误，不产出调用方对象

### Requirement: 集合往返时保留 list 顺序并忽略 set 顺序
系统 MUST 在 list 往返后保持元素顺序。set 往返后的元素顺序 MUST NOT 作为相等条件。比较 set 时 MUST 只要求元素集合一致。

#### Scenario: list 顺序不变
- **WHEN** 一个含两个以上元素的 list 按同一 Schema 编码后再解码
- **THEN** 解码后的元素顺序与编码前相同

#### Scenario: set 顺序变化仍视为相同
- **WHEN** 一个 set 往返后元素顺序与编码前不同，元素本身相同
- **THEN** 系统把这组元素视为与编码前一致

### Requirement: 状态和配置的根是 map[string]any
资源状态、Provider 配置和数据源结果的根 MUST 是 `map[string]any`。空状态 MUST 使用 `nil`。未知值 MUST 只出现在该 map 的属性中。`UpgradeResult.UpgradedState`、`ReadResourceResult.NewState`、`PlanResult.PlannedState`、`ApplyResult.NewState`、`ImportedResource.State`、`ReadDataSourceResult.State`，以及 `Executor` 上对应的 config、current、prior、proposed、planned 入参，MUST 使用 `map[string]any`。`Attribute.Type` MUST 保持 `any`，因为 TF JSON 类型是字符串或数组，不是状态对象。`ValueCodec` 的方法签名 MUST 保持 `any`。

#### Scenario: Apply 返回有状态的资源
- **WHEN** Apply 成功且插件返回的状态非空
- **THEN** `ApplyResult.NewState` 的类型是 `map[string]any`

#### Scenario: 删除后的状态
- **WHEN** 资源已删除，插件返回的状态为空
- **THEN** `ApplyResult.NewState` 为 `nil`

### Requirement: 编解码不接收 Private，也不改既有资源路径
系统 MUST NOT 给编解码增加 `Private` 方法，也 MUST NOT 让 `Encode` 或 `Decode` 接收 `Private`。系统 MUST NOT 解码或持久化这些字节。本变更 MUST NOT 修改 `PluginExecutor` 对插件方法的选择，MUST NOT 修改 `NewProcessManager` 的参数，也 MUST NOT 在进程管理中自动注入编解码。本变更的代码差异 MUST NOT 包含既有资源 Handler 目录和 `pkg/adaptor` 下的文件。

#### Scenario: Private 字节不进入编解码
- **WHEN** 调用方持有私有字节 `0x01 0x02`，并只通过执行器传递该字节
- **THEN** 编解码接口不接收这组字节，本变更也不保存它的解码结构

#### Scenario: 统计既有目录的变更
- **WHEN** 统计本变更里既有资源 Handler 与 `pkg/adaptor` 的变更文件
- **THEN** 这些目录的变更文件数为 0
