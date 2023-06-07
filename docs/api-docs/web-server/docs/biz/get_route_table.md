### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询路由表详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/route_tables/{id}

### 输入参数

| 参数名称       | 参数类型   | 必选  | 描述     |
|------------|--------|-----|--------|
| bk_biz_id  | int64  | 是   | 业务ID   |
| id         | string | 是   | 路由表的ID |

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
    "cloud_vpc_id": "vpc-xxxxxxxx",
    "cloud_id": "rtb-xxxxxxxx",
    "name": "test",
    "region": "ap-guangzhou",
    "memo": "test route table",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "main": false
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
    "cloud_vpc_id": "vpc-xxxxxxxx",
    "cloud_id": "rtb-xxxxxxxx",
    "name": "test",
    "region": "us-east-1",
    "memo": "test route table",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "main": false
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
    "cloud_vpc_id": "https://www.googleapis.com/compute/v1/projects/xxx/global/networks/test",
    "cloud_id": "system_generated(1234567890)",
    "name": "系统生成(test)",
    "region": "",
    "memo": "test route table",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20"
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
    "cloud_vpc_id": "vpc-xxxxxxxx",
    "cloud_id": "/subscriptions/xxx/resourceGroups/testrg/providers/Microsoft.Network/routeTables/test",
    "name": "test",
    "region": "centralindia",
    "memo": "test route table",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "cloud_subscription_id": "xxx",
      "resource_group": "testrg"
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
    "cloud_vpc_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "cloud_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "name": "test",
    "region": "cn-south-1",
    "memo": "test route table",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "default": true,
      "tenant_id": "xxxxxxxxxxxxxxxxxxxxxx"
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

| 参数名称         | 参数类型   | 描述                            |
|--------------|--------|-------------------------------|
| id           | string | 路由表的ID                        |
| vendor       | string | 云厂商                           |
| account_id   | string | 账号ID                          |
| cloud_id     | string | 路由表的云ID                       |
| cloud_vpc_id | string | VPC的云ID                       |
| name         | string | 路由表名称                         |
| region       | string | 地域                            |
| memo         | string | 备注                            |
| vpc_id       | string | VPC的云ID                       |
| bk_biz_id    | int64  | 业务ID，-1表示没有分配到业务              |
| creator      | string | 创建者                           |
| reviser      | string | 更新者                           |
| created_at   | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at   | string | 更新时间，标准格式：2006-01-02T15:04:05Z |
| extension    | object | 云厂商私有结构，gcp没有该字段              |

#### data.extension(tcloud)

| 参数名称 | 参数类型    | 描述       |
|------|---------|----------|
| main | boolean | 是否是默认路由表 |

#### data.extension(aws)

| 参数名称 | 参数类型    | 描述      |
|------|---------|---------|
| main | boolean | 是否是主路由表 |

#### data.extension(azure)

| 参数名称                  | 参数类型   | 描述     |
|-----------------------|--------|--------|
| cloud_subscription_id | string | 云上订阅ID |
| resource_group        | string | 资源组    |

#### data.extension(huawei)

| 参数名称      | 参数类型    | 描述       |
|-----------|---------|----------|
| default   | boolean | 是否是默认路由表 |
| tenant_id | string  | 项目ID     |
