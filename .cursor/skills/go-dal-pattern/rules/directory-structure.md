# 目录结构与文件命名规范

## 1. 整体目录结构

```
pkg/dal/
├── dao/                          # DAO 实现层
│   ├── dao.go                    # DAO Set 定义和初始化
│   ├── audit/                   # 辅助工具函数
│   │   ├── audit.go               # 审计表 DAO
│   ├── orm/                      # MySQL ORM 封装
│   │   └── orm.go
│   │   └── sqlx.go
│   ├── types/                    # DAO 层类型定义
│   │   ├── types.go             # 通用类型
│   │   ├── page.go              # 分页相关
│   │   └── */*.go              # 基础选项类型
│   ├── id-generator/             # ID 生成器
│   │   └── id_generator.go
│   └── <entity-name>/            # 具体实体 DAO
│       └── <entity_name>.go
├── table/                        # 表结构定义
│   └── <table_name>/<table_name>.go
├── audit/                        # 审计日志（50+ 文件）
└── objectstore/                  # 对象存储
```

## 2. 命名规范

| 类型 | 规则 | 示例 |
|-----|------|------|
| 目录 | kebab-case | `account-set`, `cloud/cvm`, `task` |
| 文件 | snake_case | `main_account.go`, `cvm.go`, `management.go` |

## 3. 文件拆分规范

对于复杂实体，按功能拆分文件：

| 后缀 | 用途 | 示例 |
|-----|------|------|
| `<entity>.go` | 主文件、接口定义 | `obs_bill.go` |
| `_impl.go` | 基础实现 | `obs_bill_impl.go` |
| `_brief.go` | 简要/轻量查询 | `obs_bill_brief.go` |
| `_daily.go` | 按日粒度数据 | `obs_bill_daily.go` |
| `_summary.go` | 汇总数据 | `obs_bill_summary.go` |
| `_archive.go` | 归档数据 | `obs_bill_archive.go` |
| `_alloc.go` | 分摊逻辑 | `obs_bill_alloc.go` |
| `_analysis.go` | 分析查询 | `obsbill_analysis.go` |

## 4. 复杂实体示例

`pkg/dal/dao/cloud/load-balancer/` 目录结构：

```
load-balancer/
├── load_balancer.go                    # 负载均衡表的接口定义 && 基础实现
├── load_balancer_listener.go           # 监听器表的接口定义 && 基础实现
├── load_balancer_target_group.go       # 目标组表的接口定义 && 基础实现
└── target_group_listener_rule_rel.go   # 目标组跟监听器对应关系表的接口定义 && 基础实现
```

## 5. 何时拆分文件

✅ **应该拆分**：
- 接口方法超过 20 个
- 单文件超过 500 行
- 存在明显独立的功能模块

❌ **不需要拆分**：
- 简单 CRUD 实体
- 方法数量少（< 10 个）
- 逻辑紧密耦合

## 6. 关键文件说明

| 文件 | 职责 |
|-----|------|
| `dao.go` | DAO Set 定义，初始化所有 DAO 实例 |
| `types/types.go` | ListOption、ListResult 等通用类型 |
| `types/page.go` | 分页相关类型和逻辑 |
| `id-generator/id_generator.go` | ID 生成接口和实现 |
