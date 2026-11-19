# iac-server-config Specification

## Purpose
定义 `iac-server` 微服务在配置中心（`pkg/cc`）与监控指标包（`pkg/metrics`）中的服务标识常量、最小无状态配置结构 `IacServerSetting`（不含数据库直连配置）及默认配置文件规范。

## Requirements
### Requirement: Service Identity and Metrics Component Constant
配置中心 `pkg/cc` SHALL 定义服务名常量 `IacServerName Name = "iac-server"`，且监控包 `pkg/metrics` SHALL 定义对应的组件标签常量 `ComponentIacServer = string(cc.IacServerName)`。

#### Scenario: Service name initialization and metrics component label alignment
- **WHEN** 进程启动阶段调用 `cc.InitService(cc.IacServerName)`
- **THEN** `cc.ServiceName()` 返回 `"iac-server"`，且 `metrics.ComponentIacServer` 的值等于 `"iac-server"`

### Requirement: Minimal Stateless Configuration Model Without Direct Database Dependency
`pkg/cc` SHALL 定义 `IacServerSetting` 结构体并实现 `Setting` 接口（`trySetFlagBindIP`、`trySetDefault`、`Validate`、`TenantEnable`），且 `IacServerSetting` MUST 仅包含 `Network`（`yaml:"network"`）、`Service`（`yaml:"service"`）、`Log`（`yaml:"log"`）、`Tenant`（`yaml:"tenant"`）四个配置段，SHALL NOT 包含 `Database` 直连数据库配置段。同时 `pkg/cc` SHALL 提供 `cc.IacServer() IacServerSetting` 全局读取函数并在 `loadFromFile` 中支持 `IacServerName` 分支。

#### Scenario: Load valid minimal configuration and populate default values (AC-001)
- **WHEN** 调用 `cc.InitService(cc.IacServerName)` 并使用仅配置 `network.port: 9605`（未显式设置 `bindIP`、`etcd.endpoints`、`log`）的 YAML 文件调用 `cc.LoadSettings(sysOpt)`
- **THEN** `cc.LoadSettings` 返回 `nil` 错误，且 `cc.IacServer().Network.BindIP` 自动填充为 `"127.0.0.1"`、`cc.IacServer().Service.Etcd.Endpoints` 自动填充为 `["127.0.0.1:2379"]`、`cc.IacServer().Service.Etcd.DialTimeoutMS` 自动填充为 `200`

#### Scenario: Reject missing or non-existent configuration file (AC-002-1)
- **WHEN** `cc.InitService(cc.IacServerName)` 已初始化，且以空的 `sysOpt.ConfigFile` 或不存在的文件路径调用 `cc.LoadSettings(sysOpt)`
- **THEN** `cc.LoadSettings` 返回非 `nil` 错误，错误消息包含 `"configuration file path is not configured"` 或 `"load setting from file"`

#### Scenario: Reject conflicting flag and configuration file bindIP (AC-002-2)
- **WHEN** YAML 配置文件中已设置 `network.bindIP: "127.0.0.1"` 且命令行参数同时传入 `sysOpt.BindIP = net.ParseIP("127.0.0.1")`
- **THEN** `cc.LoadSettings` 返回非 `nil` 错误，错误消息包含 `"bind ip only can set by one of the flags or configuration file"`

#### Scenario: Reject invalid bindIP or unpaired TLS certificate and key (AC-002-3, AC-002-4, AC-S01)
- **WHEN** 配置文件中 `network.bindIP` 为非法 IP（如 `"invalid-ip"`），或 `network.tls` / `service.etcd.tls` 中仅配置 `certFile` 而未配置 `keyFile`（或反之）
- **THEN** `cc.LoadSettings`（或 `IacServerSetting.Validate()`）返回非 `nil` 错误，错误消息分别包含 `"invalid network bindIP"` 或 `"cert file and key file must be both specified or both empty"`

#### Scenario: Verify no direct database configuration field exists (AC-008)
- **WHEN** 检查 `cc.IacServerSetting` 结构体定义及默认配置文件 `cmd/iac-server/etc/iac_server.yaml`
- **THEN** `cc.IacServerSetting` 不包含 `Database` 字段，且 `cmd/iac-server/etc/iac_server.yaml` 默认监听端口为 `9605`、不包含任何硬编码明文密码或数据库配置
