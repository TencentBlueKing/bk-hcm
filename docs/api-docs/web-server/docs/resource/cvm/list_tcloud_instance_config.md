### 描述

- 该接口提供版本：v1.8.6+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询腾讯云机型配置信息。

### URL

POST /api/v1/cloud/vendors/{vendor}/instances/config/query_from_cloud

### 请求参数
| 参数名称    | 参数类型       | 必选 | 描述                                      |
|------------|--------------|-----|-------------------------------------------|
| bk_biz_id  | int64        | 是  | 业务ID                                     |
| vendor     | string       | 是  | 云厂商（枚举值：tcloud，当前版本暂只支持tcloud）|
| account_id | string       | 是  | 账号ID                                     |
| region     | string       | 是  | 地域ID（唯一标识）                           |
| filters    | FilterExp    | 否  | 查询条件                                    |

#### FilterExp

每次请求的Filters的上限为10，Filters.Values的上限为100

| 参数名称 | 参数类型      | 必选 | 描述   |
|--------|--------------|-----|--------|
| name   | string       | 是  | 过滤条件 |
| values | string array | 是  | 过滤值  |

| 过滤条件              | 必选 | 描述                                                                                                               |
|----------------------|-----|-------------------------------------------------------------------------------------------------------------------|
| zone                 | 否  | 按照【可用区】进行过滤，可用区形如：ap-guangzhou-1                                                                       |
| instance-family      | 否  | 按照【实例机型系列】进行过滤，实例机型系列形如：S1、I1、M1等                                                                |
| instance-type        | 否  | 按照【实例机型】进行过滤                                                                                               |
| instance-charge-type | 否  | 按照【实例计费模式】进行过滤（PREPAID：表示预付费，即包年包月 | POSTPAID_BY_HOUR：表示后付费，即按量计费 | CDHPAID：表示独享子机 | SPOTPAID：表示竞价付费 | CDCPAID：表示专用集群付费）  |
| sort-keys            | 否  | 按照【关键字】进行排序，格式为排序字段加排序方式，中间用冒号分隔，例如： 按cpu数逆序排序 "cpu:desc", 按mem大小顺序排序 "mem:asc"    |

### 调用示例
#### 请求参数示例
```json
{
  "account_id": "00000001",
  "region": "ap-guangzhou",
  "filters": [
    {
      "name": "instance-type",
      "values": ["IT5.8XLARGE128"]
    }
  ]
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
            "zone": "ap-nanjing-1"
            "instance_type": "DA5.12XLARGE144",
            "instance_charge_type": "PREPAID",
            "network_card": 100,
            "externals": {
                "unsupport_networks": [
                    "BASIC",
                    "VPC1.0"
                ],
                "storage_block_attr": {
                    "max_size": 3570,
                    "min_size": 3570,
                    "type": "LOCAL_NVME"
                },
            },
            "cpu": 48,
            "memory": 144,
            "instance_family": "DA5",
            "type_name": "大数据型DA5",
            "local_disk_type_list": [
              {
                "max_size": 100,
                "min_size": 0,
                "partition_type": "ROOT",
                "required": "OPTIONAL",
                "type": "LOCAL_SSD"
              }
            ],
            "status": "SELL",
            "instance_bandwidth": 32,
            "instance_pps": 420,
            "storage_block_amount": 12,
            "cpu_type": "AMD Bergamo",
            "gpu": 0,
            "fpga": 0,
            "remark": "搭载 12 块 18627 GB  SATA HDD 本地硬盘",
            "gpu_count": 0,
            "frequency": "-/3.1GHz",
            "status_category": "EnoughStock"
          }
        ]
    }
}
```
### 响应参数说明

| 参数名称 | 参数类型 | 描述   |
|---------|--------|--------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | Data   | 响应数据 |

#### Data
| 参数名称 | 参数类型               | 描述             |
|--------|-----------------------|------------------|
| detail | InstanceConfig Array  | 查询返回的数据     |

#### InstanceConfig[n]

| 参数名称              | 参数类型            | 描述                                                               |
|----------------------|-------------------|--------------------------------------------------------------------|
| zone                 | string            | 可用区 示例值：ap-guangzhou-2                                        |
| instance_type        | string            | 实例机型 示例值：S5.LARGE4                                           |
| instance_charge_type | string            | 实例计费模式。取值范围：PREPAID、POSTPAID_BY_HOUR                     |
| network_card         | int64             | 网卡类型，例如：25代表25G网卡                                         |
| externals            | InstanceExternals | 扩展属性                                                           |
| cpu                  | int64             | 实例的CPU核数，单位：核                                              |
| memory               | int64             | 实例内存容量，单位：GB                                               |
| instance_family      | string            | 实例机型系列 示例值：S5                                              |
| type_name            | string            | 机型名称 示例值：标准型S5                                            |
| local_disk_type_list | []LocalDiskType   | 本地磁盘规格列表。当该参数返回为空值时，表示当前情况下无法创建本地盘。       |
| status               | string            | 实例是否售卖                                                         |
| instance_bandwidth   | float64           | 内网带宽，单位Gbps                                                   |
| instance_pps         | int64             | 网络收发包能力，单位万PPS                                             |
| storage_block_amount | int64             | 本地存储块数量                                                       |
| cpu_type             | string            | 处理器型号                                                           |
| gpu                  | int64             | 实例的GPU数量                                                        |
| fpga                 | int64             | 实例的FPGA数量                                                       |
| remark               | string            | 实例备注信息                                                         |
| gpu_count            | float64           | 实例机型映射的物理GPU卡数，单位：卡                                     |
| frequency            | string            | 实例的CPU主频信息                                                    |
| status_category      | string            | 描述库存情况                                                         |

##### InstanceConfig[n].InstanceExternals
| 参数名称             | 参数类型            | 描述                                                        |
|---------------------|-------------------|-------------------------------------------------------------|
| release_address     | bool              | 释放地址                                                      |
| unsupport_networks  | string array      | 不支持的网络类型，取值范围：BASIC：基础网络 VPC1.0：私有网络VPC1.0   |
| storage_block_attr  | StorageBlock      | 实例计费模式。取值范围：PREPAID、POSTPAID_BY_HOUR                |

##### InstanceConfig[n].InstanceExternals.StorageBlock
| 参数名称   | 参数类型    | 描述                              |
|-----------|-----------|-----------------------------------|
| type      | string    | HDD本地存储类型，示例值：LOCAL_PRO   |
| min_size  | int64     | HDD本地存储的最小容量。单位：GiB      |
| max_size  | int64     | HDD本地存储的最大容量。单位：GiB      |

##### InstanceConfig[n].LocalDiskType
| 参数名称        | 参数类型    | 描述                                                             |
|----------------|-----------|------------------------------------------------------------------|
| type           | string    | 本地磁盘类型                                                       |
| partition_type | string    | 本地磁盘属性                                                       |
| min_size       | int64     | 本地磁盘最小值                                                     |
| max_size       | int64     | 本地磁盘最大值                                                     |
| required       | string    | 购买时本地盘是否为必选。取值范围：REQUIRED：表示必选 OPTIONAL：表示可选   |
