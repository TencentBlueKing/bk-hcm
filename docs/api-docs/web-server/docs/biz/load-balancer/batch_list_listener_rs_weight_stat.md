### 描述

- 该接口提供版本：v1.8.0。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务下查询监听器绑定的rs的权重情况, 七层会遍历下属所有规则绑定的rs

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/listeners/rs_weight_stat

### 输入参数

| 参数名称      | 参数类型         | 必选 | 描述               |
|-----------|--------------|----|------------------|
| bk_biz_id | int          | 是  | 业务ID             |
| ids       | string array | 是  | 监听器ID数组，最大支持100个 |

### 调用示例

```json
{
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "0000000d": {
      "non_zero_weight_count": 6,
      "zero_weight_count": 0,
      "total_count": 6
    },
    "0000001a": {
      "non_zero_weight_count": 1,
      "zero_weight_count": 0,
      "total_count": 1
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型              | 描述                           |
|---------|-------------------|------------------------------|
| code    | int               | 状态码                          |
| message | string            | 请求信息                         |
| data    | map[string]object | 响应数据, key为监听器id，value为rs权重信息 |

#### data[key]参数说明

| 参数名称                  | 参数类型 | 描述        |
|-----------------------|------|-----------|
| non_zero_weight_count | int  | 非零权重的rs数量 |
| zero_weight_count     | int  | 权重为0的rs数量 |
| total_count           | int  | rs总数      |