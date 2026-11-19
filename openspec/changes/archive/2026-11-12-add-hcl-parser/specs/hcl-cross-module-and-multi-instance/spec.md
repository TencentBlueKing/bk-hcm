## ADDED Requirements

### Requirement: Pairwise Module Prefix Consumption and Cross-Module Reference Penetration
阶段一 `Parse` SHALL 支持多层级子模块递归穿透（模块 `source` 查找自动兼容带 `"./"` 与不带 `"./"` 的模块路径 key，且同级多个 `module` 调用块按跨模块 `output` 引用关系先执行拓扑排序再依次展开），`block_id` 格式 MUST 对齐 Terraform 原生标准（根模块为 `<type_name>.<name>` 或 `data.<type_name>.<name>`、`import.<type_name>.<name>`，每进入一层子模块前置追加 `module.<模块名>[可选下标].`）：
- 解析引用路径时，SHALL 基于 `hclsyntax.ParseTraversalAbs` 解析地址段（支持单段连续多维下标如 `az_map["gz"][1]` 与含点号的字符串下标如 `vpc["gz.01"]`），并从左往右按 **2 段一组（无下标 `module` 关键字 + `<模块名>[可选下标]`）** 成对消费模块前缀，直到遇到非 `"module"` 且包含 `_` 的词（或 `data` / `import`）为止；随后若遇到 `data` 或 `import` 则取 3 段，否则固定取 2 段作为模块内节点名，剩余段作为 `state` 属性访问路径；
- 跨模块传参链（如子模块 `output` → 父模块 → 另一子模块 `var` → `local` → 资源属性）SHALL 在阶段一全链路穿透，直接将下游节点的 `DependsOn` 和 `PendingExprs` 指向源头可调度节点 `block_id`。

#### Scenario: End-to-end 3-module 5-block HCL project penetration (AC-001)
- **WHEN** 输入 3 模块 HCL 工程（根模块 `.`、子模块 `./modules/network`、`./modules/app`，含 8 类顶层块）与 `vars: {"subnet_cidr": "127.0.0.1/24"}` 调用 `Parse(req)`
- **THEN** 返回 `err == nil`，`ParseResult.Blocks` 按拓扑顺序仅包含 5 个可调度节点（`module.network.import.tencentcloud_vpc.existing`、`data.bkcmdb_biz.target`、`module.network.tencentcloud_subnet.app`、`module.app.tencentcloud_instance.web`、`module.app.bkcmdb_host.web`），跨模块 `output`/`var` 引用全部穿透指向源头节点 `block_id`

#### Scenario: Disambiguate multi-level modules and resources named module via pairwise consumption (AC-007)
- **WHEN** 构造两层子模块调用，其中内层模块名取名为 `"module"`、资源名也取名为 `"module"`（完整节点 ID 为 `module.network.module.module.tencentcloud_vpc.module`），下游节点通过 `output` 和 `var` 引用其 `.id` 字段并调用 `Parse(req)`
- **THEN** 阶段一准确识别源头 `block_id` 为 `"module.network.module.module.tencentcloud_vpc.module"`、属性路径为 `"id"`，零歧义写入 `DependsOn` 与 `PendingExprs`

### Requirement: Module-Level and Resource/Data-Level count/for_each Expansion and Splat References
阶段一 `Parse` SHALL 支持 `module`、`resource`、`data` 块自身配置 `count`（非负整数 `>= 0`）或 `for_each`（map/object 或字符串集合/列表/元组，按键字典序确定性展开，且同一块上禁止同时指定 `count` 与 `for_each`）：
- `module` 块带 `count` 或 `for_each` 时，展开为带 `[index]` 或 `["key"]` 下标的模块实例前缀（如 `module.vpc[0]`、`module.vpc["gz"]`）；
- `resource` / `data` 块带 `count` 或 `for_each` 时，展开为多个独立 `Block` 节点，`block_id` 末尾附带 `[0]` 或 `["key"]` 下标；支持跨资源的静态实例数链式推导（如 `count = length(<other_resource>)` 在阶段一按需先展开目标多实例资源并静态算出实例总数，同时检测 `count`/`for_each` 循环引用）；
- 下游使用同下标精确引用（如 `sub[count.index].id` 或 `sub[each.key].id`）时，`DependsOn` SHALL 仅包含同下标的单个上游实例节点；
- 下游使用全量 `[*]` Splat 引用（如 `sub[*].id`）时，SHALL 自动将该多实例资源展开后的全部实例 `block_id` 加入当前节点的 `DependsOn`，并在 `PendingExprs` 中将 `[*]` 展开改写为阶段二可直接对各实例 `state` 求值的列表表达式（如 `[sub[0].id, sub[1].id]`）。

#### Scenario: Expand module for_each, resource count, same-index reference, and full Splat reference (AC-009)
- **WHEN** ① 子模块 `module "vpc"` 配置 `for_each = toset(["sh", "gz"])`；② 资源 `tencentcloud_subnet.sub` 配置 `count = 2`；③ 资源 `tencentcloud_instance.app` 配置 `count = 2` 且 `subnet_id = tencentcloud_subnet.sub[count.index].id`；④ 资源 `tencentcloud_clb.lb` 配置 `subnets = tencentcloud_subnet.sub[*].id`，调用 `Parse(req)`
- **THEN** ① 按键字典序展开出 `module.vpc["gz"].tencentcloud_vpc.main` 与 `module.vpc["sh"].tencentcloud_vpc.main`；② 展开出 `tencentcloud_subnet.sub[0]` 与 `tencentcloud_subnet.sub[1]`；③ `tencentcloud_instance.app[0]` 的 `DependsOn` 仅含 `["tencentcloud_subnet.sub[0]"]`，`app[1]` 的 `DependsOn` 仅含 `["tencentcloud_subnet.sub[1]"]`；④ `tencentcloud_clb.lb` 的 `DependsOn` 同时包含 `["tencentcloud_subnet.sub[0]", "tencentcloud_subnet.sub[1]"]`，且 `PendingExprs["subnets"]` 为 `"[tencentcloud_subnet.sub[0].id, tencentcloud_subnet.sub[1].id]"`

#### Scenario: Expand chained resource count referencing length of another multi-instance resource
- **WHEN** 资源 `aws_instance.a` 配置 `count = 3`，同模块资源 `aws_eip.b` 配置 `count = length(aws_instance.a)` 且 `instance = aws_instance.a[count.index].id`，调用 `Parse(req)`
- **THEN** `Parse` 在阶段一静态算出 `aws_eip.b` 的实例数为 `3`，展开出 `aws_instance.a[0..2]` 与 `aws_eip.b[0..2]` 共 6 个节点，且每个 `aws_eip.b[i]` 仅精确依赖对应的 `aws_instance.a[i]`

#### Scenario: Expand count = 0 or empty for_each into zero blocks
- **WHEN** `module`、`resource` 或 `data` 块配置了 `count = 0` 或 `for_each = toset([])` / `{}`
- **THEN** 该块在 `ParseResult.Blocks` 中展开为 0 个节点且不报错

#### Scenario: Reject count or for_each depending on runtime pending attributes or negative count (EX-007, AC-010)
- **WHEN** 块配置 `count = data.bkcmdb_biz.target.host_count`（依赖阶段一未知的运行期输出）或 `count = -1`
- **THEN** `Parse(req)` 返回非 `nil` 错误，提示 `count` / `for_each` 必须在阶段一静态可求值且取值合法

### Requirement: Cross-Module Multi-Account Provider Mapping
阶段一 `Parse` SHALL 支持根模块定义多个同插件 `provider`（默认无 `alias` 与带 `alias`），并支持子模块通过 `module` 块的 `providers = { <plugin>[.<child_alias>] = <plugin>.<parent_alias> }` 显式传递指定账号的 `plugin_instance_id`（及节点块级 `provider = <plugin>.<alias>` 显式选择）；当子模块 `module` 块未显式声明 `providers` 时，SHALL 默认继承父模块所有同名无 `alias` 的默认 `provider`。

#### Scenario: Bind explicit aliased provider and inherit default provider across child modules (AC-011)
- **WHEN** 根模块定义默认 `provider "tencentcloud" { plugin_instance_id = "00000001" }` 与 `provider "tencentcloud" { alias = "prod", plugin_instance_id = "00000009" }`，子模块 `module "network"` 显式声明 `providers = { tencentcloud = tencentcloud.prod }`，另一子模块 `module "dev"` 未声明 `providers`
- **THEN** `module.network` 下所有 `tencentcloud_*` 节点的 `PluginInstanceID` 为 `"00000009"`，而 `module.dev` 下所有 `tencentcloud_*` 节点的 `PluginInstanceID` 继承默认账号 `"00000001"`

#### Scenario: Reject unmapped plugin provider or non-existent provider alias in module mapping (AC-012)
- **WHEN** `module` 块的 `providers` 映射引用了父模块不存在的 provider alias（如 `tencentcloud.non_existent`）
- **THEN** `Parse(req)` 返回非 `nil` 错误，提示引用的 `provider` 配置不存在

### Requirement: Cycle Detection, Recursion Depth Guard, Deterministic Ordering, and Stage 1 Performance
阶段一 `Parse` SHALL 对模块递归展开深度施加安全上限（最大嵌套深度 `constant.MaxModuleDepth = 16` 层），对模块调用栈、同级兄弟模块 `output` 引用、`locals` 引用、`count`/`for_each` 实例数引用及展开后的 `Block.DependsOn` 节点依赖图执行有向环路检测（同时校验全局 `BlockID` 唯一性），对输出的 `ParseResult.Blocks` 按三级确定性拓扑序（① 拓扑层级升序 → ② 同层按 `import < data < resource` 类型优先级 → ③ 同类型按 `BlockID` 字典序升序）排序，对每个节点的 `Block.DependsOn` 去重后按字典序升序排列，并在包含 3 个模块、展开后 50 个 `Block` 节点的标准 HCL 工程下满足单次 `Parse` P99 耗时 `< 50ms`。

#### Scenario: Reject circular module calls, sibling module cycles, count cycles, or circular resource dependencies (EX-008, AC-008)
- **WHEN** 模块 A 调用模块 B 且模块 B 调用模块 A、同级模块互引 `output`、`locals` 互引、资源间 `count = length(...)` 互引，或节点 A 的属性/`depends_on` 引用节点 B 且节点 B 引用节点 A
- **THEN** `Parse(req)` 返回非 `nil` 错误，明确指出存在循环依赖

#### Scenario: Reject module nesting exceeding 16 levels and verify 100-run deterministic output (EX-008, AC-S02)
- **WHEN** 构造超过 16 层嵌套的子模块调用 `Parse(req)`，或对同一 3 模块工程连续调用 `Parse(req)` 100 次
- **THEN** 超深嵌套调用 100% 返回超过最大递归深度错误；100 次连续调用产出的 `ParseResult` JSON 序列化结果 100% 完全一致

#### Scenario: Meet Stage 1 Parse performance benchmark for 50 blocks (AC-P01)
- **WHEN** 运行基准测试 `go test -bench=BenchmarkParse_50Blocks -benchmem`（3 个模块、50 个展开后 `Block` 节点）
- **THEN** 单次 `Parse` 执行耗时 < 50ms
