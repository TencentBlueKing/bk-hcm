# plugin-process-lifecycle

## ADDED Requirements

### Requirement: 每次拉起都新建 go-plugin 客户端
`tfruntime` MUST 通过新的 `plugin.Client` 拉起插件进程，并用该客户端完成握手和本机 unix gRPC 连接。同一个 `plugin.Client` MUST NOT 被再次 `Start` 来重拉二进制。调用方需要另一个进程时，MUST 再次调用 `Start` 以创建新的 `Client`。

握手 MUST 使用以下固定值：核心协议版本 `1`；magic cookie 键 `TF_PLUGIN_MAGIC_COOKIE`；magic cookie 值 `d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2`；`VersionedPlugins` 只登记 `5`；dispense 名称为 `provider`；`AllowedProtocols` 只含 gRPC；传输为本机 unix 套接字；`GRPCDialOptions` 的收发消息上限至少 256MB；TLS 使用 `go-plugin` AutoMTLS。

#### Scenario: 成功拉起协议 5 的插件
- **WHEN** 调用方传入可执行的插件二进制、参数和工作目录，且协商版本为 5
- **THEN** 系统返回进程句柄，并且连接使用插件生成的本机 unix 套接字地址

#### Scenario: 再次拉起不复用原来的 Client
- **WHEN** 调用方需要再拉起一个插件进程
- **THEN** 系统创建新的 `plugin.Client`，而不是对已有 `Client` 再次 `Start`

### Requirement: 非协议主版本 5 的进程不得交给调用方
协商完成后，系统 MUST 读取 `NegotiatedVersion()`。结果不是 5 时，系统 MUST 立即 `Kill` 该进程并返回错误，且 MUST NOT 返回进程句柄。

#### Scenario: 插件协商到其他主版本
- **WHEN** 插件完成握手但 `NegotiatedVersion()` 不是 5
- **THEN** 系统结束该进程并返回错误，调用方拿不到句柄

### Requirement: 插件环境变量使用显式清单
系统 MUST NOT 把 hc-service 的 `os.Environ()` 原样复制给插件。插件环境 MUST 只包含执行二进制所需的最小环境，以及 `PLUGIN_PROTOCOL_VERSIONS=5` 和 `TF_LOG`。`TF_LOG` 默认 MUST 为 `ERROR`。系统 MUST 接受调用方为单次 `Start` 指定的 `TRACE`、`DEBUG`、`INFO`、`WARN`、`ERROR` 之一，并只作用于该进程。其他取值 MUST 拒绝启动。云账号密钥 MUST NOT 出现在插件环境变量中。`TF_PLUGIN_MAGIC_COOKIE` MUST 由 `go-plugin` 按握手配置写入。

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
进程句柄 MUST 提供 `Call` 与 `Close`。`Call` MUST 在门闩内把生成的 Provider 客户端交给回调。`Close` 开始后，新的 `Call` MUST 失败且不执行回调。`Close` MUST 按顺序执行：拒绝新的 `Call`，等待已经进入的 `Call` 返回，调用 `StopProvider`，然后调用 `plugin.Client.Kill()`。`StopProvider` 失败时系统 MUST 仍执行 `Kill`。同一句柄上的重复 `Close` MUST 只执行一次该顺序。

#### Scenario: 关闭时存在在途调用
- **WHEN** 一个 `Call` 尚未返回时调用方调用 `Close`
- **THEN** 系统不再接受新的 `Call`，等待该在途 `Call` 返回后才调用 `StopProvider`，然后再 `Kill`

#### Scenario: StopProvider 失败
- **WHEN** `Close` 调用 `StopProvider` 返回错误
- **THEN** 系统记录该错误并仍然 `Kill` 插件进程

#### Scenario: 重复关闭
- **WHEN** 对同一句柄调用两次 `Close`
- **THEN** 停止顺序只执行一次

### Requirement: 优雅退出时回收已拉起的插件进程
系统 MUST 在第一次 `Start` 时注册一个 `shutdown.AddNotifier`。hc-service 因 `SIGTERM` 或 `SignalShutdownGracefully` 退出时，该回调 MUST 对登记表中的每个句柄执行 `Close`，然后调用 `Done`。登记表为空时，回调 MUST 立即 `Done`。系统 MUST NOT 另注册 `RegisterFirstShutdown`，也 MUST NOT 在本变更中为 `SIGINT` 增加信号处理。

#### Scenario: 服务收到 SIGTERM
- **WHEN** hc-service 进入现有优雅退出，且仍有插件进程句柄
- **THEN** 系统关闭这些句柄，并在回调结束时调用 notifier 的 `Done`

#### Scenario: 已经没有运行中的插件
- **WHEN** 退出通知到达时登记表为空
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

### Requirement: 握手后的标准输出写入独立插件日志
系统 MUST 在 `Start` 之前装上 `SyncStdout` 和 `SyncStderr`，并持续读取，避免插件因管道写满而阻塞。握手之后插件写到 stdout 或 stderr 的内容 MUST 写入同一个文件：`{LogDir}/plugin/{safeName}-{safeVersion}-{pid}.log`。`LogDir` 默认 MUST 为 hc-service 的 `LogDir`。每一行 MUST 包含 `plugin`、`plugin_version`、`pid`，并标明来自 `stdout` 还是 `stderr`。

插件名和版本中不属于 `A-Za-z0-9._-` 的字符 MUST 替换为 `_`。版本为空时，文件名中的版本段 MUST 为 `unknown`。pid 尚未可知时，系统 MUST 缓冲已到达的输出，并在 pid 可用后写入上述文件，不得丢弃这段输出。

#### Scenario: 插件写出 stdout 和 stderr
- **WHEN** 握手完成后插件向 stdout 和 stderr 写入日志
- **THEN** 两路内容都出现在 `{LogDir}/plugin/{safeName}-{safeVersion}-{pid}.log`，且每行带有插件名、版本、pid 和对应的流名称

#### Scenario: 插件名包含不能用于文件名的字符
- **WHEN** 插件名为 `registry.terraform.io/hashicorp/aws`，版本为 `1.2.3`，pid 为 `42`
- **THEN** 日志文件名使用把非法字符替换为 `_` 后的插件名，并包含版本 `1.2.3` 和 pid `42`

#### Scenario: 版本未提供
- **WHEN** 调用方没有提供插件版本
- **THEN** 日志文件名中的版本段为 `unknown`

#### Scenario: 握手期间 pid 尚未可用
- **WHEN** 插件在 pid 被观察到之前已经写出日志
- **THEN** 这段内容在 pid 可用后进入带 pid 的日志文件，不丢失

### Requirement: 框架日志和握手前 stderr 进入 hc-service 日志
`go-plugin` 的 `Logger` 输出以及握手完成前的 stderr MUST 写入 hc-service 自身日志，并在已知时带上插件名、版本和 pid。这些内容 MUST NOT 只写入插件独立日志文件。

#### Scenario: go-plugin 报告握手或拨号
- **WHEN** `go-plugin` 通过 `Logger` 写出握手或拨号日志
- **THEN** 该日志出现在 hc-service 日志中，并带有插件标识

#### Scenario: 握手前插件写 stderr
- **WHEN** 插件在握手完成前向 stderr 写出内容
- **THEN** 该内容写入 hc-service 日志，而不是只写入插件独立日志文件

### Requirement: 进程退出后保留插件日志文件
系统 MUST NOT 在进程 `Close` 或退出时删除该进程的插件日志文件。

#### Scenario: 进程已结束
- **WHEN** 插件进程已被 `Kill` 或自行退出
- **THEN** `{LogDir}/plugin/` 下对应的日志文件仍然存在
