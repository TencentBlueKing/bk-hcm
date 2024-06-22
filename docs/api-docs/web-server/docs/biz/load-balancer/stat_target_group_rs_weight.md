### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：统计指定的目标组的权重情况。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/target_groups/targets/weight_stat

### 输入参数

| 参数名称             | 参数类型         | 必选 | 描述    |
|------------------|--------------|----|-------|
| bk_biz_id        | int          | 是  | 业务ID  |
| target_group_ids | string array | 是  | 目标组id |

### 调用示例


```json
{
  "target_group_ids":["00000001","00000002"]
}
```



### 响应示例


```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "target_group_id": "00000001",
      "rs_weight_zero_num": 0,
      "rs_weight_non_zero_num": 2
    },
    {
      "target_group_id": "00000002",
      "rs_weight_zero_num": 0,
      "rs_weight_non_zero_num": 2
    }
  ]
}
```


### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[n]

| 参数名称                   | 参数类型   | 描述         |
|------------------------|--------|------------|
| target_group_id        | string | 目标组id      |
| rs_weight_zero_num     | int    | 权重为0的RS数量  |
| rs_weight_non_zero_num | int    | 权重不为0的RS数量 |

