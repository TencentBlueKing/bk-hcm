## Context

GPU需求Excel导入预览接口文档（`excel_import_gpu_demand.md`）经过重新设计，tpl_schema结构发生了较大变化：

1. **列头分组**：从单一 `headers` 拆分为 `fix_headers`（固定列，用于排序/聚合/隐藏字段）和 `headers`（动态业务列）
2. **字段重命名**：`start` → `row_start`
3. **新增列属性**：`order_key`（标识固定字段的key）、`hidden`（前端是否隐藏）
4. **Detail格式变更**：从结构化字段（`demand_type`/`demand_year`/`demand_month`/`extension`）变更为通用的 `name` + `raw_data` 格式
5. **raw_data规则**：按 `fix_headers` + `headers` 中 `field != "-"` 且 `hidden != true` 的列顺序排列原始值

当前代码基于旧接口文档实现，Schema结构体、解析逻辑、响应类型均需重构对齐。之前两个变更（`excel-import-gpu-demand`、`gpu-excel-detail-fixed-fields`）的实现成果需要被本次重构覆盖。

## Goals / Non-Goals

**Goals:**
- Schema结构体完全对齐新API文档中tpl_schema的JSON结构
- Detail响应类型对齐新API文档中details的JSON结构
- Excel校验逻辑适配fix_headers + headers分离结构
- 解析逻辑按新规则构建raw_data（可见且有实际excel列的列）
- 所有日志和错误信息使用英文

**Non-Goals:**
- 不涉及行级数据校验逻辑（validate_result仍返回空数组）
- 不涉及create/overwrite接口的变更（这些是下游接口，由前端负责从raw_data提取结构化字段）
- 不涉及数据库tpl_schema存储格式变更（数据库中存的是JSON字符串，Schema结构体变更后反序列化自动适配）
- 不涉及前端代码变更

## Decisions

### Decision 1: Schema结构体直接重写，不引入v2

**选择**: 直接修改现有 `Schema`/`Sheet`/`Header` 结构体，添加新字段、重命名json tag

**理由**: 
- 当前只有GPU Excel导入一处使用这些结构体，不存在多版本兼容需求
- 数据库中tpl_schema以JSON字符串存储，只要json tag正确，反序列化自动适配
- 避免引入不必要的v2抽象层增加代码复杂度

**替代方案**: 新建v2包或v2结构体 → 过度设计，增加维护成本

### Decision 2: Sheet中同时保留FixHeaders和Headers两个独立切片

**选择**: `Sheet` 结构体增加 `FixHeaders []Header` 字段，与现有 `Headers` 字段并列

**理由**:
- 与API文档JSON结构一一对应，反序列化零成本
- fix_headers和headers的业务语义不同（固定字段 vs 动态字段），分离有助于逻辑清晰
- 需要提供 `AllVisibleHeaders()` 等辅助方法，遍历两组header时按fix_headers在前、headers在后的顺序

### Decision 3: raw_data构建逻辑基于field和hidden属性过滤

**选择**: 遍历 `fix_headers` + `headers`，仅包含 `field != "-"` 且 `hidden != true` 的列对应的值

**理由**:
- 严格对齐API文档定义：「raw_data按fix_headers和headers中有对应excel列（field不为"-"）的列顺序排列，不包含hidden为true的列」
- `field == "-"` 表示该列无对应excel列（由公式计算），excel中读不到值
- `hidden == true` 的列（如预算卡数、QPM峰值）不在raw_data中展示

### Decision 4: 列头校验只校验有实际excel列的header

**选择**: `validateHeaders` 校验时，跳过 `field == "-"` 的header（这些是公式计算列，excel中无对应列头）

**理由**:
- `field == "-"` 的列在excel中不存在物理列，无法校验列头
- 校验应覆盖 `fix_headers` 和 `headers` 中所有 `field != "-"` 的列

### Decision 5: ParseSheetRows按field映射读取列值

**选择**: 解析数据行时，根据每个header的 `field` 属性（如A、B、C）确定要读取的excel列，而非按顺序逐列读取

**理由**:
- 新结构中fix_headers和headers的field可能不连续（如fix_headers用A、B、C，headers从D开始）
- 存在 `field == "-"` 的列需要跳过
- 按field映射读取更准确，不受excel中实际列顺序影响

## Risks / Trade-offs

- **[Breaking Change]** Detail响应结构变更导致前端需要同步适配 → 前端团队需提前知悉，确认在同一迭代内完成适配
- **[数据兼容]** 数据库中已有的tpl_schema JSON如果缺少新字段（fix_headers、row_start等） → Go的json.Unmarshal对缺失字段使用零值，不会报错；但需要确认数据库中的模版数据已更新为新格式
- **[列映射复杂度]** 按field属性映射excel列号增加了解析复杂度 → 需要一个field字母到列索引的转换函数，excelize库本身支持列名转索引
