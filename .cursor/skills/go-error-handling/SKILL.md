---
name: go-error-handling
description: Go 错误处理模式与最佳实践。当编写错误处理代码、审查错误传播、使用 errf 包创建业务错误、或确定错误场景的日志级别时使用。
---

# Go 错误处理规范

## 概述

本 Skill 提供 hcm 项目的完整 Go 代码错误处理规范，包含详细的规则说明、代码示例和最佳实践。

SKILL.md 本身并没有提供具体的规范内容，必须根据场景和下方的规则索引，进一步读取对应的规则文件。

## 适用场景

- 编写新代码时参考错误处理模式
- Code Review 时检查错误处理规范性
- 不确定如何处理错误时查阅

## 核心原则

1. **统一错误类型**：业务错误使用 `errf` 包创建，携带错误码
2. **日志必记录**：错误返回前必须记录日志，包含 rid
3. **层次分明**：Handler 层添加错误码，DAO 层直接传递
4. **合理传播**：关键错误返回，非关键错误可 warn 后继续

## 规则索引

| 规则文件 | 内容 |
|---------|------|
| [error-creation.md](rules/error-creation.md) | 错误创建方式规范 |
| [error-checking.md](rules/error-checking.md) | 错误检查模式规范 |
| [error-logging.md](rules/error-logging.md) | 错误日志记录规范 |
| [error-propagation.md](rules/error-propagation.md) | 错误传播模式规范 |
| [error-codes.md](rules/error-codes.md) | 错误码定义和使用 |
| [database-errors.md](rules/database-errors.md) | 数据库错误处理规范 |
| [api-errors.md](rules/api-errors.md) | HTTP/API 错误处理规范 |

## FAQ

### Q: 什么时候用 `errf.New` vs `errors.New`？

A: 
- `errf.New/Newf`：API 层返回给客户端的错误，需要携带错误码
- `errors.New`：内部函数、工具函数中不需要错误码的场景

### Q: 什么时候用 `logs.Errorf` vs `logs.Warnf`？

A:
- `logs.Errorf`：错误后 return，流程终止
- `logs.Warnf`：错误后 continue/fallthrough，流程继续

### Q: DAO 层需要添加错误码吗？

A: 通常直接返回原始错误，由 Service 层添加错误码。但以下场景例外：
- 参数校验失败可使用 `InvalidParameter`
- 数据库执行失败可使用 `DBExecCmdFailed`

### Q: 如何判断特定错误类型？

A:
- 判断业务错误码：`errf.HasCode(err, errf.RecordNotFound)`
- 判断标准错误类型：`errors.Is(err, orm.ErrRecordNotFound)`

## 参考文件

项目中优秀的错误处理实践示例：

| 文件 | 特点 |
|-----|------|
| `pkg/criteria/errf/error.go` | 核心错误类型定义 |
| `pkg/criteria/errf/code.go` | 错误码定义 |
| `pkg/rest/handler.go` | Handler 包装器 |
| `cmd/cloud-server/service/cvm/query.go` | 标准 Handler 错误处理 |
| `cmd/data-service/service/cloud/cvm/create.go` | 事务错误处理 |
