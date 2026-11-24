# plugin-process-lifecycle Specification

## Purpose
定义 `tfruntime` 基于 `go-plugin` 托管单个 Terraform Provider 插件进程的拉起握手、协议版本 5 强校验、显式环境变量清单、在途请求排空与 `StopProvider`/`Kill` 关闭顺序，以及 `hc-service` 优雅退出时的子进程回收规范。

## Requirements

### Requirement: 每次拉起都新建 go-plugin 客户端
`tfruntime` MUST 通过新的 `plugin.Client` 拉起插件进程，并用该客户端完成握手和本机 unix gRPC 连接。同一个 `plugin.Client` MUST NOT 被再次 `Start` 来重拉二进制。同一个 `ProcessManager` 的 `Start` MUST 只能成功一次。再次调用 MUST 返回错误，且 MUST NOT 创建新的 `plugin.Client`。需要另一个进程时 MUST 新建 `ProcessManager`。

握手 MUST 使用以下固定值：核心协议版本 `1`；magic cookie 键 `TF_PLUGIN_MAGIC_COOKIE`；magic cookie 值 `d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2`；`VersionedPlugins` 只登记 `5`；dispense 名称为 `provider`；`AllowedProtocols` 只含 gRPC；传输为本机 unix 套接字；`GRPCDialOptions` 的收发消息上限至少 256MB；TLS 使用 `go-plugin` AutoMTLS。

#### Scenario: 成功拉起协议 5 的插件
- **WHEN** 调用方传入可执行的插件二进制、参数和工作目录，且协商版本为 5
- **THEN** 系统返回进程句柄，并且连接使用插件生成的本机 unix 套接字地址

#### Scenario: 再次拉起不复用原来的 Client
- **WHEN** 调用方需要再拉起一个插件进程
- **THEN** 系统用新的 `ProcessManager` 创建新的 `plugin.Client`，而不是对已有 `Client` 再次 `Start`

#### Scenario: 同一个管理器再次 Start
- **WHEN** 对已经 `Start` 成功的 `ProcessManager` 再次调用 `Start`
- **THEN** 系统返回错误，不创建新的 `plugin.Client`，原进程仍由该管理器关闭

### Requirement: 非协议主版本 5 的进程不得交给调用方
协商完成后，系统 MUST 读取 `NegotiatedVersion()`。结果不是 5 时，系统 MUST 立即 `Kill` 该进程并返回错误，且 MUST NOT 返回进程句柄。

#### Scenario: 插件协商到其他主版本
- **WHEN** 插件完成握手但 `NegotiatedVersion()` 不是 5
- **THEN** 系统结束该进程并返回错误，调用方拿不到句柄

### Requirement: 插件环境变量使用显式清单
系统 MUST NOT 把 hc-service 的 `os.Environ()` 原样复制给插件。插件环境 MUST 只包含执行二进制所需的最小环境，以及 `PLUGIN_PROTOCOL_VERSIONS=5` 和 `TF_LOG`。`TF_LOG` 默认 MUST 为 `ERROR`。系统 MUST 接受调用方为单次 `Start` 指定的 `TRACE`、`DEBUG`、`INFO`、`WARN`、`ERROR` 之一，并只作用于该进程。其他取值 MUST 拒绝启动。云账号密钥 MUST NOT 出现在插件环境变量中。`TF_PLUGIN_MAGIC_COOKIE` MUST 由 `go-plugin` 按握手配置写入。`PLUGIN_MIN_PORT` 与 `PLUGIN_MAX_PORT` MUST 取自 hc-service 配置 `tfPlugin.minPort` / `tfPlugin.maxPort` 这一组区间；未填写时默认 `40000`–`40999`。该区间 MUST NOT 包含 hc-service 监听端口。

#### Scenario: 默认日志级别
- **WHEN** 调用方启动插件且未指定日志级别
- **THEN** 插件进程的 `TF_LOG` 为 `ERROR`，且环境变量中没有云账号密钥

#### Scenario: 非法日志级别
- **WHEN** 调用方传入的日志级别不是 `TRACE`、`DEBUG`、`INFO`、`WARN`、`ERROR` 之一
- **THEN** 系统拒绝启动，不留下插件进程

#### Scenario: 单独提高一次启动的日志级别
- **WHEN** 调用方为这一次 `Start` 指定 `DEBUG`
- **THEN** 仅该进程的 `TF_LOG` 为 `DEBUG`，不改变 hc-service 自身日志级别

### Requirement: Close 先排空调用再停止进程
每个业务 RPC 方法 MUST 在方法内部做前置和后置门闩。前置检查发现关闭已开始时，方法 MUST 失败且不向插件发出请求。后置在方法返回时离开门闩。系统 MUST NOT 要求调用方经由单独的 `Call` 才能访问这些方法。`StopProvider` 由 `Close` 在门闩关闭后调用，MUST NOT 再进入这道门闩。`Close` MUST 按顺序执行：拒绝新的方法进入，等待已经进入的方法返回，调用 `StopProvider`，然后调用 `plugin.Client.Kill()`。`StopProvider` 失败时系统 MUST 仍执行 `Kill`。同一句柄上的重复 `Close` MUST 只执行一次该顺序。本迭代的业务 RPC 尚未实现，方法内门闩在实现这些方法时补上。

#### Scenario: 关闭时存在在途调用
- **WHEN** 一个业务 RPC 方法尚未返回时调用方调用 `Close`
- **THEN** 系统不再接受新的方法进入，等待该在途方法返回后才调用 `StopProvider`，然后再 `Kill`

#### Scenario: StopProvider 失败
- **WHEN** `Close` 调用 `StopProvider` 返回错误
- **THEN** 系统记录该错误并仍然 `Kill` 插件进程

#### Scenario: 重复关闭
- **WHEN** 对同一句柄调用两次 `Close`
- **THEN** 停止顺序只执行一次

### Requirement: 优雅退出时回收已拉起的插件进程
系统 MUST 在某个管理器第一次 `Start` 成功时，为该管理器注册一个 `shutdown.AddNotifier`。hc-service 因 `SIGTERM` 或 `SignalShutdownGracefully` 退出时，该回调 MUST 对该管理器执行 `Close`，然后调用 `Done`。进程已经关闭时，回调 MUST 立即 `Done`。系统 MUST NOT 维护跨管理器的进程登记表，MUST NOT 另注册 `RegisterFirstShutdown`，也 MUST NOT 在本变更中为 `SIGINT` 增加信号处理。

#### Scenario: 服务收到 SIGTERM
- **WHEN** hc-service 进入现有优雅退出，且该管理器仍有插件进程
- **THEN** 系统关闭这个进程，并在回调结束时调用 notifier 的 `Done`

#### Scenario: 已经没有运行中的插件
- **WHEN** 退出通知到达时该管理器的进程已经关闭
- **THEN** 回调立即 `Done`，不占用整段 20 秒退出预算
