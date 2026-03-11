### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：服务-机房裁撤。
- 该接口功能描述：查询裁撤数据中当前主机信息。

### URL

POST /api/v1/woa/dissolve/host/current/list

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述     |
|--------------|--------------|----|--------|
| group_ids      | string array | 否  | 运维小组id |
| bk_biz_names | string array | 否  | 业务名称   |
| module_names | string array | 是  | 裁撤模块名称 |
| operators    | string array | 否  | 人员名称   |
| page         | object       | 是  | 分页设置   |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                            |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                            |
| limit | uint32 | 否  | 每页限制条数，不能为0                                                                                                                                   |

### 调用示例

#### 获取详细信息请求参数示例

查询裁撤数据中运维小组id为“1111”, 裁撤模块名称为test, operator为test的当前主机信息。

```json
{
  "group_ids": ["1111"],
  "bk_biz_names": ["test"],
  "module_names": ["test"],
  "operators": ["test"],
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

查询裁撤数据中运维小组id为“1111”, 业务名称为test, 裁撤模块名称为test, operator为test的当前主机数量。

```json
{
  "group_ids": ["1111"],
  "bk_biz_names": ["test"],
  "module_names": ["test"],
  "operators": ["test"],
  "page": {
    "count": true
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
    "details": [
      {
        "server_asset_id": "TC123456",
        "ip": "127.0.0.1",
        "outer_ip": "",
        "app_name": "test",
        "module": "test",
        "device_type": "device_type",
        "svr_type_name": "svr_type_name",
        "module_name": "module_name",
        "idc_unit_name": "idc_unit_name",
        "sfw_name_version": "idc_unit_name",
        "go_up_date": "test",
        "raid_name": "test",
        "logic_area": "logic_area",
        "server_bak_operator": "test",
        "server_operator": "test",
        "device_layer": "device_layer",
        "cpu_score": 1,
        "mem_score": 1,
        "inner_net_traffic_score": 1,
        "disk_io_score": 1,
        "disk_util_score": 1,
        "is_pass": true,
        "mem4linux": 1,
        "inner_net_traffic": 1,
        "outer_net_traffic": 1,
        "disk_io": 1,
        "disk_util": 1,
        "disk_total": 1,
        "max_cpu_core_amount": 1,
        "group_name": "test",
        "center": "test",
        "project_name": "test"
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
    "count": 1
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                | 参数类型   | 描述      |
|---------------------|--------|---------|
|server_asset_id| string | 固资号     |
|ip| string | 内网IP    |
|outer_ip| string | 公网IP    |
|app_name| string | 业务名称    |
|module| string | 业务模块    |
|device_type| string | SCM设备类型 |
|svr_type_name   | string | 服务器类型 |
|module_name| string | 裁撤模块名称  |
|idc_unit_name| string | 存放机房管理单元        |
|sfw_name_version| string | 操作系统        |
|go_up_date| string |   上架时间      |
|raid_name| string |  RAID结构       |
|logic_area| string |  逻辑区域       |
|server_operator| string |  维护人       |
|server_bak_operator| string |   备份维护人      |
|device_layer| string |   设备技术分类      |
|cpu_score| int    |  CPU 得分       |
|mem_score| int    |  内存得分       |
|inner_net_traffic_score| int    |  内网流量得分       |
|disk_io_score| int    |  磁盘IO得分       |
|disk_util_score| int    |  磁盘IO使用率得分       |
|is_pass| bool   |   是否达标      |
|mem4linux| int    |  内存使用量(G)        |
|inner_net_traffic| int    | 内网流量(Mb/s)        |
|outer_net_traffic| int    | 外网流量(Mb/s)        |
|disk_io	| int    |  磁盘IO(Blocks/s)        |
|disk_util| int    | 磁盘IO使用率        |
|disk_total| int    | 磁盘总量(G)        |
|max_cpu_core_amount| int    | CPU核数        |
|group_name| string | 运维小组        |
|center	| string | 业务中心        |
|project_name| string| 主机项目类型|