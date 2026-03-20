## Why

GPU需求提报流程中，业务方需要通过Excel模版批量导入GPU需求数据。前端需要一个"导入预览"接口，上传Excel文件后解析并校验表结构，返回解析后的数据供用户确认后再提交创建。当前系统已有GPU模版CRUD能力（tpl_schema存储在数据库），但缺少Excel文件解析和表结构校验的能力。

## What Changes

- 新增 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/gpu/excel/import` 接口（接口文档：`docs/api-docs/web-server/docs/biz/scr/resource-plan/excel_import_gpu_demand.md`），接收multipart/form-data上传的Excel文件
- 从数据库获取最新GPU模版的tpl_schema，tpl_schema记录了每个sheet的列头定义和数据起始行
- 校验Excel文件的sheet名称/数量是否与tpl_schema中定义的sheets匹配，不匹配则直接返回code!=0
- 校验每个sheet的列头是否与tpl_schema中对应sheet的headers匹配，不匹配则直接返回code!=0
- 文件完整性校验通过后，按tpl_schema中每个sheet的start行号开始解析数据行
- 将解析结果构建为details数组返回给前端，extension按headers列类型做值转换（int→数值，float→浮点数，string/enum→字符串）
- 本阶段暂不校验每行数据的具体值（如枚举值范围、必填校验等），validate_result统一返回空数组
- 响应格式严格对齐接口文档：`{code, message, data: {tpl_schema, details}}`

## Capabilities

### New Capabilities
- `gpu-demand-excel-import`: GPU需求Excel导入预览能力，涵盖Excel文件上传、tpl_schema表结构校验（sheet匹配、列头匹配）、数据解析并返回结构化结果

### Modified Capabilities
- `gpu-template-crud`: 需要通过现有的ListDemandGpuTemplate接口获取最新模版的tpl_schema，无需求变更，仅作为依赖使用

## Impact

- **woa-server**: 新增handler和logic层代码，在`initBizPlanService`注册新路由
- **pkg/api/woa-server**: 新增Excel导入预览的响应类型定义，字段与接口文档一一对应
- **pkg/tools/excel**: 新增通用的Schema驱动Excel校验和解析工具包
- **依赖**: 使用现有data-service的GPU模版查询接口获取tpl_schema
- **依赖**: 使用 `github.com/xuri/excelize/v2` 库解析Excel文件（项目已有此依赖）
