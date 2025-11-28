### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台管理-自研云资源-主机申领。
- 该接口功能描述：修改资源申请主单需求。

### URL

POST /api/v1/woa/task/apply/ticket/demand/update

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述             |
|------------|--------------|------|-----------------|
| ticket_id  | int	   | 是	  | 资源申请主单号     |
| suborders  | object array | 是  | 资源申请子需求单信息       |

#### suborders[0]

| 参数名称                | 参数类型   | 必选 | 描述                                                                                                        |
|---------------------|--------|----|-----------------------------------------------------------------------------------------------------------|
| resource_type	      | string | 是	 | 需求资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器                     |
| replicas		          | int	   | 是	 | 需求资源数量                                                                                                    |
| anti_affinity_level | string | 否	 | 反亲和策略，默认值为"ANTI_NONE"。 "ANTI_NONE": 无要求, "ANTI_CAMPUS": 分Campus, "ANTI_MODULE": 分Module, "ANTI_RACK": 分机架 |
| remark	             | string | 否	 | 备注                                                                                                        |
| source              | string | 否  | 枚举类型，"business"（业务单据）、"purchase_to_resource_pool"(资源池采购)，默认值为"business"                                   |
| spec	               | object | 是	 | 资源需求声明                                                                                                    |

#### spec for QCLOUDCVM

| 参数名称       | 参数类型            | 必选 | 描述                                                                  |
|---------------|-------------------|------|----------------------------------------------------------------------|
| region        | string            | 是   | 地域                                                                  |
| zone          | string            | 是   | 可用区                                                                |
| device_type   | string            | 是   | 机型                                                                  |
| image_id      | string            | 是   | 镜像ID                                                                 |
| disk_size     | int               | 否   | 数据盘磁盘大小，单位G（已废弃，用data_disk参数替代）                         |
| disk_type	    | string            | 否   | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘（已废弃，用data_disk参数替代） |
| network_type  | string            | 是   | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆                       |
| vpc	        | string            | 否   | 私有网络，默认为空                                                       |
| subnet        | string            | 否   | 私有子网，默认为空                                                       |
| charge_type   | string            | 否   | 计费模式 (PREPAID:包年包月，POSTPAID_BY_HOUR:按量计费)，默认:包年包月        |
| charge_months | int               | 否   | 计费时长，单位：月(计费模式为包年包月时，该字段必传)                           |
| system_disk   | DiskObject        | 是   | 系统盘，磁盘大小：50G-1000G且为50的倍数（IT类型默认本地盘、50G；其他类型默认高性能云盘、100G） |
| data_disk     | array DiskObject  | 否   | 数据盘，支持多块硬盘，磁盘大小：20G-32000G且为10的倍数，数据盘数量总和不能超过20块 |

#### spec for IDCPM

| 参数名称         | 参数类型    | 必选 | 描述                                        |
|--------------|---------|----|-------------------------------------------|
| region       | string  | 是  | 地域                                        |
| zone         | string  | 是  | 可用区                                       |
| device_type  | string	 | 是  | 机型                                        |
| os_type      | string	 | 是  | 操作系统                                      |
| raid_type    | string	 | 是  | RAID类型                                    |
| network_type | string  | 是  | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆 |
| isp          | string  | 否  | 外网运营商                                     |

#### spec for QCLOUDDVM

| 参数名称         | 参数类型    | 必选 | 描述                                        |
|--------------|---------|----|-------------------------------------------|
| region       | string	 | 是  | 地域                                        |
| zone	        | string  | 是  | 可用区                                       |
| device_group | string	 | 是  | 机型类别                                      |
| device_type  | string	 | 是  | 机型                                        |
| image	       | string  | 是  | 镜像名                                       |
| mount_path   | string	 | 是  | 数据盘挂载点                                    |
| network_type | string  | 是  | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆 |
| cpu_provider | string	 | 是  | CPU类型                                     |

#### spec for IDCDVM

| 参数名称         | 参数类型    | 必选 | 描述                                        |
|--------------|---------|----|-------------------------------------------|
| region	      | string  | 是  | 地域                                        |
| zone	        | string  | 是  | 可用区                                       |
| device_group | string	 | 是  | 机型类别                                      |
| device_type  | string	 | 是  | 机型                                        |
| image	       | string	 | 是  | 镜像名                                       |
| kernel	      | string  | 是  | 内核                                        |
| mount_path   | string  | 是  | 数据盘挂载点                                    |
| network_type | string  | 是  | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆 |

#### spec for DiskObject
| 参数名称   | 参数类型  | 必选 | 描述                                                      |
|-----------|---------|------|----------------------------------------------------------|
| disk_type | string  | 是   | 磁盘类型，"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘, "LOCAL_BASIC": 本地盘（仅特定机型的系统盘支持） |
| disk_size | int     | 是   | 磁盘大小，单位G                                             |
| disk_num  | int     | 是   | 磁盘数量，所有磁盘数量之和不能超过20块                          |

### 调用示例

```json
{
  "ticket_id": 1001,
  "suborders": [
    {
      "resource_type": "QCLOUDCVM",
      "replicas": 2,
      "anti_affinity_level": "ANTI_NONE",
      "remark": "",
      "source": "business",
      "spec": {
        "region": "ap-shanghai",
        "zone": "ap-shanghai-2",
        "device_type": "S3.LARGE8",
        "image_id": "img-r5igp4bv",
        "disk_size": 200,
        "disk_type": "CLOUD_PREMIUM",
        "network_type": "TENTHOUSAND",
        "vpc": "",
        "subnet": "",
        "charge_type": "PREPAID",
        "charge_months": 1,
        "system_disk": {
          "disk_type": "CLOUD_PREMIUM",
          "disk_size": 100,
          "disk_num": 1
        },
        "data_disk": [{
          "disk_type": "CLOUD_PREMIUM",
          "disk_size": 100,
          "disk_num": 1
        }]
      }
    },
    {
      "resource_type": "QCLOUDCVM",
      "replicas": 2,
      "anti_affinity_level": "ANTI_NONE",
      "remark": "",
      "source": "purchase_to_resource_pool",
      "spec": {
        "region": "ap-shanghai",
        "zone": "ap-shanghai-2",
        "device_type": "S3.LARGE8",
        "image_id": "img-r5igp4bv",
        "disk_size": 200,
        "disk_type": "CLOUD_PREMIUM",
        "network_type": "TENTHOUSAND",
        "vpc": "",
        "subnet": "",
        "charge_type": "PREPAID",
        "charge_months": 1,
        "system_disk": {
          "disk_type": "CLOUD_PREMIUM",
          "disk_size": 100,
          "disk_num": 1
        },
        "data_disk": [
          {
            "disk_type": "CLOUD_PREMIUM",
            "disk_size": 100,
            "disk_num": 1
          }
        ]
      }
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":null
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object       | 响应数据             |
