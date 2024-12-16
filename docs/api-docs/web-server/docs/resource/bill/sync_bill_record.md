### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单创建权限。
- 该接口功能描述：账单同步。

### URL

POST /api/v1/account/bills/sync_records

### 输入参数

| 参数名称       | 类型     | 必选 | 描述   |
|------------|--------|----|------|
| vendor     | string | 是  | 云服务商 |
| bill_year  | int    | 是  | 账单年份 |
| bill_month | int    | 是  | 账单月份 |

### 调用示例

```json
{
  "vendor": "huawei",
  "bill_year": 2021,
  "bill_month": 1
}
```


### 响应示例

#### 导出成功结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "ids": ["xxxxx"]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data参数说明

| 参数名称 | 参数类型  | 描述         |
|------|-------|------------|
| ids  | array | 账单同步记录ID列表 |