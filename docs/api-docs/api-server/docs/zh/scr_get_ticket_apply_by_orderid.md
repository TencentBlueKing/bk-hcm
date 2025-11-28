### 描述

- 该接口提供版本：v1.8.5+。
- 该接口所需权限：无。
- 该接口功能描述：获取资源申请单据内容。

### URL

POST /api/v1/woa/task/get/apply/ticket

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述   |
|----------|------|----|------|
| order_id | int  | 是  | 单据ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id": 1001
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
    "order_id": 1001,
    "stage": "UNCOMMIT",
    "bk_biz_id": 3,
    "bk_username": "xx",
    "follower": "",
    "enable_notice": true,
    "require_type": 1,
    "expect_time": "2022-05-01 20:00:00",
    "remark": "",
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
          "system_disk": {
            "disk_type": "CLOUD_PREMIUM",
            "disk_size": 100,
            "disk_num": 1,
          },
          "data_disk": [{
            "disk_type": "CLOUD_PREMIUM",
            "disk_size": 100,
            "disk_num": 1,
          }]
        }
      },
      {
        "resource_type": "IDCPM",
        "replicas": 2,
        "anti_affinity_level": "ANTI_NONE",
        "remark": "",
        "source": "business",
        "spec": {
          "region": "东莞",
          "zone": "东莞-大朗",
          "device_type": "B7",
          "os_type": "XServer V16_64",
          "raid_type": "RAID1",
          "network_type": "TENTHOUSAND",
          "isp": ""
        }
      },
      {
        "resource_type": "QCLOUDDVM",
        "replicas": 2,
        "anti_affinity_level": "ANTI_NONE",
        "remark": "",
        "source": "business",
        "spec": {
          "region": "ap-shanghai",
          "zone": "ap-shanghai-2",
          "device_group": "GAMESERVER",
          "device_type": "D4-8-200-10",
          "image": "xxx.xxx.xxx/library/tlinux2.2:v1.6",
          "mount_path": "/data1",
          "network_type": "TENTHOUSAND",
          "cpu_provider": "Intel"
        }
      },
      {
        "resource_type": "IDCDVM",
        "replicas": 2,
        "anti_affinity_level": "ANTI_NONE",
        "remark": "",
        "source": "business",
        "spec": {
          "region": "上海",
          "zone": "上海-青浦",
          "device_group": "GAMESERVER",
          "device_type": "D4-8-200-10",
          "image": "xxx.xxx.xxx/library/tlinux2.2:v1.6",
          "kernel": "",
          "mount_path": "/data1",
          "network_type": "TENTHOUSAND"
        }
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

| 参数名称           | 参数类型       | 描述                                                               |
|----------------|------------|------------------------------------------------------------------|
| order_id       | int	       | 若order_id传值且非0，则更新order_id对应的申请单据草稿；若order_id未传值或为0，则创建申请单据草稿    |
| stage	         | string	    | 单据执行阶段。"UNCOMMIT": 未提交, "AUDIT": 审核中, "RUNNING": 生产中, "DONE": 已完成 |
| bk_biz_id      | int	       | CC业务ID                                                           |
| bk_username    | string     | 资源申请提单人                                                          |
| follower	      | string     | 关注人，如果有多人，以","分隔，如："name1,name2"                                 |
| enable_notice	 | bool	      | 是否通知用户单据完成，默认为false                                              |
| require_type   | int	       | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤                                   |
| expect_time    | string     | 期望交付时间                                                           |
| remark	        | string     | 备注                                                               |
| source              | string  | 枚举类型，"business"（业务单据）、"purchase_to_resource_pool"(资源池采购)，默认值为"business"    |
| suborders	     | object array | 资源申请子需求单信息                                                       |

#### data.suborders

| 参数名称                | 参数类型   | 必选                                                                                                        | 描述 |
|---------------------|--------|-----------------------------------------------------------------------------------------------------------|----|
| resource_type	      | string | 需求资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器                     |
| replicas		          | int	   | 需求资源数量                                                                                                    |
| anti_affinity_level | string | 反亲和策略，默认值为"ANTI_NONE"。 "ANTI_NONE": 无要求, "ANTI_CAMPUS": 分Campus, "ANTI_MODULE": 分Module, "ANTI_RACK": 分机架 |
| remark	             | string | 备注                                                                                                        |
| spec	               | object | 资源需求声明                                                                                                    |

#### data.suborders.spec

| 参数名称      | 参数类型            | 描述                                                  |
|--------------|-------------------|-----------------------------------------------------|
| region       | string	           | 地域                                                  |
| zone         | string	           | 可用区                                                 |
| device_group | string            | 机型类别                                                |
| device_type  | string	           | 机型                                                  |
| image_id     | string            | 镜像ID                                                |
| disk_size    | int               | 数据盘磁盘大小，单位G（已废弃，优先使用data_disk字段）                                             |
| disk_type	   | string	           | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘（已废弃，优先使用data_disk字段） |
| network_type | string	           | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆           |
| vpc	       | string            | 私有网络，默认为空                                           |
| subnet       | string            | 私有子网，默认为空                                           |
| os_type      | string	           | 操作系统                                                |
| raid_type    | string	           | RAID类型                                              |
| isp          | string	           | 外网运营商                                               |
| mount_path   | string	           | 数据盘挂载点                                              |
| cpu_provider | string	           | CPU类型                                               |
| kernel       | string            | 内核                                                  |
| system_disk  | DiskObject        | 系统盘，磁盘大小：50G-1000G且为50的倍数（IT类型默认本地盘、50G；其他类型默认高性能云盘、100G） |
| data_disk    | array DiskObject  | 数据盘，支持多块硬盘，磁盘大小：20G-32000G且为10的倍数，数据盘数量总和不能超过20块 |

#### data.suborders.spec.DiskObject
| 参数名称   | 参数类型  | 必选 | 描述                                                      |
|-----------|---------|------|----------------------------------------------------------|
| disk_type | string  | 是   | 磁盘类型，"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| disk_size | int     | 是   | 磁盘大小，单位G                                             |
| disk_num  | int     | 否   | 磁盘数量，所有磁盘数量之和不能超过20块，默认1块                  |
