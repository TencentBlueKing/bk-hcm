### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询VPC详情。

### URL

GET /api/v1/cloud/vpcs/{id}

### 输入参数

| 参数名称 | 参数类型   | 必选  | 描述     |
|------|--------|-----|--------|
| id   | string | 是   | VPC的ID |

### 调用示例

```json
```

### 腾讯云响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "tcloud",
    "account_id": "00000001",
    "cloud_id": "vpc-xxxxxxxx",
    "name": "vpc-test",
    "region": "ap-guangzhou",
    "category": "biz",
    "memo": "test vpc",
    "bk_biz_id": 100,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "cidr": [
        {
          "type": "ipv4",
          "cidr": "127.0.0.0/16",
          "category": "master"
        },
        {
          "type": "ipv6",
          "cidr": "::/56",
          "category": "master"
        }
      ],
      "is_default": true,
      "enable_multicast": false,
      "dns_server_set": [
        "127.0.0.1",
        "127.0.0.2"
      ],
      "domain_name": "aa.bb.cc"
    }
  }
}
```

### AWS响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "aws",
    "account_id": "00000001",
    "cloud_id": "vpc-xxxxxxxx",
    "name": "vpc-test",
    "region": "us-east-1",
    "category": "biz",
    "memo": "test vpc",
    "bk_biz_id": 100,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "cidr": [
        {
          "type": "ipv4",
          "cidr": "127.0.0.0/24",
          "address_pool": "",
          "state": "associated"
        },
        {
          "type": "ipv6",
          "cidr": "::/56",
          "address_pool": "Amazon",
          "state": "associated"
        }
      ],
      "state": "available",
      "instance_tenancy": "default",
      "is_default": false,
      "enable_dns_support": false,
      "enable_dns_hostnames": false
    }
  }
}
```

### GCP响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "gcp",
    "account_id": "00000001",
    "cloud_id": "123",
    "name": "vpc-test",
    "category": "biz",
    "memo": "test vpc",
    "bk_biz_id": 100,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "self_link": "https://www.googleapis.com/compute/v1/projects/xxx/global/networks/test",
      "auto_create_subnetworks": false,
      "enable_ula_internal_ipv6": true,
      "internal_ipv6_range": "::/48",
      "mtu": 1460,
      "routing_mode": "REGIONAL"
    }
  }
}
```

### Azure响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "azure",
    "account_id": "00000001",
    "cloud_id": "xxx",
    "name": "vpc-test",
    "region": "northeurope",
    "category": "biz",
    "memo": "test vpc",
    "bk_biz_id": 100,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "resource_group": "test",
      "cidr": [
        {
          "cidr": "127.0.0.0/16",
          "type": "ipv4"
        }
      ],
      "dns_servers": [
        "127.0.0.1"
      ]
    }
  }
}
```

### 华为云响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "huawei",
    "account_id": "00000001",
    "cloud_id": "123",
    "name": "vpc-test",
    "region": "ap-southeast-1",
    "category": "biz",
    "memo": "test vpc",
    "bk_biz_id": 100,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "cidr": [
        {
          "cidr": "127.0.0.0/8",
          "type": "ipv4"
        }
      ],
      "status": "ACTIVE",
      "enterprise_project_id": "0"
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

| 参数名称        | 参数类型   | 描述                                   |
|-------------|--------|--------------------------------------|
| id          | string | VPC的ID                               |
| vendor      | string | 云厂商（枚举值：tcloud、aws、azure、gcp、huawei） |
| account_id  | string | 云账号ID                                |
| cloud_id    | string | VPC的云ID                              |
| name        | string | VPC名称                                |
| region      | string | 地域                                   |
| category    | string | VPC类别（枚举值：biz【业务自用】、backbone【接入骨干网】） |
| memo        | string | 备注                                   |
| bk_biz_id   | int64  | 业务ID，-1表示没有分配到业务                     |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z        |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z        |
| extension   | object | 云厂商私有结构                              |

#### data.extension(tcloud)

| 参数名称             | 参数类型         | 描述         |
|------------------|--------------|------------|
| cidr             | object array | CIDR列表     |
| is_default       | boolean      | 是否是默认VPC   |
| enable_multicast | boolean      | 是否开启组播     |
| dns_server_set   | string array | DNS服务器列表   |
| domain_name      | string       | DHCP域名选项值。 |

#### data.extension(tcloud).cidr

| 参数名称     | 参数类型   | 描述                                            |
|----------|--------|-----------------------------------------------|
| type     | string | 地址类型（枚举值：ipv4、ipv6）                           |
| cidr     | string | CIDR                                          |
| category | string | 类别（枚举值：master【主】、assistant【辅助】、container【容器】） |

#### data.extension(aws)

| 参数名称                 | 参数类型         | 描述                               |
|----------------------|--------------|----------------------------------|
| cidr                 | object array | CIDR信息                           |
| state                | string       | 状态（枚举值：pending、available）        |
| instance_tenancy     | string       | 实例租期（枚举值：default、dedicated、host） |
| is_default           | boolean      | DNS服务器列表                         |
| enable_dns_hostnames | boolean      | 是否启用 DNS 主机名                     |
| enable_dns_support   | boolean      | 是否启用 DNS 解析                      |

#### data.extension(aws).cidr

| 参数名称         | 参数类型   | 描述                                                                         |
|--------------|--------|----------------------------------------------------------------------------|
| type         | string | 地址类型（枚举值：ipv4、ipv6）                                                        |
| cidr         | string | CIDR                                                                       |
| address_pool | string | 地址池                                                                        |
| state        | string | 状态（枚举值：associating、associated、disassociating、disassociated、failing、failed） |

#### data.extension(gcp)

| 参数名称                     | 参数类型    | 描述                          |
|--------------------------|---------|-----------------------------|
| self_link                | string  | 资源URL                       |
| auto_create_subnetworks  | boolean | 是否默认创建子网                    |
| enable_ula_internal_ipv6 | boolean | 是否启用 VPC 网络 ULA 内部 IPv6 范围  |
| internal_ipv6_range      | string  | VPC 网络 ULA 内部 IPv6 范围       |
| mtu                      | int64   | 最大传输单元                      |
| routing_mode             | string  | 动态路由模式（枚举值：REGIONAL、GLOBAL） |

#### data.extension(azure)

| 参数名称           | 参数类型         | 描述       |
|----------------|--------------|----------|
| resource_group | string       | 资源组      |
| dns_servers    | string array | DNS服务器列表 |
| cidr           | object array | CIDR列表   |

#### data.extension(azure).cidr

| 参数名称 | 参数类型   | 描述                  |
|------|--------|---------------------|
| type | string | 地址类型（枚举值：ipv4、ipv6） |
| cidr | string | CIDR                |

#### data.extension(huawei)

| 参数名称                  | 参数类型         | 描述                     |
|-----------------------|--------------|------------------------|
| cidr                  | object array | CIDR列表                 |
| status                | string       | 状态（枚举值：PENDING、ACTIVE） |
| enterprise_project_id | string       | 企业项目ID                 |

#### data.extension(huawei).cidr

| 参数名称 | 参数类型   | 描述                  |
|------|--------|---------------------|
| type | string | 地址类型（枚举值：ipv4、ipv6） |
| cidr | string | CIDR                |
