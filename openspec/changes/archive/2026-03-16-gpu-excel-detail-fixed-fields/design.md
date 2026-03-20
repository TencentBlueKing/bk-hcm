## Context

GPU需求Excel导入预览接口已实现基本的表结构校验和数据解析。当前`buildDetails`将所有列值放入extension数组，`demand_year`硬编码为`time.Now().Year()`，缺少`demand_month`。接口文档已定义这两个字段应从Excel数据行中提取。

Excel模版中**固定存在**name为`"年份"`和`"月份"`的列头，这两个列名是所有模版的约定，不会变化。

## Goals / Non-Goals

**Goals:**
- 从Excel数据行中提取`"年份"`和`"月份"`列的值到detail的demand_year和demand_month
- 新增`demand_month`字段，与接口文档对齐
- 年份/月份列不进入extension，避免数据重复

**Non-Goals:**
- 不修改Header结构体（无需新增mapping字段，通过固定列名识别）
- 不修改tpl_schema的存储和CRUD逻辑
- 不修改表结构校验逻辑

## Decisions

### 1. 通过固定列名常量识别年份/月份列

**选择**: 定义常量`headerNameDemandYear = "年份"`和`headerNameDemandMonth = "月份"`，在`buildDetails`中通过header.Name匹配来识别固定字段列。

**理由**:
- 年份和月份的列名在所有模版中是固定约定，不会变化
- 无需修改Header结构体，改动最小
- 常量定义在logic文件内部即可（非导出），避免污染通用excel包

**替代方案**:
- 在Header上加`mapping`字段 → 过度设计，需要修改Schema结构体和所有模版数据
- 按列位置（固定前N列）→ 不同模版列顺序可能不同

### 2. buildDetails遍历时一次分流

**选择**: 遍历每行数据时，根据header.Name判断：匹配`"年份"`则提取为demand_year（int），匹配`"月份"`则提取为demand_month（int），其余进入extension。

**处理逻辑**:
```
对每行数据:
  demand_year = 0, demand_month = 0
  extension = []
  for i, header in headers:
    val = row[i]
    switch header.Name:
      "年份" → demand_year = parseInt(val)
      "月份" → demand_month = parseInt(val)
      _     → extension.append(convertCellValue(val, header.Type))
```

**理由**: 一次遍历完成分流，年份/月份不进入extension

### 3. 类型转换失败时固定字段取零值

**选择**: 年份/月份列的值如果无法转为int（如空字符串、非数字），demand_year/demand_month取0。

**理由**: 当前阶段不做数据值校验，后续迭代可通过validate_result报告转换失败

## Risks / Trade-offs

- **[风险] 模版中缺少年份/月份列** → demand_year/demand_month为0，数据行全部进入extension，行为退化到当前实现
- **[取舍] 列名硬编码** → 依赖约定而非配置。如果未来列名需要变化，需修改常量。但当前这是稳定约定，可接受
