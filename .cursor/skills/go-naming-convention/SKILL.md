---
name: go-naming-convention
description: Go 代码命名规范完整指南。当需要了解项目命名约定、审查命名是否规范、或编写新代码需要参考命名模式时使用。
---

# Go 命名规范

本 Skill 提供 hcm 项目的完整命名规范，包含详细示例和代码引用。

SKILL.md 本身并没有提供具体的规范内容，必须根据场景和下方的规则索引，进一步读取对应的规则文件。

## 快速参考

| 场景 | 规则文件 |
|-----|---------|
| 包和文件命名 | [rules/package-file-naming.md](rules/package-file-naming.md) |
| 结构体命名 | [rules/struct-naming.md](rules/struct-naming.md) |
| 接口命名 | [rules/interface-naming.md](rules/interface-naming.md) |
| 函数和方法命名 | [rules/function-naming.md](rules/function-naming.md) |
| 常量和变量命名 | [rules/constant-variable-naming.md](rules/constant-variable-naming.md) |
| 缩写约定 | [rules/abbreviation.md](rules/abbreviation.md) |

## 核心原则

1. **一致性优先** - 遵循项目现有模式，而非个人偏好
2. **语义清晰** - 命名应准确反映用途，避免过于简单或模糊
3. **Go 惯例** - 遵循 Go 社区标准（如导出用 PascalCase，私有用 camelCase）

## 使用指南

- **编写新代码**：先阅读对应场景的规则文件
- **代码审查**：对照规则检查命名是否符合项目约定
- **重构**：参考规则统一命名风格

## 常见问题

### Q: 请求结构体用 `*Req` 还是 `*Request`？
A: 统一使用 `*Req` 后缀。项目中有 741 个 `*Req` 结构体，0 个 `*Request`。

### Q: DAO 接口为什么都叫 `Interface`？
A: 因为接口名称已由包名限定（如 `obsbill.Interface`），使用通用名称可保持一致性。

### Q: 工具包可以用复数形式吗？
A: 可以。工具类包参照 Go 标准库惯例，可使用复数形式（如 `times`, `strings`, `maths`）。
