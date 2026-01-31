---
name: code-review
description: 深度代码审查工具：分析 Git 变更，评估逻辑、代码规范、安全及需求一致性，支持CodeReview报告。
---

# 代码审查技能

执行超越 lint 的深度分析，识别安全漏洞、并发问题、潜在 Bug、架构设计问题，并生成专业报告。

## 工作流程

**0.配置** → **1.需求** → **2.代码变更** → **3.Lint** → **4.架构分析** → **5.深度审查** → **6.报告** → **7.Todo清单**

---

## 步骤 0: 读取配置

检查项目根目录 `.codereview` 配置（参考 `assets/.codereview.example`）：
- `exclude_paths`: 跳过的文件/目录
- `ignore_categories`/`ignore_rules`: 忽略的审查类别/规则
- `severity`: 严格程度（strict/normal/relaxed）

---

## 步骤 1: 需求收集

根据用户输入类型选择操作：

| 输入类型 | 操作                                                           |
|------|--------------------------------------------------------------|
| TAPD URL | 使用 TAPD MCP 获取需求（如不可用则提示）                                    |
| Word 文档 | `python3 scripts/parse_word.py <file.docx>`                  |
| Markdown/文本 | `read_file` 直接读取                                             |
| 其他 URL | `curl -s <URL> \| sed 's/<[^>]*>//g'`                        |
| 口头描述 | 记录并确认理解                                                      |

---

## 步骤 2: 代码变更分析

### 2.0 项目检测
检测编程语言（`go.mod`/`pom.xml`/`package.json`/`.sql`/`.cs`/`.proto`/`.lua`/`.css`等）和框架类型（go-zero检查`.api`文件）。

**语言检测规则**：
- Go: `go.mod` 文件
- Java: `pom.xml` 或 `build.gradle` 文件
- Python: `requirements.txt` 或 `setup.py` 文件
- C++: `.cpp`/`.h` 文件
- SQL: `.sql` 文件
- C#: `.cs`/`.csproj` 文件
- ProtoBuf: `.proto` 文件
- Lua: `.lua` 文件
- CSS: `.css` 文件

### 2.1 Git Commit 分析
```bash
# 单个提交
python3 scripts/analyze_git_diff.py --range <commit>~1..<commit>

# 多个提交
python3 scripts/analyze_git_diff.py --range <start>..<end>

# 指定文件
python3 scripts/analyze_git_diff.py --range HEAD~3..HEAD --files file1.go file2.go

# 输出JSON
python3 scripts/analyze_git_diff.py --range HEAD~1..HEAD --output changes.json
```

### 2.2 审查本次版本
触发词："审查本次版本"、"review当前版本"

```bash
git describe --tags --abbrev=0  # 获取上个tag
python3 scripts/analyze_git_diff.py --range <tag>..HEAD
```

### 2.3 审查指定文件
使用 `read_file`/`view_code_item` 读取，应用 `exclude_paths` 过滤。

---

## 步骤 3: Lint 检查

```bash
python3 scripts/lint_check.py -l <go|java|cpp|python> [--repo <path>]
```

**执行逻辑**：
1. 优先检测 `Makefile` 中的 `lint:` 目标 → 执行 `make lint`
2. 否则按语言选择工具：Go(tencentlint)、Java(spotless/checkstyle)、C++(clang-tidy)、Python(ruff/flake8)
3. Go配置优先级：项目`.golangci.yml` > skill内置`assets/.golangci.yml`

**输出**：成功输出"Lint Success"，失败输出错误信息。

---

## 步骤 4: 架构分析

生成规格文档 `docs/spec-<功能名称>-<日期>.md`，包含：
- 架构概览、数据流、关键组件、业务逻辑、错误处理

如有需求文档，执行对比分析：✅已实现 / ⚠️部分实现 / ❌缺失

---

## 步骤 5: 深度审查

**必须加载** `references/coding-standards/<语言>/` 下的标准和安全文档：

| 语言 | 文档                            |
|-----|-------------------------------|
| Go | `standard.md`, `security.md`  |
| C++ | `standard.md`, `security.md`  |
| Java | `standard.md`, `security.md`  |
| Python | `standard.md`, `security.md`  |
| SQL | `standard.md` |
| C# | `standard.md` |
| ProtoBuf | `standard.md` |
| Lua | `standard.md` |
| CSS | `standard.md` |

**注意**：SQL、C#、ProtoBuf、Lua、CSS的标准文档需要通过 `sync_standards.py` 脚本从外部仓库同步。首次使用前请执行：
```bash
python3 scripts/sync_standards.py --all
```

### 审查类别

**并发安全 [concurrency]**：data-race、goroutine-leak、lock-usage、channel-operation

**安全性 [security]**：sql-injection、command-injection、path-traversal、authentication、authorization、weak-crypto、key-management、input-validation、info-disclosure

**潜在Bug [bug]**：nil-pointer、loop-closure、slice-modification、integer-overflow、boundary-check

**性能 [performance]**：memory-allocation、string-concatenation、slice-preallocation、struct-copy

**编程规范 [coding-standards]**：interface-design、error-wrapping、error-checking、context-propagation、naming-convention

**资源管理 [resource-management]**：resource-close、context-lifecycle、memory-leak

**框架专项 [framework]**：如果检测到当前项目是Go项目并且go.mod中引入了github.com/zeromicro/go-zero，则被判断为go-zero项目，参考`references/coding-standards/go/go-zero-framework.md`对 API定义、logic层、配置文件审查

### 严重程度
| 级别 | 说明 | 示例 |
|------|------|-----|
| 🛑严重 | 功能缺陷、安全漏洞 | SQL注入、数据竞争 |
| ⚠️重要 | 性能、质量问题 | 资源未关闭、错误处理不当 |
| 💡建议 | 代码风格、最佳实践 | 命名规范、预分配slice |

---

## 步骤 6: 生成报告

使用 `assets/report-template.md` 模板生成报告，保存至 `docs/code-review-<功能名称>-<日期>.md`。

报告结构：基本信息 → 需求符合性 → 代码质量分析 → 深度审查结果 → 专项评估 → 总结评分

---

## 步骤 7: Todo 清单

```markdown
### 🛑 Critical (必须修复)
- [ ] [Security] 修复SQL注入 (file.go:123)

### ⚠️ Major (建议修复)
- [ ] [Performance] 预分配slice (processor.go:78)

### 💡 Minor (可选优化)
- [ ] [Style] 命名规范 (utils.go:234)
```

---

## 审查优先级

**高优先级**（必审）：安全漏洞、并发问题、业务逻辑、关键bug

**中优先级**（时间允许）：性能优化、代码组织、测试覆盖、错误处理

**低优先级**（顺带提及）：命名规范、样式问题、文档完整性

**跳过**：lint已覆盖的基础问题（未使用变量、格式问题、vet警告）

---

## 脚本用法

| 脚本 | 用途 | 命令 |
|------|------|------|
| `analyze_git_diff.py` | Git变更分析 | `--range <range> [--files ...] [--output file.json]` |
| `lint_check.py` | Lint检查 | `-l <go\|java\|cpp\|python> [--repo <path>]` |
| `parse_word.py` | Word解析 | `<file.docx>` |
| `sync_standards.py` | 同步外部标准 | `--all [--force]` 或 `--languages <lang1> <lang2>` |

---

## 资源文件

- `assets/report-template.md` - 报告模板
- `assets/.codereview.example` - 配置示例
- `assets/.golangci.yml` - Go lint默认配置
- `references/coding-standards/` - 各语言审查标准
  - 内置标准：Go、C++、Java、Python
  - 外部标准：SQL、C#、ProtoBuf、Lua、CSS（需通过 `sync_standards.py` 同步）

---

## 依赖

**必需**：Python 3.8+、Git

**按需**：tencentlint/golangci-lint(Go)、ruff/flake8(Python)、clang-tidy(C++)、Maven/Gradle(Java)

---

## 限制

- 仅静态分析，无法执行代码或运行测试
- 安全/性能分析基于代码模式，非运行时行为
- 框架专项审查依赖对应参考文档

---

## 外部标准同步

部分语言的编码标准存储在外部Git仓库中，需要通过 `sync_standards.py` 脚本同步到本地。

### 支持的外部标准

| 语言 | 仓库地址 |
|------|---------|
| SQL | https://git.woa.com/standards/sql.git |
| C# | https://git.woa.com/standards/csharp.git |
| ProtoBuf | https://git.woa.com/standards/protobuf.git |
| Lua | https://git.woa.com/standards/Lua.git |
| CSS | https://git.woa.com/standards/css.git |

### 使用方法

```bash
# 首次使用：同步所有外部标准
python3 scripts/sync_standards.py --all

# 同步指定语言
python3 scripts/sync_standards.py --languages sql csharp

# 强制更新已存在的标准
python3 scripts/sync_standards.py --all --force

# 列出所有可用标准
python3 scripts/sync_standards.py --list
```

### 工作原理

1. 脚本会将外部仓库克隆到 `.temp_repos/` 目录（已在 `.gitignore` 中忽略）
2. 提取仓库中的 `README.md` 文件
3. 复制到 `references/coding-standards/<语言>/standard.md`
4. 后续更新时会执行 `git pull` 获取最新版本

### 注意事项

- 首次审查SQL/C#/ProtoBuf/Lua/CSS代码前，必须先执行同步
- 建议定期执行 `--force` 更新以获取最新标准
- 如果无法访问外部仓库，审查将跳过这些语言的标准检查
