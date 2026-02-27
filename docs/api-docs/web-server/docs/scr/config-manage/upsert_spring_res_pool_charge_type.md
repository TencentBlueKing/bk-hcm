### 描述

- 该接口提供版本：v1.8.9.1+。
- 该接口所需权限：全局配置创建权限。
- 该接口功能描述：创建或更新春保资源池计费模式配置。如果传入 bk_biz_id，则创建或更新业务专属配置；如果不传入 bk_biz_id，则创建或更新全局配置。

### URL

POST /api/v1/woa/config/spring_res_pool/charge_type

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                                                                 |
|------------|---------|------|---------------------------------------------------------------------|
| bk_biz_id  | int     | 否   | 业务ID。如果为空则更新全局配置；如果传入则更新业务专属配置              |
| charge_type | string  | 是   | 计费模式。枚举值：PREPAID（包年包月）、POSTPAID_BY_HOUR（按量计费） |

### 调用示例

#### 创建全局配置请求参数示例

```json
{
  "charge_type": "PREPAID"
}
```

#### 创建业务专属配置请求参数示例

```json
{
  "bk_biz_id": 123,
  "charge_type": "POSTPAID_BY_HOUR"
}
```

### 响应示例

#### 成功响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型 | 描述                                    |
|-----------|---------|----------------------------------------|
| result    | bool    | 请求成功与否。true:请求成功；false请求失败 |
| code      | int     | 错误编码。 0表示success，>0表示失败错误   |
| message   | string  | 请求失败返回的错误信息                    |
| data      | object  | 请求返回的数据，成功时为 null            |

