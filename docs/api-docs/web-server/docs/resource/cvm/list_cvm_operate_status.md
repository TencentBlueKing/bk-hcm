### 描述

- 该接口提供版本：v1.8.1+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询虚拟机可进行电源操作状态列表, 如开关机、重启、重装。

### URL

POST /api/v1/cloud/cvms/list/operate/status

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述                                  |
|--------------|--------------|----|-------------------------------------|
| ids          | string array | 是  | 要查询的主机ID数组，数量最大500                  |
| operate_type | string       | 是  | 操作类型(可选值：start, stop, reboot,reset) |


### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "ids": ["00000001"],
  "operate_type": "start"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "details":[
      {
        "id": "00000001",
        "name": "xxxxxx",
        "bk_host_id":17,
        "bk_host_name": "xxxxxx",
        "bk_asset_id":"xxxxxxx",
        "private_ipv4_addresses":["10.0.0.1"],
        "private_ipv6_addresses":["10.0.0.1"],
        "public_ipv4_addresses":["10.0.0.1"],
        "public_ipv6_addresses":["10.0.0.1"],
        "operator":"xx",
        "bak_operator":"xx",
        "device_type":"D4-8-100-10",
        "region": "南京-xx",
        "zone": "南京-xx",
        "bk_os_name": "xxxxxx",
        "topo_module": "空闲机",
        "bk_svr_source_type_id": "1",
        "status":"运行中",
        "srv_status":"使用中",
        "operate_status":0
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

| 参数名称    | 参数类型         | 描述       |
|---------|--------------|----------|
| details | object array | 主机状态信息列表 |

#### data.info

| 参数名称                   | 参数类型          | 描述                                  |
|------------------------|---------------|-------------------------------------|
| id                     | string	       | 主机唯一ID                              |
| name                   | string	       | 主机实例名称                             |
| bk_host_id             | int	           | CC主机ID                              |
| bk_host_name           | string	       | CC主机名称                                |
| bk_asset_id            | string	       | 设备固资号                               |
| private_ipv4_addresses | string array	 | 内网ipv4                              |
| private_ipv6_addresses | string array  | 内网ipv6                              |
| public_ipv4_addresses	 | string array  | 外网ipv4                              |
| public_ipv6_addresses  | string array  | 外网ipv6                              |
| operator               | string	       | 主机负责人                               |
| bak_operator           | string	       | 主机备份负责人                             |
| device_type            | string	       | 机型                                  |
| region                 | string        | 地域                                  |
| zone                   | string        | 可用区                                 |
| bk_os_name             | string        | 操作系统名称                              |
| topo_module            | string	       | 模块名称                                |
| bk_svr_source_type_id  | string        | 服务来源类型ID(0:未知1:自有2:托管3:租用4:虚拟机5:容器) |
| status	                | string        | 主机状态                                |
| srv_status             | string        | CC的运营状态                             |
| operate_status         | int   	       | 可操作状态(0:正常1:不是主备负责人2:不在空闲机模块3:云服务器未处于关机状态4:云服务器未处于开机状态5:物理机不支持操作)       |
