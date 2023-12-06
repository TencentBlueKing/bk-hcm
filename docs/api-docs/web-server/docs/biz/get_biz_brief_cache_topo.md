### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询CC业务拓扑结构。

### URL

GET /api/v1/web/bizs/{bk_biz_id}/brief/cache/topo

### 输入参数

| 参数名称      | 参数类型  | 必选 | 描述   |
|-----------|-------|----|------|
| bk_biz_id | int64 | 是  | 业务ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "biz": {
      "id": 3,
      "name": "lee",
      "type": 0,
      "bk_supplier_account": "0"
    },
    "idle": [
      {
        "object_id": "set",
        "id": 3,
        "name": "空闲机池",
        "type": 1,
        "nds": [
          {
            "object_id": "module",
            "id": 7,
            "name": "空闲机",
            "type": 1,
            "nds": null
          },
          {
            "object_id": "module",
            "id": 8,
            "name": "故障机",
            "type": 2,
            "nds": null
          },
          {
            "object_id": "module",
            "id": 9,
            "name": "待回收",
            "type": 3,
            "nds": null
          }
        ]
      }
    ],
    "nds": [
      {
        "object_id": "province",
        "id": 22,
        "name": "广东",
        "nds": [
          {
            "object_id": "set",
            "id": 16,
            "name": "magic-set",
            "type": 0,
            "nds": [
              {
                "object_id": "module",
                "id": 48,
                "name": "gameserver",
                "type": 0,
                "nds": null
              },
              {
                "object_id": "module",
                "id": 49,
                "name": "mysql",
                "type": 0,
                "nds": null
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 名称      | 类型     | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data    | object | 请求返回的数据                   |

#### data.biz

| 字段                  | 类型     | 描述                                      |
|---------------------|--------|-----------------------------------------|
| id                  | int    | 业务ID                                    |
| name                | string | 业务名                                     |
| type                | int    | 业务类型，该值>=0，0: 表示该业务为普通业务。1: 表示该业务为资源池业务 |
| bk_supplier_account | string | 开发商账号                                   |

#### data.idle

idle对象中的数据表示该业务的空闲set中的数据，目前只有一个空闲set，后续可能有多个set，请勿依赖此数量。

| 字段        | 类型     | 描述                                                          |
|-----------|--------|-------------------------------------------------------------|
| object_id | string | 该资源的对象，可以是业务自定义层级对应的模块id(bk_obj_id字段值)，set, module等。        
| id        | int    | 该实例的ID                                                      |
| name      | string | 该实例的名称                                                      |
| type      | int    | 该值>=0，只有set和module有该字段，0:表示普通的集群或者模块，>1:表示为空闲机类的set或module。 |
| nds       | object | 该节点所属的子节点信息                                                 |

#### data.nds

描述该业务下除空闲set外的其它拓扑节点的拓扑数据。该对象是一个数组对象，若无其它节点，则为空。
每个节点的对象描述如下，按照拓扑层级，各节点和其对应的子节点逐个嵌套。
需要注意的是，module的"nds"节点一定为空，module是整个业务拓扑树中最底层的节点。

| 字段        | 类型     | 描述                                                          |
|-----------|--------|-------------------------------------------------------------|
| object_id | string | 该资源的对象，可以是业务自定义层级对应的模块id(bk_obj_id字段值)，set, module等。        
| id        | int    | 该实例的ID                                                      |
| name      | string | 该实例的名称                                                      |
| type      | int    | 该值>=0，只有set和module有该字段，0:表示普通的集群或者模块，>1:表示为空闲机类的set或module。 |
| nds       | object | 该节点所属的子节点信息，按照拓扑层级逐级循环嵌套。                                   |
