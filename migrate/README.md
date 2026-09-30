# hcm-migrate

HCM MySQL migration 独立二进制，落点按 HCM Migration 技术方案。

本地记忆（图文理解 + 评论）：`.cursor/docs/iwiki/4039435213-hcm-migration/understanding.md`

实现时按下面拆包。迁移文件只 import `register`、`util`，以及注册 `PENDING` 时的 `pkg/criteria/constant`，不要 import `engine`。

进度：`util/`（MetaOrm + 幂等 DDL/DML）已实现。`register/` 已改为可读 ID + `Migrator` / `Pkg`（版本语法仍在本包；常量和格式校验分别在 `pkg/criteria/constant`、`pkg/migrate`）。`engine/`：`schema.go` 的 `InitTables`、`recordstore.go`、`auditstore.go`（`NewAudit().Begin` / `End`）已写，执行器与 CLI 接线未做。`scripts/`、`migrations/` 等有代码再入库，空目录不占位。

```text
migrate/
├── migrate.go                        入口 + 子命令：init / up / status / list
├── Makefile                          本地构建 + check-imports
├── scripts/
│   └── new-migration.sh              建模板 + 写 imports.go
├── register/                         Registry + Main / Obs；Version 解析与排序
├── util/                             MetaOrm + 幂等 DDL/DML（已实现）
├── engine/                           schema（InitTables）/ recordstore / auditstore
└── migrations/
    ├── imports.go                    空白 import，把版本包拉进编译
    ├── main/                         主库 → register.Main
    │   ├── v1.9.3/                   三位通用
    │   └── v1.9.3.x/                 特有（内部 .N / tenant -label.N）
    └── obs/                          OBS 库 → register.Obs（外部没有）
```
