### 描述

- 该接口提供版本：v1.9.x。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务维度查询可用镜像列表。返回全部公共镜像、共享镜像以及当前业务的私有镜像；支持全部云厂商；类型化入参，不接受透传 filter。
- 说明：
    - `type` 统一为 `public` / `private` / `shared`（共享镜像独立，不算公共镜像；public 与 shared 均全业务可见）。
    - 本地入库与查询均使用上述统一格式。
    - 私有镜像业务归属口径主要对 tcloud 有意义；其他厂商同步侧主要为公共镜像。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/images/list

#### 路径参数说明

| 参数名称      | 参数类型   | 必选 | 描述                                              |
|-----------|--------|----|-------------------------------------------------|
| bk_biz_id | int64  | 是  | 业务 ID                                           |
| vendor    | string | 是  | 云厂商。枚举值：tcloud / aws / azure / gcp / huawei 等 |

### 输入参数

| 参数名称     | 参数类型   | 必选 | 描述                           |
|----------|--------|----|------------------------------|
| platform | string | 否  | 镜像平台，精确匹配。如 CentOS、TencentOS |
| name     | string | 否  | 镜像名称关键字，模糊匹配（不区分大小写）         |
| type     | string | 否  | 归一化镜像类型。枚举值：public / private / shared |
| region   | string | 否  | 地域，精确匹配。如 ap-guangzhou       |
| page     | Page   | 是  | 分页配置                         |

#### Page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                     |
|-------|--------|----|--------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。为 true 时仅返回 count，不返回 details，且 start/limit 需为 0；为 false 时按 start/limit 返回 details，不返回 count |
| limit | uint   | 是  | 每页限制条数，最大 500，不能为 0                                                                                    |
| start | uint   | 否  | 记录开始位置，起始值为 0                                                                                          |
| sort  | string | 否  | 排序字段                                                                                                   |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                     |

### 调用示例

#### 请求参数示例

查询业务下广州地域、名称含 ubuntu 的公共镜像：

```json
{
  "region": "ap-guangzhou",
  "name": "ubuntu",
  "type": "public",
  "page": {
    "count": false,
    "start": 0,
    "limit": 20
  }
}
```

#### 返回参数示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "vendor": "tcloud",
        "cloud_id": "img-xxx1",
        "name": "Ubuntu Server 22.04 LTS",
        "architecture": "x86_64",
        "platform": "Ubuntu",
        "state": "NORMAL",
        "type": "public",
        "os_type": "Linux",
        "region": "ap-guangzhou",
        "bk_biz_id": -1,
        "creator": "system",
        "reviser": "system",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      },
      {
        "id": "00000002",
        "vendor": "tcloud",
        "cloud_id": "img-xxx2",
        "name": "my-custom-image",
        "architecture": "x86_64",
        "platform": "CentOS",
        "state": "NORMAL",
        "type": "private",
        "os_type": "Linux",
        "region": "ap-guangzhou",
        "bk_biz_id": 100,
        "creator": "user",
        "reviser": "user",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | Data   | 响应数据 |

#### Data

| 参数名称    | 参数类型        | 描述                        |
|---------|-------------|---------------------------|
| count   | uint64      | 匹配总记录数（page.count=true 时） |
| details | Image Array | 镜像列表                      |

#### Image[n]

| 参数名称         | 参数类型   | 描述                                                |
|--------------|--------|---------------------------------------------------|
| id           | string | 镜像 HCM ID                                         |
| vendor       | string | 云厂商                                               |
| cloud_id     | string | 镜像在云厂商上的 ID                                       |
| name         | string | 镜像名称                                              |
| architecture | string | 镜像架构                                              |
| platform     | string | 镜像平台                                              |
| state        | string | 镜像状态                                              |
| type         | string | 统一镜像类型（public / private / shared）                 |
| os_type      | string | 操作系统类型                                            |
| region       | string | 地域                                                |
| bk_biz_id    | int64  | 归属业务 ID，公共/共享镜像一般为 -1                             |
| creator      | string | 创建者                                               |
| reviser      | string | 修改者                                               |
| created_at   | string | 创建时间                                              |
| updated_at   | string | 修改时间                                              |
