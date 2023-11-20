### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-选型推荐。
- 该接口功能描述：查询机房服务区域接口。

### URL

POST /api/v1/cloud/selection/idcs/services/areas/query

### 输入参数

| 参数名称      | 参数类型           | 必选 | 描述     |
|-----------|----------------|----|--------|
| idc_ids   | string array   | 是  | 机房ID列表 |
| area_topo | AreaInfo array | 是  | 地理拓扑   |

#### AreaInfo

| 参数名称     | 参数类型           | 必选 | 描述       |
|----------|----------------|----|----------|
| name     | string         | 是  | 国家/城市名称  |
| children | AreaInfo array | 是  | 下级地理纬度列表 |

### 调用示例

```json
{
  "idc_ids": [
    "1",
    "2"
  ],
  "area_topo": [
    {
      "name": "中国",
      "children": [
        {
          "name": "香港"
        },
        {
          "name": "广州"
        }
      ]
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "idc_id": "1",
      "service_areas": [
        {
          "country_name": "中国",
          "province_name": "广州",
          "network_latency": 60
        }
      ]
    },
    {
      "idc_id": "2",
      "service_areas": [
        {
          "name": "country_1",
          "children": [
            {
              "name": "province_1_1",
              "value": 60
            },
            {
              "name": "province_1_2",
              "value": 120
            }
          ]
        },
        {
          "country": "country_2",
          "children": [
            {
              "name": "province_2_1",
              "value": 80
            },
            {
              "name": "province_2_2",
              "value": 70
            }
          ]
        }
      ]
    }
  ]
}
```

### 响应参数说明

### 响应参数说明

| 参数名称    | 参数类型                    | 描述   |
|---------|-------------------------|------|
| code    | int32                   | 状态码  |
| message | string                  | 响应信息 |
| data    | IdcServiceAreaRel array | 响应数据 |

#### IdcServiceAreaRel

| 参数名称          | 参数类型              | 描述     |
|---------------|-------------------|--------|
| idc_id        | string            | 机房ID   |
| service_areas | ServiceArea array | 服务区域列表 |

#### ServiceArea

| 参数名称     | 参数类型              | 描述          |
|----------|-------------------|-------------|
| name     | string            | 地理标签，国家/州省等 |
| value    | int               | 网络延迟数据      |
| children | ServiceArea array | 下级地理标签      |
