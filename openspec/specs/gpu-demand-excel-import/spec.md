# gpu-demand-excel-import

## Purpose

Defines the GPU demand Excel import preview flow: file upload, template schema validation, sheet/header matching, and data row parsing.

## Requirements

### Requirement: GPU需求Excel导入预览
系统 SHALL 提供Excel导入预览接口，接收用户上传的Excel文件，基于最新GPU模版的tpl_schema进行表结构校验并解析数据，返回tpl_schema定义和解析后的details数据。接口路径为 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/excel/import`，请求格式为multipart/form-data。实现 SHALL 位于 `cmd/woa-server/service/plan/` 包内，由 service 结构体直接持有业务逻辑方法，不再通过 logics 层 `plan.Logics` 接口中转。

#### Scenario: 成功导入并解析Excel文件
- **GIVEN** 用户上传的Excel文件的sheet名称和列头与最新tpl_schema完全匹配
- **WHEN** 向 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/excel/import` 上传Excel文件
- **THEN** 系统返回code=0，data包含tpl_schema定义和按每行解析的details数组，每条detail包含demand_type（对应sheet名称）、demand_year、extension（按列顺序的值数组）和validate_result（空数组）

#### Scenario: 上传文件为空或格式非Excel
- **GIVEN** 用户未上传文件或上传的文件不是有效的Excel格式
- **WHEN** 向导入预览接口发送请求
- **THEN** 系统返回code!=0，message说明文件无效

#### Scenario: service层直接处理Excel解析
- **GIVEN** service层持有client依赖可直接调用data-service
- **WHEN** handler接收到Excel导入预览请求
- **THEN** handler SHALL 直接调用service内部方法完成模版查询、文件校验和数据解析，不通过logics层Controller

### Requirement: Sheet匹配校验
系统 SHALL 校验上传Excel文件的所有sheet名称是否与tpl_schema中定义的sheets名称一一匹配。sheet的数量和名称必须完全一致。

#### Scenario: Sheet名称完全匹配
- **GIVEN** Excel文件包含的sheet名称与tpl_schema.sheets中定义的name字段完全一致（数量和名称均匹配）
- **WHEN** 系统执行sheet匹配校验
- **THEN** 校验通过，继续执行列头校验

#### Scenario: Excel缺少tpl_schema定义的sheet
- **GIVEN** Excel文件缺少tpl_schema中定义的某个sheet（如tpl_schema定义了"混元精调"和"传统AI训练"，但Excel只有"混元精调"）
- **WHEN** 系统执行sheet匹配校验
- **THEN** 系统返回code!=0，message明确指出缺少哪些sheet

#### Scenario: Excel包含tpl_schema未定义的多余sheet
- **GIVEN** Excel文件包含tpl_schema中未定义的额外sheet
- **WHEN** 系统执行sheet匹配校验
- **THEN** 系统返回code!=0，message明确指出多余的sheet名称

### Requirement: 列头匹配校验
系统 SHALL 对每个sheet的列头行进行校验，确保列头与tpl_schema中该sheet的headers定义匹配。列头通过name字段与headers的name字段进行比对。

#### Scenario: 所有sheet列头均匹配
- **GIVEN** 每个sheet的列头行中的列名与tpl_schema中对应sheet的headers.name字段完全一致
- **WHEN** 系统执行列头匹配校验
- **THEN** 校验通过，开始数据解析

#### Scenario: 某个sheet列头不匹配
- **GIVEN** 某个sheet的列头行缺少tpl_schema中定义的列或包含多余的列
- **WHEN** 系统执行列头匹配校验
- **THEN** 系统返回code!=0，message明确指出哪个sheet的哪些列不匹配

### Requirement: 数据行解析
系统 SHALL 按tpl_schema中每个sheet定义的start行号开始读取数据行，将每行数据构建为一条detail记录。每条detail包含demand_type（sheet名称）、demand_year（当前年份）、extension（按headers定义的列顺序提取的值数组）。对于 schema 中定义了 Formula 的列，当缓存值为空时，系统 SHALL 先执行公式校验，校验通过后调用 CalcCellValue 补算值填入行数据，使补算后的值参与后续校验和前端预览展示。

#### Scenario: 按start行号解析数据
- **GIVEN** tpl_schema定义某sheet的start为2
- **WHEN** 系统解析该sheet的数据
- **THEN** 系统从Excel的第2行开始读取数据（第1行为列头），每行生成一条detail记录

#### Scenario: start行号大于2的sheet
- **GIVEN** tpl_schema定义某sheet的start为3（表示第1-2行为表头区域）
- **WHEN** 系统解析该sheet的数据
- **THEN** 系统从Excel的第3行开始读取数据，跳过前2行

#### Scenario: sheet中无数据行
- **GIVEN** 某个sheet在start行号之后没有任何数据行
- **WHEN** 系统解析该sheet的数据
- **THEN** 该sheet不产生任何detail记录，不报错

#### Scenario: 数据行中存在空行
- **GIVEN** 数据区域中存在完全空白的行
- **WHEN** 系统解析该行数据
- **THEN** 系统跳过空行，不生成对应的detail记录

#### Scenario: 公式列缓存缺失的数据行正常解析
- **GIVEN** 某数据行的公式列（如 C 列）缓存值为空，但公式正确且 CalcCellValue 成功补算
- **WHEN** 系统解析该行数据
- **THEN** 该行的 detail 记录中，公式列的值为 CalcCellValue 的计算结果（非空），rawData 展示补算后的值，validate_result 不因该列报 required 错误

#### Scenario: 公式列缓存缺失且补算失败的数据行
- **GIVEN** 某数据行的公式列缓存值为空，公式正确但 CalcCellValue 补算失败
- **WHEN** 系统解析该行数据
- **THEN** 该行的 detail 记录中，公式列的值为空字符串，validate_result 中包含该列的 "必填项不能为空" 错误

### Requirement: 获取最新GPU模版的tpl_schema
系统 SHALL 通过data-service的模版查询接口获取最新创建的GPU模版记录，从中提取tpl_schema作为表结构校验和数据解析的依据。

#### Scenario: 成功获取最新模版
- **GIVEN** 数据库中存在至少一条GPU模版记录
- **WHEN** 系统查询最新模版
- **THEN** 系统获取按创建时间倒序排列的第一条记录的tpl_schema

#### Scenario: 数据库中无模版记录
- **GIVEN** 数据库中不存在任何GPU模版记录
- **WHEN** 系统查询最新模版
- **THEN** 系统返回code!=0，message说明未找到GPU需求模版

### Requirement: Schema结构体对齐新tpl_schema定义
系统 SHALL 提供与API文档tpl_schema JSON结构完全对应的Go结构体定义。Schema包含Sheets数组，每个Sheet包含Name、RowStart（json: row_start）、FixHeaders（json: fix_headers）和Headers两个独立的Header切片。Header结构体 SHALL 包含Name、Type、Field、OrderKey（json: order_key）、Value、Hidden、Required、Formula、Readonly字段。

#### Scenario: 反序列化包含fix_headers和headers的tpl_schema JSON
- **GIVEN** 数据库中存储的tpl_schema JSON包含fix_headers和headers两个独立数组
- **WHEN** 系统将JSON反序列化为Schema结构体
- **THEN** FixHeaders和Headers分别填充为对应的Header切片，RowStart正确映射row_start字段

#### Scenario: Header中包含order_key和hidden属性
- **GIVEN** fix_headers中的列定义包含order_key和hidden字段
- **WHEN** 系统反序列化该header
- **THEN** OrderKey和Hidden字段正确赋值

### Requirement: Excel文件完整性校验适配新结构
系统 SHALL 校验上传的Excel文件与Schema的完整性：sheet名称/数量匹配，以及每个sheet中fix_headers和headers中有实际excel列（field不为"-"）的列头名称匹配。列头行 SHALL 取每个sheet的第(row_start - 1)行。

#### Scenario: 校验fix_headers和headers中有实际列的列头
- **GIVEN** Schema定义某sheet有fix_headers=[{name:"年份",field:"A"},{name:"QPM峰值",field:"-"}]和headers=[{name:"使用场景",field:"C"}]
- **WHEN** 系统校验该sheet的列头行
- **THEN** 系统仅校验"年份"和"使用场景"是否存在于excel列头行，跳过field为"-"的"QPM峰值"

#### Scenario: 使用row_start确定列头行位置
- **GIVEN** Schema定义某sheet的row_start为4
- **WHEN** 系统读取该sheet的列头行
- **THEN** 系统读取第3行（row_start - 1）作为列头行

### Requirement: 数据行解析适配新结构
系统 SHALL 按Sheet定义的row_start行号开始逐行读取数据，根据fix_headers和headers中每个header的field属性（如A、B、C列号）映射读取对应excel列的值，跳过field为"-"的列。每行数据 SHALL 规范化为与有实际excel列的header数量一致的字符串数组。

#### Scenario: 按field映射读取列值
- **GIVEN** Schema定义fix_headers的field分别为A、B、"-"，headers的field分别为C、D、E
- **WHEN** 系统解析一行excel数据
- **THEN** 返回的行数据依次包含A列、B列、C列、D列、E列的值，跳过field为"-"的列

#### Scenario: row_start大于等于2时正常解析
- **GIVEN** Schema定义某sheet的row_start为2
- **WHEN** 系统解析该sheet
- **THEN** 从第2行开始读取数据行，跳过空行

### Requirement: 日志和错误信息英文化
系统中所有与GPU Excel导入相关的日志输出和错误返回信息 SHALL 使用英文。

#### Scenario: 解析失败时返回英文错误
- **GIVEN** 上传的Excel文件格式不合法
- **WHEN** 系统尝试解析该文件
- **THEN** 返回的错误信息为英文描述，如 "invalid excel file format"

### Requirement: AllVisibleHeaders辅助方法
Sheet结构体 SHALL 提供AllVisibleHeaders方法，返回fix_headers和headers中所有field不为"-"且hidden不为true的Header列表，顺序为fix_headers在前、headers在后。

#### Scenario: 获取可见列定义
- **GIVEN** 某Sheet有fix_headers=[{field:"A",hidden:false},{field:"-",hidden:true}]和headers=[{field:"C",hidden:false},{field:"D",hidden:false}]
- **WHEN** 调用AllVisibleHeaders()
- **THEN** 返回field为A、C、D的三个Header（跳过field为"-"的和hidden为true的）

### Requirement: AllExcelHeaders辅助方法
Sheet结构体 SHALL 提供AllExcelHeaders方法，返回fix_headers和headers中所有field不为"-"的Header列表（不考虑hidden属性），用于列头校验场景。

#### Scenario: 获取所有有excel列的header定义
- **GIVEN** 某Sheet有fix_headers=[{field:"A",hidden:true},{field:"-"}]和headers=[{field:"C",hidden:false}]
- **WHEN** 调用AllExcelHeaders()
- **THEN** 返回field为A和C的两个Header（跳过field为"-"的，但保留hidden为true的）
