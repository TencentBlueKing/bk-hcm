## Why

在海垒（HCM）插件化改造与 IaC 编译调度架构（M2 里程碑）中，终端用户提交的是包含根模块 `"."` 与多个子模块 `module` 的多模块 HCL 工程（涉及 `variable`、`locals`、`provider`、`module`、`output`、`resource`、`data`、`import` 八类顶层块，以及跨模块传参、`count`/`for_each` 多实例展开、内置函数计算与 `lifecycle` 元参数），而底层 `hc-service` 北向接口仅接受单个节点的纯 JSON 配置（`config`）与北向定位三元组（`plugin`、`plugin_instance_id`、`type_name`）。目前仓库已落地 `hc-service` 北向接口与 `iac-server` 服务骨架，亟需在公共包 `pkg/iacparser` 下交付两阶段 HCL 语法解析与增量求值器（阶段一 `Parse` 一次性静态展开穿透生成可调度 `Block` 列表供构建 Plan/Apply DAG；阶段二 `EvalBlock` 在上游依赖节点执行完成后增量求值 `pending_exprs` 并按顶层键或点号路径回填产出纯 JSON `config`），打通从 HCL 工程到 DAG 编排与北向插件执行的编译闭环。

## What Changes

### 本期包含（In Scope）

1. **`pkg/iacparser` 对外接口契约与核心数据模型（纯 Go Library 交付）**：
   - 在公共枚举与常量包 `pkg/criteria/enumor`（`iac.go`）中定义 `BlockKind` 枚举类型、常量（`BlockKindResource = "resource"`、`BlockKindData = "data"`、`BlockKindImport = "import"`）及 `Validate()` 方法，在 `pkg/criteria/constant`（`iac.go`）中定义 `MaxModuleDepth = 16` 与 `RootModuleKey = "."`；在公共包 `pkg/iacparser`（`package iacparser`）下交付两阶段 HCL 语法解析与增量求值器，定义 `Parser` 接口（含 `Parse(req *ParseReq) (*ParseResult, error)` 与 `EvalBlock(req *EvalReq) (map[string]any, error)`）、默认无状态构造函数 `NewParser() Parser`，以及 `ParseReq`（含 `Validate()`）、`ParseResult`、`Block`（`Kind` 为 `enumor.BlockKind`）、`Lifecycle`（含 `create_before_destroy`、`prevent_destroy`、`ignore_changes`、`replace_triggered_by` 与海垒扩展字段 `recompute`）、`EvalReq`（含 `Validate()`）全部数据模型。
2. **HCL 依赖引入与安全纯内存内置函数上下文**：
   - 在 `go.mod` 中引入官方 `github.com/hashicorp/hcl/v2 v2.24.0`，结合仓库已有的 `github.com/zclconf/go-cty v1.16.4` 构建阶段一与阶段二共用的 HCL 解析与标准内置函数求值上下文：启用 `cty/function/stdlib` 纯内存标准函数，并补齐常用 Terraform 类型转换、集合与编解码/哈希函数（`tostring`、`tonumber`、`tobool`、`tolist`、`toset`、`tomap`、`lookup`、`length`、`index`、`try`、`can`、`base64encode`、`base64decode`、`md5`、`sha1`、`sha256`、`urlencode` 等），同时维护显式黑名单 `forbiddenFunctions` **严禁**调用任何访问宿主机文件系统、路径或外部模板渲染的函数（如 `file`、`fileexists`、`fileset`、`filebase64`、`templatefile`、`pathexpand`、`abspath`）。
3. **阶段一 `Parse` 基础解析与单块属性四分类拆解**：
   - 解析各模块目录下的 `.tf` 文件字节流（同模块多 `.tf` 文件按文件名排序合并，拦截顶层游离属性与不支持的顶层块类型），识别 8 种顶层块，内部穿透消化 `variable`、`locals`、`provider`、`module`、`output` 5 种桥梁块，生成 `resource`、`data`、`import` 3 种可调度 `Block` 节点并提取北向定位三元组（`plugin`、`plugin_instance_id`、`type_name`；支持块级 `provider` 元参数覆盖；其中 `import` 节点的 `config` 固定为 `{"import_id": "<计算后的 id>"}`，同模块同时定义 `import` 与同名 `resource` 时自动建立依赖关联，仅定义 `import` 时自动映射引用）。
   - 实现单块属性四分类拆解：① 普通静态属性直接求值写入 `Block.Config`；② 顶层未决属性经桥梁穿透改写为源头 `block_id` 表达式后写入 `Block.PendingExprs` 并收集 `Block.DependsOn`（支持静态变量与未决属性混合表达式的静态部分内联替换，以及静态三元条件短路裁剪）；③ 嵌套子块（按块名合并为 `[]any` 切片）与内联对象 `{...}` / 内联对象列表 `[{...}]` 内部的未决叶子属性在 `Block.Config` 中保留结构骨架并以 `nil`（JSON `null`）占位，以点号路径（如 `"tags.0.value"`、`"inline_obj.dyn_k"`、`"inline_arr.0.v"`）为 key 写入 `Block.PendingExprs`；④ 剥离 `lifecycle` 元参数子块进 `Block.Lifecycle`（其中 `replace_triggered_by` 引用的源头资源自动并入 `Block.DependsOn`）并处理显式 `depends_on` 元参数，确保元参数绝不进入业务 `Block.Config`。
4. **阶段一跨模块与多实例 4 大高阶维度展开**：
   - **维度 1（多层级 `module` 穿透与 `module.<模块名>` 成对消费无歧义）**：兼容 `"./"` 模块路径匹配与同级 `module` 调用拓扑排序，按两段一组（`module` + `<模块名>[可选下标]`）从左往右成对消费模块前缀，零歧义切分源头 `block_id` 与属性路径（即使模块名、资源名、属性名取名为 `"module"`），全链路穿透跨模块 `var`/`local`/`output` 传参链，支持模块调用环路、`locals` 环路、节点依赖环路检测及最大嵌套深度 `constant.MaxModuleDepth = 16` 保护，并对最终 `ParseResult.Blocks` 按“拓扑层级升序 → 同层 `import < data < resource` → 同类型 `BlockID` 字典序”执行确定性拓扑排序。
   - **维度 2（`module` 块自身 `count` / `for_each` 展开）**：展开为带 `[index]` 或 `["key"]` 下标的模块实例前缀（如 `module.vpc["gz"].tencentcloud_vpc.main`）。
   - **维度 3（`resource` / `data` 块自身 `count` / `for_each` 与 `[*]` Splat 引用）**：展开为带下标的独立 `Block` 节点（支持 `count = length(<other_resource>)` 跨资源静态实例数链式推导及 `count`/`for_each` 循环检测）；支持同下标精确引用（`sub[count.index].id` / `sub[each.key].id` 仅依赖同下标单实例）与 `[*]` Splat 全量扇入引用（将全部实例 `block_id` 加入 `DependsOn` 并改写 `PendingExprs` 为列表表达式）。
   - **维度 4（跨模块 `providers` 多账号映射）**：支持根模块多 `alias` `provider` 定义及子模块 `providers = { <plugin> = <plugin>.<alias> }` 显式传递 `plugin_instance_id`，未显式声明时默认继承父模块同名无 alias 默认 `provider`。
5. **阶段二 `EvalBlock` 单节点增量求值与点号路径精准回填**：
   - 校验非空 `EvalReq` 并深拷贝 `EvalReq.Config` 防止原地修改污染；校验 `EvalReq.DependencyStates` 键合法性并按 `module.<模块名>` 成对消费规则与多实例下标（区分对象字典与元组切片）还原为 `hcl.EvalContext` 嵌套 `cty.Value` 变量树；二次校验表达式函数安全性后挂载安全纯内存内置函数库求值 `PendingExprs`（保持整数语义精度），并按顶层 key 或点号路径（支持 map 键值与切片元素下标回填，如 `"tags.0.value"`、`"arr_leaf.0"`）精准回填，产出完整纯 JSON `config`。
6. **完备的单元测试、并发安全与性能基准测试套件**：
   - 交付覆盖全部正反例（AC-001 ~ AC-014）、安全红线（AC-S01 ~ AC-S02）、iWiki §3 端到端全链路示例、100 次确定性输出验证、50 协程并发安全（`go test -race`）以及阶段一 3 模块 50 节点 `BenchmarkParse_50Blocks`（P99 < 50ms）与阶段二 `BenchmarkEvalBlock`（P99 < 5ms）的测试套件。

### 本期不包含（Out of Scope）

1. **不包含独立 HTTP 接口暴露**：本期仅作为 `pkg/iacparser` 纯 Go Library 交付，不单独在 `iac-server` 暴露 `/parse` 或 `/eval_block` HTTP 接口（由后续 Plan/Apply 编排逻辑直接通过 Go 包调用）。
2. **不包含 Plan / Apply DAG 图构建与拓扑编排**：如 CBD（`create_before_destroy`）标记向上游传递、替换动作拆分为创建与销毁节点、销毁边反向等，属于 `iac-server` 的 DAG 编排需求范围，不在本语法解析需求中实现。
3. **不包含 HCL Spec 版本托管、草稿/发布管理与 State 数据库读写**：属于 `iac-server` Spec / State 管理需求与 `data-service` 范围。
4. **不包含 `task-server` 任务调度与对 `hc-service` 北向接口的 HTTP 实际调用**：本需求仅产出供编排与调度消费的 `ParseResult` 与 `EvalBlock` 纯 JSON `config`，不直接发起网络调用。
5. **不包含两个 HCL 版本的文本对比与参数模版渲染**：属于 M3 里程碑范围。

## Capabilities

### New Capabilities

- `hcl-parser-contract-and-single-module`: `pkg/iacparser` 对外接口契约与数据模型定义（`Parser`、`NewParser()`、`ParseReq`、`ParseResult`、`Block`、`Lifecycle`、`EvalReq`）及 `pkg/criteria/enumor.BlockKind`、`pkg/criteria/constant.{MaxModuleDepth,RootModuleKey}` 公共枚举常量定义、安全纯内存内置函数注册表与黑名单拦截、阶段一 8 类顶层块识别与桥梁块（`variable`/`locals`/`provider`/`module`/`output`）内部消化、`resource`/`data`/`import` 北向三元组定位与同名 `import`/`resource` 关联，以及单块属性四分类拆解（静态 `config`、顶层 `pending_exprs`、嵌套子块 `[]any` 与内联对象/对象列表骨架 `nil` 占位及点号路径提取、`lifecycle` 5 字段剥离与显式 `depends_on` 收集）。
- `hcl-cross-module-and-multi-instance`: 阶段一跨模块与多实例四维展开能力，含多层级 `module` 递归穿透与 `module.<模块名>` 成对消费无歧义路径切分、`module` 级与 `resource`/`data` 级 `count`/`for_each` 多实例展开（含 `count = length(<resource>)` 链式实例数推导）、同下标精确引用与 `[*]` Splat 扇入引用改写、跨模块 `providers` 多账号映射传递、全维度环路检测与 16 层递归深度保护、三级确定性拓扑排序及 `BenchmarkParse_50Blocks` 性能基准。
- `hcl-stage2-eval-and-backfill`: 阶段二 `EvalBlock` 单节点增量求值与点号路径回填能力，含 `EvalReq` 校验与 `Config` 深拷贝隔离、基于成对消费规则与多实例下标将 `DependencyStates` 构建为 `hcl.EvalContext` 嵌套 `cty.Value` 变量树、`PendingExprs` AST 函数安全校验与内置函数求值（保持整数精度）、顶层键及点号路径（支持 map 键值与切片下标，如 `tags.0.value`、`arr_leaf.0`）精准回填、50 协程并发安全验证及 `BenchmarkEvalBlock` 性能基准。

### Modified Capabilities

无（本期为全新公共编译解析包 `pkg/iacparser` 及配套 IaC 枚举常量定义，不改变任何既有 capability 的规范要求）。

## Impact

- **新增公共常量与枚举**：`pkg/criteria/constant/iac.go`（`MaxModuleDepth`、`RootModuleKey`）、`pkg/criteria/enumor/iac.go`（`BlockKind` 及 `Validate()`）。
- **新增公共代码包**：`pkg/iacparser/`（`types.go`、`parser.go`、`cty_conv.go`、`functions.go`、`address.go`、`module_loader.go`、`expr_eval.go`、`block_decomposer.go`、`expander.go`、`dag_validator.go`、`eval_block.go`，以及完整单元测试与 Benchmark 测试文件 `parser_single_module_test.go`、`parser_cross_module_test.go`、`eval_block_test.go`、`benchmark_and_race_test.go`）。
- **依赖管理（`go.mod` / `go.sum`）**：将 `github.com/hashicorp/hcl/v2 v2.24.0` 声明为直接依赖，并复用仓库已有的 `github.com/zclconf/go-cty v1.16.4`。
- **上下游服务与数据库影响**：纯内存无状态 Go Library，不涉及任何配置文件（`etc/*.yaml` / Helm）、HTTP 路由、外部网络调用或 MySQL 数据库 DDL 变更，对存量业务服务零侵入、零破坏性影响。
