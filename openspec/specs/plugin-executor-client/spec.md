# plugin-executor-client Specification

## Purpose
定义 Terraform 插件协议主版本 5 连接的客户端执行器适配、gRPC 请求构造、关闭门闩保护、编解码前置拦截以及 `Private` 字节透传规范。

## Requirements

### Requirement: 执行器只绑定协议主版本 5 的连接
系统 MUST 只在 `NegotiatedVersion()` 为 5 的连接上构造可使用的执行器。协商结果不是 5 时，系统 MUST 结束该进程并返回错误，调用方 MUST NOT 拿到执行器。本期 MUST 只使用已提交的协议 5.9 桩，MUST NOT 另生成一套协议 5.3 客户端。

#### Scenario: 协商结果为主版本 5
- **WHEN** 进程管理完成握手且 `NegotiatedVersion()` 为 5
- **THEN** 调用方拿到绑定该连接的执行器

#### Scenario: 协商结果不是 5
- **WHEN** 插件完成握手但 `NegotiatedVersion()` 不是 5
- **THEN** 系统结束该进程并返回错误，不返回执行器

### Requirement: 执行器直接填写生成的协议请求
`GRPCClient` MUST 接收已建立的 `*grpc.ClientConn`，并持有 `tfplugin5` 生成的 Provider 客户端。执行器 MUST 直接填写生成的 Request。`ProviderMeta`、identity、`ClientCapabilities` 和 deferred MUST 留空。响应里的 `DynamicValue` MUST 直接交给编解码。对外结果仍要使用的 Schema MUST 从生成的 Response 转换。`RequiresReplace` 与 Diagnostics MUST 直接返回生成的 `tfplugin5` 结构。`GRPCServer` MUST 返回错误。hc-service MUST NOT 作为该协议的 gRPC 服务端。

#### Scenario: 建立客户端
- **WHEN** `go-plugin` 以已连接的 `*grpc.ClientConn` 调用 `GRPCClient`
- **THEN** 返回的执行器可以通过生成的客户端发送 Provider 请求并解析响应

#### Scenario: 拒绝作为服务端
- **WHEN** 调用该 `GRPCPlugin` 的 `GRPCServer`
- **THEN** 方法返回错误，不注册 Provider 服务

### Requirement: 关闭开始后不再向插件发出业务 RPC
每个会向插件发 RPC 的执行器方法 MUST 在方法内部做前置和后置门闩。前置检查发现 `Close` 已开始时，方法 MUST 失败，且插件侧该次 RPC 次数为 0。后置在方法返回时离开门闩。系统 MUST NOT 要求调用方经由单独的 `Call` 才能访问这些方法。`StopProvider` 由 `Close` 在门闩关闭后调用，MUST NOT 再进入这道门闩。

#### Scenario: 关闭之后调用读取
- **WHEN** `Close` 已经开始，调用方再调用 `ReadResource`
- **THEN** 调用失败，插件侧 `ReadResource` 次数为 0

#### Scenario: 关闭时存在在途方法
- **WHEN** 一个执行器方法尚未返回时调用方调用 `Close`
- **THEN** 系统不再接受新的方法进入，等待该在途方法返回后才调用 `StopProvider`

### Requirement: 编解码拒绝入参时不发 RPC
执行器 MUST 在把配置或状态发给插件之前调用编解码。编解码返回错误时，执行器 MUST 使该次调用失败，且插件侧该次 RPC 次数为 0。执行器 MUST NOT 在本变更中实现 `DynamicValue` 的类型构造、msgpack 编码或按 Schema 解码。

#### Scenario: 计划入参被编解码拒绝
- **WHEN** 编解码拒绝 `PlanResourceChange` 的入参
- **THEN** 该次计划失败，插件侧 RPC 次数为 0

### Requirement: 不传 ProviderMeta，Private 按字节原样往返
执行器发出的协议请求 MUST NOT 设置 `ProviderMeta`。`Private`、`PriorPrivate` 和 `PlannedPrivate` MUST 按字节原样传入和返回。执行器 MUST NOT 解码这些字节，也 MUST NOT 把它们持久化到执行器之外。

#### Scenario: 读取时带上上次的 Private
- **WHEN** 调用方以一组私有字节调用 `ReadResource`，且插件在响应中返回另一组私有字节
- **THEN** 插件收到的字节与调用方传入的相同，调用方收到的字节与插件返回的相同，执行器不保存这两段字节

### Requirement: 本变更不改既有资源路径
本变更的代码差异 MUST NOT 包含既有资源 Handler 目录和 `pkg/adaptor` 下的文件。系统 MUST NOT 新增北向 HTTP API，也 MUST NOT 实现进程池调度。

#### Scenario: 统计既有目录的变更
- **WHEN** 统计本变更里既有资源 Handler 与 `pkg/adaptor` 的变更文件
- **THEN** 这些目录的变更文件数为 0
