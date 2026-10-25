## ADDED Requirements

### Requirement: 池注册表按插件启停管理池
`PluginRegistry` SHALL 为本副本每个已启用的 tf_provider 插件管理一个池。服务启动且插件注册完成后，注册表 MUST 为已启用的插件建池并预热；运行期按 `registryCheckInterval` 周期读取插件状态：读到已启用且无池则建池并预热；读到停用（`disabled`）则对该池调用 `Close(ctx)` 后移除。读取状态失败 MUST 只记日志并跳过本轮，MUST NOT 误关池。注册期读取 Schema 用的临时进程 MUST NOT 进入池。池的创建 MUST 按插件名唯一，同一插件 MUST NOT 同时存在两个池。通过注册表 `GetPool(pluginName)` 查不到池时调用方 MUST 得到明确的"插件未启用"错误（`ErrPluginNotEnabled`），MUST NOT 隐式创建池。

#### Scenario: 启动时建池
- **WHEN** 本副本注册完成，注册表中有 2 个已启用的 tf 插件
- **THEN** 建立 2 个池并各自预热

#### Scenario: 运行时启用
- **WHEN** 某插件在运行期被启用，下一轮检查读到该状态
- **THEN** 建池并预热

#### Scenario: 运行时停用
- **WHEN** 某插件被停用
- **THEN** 对应池停止接收新请求，排队请求收到"池已关闭"，进程被关闭，池被移出注册表

#### Scenario: 读取状态失败
- **WHEN** data-service 请求失败
- **THEN** 本轮跳过，已有池保持不变

#### Scenario: 查询未启用插件
- **WHEN** 调用方对未建池的插件名获取池
- **THEN** 返回"插件未启用"错误，不新建池

### Requirement: hc-service 接入与退出清理
hc-service SHALL 在插件注册完成并标记就绪之后建立池注册表，将其提供给执行代理使用；服务退出时 MUST 先停止池注册表巡检，再对所有池调用 `Close(ctx)` 并等待池内任务结束。接入 MUST 由配置开关控制，默认关闭；开关关闭时 MUST NOT 创建任何池或预热进程，行为与接入前一致。进程退出信号不会转发给子进程，因此服务退出路径 MUST 逐个关闭池内进程。接入 MUST NOT 修改 P2 注册 / 版本检查的既有逻辑，只新增调用。

#### Scenario: 开关关闭
- **WHEN** 池接入开关为关
- **THEN** 不建池、不拉起任何插件进程

#### Scenario: 正常退出
- **WHEN** hc-service 收到优雅退出通知
- **THEN** 注册表停止巡检并关闭所有池，池内进程在退出前被结束

#### Scenario: 注册失败
- **WHEN** 插件注册失败导致进程退出
- **THEN** 不创建池注册表

### Requirement: 池配置项与 helm 同步
`pkg/cc` 的 `IaCPluginConfig` SHALL 增加池相关配置项与接入开关，并提供默认值与校验：`maxProcesses`、`maxLeasesPerProcess`（默认 1）、`minIdleProcesses`（默认 1）、`startTimeout`（默认 30 秒）、`configureTimeout`（默认 30 秒）、`closeTimeout`（默认 30 秒）、`maxRetries`（默认 2）、`maxQueueWait`、`maxQueueLength`、`idleReclaimTimeout`（默认 10 分钟）、`maxConfigAge`（默认 30 分钟）、`poolCheckInterval`（默认 30 秒）、`registryCheckInterval`（默认 5 秒）、`leaseHoldWarn`、接入开关（默认关闭）。校验规则 MUST 与 `PoolConfig` 的校验一致：拒绝 `maxProcesses < 1`、`maxLeasesPerProcess < 1`、`minIdleProcesses > maxProcesses`、任何时长为 0。`hc_service.yaml` 与 helm 的 hcservice configmap / values MUST 同步，注释说明各项含义与取值约束：`maxConfigAge` 须短于云凭证有效期；`minIdleProcesses` 是空闲进程数下限，回收只处理超出下限的部分。

#### Scenario: 缺省值
- **WHEN** 配置文件未设置池相关项
- **THEN** 使用默认值，校验通过，接入开关为关

#### Scenario: 非法组合
- **WHEN** `minIdleProcesses` 大于 `maxProcesses`
- **THEN** 配置校验失败，服务不启动

#### Scenario: helm 渲染
- **WHEN** 使用默认 values 渲染 hcservice 的 configmap
- **THEN** 渲染结果包含池配置项且接入开关为关闭
