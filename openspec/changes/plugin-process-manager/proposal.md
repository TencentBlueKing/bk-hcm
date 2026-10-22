## Why

hc-service 即将以并行路径托管 Terraform Provider 插件进程，用来调用云 API。当前仓库没有插件进程生命周期、协议桩代码和插件日志出口。没有这些基础，后续进程池和执行器无法拉起进程、完成握手，也无法在不阻塞插件的前提下拿到运行日志。

本次变更对应 TAPD 需求 [【进程管理与插件日志】](https://tapd.woa.com/tapd_fe/69995598/story/detail/1069995598139221708)，依据 iWiki《技术方案》[第 4 节](https://iwiki.woa.com/p/4047720169) 及其中的进程生命周期（方案 1）和插件日志（方案 2）。

## What Changes

- 在 `cmd/hc-service/logics/tfruntime` 新增进程管理：用 `go-plugin` 拉起插件进程、完成握手和本机 unix gRPC 连接，对外提供 `Start`、`Close`。
- 按协议 5.9 的 `tfplugin5.proto` 在 `pkg/plugin/tfplugin5` 生成 gRPC 桩代码。只登记协议主版本 5；`NegotiatedVersion()` 不是 5 的进程不交给上游。`GRPCServer` 返回错误，hc-service 只做客户端。
- 插件日志采用独立文件：握手后的 stdout / stderr 按插件落盘，并带上插件名、插件版本和 pid。`go-plugin` 自身日志和握手前 stderr 仍写入 hc-service 日志。生产环境拉起时写死 `TF_LOG=ERROR`，不从 hc-service 环境继承日志级别，也不把密钥写入插件环境变量。
- `Close` 先等待已发出的 RPC 返回且不再接受新请求，再调用 `StopProvider`，最后调用 `go-plugin` 的 `Kill`。hc-service 退出前对每个已拉起的子进程执行同样的终止，因为退出信号不会转给插件。
- 不改现有 `service` → `logics` → `pkg/adaptor` 路径，不改已有资源 Handler。
- 不包含北向 `service/plugin`、进程池 `logics/plugin`，也不包含执行器全量 RPC（GetSchema、Configure、Plan、Apply 等）和 `DynamicValue` 编解码。上游负责传入插件二进制、参数和工作目录。

## Capabilities

### New Capabilities

- `plugin-process-lifecycle`: 用 `go-plugin` 管理单个 Terraform Provider 进程的拉起、协议版本约束、在途请求排空和终止，包括 hc-service 退出时清理子进程。
- `tfplugin5-grpc-stub`: 由 `tfplugin5.proto` 生成协议 5 的 gRPC 客户端桩，并提供只支持客户端的 `GRPCPlugin` 适配。
- `plugin-runtime-log`: 把插件握手后的标准输出分开落到独立日志文件，并把框架日志和握手前 stderr 留在 hc-service 日志中。

### Modified Capabilities

无。`openspec/specs/` 下没有插件进程、协议桩或插件日志相关的正式能力。

## Impact

- 桩代码位于 `pkg/plugin/tfplugin5`。进程管理位于 `cmd/hc-service/logics/tfruntime`，只被 hc-service 内部后续的进程池调用，本期不暴露新的北向 HTTP API。
- 新增依赖：`github.com/hashicorp/go-plugin`、`github.com/hashicorp/terraform-plugin-go/tfprotov5`。桩代码由 `tfplugin5.proto` 生成，生成结果纳入仓库，供版本 5 的客户端使用。
- 日志：在 hc-service 现有日志目录下增加插件日志子目录。文件按插件名、版本和 pid 区分，采集仍按目录进行。
- 运行时：子进程使用本机 unix 套接字；连接关闭即终止进程。hc-service 退出路径需要主动结束这些子进程，避免插件残留。
- 不影响现有云资源同步、账单和账号等 adaptor 调用。
