## ADDED Requirements

### Requirement: Internal Client SDK Skeleton for iac-server
`pkg/client/iac-server` SHALL 定义 `Client` 结构体与构造函数 `NewClient(c *client.Capability, version string) *Client`（基础路径 MUST 为 `/api/%s/iac`），且 `pkg/client/api.go` 的 `ClientSet` SHALL 提供 `IacServer() *iacserver.Client` 方法，通过 `cs.discovery(cc.IacServerName)` 获取服务发现能力。

#### Scenario: ClientSet returns iac-server client when discovery is registered (AC-003)
- **WHEN** 调用方的 `serviced.Discover` 已声明包含 `cc.IacServerName`，并通过 `clientSet.IacServer()` 获取客户端实例
- **THEN** 返回非空的 `*iacserver.Client` 实例，且其内部 REST 客户端基础路径绑定为 `/api/v1/iac`

#### Scenario: ClientSet blocks request when iac-server is not in discovery list (AC-004)
- **WHEN** 调用方的 `serviced.Discover` 未声明 `cc.IacServerName`，调用 `clientSet.IacServer()` 并发起请求
- **THEN** `cs.discovery(cc.IacServerName)` 返回 `rdisc.DeniedServers(cc.IacServerName)`，拒绝请求并返回服务未授权发现错误

### Requirement: Upstream Service Discovery and Gateway Reverse Proxy Routing
`cloud-server`、`api-server` 与 `web-server` SHALL 在各自的 `DiscoveryOption.Services` 中注册 `cc.IacServerName`；且 `api-server`（`cmd/api-server/service/proxy.go`）与 `web-server`（`cmd/web-server/service/proxy.go`）SHALL 在 `discoverServices` 列表及 URL 路径第 3 段（`servicePath == "iac"`）路由分发中支持将 `/api/v1/iac/*` 请求代理转发至 `iac-server`。

#### Scenario: Gateway recognizes /api/v1/iac prefix and routes to iac-server (AC-003)
- **WHEN** 客户端向 `api-server` 或 `web-server` 发起 `/api/v1/iac/*` 请求，且 Etcd 中存在在线的 `iac-server` 实例
- **THEN** `api-server` 与 `web-server` 将 `servicePath == "iac"` 解析映射为 `cc.IacServerName`，从服务发现中获取 `iac-server` 节点地址并完成反向代理转发

#### Scenario: Gateway returns no-servers error when no iac-server instance is online (AC-004)
- **WHEN** 客户端向 `api-server` 或 `web-server` 发起 `/api/v1/iac/*` 请求，但 Etcd 中无在线的 `iac-server` 实例
- **THEN** `api-server` 与 `web-server` 拒绝转发并返回无可用服务节点错误（`has no servers` / `Servers Not Found`）
