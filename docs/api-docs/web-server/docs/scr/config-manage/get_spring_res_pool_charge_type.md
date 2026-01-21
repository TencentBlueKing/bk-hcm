### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问权限。
- 该接口功能描述：查询春保资源池计费模式配置。优先返回业务专属配置，如果业务专属配置不存在，则返回全局配置。如果全局配置也不存在，则返回默认值 POSTPAID_BY_HOUR（按量计费）。

### URL

GET /api/v1/woa/bizs/{bk_biz_id}/config/spring_res_pool/charge_type

### 路径参数

| 参数名称   | 参数类型 | 必选 | 描述   |
|-----------|---------|------|--------|
| bk_biz_id | int     | 是   | 业务ID |

### 调用示例

#### 请求示例

```
GET /api/v1/woa/bizs/123/config/spring_res_pool/charge_type
```

### 响应示例

#### 成功响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "charge_type": "PREPAID"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型 | 描述                                    |
|-----------|---------|----------------------------------------|
| result    | bool    | 请求成功与否。true:请求成功；false请求失败 |
| code      | int     | 错误编码。 0表示success，>0表示失败错误   |
| message   | string  | 请求失败返回的错误信息                    |
| data      | object  | 响应数据                                |

#### data

| 参数名称     | 参数类型 | 描述                                                                 |
|------------|---------|---------------------------------------------------------------------|
| charge_type | string  | 计费模式。枚举值：PREPAID（包年包月）、POSTPAID_BY_HOUR（按量计费） |

