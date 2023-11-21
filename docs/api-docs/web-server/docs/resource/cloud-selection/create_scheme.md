### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-选型推荐。
- 该接口功能描述：保存云资源选型方案。

### URL

POST /api/v1/cloud/selections/schemes/create

### 输入参数

| 参数名称                    | 参数类型           | 必选 | 描述                              |
|-------------------------|----------------|----|---------------------------------|
| bk_biz_id               | int            | 是  | 业务id                            |
| name                    | string         | 是  | 方案名称                            |
| cover_ping              | number         | 是  | 网络延迟ping值容忍                     |
| biz_type_id             | string         | 是  | 业务类型                            |
| deployment_architecture | string array   | 是  | 部署架构 取值：distributed,centralized |
| user_distribution       | AreaInfo array | 是  | 用户分布                            |
| cover_rate              | double         | 是  | 覆盖率                             |
| composite_score         | double         | 是  | 综合评分                            |
| net_score               | double         | 是  | 网络评分                            |
| cost_score              | double         | 是  | 成本评分                            |
| result_idc_ids          | string array   | 是  | 推荐结果机房ID列表                      |

#### AreaInfo

| 参数名称     | 参数类型           | 必选 | 描述          |
|----------|----------------|----|-------------|
| name     | string         | 是  | 地理标签，国家/州省等 |
| value    | double         | 否  | 人口占比        |
| children | AreaInfo array | 否  | 下级地理标签      |

### 调用示例

```json
{
  "bk_biz_id": 2,
  "name": "方案1",
  "cover_ping": 180,
  "biz_type_id": "0000001",
  "deployment_architecture": [
    "distributed"
  ],
  "user_distribution": [
    {
      "name": "country_1",
      "children": [
        {
          "name": "province_1_1",
          "value": 0.1
        },
        {
          "name": "province_1_2",
          "value": 0.1
        }
      ]
    },
    {
      "name": "country_2",
      "children": [
        {
          "name": "province_2_1",
          "value": 0.1
        },
        {
          "name": "province_2_2",
          "value": 0.1
        }
      ]
    }
  ],
  "cover_rate": 0.9,
  "composite_score": 75,
  "net_score": 50,
  "cost_score": 100,
  "result_idc_ids": [
    "0000001",
    "0000002",
    "0000003"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |