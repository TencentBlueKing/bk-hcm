## Context

GPU需求提报流程要求业务方通过Excel模版批量导入需求数据。系统已具备GPU模版CRUD能力，数据库中存储了tpl_schema（JSON格式的模版定义，包含sheet结构、列头、起始行等）。现在需要新增一个Excel导入预览接口，在woa-server层完成文件接收、表结构校验和数据解析，返回结构化数据供前端预览确认。

**接口文档**: `docs/api-docs/web-server/docs/biz/scr/resource-plan/excel_import_gpu_demand.md`

当前系统已有类似的Excel导入模式：
- `cloud-server`的负载均衡Excel导入预览（`import_excel_preview.go`），使用`excelize.OpenReader`解析
- `woa-server`的需求周数据导入（`meta.go: ImportDemandWeek`），使用`FormFile`接收文件
- 项目已有`excelize/v2`依赖

本接口仅在woa-server层处理，不需要经过data-service做数据持久化（预览阶段），但需要调用data-service的模版查询接口获取tpl_schema。

## Goals / Non-Goals

**Goals:**
- 实现Excel文件上传和多sheet解析能力
- 基于tpl_schema进行表结构完整性校验（sheet匹配、列头匹配），校验失败直接返回code!=0
- 将Excel数据按tpl_schema定义解析为结构化的details数组返回前端
- extension中的值根据header.type做类型转换（int→int64, float(1)→float64, string/enum→string）
- 响应格式严格对齐接口文档中的字段定义和JSON结构
- 代码具备通用性，tpl_schema驱动的解析逻辑可复用于不同模版

**Non-Goals:**
- 本阶段不校验每行数据的具体值（枚举范围、数值范围、必填校验等），validate_result统一返回空数组
- 不涉及数据持久化，预览数据由前端持有后回传给创建接口
- 不处理前端展示逻辑

## Decisions

### 1. Excel解析逻辑放在woa-server的logics层

**选择**: 在`cmd/woa-server/logics/plan/`下新增GPU Excel导入的logic模块

**理由**: 
- 与现有plan handler/logic分层保持一致
- handler层负责参数提取和路由，logic层负责业务逻辑
- logic层解耦后便于单元测试

**替代方案**: 直接在handler中处理 → 违反现有分层模式，不利于测试和复用

### 2. tpl_schema解析为Go结构体，放在pkg/tools/excel/

**选择**: 在`pkg/tools/excel/schema.go`中定义Schema、Sheet、Header等Go结构体，与接口文档中tpl_schema的JSON结构一一对应。响应类型定义在`pkg/api/woa-server/`中引用此结构体。

**理由**:
- Schema结构体是通用的Excel模版定义，放在tools包可被多处复用
- 响应体中直接引用`*excel.Schema`，JSON序列化后自动匹配接口文档格式
- Header结构体包含name、type、field、value、formula、readonly、required全部字段

### 3. 抽取通用的schema驱动Excel解析器到pkg/tools/excel/

**选择**: 在`pkg/tools/excel/`下新建schema驱动的Excel读取解析包，提供`ValidateFileIntegrity`和`ParseSheetRows`两个通用函数。业务编排留在logic层。

**理由**:
- tpl_schema本身是动态的，校验和解析逻辑不依赖具体业务字段，是纯粹的"按schema读Excel"能力
- 未来其他模版（如API需求、推理需求等）也可能需要Excel导入，放在pkg/tools/下可直接复用
- 与现有项目`pkg/tools/`的工具包定位一致（converter、json、slice等均为通用工具）

**包内设计**:
- `schema.go`: 定义通用的`Schema`、`Sheet`、`Header`结构体（JSON tag与接口文档tpl_schema完全一致）
- `reader.go`: 实现`ValidateFileIntegrity(excelFile, schema)`校验函数和`ParseSheetRows(excelFile, sheet)`解析函数

### 4. 文件完整性校验分两步，失败即终止返回code!=0

**选择**: `ValidateFileIntegrity`内部执行顺序：
1. 先校验sheet名称/数量是否与tpl_schema匹配 → 不匹配直接返回error
2. 再校验每个sheet的列头是否与schema的headers.name匹配 → 不匹配直接返回error

handler层通过`errf.NewFromErr(errf.InvalidParameter, err)`将error转为code!=0的响应。

**理由**:
- 符合用户需求：sheet不匹配直接报错，不继续校验列头
- 校验失败的错误信息明确指出缺少/多余的sheet或列名
- 调用方只需调用一次，逻辑封装在工具包内

### 5. extension值按header.type做类型转换

**选择**: `buildDetails`方法根据每列header的type定义，将原始字符串转换为对应Go类型存入`[]interface{}`数组：
- `string`/`enum` → `string`
- `int` → `int64`（转换失败保留原字符串）
- `float(1)` → `float64`（转换失败保留原字符串）

**理由**:
- 接口文档示例中extension包含混合类型值（如`["文生图","H100",1]`），前端需要正确的类型
- 转换失败时保留原字符串而非报错，因为本阶段不做具体值校验
- 后续迭代加入值校验时，类型转换失败可以写入validate_result

### 6. 获取最新模版的策略

**选择**: 调用data-service的`ListDemandGpuTemplate`接口，按`created_at`倒序排列取第一条

**理由**:
- 复用现有的模版CRUD能力
- 最新创建的模版即为当前生效模版

## Risks / Trade-offs

- **[风险] tpl_schema格式不规范** → 在反序列化时做完整性校验，tpl_schema解析失败返回服务端错误
- **[风险] Excel文件过大导致内存问题** → 使用excelize的流式读取（`Rows()`迭代器），与现有模式一致
- **[取舍] 暂不校验每行数据** → 降低首版复杂度，后续迭代添加validate_result校验逻辑
- **[取舍] 类型转换失败保留原字符串** → 不中断解析流程，后续可通过validate_result反馈给用户
- **[取舍] extension使用数组而非map** → 按API文档定义，extension为按列顺序的值数组，保持与文档一致
