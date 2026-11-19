## Context

`executor.ValueCodec` 已经定下边界：按一份 `*tfplugin5.Schema` 把调用方对象编成 `*tfplugin5.DynamicValue`，再用同一份 Schema 解回来。`PluginExecutor.encode` 在值为 `nil` 时不调用编解码；`decode` 在 msgpack 和 json 都为空时直接返回 `nil`。`Private` 在 `resource.go` 里按字节克隆，不进入编解码。

协议 Schema 的属性 `Type` 是 TF JSON 类型字节。块上的 `BlockTypes` 另用 `NestingMode` 表示 single、list、set、map、group。msgpack 不保留这些差别，解码必须用编码时的 Schema。

仓库当前是 Go 1.24，直接依赖 `github.com/hashicorp/go-plugin v1.6.3`。还没有 `terraform-plugin-go`。执行器测试里的 `jsonCodec` 只证明调用顺序，不是协议编码。

## Goals / Non-Goals

**Goals:**

- 提供可注入 `SetValueCodec` 的实现，满足现有 `Encode` / `Decode`。
- 编码前按 Schema 校验，类型不匹配或未声明字段时返回带属性路径的错误，不进入会 panic 的构造。
- 缺省字段编码为 null；调用方标出的未知字段编码为未知值。
- 解码只认 msgpack，类型取自传入的 Schema。
- 用固定 Schema 和样例值覆盖往返、未声明字段、未知值，以及 List 与 Set 必须靠 Schema 区分。

**Non-Goals:**

- 不增加 `Private` 方法，不解码、不持久化 `Private`。
- 不改 `PluginExecutor` 的方法选择、空值短路和 `NewProcessManager` 的参数。不在进程管理里自动注入。
- 不实现 `UpgradeResourceState` 的 `RawState`。那条路径不是 `DynamicValue`。
- 不解释 `required`、`optional`、`computed`、`sensitive`、`write_only` 以外的协议标志来改编码规则。`sensitive` 只约束错误和日志不得带出字段值。
- 不支持 TF JSON 里的 `dynamic`。遇到该类型时按属性路径失败。
- 不改资源 Handler 和 `pkg/adaptor`。不新增时延指标。

## Decisions

### 1. 编解码独立成包，只实现现有接口

新包放在 `cmd/hc-service/logics/tfruntime/codec`。它实现 `executor.ValueCodec`，执行器包不引用它，因此 `tftypes` 不会进执行器。类型无状态，可以并发使用。进程池以后自己调用已有的 `SetValueCodec`。

备选是把实现写进 `executor`。那样执行器文件会直接依赖 `terraform-plugin-go`，和「库只留在编解码里」不一致。

### 2. 依赖停在 terraform-plugin-go v0.28.0

使用 `github.com/hashicorp/terraform-plugin-go v0.28.0` 的 `tftypes`。这一版要求 Go 1.23，并依赖 `go-plugin v1.6.3`，与仓库里的进程管理依赖相同。属性类型用 `tftypes.ParseJSONType` 解析。构造前用 `tftypes.ValidateValue` 检查，通过后再调用 `tftypes.NewValue`。msgpack 使用 `Value.MarshalMsgPack` 和 `tftypes.ValueFromMsgPack`。结果写入 `tfplugin5.DynamicValue.Msgpack`。

不使用 v0.29.0 及更新版本。那些版本把 `go-plugin` 升到 v1.7.0，会带动进程管理的依赖。仓库里的 gRPC 已经高于 v0.28.0 所声明的版本，按 Go 的最小版本选择保持现有 gRPC，不降级。

备选是手写 msgpack。List 与 Set、Object 与 Map 的字节差别要自己维持，也要自己避开 `NewValue` 的 panic，收益低于复用 `tftypes`。

### 3. Schema 先变成 tftypes.Object，再校验调用方对象

一份 `tfplugin5.Schema` 的顶层块转换成 `tftypes.Object`。属性的 `Type` 字节交给 `ParseJSONType`。嵌套块按 `NestingMode` 并进同一个对象：

| NestingMode | 对象里的属性类型 | 调用方形状 |
|---|---|---|
| SINGLE、GROUP | 内层块对应的 object | `map[string]any` 或 null |
| LIST | list(object) | `[]any` 或 null |
| SET | set(object) | `[]any` 或 null |
| MAP | map(object) | `map[string]any` 或 null |
| INVALID | 不编码 | 返回错误 |

属性名和嵌套块名冲突时，构造类型失败。`min_items`、`max_items` 和 Schema 的 `version` 不参与编解码。

调用方对象在进入 `tftypes` 之前先走一遍这棵类型：

- 对象和 map 必须是 `map[string]any`。多出来的键失败，错误带上从根开始的路径，例如 `tags.extra`。
- list 和 set 必须是 `[]any`。下标写进路径，例如 `network[0].name`。
- 键不存在，或值为 `nil`，编码为该类型的 null。`required` 的键缺失同样是 null，不在编解码里另报「缺少必填」。必填由插件的校验 RPC 判定。
- 值为编解码包导出的 `Unknown` 时，编码为该类型的未知值。不因为属性带有 `computed` 就把缺省字段改成未知值。
- 标量只接受该类型对应的 Go 值。number 接受 `tftypes` 已接受的整数、`float64` 和 `*big.Float`。类型不符时返回路径，不把原值交给 `NewValue`。

通过后，自底向上做成 `tftypes.Value`：对象和 map 使用 `map[string]tftypes.Value`，list、set 和 tuple 使用 `[]tftypes.Value`。每一层先 `ValidateValue`，再 `NewValue`。

`dynamic` 以及 `ParseJSONType` 失败的类型，在构造类型时失败，错误带上属性路径。

### 4. 解码按 Schema 走回调用方对象，数字统一为 big.Float

`Decode` 只读取 `Msgpack`。只有 `Json`、没有 `Msgpack` 时返回错误，不把 JSON 当成成功路径。Schema 或顶层块为空时返回错误，不产出对象。

`ValueFromMsgPack` 使用编码时的同一份 `tftypes.Type`。随后：

| tftypes 值 | 调用方对象 |
|---|---|
| null | `nil` |
| unknown | `codec.Unknown` |
| string、bool | `string`、`bool` |
| number | `*big.Float` |
| list、set、tuple | `[]any` |
| object、map | `map[string]any` |

因此用 float64 编码再解码，数值相等，但类型变成 `*big.Float`。往返相等以「已知字段的数值和结构一致」为准，不用 `reflect.DeepEqual` 比较原来的 float64。set 的顺序不作为相等条件。

用 List 的 Schema 解码一段由 Set 编出的字节，结果按 List 返回。这是协议限制，不是编解码可以在字节里恢复的信息。

资源、Provider 配置和数据源的 Schema 根都是块，所以一次编解码的根只有两种：空状态是 `nil`，有状态是 `map[string]any`。未知值只出现在这个 map 的属性里，不作为整份状态的根。`cmd/hc-service/types/plugin.go` 里下面这些字段，以及 `Executor` 上对应的 config、current、prior、proposed、planned 入参，改为 `map[string]any`：

| 字段 | 类型 |
|---|---|
| `UpgradeResult.UpgradedState` | `map[string]any` |
| `ReadResourceResult.NewState` | `map[string]any` |
| `PlanResult.PlannedState` | `map[string]any` |
| `ApplyResult.NewState` | `map[string]any` |
| `ImportedResource.State` | `map[string]any` |
| `ReadDataSourceResult.State` | `map[string]any` |

`Attribute.Type` 保持 `any`。它是 TF JSON 类型，简单类型解码后是 `string`（`"string"`、`"number"`、`"bool"`），复合类型解码后是数组（`["list", T]`、`["object", {...}]`）。它不是状态对象，没有一种 map 或 slice 能同时表示这两种写法。`ValueCodec` 的签名仍是 `any`；执行器把根为对象的结果放进上述 `map[string]any` 字段。

### 5. 错误只带路径，不带字段值

校验和 msgpack 错误都带属性路径。错误文本不拼接字段值，编解码包自身不打日志。这样 `sensitive` 字段不会进编解码自己的日志或错误明文。

## Risks / Trade-offs

- [Risk] v0.28.0 在高于它所声明的 gRPC 版本上编译失败。→ Mitigation：只让 `codec` 包引用 `tftypes`，用该包测试确认能编译。不降级仓库里的 gRPC，也不升级 `go-plugin`。
- [Risk] 缺省字段一律为 null，调用方若忘了传入 `Unknown`，计划里的未知值会变成 null。→ Mitigation：`computed` 不自动改写。未知值只有哨兵一条路径，测试覆盖哨兵不会被编成 null 或空字符串。
- [Risk] 数字解码为 `*big.Float`，和 JSON 解码出的 float64 不能直接深度相等。→ Mitigation：往返测试比较数值。调用方保存状态时按 `*big.Float` 理解 number。
- [Risk] set 往返后顺序变化，被误当成状态漂移。→ Mitigation：相等规则忽略 set 顺序。List 保留顺序。
- [Risk] 手写 Schema 漏掉嵌套块，对象类型与插件不一致。→ Mitigation：类型来自执行器缓存的协议 Schema，不来自适配后的文档。测试同时覆盖属性和 nested block。

## Migration Plan

1. 增加 `terraform-plugin-go v0.28.0`。确认 `go-plugin` 仍是 v1.6.3，gRPC 仍是仓库现有版本。
2. 实现 `codec` 包并通过固定 Schema 测试。此时没有调用方注入它，现有进程行为不变。
3. 进程池接入时调用 `SetValueCodec`。未注入时，执行器维持现在的「需要编解码则失败且不发 RPC」。
4. 回滚时删除 `codec` 包和这条依赖。没有表、配置和北向 API。

## Open Questions

无。依赖版本、未知值哨兵、缺省即 null，以及数字解码为 `*big.Float` 已在上面定下。
