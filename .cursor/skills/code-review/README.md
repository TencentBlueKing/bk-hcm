# Code Review Skill

深度代码审查工具，超越传统 Lint 检查，提供全方位的代码质量分析和安全审查。

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [工作流程](#工作流程)
- [配置说明](#配置说明)
- [脚本工具](#脚本工具)
- [审查维度](#审查维度)
- [支持语言](#支持语言)
- [CI/CD 集成](#cicd-集成)
- [常见问题](#常见问题)

---

## 🎯 功能特性

### 核心能力

- **🔍 深度代码分析**：识别安全漏洞、并发问题、潜在 Bug、架构设计问题
- **📊 多维度审查**：并发安全、安全性、性能、编程规范、资源管理、框架专项
- **📝 专业报告生成**：基于模板生成结构化的 Code Review 报告
- **🔧 智能 Lint 集成**：自动检测并执行项目的 Lint 工具
- **📦 需求符合性分析**：对比需求文档，验证功能实现完整性
- **🎨 多语言支持**：Go、Java、Python、C++、SQL、C#、ProtoBuf、Lua、CSS 基于腾讯代码规范：https://git.woa.com/groups/standards/-/projects/list

### 审查优势

✅ **超越 Lint**：不仅检查代码格式，更关注逻辑、安全、性能  
✅ **上下文感知**：理解业务逻辑，提供针对性建议  
✅ **框架专项**：针对 go-zero 等框架的特定审查规则  
✅ **可配置**：灵活的配置系统，适应不同团队规范  
✅ **自动化友好**：支持 CI/CD 集成，可作为质量门禁

---

## 🚀 快速开始

### 基本使用

在 CodeBuddy 中激活 code-review skill 后，可以通过以下方式触发审查：

```bash
# 审查最近一次提交
"请审查最近一次提交，对应的需求在https://www.tapd.cn/XXX/markdown_wikis/show/#XXX"

# 审查指定提交范围
"CodeReview从 v1.0.0 到 HEAD 的代码变更，对应需求文档在./docs/requirement1.docx"

# 审查本次版本
"CodeReview 本次版本，需求和设计参考：./docs/requirement.md"

# 审查指定文件
"审查 internal/logic/user.go 文件，对应文档参考: http://www.xxx.com/docs/index.html"
```

### 首次使用准备

如果需要审查 SQL、C#、ProtoBuf、Lua、CSS 代码，需要先同步标准，请保证当前环境为腾讯内网，AI会执行：

```bash
python3 scripts/sync_standards.py --all
```

---

## 🔄 工作流程

Code Review Skill 遵循以下 7 步工作流程：

```
配置读取 → 需求收集 → 代码变更分析 → Lint检查 → 架构分析 → 深度审查 → 报告生成 → Todo清单
```

### 步骤详解

#### 0️⃣ 配置读取

自动读取项目根目录的 `.codereview` 配置文件（可选），支持：
- 排除特定文件/目录
- 忽略特定审查类别或规则
- 设置审查严格程度
- 自定义报告输出路径

#### 1️⃣ 需求收集

支持多种需求输入方式：
- **TAPD URL**：通过 TAPD MCP 工具获取需求详情
- **Word 文档**：使用 `parse_word.py` 解析 `.docx` 文件
- **Markdown/文本**：直接读取需求文档
- **URL**：通过 curl 获取在线文档
- **口头描述**：记录用户口头描述的需求

#### 2️⃣ 代码变更分析

**项目检测**：
- 自动识别编程语言（通过 `go.mod`、`pom.xml`、`package.json` 等）
- 检测框架类型（如 go-zero 通过 `.api` 文件识别）

**Git 变更分析**：
```bash
# 使用 analyze_git_diff.py 脚本
python3 scripts/analyze_git_diff.py --range HEAD~1..HEAD
python3 scripts/analyze_git_diff.py --range v1.0.0..HEAD --output changes.json
```

#### 3️⃣ Lint 检查

```bash
# 使用 lint_check.py 脚本
python3 scripts/lint_check.py -l go
python3 scripts/lint_check.py -l java --repo /path/to/project
```

**执行逻辑**：
1. 优先检测 `Makefile` 中的 `lint:` 目标
2. 否则按语言选择默认工具：
   - Go: `tencentlint` (golangci-lint)
   - Java: Maven/Gradle spotless/checkstyle
   - C++: clang-tidy
   - Python: ruff/flake8

#### 4️⃣ 架构分析

生成规格文档 `docs/spec-<功能名称>-<日期>.md`，包含：
- 架构概览
- 数据流分析
- 关键组件说明
- 业务逻辑梳理
- 错误处理机制

如有需求文档，执行需求符合性对比分析。

#### 5️⃣ 深度审查

基于编码标准文档进行多维度审查：

| 审查类别 | 检查项 |
|---------|--------|
| **并发安全** | data-race, goroutine-leak, lock-usage, channel-operation |
| **安全性** | sql-injection, command-injection, path-traversal, weak-crypto, authentication, authorization |
| **潜在Bug** | nil-pointer, loop-closure, slice-modification, integer-overflow, boundary-check |
| **性能** | memory-allocation, string-concatenation, slice-preallocation, struct-copy |
| **编程规范** | interface-design, error-wrapping, error-checking, context-propagation, naming-convention |
| **资源管理** | resource-close, context-lifecycle, memory-leak |
| **框架专项** | go-zero API设计、logic层、配置文件审查 |

**严重程度分级**：
- 🛑 **严重**：功能缺陷、安全漏洞（必须修复）
- ⚠️ **重要**：性能问题、质量问题（建议修复）
- 💡 **建议**：代码风格、最佳实践（可选优化）

#### 6️⃣ 报告生成

使用 `assets/report-template.md` 模板生成报告，保存至 `docs/code-review-<功能名称>-<日期>.md`。

报告包含：
- 基本信息和审查范围
- 需求符合性分析
- 代码质量与复杂度
- 深度审查结果（按严重程度分类）
- 专项评估（安全、性能）
- 总结评分

#### 7️⃣ Todo 清单

生成可操作的修复清单：

```markdown
### 🛑 Critical (必须修复)
- [ ] [Security] 修复SQL注入 (user.go:123)

### ⚠️ Major (建议修复)
- [ ] [Performance] 预分配slice (processor.go:78)

### 💡 Minor (可选优化)
- [ ] [Style] 命名规范 (utils.go:234)
```

---

## ⚙️ 配置说明

### 配置文件

在项目根目录创建 `.codereview` 文件（参考 `assets/.codereview.example`）：

```yaml
# 审查严格程度
severity: standard  # strict | standard | loose

# 排除路径（文件级跳过）
exclude_paths:
  - "vendor/"
  - "node_modules/"
  - "*.pb.go"
  - "*_test.go"
  - "internal/types/types.go"  # go-zero 生成的代码

# 忽略特定审查类别
ignore_categories:
  # - concurrency
  # - performance
  # - security

# 忽略特定审查规则
ignore_rules:
  # - data-race
  # - sql-injection
  # - naming-convention

# 文件级规则忽略
file_ignore_rules:
  - file: "internal/legacy/**/*.go"
    rules:
      - naming-convention
      - exported-comment

# 代码质量阈值
quality_thresholds:
  cyclomatic_complexity: 10
  function_lines: 80
  file_lines: 800
  nesting_depth: 4

# 报告输出配置
output: code-review/CR-${requirement}-${date}.md

# 自定义编程规范文档路径
coding_standards:
  go: /path/to/custom/go-standard.md
```

### 配置优先级

1. 项目根目录 `.codereview` 配置
2. Skill 内置默认配置
3. 命令行参数覆盖

---

## 🛠️ 脚本工具

### 1. analyze_git_diff.py

分析 Git 代码变更，支持多种输出格式。

```bash
# 基本用法
python3 scripts/analyze_git_diff.py --range HEAD~1..HEAD

# 分析指定范围
python3 scripts/analyze_git_diff.py --range v1.0.0..HEAD

# 分析指定文件
python3 scripts/analyze_git_diff.py --range HEAD~3..HEAD --files file1.go file2.go

# 输出为 JSON
python3 scripts/analyze_git_diff.py --range HEAD~1..HEAD --output changes.json

# 输出为 Markdown
python3 scripts/analyze_git_diff.py --range HEAD~1..HEAD --format markdown --output changes.md

# 详细模式
python3 scripts/analyze_git_diff.py --range HEAD~1..HEAD --verbose
```

**输出示例**：
```json
[
  {
    "file_path": "internal/logic/user.go",
    "language": "go",
    "added_lines": 45,
    "deleted_lines": 12,
    "total_changes": 57,
    "hunk_count": 3
  }
]
```

### 2. lint_check.py

执行 Lint 检查，支持多种语言。

```bash
# Go 项目
python3 scripts/lint_check.py -l go

# Java 项目
python3 scripts/lint_check.py -l java

# C++ 项目
python3 scripts/lint_check.py -l cpp

# Python 项目
python3 scripts/lint_check.py -l python

# 指定仓库路径
python3 scripts/lint_check.py -l go --repo /path/to/project
```

**执行逻辑**：
1. 检测 `Makefile` 中的 `lint:` 目标 → 执行 `make lint`
2. 否则使用语言默认工具
3. Go 配置优先级：项目 `.golangci.yml` > skill 内置配置

**输出**：
- 成功：`Lint Success`
- 失败：输出错误信息

### 3. sync_standards.py

同步外部编码标准仓库。

```bash
# 同步所有标准
python3 scripts/sync_standards.py --all

# 同步指定语言
python3 scripts/sync_standards.py --languages sql csharp

# 强制更新已存在的标准
python3 scripts/sync_standards.py --all --force

# 列出所有可用标准
python3 scripts/sync_standards.py --list
```

**支持的外部标准**：
- SQL: `https://git.woa.com/standards/sql.git`
- C#: `https://git.woa.com/standards/csharp.git`
- ProtoBuf: `https://git.woa.com/standards/protobuf.git`
- Lua: `https://git.woa.com/standards/Lua.git`
- CSS: `https://git.woa.com/standards/css.git`

### 4. parse_word.py

解析 Word 需求文档。

```bash
python3 scripts/parse_word.py requirement.docx
```

**功能**：
- 提取段落文本
- 提取表格内容
- 自动安装 `python-docx` 依赖（如未安装）

---

## 📐 审查维度

### 并发安全 [concurrency]

- **data-race**：数据竞争检测
- **goroutine-leak**：Goroutine 泄漏
- **lock-usage**：锁使用不当
- **channel-operation**：Channel 操作错误

### 安全性 [security]

- **sql-injection**：SQL 注入
- **command-injection**：命令注入
- **path-traversal**：路径遍历
- **authentication**：认证问题
- **authorization**：授权问题
- **weak-crypto**：弱加密算法
- **key-management**：密钥管理
- **input-validation**：输入验证
- **info-disclosure**：信息泄露

### 潜在 Bug [bug]

- **nil-pointer**：空指针引用
- **loop-closure**：循环闭包变量捕获
- **slice-modification**：切片并发修改
- **integer-overflow**：整数溢出
- **boundary-check**：边界检查

### 性能 [performance]

- **memory-allocation**：内存分配优化
- **string-concatenation**：字符串拼接
- **slice-preallocation**：切片预分配
- **struct-copy**：结构体复制

### 编程规范 [coding-standards]

- **interface-design**：接口设计
- **error-wrapping**：错误包装
- **error-checking**：错误检查
- **context-propagation**：Context 传播
- **naming-convention**：命名规范

### 资源管理 [resource-management]

- **resource-close**：资源关闭
- **context-lifecycle**：Context 生命周期
- **memory-leak**：内存泄漏

### 框架专项 [framework]

**go-zero 框架**（检测到 `github.com/zeromicro/go-zero` 依赖时）：
- API 定义规范
- Logic 层实现
- 配置文件审查
- JWT 认证配置
- Context 传递
- 错误处理
- 数据库操作
- 缓存使用
- RPC 调用

---

## 🌍 支持语言

### 内置标准

| 语言 | 标准文档 | 安全文档 | Lint 工具 |
|------|---------|---------|----------|
| **Go** | ✅ | ✅ | tencentlint/golangci-lint |
| **Java** | ✅ | ✅ | spotless/checkstyle |
| **Python** | ✅ | ✅ | ruff/flake8 |
| **C++** | ✅ | ✅ | clang-tidy |

### 其他标准（需同步）

| 语言 | 标准文档 | 同步命令 |
|------|---------|---------|
| **SQL** | ✅ | `sync_standards.py --languages sql` |
| **C#** | ✅ | `sync_standards.py --languages csharp` |
| **ProtoBuf** | ✅ | `sync_standards.py --languages protobuf` |
| **Lua** | ✅ | `sync_standards.py --languages lua` |
| **CSS** | ✅ | `sync_standards.py --languages css` |

---

## 🔗 CI/CD 集成

```
执行CodeBuddy Code命令:
codebuddy -y -p "使用code-review这个SKILL对本次提交进行代码审查，需求为{需求XXX的链接或文档路径，从git commit信息中提取即可}。"
```

### 退出码

- `0`：无严重问题
- `1`：存在严重问题
- `2`：执行错误

---

## ❓ 常见问题

### Q1: 如何跳过测试文件的审查？

在 `.codereview` 配置中添加：
```yaml
exclude_paths:
  - "*_test.go"
  - "test/"
  - "tests/"
```

### Q2: 如何忽略特定的审查规则？

```yaml
ignore_rules:
  - naming-convention
  - exported-comment
```

或针对特定文件：
```yaml
file_ignore_rules:
  - file: "internal/legacy/**/*.go"
    rules:
      - naming-convention
```

### Q3: Go 项目 Lint 失败怎么办？

1. 检查是否安装了 `tencentlint` 或 `golangci-lint`
2. 确认 `.golangci.yml` 配置是否正确
3. 使用 `--verbose` 查看详细错误信息

### Q4: 如何自定义编码标准？

在 `.codereview` 中指定自定义标准文档：
```yaml
coding_standards:
  go: /path/to/custom/go-standard.md
  java: /path/to/custom/java-standard.md
```

### Q5: 审查报告保存在哪里？

默认保存在 `docs/code-review-<功能名称>-<日期>.md`，可通过配置自定义：
```yaml
output: code-review/CR-${requirement}-${date}.md
```

### Q6: 如何审查 go-zero 项目？

Skill 会自动检测 go-zero 项目（通过 `go.mod` 和 `.api` 文件），并应用框架专项审查规则。

### Q7: 外部标准同步失败怎么办？

1. 检查网络连接和 Git 仓库访问权限
2. 使用 `--force` 强制重新同步
3. 手动下载标准文档并放置到对应目录

---

## 📚 资源文件

```
.codebuddy/skills/code-review/
├── SKILL.md                          # Skill 定义文档
├── README.md                         # 本文档
├── .gitignore                        # Git 忽略规则
├── assets/                           # 资源文件
│   ├── .codereview.example          # 配置示例
│   ├── .golangci.yml                # Go Lint 默认配置
│   └── report-template.md           # 报告模板
├── references/                       # 编码标准参考
│   └── coding-standards/
│       ├── go/                      # Go 标准（内置）
│       ├── java/                    # Java 标准（内置）
│       ├── python/                  # Python 标准（内置）
│       ├── cpp/                     # C++ 标准（内置）
│       ├── sql/                     # SQL 标准（需同步）
│       ├── csharp/                  # C# 标准（需同步）
│       ├── protobuf/                # ProtoBuf 标准（需同步）
│       ├── lua/                     # Lua 标准（需同步）
│       └── css/                     # CSS 标准（需同步）
└── scripts/                          # 工具脚本
    ├── analyze_git_diff.py          # Git 变更分析
    ├── lint_check.py                # Lint 检查
    ├── parse_word.py                # Word 文档解析
    └── sync_standards.py            # 外部标准同步
```

---

## 🤝 贡献

欢迎贡献新的审查规则、语言支持或改进建议！

---

## 📄 许可

本 Skill 遵循项目许可协议。

---

## 📞 支持

如有问题或建议，请联系 devinyzeng 或提交 Issue。
