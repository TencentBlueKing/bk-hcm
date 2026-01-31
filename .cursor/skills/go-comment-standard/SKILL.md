---
name: go-comment-standard
description: Go 代码注释规范完整指南。当编写函数、结构体、字段注释，审查注释风格，或确保注释遵循英文优先和名称开头约定时使用。
---

# Go 代码注释规范

## 概述

本 Skill 提供 hcm 项目的完整 Go 代码注释规范，包含详细的规则说明、代码示例和最佳实践。

SKILL.md 本身并没有提供具体的规范内容，必须根据场景和下方的规则索引，进一步读取对应的规则文件。

## 适用场景

- 编写新代码时参考注释格式
- Code Review 时检查注释规范性
- 不确定注释风格时查阅

## 核心原则

1. **英文为主**：代码注释必须使用英文
2. **名称开头**：函数、结构体、常量等的注释以其名称开头
3. **位置一致**：字段注释放在字段上方，不放行尾
4. **简洁明了**：注释说明"是什么"或"做什么"，避免冗余

## 规则索引

| 规则文件 | 内容 |
|---------|------|
| [function-comment.md](rules/function-comment.md) | 函数和方法注释规范 |
| [struct-comment.md](rules/struct-comment.md) | 结构体和字段注释规范 |
| [interface-comment.md](rules/interface-comment.md) | 接口注释规范 |
| [package-comment.md](rules/package-comment.md) | 包注释规范 |
| [constant-variable-comment.md](rules/constant-variable-comment.md) | 常量和变量注释规范 |
| [special-comment.md](rules/special-comment.md) | TODO/FIXME/Deprecated 等特殊注释 |
| [language-policy.md](rules/language-policy.md) | 中英文使用策略 |

## FAQ

### Q: 什么时候可以使用中文注释？

A: 仅在以下场景允许：
- 业务数据映射（如中文名称、中文错误提示）
- 复杂业务逻辑的补充说明
- 第三方系统字段说明

### Q: 非导出函数需要注释吗？

A: 建议添加，但简单明确的私有函数可省略。复杂的私有函数必须有注释。

### Q: 接口方法需要注释吗？

A: 接口定义处可省略，但实现处必须添加注释。

## 参考文件

项目中优秀的注释实践示例：

| 文件 | 特点 |
|-----|------|
| `pkg/tools/maps/maps.go` | 规范的函数注释，多行返回值说明 |
| `pkg/dal/table/cloud/cvm/cvm.go` | 完整的结构体和字段注释 |
| `pkg/dal/dao/dao.go` | 接口和非导出函数注释 |
| `pkg/criteria/constant/clb.go` | 常量组注释规范 |
| `pkg/tools/times/time.go` | 详细的包注释 |
