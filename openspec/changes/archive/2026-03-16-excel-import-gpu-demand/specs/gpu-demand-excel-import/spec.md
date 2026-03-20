## ADDED Requirements

### Requirement: GPU需求Excel导入预览
系统 SHALL 提供Excel导入预览接口，接收用户上传的Excel文件，基于最新GPU模版的tpl_schema进行表结构校验并解析数据，返回tpl_schema定义和解析后的details数据。接口路径为 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/gpu/excel/import`，请求格式为multipart/form-data。

#### Scenario: 成功导入并解析Excel文件
- **GIVEN** 用户上传的Excel文件的sheet名称和列头与最新tpl_schema完全匹配
- **WHEN** 向 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/gpu/excel/import` 上传Excel文件
- **THEN** 系统返回code=0，data包含tpl_schema定义和按每行解析的details数组，每条detail包含demand_type（对应sheet名称）、demand_year、extension（按列顺序的值数组）和validate_result（空数组）

#### Scenario: 上传文件为空或格式非Excel
- **GIVEN** 用户未上传文件或上传的文件不是有效的Excel格式
- **WHEN** 向导入预览接口发送请求
- **THEN** 系统返回code!=0，message说明文件无效

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
系统 SHALL 按tpl_schema中每个sheet定义的start行号开始读取数据行，将每行数据构建为一条detail记录。每条detail包含demand_type（sheet名称）、demand_year（当前年份）、extension（按headers定义的列顺序提取的值数组）。

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
