## Context

方案 5.3：启停接口只改注册表 `state`，不直接操作进程；各 hc-service 副本以注册表 `state` 为准自行调整本地进程池（请求入口读状态、后台巡检关闭进程池），这部分属运行态，不在本 change。方案还要求接口持插件名的锁，否则启停发生在新版本"读取上一版本状态"和"写入状态"之间时，新版本会拿到过期的继承状态。

仓库现状：cloud-server 新增管理接口的参照写法是 `permission-policy-library`（`InitService`、`h.Add`、`authorizer.AuthorizeWithPerm`、`audit.ResUpdateAudit`）；平台级资源的 IAM 映射在 `auth-server/.../gen_id.go` 中直接映射到现有 `sys.ActionID`，无需新增 IAM 动作；cloud-server 的 `cc.Service` 同样带 `Etcd` 配置，可构造 etcd 客户端。

Schema 存在性由对接方提供，本 change 用一个包级函数变量调用，不声明接口；交付前默认放行。

## Goals / Non-Goals

**Goals:**
- 启停语义与方案表格逐格一致，且可用表驱动单测验证。
- 与注册 / 升级 / 修改串行，不出现状态继承读到过期值。
- 幂等：重复请求不重复写库、不重复写审计。

**Non-Goals:**
- 进程池响应、请求入口接入、后台巡检。
- http 插件的注册 / 升级 / 修改（P4）。
- 新增 IAM 动作（沿用 `global_configuration`；方案允许后续新增 `iac_plugin_find` / `iac_plugin_manage`）。

## Decisions

### D1 状态规则抽成纯函数
`DecideSwitch(current *IaCPlugin, target State, schemaExists bool) (Action, error)`，输出"无需处理 / 批量改状态 / 拒绝(原因)"。方案表格的 8 个格子一一对应单测；handler 只负责取数与执行。

### D2 取"当前版本"与批量更新的两步
流程：加锁（`plugin_name`） → 列出该插件名下所有记录 → 取最高版本 → `DecideSwitch` → 调 P1 的 `batch_update_state(plugin_name, from_states, to_state)`。`from_states` 目标 disabled 时为 `[enabled, failed]`，目标 enabled 时为 `[disabled]`。持锁保证读与写之间无新版本插入。

### D3 拿不到锁不等待
管理接口是同步 HTTP，等待会拖到上游网关超时。使用 `TryLock`，拿不到返回明确的"进行中"错误，调用方稍后重试。锁租约取较短值（秒级到分钟级）并可配置，因为启停只有两次 data-service 调用。

### D4 启用前检查 Schema 存在性：函数变量 + 交付前默认放行
仅在 `disabled → enabled` 时调用 `isIaCPluginSchemaExists(kt, pluginName, version)`，它当前是 `svc` 的方法，对接方交付后改为调用其查询接口。

交付前返回 `(true, nil)` 并打一条 warn：此时 Schema 尚未接入存储，返回 `false` 或错误会让所有"启用"操作永久失败，联调无法进行；放行的风险是可能启用一个 Schema 缺失的插件，运行期由插件自身报错，属于可接受降级。检查出错（非"不存在"）按系统错误返回，不当作"不存在"，避免误判。

### D5 审计在变化成功后写
批量改状态成功且受影响行数 > 0 才写审计，`res_id` 取当前版本记录；变更前快照取自加锁后读到的记录。

### D6 可用性判定函数放 `pkg/iacplugin`
纯函数，无 I/O，运行态阶段直接 import；错误类型沿用 `errf`，`failed` 的错误信息拼接记录 `message`。

### D7 权限映射沿用 `global_configuration`
`meta.ResourceType` 新增 `iac_plugin`；`gen_id.go` 增加映射函数返回 `sys.GlobalConfiguration`（以仓库实际常量名为准，开发时对照 `CloudVendorConfig` 样例确认）。若后续需要独立 IAM 动作，只改映射函数，不影响接口。

## Risks / Trade-offs

- [`TryLock` 拿不到时用户需手动重试] → 错误信息明确；启停耗时极短，冲突概率低。
- [`from_states` 目标 disabled 包含 `failed`，会把异常版本一并停用] → 与方案一致（异常版本无副本使用，停用无副作用）。
- [当前版本为 `init` 时一律拒绝，管理员无法在注册卡住时停用] → 方案规则如此；`init` 重入由注册 / 升级流程收敛。
- [Schema 查询未交付] → 默认放行并告警，联调可继续；交付后只替换 `isIaCPluginSchemaExists` 一个方法。

## Open Questions

- IAM 是否需要独立动作（`iac_plugin_find` / `iac_plugin_manage`）：方案写"或新增"，先沿用 `global_configuration`，上线前由权限负责人确认。
- 锁租约的具体默认值：开发时按 data-service 调用超时取值。
