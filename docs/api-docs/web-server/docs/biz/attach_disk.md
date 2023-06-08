### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：挂载云盘。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/disks/attach

### 输入参数

#### TCloud
| 参数名称               | 参数类型     | 必选  | 描述     |
|--------------------|----------|-----|--------|
| bk_biz_id          | int64    | 是   | 业务ID   |
| account_id         | string   | 是   | 账号 ID  |
| cvm_id             | string   | 是   | 虚拟机 ID |
| disk_id            | string   | 是   | 云盘 ID  |

#### Aws
| 参数名称        | 参数类型    | 必选  | 描述     |
|-------------|---------|-----|--------|
| bk_biz_id   | int64   | 是   | 业务ID   |
| disk_id     | string  | 是   | 云盘 ID  |
| cvm_id      | string  | 是   | 虚拟机 ID |
| device_name | string  | 是   | 设备名称   |

#### HuaWei
| 参数名称        | 参数类型    | 必选   | 描述     |
|-------------|---------|------|--------|
| bk_biz_id   | int64   | 是    | 业务ID   |
| disk_id     | string  | 是    | 云盘 ID  |
| cvm_id      | string  | 是    | 虚拟机 ID |

#### Gcp
| 参数名称         | 参数类型     | 必选    | 描述     |
|--------------|----------|-------|--------|
| bk_biz_id    | int64    | 是     | 业务ID   |
| disk_id      | string   | 是     | 云盘 ID  |
| cvm_id       | string   | 是     | 虚拟机 ID |

#### Azure
| 参数名称         | 参数类型      | 必选   | 描述                                |
|--------------|-----------|------|-----------------------------------|
| bk_biz_id    | int64     | 是    | 业务ID                              |
| disk_id      | string    | 是    | 云盘 ID                             |
| cvm_id       | string    | 是    | 虚拟机 ID                            |
| caching_type | string    | 是    | 缓存类型（枚举值：None、ReadOnly、ReadWrite） |

### 调用示例

如挂载云厂商是 tcloud , 账号 ID 是 00000003 ，虚拟机 ID 是 00000050，云盘 ID 是00000050的云盘。

```json
{
    "account_id": "00000003", 
    "cvm_id": "00000050",
    "disk_id": "00000050"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称     | 参数类型   | 描述   |
|----------|--------|------|
| code     | int    | 状态码  |
| message  | string | 请求信息 |
