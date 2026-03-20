### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：GPU需求模版excel导入预览，上传文件并返回解析校验结果。表结构为动态结构，sheet和列由模版的tpl_schema定义。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/excel/import

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述       |
|-----------|--------|----|----------|
| file      | file   | 是  | excel文件  |

### 调用示例

```multipart/form-data
{
  "file": "gpu_demand.xlsx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
      "sheets":[
       {
          "name":"混元精调",
          "row_start": 3,
          "head_row":2,
          "fixed_headers":[
              {"name": "年份", "type": "enum", "field": "A","db_field":"demand_year","value": [2026, 2027,2028],"hidden":false}, 
              {"name": "月份", "type": "enum", "field": "B","db_field":"demand_month","value": [1,2,3,4,5,6,7,8,9,10,11,12],"hidden":false}, 
              {"name": "预算卡数", "type": "int", "field": "C","db_field":"gpu_num","hidden":true,"required":true}
          ],
          "headers":[
            {"name": "使用场景", "type": "string", "field": "D","hidden":false}, 
            {"name": "卡型", "type": "enum", "field": "E", "value": ["H20", "L20"],"hidden":false}, 
            {"name": "混元模型名称", "type": "string", "field": "F","hidden":false},
            {"name": "主要目标例子：在【具体任务】（如，文本分类、情感分析、问答）上的符合业务场景的要求。", "type": "string", "field": "G","hidden":false}, 
            {"name": "现有数据量", "type": "string", "field": "H","hidden":false}, 
            {"name": "是否有数据采集计划，如有请解释", "type": "string", "field": "I","hidden":false},
            {"name": "单次训练数据量（GB）", "type": "int", "field": "J","hidden":false},
            {"name": "训练周期如，每周一次", "type": "string", "field": "K","hidden":false}, 
            {"name": "精调方式(全参数、LoRA)", "type": "string", "field": "L","hidden":false},
            {"name": "业务逻辑(详细说明qpm和token长度的估算逻辑)", "type": "string", "field": "M","hidden":false}
          ]
        },
        {
          "name":"DS火山api调用",
          "row_start": 4,
          "head_row":2,
          "fixed_headers":[
              {"name": "年份", "type": "enum", "field": "A","db_field":"demand_year","value": [2026, 2027,2028], "hidden":false}, 
              {"name": "月份", "type": "enum", "field": "B","db_field":"demand_month","value": [1,2,3,4,5,6,7,8,9,10,11,12],"hidden":false}, 
              {"name": "QPM峰值", "type": "int", "field": "-","db_field":"qpm_max","hidden":true,"required":true,"formula":"MAX(H3,I3,J3,K3,0)", "readonly": true}, 
          ],
          "headers":[
            {"name": "使用场景", "type": "string", "field": "C"}, 
            {"name": "API类型", "type": "string", "field": "D"}, 
            {"name": "存量额度TPM", "type": "enum", "field": "E", "value": ["CV", "语音", "NLP", "ML"]},
            {"name": "模型名称（模型结构）", "type": "int", "field": "F"},
            {"name": "实际使用TPM", "type": "float", "field": "G"},
            {"name": "Q1峰值TPM/QPM（需同时填写TPM和QPM）", "type": "float", "field": "H"}, 
            {"name": "Q2峰值TPM/QPM（需同时填写TPM和QPM）", "type": "float", "field": "I"}, 
            {"name": "Q3峰值TPM/QPM（需同时填写TPM和QPM）", "type": "float", "field": "J"}, 
            {"name": "Q4峰值TPM/QPM（需同时填写TPM和QPM）", "type": "float", "field": "K"}, 
            {"name": "业务逻辑(详细推导说明)", "type": "string", "field": "L","required": true} 
          ]
        }
      ],
    "details": [
      {
        "name":"混元精调",
        "raw_data":[2026,9,12,"使用场景是文生图","H20","混元12","模块准确","记忆抽取","是",12,"每周一次","全参数","业务逻辑是xx"],
        "validate_result": ["使用场景类型错误"]
      },
      {
        "name":"混元精调",
        "raw_data":[2026,12,112,"使用场景是文生图","H20","混元12","模块准确","记忆抽取","是",12,"每周一次","精调ROLA","业务逻辑是xx"],
        "validate_result": ["当前额度>=0"]
      },
      {
        "name":"DS火山api调用",//DS火山api调用的qpm_max是hidden，从第三个开始是动态字段了
        "raw_data":[2026,9,"使用场景是文生图","火山API-DS-R1",22.1,"deepseekv3",12.0,12.1,12.2,13.2,"业务逻辑是xx"],
        "validate_result": []
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data    | object | 响应数据                      |

### data参数说明

| 参数名称     | 参数类型            | 描述                                |
|----------|-----------------|-----------------------------------|
| sheets   | object array    | 模版schema定义，包含sheets数组，描述各sheet的结构 |
| details  | object array    | excel导入数据详情，每行数据一个元素              |

#### tpl_schema 参数说明

| 参数名称   | 参数类型         | 描述                         |
|--------|--------------|----------------------------|
| sheets | object array | sheet定义列表，每个元素描述一个sheet的结构 |

#### sheets[n] 参数说明

| 参数名称          | 参数类型         | 描述                          |
|---------------|--------------|-----------------------------|
| name          | string       | sheet名称                     |
| head_row      | int          | 表头所在的行  |
| row_start     | int          | 数据起始行号（excel中数据从第几行开始，不含表头） |
| fixed_headers | object array | 固定列定义列表，用于定义排序和聚合等固定字段      |
| headers       | object array | 动态列定义列表，用于定义业务扩展字段          |

#### tpl_schema.sheets[n].fixed_headers[n] / headers[n] 列定义参数说明

| 参数名称      | 参数类型         | 必选 | 描述                                       |
|-----------|--------------|----|------------------------------------------|
| name      | string       | 是  | 列展示名称，对应excel表头                          |
| type      | string       | 是  | 列数据类型，枚举：string、int、float、enum           |
| field     | string       | 是  | 列在excel中的列号，如A、B、C；"-"表示无对应excel列（由公式计算） |
| db_field  | string       | 否  | 固定字段名,为空代表动态字段                           |
| value     | string array | 否  | 当type为enum时，定义可选枚举值列表                    |
| hidden    | bool         | 否  | 是否在前端隐藏该列，默认false                        |
| required  | bool         | 否  | 是否必填，默认为false                            |
| formula   | string       | 否  | excel公式，当该列由公式自动计算时提供，默认为空               |
| readonly  | bool         | 否  | 是否只读，为true时该列由公式自动计算，用户不可编辑，默认为false     |

#### details[n] 参数说明

| 参数名称            | 参数类型         | 描述                                                                          |
|-----------------|--------------|-----------------------------------------------------------------------------|
| name            | string       | sheet名称，对应tpl_schema中的sheet名称                                               |
| raw_data        | array        | 原始行数据数组，按fixed_headers和headers中有对应excel列（field不为"-"）的列顺序排列，不包含hidden为true的列 |
| validate_result | string array | 校验结果详情，空数组表示校验通过，非空时包含具体的校验错误描述                                             |
