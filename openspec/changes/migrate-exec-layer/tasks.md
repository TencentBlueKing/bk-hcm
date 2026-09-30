## 1. 版本解析与比较器

- [x] 1.1 在 `migrate/register/version.go` 定义 `Suffix` / `Version` 结构与占位符常量，值为 `PENDING`
- [x] 1.2 实现 `Parse`：一条正则拆三种写法，额外拒绝前导零，非法输入返回错误
- [x] 1.3 实现 `IsPending`：只认常量 `PENDING`；`Parse("PENDING")` 返回错误；小写 `pending` 既不是合法版本也不是占位符
- [x] 1.4 实现 `Compare(a, b Version)`：P1/P2/P3 数值 → 第四段（nil 最小，先 Label 后 Seq）。时间戳与 Migration ID 是迁移的字段不是版本的字段，由第 2 组的 `All` 在 `Compare` 返回 0 时接着比（见任务 2.7）
- [x] 1.5 在 `version_test.go` 写 `Parse` 表驱动测试，覆盖三种合法写法与全部非法用例（缺 `v`、`v1.9`、`v1.9.3.1.2`、`v1.9.3.tenant.1`、`v1.9.3-Tenant.1`、`v1.9.3-tenant`、`v01.9.3`）
- [x] 1.6 写 `Compare` 表驱动测试：`v1.9.3` 与 `v1.9.3.0` 相等、`v1.9.3 < v1.9.3-tenant.0`、`v1.9.9 < v1.9.10`、三位先于 `.1`、数字第四段先于带标签、同标签比 Seq 数值、跨标签比字典序，并断言反对称
- [x] 1.7 写排序测试：对打乱的完整版本列表排序，断言结果等于设计文档中的顺序示例，并断言多次排序结果一致

## 2. 注册表

- [x] 2.1 在 `migrate/register/register.go` 定义 `Migration` 结构与 `UpFunc`（参数为 `orm.Interface` 而非占位类型）
- [x] 2.2 实现 `Registry` 与 `Regist`：校验 UUID（小写标准写法）、版本（合法或常量 `PENDING`）、14 位时间戳、`up` 非空
- [x] 2.3 同库相同 Migration ID 允许注册，不在启动期拒绝
- [x] 2.4 校验失败时在 `init()` 阶段 panic 终止进程，panic 文本带上全部五个注册字段以定位到具体迁移文件
- [x] 2.5 暴露 `Main` 与 `Obs` 两个实例；实现返回排序后副本的 `All`，`PENDING` 排在最后
- [x] 2.6 写 `register_test.go`：合法注册、各类非法参数拒绝、同库相同 ID 两条都注册成功、不同库相同 ID 注册成功、`All` 返回副本
- [x] 2.7 写测试断言 `All` 的顺序与第 1 组比较器一致（版本 → 时间戳 → ID）

## 3. 记录表与记录读写

- [x] 3.1 `util.CreateTableIfNotExists` 增加返回值 `created bool`（表已存在时为 false 且不执行语句）。`migrate/schema/schema.go` 放不含 `IF NOT EXISTS` 的 `CREATE TABLE`；`init` 只调用这个函数，不自己探测 `information_schema`。仅 `created && mode==adopt` 时插入基线记录
- [x] 3.2 在 `migrate/schema/recordstore.go` 定义记录结构与状态常量（`running` / `success` / `failed`）
- [x] 3.3 实现按库读取全部记录，并按 `migration_id` 建索引供判定使用
- [x] 3.4 实现 `running` 写入：`INSERT ... ON DUPLICATE KEY UPDATE` 回到 `running` 并清空 `message`
- [x] 3.5 实现 `success` 写入：更新状态并把 `version` 刷成此刻注册的版本
- [x] 3.6 实现 `failed` 写入：更新状态与 `message`，按字节安全截断到 1024 且不切断多字节字符，完整错误进日志
- [x] 3.7 实现库当前版本计算：解析全部 `success` 记录的版本后用比较器取最大；无记录返回空；解析失败返回可映射为退出码 3 的错误
- [x] 3.8 实现版本漂移 WARN：`success` 记录的版本与注册版本不一致时告警但不失败
- [x] 3.9 写 `recordstore` 单元测试：用 `migrate/util` 里已有的 `orm.Interface` 假实现思路覆盖库当前版本计算（含 `v1.9.9` 与 `v1.9.10`、补跑低版本不降低库当前、空记录、解析失败）与 `message` 截断

## 4. 数据库连接管理

- [x] 4.1 启动时 `cc.InitService(cc.DataServiceName)`，再 `cc.LoadSettings` 按 `--config-file` 读文件。从载入结果取 `Database` 与 `OBSDatabase`，各调一次 `dao.NewDaoSet`。不 import `cmd/data-service`
- [x] 4.2 取 `GetOrm()` 的裸 `Do()` 交给迁移，确认未挂 `ModifySQLOpts`
- [x] 4.3 实现 `OBSDatabase` 为 `nil` 时整段跳过并打 Warn 日志，退出码保持 0
- [x] 4.4 实现 `--database` 库名列表（可重复、可逗号分隔），不传表示已启用的全部库，未知名字映射为退出码 2
- [x] 4.5 实现单次连接探测，失败返回可定位的错误，不做进程内重试轮询
- [x] 4.6 写测试覆盖库选择与 OBS 缺省跳过的分支判断

## 5. 执行器

- [x] 5.1 在 `migrate/engine/executor.go` 实现执行前扫描：未开 `--allow-pending` 时 PENDING → 退出码 5；两个及以上不同非空标签 → 退出码 5；与漏执行 / ID 复用同一轮汇总（不再「连库前退出码 3」）。引擎层已由 `adjust-migration-specs` 落地，本项随 CLI 接线关闭
- [x] 5.1.1 实现记录表缺失检查：`up` 在表不存在时返回映射为退出码 3 的错误并提示先跑 `init`，不建表、不执行任何迁移
- [x] 5.2 实现版本上限解析与过滤（`Compare(m.Version, ceiling) <= 0`），上限非法映射为退出码 2
- [x] 5.3 实现逐条判定循环：已 success 跳过、默认模式漏执行失败、补跑模式一律执行，`running` 与 `failed` 同等视为无成功记录
- [x] 5.4 实现执行单条迁移：写 `running` → 调 `up()` → 写 `success` 或 `failed`，记录写入在迁移事务之外，且不包裹外层事务
- [x] 5.5 实现失败即中断：该库后续迁移不执行，错误向上返回
- [x] 5.6 实现按库循环：固定先主库后 OBS，主库失败不再连 OBS，OBS 失败不回滚主库
- [x] 5.7 实现 `--plan`：打印 `EXECUTE` / `SKIP-SUCCESS` / `ABOVE-MAX-VERSION` / `MISSING`，不建表、不写记录、不调 `up()`
- [x] 5.8 写判定循环单元测试：表驱动覆盖判定表全部分支（已成功、超上限、默认模式高于/低于库当前、补跑模式、running、failed）——`TestBuildPlan`
- [x] 5.9 写测试覆盖断点续跑：模拟第三条失败后重跑，断言前两条被跳过——`TestExecute`（失败即停）+ `TestLocalMySQLExecutor`（failing up / rerun resumes）+ `TestBuildPlan`（success 跳过）
- [x] 5.10 写测试覆盖 `--plan` 在默认模式与补跑模式下对同一批迁移给出不同判定——`TestBuildPlan`（missed vs catch-up）
- [x] 5.11 写测试覆盖相同 ID 的两条迁移：靠前的执行成功后，后一条被跳过且不调用 `up()`——`TestBuildPlan`（execute claims id second same suffix skip success）

## 6. 命令行入口

- [x] 6.1 在 `migrate/migrate.go` 装配入口（stderr 日志 + `cli.Run`）；子命令与全局参数在包 `migrate/cli`（`--config-file` / `--database` / `--allow-pending`）
- [x] 6.1.1 实现 `init`：`--mode` 必填（`empty` / `adopt`）、`--baseline 库名=版本` 可重复、`--plan`；表已存在即对该库 no-op；`adopt` 缺 `--baseline` 退出码 2；被选中但无基线条目的库按 `empty` 建空表
- [x] 6.2 实现 `up`：`--to` / `--catch-up` / `--plan`，执行前打印本次命令收到的模式
- [x] 6.3 实现 `status`：按库打印当前版本、各状态计数、未执行清单，并标出 `version ≤ 库当前` 的未执行项；不接收也不推断模式；只读不建表
- [x] 6.4 实现 `list`：按执行顺序打印，不连库，版本为 `PENDING` 的项照常打印且退出码 0
- [x] 6.5 核对 `init --mode=adopt` 只写记录、不调用任何 `up()`
- [x] 6.6 统一退出码映射：0 / 1 / 2 / 3 / 4 / 5 / 6，确认漏执行、注册表问题、ID 复用与执行失败可区分（`pkg/migrate.ExitCode`）
- [x] 6.7 分离输出：判定结果与汇总走 stdout，过程日志与错误详情走 `pkg/logs`
- [x] 6.8 写命令层测试：参数校验、非法 `--database`、`--to` 非法、`init` 缺 `--mode`、`adopt` 缺 `--baseline`、`up` 缺记录表、退出码映射——`migrate/cli/cli_test.go`、`migrate/cli/mysql_test.go`、`pkg/migrate/errors_test.go`

## 7. 本地 MySQL 集成测试

- [ ] 7.1 约定集成测试开关：读取 DSN 环境变量，未设置则整体 `t.Skip`，避免在 CI 与生产环境误连（现实现为固定连 `127.0.0.1:3306`，多数用例连不上即 fail；仅 `TestLocalMySQLExecutor` 对不可用 Skip）
- [x] 7.2 实现测试库隔离：每次用随机名建临时库，测试结束无条件 `DROP DATABASE`，禁止连已有业务库——`openIsolatedMySQL`（engine / util）
- [x] 7.3 写建表集成测试：`init --mode=empty` 建空表、重复 `init` 无副作用、`up` 在缺表时失败且不建表——`TestLocalMySQLEngine` / `TestLocalMySQLExecutor` / `TestLocalMySQLCLI`（`status` 在未初始化库上退出码 3 且不建表）
- [x] 7.4 写记录读写集成测试：`running` → `success`、`running` → `failed`、重跑回到 `running`、唯一键保证一条迁移只有一行——`TestLocalMySQLEngine`（mark running failed and success）
- [ ] 7.5 写执行器集成测试：注册若干假迁移，覆盖空库全跑、断点续跑、默认模式漏执行失败、补跑模式补齐、版本上限（`TestLocalMySQLExecutor` 已覆盖空库全跑与断点续跑；漏执行 / 补跑 / 版本上限仍在 `plan_test` 单元层）
- [x] 7.6 写 `init` 集成测试：`adopt` 垫起库版本且 `up()` 未被调用、表已存在时 no-op——`TestLocalMySQLEngine`、`TestLocalMySQLCLI`（init adopt / second init / `init` 后 `up`；混合主库 adopt、OBS empty 尚无独立用例）
- [x] 7.7 按同一隔离方式（临时库 + 无条件清理）补 `migrate/util` 的本地 MySQL 集成测试，验证幂等助手在真实 MySQL 上的执行正确性与重复执行无副作用——`TestLocalMySQLIdempotent`（同样无 DSN 环境变量开关）

## 8. 构建与部署接入

- [ ] 8.1 写 `migrate/Makefile`：`build` / `test` 目标
- [ ] 8.2 加 `check-imports` 门禁：校验 `migrations/` 下每个版本目录都已被 `imports.go` 空白 import
- [ ] 8.3 加 `check-migration-deps` 门禁：校验 `migrations/` 子树未 import `migrate/engine`
- [ ] 8.4 在 `cmd/Makefile` 与根 `Makefile` 挂接独立迁移镜像的构建目标
- [ ] 8.5 建 `migrations/` 目录骨架与初始 `imports.go`（`main/` 与 `obs/` 各留空目录结构）
- [ ] 8.6 编写 Helm hook Job：命令为 `init && up`、三个 hook 注解、`backoffLimit: 0`、`parallelism: 1`、`restartPolicy: Never`、`ttlSecondsAfterFinished`、initContainer 等 DB 就绪、复用 dataservice ConfigMap 与 chart 的 ServiceAccount。迁移镜像不打进业务 Pod
- [ ] 8.7 在 chart values 就近注释说明补跑开关是按包一次性开关，并注明 `parallelism` 必须保持为 1
- [ ] 8.8 在 values 里暴露 `--mode` 与 `--baseline`，并写接入说明：空库填 `empty`、存量库填 `adopt` 并如何选定各库基线、误填 `adopt` 到空库的后果、`--plan` 如何预演

## 9. 收尾校验

- [ ] 9.1 通读 `migrate/` 全部新增代码，核对命名、导入分组、注释、错误处理与日志规范（含 rid）
- [ ] 9.2 核对迁移文件只依赖 `register` 与 `util`，`engine` 未被 `migrations/` 引用
- [ ] 9.3 跑通空库全量执行与增量执行两条路径，核对记录表内容与 `status` 输出一致
- [ ] 9.4 逐条核对 `design.md` 的偏离清单是否都已按决策落地
