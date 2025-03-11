### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单管理权限。
- 该接口功能描述：重新核算一级账号账单。

### URL

POST /api/v1/account/bills/root_account_summarys/reaccount

### 输入参数

| 参数名称            | 类型     | 必选 | 描述     |
|-----------------|--------|----|--------|
| root_account_id | string | 是  | 一级账号ID |
| bill_year       | int    | 是  | 账单年份   |
| bill_month      | int    | 是  | 账单月份   |

### 调用示例

```json
{
  "root_account_id": "xxxx",
  "bill_year": 2021,
  "bill_month": 1
}
```


### 响应示例

#### 导出成功结果示例

```json
{
	"code": 0,
    "message": ""
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |

