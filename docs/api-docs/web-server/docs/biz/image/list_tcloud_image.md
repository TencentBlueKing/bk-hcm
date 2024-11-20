### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询腾讯云镜像列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/images/query_from_cloud

### 请求参数
| 参数名称       | 参数类型         | 必选 | 描述                                   |
|------------|--------------|----|--------------------------------------|
| bk_biz_id  | int64        | 是  | 业务ID                                 |
| account_id | string       | 是  | 账号ID                                 |
| region     | string       | 是  | 地域ID（唯一标识）                           |
| cloud_ids  | string array | 否  | 镜像ID列表, 不能和filters同时使用               |
| filters    | FilterExp    | 否  | 查询条件。不传时表示查询所有公共镜像, 不能和cloud_ids同时使用 |
| page       | Page         | 是  | 分页设置                                 |


#### Page

| 参数名称   | 参数类型   | 必选 | 描述                  |
|--------|--------|----|---------------------|
| offset | uint32 | 否  | 记录开始位置，offset 起始值为0 |
| limit  | uint32 | 否  | 偏移量, 默认为0, 最大值为100  |

#### FilterExp

每次请求的Filters的上限为10，Filters.Values的上限为5

| 参数名称   | 参数类型         | 必选 | 描述   |
|--------|--------------|----|------|
| name   | string       | 是  | 过滤条件 |
| values | string array | 是  | 过滤值  |

| 过滤条件        | 必选 | 描述                                                                                                                |
|-------------|----|-------------------------------------------------------------------------------------------------------------------|
| image-id    | 否  | 按照【镜像ID】进行过滤                                                                                                      |
| image-type  | 否  | 按照【镜像类型】进行过滤, 可选项：PRIVATE_IMAGE: 私有镜像 (本账户创建的镜像), PUBLIC_IMAGE: 公共镜像 (腾讯云官方镜像) ,SHARED_IMAGE: 共享镜像(其他账户共享给本账户的镜像) |
| image-name  | 否  | 按照【镜像名称】进行过滤                                                                                                      |
| platform    | 否  | 按照【镜像平台】进行过滤, 如CentOS                                                                                             |
| tag-key     | 否  | 按照【标签键】进行过滤                                                                                                       |
| tag-value   | 否  | 按照【标签值】进行过滤                                                                                                       |
| tag:tag-key | 否  | 按照【标签键值对】进行过滤, tag-key使用具体的标签键进行替换                                                                                |


### 调用示例
#### 请求参数示例
查询"公共"镜像列表
```json
{
  "account_id": "00000001",
  "region": "ap-guangzhou",
  "filters": [
    {
      "name": "image-type",
      "values": ["PRIVATE_IMAGE"]
    }
  ],
  "page": {
    "offset": 0,
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
        "details": [
          {
            "cloud_id": "img-j5fjvgpw",
            "name": "test",
            "architecture": "x86_64",
            "platform": "TencentOS",
            "state": "NORMAL",
            "type": "public",
            "image_size": 100,
            "image_source": "CREATE_IMAGE",
            "os_type": "Linux"
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
| 参数名称   | 参数类型        | 描述             |
|--------|-------------|----------------|
| count  | uint64      | 当前规则能匹配到的总记录条数 |
| detail | Image Array | 查询返回的数据        |

#### Image[n]

| 参数名称         | 参数类型   | 描述                                                                                                  |
|--------------|--------|-----------------------------------------------------------------------------------------------------|
| name         | string | 镜像名                                                                                                 |
| cloud_id     | string | 镜像在云厂商上的 ID                                                                                         |
| architecture | string | 镜像架构                                                                                                |
| platform     | string | 镜像平台                                                                                                |
| state        | string | 镜像状态,CREATING-创建中,NORMAL-正常,CREATEFAILED-创建失败,USING-使用中,SYNCING-同步中,IMPORTING-导入中,IMPORTFAILED-导入失败 |
| type         | string | 镜像类型(私有、公共或其他类型)                                                                                    |
| image_size   | int    | 镜像大小                                                                                                |
| image_source | string | 镜像来源 示例值：CREATE_IMAGE                                                                               |
| os_type      | string | 镜像操作系统 示例值：CentOS 8.0 64位                                                                           |
