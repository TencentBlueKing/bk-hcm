## ADDED Requirements

### Requirement: 启动时序与探针
hc-service SHALL 按以下顺序启动：目录检查 → 启动 HTTP 服务但暂不注册 etcd → 逐个插件注册 → 检查注册表中的 tf 插件在目录里是否都在 → 注册 etcd。注册完成前 `/healthz` MUST 返回 HTTP 503，`/alivez` MUST 保持 200。目录检查失败 MUST 使进程退出且不启动服务。任一插件处理失败 MUST 释放锁并以非 0 退出码退出。全部通过后 `/healthz` 返回 200，且服务才对调用方可见。

#### Scenario: 注册期间未就绪
- **WHEN** 插件注册进行中
- **THEN** `/healthz` 返回 503，`/alivez` 返回 200，服务未注册 etcd

#### Scenario: 全部通过后就绪
- **WHEN** 所有插件处理通过
- **THEN** 服务注册 etcd，`/healthz` 返回 200

#### Scenario: 目录检查失败
- **WHEN** 目录检查失败
- **THEN** 进程以非 0 退出，HTTP 服务未启动

#### Scenario: 无插件目录
- **WHEN** 插件根目录不存在或为空
- **THEN** 无插件需注册，按"注册表缺失检查"继续（已启用插件缺失则退出）

### Requirement: 持插件名锁处理单个插件
对每个插件，副本 SHALL 获取以插件名为键的 etcd 锁（带租约、不续约），拿不到则在 `lockWaitTimeout` 内等待（监听 key 删除，不轮询），超时即注册失败；拿到后读取该插件名下全部记录处理，处理完释放锁。令 M 为已完成记录（`enabled`、`disabled`）中的最高版本，V 为目录中二进制版本。插件名在注册表中唯一，该插件名下的记录 MUST 都是 `protocol=tf_provider`；存在其它协议的记录时 MUST 直接失败退出、MUST NOT 修改任何记录（否则会把 tf 的 sha256 写进别的协议的 extension）。对记录的修改 MUST 为条件更新（带读取时的 `state`）。注册、升级、启停共用同一把插件名锁，持锁期间不存在正常的并发写，因此影响 0 行时 MUST 直接失败退出、MUST NOT 重新读取后继续修改：此时本副本已不持有锁（租约失效）或有人绕过锁改记录，继续写会造成无锁写入。

#### Scenario: 并发启动串行处理
- **WHEN** 副本 A、B 同时启动并发现同一新版本
- **THEN** A 先拿锁完成注册并置 `enabled`；B 拿锁后读到 `enabled`，只校验 sha256，不拉起进程，也不同步 Schema

#### Scenario: 条件更新冲突
- **WHEN** 条件更新影响 0 行
- **THEN** 直接失败退出，不修改记录、不重试；由编排重启后在新的锁下重新走一遍

### Requirement: 注册与升级判定
持锁后 SHALL 按以下规则处理：V 低于 M → 退出，不修改任何记录；没有已完成记录或 V 不低于 M 时，看 V 的记录：不存在 → 插入 `init` 记录并走同步流程；为 `init` 或 `failed` → 覆盖改为 `init` 并走同步流程（不设超时判断）；为 `enabled` 或 `disabled` → 只校验 sha256，不拉起进程。目录版本高于 M 即升级：插入新版本记录，旧版本记录 MUST 保持原状。注册 MUST NOT 修改已有记录的 `state`。

#### Scenario: 首次注册
- **WHEN** 注册表中没有该插件
- **THEN** 插入 V 的 `init` 记录并走同步流程

#### Scenario: 升级
- **WHEN** 目录版本 `1.83.34` 高于已完成最高版本 `1.83.30`
- **THEN** 插入 `1.83.34` 的 `init` 记录，`1.83.30` 记录保持原状

#### Scenario: 版本回退
- **WHEN** 目录版本 `1.83.30` 低于注册表已完成最高版本 `1.83.34`
- **THEN** 进程退出，注册表不变

#### Scenario: 上次中断留下的 init
- **WHEN** V 的记录为 `init`（上次注册中途崩溃）
- **THEN** 重新走同步流程，不设超时判断

#### Scenario: 上次失败留下的 failed
- **WHEN** V 的记录为 `failed`
- **THEN** 改为 `init` 后重新走同步流程

#### Scenario: 已停用的插件重启
- **WHEN** V 的记录为 `disabled`
- **THEN** 只校验 sha256，状态保持 `disabled`

### Requirement: 拉起进程并同步 Schema
同步流程 SHALL 由注册流程驱动：通过对接接口 `ProcessStarter`（对接方实现）按「插件名、版本、可执行文件路径」取得 `ProcessManager`，调用 `Start` 取得 `Executor`，把执行器的 `GetProviderSchema` 适配为 `SchemaGetter` 后传给 `SchemaSyncer.Sync`（对接方）。注册流程 MUST 在同步结束后无论成败都调用 `ProcessManager.Close`（`Close` 在进程管理上，不在执行器上），该进程 MUST NOT 进入进程池。进程的握手、协议版本协商、gRPC 通信与关闭细节不属于本流程。拉起或握手失败时 MUST NOT 调用 `Sync`，直接把记录置 `failed` 并写入 `message`。同步成功后 MUST 将记录从 `init` 条件更新为 `enabled`（上一版本为 `disabled` 则为 `disabled`），并写入本架构 sha256。同版本记录已是 `enabled` 或 `disabled` 时 MUST NOT 拉起进程、MUST NOT 调用 `Sync`。

#### Scenario: 同步成功
- **WHEN** 进程拉起成功且 `Sync` 返回 nil
- **THEN** 进程被关闭，记录改为 `enabled`，`extension` 写入本架构 sha256

#### Scenario: 继承停用状态
- **WHEN** 上一版本为 `disabled`，新版本同步成功
- **THEN** 新版本记录为 `disabled`

#### Scenario: 握手失败
- **WHEN** 拉起进程或握手失败
- **THEN** 不调用 `Sync`，记录改为 `failed` 并写 `message`，进程以非 0 退出

#### Scenario: 同步失败
- **WHEN** `Sync` 返回错误
- **THEN** 进程仍被关闭，记录改为 `failed` 并写 `message`（含失败原因），进程以非 0 退出

#### Scenario: 已完成记录不重复同步
- **WHEN** 同版本记录为 `enabled`
- **THEN** 不拉起进程，不调用 `Sync`

### Requirement: 注册表有记录而目录里没有
启动完成注册后，副本 SHALL 检查注册表中 `protocol=tf_provider` 的插件（只读，不加锁），以该插件名下版本最高记录的 `state` 为准：`enabled` → 进程以非 0 退出，滚动发布停住、旧副本继续服务；`disabled`、`init`、`failed` → 跳过。

#### Scenario: 已启用插件缺失
- **WHEN** 注册表中 `tencentcloud` 为 `enabled`，但目录里没有该插件
- **THEN** 进程以非 0 退出，不注册 etcd

#### Scenario: 已停用插件缺失
- **WHEN** 注册表中该插件为 `disabled`，目录里没有
- **THEN** 跳过，不影响启动

#### Scenario: 注册失败留下的记录
- **WHEN** 该插件最高版本记录为 `failed`，目录里没有
- **THEN** 跳过

### Requirement: 失败一律退出，不做进程内重试
系统 SHALL 区分以下几类失败，但处理方式统一为"退出、不注册 etcd"，MUST NOT 在进程内对失败做退避重试：

- 目录检查失败 → 退出、不启动服务；
- 确定性注册失败（握手失败、取 Schema 失败、同步返回错误）→ 记录置 `failed` 并写 `message`（写不进去时保持 `init`）、释放锁、非 0 退出；
- 与已有记录冲突（版本低于已完成最高版本、sha256 不一致）→ 不修改任何记录、释放锁、非 0 退出；
- 依赖不可用（data-service 报错、etcd 不可用）→ 直接返回错误、非 0 退出。

由编排层重启进程来完成重试：helm 已保证 dataservice 就绪后才启动 hc-service，etcd 不可用在启动准备阶段即失败，二者都不需要在进程内等待。等锁超时同样按注册失败处理。

#### Scenario: 与已有记录冲突不改记录
- **WHEN** sha256 不一致导致退出
- **THEN** 所有 `iac_plugin` 记录保持不变

#### Scenario: 依赖不可用时退出而非等待
- **WHEN** 调用 data-service 返回错误
- **THEN** 注册直接失败、进程非 0 退出，`/healthz` 保持 503，由编排重启重试

#### Scenario: 等锁超时
- **WHEN** 锁一直被占用，等待超过 `lockWaitTimeout`
- **THEN** 按注册失败处理，非 0 退出

### Requirement: 注册审计
注册与升级 SHALL 写审计，操作人为 `system`：新增版本记录写 `create`（`res_type=iac_plugin`，`bk_biz_id=-1`，`detail.data` 为新记录）；状态变化（注册失败置 `failed` 等）写 `update`，`detail.changed` 为 `{"state":"<新状态>"}`。同版本只校验 sha256 的启动 MUST NOT 写审计。

#### Scenario: 升级写审计
- **WHEN** 新版本记录插入成功
- **THEN** 产生一条 `create` 审计，操作人 `system`

#### Scenario: 无变化不写审计
- **WHEN** 启动时记录已是 `enabled`，仅校验 sha256
- **THEN** 不产生审计

### Requirement: 启动扫描与运行期检查的分工
启动时扫描用于**发现并注册**新版本：变更插件目录后 MUST 重启 hc-service，新版本才会被该副本扫描到并走升级流程。运行期由独立的版本检查负责**发现本副本落后**（见 `iac-plugin-tf-version-watch`）：本副本二进制版本低于注册表已完成最高版本并连续 3 个周期成立时退出。注册流程 MUST NOT 在运行期重新扫描目录。

#### Scenario: 重启后感知新版本
- **WHEN** 插件目录新增 `1.83.34` 后重启 hc-service
- **THEN** 启动扫描发现新版本并走升级流程

#### Scenario: 热更新目录不重启
- **WHEN** 运行期间替换插件目录内容但未重启
- **THEN** 该副本的注册流程不感知；若其它副本已注册更高版本，本副本在 3 个检查周期内因版本落后退出

#### Scenario: 重启时本地版本落后
- **WHEN** 注册表已完成最高版本为 `1.83.34`，本副本目录最高版本为 `1.83.30` 并重启
- **THEN** 启动判定为版本回退，进程以非 0 退出
