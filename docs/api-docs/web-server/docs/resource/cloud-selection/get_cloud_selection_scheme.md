### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-方案查看。
- 该接口功能描述：查询云选型方案详情。

### URL

GET /api/v1/cloud/selections/schemes/{id}

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述   |
|------|--------|----|------|
| id   | string | 是  | 方案ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "bk_biz_id": 310,
    "name": "scheme1",
    "biz_type": "游戏",
    "vendors": [
      "aws",
      "tcloud"
    ],
    "deployment_architecture": [
      "xxxx"
    ],
    "cover_ping": 120,
    "composite_score": 80,
    "net_score": 70,
    "cost_score": 90,
    "cover_rate": 80,
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
        "country": "country_2",
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
    ,
    "result_idc_ids": [
      "00000002",
      "00000007"
    ],
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2023-02-05T15:29:15Z",
    "updated_at": "2023-02-05T15:29:15Z"
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

| 参数名称                    | 参数类型                   | 描述                                   |
|-------------------------|------------------------|--------------------------------------|
| id                      | uint64                 | 审计ID                                 |
| bk_biz_id               | int                    | 业务ID                                 |
| name                    | string                 | 方案名                                  |
| biz_type                | string                 | 业务类型                                 |
| vendors                 | string array           | 供应商（枚举值：tcloud、aws、azure、gcp、huawei） |
| deployment_architecture | string array           | 部署架构                                 |
| cover_ping              | double                 | 用户容忍网络延迟                             |
| composite_score         | double                 | 综合评分                                 |
| net_score               | double                 | 网络评分                                 |
| cost_score              | double                 | 成本评分                                 |
| cover_rate              | double                 | 覆盖率                                  |
| user_distribution       | UserDistribution array | 用户占比                                 |
| result_idc_ids          | string array           | 推荐结果机房ID列表                           |
| creator                 | string                 | 创建者                                  |
| reviser                 | string                 | 更新者                                  |
| created_at              | string                 | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at              | string                 | 更新时间，标准格式：2006-01-02T15:04:05Z       |

#### UserDistribution

| 参数名称     | 参数类型                   | 描述          |
|----------|------------------------|-------------|
| name     | string                 | 地理标签，国家/州省等 |
| value    | double                 | 分布权重        |
| children | UserDistribution array | 下级地理标签      |
