### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询安全组详情。

### URL

GET /api/v1/cloud/security_groups/{id}

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述    |
|------|----------|----|-------|
| id   | string   | 是  | 安全组ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "tcloud",
    "cloud_id": "sg-xxxxx",
    "region": "ap-guangzhou",
    "name": "sg-demo",
    "memo": "安全组测试demo",
    "account_id": "00000001",
    "bk_biz_id": -1,
    "creator": "jim",
    "reviser": "jim",
    "created_at": "2022-12-26T15:49:40Z",
    "updated_at": "2023-01-11T19:01:15Z",
    "network_interface_count": 0,
    "subnet_count": 0,
    "cvm_count": 0,
    "extension": {
      "cloud_project_id": "0"
    },
    "cloud_created_time": "2024-01-20 15:37:46",
    "cloud_update_time": "2024-01-20 15:46:45",
    "tags": {
      "abc": "123123",
      "module": "vpc"
    }
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

| 参数名称                    | 参数类型              | 描述                                 |
|-------------------------|-------------------|------------------------------------|
| id                      | string            | 安全组ID                              |
| vendor                  | string            | 云厂商                                |
| cloud_id                | string            | 安全组云ID                             |
| bk_biz_id               | int64             | 业务ID                               |
| region                  | string            | 地域                                 |
| name                    | string            | 安全组名称                              |
| memo                    | string            | 备注                                 |
| account_id              | string            | 账号ID                               |
| bk_biz_id               | int64             | 业务ID, -1代表未分配业务。                   |
| cvm_count               | uint64            | 关联虚拟机数量。（tcloud、aws、huawei专属）      |
| network_interface_count | uint64            | 关联网络接口数量。（azure专属）                 |
| subnet_count            | uint64            | 关联子网数量。（azure专属）                   |
| extension               | object[vendor]    | 混合云差异字段                            |
| creator                 | string            | 创建者                                |
| reviser                 | string            | 最后一次修改的修改者                         |
| created_at              | string            | 创建时间，标准格式：2006-01-02T15:04:05Z     |
| updated_at              | string            | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z |
| tags                    | map[string]string | 标签字典                               |
| cloud_created_time      | string            | 安全组云上创建时间，标准格式：2006-01-02 15:04:05 |
| cloud_update_time       | string            | 安全组云上更新时间，标准格式：2006-01-02 15:04:05 |

#### extension[tcloud]

| 参数名称             | 参数类型   | 描述        |
|------------------|--------|-----------|
| cloud_project_id | string | 项目id，默认0。 |

#### extension[aws]

| 参数名称           | 参数类型   | 描述                |
|----------------|--------|-------------------|
| vpc_id         | string | vpc主键ID。          |
| cloud_vpc_id   | string | vpc云主键ID。         |
| cloud_owner_id | string | 拥有该安全组的Amazon账号ID。 |

#### extension[azure]

| 参数名称                        | 参数类型         | 描述                                              |
|-----------------------------|--------------|-------------------------------------------------|
| etag                        | string       | 唯一只读字符串，每当资源更改都会更新。                             |
| flush_connection            | string       | 启用后，在更新规则时，将重新评估从网络安全组连接创建的流。初始启用将触发重新评估。       |
| resource_guid               | string       | 网络安全组资源的资源GUID。                                 |
| provisioning_state          | string       | 资源调配状态。（枚举值：Deleting、Failed、Succeeded、Updating） |
| cloud_network_interface_ids | string array | 关联网络接口云ID列表。                                    |
| cloud_subnet_ids            | string array | 关联子网云ID列表。                                      |

#### extension[huawei]

| 参数名称                        | 参数类型   | 描述                                                               |
|-----------------------------|--------|------------------------------------------------------------------|
| cloud_project_id            | string | 安全组所属的项目ID。                                                      |
| cloud_enterprise_project_id | string | 安全组所属的企业项目ID。取值范围：最大长度36字节，带“-”连字符的UUID格式，或者是字符串“0”。“0”表示默认企业项目。 |

