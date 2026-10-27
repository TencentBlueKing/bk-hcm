## 1. 协议桩与依赖

- [x] 1.1 选定 `github.com/hashicorp/terraform-plugin-go` v0.28.0 与 `github.com/hashicorp/go-plugin` v1.6.3，proto 文件头必须标明协议 5.9。这两个模块在进程管理代码引用它们时再写入 `go.mod`。
- [x] 1.2 将 `tfplugin5.proto` 放入 `pkg/plugin/tfplugin5/`，用 `protoc-gen-go` 和 `protoc-gen-go-grpc` 生成桩代码并提交，生成命令写入 `//go:generate`。
- [x] 1.3 确认生成的 Provider 客户端包含 `StopProvider`，且本包不封装 GetSchema、Configure、Plan、Apply、Read、Import、ReadDataSource，也不实现 `DynamicValue` 编解码。

## 2. 客户端插件适配

- [x] 2.1 实现协议版本 5 的 `plugin.GRPCPlugin`：`GRPCClient` 接收 `*grpc.ClientConn` 并返回生成的 Provider 客户端，`GRPCServer` 返回错误。
- [x] 2.2 固定握手参数：核心协议 `1`、`TF_PLUGIN_MAGIC_COOKIE` 及其常量值、`VersionedPlugins` 只登记 `5`、dispense 名称 `provider`、只允许 gRPC、本机 unix 套接字、收发消息上限至少 256MB、AutoMTLS。

## 3. 进程拉起与关闭

- [x] 3.1 实现 `Start`：成功时创建新的 `plugin.Client`，不复用已有 `Client`。同一个 `ProcessManager` 只能成功一次，再次调用返回错误且不创建 `Client`。协商版本为 5 时返回进程句柄；不是 5 时立即 `Kill` 并返回错误，不返回句柄。
- [ ] 3.2 实现各业务 RPC 时，在方法内部加前置和后置门闩。不提供要求调用方使用的 `Call`。本迭代方法未实现，只记录 TODO。
- [x] 3.3 实现 `Close`：拒绝新的方法进入，等待已进入门闩的调用返回，调用 `StopProvider`，再调用 `plugin.Client.Kill()`。`StopProvider` 失败仍要 `Kill`。同一句柄重复 `Close` 只执行一次。`StopProvider` 不再进入这道门闩。
- [x] 3.4 组装插件环境变量：不复制 `os.Environ()`，包含执行所需的最小环境、`PLUGIN_PROTOCOL_VERSIONS=5` 和 `TF_LOG`。默认 `TF_LOG=ERROR`，只接受 `TRACE`、`DEBUG`、`INFO`、`WARN`、`ERROR`，其他取值拒绝启动。不传入云账号密钥。若该版本 `go-plugin` 读取端口区间，则使用配置项 `tfPlugin.minPort` / `tfPlugin.maxPort` 这一组区间，且不得包含 hc-service 监听端口。

## 4. 插件日志

- [x] 4.1 在 `Start` 前装上 `SyncStdout` 和 `SyncStderr`，并持续读取。握手后的两路输出写入 `{日志目录}/plugin` 下的独立日志文件。文件名、大小上限和保留个数复用 glog。
- [x] 4.2 每行包含 `plugin`、`plugin_version`、`pid` 以及 `stdout` 或 `stderr`。版本为空时 `plugin_version` 使用 `unknown`。pid 未知时该行 pid 为空。文件名在 glog 的 `logName` 中插入插件名。
- [x] 4.3 将 `go-plugin` 的 `Logger` 和握手前 stderr 写入 hc-service 日志，已知时带上插件名、版本和 pid。插件日志文件在进程关闭时保留，滚动和删除沿用 glog 的大小和个数。

## 5. 服务退出回收

- [x] 5.1 在管理器第一次 `Start` 成功时注册它自己的 `shutdown.AddNotifier`。收到 `SIGTERM` 或优雅退出通知后，对该管理器执行 `Close`，然后 `Done`。进程已经关闭时立即 `Done`。不维护跨管理器的进程登记表。
- [x] 5.2 不注册新的 `RegisterFirstShutdown`，不为 `SIGINT` 增加信号处理，不改现有 adaptor 与资源 Handler，不注册北向路由。

## 6. 验证

- [ ] 6.1 为协商版本不是 5、`Close` 顺序、`StopProvider` 失败仍 `Kill`、重复 `Close`、非法 `TF_LOG`、环境变量不含密钥补充单元测试。
- [ ] 6.2 为空版本、行前缀、`GRPCServer` 返回错误补充单元测试。
