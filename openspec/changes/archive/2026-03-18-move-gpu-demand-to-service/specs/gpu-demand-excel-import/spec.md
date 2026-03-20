## MODIFIED Requirements

### Requirement: GPU需求Excel导入预览
系统 SHALL 提供Excel导入预览接口，接收用户上传的Excel文件，基于最新GPU模版的tpl_schema进行表结构校验并解析数据，返回tpl_schema定义和解析后的details数据。接口路径为 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/gpu/excel/import`，请求格式为multipart/form-data。实现 SHALL 位于 `cmd/woa-server/service/plan/` 包内，由 service 结构体直接持有业务逻辑方法，不再通过 logics 层 `plan.Logics` 接口中转。

#### Scenario: 成功导入并解析Excel文件
- **GIVEN** 用户上传的Excel文件的sheet名称和列头与最新tpl_schema完全匹配
- **WHEN** 向 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/gpu/excel/import` 上传Excel文件
- **THEN** 系统返回code=0，data包含tpl_schema定义和按每行解析的details数组，每条detail包含demand_type（对应sheet名称）、demand_year、extension（按列顺序的值数组）和validate_result（空数组）

#### Scenario: 上传文件为空或格式非Excel
- **GIVEN** 用户未上传文件或上传的文件不是有效的Excel格式
- **WHEN** 向导入预览接口发送请求
- **THEN** 系统返回code!=0，message说明文件无效

#### Scenario: service层直接处理Excel解析
- **GIVEN** service层持有client依赖可直接调用data-service
- **WHEN** handler接收到Excel导入预览请求
- **THEN** handler SHALL 直接调用service内部方法完成模版查询、文件校验和数据解析，不通过logics层Controller
