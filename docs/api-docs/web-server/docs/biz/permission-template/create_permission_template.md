### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-云权限模板操作。
- 该接口功能描述：创建云权限模板。为指定二级账号创建审批单，审批通过后从权限策略库获取策略内容，调用云API创建权限策略并在本地记录云权限模板。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/create_permission_template

### 路径参数

| 参数名称      | 参数类型   | 必选 | 描述                        |
|-----------|--------|----|---------------------------|
| bk_biz_id | int64  | 是  | 业务ID                      |
| vendor    | string | 是  | 云厂商（枚举值：tcloud） |

### 输入参数

| 参数名称              | 参数类型   | 必选 | 描述                            |
|-------------------|--------|----|-------------------------------|
| account_id        | string | 是  | 目标二级账号ID，需为当前业务的管理账号          |
| policy_library_id | string | 是  | 选择的权限策略库ID                    |
| name              | string | 是  | 云权限模板名称                       |
| memo              | string | 否  | 备注                            |

### 调用示例

```json
{
  "account_id": "00000001",
  "policy_library_id": "00000010",
  "name": "my-permission-template",
  "memo": "用于xxx业务的权限模板"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001"
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

| 参数名称 | 参数类型   | 描述     |
|------|--------|--------|
| id   | string | 审批单据ID |
