# tfplugin5-grpc-stub Specification

## Purpose
定义 Terraform Plugin Protocol 5.9 的 proto 文件位置、gRPC Go 桩代码生成约定及仅支持客户端模式的 `plugin.GRPCPlugin` 适配规范。

## Requirements

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
