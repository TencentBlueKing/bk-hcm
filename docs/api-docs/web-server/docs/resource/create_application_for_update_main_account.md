### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：二级账号更新。
- 该接口功能描述：二级账号修改。

### URL

POST /api/v1/cloud/applications/types/update_main_account

## 请求参数
| 参数名称          | 参数类型         | 必选 | 描述                |
|---------------|--------------|----|-------------------|
| id            | string       | 是  | 要变更的账号的id，不可修改    |
| vendor        | string       | 是  | 要变更的账号vendor，不可修改 |
| managers      | string array | 否  | 要变更成为的管理员列表       |
| bak_managers  | string array | 否  | 要变更成为的备份负责人列表     |
| dept_id       | int          | 否  | 要变成成为的部门id        |
| op_product_id | int          | 否  | 要变更成为的运营产品id      |
| bk_biz_id     | int          | 否  | 要变更成为的业务id        |


### 响应数据
```
{
    "code": 0,
    "message": "",
    "data": {
        "id": "xxxxxx"              // string, 海垒申请单ID，非ITSM申请单ID
    }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| id   | string | 单据ID |
