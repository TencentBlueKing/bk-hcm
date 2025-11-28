### 描述

- 该接口提供版本：v1.8.5+。
- 该接口所需权限：平台管理-自研云资源-主机申领。
- 该接口功能描述：创建资源申请单据。

### URL

POST /api/v1/woa/task/create/apply

### 输入参数

| 参数名称           | 参数类型         | 必选 | 描述                                               |
|----------------|--------------|----|--------------------------------------------------|
| bk_biz_id      | int	         | 是	 | CC业务ID                                           |
| bk_username    | string       | 是	 | 资源申请提单人                                          |
| follower	      | string array | 否	 | 关注人，如果有多人，以","分隔，如："name1,name2"                 |
| enable_notice	 | bool	        | 否	 | 是否通知用户单据完成，默认为false                              |
| require_type   | int	         | 是	 | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤; 6: 滚服项目; 7: 小额绿通 |
| expect_time    | string       | 是	 | 期望交付时间                                           |
| remark	        | string       | 否	 | 备注                                               |
| suborders	     | object array | 是  | 资源申请子需求单信息                                       |

#### suborders

| 参数名称                | 参数类型   | 必选 | 描述                                                                                                        |
|---------------------|--------|----|-----------------------------------------------------------------------------------------------------------|
| resource_type	      | string | 是	 | 需求资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器                     |
| replicas		          | int	   | 是	 | 需求资源数量                                                                                                    |
| anti_affinity_level | string | 否	 | 反亲和策略，默认值为"ANTI_NONE"。 "ANTI_NONE": 无要求, "ANTI_CAMPUS": 分Campus, "ANTI_MODULE": 分Module, "ANTI_RACK": 分机架 |
| enable_disk_check	  | bool   | 否  | 交付前是否执行本地盘压测，默认值为false                                                                                    |
| remark	             | string | 否	 | 备注                                                                                                        |
| source              | string | 否  | 枚举类型，"business"（业务单据）、"purchase_to_resource_pool"(资源池采购)，默认值为"business"                                   |
| spec	               | object | 是	 | 资源需求声明                                                                                                    |

#### spec for QCLOUDCVM

| 参数名称              | 参数类型          | 必选 | 描述                                                      |
|---------------------|------------------|-----|-----------------------------------------------------------|
| region              | string	         | 是  | 地域                                                       |
| zone                | string	         | 是  | 可用区（跟zones参数，需要传其中一个，该参数即将废弃）             |
| resource_mode	      | int              | 是  | 1: 按机型族申领, 0: 按机型申领                                |
| device_group        | string           | 否  | 机型族。当resource_mode为1时必填                             |
| model_type          | string           | 否  | 配置类型。当resource_mode为1时必填                            |
| device_type         | string	         | 是  | 机型。当resource_mode为0时必填                               |
| image_id            | string           | 是  | 镜像ID                                                     |
| disk_size           | int              | 否  | 数据盘磁盘大小，单位G                                         |
| disk_type	          | string	         | 否  | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| network_type        | string	         | 是  | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆           |
| vpc	              | string           | 否  | 私有网络，默认为空                                           |
| subnet              | string           | 否  | 私有子网，默认为空                                           |
| charge_type         | string           | 否  | 计费模式 (PREPAID:包年包月，POSTPAID_BY_HOUR:按量计费)，默认:包年包月 |
| charge_months       | int              | 否  | 计费时长，单位：月(计费模式为包年包月时，该字段必传)                    |
| inherit_instance_id | string           | 否  | 被继承云主机实例ID（同一批次只支持一台），如果是滚服项目，该字段必传       |
| system_disk         | DiskObject       | 是  | 系统盘，磁盘大小：50G-1000G且为50的倍数（IT类型默认本地盘、50G；其他类型默认高性能云盘、100G） |
| data_disk           | DiskObject array | 否  | 数据盘，支持多块硬盘，磁盘大小：20G-32000G且为10的倍数，数据盘数量总和不能超过20块 |
| zones               | string array     | 否  | 多可用区(选“全部”时传all)                                      |
| res_assign          | int              | 否  | 资源分配方式(1表示“有资源区域优先”、2表示“分Campus生产”)           |

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
| disk_type | string  | 是   | 磁盘类型，"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| disk_size | int     | 是   | 磁盘大小，单位G                                             |
| disk_num  | int     | 是   | 磁盘数量，所有磁盘数量之和不能超过20块                          |

### 调用示例

#### 获取详细信息请求参数示例

#### CVM申请示例

```json
{
  "bk_biz_id": 3,
  "bk_username": "xx",
  "follower": [],
  "enable_notice": true,
  "require_type": 1,
  "expect_time": "2022-05-01 20:00:00",
  "remark": "",
  "suborders": [
    {
      "resource_type": "QCLOUDCVM",
      "replicas": 2,
      "anti_affinity_level": "ANTI_NONE",
      "enable_disk_check": false,
      "remark": "",
      "source": "business",
      "spec": {
        "region": "ap-shanghai",
        "device_type": "S3.LARGE8",
        "image_id": "img-r5igp4bv",
        "disk_size": 200,
        "disk_type": "CLOUD_PREMIUM",
        "network_type": "TENTHOUSAND",
        "vpc": "",
        "subnet": "",
        "charge_type": "PREPAID",
        "charge_months": 1,
        "inherit_instance_id": "ins-111",
        "system_disk": {
        "disk_type": "CLOUD_PREMIUM",
        "disk_size": 100,
        "disk_num": 1,
        },
        "data_disk": [{
        "disk_type": "CLOUD_PREMIUM",
        "disk_size": 100,
        "disk_num": 1
        }],
        "zones": [
          "ap-nanjing-1",
          "ap-nanjing-2"
        ],
        "res_assign": 1
      }
    }
  ]
}
```

#### PM申请示例

```json
{
  "bk_biz_id": 3,
  "bk_username": "xx",
  "follower": [],
  "enable_notice": true,
  "require_type": 1,
  "expect_time": "2022-05-01 20:00:00",
  "remark": "",
  "suborders": [
    {
      "resource_type": "IDCPM",
      "replicas": 2,
      "anti_affinity_level": "ANTI_NONE",
      "remark": "",
      "spec": {
        "region": "东莞",
        "zone": "东莞-大朗",
        "device_type": "B7",
        "os_type": "XServer V16_64",
        "raid_type": "RAID1",
        "network_type": "TENTHOUSAND",
        "isp": ""
      }
    }
  ]
}
```

#### QCLOUDDVM申请示例

```json
{
  "bk_biz_id": 3,
  "bk_username": "xx",
  "follower": [],
  "enable_notice": true,
  "require_type": 1,
  "expect_time": "2022-05-01 20:00:00",
  "remark": "",
  "suborders": [
    {
      "resource_type": "QCLOUDDVM",
      "replicas": 2,
      "anti_affinity_level": "ANTI_NONE",
      "remark": "",
      "spec": {
        "region": "ap-shanghai",
        "zone": "ap-shanghai-2",
        "device_group": "GAMESERVER",
        "device_type": "D4-8-200-10",
        "image": "test.xxx/library/tlinux2.2:v1.6",
        "mount_path": "/data1",
        "network_type": "TENTHOUSAND",
        "cpu_provider": "Intel"
      }
    }
  ]
}
```

#### IDCDVM申请示例

```json
{
  "bk_biz_id": 3,
  "bk_username": "xx",
  "follower": [],
  "enable_notice": true,
  "require_type": 1,
  "expect_time": "2022-05-01 20:00:00",
  "remark": "",
  "suborders": [
    {
      "resource_type": "IDCDVM",
      "replicas": 2,
      "anti_affinity_level": "ANTI_NONE",
      "remark": "",
      "spec": {
        "region": "上海",
        "zone": "上海-青浦",
        "device_group": "GAMESERVER",
        "device_type": "D4-8-200-10",
        "image": "test.xxx/library/tlinux2.2:v1.6",
        "kernel": "",
        "mount_path": "/data1",
        "network_type": "TENTHOUSAND"
      }
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "order_id": 1001
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

| 参数名称     | 参数类型 | 描述   |
|----------|------|------|
| order_id | int  | 单据ID |
