## Context

hc-service 今天的云 API 调用走 `service` → `logics` → `pkg/adaptor`。插件化是一条并行路径。协议 5.9 的 gRPC 桩放在 `pkg/plugin/tfplugin5`，进程管理和日志放在 `cmd/hc-service/logics/tfruntime`。上游进程池尚未实现，本期没有任何业务调用 `Start`。

进程模型已经在技术方案里选定：插件进程是 `go-plugin` 服务端，hc-service 用 `go-plugin` 客户端拉起、握手、建立本机 unix gRPC，并在退出时结束进程。`terraform-plugin-go/tfprotov5` 只有公开的请求和响应结构，没有可直接使用的 gRPC 客户端，桩代码必须由 `tfplugin5.proto`（协议 5.9）在本仓库生成。

现有退出机制：

- `pkg/runtime/shutdown` 只监听 `SIGTERM` 和 `SignalShutdownGracefully`。`SIGINT` 不进入这套流程。
- `RegisterFirstShutdown` 只能注册一个回调，已被 `cmd/hc-service/app/app.go` 的 `finalizer`（etcd 反注册）占用，且要求尽快返回。
- `WaitShutdown(20)` 最多等 20 秒，让 `shutdown.AddNotifier` 的回调结束，然后写日志并 `os.Exit`。

日志目录来自 `cc.HCService().Log.LogDir`（`pkg/cc.LogOption`），hc-service 自身日志由 `pkg/logs` 写入该目录。

## Goals / Non-Goals

**Goals:**

- 用新的 `plugin.Client` 拉起一个插件进程，完成握手，只接受协商结果为主版本 5 的进程。
- 将 `tfplugin5.proto` 生成的 gRPC 客户端纳入仓库，并用它实现只支持客户端的 `plugin.GRPCPlugin`。
- `Close` 时先拒绝新调用并等待已发出的调用返回，再调用 `StopProvider`，最后 `Kill`。
- hc-service 走现有优雅退出时，结束仍在运行的插件进程。
- 握手后的 stdout / stderr 写入独立插件日志文件，命名、大小和保留个数复用 glog。`go-plugin` 自身日志和握手前 stderr 写入 hc-service 日志。
- 插件环境变量使用显式清单。生产默认 `TF_LOG=ERROR`，不继承 hc-service 的日志级别，不传入云账号密钥。

**Non-Goals:**

- 不实现北向 `service/plugin` 和进程池 `logics/plugin`。二进制、参数、工作目录、插件名和版本由调用方传入。
- 不实现执行器，不封装 GetSchema、Configure、Plan、Apply、Read、Import、ReadDataSource，不做 `DynamicValue` 编解码。
- 不替换 `pkg/adaptor`，不修改已有资源 Handler。
- 不在本变更里扩大进程监听的信号集合。`SIGINT` 仍不进入 `pkg/runtime/shutdown`。
- 不改 glog 的滚动和采集配置。

## Decisions

### 1. 进程生命周期使用 go-plugin，而不是自行握手拨号

采用技术方案的方案 1。一个 `plugin.Client` 对应一个进程；连接关闭即终止进程。再次拉起必须 `plugin.NewClient`，不能对同一个 `Client` 再次 `Start`。同一个 `ProcessManager` 只成功 `Start` 一次，再次调用返回错误，不新建 `Client`。另一个进程用另一个 `ProcessManager`。

备选方案是自行解析握手行 `1|5|unix|路径|grpc` 并拨号。进程池的槽位本来就是一条连接绑一个进程，自行拨号换来的重连和保留进程没有收益，却要自己维护 cookie、stdio 和退出。

固定握手参数：

| 项 | 值 |
|---|---|
| `HandshakeConfig.ProtocolVersion` | `1` |
| Magic cookie key | `TF_PLUGIN_MAGIC_COOKIE` |
| Magic cookie value | `d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2`（插件握手常量，不是云账号凭证） |
| `VersionedPlugins` | 只登记 `5` |
| Dispense 名称 | `provider` |
| `AllowedProtocols` | 仅 gRPC |
| 传输 | 本机 unix 套接字，地址由插件生成 |
| 消息上限 | `GRPCDialOptions` 中收发至少 256MB |
| TLS | 使用 `go-plugin` 的 AutoMTLS，不另建证书 |

`NegotiatedVersion()` 不是 5 时，立即 `Kill` 并返回错误，不把句柄交给调用方。

### 2. 桩代码在仓库内由 proto 生成，GRPCPlugin 只做客户端

从所固定的 `terraform-plugin-go` 版本取出声明为协议 5.9 的 `tfplugin5.proto`，放入 `pkg/plugin/tfplugin5/`，用 `protoc-gen-go` 与 `protoc-gen-go-grpc` 生成桩代码并提交。`//go:generate` 固定生成命令，避免手写与 proto 漂移。桩代码不放进 `pkg/adaptor`：adaptor 是现有云 SDK 封装，插件协议是另一条路径。进程拉起、关闭和日志仍留在 `cmd/hc-service/logics/tfruntime`，由它引用该桩。

`GRPCClient` 收下 `*grpc.ClientConn`，返回生成的 Provider 客户端。`GRPCServer` 返回错误。hc-service 不承载插件服务端。

本期业务代码只通过该客户端调用 `StopProvider`。生成结果仍包含 proto 里的全部 RPC，供后续执行器使用，本变更不包装这些 RPC。

`tfprotov5` 的公开结构体本期不负责和生成桩互相拷贝。那一步属于执行器，不在本变更。

### 3. 每个 RPC 方法自带门闩，Close 按固定顺序执行

调用方直接调用 `PluginExecutor` 的方法。不提供 `Call` 包装，因为包装约束不住直接调用的人。每个业务 RPC 在方法内部做前置和后置门闩：关闭已开始则失败，通过后记入在途调用，返回时离开。`StopProvider` 由 `Close` 在门闩关闭后调用，不再进入这道门闩。本迭代这些业务方法尚未实现，门闩记在 TODO 里。

`Close` 在同一句柄上重复调用只执行一次。顺序：

1. 标记关闭，拒绝新的方法进入。
2. 等待已经进入门闩的方法返回。
3. 调用 `StopProvider`。该调用失败只记日志，不跳过下一步。
4. 调用 `plugin.Client.Kill()`。按 `go-plugin` 的 gRPC 客户端实现，这一步先请求插件 `Shutdown`，最多等待约 2 秒，再发送 `SIGKILL`。

### 4. 服务退出时用 shutdown notifier 回收进程，不占用 FirstShutdown

每个 `ProcessManager` 在自己第一次 `Start` 成功时注册一个 `shutdown.AddNotifier`。收到退出通知后只关闭自己的进程，然后 `Done`。进程已经关闭时回调立即 `Done`，避免占满 `WaitShutdown(20)` 的 20 秒。多个管理器的统一回收属于进程池，不在本期。

不把回收放进 `hcService.finalizer`。该回调已被占用，且文档要求它尽快返回；进程 `Close` 需要等待在途 RPC 和 `Kill`。

本地 `SIGINT` 仍会跳过这套回收。生产环境由 `SIGTERM` 触发，与现有 hc-service 一致。

### 5. 插件标准输出复用 glog 的文件管理，单独落盘

| 出口 | 去向 |
|---|---|
| `ClientConfig.Logger` | hc-service 日志，带插件名、版本、pid |
| `ClientConfig.Stderr`（握手前） | hc-service 日志 |
| `SyncStdout` / `SyncStderr`（握手后） | `{日志目录}/plugin` 下的插件日志文件，行内标明 `stdout` 或 `stderr` |

插件文件复用 glog 现有的 `logName`、`MaxSize` 和 `logKeeper`。文件名是 `{程序名}.{插件名}.{主机名}.{用户名}.log.INFO.{YYYYMMDD-HHMMSS}.{pid}`，后缀与服务日志相同，只在程序名后插入插件名。插件名、版本和插件进程 pid 写在行内，版本为空时用 `unknown`。pid 尚未可知的行把 pid 留空。进程关闭只关闭文件描述符，不删除文件。文件放在 `plugin` 子目录：服务日志的 `logKeeper.load` 只扫描日志目录本身，这样不必改它的匹配规则。

### 6. 环境变量使用显式清单，日志级别默认 ERROR

`exec.Cmd.Env` 不复制 `os.Environ()`。只设置执行二进制所需的最小环境（如 `PATH`、`TMPDIR`）以及：

- `TF_LOG`：默认 `ERROR`。调用方可以在单次 `Start` 传入 `TRACE`、`DEBUG`、`INFO`、`WARN`、`ERROR` 之一，用于单独拉起一个更高日志级别的进程。其他取值拒绝启动。
- `PLUGIN_PROTOCOL_VERSIONS=5`。
- 若所固定的 `go-plugin` 仍读取 `PLUGIN_MIN_PORT` / `PLUGIN_MAX_PORT`，使用 hc-service 配置 `tfPlugin.minPort` / `tfPlugin.maxPort` 这一组区间。未填写时默认 `40000`–`40999`。该区间不得包含 `network.port`，否则配置校验失败。unix 套接字数据面不使用这两个值。

`TF_PLUGIN_MAGIC_COOKIE` 由 `go-plugin` 按握手配置写入，业务代码不把云密钥放进环境变量。

## Risks / Trade-offs

- [Risk] 握手后没有读者时，插件写 stdout / stderr 会堵住。→ Mitigation：`SyncStdout` / `SyncStderr` 在 `Start` 前装上，并持续写入插件日志文件。
- [Risk] `Kill` 会打断仍在执行的插件调用，连接关闭即进程退出。→ Mitigation：每个业务 RPC 在方法内部进入门闩；`Close` 先拒绝新的进入并等待已进入的方法返回。
- [Risk] `StopProvider` 失败或插件不响应时进程残留。→ Mitigation：无论 `StopProvider` 是否成功都执行 `Kill`；`go-plugin` 在约 2 秒后 `SIGKILL`。
- [Risk] 优雅退出只有 20 秒，排空过久会被 `os.Exit` 打断。→ Mitigation：`Close` 服从这段预算；超时后进程随 hc-service 一起退出。不在本变更里加长全局退出时间。
- [Risk] `SIGINT` 不触发 notifier，本地 Ctrl-C 可能留下插件进程。→ Mitigation：与现有服务一致，回收绑定 `SIGTERM` 和优雅退出。不在本变更修改全局信号。
- [Risk] 插件标准输出增大日志目录占用。→ Mitigation：单文件大小和保留个数沿用 glog。文件在 `plugin` 子目录，行内带插件名。
- [Risk] 256MB 的 gRPC 消息上限提高单连接内存。→ Mitigation：按协议要求设置，一连接一进程，不在本变更再放大。
- [Risk] 生成桩与上游 proto 不一致。→ Mitigation：proto 与生成文件一并提交，生成命令写入 `//go:generate`，版本以 proto 文件头中的 5.9 为准。

## Migration Plan

1. 固定 `terraform-plugin-go` 与 `go-plugin` 版本，放入协议 5.9 的 `tfplugin5.proto` 并生成桩代码。
2. 实现客户端 `GRPCPlugin`、进程句柄和日志写入器。此时尚无业务调用，运行行为不变。
3. 在第一次 `Start` 时登记 shutdown notifier。部署后若没有 `Start`，notifier 不会注册。
4. 回滚时移除该包及其依赖即可。没有表结构、配置项和北向 API 变更。已写出的插件日志文件留在 `{日志目录}/plugin`。

## Open Questions

无。模块小版本按「proto 文件头为协议 5.9」选定为 `terraform-plugin-go` v0.28.0 与 `go-plugin` v1.6.3，在进程管理代码引用它们时再写入 `go.mod`。
