## ADDED Requirements

### Requirement: 批量创建GPU模版
系统 SHALL 支持通过data-service API批量创建GPU需求提报模版记录。每条记录包含schema（JSON格式模版内容）、remark（可选备注）、creator（创建人）。系统自动生成唯一ID并设置创建/更新时间。

#### Scenario: 成功批量创建模版
- **GIVEN** 调用方提供包含至少1条、至多100条有效模版数据的请求
- **WHEN** 向 `POST /res_plans/demand_gpu_templates/batch/create` 发送请求
- **THEN** 系统为每条记录生成唯一ID，写入数据库，返回所有生成的ID列表

#### Scenario: 创建时schema为空
- **GIVEN** 调用方提供的模版数据中schema字段为空
- **WHEN** 向创建接口发送请求
- **THEN** 系统返回参数校验错误

### Requirement: 列表查询GPU模版
系统 SHALL 支持通过data-service API按条件查询GPU需求提报模版列表，支持过滤、分页、字段选择和计数。

#### Scenario: 按条件分页查询
- **GIVEN** 调用方提供过滤条件和分页参数
- **WHEN** 向 `POST /res_plans/demand_gpu_templates/list` 发送请求
- **THEN** 系统返回符合条件的模版列表和总数

#### Scenario: 仅查询总数
- **GIVEN** 调用方设置分页参数为count模式
- **WHEN** 向列表查询接口发送请求
- **THEN** 系统仅返回符合条件的记录总数，不返回详情

### Requirement: 批量更新GPU模版
系统 SHALL 支持通过data-service API按ID批量更新GPU需求提报模版记录。可更新字段包括schema、remark、reviser。

#### Scenario: 成功批量更新模版
- **GIVEN** 调用方提供包含至少1条、至多100条更新数据的请求，每条包含有效ID
- **WHEN** 向 `PATCH /res_plans/demand_gpu_templates/batch` 发送请求
- **THEN** 系统更新指定记录的对应字段，updated_at自动更新

#### Scenario: 更新不存在的记录
- **GIVEN** 调用方提供的ID在数据库中不存在
- **WHEN** 向更新接口发送请求
- **THEN** 该条更新操作影响行数为0，不返回错误（与现有模式一致）

### Requirement: 批量删除GPU模版
系统 SHALL 支持通过data-service API按过滤条件批量删除GPU需求提报模版记录。

#### Scenario: 按条件删除模版
- **GIVEN** 调用方提供有效的过滤条件
- **WHEN** 向 `DELETE /res_plans/demand_gpu_templates/batch` 发送请求
- **THEN** 系统先查询符合条件的所有记录ID，再分批删除，使用事务保证一致性

#### Scenario: 过滤条件匹配零条记录
- **GIVEN** 调用方提供的过滤条件不匹配任何记录
- **WHEN** 向删除接口发送请求
- **THEN** 系统正常返回，不报错
