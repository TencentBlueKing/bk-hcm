### 描述

- 该接口提供版本：v1.8.2+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询指定条件的任务管理列表(内部接口，只给特定业务使用，不对外公开)。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/task_managements/list_by_cond

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述                       |
|------------------|--------------|------|---------------------------|
| bk_biz_id        | int64        | 是   | 业务ID                     |
| resource         | string       | 是   | 资源类型，如: 负载均衡：clb、主机：host   |
| start_time       | string       | 是   | 任务最早开始时间，格式：2025-01-01 00:00:00 |
| end_time         | string       | 是   | 任务最晚开始时间，格式：2025-01-01 23:59:59，任务时间跨度最大支持1个月 |
| page             | object       | 是   | 分页设置                    |
| account_ids      | string array | 否   | 账号ID，最大查询数量10个      |
| operations       | string array | 否   | 任务操作类型，最大查询数量10个（创建4层监听器：create_layer4_listener、创建7层监听器：create_layer7_listener、绑定4层RS：binding_layer4_rs、绑定7层RS：binding_layer7_rs、创建URL规则：create_layer7_rule、解绑4层RS：listener_layer4_unbind_rs、解绑7层RS：listener_layer7_unbind_rs、调整4层权重：listener_layer4_rs_weight、调整7层权重：listener_layer7_rs_weight、删除监听器：listener_delete、开机：start_cvm、关机：stop_cvm、重启：reboot_cvm、重装：cvm_reset_system） |
| source           | string       | 否   | 任务来源，如：标准运维插件(sops)、excel导入(excel)、api调用(api) |
| state            | string       | 否   | 任务状态，如：为running（运行中）、failed（失败）、success（成功）、deliver_partial（部分成功）、cancel（取消） |
| creator          | string       | 否   | 操作人                      |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "resource": "clb",
  "start_time": "2025-01-01 00:00:00",
  "end_time": "2025-01-01 23:59:59",
  "page": {
    "count": false,
    "start": 0,
    "limit": 10
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "bk_biz_id": 2,
        "source": "sops",
        "vendors": ["tcloud"],
        "state": "running",
        "account_ids": ["00000001"],
        "resource": "clb",
        "operators": ["create_layer4_listener"],
        "flow_ids": ["00000001"],
        "extension": {},
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z"
      }
    ]
  }
}
```

#### 获取数量返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 100
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

| 参数名称 | 参数类型 | 描述                    |
|---------|--------|-------------------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据             |

#### data.details[n]

| 参数名称     | 参数类型       | 描述                                                                          |
|-------------|--------------|-------------------------------------------------------------------------------|
| id          | string       | 任务管理ID                                                                     |
| bk_biz_id   | int          | 业务id                                                                         |
| source      | string       | 任务来源，如：标准运维插件(sops)、excel导入(excel)                                  |
| vendors     | string array | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                                  |
| state       | string       | 任务状态，如：为running（运行中）、failed（失败）、success（成功）、deliver_partial（部分成功）、cancel（取消） |
| account_ids | string array | 账号ID                                                                         |
| resource    | string       | 资源类型，如: clb、host                                                          |
| operations  | string array | 操作（创建4层监听器：create_layer4_listener、创建7层监听器：create_layer7_listener、绑定4层RS：binding_layer4_rs、绑定7层RS：binding_layer7_rs、创建URL规则：create_layer7_rule、解绑4层RS：listener_layer4_unbind_rs、解绑7层RS：listener_layer7_unbind_rs、调整4层权重：listener_layer4_rs_weight、调整7层权重：listener_layer7_rs_weight、删除监听器：listener_delete、开机：start_cvm、关机：stop_cvm、重启：reboot_cvm、重装：cvm_reset_system）|
| flow_ids    | string array | 关联的后台异步任务flow id数组                                                      |
| extension   | object       | 扩展字段                                                                         |
| creator     | string       | 创建者                                       |
| reviser     | string       | 修改者                                       |
| created_at  | string       | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at  | string       | 修改时间，标准格式：2006-01-02T15:04:05Z       |
