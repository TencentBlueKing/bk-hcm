## ADDED Requirements

### Requirement: 从Excel数据行提取年份和月份到detail固定字段
系统 SHALL 在构建detail记录时，识别tpl_schema中header.name为"年份"的列并将其值提取为demand_year（int），识别header.name为"月份"的列并将其值提取为demand_month（int）。这两列的值不进入extension数组。

#### Scenario: Excel包含年份和月份列
- **GIVEN** tpl_schema的headers中存在name="年份"和name="月份"的列定义
- **WHEN** 系统解析某行数据，该行年份列值为"2026"，月份列值为"3"
- **THEN** 生成的detail记录中demand_year=2026，demand_month=3，extension不包含这两列的值

#### Scenario: 年份或月份列值为空
- **GIVEN** tpl_schema的headers中存在name="年份"的列定义
- **WHEN** 系统解析某行数据，该行年份列值为空字符串
- **THEN** 生成的detail记录中demand_year=0

#### Scenario: 年份或月份列值非数字
- **GIVEN** tpl_schema的headers中存在name="月份"的列定义
- **WHEN** 系统解析某行数据，该行月份列值为"abc"
- **THEN** 生成的detail记录中demand_month=0（转换失败取零值）

#### Scenario: 模版中无年份/月份列（向后兼容）
- **GIVEN** tpl_schema的headers中不存在name="年份"或name="月份"的列定义
- **WHEN** 系统解析数据行
- **THEN** demand_year=0，demand_month=0，所有列值进入extension
