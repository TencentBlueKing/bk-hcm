# hcl-stage2-eval-and-backfill Specification

## Purpose
定义 `pkg/iacparser` 阶段二 `EvalBlock` 单节点增量求值与点号路径回填能力规范，涵盖 `EvalReq` 校验与 `Config` 深拷贝隔离、基于成对消费规则与多实例下标构建 `hcl.EvalContext` 嵌套变量树、`PendingExprs` 安全校验与内置函数求值（保持整数精度）、顶层键及点号路径回填，以及并发安全性与阶段二性能基准。

## Requirements
### Requirement: Stage 2 Deep-Copy Isolation and DependencyState Variable Tree Construction
阶段二 `EvalBlock(req *EvalReq)` SHALL 首先调用 `req.Validate()` 校验非空请求，并对 `req.Config` 进行递归深拷贝以避免原地修改污染调用方入参（当 `req.Config == nil` 时返回空 `map[string]any{}`）；当 `req.PendingExprs` 为空时直接返回深拷贝后的 `Config`。当 `req.PendingExprs` 非空时，`EvalBlock` SHALL 校验 `req.DependencyStates`（形如 `map[block_id]state`）中的每个 `block_id` 为不含多余属性段的合法节点地址，并按 `module.<模块名>` 两段一组成对消费规则（兼容多层级子模块、`data.`/`import.` 前缀及 `[index]`/`["key"]` 实例下标，其中字符串属性/键组装为 `cty.Object`、整数下标组装为 `cty.Tuple`）还原为 `hcl.EvalContext` 中的嵌套 `cty.Value` 变量树，同时在 Go 数值（含 `int64`、整数语义 `float64`、`json.Number`）与 `cty.Number` 双向转换时保持整数语义精度（如 `213` 保持为 `int64(213)`）。

#### Scenario: Return deep-copied Config without mutating caller input (AC-013)
- **WHEN** 调用 `EvalBlock(req)` 对包含嵌套切片/映射 `tags: []any{map[string]any{"key": "subnet_id", "value": nil}}` 的 `req.Config` 执行回填
- **THEN** 返回的新 `config` 中 `"tags.0.value"` 已更新为真实计算值，而原入参 `req.Config["tags"][0]["value"]` 仍保持为 `nil` 未被修改

#### Scenario: Reconstruct multi-level module block_id with pairwise consumption in Stage 2 (AC-007)
- **WHEN** `PendingExprs` 引用 `"module.network.module.module.tencentcloud_vpc.module.id"` 且 `DependencyStates["module.network.module.module.tencentcloud_vpc.module"]["id"] = "vpc-mod-001"`，调用 `EvalBlock(req)`
- **THEN** `EvalBlock` 准确按成对消费规则构建 `hcl.EvalContext` 变量树并求值回填出 `"vpc-mod-001"`

### Requirement: Pending Expression Evaluation and Top-Level / Dot-Path Backfill
阶段二 `EvalBlock` SHALL 在构建好的 `hcl.EvalContext`（挂载与阶段一相同的安全纯内存内置函数注册表 `safeFunctions`）中按字段路径字典序遍历求值 `req.PendingExprs` 的每一项 `(targetPath, exprStr)`：
- 求值前先对表达式语法树执行安全校验（拦截 `forbiddenFunctions` 黑名单与未注册函数），求值后将结果 `cty.Value` 转换为标准 Go 纯 JSON 类型（`any`）；
- 若 `targetPath` 不含点号 `.`，直接写入 `config[targetPath] = val`；
- 若 `targetPath` 含点号 `.`（如 `"tags.0.value"`、`"fields.0.value"` 或 `"arr_leaf.0"`），按 `.` 切分路径段：中间容器段遇到纯数字段（如 `"0"`）作为切片下标索引 `[]any` 或 `[]map[string]any`，遇到字符串段作为 `map[string]any` 的 key；最后一段叶子位置既支持字符串段写入 `map[string]any[key] = val`，也支持纯数字段写入 `[]any[idx] = val`。

#### Scenario: Evaluate pending expressions with built-in functions and backfill top-level keys and nested dot-paths (AC-013)
- **WHEN** 传入 `EvalReq`（`Config` 含 `"operator": "user"` 与 `"tags": []any{map[string]any{"key": "subnet_id", "value": nil}}`，`PendingExprs` 含 `"bk_host_innerip": "module.app.tencentcloud_instance.web.private_ip"`、`"bk_biz_id": "tostring(data.bkcmdb_biz.target.bk_biz_id)"`、`"tags.0.value": "module.network.tencentcloud_subnet.app.id"`，以及对应的 `DependencyStates`）并调用 `EvalBlock(req)`
- **THEN** 返回 `err == nil`，输出的 `config` 中 `"bk_host_innerip": "127.0.0.1"`、`"bk_biz_id": "213"`、`"tags": []any{map[string]any{"key": "subnet_id", "value": "subnet-demo-01"}}`

#### Scenario: Reject nil EvalReq, missing dependency block_id, missing state attribute, or out-of-bounds dot-path (EX-009, EX-010, AC-014)
- **WHEN** 调用 `EvalBlock(req)` 时传入 `req == nil`，或 `EvalReq.PendingExprs` 引用的 `block_id` 在 `EvalReq.DependencyStates` 中缺失、或对应 `state` 中缺失被引用的字段、或点号路径（如 `"tags.5.value"`）在 `Config` 骨架中越界/类型不匹配
- **THEN** `EvalBlock(req)` 返回非 `nil` 错误，明确指出请求为空、缺失的依赖节点 ID、属性名或路径越界/类型不匹配错误

#### Scenario: Reject forbidden host file functions in Stage 2 PendingExprs (AC-S01)
- **WHEN** `EvalReq.PendingExprs` 中包含调用 `file("config.txt")` 或 `templatefile(...)` 的表达式并调用 `EvalBlock(req)`
- **THEN** `EvalBlock(req)` 100% 返回禁用/不支持函数错误，不发生任何宿主机磁盘文件读取

### Requirement: Concurrency Safety and Stage 2 Evaluation Performance
`Parser` 实例 SHALL 无可变全局共享状态，支持多协程并发安全调用 `Parse` 与 `EvalBlock`，且在包含 5 条未决表达式（含内置函数与点号路径回填）及 5 个依赖节点 `state` 的单节点 `EvalReq` 下，单次 `EvalBlock` P99 耗时 MUST `< 5ms`。

#### Scenario: Verify 50-goroutine concurrent Parse and EvalBlock execution under race detector (AC-S02)
- **WHEN** 运行 `go test -race ./pkg/iacparser/...` 启动 50 个并发协程同时调用同一 `Parser` 实例的 `Parse` 与 `EvalBlock`
- **THEN** 全部协程执行成功，零数据竞争（data race）、零 panic

#### Scenario: Meet Stage 2 EvalBlock performance benchmark (AC-P02)
- **WHEN** 运行基准测试 `go test -bench=BenchmarkEvalBlock -benchmem`（单节点含 5 条未决表达式与 5 个依赖节点 `state`）
- **THEN** 单次 `EvalBlock` 执行耗时 < 5ms
