# plugin-runtime-log Specification

## Purpose
定义 Terraform 插件握手前后标准输出与框架日志的分流策略、独立插件日志文件命名格式、行级标识字段以及沿用 `glog` 的日志滚动与保留规范。

## Requirements

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
