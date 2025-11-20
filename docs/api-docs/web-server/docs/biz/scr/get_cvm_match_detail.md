### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：业务访问。
- 该接口功能描述：亲和性检查接口。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/apply/match/check

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述                            |
|--------------|--------------|------|--------------------------------|
| bk_biz_id    | int	      | 是	 | 业务ID                          |
| specs	       | object array | 是   | 资源申请子需求单信息               |

#### specs
| 参数名称       | 参数类型       | 必选 | 描述                                                |
|---------------|--------------|-----|-----------------------------------------------------|
| region        | string       | 是  | 地域                                                 |
| zones         | string array | 是  | 可用区数组，最大最多传100个 (选“全部”时传all)             |
| device_type   | string       | 是  | 机型                                                 |
| replicas      | int          | 是  | 该机型对应的数量                                       |

### 调用示例

```json
{
  "bk_biz_id": 3,
  "specs": [
    {
        "region": "ap-shanghai",
        "zones": ["ap-shanghai-2"],
        "device_type": "S3.LARGE8",
        "replicas": 2
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "region": "ap-shanghai",
        "zone": "ap-shanghai-2",
        "device_type": "S3.LARGE8",
        "replicas": 2
        "status": 1,
        "max_cut_num": 4,
        "ips": ["x.x.x.x"]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	  | object | 响应数据                      |

#### data

| 参数名称 | 参数类型 | 描述                    |
|---------|--------|-------------------------|
| details | array  | 查询返回的数据             |

#### data.details[n]

| 参数名称            | 参数类型         | 描述                                      |
|--------------------|----------------|-------------------------------------------|
| region             | string         | 地域                                       |
| zone               | string         | 可用区                                     |
| device_type        | string         | 机型                                       |
| replicas           | int            | 该机型对应的数量                             |
| status             | int            | 匹配状态（1:亲和性预检有数据 2:亲和性预检无数据） |
| max_cut_num        | int            | 最大切片数量                                 |
| ips                | string array   | 申请后的母机IP分布                           |
