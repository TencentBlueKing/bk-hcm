## Why

执行器已经把配置和状态交给 `executor.ValueCodec`，但这个接口还没有按 Schema 做 msgpack 编解码。插件协议里的 config、state、result 都是 `DynamicValue`，字节本身分不出 List 与 Set、Object 与 Map。没有这一层，执行器无法把调用方对象送进官方插件，也无法把插件响应解回对象。

本次变更对应 TAPD 需求「DynamicValue 编解码」（story 139221653），依据 iWiki《技术方案》第 6.4 节。执行器变更已经留下调用边界，本变更补上边界后面的实现。

## What Changes

- 新增满足现有 `Encode` / `Decode` 的编解码实现。入参是调用方对象和一份 `*tfplugin5.Schema`，产出是 `*tfplugin5.DynamicValue` 的 msgpack；解码使用编码时的同一份 Schema，不从字节推断类型。
- 调用方对象用 `nil` 表示 null，对象和 map 用 `map[string]any`，list 和 set 用 `[]any`。资源状态、Provider 配置和数据源结果的根就是这个 map。`plugin.go` 里对应的状态字段和 `Executor` 的对应入参改为 `map[string]any`。`Attribute.Type` 保持 `any`。具体集合类型由 Schema 决定。编码前校验字段是否存在、类型是否匹配；Schema 未声明的字段直接失败，错误信息带上字段名。类型不匹配时在构造带类型的值之前返回错误。
- 可选且未设置的字段编码为 null。Computed 且当前未知的字段编码为未知值，用编解码包内的固定哨兵表示，不编成 null 或空字符串。
- 不给编解码增加 `Private` 方法，也不把 `Private` 放进 `Encode` / `Decode`。执行器已经按字节原样传递 `Private`；本变更不解码、不持久化这些字节。空的配置和状态仍由执行器挡在编解码之外。
- 不改 `PluginExecutor` 对插件方法的选择，也不在 `NewProcessManager` 里自动注入编解码。进程池以后自己调用已有的 `SetValueCodec`。
- 用固定 Schema 和样例值验证编码、未知字段、未知值、往返和解码缺 Schema。不改已有资源 Handler，不改 `pkg/adaptor`。

## Capabilities

### New Capabilities

- `dynamic-value-codec`: 按一份 Schema 把调用方对象编码为 `DynamicValue`，并用同一份 Schema 解回对象，覆盖 null、未知值和未声明字段。

### Modified Capabilities

无。`openspec/specs/` 下没有 DynamicValue 编解码或执行器的正式能力。执行器变更里「不实现 msgpack」只约束那次变更，不改写。

## Impact

- 实现放在 `cmd/hc-service/logics/tfruntime`，对外仍是现有 `executor.ValueCodec`。执行器继续使用已提交的 `tfplugin5` 结构；`tftypes` 只留在编解码实现内部。
- 重新引入 `github.com/hashicorp/terraform-plugin-go`，依赖范围限于编解码。执行器、进程管理和资源 Handler 不直接使用该库。
- 不新增北向 HTTP API、表结构或进程环境变量。sensitive 字段不写入编解码自己的日志明文。
- 不影响现有云资源同步、账单和账号等 adaptor 调用。
