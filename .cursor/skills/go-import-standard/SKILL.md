---
name: go-import-standard
description: Go 导入规范完整指南。当编写新 Go 文件、审查导入顺序、配置 IDE 自动格式化、或查阅包别名约定时使用。
---

# Go 导入规范

本 Skill 提供 hcm 项目中 Go 代码导入的完整规范指南。

SKILL.md 本身并没有提供具体的规范内容，必须根据场景和下方的规则索引，进一步读取对应的规则文件。

## 适用场景

- 编写新的 Go 文件时需要了解导入规范
- 审查代码导入是否符合项目规范
- 了解项目内部包的别名约定
- 配置 IDE 自动格式化导入

## 规则索引

| 文件 | 内容 |
|-----|------|
| [import-grouping.md](rules/import-grouping.md) | 导入分组规范：三组顺序、空行分隔 |
| [alias-convention.md](rules/alias-convention.md) | 别名命名规范：何时使用、命名模式、常用对照表 |
| [special-imports.md](rules/special-imports.md) | 特殊导入规范：下划线导入、禁止点导入 |
| [common-patterns.md](rules/common-patterns.md) | 常见导入模式：Service/DAO/工具包/测试文件 |

## 快速检查清单

- [ ] 导入分为三组：标准库 → 内部包 → 第三方包
- [ ] 组与组之间有空行分隔
- [ ] 仅在必要时使用别名（冲突、连字符、语义不清）
- [ ] 别名使用项目约定的命名（如 `dsset`, `cstypes`）
- [ ] 下划线导入有注释说明用途
- [ ] 未使用点导入
- [ ] 使用 `goimports` 自动格式化

## FAQ

### Q: 什么时候需要给内部包起别名？

当满足以下任一条件时：
1. 目录名含连字符（如 `client-set`）
2. 包名与其他导入冲突
3. 默认包名语义不清晰（如通用的 `types`）

### Q: `cstypes` 和 `ctypes` 应该用哪个？

两者都指向 `hcm/pkg/client/types`。`cstypes` 使用更广泛，建议新代码统一使用 `cstypes`。

### Q: 如何自动格式化导入？

运行 `goimports -w .`，或在 IDE 中配置保存时自动运行 `goimports`。
