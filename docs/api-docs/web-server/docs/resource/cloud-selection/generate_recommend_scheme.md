### 描述

- 该接口提供版本：v1.3.0+。
- 该接口所需权限：资源选型-选型推荐。
- 该接口功能描述：生成云资源选型方案。

### URL

POST /api/v1/cloud/selections/schemes/generate

### 输入参数

| 参数名称                    | 参数类型           | 必选 | 描述                              |
|-------------------------|----------------|----|---------------------------------|
| cover_ping              | number         | 是  | 网络延迟ping值容忍                     |
| deployment_architecture | string array   | 是  | 部署架构 取值：distributed,centralized |
| biz_type                | string         | 否  | 业务类型                            |
| user_distribution       | AreaInfo array | 是  | 用户分布                            |

#### AreaInfo

| 参数名称     | 参数类型           | 必选 | 描述          |
|----------|----------------|----|-------------|
| name     | string         | 是  | 地理标签，国家/州省等 |
| value    | double         | 否  | 人口占比        |
| children | AreaInfo array | 否  | 下级地理标签      |

### 调用示例

```json
{
  "cover_ping": 180,
  "biz_type": "射击",
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
  ]
}

```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "cover_rate": 0.9,
      "composite_score": 75,
      "avg_ping": 100.1,
      "net_score": 50,
      "cost_score": 100,
      "vendors":["tcloud","huawei"],
      "deployment_architecture": "distributed",
      "result_idc_ids": [
        "0000001",
        "0000002",
        "0000003"
      ]
    }
  ]
}
```

### 响应参数说明

| 字段名称    | 字段类型         | 描述   |
|---------|--------------|------|
| code    | int32        | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[n]

| 字段名称                    | 字段类型         | 描述         |
|-------------------------|--------------|------------|
| cover_rate              | double       | 覆盖率        |
| avg_ping                | double       | 平均延迟       |
| composite_score         | double       | 综合评分       |
| net_score               | double       | 网络评分       |
| cost_score              | double       | 成本评分       |
| deployment_architecture | string       | 部署架构       |
| vendors                 | string array | 机房厂商列表     |
| result_idc_ids          | string array | 推荐结果机房ID列表 |
