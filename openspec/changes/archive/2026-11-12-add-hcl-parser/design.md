## Context

在海垒（HCM）插件化改造与 IaC 编译调度架构（M2 里程碑）中，终端用户提交的是包含根模块 `"."` 与多个子模块 `module` 的多模块 HCL 工程（涉及 `variable`、`locals`、`provider`、`module`、`output`、`resource`、`data`、`import` 八类顶层块，以及跨模块传参、`count`/`for_each` 多实例展开、内置函数计算与 `lifecycle` 元参数）。而底层 `hc-service` 北向接口（`POST /api/v1/hc/iac_plugins/{execute|read_data|import|plan|validate}`）只接受**单个节点的纯 JSON 配置（`config`）**与北向定位三元组（`plugin`、`plugin_instance_id`、`type_name`）。

若每次上游依赖节点执行完都重新解析整份 `.tf` 工程，不仅产生大量重复 AST 解析与静态计算开销，也无法为 `iac-server` DAG 编排提前提供全图拉平的节点依赖关系（`depends_on`）与生命周期控制属性（`lifecycle`）。因此，需在公共包 `pkg/iacparser` 下设计并实现无状态、纯内存的两阶段 HCL 语法解析与增量求值器：
- **阶段一 `Parse(req *ParseReq) (*ParseResult, error)`**：在编排前调用一次，静态消化全部桥梁块、穿透跨模块引用、展开 `count`/`for_each` 多实例，产出拉平的可调度 `Block` 列表（含已算好的静态 `config`、未决字段表达式字典 `pending_exprs`、源头 `depends_on` 与 `lifecycle`）；
- **阶段二 `EvalBlock(req *EvalReq) (map[string]any, error)`**：在运行期上游依赖节点执行完成后调用，将 `dependency_states` 注入求值上下文计算 `pending_exprs`，并按顶层键或点号路径（如 `tags.0.value`）精准回填至 `config` 副本，产出 100% 纯 JSON 的北向 `config`。

## Goals / Non-Goals

**Goals:**
- 在 `pkg/criteria/enumor` 与 `pkg/criteria/constant` 中定义 `BlockKind` 枚举及 `MaxModuleDepth`、`RootModuleKey` 常量；在 `pkg/iacparser`（`package iacparser`）下交付纯 Go Library 形态的 `Parser` 接口、`NewParser()` 构造函数及 `ParseReq`、`ParseResult`、`Block`、`Lifecycle`、`EvalReq` 核心数据模型。
- 引入官方 `github.com/hashicorp/hcl/v2 v2.24.0` 并结合仓库已有的 `github.com/zclconf/go-cty v1.16.4`，构建阶段一与阶段二共用的安全纯内存内置函数注册表（`cty/function/stdlib` 纯内存函数 + `tostring`/`tonumber`/`tobool`/`tolist`/`toset`/`tomap`/`lookup`/`length`/`index`/`try`/`can` 与哈希/编解码等常用 Terraform 函数，显式黑名单拦截宿主机文件/路径与模板类函数）。
- 完整实现阶段一 8 类顶层块解析、5 类桥梁块内部消化、3 类可调度节点（`resource`/`data`/`import`）北向三元组定位、单块属性四分类拆解（静态 `config`、顶层 `pending_exprs`、嵌套子块 `[]any` 与内联对象/对象列表骨架 `nil` 占位及点号路径、`lifecycle` 5 字段剥离）。
- 完整实现跨模块与多实例四维展开：多层级 `module` 穿透与 `module.<模块名>` 成对消费无歧义路径切分、`module`/`resource`/`data` 级 `count`/`for_each` 展开（含 `count = length(<resource>)` 链式实例数推导）、同下标精确引用与 `[*]` Splat 全量扇入引用、跨模块 `providers` 多账号映射，以及全维度环路检测、16 层递归深度上限与三级确定性拓扑排序。
- 完整实现阶段二 `EvalBlock`：`EvalReq` 校验与 `Config` 深拷贝隔离、`DependencyStates` 到 `hcl.EvalContext` 嵌套 `cty.Value` 变量树还原（保持整数精度）、`PendingExprs` 函数安全二次校验与求值及顶层键/点号路径（含 map 键值与切片元素下标）精准回填。
- 满足性能与并发安全指标：3 模块 50 节点 `Parse` P99 `< 50ms`、单节点 `EvalBlock` P99 `< 5ms`、50 协程 `-race` 零数据竞争。

**Non-Goals:**
- 不单独暴露 `/parse` 或 `/eval_block` HTTP 接口（由 `iac-server` / `task-server` 直接通过 Go 包调用）。
- 不负责 Plan / Apply DAG 拓扑编排（如 CBD 向上游传递、替换动作拆分为创建与销毁节点、销毁边反向等）。
- 不涉及 Spec 版本托管、State 数据库读写或对 `hc-service` 北向 HTTP 接口的实际网络调用。
- 不涉及两个 HCL 版本的文本对比与参数模版渲染（属 M3 范围）。

## Decisions

### 1. 代码包职责切分与编码规范约束
- **决策**：遵循项目公共规范将跨包复用的枚举与常量沉淀至 `pkg/criteria`，并将 `pkg/iacparser` 按职责拆分为高内聚、低耦合的多个源文件，严格保证每个函数长度不超过 80 行、非必要不使用 `else`、指针转换统一使用 `cvt "hcm/pkg/tools/converter"`、map 克隆/合并/键提取优先使用 `hcm/pkg/tools/maps`：
  - `pkg/criteria/constant/iac.go`：定义子模块最大嵌套深度 `MaxModuleDepth = 16` 与根模块固定键 `RootModuleKey = "."`；
  - `pkg/criteria/enumor/iac.go`：定义 `BlockKind` 枚举类型及常量（`BlockKindResource`、`BlockKindData`、`BlockKindImport`）与 `Validate()` 方法；
  - `pkg/iacparser/types.go`：定义 `ParseReq`、`ParseResult`、`Block`（`Kind` 为 `enumor.BlockKind`）、`Lifecycle`、`EvalReq` 结构体及 `Validate()` 基础入参校验；
  - `pkg/iacparser/parser.go`：定义 `Parser` 接口、无状态实现 `parserImpl`、`NewParser()` 构造函数，以及阶段一 `Parse` 与阶段二 `EvalBlock` 的顶层主流程编排；
  - `pkg/iacparser/cty_conv.go`：封装 Go `any`（含 `*bool`/`*string`/`*int64` 指针、各宽度整数/浮点/`json.Number`、切片、map）与 `cty.Value` 的双向无损转换、`ctyValueToHCLString` 字面量序列化及 `deepCopyMap`/`deepCopyAny` 深拷贝工具；
  - `pkg/iacparser/functions.go`：预构建只读的安全纯内存 `cty` 内置函数表 `safeFunctions map[string]function.Function` 与禁用函数黑名单 `forbiddenFunctions`；
  - `pkg/iacparser/address.go`：基于 `hclsyntax.ParseTraversalAbs` 实现支持单段连续多维下标（`addrSegment.Indices`）的地址段解析、`consumePairwiseAddress` 两段一组成对消费路径切分器、`block_id` 拼接与 `extractPluginFromTypeName` 工具；
  - `pkg/iacparser/module_loader.go`：负责单模块多 `.tf` 文件按文件名升序解析、拦截顶层游离属性与不支持的顶层块类型，并将 8 类顶层块分类提取至 `rawModule`；
  - `pkg/iacparser/expr_eval.go`：负责阶段一基于 `hclsyntax.Expression` AST 的统一静态求值与动态引用改写（`evalExpr`、`tryParseRefChain`、`resolveRefChain`），覆盖字面量、插值模板、函数调用（含 `...` 可变参数展开与 `length(<resource>)` 静态试算收敛）、列表/对象构造、`[*]` Splat、一元/二元/三元运算（含静态条件分支短路裁剪）及桥梁引用穿透与源头 `block_id` 依赖收集；
  - `pkg/iacparser/block_decomposer.go`：负责北向三元组定位（含块级 `provider` 元参数）、`import` 块拆解与同名 `resource` 自动依赖关联、单块属性四分类拆解（静态属性、顶层未决属性、内联对象 `{...}` / 对象列表 `[{...}]` 与嵌套子块 `[]any` 的 `nil` 占位及点号路径提取、`lifecycle` 5 字段剥离与显式 `depends_on` 收集）；
  - `pkg/iacparser/expander.go`：负责带 AST 缓存 `parsedCache` 的多层级 `module` 递归展开、`"./"` 模块路径双向兼容查找、同级 `module` 调用拓扑排序、模块 `variable`/`provider` 绑定、按需节点实例展开（含 `count`/`for_each` 循环检测）、`locals` 惰性求值与环检测、跨模块 `providers` 多账号映射，以及最大递归深度（`<= constant.MaxModuleDepth`）保护；
  - `pkg/iacparser/dag_validator.go`：负责节点 `BlockID` 唯一性检查、展开后 `Block.DependsOn` 节点依赖图的 DFS 三色标记环路检测与拓扑层级计算，以及三级确定性拓扑排序（拓扑层级升序 → 同层 `import < data < resource` → 同类型 `BlockID` 字典序）；
  - `pkg/iacparser/eval_block.go`：负责阶段二基于 `stateTreeNode` 前缀树（区分对象字典 `children` 与数组元组 `elemList`）将 `DependencyStates` 组装为 `hcl.EvalContext`、`validatePendingExprAST` 禁用/未知函数二次拦截、`evalPendingExpr` 求值与 `backfillConfigField` 顶层键/点号路径回填。
- **理由**：将公共枚举常量统一收敛到 `pkg/criteria`，并按阶段一（词法语法加载 → 模块与多实例展开 → 表达式求值改写与单块四分类拆解 → DAG 校验与拓扑排序）和阶段二（变量树还原 → 安全校验与增量求值回填）流水线切分文件，确保单函数严格控制在 80 行以内且易于编写表驱动单测。
- **备选方案**：将全部解析与求值逻辑堆叠在单个 `parser.go` 文件中——会导致文件超千行、函数过长且违反团队编码规范，予以排除。

### 2. 安全纯内存内置函数注册表与预构建复用
- **决策**：在包初始化时一次性预构建只读 `safeFunctions map[string]function.Function` 与 `forbiddenFunctions map[string]bool`：
  - 纳入 `github.com/zclconf/go-cty/cty/function/stdlib` 中的纯内存计算函数：数学计算（`abs`、`ceil`、`floor`、`log`、`max`、`min`、`parseint`、`pow`、`signum`）、字符串处理（`chomp`、`format`、`formatlist`、`indent`、`join`、`lower`、`regex`、`regexall`、`replace`、`split`、`strrev`、`substr`、`title`、`trim`、`trimprefix`、`trimsuffix`、`trimspace`、`upper`）、集合操作（`chunklist`、`coalesce`、`coalescelist`、`compact`、`concat`、`contains`、`distinct`、`element`、`flatten`、`keys`、`lookup`、`merge`、`range`、`reverse`、`setintersection`、`setproduct`、`setsubtract`、`setunion`、`slice`、`sort`、`values`、`zipmap`）、编解码与时间格式化（`csvdecode`、`formatdate`、`jsondecode`、`jsonencode`）；
  - 补充注册 Terraform 常用扩展函数：基于 `stdlib.MakeToFunc` 封装类型转换（`tostring`、`tonumber`、`tobool`、`tolist`、`toset`、`tomap`），自定义封装 `length`（同时支持字符串长度与集合元素个数）、`index`（列表元素下标查找）、容错函数 `try` 与 `can`，以及纯内存哈希/编解码函数 `base64encode`、`base64decode`、`md5`、`sha1`、`sha256`、`urlencode`；
  - **显式黑名单拦截**任何读取宿主机磁盘文件、路径展开或外部模板渲染的函数（`forbiddenFunctions`：`file`、`fileexists`、`fileset`、`filebase64`、`templatefile`、`pathexpand`、`abspath`），并在阶段一 `evalFunctionCallExpr` 与阶段二 `validatePendingExprAST` 中双重拦截黑名单及任何未注册函数。
- **理由**：多租户服务端环境下，若允许 HCL 调用 `file("config.txt")` 等宿主机文件函数将造成任意文件读取风险；预构建只读函数表可在并发调用 `Parse` 与 `EvalBlock` 时零锁复用，消除重复构建开销，保障 `EvalBlock` `< 5ms` 性能目标。

### 3. `module.<模块名>` 两段一组成对消费算法与全链路引用穿透
- **决策**：基于 `hclsyntax.ParseTraversalAbs` 将引用地址解析为支持连续多维下标（`addrSegment.Indices`，如 `az_map["gz"][1]`、`module.vpc["gz.01"]`）的 `[]addrSegment`，在阶段一重写未决引用与阶段二将 `DependencyStates` 还原为 `hcl.EvalContext` 变量树时，统一采用确定性的左向右扫描状态机（`consumePairwiseAddress`）：
  1. 当当前段为 `"module"`（无下标）且存在下一段时，将 `"module"` 与下一段 `<模块名>[可选下标]` 作为一对模块前缀消费，循环继续；
  2. 遇到第一个非 `"module"` 段（或虽名为 `"module"` 但已处于模块前缀消费完毕后的类型位置）时停止模块前缀消费：
     - 若该段为 `"data"` 或 `"import"`（无下标），则校验下一段 `type_name` 包含 `_`，并消费 `[kind, type_name, name(含可选下标)]` 共 3 段作为模块内节点标识；
     - 否则要求该段包含 `_`（`<plugin>_<res>`），消费 `[type_name, name(含可选下标)]` 共 2 段作为模块内资源节点标识；
  3. 剩余段全部作为节点的 `state` 属性访问路径（如 `.id`、`.private_ip`、`.matrix[0][1]`）。
  4. 对于同模块内仅定义了 `import { to = <type>.<name> ... }` 而未定义同名 `resource "<type>" "<name>"` 的纯导入场景，符号表将该模块内的 `<type>.<name>` 自动绑定至 `import.<type>.<name>` 节点；若同时定义了 `import` 与同名 `resource`，则 `resource` 自动将同模块 `import.<type>.<name>` 加入其 `DependsOn`。
- **理由**：即使模块名、资源名或属性名本身取名为 `"module"`（如 `module.network.module.module.tencentcloud_vpc.module.id`），或 `for_each` 字符串键包含点号，基于 HCL 原生 Traversal 解析 + 两段一组成对消费 + 下划线类型特征判定也能 100% 零歧义切分出 `block_id` 与属性路径。

### 4. 混合表达式静态内联、结构化骨架点号路径占位与 `[*]` Splat 改写策略
- **决策**：
  - **静态 vs 未决判定与条件裁剪**：在 `expr_eval.go` 中递归求值 `hclsyntax.Expression`：若所有引用的 `var`/`local`/`count.index`/`each.*` 均为静态 `cty.Value` 且不依赖未决节点属性，则直接算出 `cty.Value`；对于三元表达式 `cond ? trueExpr : falseExpr`，若 `cond` 静态可求值为布尔值，则仅求值命中分支以避免引入未命中分支的多余依赖；
  - **未决表达式重写与链式实例数试算**：若表达式直接或经由 `var`/`local`/`module.output` 间接引用了未决节点，则将静态已知子表达式通过 `ctyValueToHCLString` 内联替换为字面量，将未决桥梁引用替换为源头节点的完整 `block_id` 路径；对于多实例节点整体引用或 `sub[*].id` 全量 Splat 引用，将该多实例资源展开后的全部实例 `block_id` 加入 `DependsOn`，改写为列表表达式 `[sub[0].id, sub[1].id, ...]`，同时在 `evalValue.ctyVal` 中附带等长 `cty.DynamicVal` 占位元组，从而使 `count = length(aws_instance.a)` 在阶段一试算时通过 `IsWhollyKnown()` 直接收敛为静态整数实例数；
  - **嵌套子块与内联对象/对象列表骨架拆解**：同名嵌套子块（如多个 `tags { ... }`）在 `Block.Config` 中按出现顺序组装为 `[]any`（每个元素为 `map[string]any`）；同时对属性值为内联花括号对象 `{...}`（`ObjectConsExpr`）或内联对象列表 `[{...}, {...}]`（元素均为 `ObjectConsExpr` 的 `TupleConsExpr`）的属性执行逐字段骨架拆解（`tryDecomposeStructuredAttr`）。递归拆解时维护当前点号路径前缀（如 `tags.0`、`inline_obj`、`inline_arr.0`），若叶子属性未决，则在 `Config` 对应位置写入 `nil` 占位，并将改写后的表达式登记到 `PendingExprs["tags.0.value"]` 等点号路径键下。
- **理由**：确保阶段二 `EvalBlock` 无需再次访问 `.tf` 源码或模块符号表，仅凭单节点的 `Config`、`PendingExprs` 与 `DependencyStates` 即可用标准 `hclsyntax.ParseExpression` 直接求值并按点号路径回填。

### 5. 阶段二 `DependencyStates` 变量树构建与数值精度保持
- **决策**：
  - **入参校验与深拷贝隔离**：在 `EvalBlock` 入口调用 `req.Validate()` 校验非空请求，并对 `req.Config` 进行递归深拷贝（对 `map[string]any` 使用 `maps.Clone` 结合递归切片/子 map 拷贝，`nil` 安全转为空 `map[string]any{}`），保证并发或重复调用绝不污染原 `req.Config`；
  - **`stateTreeNode` 多实例与多模块 `cty.Value` 树组装**：校验 `DependencyStates` 中每个 `block_id` 为不含多余属性段的合法节点地址后，使用 `stateTreeNode` 前缀树分别维护字符串键子节点（`children map[string]*stateTreeNode` → 转为 `cty.ObjectVal`）与整数下标子节点（`elemList []*stateTreeNode` → 转为 `cty.TupleVal`），注入 `hcl.EvalContext.Variables`，使得 `hclsyntax.ParseExpression` 原生支持 `.` 属性访问、`[0]` 整数索引与 `["gz"]` 字符串键索引；
  - **点号路径容器与叶子回填**：按字典序求值 `PendingExprs`（求值前经 `validatePendingExprAST` 拦截禁用或未知函数），回填时中间路径兼容 `map[string]any`、`[]any` 与 `[]map[string]any`，最后一段叶子既支持写入 `map[string]any` 键值（如 `"tags.0.value"`），也支持写入 `[]any` 切片元素（如 `"arr_leaf.0"`）；
  - **整数精度保持**：当将 Go `float64`（如 JSON 反序列化产生的 `float64(213)`）转入 `cty.Number` 时，对无小数部分且在 `[math.MinInt64, math.MaxInt64)` 范围内的值转为精确整数 `cty.NumberIntVal`；当将 `cty.Number` 转换回 Go `any` 时，使用 `bf := val.AsBigFloat()` 检查 `bf.IsInt()` 且 `bf.Int64()` 为 `big.Exact` 时返回 `int64`，否则返回 `float64`，确保 `bk_biz_id: 213` 经过 `cty` 计算或 `tostring()` 转换时保持整数语义，不产生科学计数法或浮点精度丢失。

## Risks / Trade-offs

- **[Risk] `locals`、同级 `module` 或资源 `count`/`for_each` 之间存在复杂依赖顺序或循环引用** → **Mitigation**：对模块内 `locals`（`visitingLocals`）与资源实例展开（`visitingNodes`）采用按需惰性求值 + 访问栈环检测；对同级 `module` 调用与最终 `Block.DependsOn` 采用 DFS 三色标记拓扑排序与环检测；对子模块调用栈维护 `depth` 计数器（上限 `constant.MaxModuleDepth = 16` 层）与调用路径集合，一旦发现环路或超限立即返回清晰的错误诊断。
- **[Risk] Go `map` 遍历无序导致多次 `Parse` 产出的 `Blocks` 或 `DependsOn` 顺序抖动** → **Mitigation**：对模块内 `.tf` 文件名、属性名、`for_each` 的键集合、每个节点的 `DependsOn` 去重切片均按字典序排序，并对最终 `ParseResult.Blocks` 按“拓扑层级升序 → 同层 `import < data < resource` → 同类型 `BlockID` 字典序”显式排序，通过连续 100 次调用 JSON 序列化一致性单测（AC-S02）强校验。
- **[Risk] `EvalBlock` 点号路径回填时遇到 `[]map[string]any` 与 `[]any` 类型差异** → **Mitigation**：点号路径寻址器同时兼容 `[]any` 与 `[]map[string]any` 切片下标访问及 `[]any` 叶子元素回填，并在下标越界、非数字下标访问切片或对非容器节点继续下钻时返回明确的回填错误（EX-010）。

## Migration Plan

1. **依赖引入与代码合入**：在 `go.mod` 中引入 `github.com/hashicorp/hcl/v2 v2.24.0` 直接依赖，合入 `pkg/criteria/constant/iac.go`、`pkg/criteria/enumor/iac.go` 及 `pkg/iacparser` 全量实现与单元测试。
2. **下游消费集成**：后续 `iac-server` 的 Plan/Apply DAG 编排模块与 `task-server` 调度模块直接 `import "hcm/pkg/iacparser"` 并调用 `NewParser()` 使用。
3. **回滚策略**：本变更为纯新增公共库 `pkg/iacparser` 及配套枚举常量，不涉及任何微服务启动配置、HTTP 路由或数据库表结构变更；若需回滚，直接 revert 对应 commit 即可，对存量系统零影响。

## Open Questions

无（代码包路径 `pkg/iacparser`、纯 Go Library 交付形态、安全内置函数白名单范围及 `import` 节点引用穿透规则均已在需求澄清阶段全部确认闭环）。
