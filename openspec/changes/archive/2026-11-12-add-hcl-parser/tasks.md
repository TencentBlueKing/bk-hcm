## 1. 契约定义、依赖引入与安全纯内存函数库（`hcl-parser-contract-and-single-module`）

- [x] 1.1 在 `go.mod` 中引入 `github.com/hashicorp/hcl/v2 v2.24.0` 直接依赖（复用已有的 `github.com/zclconf/go-cty v1.16.4`）并执行 `go mod tidy`
- [x] 1.2 新增 `pkg/criteria/constant/iac.go`（定义 `MaxModuleDepth = 16`、`RootModuleKey = "."`）、`pkg/criteria/enumor/iac.go`（定义 `BlockKind` 枚举常量 `resource`/`data`/`import` 及 `Validate()`）、`pkg/iacparser/types.go`（定义 `ParseReq`、`ParseResult`、`Block`、`Lifecycle`、`EvalReq` 结构体及参数 `Validate()`）与 `pkg/iacparser/parser.go`（定义 `Parser` 接口、`NewParser()` 构造函数及 `Parse`/`EvalBlock` 顶层编排入口）（覆盖 F-001、EX-001）
- [x] 1.3 新增 `pkg/iacparser/cty_conv.go`，实现 Go `any`（含 `*bool`/`*string`/`*int64` 指针、`json.Number`、`float64`、`int64`、切片、map）与 `cty.Value` 的双向无损转换、`ctyValueToHCLString` 字面量序列化及 `deepCopyMap` 递归深拷贝，确保整数值在双向转换时保持精确整数语义（覆盖 F-001、F-005）
- [x] 1.4 新增 `pkg/iacparser/functions.go`，预构建阶段一与阶段二共用的只读安全纯内存内置函数注册表 `safeFunctions`（集成 `go-cty/cty/function/stdlib` 纯内存函数并封装 `tostring`、`tonumber`、`tobool`、`tolist`、`toset`、`tomap`、`lookup`、`length`、`index`、`try`、`can`、`base64encode`、`base64decode`、`md5`、`sha1`、`sha256`、`urlencode` 等常用函数）及禁用函数黑名单 `forbiddenFunctions`（显式拦截 `file`、`fileexists`、`fileset`、`filebase64`、`templatefile`、`pathexpand`、`abspath`，覆盖 F-001、AC-S01）
- [x] 1.5 新增 `pkg/iacparser/address.go`，基于 `hclsyntax.ParseTraversalAbs` 实现支持连续多维下标（`segmentIndex`）与 `module.<模块名>[下标]` 两段一组成对消费的路径解析器（`consumePairwiseAddress`）及 `block_id` 格式化/切分工具，支持 `resource`/`data`/`import` 节点标识与属性路径零歧义提取（覆盖 R-004、AC-007）

## 2. 阶段一单模块解析与单块属性四分类拆解（`hcl-parser-contract-and-single-module`）

- [x] 2.1 新增 `pkg/iacparser/module_loader.go`，使用 `hclsyntax.ParseConfig` 按文件名排序合并解析模块内多个 `.tf` 文件字节流，拦截顶层游离属性与不支持的顶层块类型，识别并归档 8 类顶层块（`variable`、`locals`、`provider`、`module`、`output`、`resource`、`data`、`import`），对词法/语法错误返回含文件路径与行列号的诊断错误（覆盖 F-002、R-001、AC-002、EX-002）
- [x] 2.2 在 `pkg/iacparser/expander.go` 与 `pkg/iacparser/block_decomposer.go` 中实现单模块桥梁块消化与北向三元组定位：解析 `variable`（`default` 与外部 `Vars` 覆盖，缺失必填变量报错 EX-003）、`locals` 按需惰性求值与环检测、`provider` 的 `(plugin, alias) -> plugin_instance_id` 映射，以及 `resource`/`data`/`import` 的 `plugin`、`plugin_instance_id`、`type_name` 提取（支持块级 `provider` 元参数）与纯 `import` 自动映射 / 同名 `import`+`resource` 自动依赖关联（覆盖 F-002、R-001、AC-001、AC-012、EX-004）
- [x] 2.3 新增 `pkg/iacparser/expr_eval.go`，基于 `hclsyntax.Expression` AST 实现阶段一统一表达式求值与改写引擎（`evalExpr`、`tryParseRefChain`、`resolveRefChain`），支持桥梁引用（`var`/`local`/`output`/纯 `import`）穿透改写、混合表达式中已知静态子表达式内联替换、静态三元条件短路裁剪、`length(<multi_instance_res>)` 试算收敛及源头 `block_id` 依赖收集，拦截未定义引用与禁用/未知函数（覆盖 F-003、R-002、AC-003、AC-004、EX-005、EX-006）
- [x] 2.4 新增 `pkg/iacparser/block_decomposer.go`，实现单块属性四分类拆解：① 普通静态属性求值进 `Block.Config`；② 顶层未决属性改写进 `Block.PendingExprs`；③ 嵌套子块（组装为 `[]any`）与内联对象 `{...}` / 对象列表 `[{...}]` 内部未决叶子在 `Block.Config` 填 `nil` 占位并以点号路径（如 `"tags.0.value"`、`"inline_obj.dyn_k"`、`"inline_arr.0.v"`）写入 `Block.PendingExprs`；④ 剥离 `lifecycle` 5 字段（`create_before_destroy`、`prevent_destroy`、`ignore_changes`、`replace_triggered_by` 自动并入 `DependsOn`、`recompute`）及显式 `depends_on`，确保元参数不进入 `Block.Config`（覆盖 F-003、R-002、R-003、AC-003、AC-005、AC-006）

## 3. 阶段一跨模块穿透与多实例四维展开（`hcl-cross-module-and-multi-instance`）

- [x] 3.1 新增 `pkg/iacparser/expander.go`，实现从根模块 `constant.RootModuleKey`（`"."`）出发带 AST 缓存的多层级 `module` 递归展开、`"./"` 模块路径兼容查找、同级 `module` 调用拓扑排序、跨模块 `var`/`local`/`output` 全链路依赖穿透以及跨模块 `providers` 多账号映射传递与默认 `provider` 继承，限制最大递归嵌套深度为 `constant.MaxModuleDepth`（`16` 层）（覆盖 F-004 维度 1 & 维度 4、R-004、R-006、AC-001、AC-007、AC-011、AC-012、AC-S02）
- [x] 3.2 在 `pkg/iacparser/expander.go` 与 `pkg/iacparser/expr_eval.go` 中实现 `module` 级与 `resource`/`data` 级 `count`（非负整数，含 `count = length(<resource>)` 按需展开与循环检测）和 `for_each`（按键字典序展开）多实例展开，支持同下标精确引用（`sub[count.index].id` / `sub[each.key].id`）具体化及 `[*]` Splat 全量扇入引用改写，拦截依赖运行期未决属性的 `count`/`for_each`（覆盖 F-004 维度 2 & 维度 3、R-005、AC-009、AC-010、EX-007）
- [x] 3.3 新增 `pkg/iacparser/dag_validator.go`，实现节点 `BlockID` 唯一性检查、展开后 `Block.DependsOn` 节点依赖图的 DFS 三色标记环路检测与拓扑层级计算（拦截循环依赖 AC-008、EX-008），并对 `ParseResult.Blocks`（按拓扑层级升序 → 同层 `import < data < resource` → 同类型 `BlockID` 字典序）及每个节点的 `DependsOn` 执行确定性排序（覆盖 NFR-A02、AC-S02）

## 4. 阶段二增量求值与点号路径精准回填（`hcl-stage2-eval-and-backfill`）

- [x] 4.1 在 `pkg/iacparser/parser.go` 与 `pkg/iacparser/eval_block.go` 中实现 `EvalReq.Validate()` 校验、`EvalReq.Config` 深拷贝隔离，并将 `EvalReq.DependencyStates` 经纯节点地址校验后按 **R-004** `module.<模块名>` 两段一组成对消费规则与 `[index]`/`["key"]` 多实例下标通过 `stateTreeNode` 还原构建为 `hcl.EvalContext` 嵌套 `cty.Value` 变量树（覆盖 F-005、R-004、R-007、AC-007、AC-013）
- [x] 4.2 在 `pkg/iacparser/eval_block.go` 中实现 `PendingExprs` 函数安全性 AST 校验（`validatePendingExprAST`）、按字典序遍历求值（挂载 `safeFunctions`）与顶层 key / 点号路径（中间容器兼容 `map[string]any`、`[]any`、`[]map[string]any`，叶子节点支持 map 键值与 `[]any` 切片下标）精准回填，对缺失依赖节点 `state`、缺失属性字段或点号路径越界/类型不匹配返回明确错误（覆盖 F-005、R-007、AC-013、AC-014、EX-009、EX-010）

## 5. 单元测试、并发安全与性能基准自动化验证

- [x] 5.1 新增 `pkg/iacparser/parser_single_module_test.go` 表驱动单元测试，覆盖单模块顶层块解析与同名 `import`/`resource` 自动依赖关联、北向三元组定位、单块属性四分类拆解（含内联对象与嵌套子块）、`lifecycle` 5 字段剥离与 `replace_triggered_by` 并入 `DependsOn`、内置函数与算术/逻辑/三元运算符、非法语法/缺失变量/未定义引用拦截及 `file()`/`templatefile()` 安全拦截（覆盖 AC-001 ~ AC-006、AC-S01）
- [x] 5.2 新增 `pkg/iacparser/parser_cross_module_test.go` 表驱动单元测试，覆盖 3 模块 5 节点工程端到端解析、模块名与资源名同为 `"module"` 的无歧义穿透、`module`/`resource`/`data` 级 `count`（含 `count = length(<resource>)` 链式推导）与 `for_each` 展开、同下标引用与 `[*]` Splat 扇入引用、跨模块 `providers` 多账号映射、模块/兄弟模块/资源/`count` 环路检测、超过 16 层递归拦截及连续 100 次调用输出 100% 确定性验证（覆盖 AC-001、AC-007 ~ AC-012、AC-S02）
- [x] 5.3 新增 `pkg/iacparser/eval_block_test.go` 表驱动单元测试，覆盖阶段二 `EvalBlock` 内置函数计算、顶层与点号路径（含 map 键值与切片下标）回填、原入参 `Config` 深拷贝零污染、多层级 `"module"` 同名节点还原、空请求/缺失依赖 `state`/缺失字段/路径越界拦截及阶段二 `file()`/`templatefile()` 禁用函数拦截（覆盖 AC-007、AC-013、AC-014、AC-S01）
- [x] 5.4 新增 `pkg/iacparser/benchmark_and_race_test.go`，实现 50 协程并发调用 `Parse` 与 `EvalBlock` 的并发安全测试、3 模块 50 节点 `BenchmarkParse_50Blocks`（断言 P99 < 50ms）以及单节点 `BenchmarkEvalBlock`（断言 P99 < 5ms）（覆盖 AC-P01、AC-P02、AC-S02）
- [x] 5.5 执行 `go test -v -race ./pkg/iacparser/...` 与 `go test -bench=. -benchmem ./pkg/iacparser/...`，确认全部单测、并发竞态检测、代码覆盖率与性能基准 100% 达标，且代码严格符合单函数 ≤ 80 行、`cvt`/`pkg/tools/maps` 使用规范

> **说明（外部环境/人工验证项）**：
> - 本需求为纯内存 Go Library（`pkg/iacparser`），无独立进程、无数据库 DDL、无外部网络或 Etcd 依赖，全部功能、安全、并发（`-race`）与性能基准（Benchmark）验收均可通过本地/CI `go test` 100% 自动化完成，无需外部真实环境人工联调。
