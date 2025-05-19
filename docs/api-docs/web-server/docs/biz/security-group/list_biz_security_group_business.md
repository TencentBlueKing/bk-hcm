### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询安全组关联资源所属的业务列表，目前仅支持查询关联的CVM和CLB资源。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/{sg_id}/related_resources/bizs/list

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述      |
|-----------|--------|----|---------|
| bk_biz_id | int64  | 是  | 安全组业务ID |
| sg_id     | string | 是  | 安全组ID   |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "cvm": [
      {
        "bk_biz_id": 123,
        "res_count": 0
      },
      {
        "bk_biz_id": 234,
        "res_count": 10
      }
    ],
    "load_balancer": [
      {
        "bk_biz_id": 123,
        "res_count": 600
      }
    ]
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

##### 说明：返回的业务列表中，一定包含管理业务，且一定排在第一个（即使为空）

| 参数名称          | 参数类型         | 描述                         |
|---------------|--------------|----------------------------|
| cvm           | object array | 安全组关联的CVM所属的业务列表           |
| load_balancer | object array | 安全组关联的load balancer所属的业务列表 |

##### cvm[n] && load_balancer[n]

| 参数名称      | 参数类型 | 描述              |
|-----------|------|-----------------|
| bk_biz_id | int  | 资源所属业务ID        |
| res_count | int  | 该业务下的CVM或LB资源总数 |