### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务审计查看。
- 该接口功能描述：查询审计详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/audits/{id}

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述     |
|-------------|--------|----|--------|
| bk_biz_id   | int64  | 是  | 业务ID   |
| id          | uint64 | 是  | 审计ID   |

### 调用示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": 1,
    "res_id": "00000001",
    "cloud_res_id": "sg-xxxxxx",
    "res_name": "test-update",
    "res_type": "security_group",
    "associated_res_id": "",
    "associated_cloud_res_id": "",
    "associated_res_name": "",
    "associated_res_type": "",
    "action": "update",
    "bk_biz_id": -1,
    "vendor": "tcloud",
    "account_id": "00000001",
    "operator": "Jim",
    "source": "api_call",
    "rid": "xxxxxx",
    "app_code": "xxxxxx",
    "detail": {
      "changed": {
        "memo": "update sg test",
        "name": "test-update"
      },
      "data": {
        "account_id": "00000001",
        "bk_biz_id": -1,
        "cloud_id": "sg-xxxxxx",
        "created_at": "2022-12-26T15:49:40Z",
        "creator": "Jim",
        "extension": "{\"cloud_project_id\": \"0\"}",
        "id": "00000001",
        "memo": "update sg test",
        "name": "test-update",
        "region": "ap-guangzhou",
        "reviser": "Jim",
        "updated_at": "2023-02-05T15:29:08Z",
        "vendor": "tcloud"
      }
    },
    "created_at": "2023-02-05T15:29:15Z"
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

| 参数名称                    | 参数类型    | 描述                                                                                                                  |
|-------------------------|---------|---------------------------------------------------------------------------------------------------------------------|
| id                      | uint64  | 审计ID                                                                                                                |
| res_id                  | string  | 资源ID                                                                                                                |
| cloud_res_id            | string  | 云资源ID                                                                                                               |
| res_name                | string  | 资源名称                                                                                                                |
| res_type                | string  | 资源类型                                                                                                                |
| associated_res_id       | string  | 关联资源ID                                                                                                              |
| associated_cloud_res_id | string  | 关联云资源ID                                                                                                             |
| associated_res_name     | string  | 关联资源名称                                                                                                              |
| associated_res_type     | string  | 关联资源类型                                                                                                              |
| action                  | string  | 动作（枚举值：create、update、delete、assign、recycle、recover、reboot、start、stop、reset_pwd、associate、disassociate、bind、deliver） |
| bk_biz_id               | string  | 业务ID                                                                                                                |
| vendor                  | string  | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                                                                                |
| account_id              | string  | 账号ID                                                                                                                |
| operator                | string  | 操作者                                                                                                                 |
| source                  | string  | 请求来源（枚举值：api_call[API调用]、background_sync[后台同步]）                                                                     |
| rid                     | string  | 请求ID                                                                                                                |
| app_code                | string  | 应用代码                                                                                                                |
| detail                  | object  | 审计详情                                                                                                                |
| created_at              | string  | 创建时间，标准格式：2006-01-02T15:04:05Z                                                                                      |

#### detail

| 参数名称     | 参数类型   | 描述                                  |
|----------|--------|-------------------------------------|
| data     | object | 创建资源信息/资源更新前信息/资源删除前信息，且不同资源审计该字段不同 |
| changed  | object | 资源更新信息                              |
