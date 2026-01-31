---
name: go-dal-pattern
description: 数据访问层（DAO/DAL）模式规范。当编写 DAO 层代码、设计数据访问接口、处理数据库操作或实现 ES 查询时使用。
---

# 数据访问层（DAO/DAL）模式规范

## 快速参考

| 场景 | 规则文件 |
|-----|---------|
| 目录和文件组织 | [rules/directory-structure.md](rules/directory-structure.md) |
| 接口设计 | [rules/interface-design.md](rules/interface-design.md) |
| 实现模式 | [rules/implementation-pattern.md](rules/implementation-pattern.md) |
| 事务处理 | [rules/transaction-handling.md](rules/transaction-handling.md) |
| ES DAO | [rules/es-dao.md](rules/es-dao.md) |

## 核心原则

1. **接口驱动**：所有 DAO 通过 `Interface` 暴露能力
2. **依赖注入**：通过构造函数注入 ORM、ID 生成器等依赖
3. **职责单一**：每个 DAO 只负责单一数据表/索引的操作
4. **组合复用**：通过 share 包实现跨表查询逻辑复用

## 目录结构概览

```
pkg/dal/
├── dao/                    # DAO 实现层
│   ├── dao.go             # DAO Set 入口
│   ├── audit/             # 审计表 入口
│   ├── orm/               # MySQL ORM
│   ├── types/             # 类型定义
│   └── <entity>/          # 具体实体 DAO
└── table/               # 表结构定义
```

## 接口设计要点

```go
// 接口命名为 Interface
type Interface interface {
    List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[T], error)
    Create(kt *kit.Kit, tx *sqlx.Tx, model *table.Entity) (string, error)
    BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []table.Entity) ([]string, error)
    // ...
}
```

**方法命名规范**：
- 单条查询：`Get*`
- 列表查询：`List*`
- 批量操作：`Batch*`
- 事务内方法：`*Txn`

## 实现模式要点

```go
// 简单实体
type Dao struct {
    Orm   orm.Interface
    IDGen idgenerator.IDGenInterface
    Audit audit.Interface
}

// 复杂实体（带审计/缓存）
type DiskDao struct {
    Orm   orm.Interface
    IDGen idgenerator.IDGenInterface
    Audit audit.Interface
}
```

## 常见问题

### Q: 何时使用事务？
A: 涉及多表写入或需要原子性保证的操作。通过 `*sqlx.Tx` 参数传递，或使用 `AutoTxn` 自动管理。

### Q: 结构体命名用 `dao` 还是 `xxxDao`？
A: 简单 CRUD 实体用 `dao`，复杂实体（带审计、缓存）用描述性名称如 `systemConfigDao`。

### Q: ES 和 MySQL DAO 有何区别？
A: ES DAO 使用 `ormes.Interface`，通常不需要 ID 生成器（使用文档唯一键）。

## 关键文件

| 文件 | 说明 |
|-----|------|
| `pkg/dal/dao/dao.go` | DAO Set 入口 |
| `pkg/dal/dao/user/user_collection.go` | 标准 CRUD 示例 |
| `pkg/dal/dao/cloud/load-balancer/load_balancer.go` | 复杂 DAO 示例 |
