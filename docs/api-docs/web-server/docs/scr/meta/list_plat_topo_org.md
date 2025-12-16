### 描述

- 提供版本：v1.7.3+。
- 所需权限：服务-机房裁撤。
- 功能描述：查询指定视角下的组织拓扑以及各子节点的拓扑树。

### URL

POST /api/v1/woa/metas/org_topos/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述                             |
|---------|--------|------|---------------------------------|
| view    | string | 是   | 视角，枚举值：ieg(IEG互动娱乐事业群) |

### 调用示例

```json
{
  "view": "ieg"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "id": "11111",
      "name": "腾讯公司",
      "full_name": "腾讯公司",
      "level": 0,
      "children": [
        {
          "id": "22222",
          "name": "XXXX事业群",
          "full_name": "腾讯公司/XXXX事业群",
          "level": 1,
          "tof_dept_id": 1001,
          "children": [
            {
              "id": "33333",
              "name": "XXXX部",
              "full_name": "腾讯公司/XXXX事业群/XXXX部",
              "level": 2,
              "tof_dept_id": 1002,
              "children": null
            },
            {
              "id": "44444",
              "name": "XX工作室群",
              "full_name": "腾讯公司/XXXX事业群/XXXX群",
              "level": 2,
              "tof_dept_id": 1003,
              "children": [
                {
                  "id": "55555",
                  "name": "XXXX部",
                  "full_name": "腾讯公司/XXXX事业群/XX工作室群/XXXX部",
                  "level": 3,
                  "tof_dept_id": 1004,
                  "children": null
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称  | 参数类型      | 描述    |
|---------|--------------|---------|
| code    | int          | 状态码   |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[i], data[i].children[i]

| 参数名称       | 参数类型      | 描述                              |
|--------------|--------------|-----------------------------------|
| id           | string       | 节点ID                             |
| name         | string       | 节点名称                            |
| full_name    | string       | 到当前节点的全路径名称                |
| level        | int          | 节点所在层级                         |
| tof_dept_id  | int          | 节点在tof平台的ID                    |
| children     | object array | 子节点列表，数据结构类型和data一致      |
