## ADDED Requirements

### Requirement: 查询插件
cloud-server SHALL 提供 `POST /api/v1/cloud/iac_plugins/list`，沿用项目通用 `filter` + `page` 请求格式，MUST 支持按 `id`、`plugin_name`、`protocol`、`state` 过滤。列表与详情共用同一接口，详情以 `id` 过滤。返回字段为注册表各列（含 `extension`）。接口只返回注册表内容，MUST NOT 返回进程池等运行态信息。同名插件有多个版本时返回多条记录。

#### Scenario: 按插件名查全部版本
- **WHEN** 以 `plugin_name=tencentcloud` 过滤
- **THEN** 返回该插件名下所有版本的记录

#### Scenario: 详情
- **WHEN** 以 `id` 过滤
- **THEN** 返回至多一条记录

#### Scenario: 分页
- **WHEN** 传入 `page`
- **THEN** 按分页返回，并支持 count 查询

### Requirement: 查询权限
插件查询接口 SHALL 要求调用者具备平台级权限：资源类型 `iac_plugin`，动作 `find`，映射到 IAM `global_configuration`（平台级，无关联资源实例）。无权限 MUST 返回无权限错误，不返回数据。

#### Scenario: 有权限
- **WHEN** 调用者具备平台全局配置权限
- **THEN** 正常返回列表

#### Scenario: 无权限
- **WHEN** 调用者无该权限
- **THEN** 返回无权限错误，不访问 data-service
