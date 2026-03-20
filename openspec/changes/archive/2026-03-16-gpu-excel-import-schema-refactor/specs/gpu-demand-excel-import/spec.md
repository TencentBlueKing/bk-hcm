## MODIFIED Requirements

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

## ADDED Requirements

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
