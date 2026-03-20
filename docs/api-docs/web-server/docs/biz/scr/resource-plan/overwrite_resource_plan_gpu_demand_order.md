### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：覆盖GPU需求提报主单，批量删除主单下面所有子单，并批量创建子单。前端通过excel导入预览接口获取解析数据后，将details回传至本接口完成子单的覆盖。

### URL

PATCH /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/order/overwrite

### 输入参数

| 参数名称          | 参数类型             | 必选   | 描述                                     |
|---------------|------------------|------|----------------------------------------|
| order_id      | string           | 是    | 要覆盖上传的主单ID                             |
| details       | object array     | 是    | 子单列表，来自excel导入预览接口返回的details，校验通过的记录回传 |



### details[n]

| 参数名称         | 参数类型   | 必选 | 描述                                                            |
|--------------|--------|----|---------------------------------------------------------------|
| demand_type  | string | 是  | 需求分类，对应tpl_schema中的sheet名称                                    |
| demand_year  | int    | 是  | 需求年份                                                          |
| demand_month | int    | 是  | 需求月份                                                          |
| gpu_num      | int    | 否  | GPU预算卡数,默认为0                                                  |
| qpm_max      | int    | 否  | 峰值QPM，默认为0                                                    |
| extension    | array  | 是  | 扩展数据数组，按tpl_schema中headers定义的列顺序排列，对应非固定列（非fixed_headers）的填写值 |

### 调用示例

```json
{
  "order_id": "00000001",
  "details": [
    {
      "demand_type": "混元精调",
      "demand_year": 2026,
      "demand_month": 1,
      "gpu_num": 100,
      "qpm_max": 0,
      "extension": ["使用场景是文生图","H20","混元12","模块准确","记忆抽取","是",12,"每周一次","全参数","业务逻辑是xx"]
    },
    {
      "demand_type": "混元精调",
      "demand_year": 2026,
      "demand_month": 1,
      "gpu_num": 50,
      "qpm_max": 0,
      "extension": [112,"使用场景是文生图","H20","混元12","模块准确","记忆抽取","是",12,"每周一次","精调ROLA","业务逻辑是xx"],
    },
    {
      "demand_type": "DS火山api调用",
      "demand_year": 2026,
      "demand_month": 1,
      "gpu_num": 0,
      "qpm_max": 13.2,
      "extension": ["使用场景是文生图","火山API-DS-R1",22.1,"deepseekv3",12.0,12.1,12.2,13.2,"业务逻辑是xx"]
    }
  ]
}
```

### 响应示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

无
