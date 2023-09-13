### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询虚拟机机型列表。

### URL

POST /api/v1/web/instance_types/list

### 请求参数

| 参数名称                 | 参数类型   | 必选 | 描述                                                                                                                    |
|----------------------|--------|----|-----------------------------------------------------------------------------------------------------------------------|
| account_id           | string | 是  | 账号ID                                                                                                                  |
| vendor               | string | 是  | 供应商                                                                                                                   |
| region               | string | 是  | 地域                                                                                                                    |
| zone                 | string | 是  | 可用区                                                                                                                   |
| instance_charge_type | string | 是  | 计费类型（PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：表示专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费。） |

### 调用示例

#### 请求参数示例

查询腾讯云机型列表。

```json
{
  "account_id": "00000003",
  "vendor": "tcloud",
  "region": "ap-guangzhou",
  "zone": "ap-guangzhou-4"
}
```

#### 返回参数示例

查询腾讯云机型列表返回参数。

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "zone": "ap-guangzhou-4",
      "instance_type": "S4.MEDIUM2",
      "instance_family": "S4",
      "gpu": 0,
      "cpu": 2,
      "memory": 2048,
      "fpga": 0,
      "status": "SELL"
    },
    {
      "zone": "ap-guangzhou-4",
      "instance_type": "S4.MEDIUM2",
      "instance_family": "S4",
      "gpu": 0,
      "cpu": 2,
      "memory": 2048,
      "fpga": 0,
      "status": "SELL"
    },
    {
      "zone": "ap-guangzhou-4",
      "instance_type": "S4.MEDIUM2",
      "instance_family": "S4",
      "gpu": 0,
      "cpu": 2,
      "memory": 2048,
      "fpga": 0,
      "status": "SELL"
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[tcloud]

| 参数名称            | 参数类型   | 描述                                         |
|-----------------|--------|--------------------------------------------|
| instance_type   | string | 实例机型。     |
| instance_family | string | 实例机型系列。                                      |
| gpu             | int64  | 实例的GPU数量。                                    |
| cpu             | int64  | 实例的CPU核数，单位：核。                                   |
| memory          | int64  | 实例内存容量，单位：GB。                                   |
| fpga            | int64  | 实例的FPGA数量。                                   |
| status          | string | 实例是否售卖。取值范围：SELL：表示实例可购买、SOLD_OUT：表示实例已售罄。 |
| cpu_type          | string  | cpu类型。 |
| InstanceBandwidth | float64 | 最大带宽。 |
| InstancePps       | int64   | 内网收发包。 |
| Price             | object  | 实例价格。 |

#### data[huawei]

| 参数名称            | 参数类型   | 描述                                         |
|-----------------|--------|--------------------------------------------|
| instance_type   | string | 实例机型。                                      |
| instance_family | string | 实例机型系列。                                    |
| gpu             | int64  | 实例的GPU数量。                                  |
| cpu             | int64  | 实例的CPU核数，单位：核。                             |
| memory          | int64  | 实例内存容量，单位：GB。                              |
| fpga            | int64  | 实例的FPGA数量。                                 |
| ecsperformancetype           | string  | 机型族。                                 |
| name                         | string  | 机型名称。                                 |
| ecsinstance_architecture     | string  | CPU架构。                                 |
| infocpuname                  | string  | CPU型号。                                 |
| quotamax_rate                | string  | 最大带宽。                                 |
| quotamax_pps                 | string  | 内网收发包。                                 |

#### data[gcp]

| 参数名称            | 参数类型   | 描述                                         |
|-----------------|--------|--------------------------------------------|
| instance_type   | string | 实例机型。     |
| kind            | string | 实例机型系列。                                      |
| gpu             | int64  | 实例的GPU数量。                                    |
| cpu             | int64  | 实例的CPU核数，单位：核。                                   |
| memory          | int64  | 实例内存容量，单位：GB。                                   |
| fpga            | int64  | 实例的FPGA数量。                                   |

#### data[azure]

| 参数名称            | 参数类型   | 描述                                         |
|-----------------|--------|--------------------------------------------|
| instance_type   | string | 实例机型。     |
| kind            | string | 实例机型系列。                                      |
| gpu             | int64  | 实例的GPU数量。                                    |
| cpu             | int64  | 实例的CPU核数，单位：核。                                   |
| memory          | int64  | 实例内存容量，单位：GB。                                   |
| fpga            | int64  | 实例的FPGA数量。                                   |
| architecture    | string | cpu架构。                                   |

#### data[aws]

| 参数名称            | 参数类型   | 描述                                         |
|-----------------|--------|--------------------------------------------|
| instance_type   | string | 实例机型。     |
| kind            | string | 实例机型系列。                                      |
| gpu             | int64  | 实例的GPU数量。                                    |
| cpu             | int64  | 实例的CPU核数，单位：核。                                   |
| memory          | int64  | 实例内存容量，单位：GB。                                   |
| fpga            | int64  | 实例的FPGA数量。                                   |
| architecture    | string | cpu架构。                                   |
| network_performance | string | 网络性能。                                   |
| disk_size_in_gb     | int64  | 磁盘大小。                                   |
| disk_type           | string | 磁盘类型。                                   |