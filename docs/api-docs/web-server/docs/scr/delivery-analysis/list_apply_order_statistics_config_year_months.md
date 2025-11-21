### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：查询单据配置里的月份列表

### URL

POST /api/v1/woa/task/config/findmany/apply/order/statistics/year_months

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 3,
    "details": [
      {
        "stat_month": "2025-01"
      },
      {
        "stat_month": "2025-02"
      },
      {
        "stat_month": "2025-03"
      }
    ]
  }
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

| 参数名称  | 参数类型         | 描述           |
|-------|--------------|--------------|
| count | int          | 月份总数         |
| details | object array | 月份列表，按年月降序排列 |

#### data.details[n]

| 参数名称 | 参数类型   | 描述            |
|------|--------|---------------|
| stat_month | string | 年月，格式：YYYY-MM |
