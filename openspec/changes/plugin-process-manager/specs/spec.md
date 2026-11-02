# plugin-process-lifecycle

## ADDED Requirements

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

# tfplugin5-grpc-stub

## ADDED Requirements

### Requirement: 由协议 5.9 的 proto 生成并提交 gRPC 桩
仓库 MUST 在 `pkg/plugin/tfplugin5/` 包含声明为协议 5.9 的 `tfplugin5.proto`，以及由 `protoc-gen-go` 和 `protoc-gen-go-grpc` 生成的 Go 桩代码。生成命令 MUST 记在 `//go:generate` 中。生成的 Provider 客户端 MUST 包含 `StopProvider`。桩代码 MUST NOT 放在 `pkg/adaptor`。

#### Scenario: 桩代码覆盖 StopProvider
- **WHEN** 查看已提交的 `tfplugin5` 生成客户端
- **THEN** Provider 客户端提供 `StopProvider`，且 proto 文件标明协议 5.9

### Requirement: GRPCPlugin 只支持客户端
版本 5 的 `plugin.GRPCPlugin` MUST 在 `GRPCClient` 中接收 `*grpc.ClientConn` 并返回生成的 Provider 客户端。`GRPCServer` MUST 返回错误。hc-service MUST NOT 作为该协议的 gRPC 服务端。

#### Scenario: 建立客户端
- **WHEN** `go-plugin` 以已连接的 `*grpc.ClientConn` 调用 `GRPCClient`
- **THEN** 返回的对象是生成的 Provider 客户端，可以发起 `StopProvider`

#### Scenario: 拒绝作为服务端
- **WHEN** 调用该 `GRPCPlugin` 的 `GRPCServer`
- **THEN** 方法返回错误，不注册 Provider 服务

### Requirement: 本变更不封装执行器 RPC
除 `Close` 路径上的 `StopProvider` 外，`tfruntime` MUST NOT 调用或封装 `GetProviderSchema`、`PrepareProviderConfig`、`ConfigureProvider`、`ValidateResourceTypeConfig`、`ValidateDataSourceConfig`、`UpgradeResourceState`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`、`ImportResourceState`、`ReadDataSource`。系统 MUST NOT 在本变更中实现 `DynamicValue` 编解码。

#### Scenario: 进程关闭只使用 StopProvider
- **WHEN** 进程句柄执行 `Close`
- **THEN** 对插件业务 RPC 的调用只有 `StopProvider`，不调用 Plan、Apply、GetSchema 或其他执行器方法

# plugin-runtime-log

## ADDED Requirements

### Requirement: 握手后的标准输出写入独立的插件日志文件
系统 MUST 在 `Start` 之前装上 `SyncStdout` 和 `SyncStderr`，并持续读取，避免插件因管道写满而阻塞。握手之后插件写到 stdout 或 stderr 的内容 MUST 写入独立的插件日志文件。该文件 MUST 放在 hc-service 日志目录下的 `plugin` 子目录。文件名 MUST 为 `{程序名}.{插件名}.{主机名}.{用户名}.log.INFO.{YYYYMMDD-HHMMSS}.{pid}`，后缀与 glog 的 `logName` 相同。单文件大小、重启续写和保留个数 MUST 沿用 glog 现有的 `logKeeper`。每一行 MUST 包含 `plugin`、`plugin_version`、`pid`，并标明来自 `stdout` 还是 `stderr`。pid 尚未可知时，该行的 pid 为空，后续行写入已知 pid。版本为空时，行内的 `plugin_version` MUST 为 `unknown`。

#### Scenario: 插件写出 stdout 和 stderr
- **WHEN** 握手完成后插件向 stdout 和 stderr 写入日志
- **THEN** 两路内容都出现在该插件自己的日志文件中，且每行带有插件名、版本、pid 和对应的流名称

#### Scenario: 版本未提供
- **WHEN** 调用方没有提供插件版本
- **THEN** 日志行中的 `plugin_version` 为 `unknown`

### Requirement: 框架日志和握手前 stderr 进入 hc-service 日志
`go-plugin` 的 `Logger` 输出以及握手完成前的 stderr MUST 写入 hc-service 的 glog，并在已知时带上插件名、版本和 pid。

#### Scenario: go-plugin 报告握手或拨号
- **WHEN** `go-plugin` 通过 `Logger` 写出握手或拨号日志
- **THEN** 该日志出现在 hc-service 日志中，并带有插件标识

#### Scenario: 握手前插件写 stderr
- **WHEN** 插件在握手完成前向 stderr 写出内容
- **THEN** 该内容写入 hc-service 的 glog

### Requirement: 插件日志文件跟随 glog 保留和滚动
系统 MUST NOT 在进程 `Close` 或退出时删除插件日志文件。单个插件日志文件达到 glog 的大小上限时 MUST 按服务日志同样的方式滚动。保留个数 MUST 使用 glog 现有的 `logKeeper`。

#### Scenario: 进程已结束
- **WHEN** 插件进程已被 `Kill` 或自行退出
- **THEN** 该插件的日志文件仍留在 `{日志目录}/plugin`，并按 glog 的大小和个数规则保留
