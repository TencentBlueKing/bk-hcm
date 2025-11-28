### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源申请单据列表查询。

### URL

POST /api/v1/woa/task/findmany/apply

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述                                                                                             |
|--------------|--------------|----|------------------------------------------------------------------------------------------------|
| bk_biz_id    | int array    | 否  | 业务ID                                                                                           |
| order_id	    | int array    | 否  | 资源申请单号，数量上限20个                                                                                 |
| suborder_id  | string array | 否  | 资源申请子单号，数量上限20个                                                                                |
| bk_username  | string	array | 否  | 提单人，数量上限20个                                                                                    |
| require_type | int array    | 否  | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤; 6: 滚服项目; 7: 小额绿通，数量上限20个                                       |
| stage        | string array | 否  | 单据执行阶段。"UNCOMMIT": 未提交, "AUDIT": 审核中, "RUNNING": 生产中, "DONE": 已完成, "CONFIRMING": 待用户确认，数量上限20个 |
| start        | string	      | 否  | 单据创建时间过滤条件起点日期，格式如"2022-05-01"                                                                 |
| end          | string	      | 否  | 单据创建时间过滤条件终点日期，格式如"2022-05-01"                                                                 |
| page         | object	      | 否  | 分页信息                                                                                           |
| get_product  | bool         | 否  | 是否获取CVM生产数据                                                                                    |
| source       | string array | 否  | 枚举类型，"business"（业务单据）、"purchase_to_resource_pool"(资源池采购)，不传时默认值为所有类型的单据                        |

#### page

| 参数名称  | 参数类型 | 必选 | 描述                 |
|-------|------|----|--------------------|
| start | int  | 否  | 记录开始位置，start 起始值为0 |
| limit | int  | 是  | 每页限制条数，最大200       |

说明：默认按create_at降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_biz_id": [213],
  "order_id": [1001],
  "suborder_id": ["1001-1"],
  "bk_username": ["xxx"],
  "require_type": [
    1
  ],
  "stage": [
    "UNCOMMIT",
    "AUDIT",
    "RUNNING"
  ],
  "start": "2022-04-18",
  "end": "2022-04-25",
  "page": {
    "start": 0,
    "limit": 20
  },
  "get_product": false,
  "source": ["business"]
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
    "count": 1,
    "info": [
      {
        "order_id": 1001,
        "suborder_id": "1001-1",
        "bk_biz_id": 2,
        "bk_username": "admin",
        "require_type": 1,
        "resource_type": "QCLOUDCVM",
        "expect_time": "2022-05-01 20:00:00",
        "remark": "",
        "spec": {
          "device_type": "S3.6XLARGE64",
          "cpu": "",
          "mem": "",
          "disk": "",
          "image": "Tencent Linux Release 1.2 (tkernel2)",
          "network": "TENTHOUSAND",
          "region": "ap-shanghai",
          "zone": "ap-shanghai-2"
        },
        "anti_affinity_level": "ANTI_NONE",
        "stage": "RUNNING",
        "status": "Matching",
        "origin_num": 10,
        "total_num": 10,
        "success_num": 5,
        "pending_num": 5,
        "product_num": 5,
        "source": "business",
        "create_at": "2022-01-02T15:04:05.004Z",
        "update_at": "2022-01-02T15:04:05.004Z"
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

| 参数名称  | 参数类型         | 描述             |
|-------|--------------|----------------|
| count | int          | 当前规则能匹配到的总记录条数 |
| info  | object array | 镜像信息列表         |

#### data.info

| 参数名称                | 参数类型      | 描述                                                                                                        |
|---------------------|-----------|-----------------------------------------------------------------------------------------------------------|
| order_id            | int       | 资源申请单号                                                                                                    |
| suborder_id         | string    | 资源申请子单号                                                                                                   |
| bk_biz_id	          | int	      | 业务ID                                                                                                      |
| bk_username         | 	string   | 提单人                                                                                                       |
| require_type        | 	int	     | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤; 6: 滚服项目; 7: 小额绿通                                                          |
| resource_type       | 	string   | 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器                       |
| expect_time         | 	string   | 期望交付时间                                                                                                    |
| remark	             | string    | 备注                                                                                                        |
| spec	               | object    | 资源需求明细                                                                                                    |
| anti_affinity_level | 	string   | 反亲和策略，默认值为"ANTI_NONE"。 "ANTI_NONE": 无要求, "ANTI_CAMPUS": 分Campus, "ANTI_MODULE": 分Module, "ANTI_RACK": 分机架 |
| stage	              | string    | 单据执行阶段。"UNCOMMIT": 未提交, "AUDIT": 审核中, "RUNNING": 生产中, "DONE": 已完成, "CONFIRMING": 待用户确认                             |
| status              | 	string   | 单据状态。WaitForMatch：待匹配，Matching：匹配执行中，MatchedSome：已完成部分资源匹配，Paused：已暂停，Done：完成, "CONFIRMING": 待用户确认   |
| origin_num          |     int      | 原始需求总数(不会变动)                                                                                           |
| total_num           | 	int	     | 资源需求总数                                                                                                    |
| success_num         | 	int	     | 已交付的资源数量                                                                                                  |
| pending_num         | 	int	     | 待匹配的资源数量                                                                                                  |
| product_num         | 	int	     | 已生产的资源数量(get_product为true时返回该字段)                                                                    |
| start_at	           | timestamp | 	步骤开始时间                                                                                                   |
| end_at	             | timestamp | 	步骤结束时间                                                                                                   |

#### spec

| 参数名称      | 参数类型     | 描述                                                |
|-----------|----------|---------------------------------------------------|
| region    | string	  | 地域                                                |
| zone      | string	  | 可用区                                               |
| device_group | string	  | 机型类别                                              |
| device_type | string	  | 机型                                                |
| image_id  | string   | 镜像ID                                              |
| image     | string   | 镜像名                                               |
| disk_size | int      | 数据盘磁盘大小，单位G                                       |
| disk_type	 | string	  | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| network_type | string	  | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆         |
| vpc	      | string   | 私有网络，默认为空                                         |
| subnet    | string   | 私有子网，默认为空                                         |
| os_type	  | string	  | 操作系统                                              |
| raid_type	 | string	  | RAID类型                                            |
| isp	      | string	  | 外网运营商                                             |
| mount_path | string	  | 数据盘挂载点                                            |
| cpu_provider | string	  | CPU类型                                             |
| kernel	   | string	  | 内核                                                |
| source    | string   | 枚举类型，"business"（业务单据）、"purchase_to_resource_pool"(资源池采购)     |