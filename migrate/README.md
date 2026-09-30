# hcm-migrate

HCM MySQL migration 独立二进制，落点按 HCM Migration 技术方案。

本地记忆（图文理解 + 评论）：`.cursor/docs/iwiki/4039435213-hcm-migration/understanding.md`

实现时按下面拆包。迁移文件只 import `register` 和 `util`，不要 import `engine`。

进度：`util/`（MetaOrm + 幂等 DDL/DML）已实现。`register/`、`engine/`、`scripts/`、`migrations/` 等有代码再入库，空目录不占位。

```text
migrate/
├── migrate.go                        入口 + 子命令：init / up / status / list
├── Makefile                          本地构建 + check-imports
├── scripts/
│   └── new-migration.sh              建模板 + 写 imports.go
├── register/                         Registry + Main / Obs；Version 解析与排序
├── util/                             MetaOrm + 幂等 DDL/DML（已实现）
├── engine/                           executor / recordstore / schema
└── migrations/
    ├── imports.go                    空白 import，把版本包拉进编译
    ├── main/                         主库 → register.Main
    │   ├── v1.9.3/                   三位通用
    │   └── v1.9.3.x/                 特有（内部 .N / tenant -label.N）
    └── obs/                          OBS 库 → register.Obs（外部没有）
```
