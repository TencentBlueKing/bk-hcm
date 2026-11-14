## 1. 依赖与包边界

- [x] 1.1 在 `go.mod` 增加 `github.com/hashicorp/terraform-plugin-go v0.28.0`。确认直接依赖的 `go-plugin` 仍是 v1.6.3，gRPC 仍是仓库现有版本，不降级。
- [x] 1.2 新建 `cmd/hc-service/logics/tfruntime/codec`，实现 `executor.ValueCodec`。执行器包不引用该包。`tftypes` 只出现在 `codec` 包内。

## 2. Schema 转 tftypes

- [x] 2.1 把 `tfplugin5.Schema` 的顶层块转成 `tftypes.Object`。属性 `Type` 使用 `tftypes.ParseJSONType`。嵌套块按 `NestingMode` 并入同一对象：`SINGLE` 与 `GROUP` 为 object，`LIST` 为 list，`SET` 为 set，`MAP` 为 map。
- [x] 2.2 `min_items`、`max_items` 和 Schema 的 `version` 不参与类型。属性名与嵌套块名冲突、`NestingMode` 为 `INVALID`、类型为 `dynamic` 或类型字节无法解析时，返回带属性路径的错误，不产出 `DynamicValue`。

## 3. 编码

- [x] 3.1 导出固定哨兵 `Unknown`。对象和 map 只接受 `map[string]any`，list 和 set 只接受 `[]any`。
- [x] 3.2 编码前走完类型树。未声明字段失败。键缺失或值为 `nil` 时编码为 null，`required` 缺失也是 null。值为哨兵时编码为未知值，不编成 null 或空字符串。不因 `computed` 把缺省字段改成未知值。number 接受 `tftypes` 已接受的整数、`float64` 和 `*big.Float`。
- [x] 3.3 类型不符时在 `tftypes.NewValue` 之前返回错误。错误只含属性路径，list 和 set 含下标，不含字段值。通过后自底向上做成 `tftypes.Value`，每层先 `ValidateValue` 再 `NewValue`。
- [x] 3.4 `Encode` 成功时用 `Value.MarshalMsgPack` 只写 `DynamicValue.Msgpack`。编解码包自身不打日志。

## 4. 解码

- [x] 4.1 `Decode` 用 `tftypes.ValueFromMsgPack` 和这次传入的 Schema 类型。Schema 或顶层块为空，或只有 `Json` 没有 `Msgpack` 时返回错误，不产出调用方对象。
- [x] 4.2 null 解为 `nil`，未知值解为同一个哨兵。string 和 bool 保持原类型。number 解为 `*big.Float`。list、set 和 tuple 解为 `[]any`。object 和 map 解为 `map[string]any`。
- [x] 4.3 用 set 编出的字节改用元素类型相同的 list Schema 解码时，按 list 返回 `[]any`，不从字节恢复 set。

## 5. 对外类型与现有边界

- [x] 5.1 把 `UpgradeResult.UpgradedState`、`ReadResourceResult.NewState`、`PlanResult.PlannedState`、`ApplyResult.NewState`、`ImportedResource.State`、`ReadDataSourceResult.State` 改为 `map[string]any`。`Executor` 上对应的 config、current、prior、proposed、planned 入参一并改为 `map[string]any`。`nil` 表示空。`Attribute.Type` 保持 `any`。`ValueCodec` 的签名保持 `any`。
- [x] 5.2 不增加 `Private` 方法，`Encode` 和 `Decode` 不接收 `Private`。不修改 `PluginExecutor` 的方法选择，不修改 `NewProcessManager` 的参数，不在进程管理里自动注入编解码。
- [x] 5.3 不修改既有资源 Handler 和 `pkg/adaptor`。

## 6. 验证

- [x] 6.1 用固定 Schema 覆盖：可选字段缺省与显式 `nil` 都是 null；`computed` 缺省是 null；哨兵编码为未知值。同时覆盖 `LIST` 与 `SET` 嵌套块都接受 `[]any`，但编码类型分别是 list 和 set。
- [x] 6.2 覆盖未声明字段 `extra`、list 元素类型不匹配、`dynamic` 类型和 `INVALID` 嵌套模式。确认不返回 `DynamicValue`，错误含属性路径且不含字段值。
- [x] 6.3 覆盖同一 Schema 往返，number 以 `*big.Float` 的数值相等为准。list 保持顺序。set 用元素集合比较，不比较下标。覆盖 set 字节用 list Schema 解码、空 Schema，以及只有 `Json` 时不产出对象。
- [x] 6.4 确认 `codec` 的导出方法只有 `Encode` 和 `Decode`，本变更不改资源 Handler 与 `pkg/adaptor`，`go-plugin` 仍是 v1.6.3。
