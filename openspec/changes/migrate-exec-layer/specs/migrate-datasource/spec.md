## ADDED Requirements

### Requirement: 从配置构造库连接

系统 SHALL 复用 dataservice 的配置文件。迁移 Job 把与 dataservice Deployment 相同的 ConfigMap 挂成一个文件，经 `--config-file` 把该路径交给 `cc.LoadSettings`。加载前 MUST 调用 `cc.InitService(cc.DataServiceName)`，否则文件不会反序列化成 `DataServiceSetting`。`cc.DataService()` 只读取已经载入内存的结果，MUST NOT 再去读文件，也 MUST NOT 访问 data-service 进程。系统 MUST NOT import `cmd/data-service`。从载入结果取 `Database` 与 `OBSDatabase`，各调用一次 `dao.NewDaoSet`。MUST NOT 为 migrate 新增独立的配置段。

#### Scenario: 两个库各建一份连接

- **WHEN** 配置文件中同时提供 `database` 与 `obsDatabase`，且进程已用 `data-service` 这个名字加载该文件
- **THEN** 系统建立两份 `dao.Set`，分别指向主库与 OBS 库

#### Scenario: 未登记服务名则读不到库配置

- **WHEN** 未调用 `cc.InitService(cc.DataServiceName)` 就加载配置文件
- **THEN** 命令失败，MUST NOT 拿到 `Database` 或 `OBSDatabase`

#### Scenario: 不依赖 data-service 镜像

- **WHEN** 检查迁移镜像的内容与启动命令
- **THEN** 其中没有 data-service 二进制，配置只来自 `--config-file` 指向的文件

#### Scenario: 配置缺失必要字段即失败

- **WHEN** 配置中 `Database` 缺少必要字段
- **THEN** 命令以退出码 2 失败并指出缺失项

### Requirement: 裸 Do 执行

迁移函数拿到的 ORM SHALL 是未挂 `ModifySQLOpts` 的裸 `Do()`。MUST NOT 启用租户 SQL 改写，否则语句会被重写到其他表上。

#### Scenario: 迁移执行不经租户改写

- **WHEN** 某迁移通过 `util` 执行一条 DDL
- **THEN** 实际下发的 SQL 与拼装结果一致，表名未被改写

### Requirement: OBS 库缺省跳过

`OBSDatabase` 未配置时系统 SHALL 整段跳过 OBS 库，打印一条显眼的 Warn 日志说明跳过原因，并以退出码 0 结束。MUST NOT 静默跳过，也 MUST NOT 因此失败。

#### Scenario: 外部版无 OBS 配置

- **WHEN** 配置中未提供 `OBSDatabase`，执行 `up`
- **THEN** 主库正常执行，OBS 段被跳过并打印 Warn 日志，命令退出码为 0

#### Scenario: 显式选择 obs 但无配置

- **WHEN** 未提供 `OBSDatabase`，执行 `up --database obs`
- **THEN** 命令打印 Warn 日志说明未配置 OBS 库并以退出码 0 结束

### Requirement: 执行顺序与失败隔离

多个库之间无顺序依赖，系统 SHALL 固定先主库后 OBS，使日志与排障可预期。主库失败时 MUST 直接退出且不再连接 OBS。OBS 失败时主库已写入的 `success` 记录 MUST 保留，MUST NOT 回滚，也 MUST NOT 做跨库补偿。任一库失败时进程 MUST 以非 0 退出。

#### Scenario: 主库失败不波及 OBS

- **WHEN** 主库某条迁移执行失败
- **THEN** 命令立即以非 0 退出，OBS 库不被连接

#### Scenario: OBS 失败不回滚主库

- **WHEN** 主库全部成功而 OBS 某条迁移失败
- **THEN** 主库的 `success` 记录保持不变，命令以非 0 退出

#### Scenario: 按库名子集只作用于指定库

- **WHEN** 执行 `up --database main`
- **THEN** 只处理主库，OBS 库不被连接

#### Scenario: 三个库里只跑两个

- **WHEN** 配置启用了三个库，执行 `up --database main,obs`
- **THEN** 只连接并处理被点名的两个库

### Requirement: 数据库就绪

部署顺序 SHALL 由 Job 的 initContainer 保证数据库端口可达。程序自身 SHALL 只做一次连接探测，失败即以明确错误退出，MUST NOT 在进程内重试轮询等待数据库就绪。

#### Scenario: 数据库不可达时快速失败

- **WHEN** 数据库不可达，执行 `up`
- **THEN** 命令在一次探测后失败，输出可定位的连接错误，不进入等待循环

### Requirement: 扩展到更多库

新增一个被迁移的库 SHALL 只需要增加一个配置对象、一个 `Registry` 实例和一棵迁移目录。执行器 SHALL 按「有连接才跑」循环，MUST NOT 需要改动判定逻辑。

#### Scenario: 新增库不改判定逻辑

- **WHEN** 增加第三个库的配置、注册表实例与目录
- **THEN** 该库进入执行循环，逐条判定规则与现有两个库一致
