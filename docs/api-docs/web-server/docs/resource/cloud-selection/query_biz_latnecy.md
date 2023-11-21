### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-选型推荐。
- 该接口功能描述：查询业务延迟数据。

### URL

POST /api/v1/cloud/selections/latency/biz/query

### 输入参数

| 参数名称      | 参数类型           | 必选 | 描述        |
|-----------|----------------|----|-----------|
| area_topo | AreaInfo array | 是  | 需要查询的拓扑列表 |

#### AreaInfo

| 参数名称     | 参数类型           | 必选 | 描述          |
|----------|----------------|----|-------------|
| name     | string         | 是  | 地理标签，国家/州省等 |
| children | AreaInfo array | 否  | 下级地理标签      |

### 调用示例

```json
{
  "area_topo": [
    {
      "name": "country_1",
      "children": [
        {
          "name": "province_1_1"
        },
        {
          "name": "province_1_2"
        }
      ]
    },
    {
      "name": "country_2",
      "children": [
        {
          "name": "province_2_1"
        },
        {
          "name": "province_2_2"
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
  "data": {
    "details": [
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
}
```

### 响应参数说明

| 参数名称    | 参数类型           | 描述   |
|---------|----------------|------|
| code    | int32          | 状态码  |
| message | string         | 请求信息 |
| data    | AreaInfo array | 响应数据 |

#### AreaInfo

| 参数名称     | 参数类型           | 描述          |
|----------|----------------|-------------|
| name     | string         | 地理标签，国家/州省等 |
| value    | double         | 分布权重        |
| children | AreaInfo array | 下级地理标签      |
