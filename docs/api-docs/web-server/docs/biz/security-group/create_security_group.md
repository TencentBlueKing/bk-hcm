### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建安全组。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/create

### 输入参数

| 参数名称          | 参数类型      | 必选 | 描述                               |
|---------------|-----------|----|----------------------------------|
| bk_biz_id     | int64     | 是  | 业务ID                             |
| vendor        | string    | 是  | 供应商（枚举值：tcloud、aws、azure、huawei） |
| account_id    | string    | 是  | 账号ID                             |
| region        | string    | 是  | 地域                               |
| name          | string    | 是  | 安全组名称                            |
| manager       | string    | 是  | 负责人                              |
| bak_manager   | string    | 是  | 备份负责人                            |
| usage_biz_ids | int array | 否  | 不传默认为当前业务，不能传入-1, 传入的业务可以不包含当前业务 |
| memo          | string    | 否  | 备注                               |
| extension     | object    | 否  | 混合云资源差异字段（aws、azure必填）           |

#### extension[aws]

| 参数名称         | 参数类型   | 必选 | 描述     |
|--------------|--------|----|--------|
| cloud_vpc_id | string | 是  | 云VpcID |

#### extension[azure]

| 参数名称                | 参数类型   | 必选 | 描述    |
|---------------------|--------|----|-------|
| resource_group_name | string | 是  | 资源组名称 |

### 调用示例

#### 创建TCloud安全组。

```json
{
  "vendor": "tcloud",
  "account_id": "00000003",
  "region": "ap-guangzhou",
  "name": "sg-create-test",
  "manager": "u1",
  "bak_manager": "u2",
  "usage_biz_ids": [
    123,
    234
  ],
  "memo": "sg test"
}
```

#### 创建Aws安全组。

```json
{
  "vendor": "aws",
  "account_id": "00000012",
  "region": "us-west-2",
  "name": "sg-create-test",
  "memo": "sg test",
  "manager": "u1",
  "bak_manager": "u2",
  "usage_biz_ids": [
    123,
    234
  ],
  "extension": {
    "cloud_vpc_id": "vpc-xxxxx"
  }
}
```

#### 创建HuaWei安全组。

```json
{
  "vendor": "huawei",
  "account_id": "0000001z",
  "region": "ap-southeast-1",
  "name": "sg-create-test",
  "manager": "u1",
  "bak_manager": "u2",
  "usage_biz_ids": [
    123,
    234
  ],
  "memo": "sg test"
}
```

#### 创建Azure安全组。

```json
{
  "vendor": "azure",
  "account_id": "00000024",
  "region": "westus",
  "name": "sg-create-test",
  "memo": "sg test",
  "manager": "u1",
  "bak_manager": "u2",
  "usage_biz_ids": [
    123,
    234
  ],
  "extension": {
    "resource_group_name": "bk"
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001"
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

| 参数名称 | 参数类型   | 描述    |
|------|--------|-------|
| id   | string | 安全组ID |
