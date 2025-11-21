### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：查询指定月份的配置列表。

### URL

POST /api/v1/woa/task/config/findmany/apply/order/statistics

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述                                    |
|------|--------|----|---------------------------------------|
| stat_month | string | 是  | 年月，格式：YYYY-MM，如 "2025-01"。|

### 调用示例

#### 查询指定月份配置示例

```json
{
  "stat_month": "2025-09"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "id": "config-002",
        "stat_month": "2025-09",
        "bk_biz_id": 104,
        "sub_order_ids": ["000001", "000002", "000003"],
        "start_at": null,
        "end_at": null,
        "memo": ""
      },
      {
        "id": "config-003",
        "stat_month": "2025-09",
        "bk_biz_id": 105,
        "sub_order_ids": ["000001", "000002", "000003"],
        "start_at": "2022-04-30",
        "end_at": "2024-04-30",
        "memo": ""
      },
      {
        "id": "config-004",
        "stat_month": "2025-09",
        "bk_biz_id": 106,
        "sub_order_ids": ["000001", "000002", "000003", "000004", "000005", "000006", "000007"],
        "start_at": null,
        "end_at": null,
        "memo": ""
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

| 参数名称 | 参数类型         | 描述      |
|------|--------------|---------|
| details | object array | 查询返回的数据 |

#### data.details[n]

| 参数名称          | 参数类型        | 描述                                                                   |
|---------------|-------------|----------------------------------------------------------------------|
| id            | string      | 配置ID                                                                 |
| stat_month    | string      | 年月，格式：YYYY-MM                                                        |
| bk_biz_id     | int         | 业务ID                                                                 |
| sub_order_ids | array       | 子单号列表，前端会转换为逗号分隔的字符串显示（如"000001, 000002, 000003"），如果数量超过显示限制，会显示"+N" |
| start_at      | string/null | 开始时间，格式：YYYY-MM-DD。如果为null，前端显示"-"                     |
| end_at        | string/null | 结束时间，格式：YYYY-MM-DD。如果为null，前端显示"-"                     |
| memo          | string      | 备注，如果没有值前端显示"-"                                                      |