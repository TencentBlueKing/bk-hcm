### 描述

- 该接口提供版本：v1.7.0+。
- 该接口所需权限：账号访问。
- 该接口功能描述：同步指定账号下指定资源。

### URL

POST /api/v1/cloud/vendors/{vendor}/accounts/{account_id}/resources/{res}/sync_by_cond

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述                                                  |
|------------|--------|----|-----------------------------------------------------|
| vendor     | string | 是  | 云厂商                                                 |
| account_id | string | 是  | 账号ID                                                |
| res        | string | 是  | 资源名称 目前仅支持 security_group, load_balancer(仅支持tcloud) |

> `res=load_balancer` 且 vendor 为 `tcloud` 时走异步任务：接口立即返回 `task_management_id`，实际同步在任务中心执行。其他资源仍为同步接口，`data` 为空。

#### vendor=tcloud

| 参数名称        | 参数类型                | 必选 | 描述               |
|-------------|---------------------|----|------------------|
| regions     | []string            | 是  | 指定资源同步地域，最少1，最大5 |
| cloud_ids   | []string            | 否  | 资源id，数量上限20      |
| tag_filters | map[string][]string | 否  | 指定同步标签过滤器标签，上限5  |


#### vendor=huawei

| 参数名称        | 参数类型                | 必选 | 描述               |
|-------------|---------------------|----|------------------|
| regions     | []string            | 是  | 指定资源同步地域，最少1，最大5 |

#### vendor=azure

| 参数名称                 | 参数类型     | 必选 | 描述                 |
|----------------------|----------|----|--------------------|
| resource_group_names | []string | 是  | 指定资源同步的资源组，最少1，最大5 |


#### vendor=aws

| 参数名称        | 参数类型                | 必选 | 描述               |
|-------------|---------------------|----|------------------|
| regions     | []string            | 是  | 指定资源同步地域，最少1，最大5 |

### 调用示例

1. 同步'ap-guangzhou'地域下，id 为"id-abcdefg"的资源

```json
{
  "regions": [
    "ap-guangzhou"
  ],
  "cloud_ids": [
    "id-abcdefg"
  ]
}
```

2. 同步'ap-guangzhou'地域下，标签biz='1234'的资源

```json
{
  "regions": [
    "ap-guangzhou"
  ],
  "tag_filters": {
    "biz": [
      "1234"
    ]
  }
}
```

### 响应示例

#### 同步资源（非 CLB）

```json
{
  "code": 0,
  "message": "ok"
}
```

#### 异步同步负载均衡（res=load_balancer）

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_management_id": "000000xw"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | CLB 异步同步时返回任务信息，其他资源为空 |

#### data（仅 res=load_balancer）

| 参数名称               | 参数类型   | 描述                          |
|--------------------|--------|-----------------------------|
| task_management_id | string | 任务管理 ID。请求条件下没有任何可处理的 CLB 时不创建任务，该字段为空字符串 |
