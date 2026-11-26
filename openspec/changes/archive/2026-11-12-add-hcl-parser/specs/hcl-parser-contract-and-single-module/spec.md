## ADDED Requirements

### Requirement: Public Parser Contract and Core Data Models
公共包 `pkg/iacparser`（包名 `package iacparser`）及配套公共枚举常量包 `pkg/criteria/enumor`、`pkg/criteria/constant` SHALL 定义两阶段 HCL 语法解析器的对外接口契约与核心数据模型，作为纯 Go Library 交付（不暴露独立 HTTP 接口）：
- 在 `pkg/criteria/enumor` 中定义 `BlockKind` 枚举类型、常量（`BlockKindResource = "resource"`、`BlockKindData = "data"`、`BlockKindImport = "import"`）及 `(k BlockKind) Validate() error` 校验方法；在 `pkg/criteria/constant` 中定义 `RootModuleKey = "."` 与 `MaxModuleDepth = 16`；
- 在 `pkg/iacparser` 中定义 `Parser` 接口，包含 `Parse(req *ParseReq) (*ParseResult, error)` 与 `EvalBlock(req *EvalReq) (map[string]any, error)` 两个方法，并提供无状态构造函数 `NewParser() Parser`；
- 在 `pkg/iacparser` 中定义 `ParseReq`（含 `Validate() error`）、`ParseResult`、`Block`（`Kind` 字段类型为 `enumor.BlockKind`）、`Lifecycle`、`EvalReq`（含 `Validate() error`）结构体，其字段名、类型与 JSON Tag MUST 严格对齐技术方案契约；当节点未配置 `lifecycle` 时 `Block.Lifecycle` MUST 为 `nil`，当节点无未决属性时 `Block.PendingExprs` MUST 为 `nil`，当节点无依赖时 `Block.DependsOn` MUST 为非 `nil` 空切片 `[]string{}`。

#### Scenario: Instantiate stateless parser and validate default empty fields
- **WHEN** 调用 `NewParser()` 创建解析器并解析不含未决属性与 `lifecycle` 块的单资源根模块，同时调用 `enumor.BlockKindResource.Validate()`
- **THEN** `BlockKind.Validate()` 返回 `nil`（非法枚举值返回错误）；返回的 `Block` 中 `Lifecycle` 为 `nil`，`DependsOn` 为 `[]string{}`，`PendingExprs` 为 `nil`

#### Scenario: Reject nil ParseReq, empty Modules, or missing root module (EX-001, AC-002)
- **WHEN** 调用 `Parse(req)` 时传入 `req == nil`、`len(req.Modules) == 0` 或 `req.Modules[constant.RootModuleKey]`（`"."`）缺失/为空
- **THEN** `Parse` 返回非 `nil` 参数校验错误，明确提示请求为空或缺失根模块 `"."`

### Requirement: Safe In-Memory Built-in Function Registry
`pkg/iacparser` SHALL 构建并在阶段一与阶段二之间复用纯内存、无副作用的只读 `cty` 内置函数注册表（`safeFunctions`），包含 `go-cty/cty/function/stdlib` 纯内存标准函数以及常用 Terraform 类型转换、集合与编解码/哈希函数（`tostring`、`tonumber`、`tobool`、`tolist`、`toset`、`tomap`、`lookup`、`length`、`index`、`try`、`can`、`base64encode`、`base64decode`、`md5`、`sha1`、`sha256`、`urlencode` 等），并维护禁用函数黑名单（`forbiddenFunctions`：`file`、`fileexists`、`fileset`、`filebase64`、`templatefile`、`pathexpand`、`abspath`），MUST NOT 注入任何访问宿主机文件系统、路径展开或外部模板网络的函数。

#### Scenario: Evaluate standard pure built-in functions in static expressions
- **WHEN** 模块属性表达式在阶段一使用 `upper()`、`lower()`、`join()`、`lookup()`、`tostring()`、`tonumber()`、`tobool()`、`tolist()`、`toset()`、`tomap()`、`index()`、`try()`、`can()`、`base64encode()`、`md5()`、`sha256()`、`urlencode()` 等内置函数对静态字面量或输入变量求值
- **THEN** `Parse` 成功计算表达式结果并写入 `Block.Config`

#### Scenario: Reject forbidden host file or template functions (EX-005, AC-S01)
- **WHEN** HCL 属性表达式中调用了 `file("config.txt")` 或 `templatefile(...)` 等宿主机文件/模板函数并调用 `Parse(req)`
- **THEN** `Parse` 立即返回非 `nil` 的禁用/不支持函数错误，且不发生任何宿主机磁盘文件读取

### Requirement: Stage 1 Top-Level Block Classification, Bridge Block Digestion, and Northbound Triplet Extraction
阶段一 `Parse` SHALL 使用 `github.com/hashicorp/hcl/v2/hclsyntax` 解析各模块下的 `.tf` 文件字节流（同模块存在多个 `.tf` 文件时按文件名升序合并解析，并拦截顶层游离属性与不支持的顶层块类型），识别 8 类顶层块：
- 将 `variable`、`locals`、`provider`、`module`、`output` 5 类桥梁块在阶段一内部穿透消化，MUST NOT 将桥梁块输出到 `ParseResult.Blocks` 中；
- 仅将 `resource`、`data`、`import` 3 类可调度块生成 `Block` 节点，提取北向定位三元组（默认 `plugin` 取 `type_name` 首个 `_` 前的前缀，若块内显式声明 `provider = <plugin>` 或 `provider = <plugin>.<alias>` 元参数则按指定插件与 `alias` 从当前模块匹配的 `provider` 获取 `plugin_instance_id`，`type_name` 为块类型名）；
- `import` 块的 `block_id` SHALL 带 `import.` 前缀，`Config` MUST 固定为 `{"import_id": "<计算后的静态字符串 id>"}`；若同模块内仅定义了 `import { to = <type>.<name> ... }` 而未定义同名 `resource` 块，其他节点对 `<type>.<name>.<attr>` 的引用 SHALL 自动映射至 `import.<type>.<name>.<attr>`；若同模块内同时定义了 `import` 与同名 `resource "<type>" "<name>"`，该 `resource` 节点 SHALL 自动将同模块的 `import.<type>.<name>` 加入其 `DependsOn`。

#### Scenario: Parse 8 top-level block types and emit only resource, data, and import blocks (AC-001)
- **WHEN** 输入包含 `variable`、`locals`、`provider`、`output`、`data`、`import`、`resource` 的 HCL 模块及必填 `vars` 调用 `Parse(req)`
- **THEN** `variable`、`locals`、`provider`、`output` 被内部消化不生成 `Block`；`ParseResult.Blocks` 准确产出 `"import"`、`"data"`、`"resource"` 三类节点，北向三元组（`plugin`、`plugin_instance_id`、`type_name`）提取准确，且 `import` 节点的 `Config` 为 `{"import_id": "vpc-existing-001"}`

#### Scenario: Automatically link resource dependency when both import and matching resource are defined
- **WHEN** 同一模块中同时定义了 `import { to = tencentcloud_vpc.legacy, id = "vpc-legacy-001" }` 与同名 `resource "tencentcloud_vpc" "legacy" { name = "corp-vpc" }` 并调用 `Parse(req)`
- **THEN** `ParseResult.Blocks` 依次产出 `import.tencentcloud_vpc.legacy` 与 `tencentcloud_vpc.legacy` 两个节点，且 `tencentcloud_vpc.legacy` 的 `DependsOn` 自动包含 `["import.tencentcloud_vpc.legacy"]`

#### Scenario: Reject HCL syntax errors or missing required variable without default (EX-002, EX-003, AC-002)
- **WHEN** 输入的 `.tf` 内容包含未闭合花括号语法错误，或定义了无 `default` 的 `variable "subnet_cidr" {}` 但未传入 `"subnet_cidr"` 变量值
- **THEN** `Parse(req)` 返回非 `nil` 错误，错误消息明确包含语法错误位置或缺失变量名 `"subnet_cidr"`

#### Scenario: Reject unbound provider or invalid resource type_name without underscore (EX-004, AC-012)
- **WHEN** 资源 `type_name` 不包含下划线 `_`，或资源对应的 `plugin` 在当前模块未能匹配到任何 `provider` 的 `plugin_instance_id`
- **THEN** `Parse(req)` 返回非 `nil` 错误，提示 `type_name` 格式非法或缺失对应的 `provider` 配置

### Requirement: Stage 1 Four-Category Attribute Decomposition and Lifecycle Extraction
阶段一 `Parse` SHALL 对每个可调度块的 HCL Body 执行四分类拆解：
1. **普通静态属性**：仅依赖字面量、已决 `var`/`local`、`count.index`/`each.key`/`each.value`、运算符（含条件静态为布尔值时的三元表达式命中分支）及内置函数的属性，在阶段一直接求值并转换为标准 Go 基础类型写入 `Block.Config`；
2. **顶层未决属性**：直接或间接引用任何 `resource`/`data`/`import` 节点属性的顶层标量/表达式属性，将其中的桥梁引用全部替换为源头节点的完整 `block_id` 属性路径（混合插值、函数参数或运算表达式中的已知静态子表达式内联替换为字面量），存入 `Block.PendingExprs["<attr_name>"]`，且不在 `Block.Config` 中设置该顶层 key，同时将引用的源头 `block_id` 去重加入 `Block.DependsOn`；
3. **嵌套子块与内联对象/对象列表内部的未决属性**：同名嵌套子块按出现顺序组装为 `[]any`（元素为 `map[string]any`）放入 `Block.Config`，且属性值为内联花括号对象 `{...}` 或内联对象列表 `[{...}, {...}]` 的结构化属性同样按字段拆解为 `map[string]any` 或 `[]any` 骨架放入 `Block.Config`；已决叶子字段填入计算值，未决叶子字段在 `Block.Config` 对应位置填 `nil`（JSON `null`）占位，并以点号路径（如 `"tags.0.value"`、`"inline_obj.dyn_k"`、`"inline_arr.0.v"`）为 key 将穿透改写后的源头引用表达式存入 `Block.PendingExprs`，同时将源头 `block_id` 加入 `Block.DependsOn`；
4. **`lifecycle` 子块与元参数剥离**：将 `lifecycle` 块解析进 `Block.Lifecycle`（支持 `create_before_destroy`、`prevent_destroy`、`ignore_changes`、`replace_triggered_by`、`recompute` 五个字段），其中 `ignore_changes` 兼容无引号属性遍历表达式（如 `[operator, tags[0].value]` 规范化为 `["operator", "tags.0.value"]`）与关键字 `all`，`replace_triggered_by` 中的引用补齐所属模块前缀与下标并将源头资源 `block_id` 自动并入 `Block.DependsOn`；显式 `depends_on` 解析并入 `Block.DependsOn`（若指向不带下标的多实例资源则自动展开为其全部实例 `block_id`，若不关联任何 `resource`/`data`/`import` 则返回错误）；`lifecycle`、`depends_on`、`provider`、`count`、`for_each` 元参数 MUST NOT 进入 `Block.Config`。

#### Scenario: Decompose static attributes, top-level pending attributes, and nested/inline skeleton dot-path placeholders (AC-003)
- **WHEN** `resource "bkcmdb_host" "web"` 包含静态属性 `operator = "user"`、混合插值未决属性 `host_tag = "${var.env}-${tencentcloud_instance.web.id}"`（`var.env = "prod"`）、顶层未决属性 `bk_host_innerip` 与 `bk_biz_id`、以及嵌套子块 `tags { key = "subnet_id", value = tencentcloud_subnet.app.id }` 并调用 `Parse(req)`
- **THEN** 该节点的 `Config` 为 `{"operator": "user", "tags": []any{map[string]any{"key": "subnet_id", "value": nil}}}`，`PendingExprs` 包含 `"host_tag": "\"prod-${tencentcloud_instance.web.id}\""`、`"bk_host_innerip"`、`"bk_biz_id"` 与 `"tags.0.value": "tencentcloud_subnet.app.id"`，且所有被引用的源头节点 `block_id` 已加入 `DependsOn`

#### Scenario: Reject references to undefined resources, locals, or variables (EX-006, AC-004)
- **WHEN** 资源属性表达式引用了不存在的资源 `tencentcloud_vpc.not_exist.id` 或未定义的局部变量 `local.unknown_val`
- **THEN** `Parse(req)` 返回非 `nil` 错误，明确指出未定义的引用目标名称

#### Scenario: Extract lifecycle block with 5 fields and merge replace_triggered_by into DependsOn (AC-005)
- **WHEN** 子模块 `./modules/app` 内的 `resource "bkcmdb_host" "web"` 配置了 `lifecycle { create_before_destroy = true, prevent_destroy = true, ignore_changes = [operator, tags[0].value], replace_triggered_by = [tencentcloud_instance.web.image_id], recompute = ["bk_module_id"] }` 并调用 `Parse(req)`
- **THEN** `Block.Config` 中不包含 `"lifecycle"` 键；`Block.Lifecycle` 中 `CreateBeforeDestroy=true`、`PreventDestroy=true`、`IgnoreChanges=["operator", "tags.0.value"]`、`ReplaceTriggeredBy=["module.app.tencentcloud_instance.web.image_id"]`、`Recompute=["bk_module_id"]`；且 `"module.app.tencentcloud_instance.web"` 已并入 `Block.DependsOn`

#### Scenario: Reject unsupported lifecycle attributes or non-existent replace_triggered_by targets (EX-006, AC-006)
- **WHEN** `lifecycle` 块中包含不支持的属性名（如 `unknown_opt = true`）或 `replace_triggered_by` 引用了不存在的资源 `tencentcloud_instance.none.id`
- **THEN** `Parse(req)` 返回非 `nil` 错误，提示非法 `lifecycle` 属性或引用目标不存在
