### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询华为云账号地域配额。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/huawei/accounts/{account_id}/regions/quotas

### 请求参数

| 参数名称       | 参数类型   | 必选  | 描述   |
|------------|--------|-----|------|
| bk_biz_id  | int64  | 是   | 业务ID |
| account_id | string | 是   | 账号ID |
| vendor     | string | 是   | 供应商  |
| region     | string | 是   | 地域   |

### 调用示例

#### 请求参数示例

```json
{
  "region": "cn-south-1"
}
```

#### 返回参数示例

查询腾讯云机型列表返回参数。
```json
{
  "code": 0,
  "message": "",
  "data": {
    "max_image_meta": 128,
    "max_personality": 5,
    "max_personality_size": 10240,
    "max_security_group_rules": 20,
    "max_security_groups": 10,
    "max_server_group_members": 16,
    "max_server_groups": 10,
    "max_server_meta": 128,
    "max_total_cores": 800,
    "max_total_floating_ips": 10,
    "max_total_instances": 200,
    "max_total_keypairs": 100,
    "max_total_ram_size": 1638400,
    "max_total_spot_instances": 20,
    "max_total_spot_cores": 320,
    "max_total_spot_ram_size": 655360
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[tcloud]

| 参数名称                     | 参数类型   | 描述                                         |
|--------------------------|--------|--------------------------------------------|
| max_image_meta           | int32 | 镜像元数据最大的长度。     |
| max_personality          | int32 | 可注入文件的最大个数。     |
| max_personality_size     | int32 | 注入文件内容的最大长度（单位：Byte）。     |
| max_security_group_rules | int32 | 安全组中安全组规则最大的配置个数。   > 说明：  - 具体配额限制请以VPC配额限制为准。     |
| max_security_groups      | int32 | 安全组最大使用个数。  > 说明：  - 具体配额限制请以VPC配额限制为准。     |
| max_server_group_members | int32 | 服务器组中的最大虚拟机数。     |
| max_server_groups        | int32 | 服务器组的最大个数。     |
| max_server_meta          | int32 | 可输入元数据的最大长度。     |
| max_total_cores          | int32 | CPU核数最大申请数量。     |
| max_total_floating_ips   | int32 | 最大的浮动IP使用个数。     |
| max_total_instances      | int32 | 云服务器最大申请数量。     |
| max_total_keypairs       | int32 | 可以申请的SSH密钥对最大数量。     |
| max_total_ram_size       | int32 | 内存最大申请容量（单位：MB）。     |
| max_total_spot_instances | int32 | 竞价实例的最大申请数量。     |
| max_total_spot_cores     | int32 | 竞价实例的CPU核数最大申请数量。     |
| max_total_spot_ram_size  | int32 | 竞价实例的内存最大申请容量（单位：MB）。     |
